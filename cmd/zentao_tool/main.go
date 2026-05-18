// Package main 是应用程序入口
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/jan2xue/zentao_import_story/internal/config"
	"github.com/jan2xue/zentao_import_story/internal/excel"
	"github.com/jan2xue/zentao_import_story/internal/logger"
	"github.com/jan2xue/zentao_import_story/internal/zentao"
	"github.com/jan2xue/zentao_import_story/pkg/story"
)

// envVarOverrides maps environment variable names to config field setters.
// Environment variables take precedence over config.yaml values.
var envVarOverrides = []struct {
	envKey string
	apply  func(cfg *config.Config, val string)
}{
	{"ZENTAO_URL", func(cfg *config.Config, val string) { cfg.ZentaoURL = val }},
	{"ZENTAO_USERNAME", func(cfg *config.Config, val string) { cfg.ZentaoUsername = val }},
	{"ZENTAO_PASSWORD", func(cfg *config.Config, val string) { cfg.ZentaoPassword = val }},
	{"ZENTAO_REVIEWER", func(cfg *config.Config, val string) { cfg.DefaultReviewer = val }},
}

func main() {
	// 初始化日志记录器
	log, err := logger.NewLogger()
	if err != nil {
		fmt.Printf("初始化日志记录器失败: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	excelPath := flag.String("excel", "requirements.xlsx", "Excel文件路径")
	action := flag.String("action", "import", "操作类型: import(导入)、delete(删除)")
	productID := flag.Int("product", 0, "产品ID（删除时必填）")
	titleFilter := flag.String("title", "", "标题筛选（删除时可选，部分匹配）")
	openedByFilter := flag.String("openedBy", "", "创建者筛选（删除时可选，精确匹配账号名）")
	flag.Parse()

	// 加载配置文件
	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatal("加载配置文件失败: %v", err)
	}

	// 如果命令行指定了Excel文件，覆盖配置文件中的设置
	if *excelPath != "" {
		cfg.ExcelFile = *excelPath
	}

	// 验证配置
	if err := validateConfig(cfg); err != nil {
		log.Fatal("%v", err)
	}

	// 根据操作类型执行相应功能
	switch *action {
	case "import":
		handleImport(cfg, log)
	case "delete":
		handleDelete(cfg, log, *productID, *titleFilter, *openedByFilter)
	default:
		log.Fatal("不支持的操作类型: %s，仅支持 import 或 delete", *action)
	}
}

// handleImport 处理导入操作
func handleImport(cfg *config.Config, log *logger.Logger) {
	// 创建Excel读取器
	reader, err := excel.NewReader(cfg.ExcelFile)
	if err != nil {
		log.Fatal("创建Excel读取器失败: %v", err)
	}
	defer reader.Close()

	// 读取需求数据
	stories, err := reader.ReadStories(cfg.DefaultPriority)
	if err != nil {
		log.Fatal("读取Excel数据失败: %v", err)
	}

	log.Info("从Excel中读取到 %d 个需求", len(stories))

	// 统计各类型数量
	epicCount, reqCount, storyCount := 0, 0, 0
	for _, s := range stories {
		switch s.Type {
		case story.StoryTypeEpic:
			epicCount++
		case story.StoryTypeRequirement:
			reqCount++
		case story.StoryTypeStory:
			storyCount++
		}
	}

	// 提取所有唯一的产品ID
	productIDSet := make(map[int]bool)
	for _, s := range stories {
		productIDSet[s.ProductID] = true
	}
	productIDs := make([]int, 0, len(productIDSet))
	for id := range productIDSet {
		productIDs = append(productIDs, id)
	}

	// 创建禅道客户端
	client, err := zentao.NewClient(cfg)
	if err != nil {
		log.Fatal("创建禅道客户端失败: %v", err)
	}

	// 获取产品名称信息
	productInfo, err := client.Product.GetProductInfo(productIDs)
	if err != nil {
		log.Error("获取产品信息失败: %v，将仅显示产品ID", err)
	}

	// 显示确认信息
	separator := strings.Repeat("=", 60)
	fmt.Printf("\n%s\n", separator)
	fmt.Printf("               导入确认\n")
	fmt.Printf("%s\n\n", separator)

	fmt.Printf("需求类型分布:\n")
	fmt.Printf("  业务需求(Epic):        %d 条\n", epicCount)
	fmt.Printf("  用户需求(Requirement): %d 条\n", reqCount)
	fmt.Printf("  研发需求(Story):      %d 条\n", storyCount)
	fmt.Printf("  合计:                  %d 条\n\n", len(stories))

	fmt.Printf("导入顺序: Epic → Requirement → Story\n\n")

	fmt.Printf("涉及产品:\n")
	fmt.Printf("%-10s %-40s %-10s\n", "产品ID", "产品名称", "需求数量")
	fmt.Printf("%-10s %-40s %-10s\n", "------", "----------------------------------------", "------")

	productCount := make(map[int]int)
	for _, s := range stories {
		productCount[s.ProductID]++
	}
	for _, productID := range productIDs {
		productName := productInfo[productID]
		if productName == "" {
			productName = "[未知产品]"
		}
		fmt.Printf("%-10d %-40s %-10d\n", productID, productName, productCount[productID])
	}

	fmt.Printf("\n%s\n", separator)
	fmt.Printf("\n⚠️  重要提示：\n")
	fmt.Printf("   1. 请仔细核对上述产品信息，错误的产品ID会导致数据导入错误产品\n")
	fmt.Printf("   2. 父需求引用(@行号)将在导入时自动解析为实际禅道ID\n")
	fmt.Printf("   3. 导入顺序为 Epic → Requirement → Story，确保父需求先创建\n")
	fmt.Printf("\n是否确认导入? (yes/no): ")

	confirmReader := bufio.NewReader(os.Stdin)
	response, err := confirmReader.ReadString('\n')
	if err != nil {
		log.Fatal("读取用户输入失败: %v", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "yes" && response != "y" {
		log.Info("取消导入操作")
		return
	}

	// 创建导入器
	importer := zentao.NewImporter(client, log)

	// 层级导入
	results := importer.ImportStories(stories)

	// 生成并打印报告
	report := importer.GenerateReport(results)
	log.Info("\n%s", report)

	log.Info("日志文件已保存至: %s", log.GetLogFilePath())

	hasFailure := false
	for _, result := range results {
		if !result.Success {
			hasFailure = true
		}
	}
	if hasFailure {
		os.Exit(1)
	}
}

// handleDelete 处理删除操作
// 必须指定产品ID，支持标题（部分匹配）和创建者作为可选过滤条件
func handleDelete(cfg *config.Config, log *logger.Logger, productID int, titleFilter, openedByFilter string) {
	if productID <= 0 {
		log.Fatal("删除操作必须指定产品ID (-product 参数)")
	}

	separator := strings.Repeat("=", 60)

	// 显示筛选条件
	fmt.Printf("\n%s\n", separator)
	fmt.Printf("           删除需求 — 筛选条件\n")
	fmt.Printf("%s\n\n", separator)
	fmt.Printf("  产品ID: %d\n", productID)
	if titleFilter != "" {
		fmt.Printf("  标题筛选: \"%s\"（部分匹配）\n", titleFilter)
	} else {
		fmt.Printf("  标题筛选: (无)\n")
	}
	if openedByFilter != "" {
		fmt.Printf("  创建者筛选: \"%s\"（精确匹配）\n", openedByFilter)
	} else {
		fmt.Printf("  创建者筛选: (无)\n")
	}

	// 创建禅道客户端
	client, err := zentao.NewClient(cfg)
	if err != nil {
		log.Fatal("创建禅道客户端失败: %v", err)
	}

	// 获取产品信息
	productInfo, err := client.Product.GetProductInfo([]int{productID})
	if err != nil {
		log.Error("获取产品信息失败: %v，继续执行", err)
	} else {
		if name, ok := productInfo[productID]; ok && name != "" {
			fmt.Printf("  产品名称: %s\n", name)
		}
	}

	// 查询匹配的需求
	log.Info("正在查询匹配的需求...")
	deleter := zentao.NewDeleter(client, log)
	filter := zentao.DeleteFilter{
		ProductID: productID,
		Title:     titleFilter,
		OpenedBy:  openedByFilter,
	}
	matchedItems := deleter.FetchByFilter(filter)

	if len(matchedItems) == 0 {
		fmt.Printf("\n未找到匹配的需求。\n")
		log.Info("未找到匹配的需求，产品ID=%d，标题=%s，创建者=%s", productID, titleFilter, openedByFilter)
		return
	}

	// 显示匹配结果
	fmt.Printf("\n%s\n", separator)
	fmt.Printf("           匹配结果\n")
	fmt.Printf("%s\n\n", separator)
	fmt.Print(zentao.FormatMatchedList(matchedItems))

	// 二次确认
	fmt.Printf("\n%s\n", separator)
	fmt.Printf("\n⚠️  警告: 即将删除以上 %d 个需求！\n", len(matchedItems))
	conditions := []string{fmt.Sprintf("产品ID=%d", productID)}
	if titleFilter != "" {
		conditions = append(conditions, fmt.Sprintf("标题包含\"%s\"", titleFilter))
	}
	if openedByFilter != "" {
		conditions = append(conditions, fmt.Sprintf("创建者=\"%s\"", openedByFilter))
	}
	fmt.Printf("   筛选条件: %s\n", strings.Join(conditions, ", "))
	fmt.Printf("   此操作不可撤销！\n")
	fmt.Printf("\n请输入 \"yes\" 确认删除: ")

	reader := bufio.NewReader(os.Stdin)
	confirmInput, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("读取用户输入失败: %v", err)
	}

	confirmInput = strings.TrimSpace(strings.ToLower(confirmInput))
	if confirmInput != "yes" && confirmInput != "y" {
		log.Info("取消删除操作")
		return
	}

	// 执行删除（大批量时使用并发）
	var results []zentao.DeleteResult
	if len(matchedItems) > 20 {
		log.Info("大批量删除(>20条)，使用并发模式")
		results = deleter.DeleteStoriesConcurrent(matchedItems, 5)
	} else {
		results = deleter.DeleteStories(matchedItems)
	}

	// 生成并打印报告
	report := deleter.GenerateDeleteReport(results)
	log.Info("\n%s", report)

	log.Info("日志文件已保存至: %s", log.GetLogFilePath())

	hasFailure := false
	for _, result := range results {
		if !result.Success {
			hasFailure = true
		}
	}
	if hasFailure {
		os.Exit(1)
	}
}

// loadConfig 从YAML文件加载配置，支持环境变量覆盖敏感字段
func loadConfig(configFile string) (*config.Config, error) {
	// 首先创建默认配置
	cfg := config.NewDefaultConfig()

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 环境变量覆盖（优先级高于YAML文件）
	for _, ov := range envVarOverrides {
		if val, ok := os.LookupEnv(ov.envKey); ok && val != "" {
			ov.apply(cfg, val)
		}
	}

	return cfg, nil
}

// validateConfig 验证配置是否完整
func validateConfig(cfg *config.Config) error {
	if cfg.ZentaoURL == "" {
		return fmt.Errorf("禅道URL不能为空")
	}
	if cfg.ZentaoUsername == "" {
		return fmt.Errorf("禅道用户名不能为空")
	}
	if cfg.ZentaoPassword == "" {
		return fmt.Errorf("禅道密码不能为空")
	}
	if cfg.ExcelFile == "" {
		return fmt.Errorf("Excel文件路径不能为空")
	}
	if !filepath.IsAbs(cfg.ExcelFile) {
		absPath, err := filepath.Abs(cfg.ExcelFile)
		if err != nil {
			return fmt.Errorf("转换Excel文件路径为绝对路径失败: %w", err)
		}
		cfg.ExcelFile = absPath
	}
	return nil
}

