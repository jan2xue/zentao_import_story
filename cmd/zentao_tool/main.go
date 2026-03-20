// Package main 是应用程序入口
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/jan2xue/zentao_import_story/internal/config"
	"github.com/jan2xue/zentao_import_story/internal/excel"
	"github.com/jan2xue/zentao_import_story/internal/logger"
	"github.com/jan2xue/zentao_import_story/internal/zentao"
)

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
	action := flag.String("action", "import", "操作类型: import(导入) 或 delete(删除)")
	productID := flag.Int("product", 0, "产品ID(导出时使用)")
	storyType := flag.String("type", "story", "需求类型: epic(业务需求)/requirement(用户需求)/story(研发需求)")
	storyIDs := flag.String("ids", "", "要删除的需求ID列表，多个ID用逗号分隔(删除时使用)")
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

	// 验证需求类型
	if *storyType != "epic" && *storyType != "requirement" && *storyType != "story" {
		log.Fatal("无效的需求类型: %s，仅支持 epic/requirement/story", *storyType)
	}

	// 根据操作类型执行相应功能
	switch *action {
	case "import":
		log.Info("执行导入操作，需求类型: %s", *storyType)
		handleImport(cfg, log, *storyType)
	case "delete":
		log.Info("执行删除操作，需求类型: %s", *storyType)
		handleDelete(cfg, log, *storyType, *storyIDs, *productID)
	default:
		log.Fatal("不支持的操作类型: %s，仅支持 import 或 delete", *action)
	}
}

// handleImport 处理导入操作
func handleImport(cfg *config.Config, log *logger.Logger, storyType string) {
	// 创建Excel读取器
	reader, err := excel.NewReader(cfg.ExcelFile)
	if err != nil {
		log.Fatal("创建Excel读取器失败: %v", err)
	}
	defer reader.Close()

	// 读取需求数据
	stories, err := reader.ReadStories(cfg.DefaultPriority, storyType)
	if err != nil {
		log.Fatal("读取Excel数据失败: %v", err)
	}

	log.Info("从Excel中读取到 %d 个需求", len(stories))

	// 创建禅道客户端
	client, err := zentao.NewClient(cfg)
	if err != nil {
		log.Fatal("创建禅道客户端失败: %v", err)
	}

	// 创建导入器
	importer := zentao.NewImporter(client, log)

	// 导入需求
	results := importer.ImportStories(stories)

	// 生成并打印报告
	report := importer.GenerateReport(results)
	log.Info("\n%s", report)

	log.Info("日志文件已保存至: %s", log.GetLogFilePath())

	// 如果有失败的导入，使用非零状态码退出
	for _, result := range results {
		if !result.Success {
			os.Exit(1)
		}
	}
}

// handleDelete 处理删除操作
func handleDelete(cfg *config.Config, log *logger.Logger, storyType string, storyIDsStr string, productID int) {
	var storyIDs []int

	// 解析需求ID列表
	if storyIDsStr != "" {
		// 从命令行参数解析ID列表
		idStrs := strings.Split(storyIDsStr, ",")
		for _, idStr := range idStrs {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}
			id, err := strconv.Atoi(idStr)
			if err != nil {
				log.Fatal("无效的需求ID: %s", idStr)
			}
			storyIDs = append(storyIDs, id)
		}
	} else if productID > 0 {
		// 从产品ID获取所有需求ID
		storyIDs = fetchStoryIDsFromProduct(cfg, log, productID, storyType)
		if len(storyIDs) == 0 {
			log.Info("产品 %d 下没有找到任何需求", productID)
			return
		}
	} else {
		log.Fatal("删除操作需要指定需求ID列表 (-ids 参数) 或产品ID (-product 参数)")
	}

	// 显示确认提示
	fmt.Printf("\n⚠️  警告: 即将删除 %d 个需求\n", len(storyIDs))
	fmt.Printf("需求类型: %s\n", storyType)
	fmt.Printf("需求ID列表: %v\n", storyIDs)
	fmt.Printf("\n此操作不可撤销，是否继续? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("读取用户输入失败: %v", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "yes" && response != "y" {
		log.Info("取消删除操作")
		return
	}

	// 创建禅道客户端
	client, err := zentao.NewClient(cfg)
	if err != nil {
		log.Fatal("创建禅道客户端失败: %v", err)
	}

	// 创建删除器
	deleter := zentao.NewDeleter(client, log)

	// 执行删除
	results := deleter.DeleteStories(storyIDs, storyType)

	// 生成并打印报告
	report := deleter.GenerateDeleteReport(results)
	log.Info("\n%s", report)

	log.Info("日志文件已保存至: %s", log.GetLogFilePath())

	// 如果有失败的删除，使用非零状态码退出
	for _, result := range results {
		if !result.Success {
			os.Exit(1)
		}
	}
}

// fetchStoryIDsFromProduct 从产品获取所有需求ID
func fetchStoryIDsFromProduct(cfg *config.Config, log *logger.Logger, productID int, storyType string) []int {
	// 创建禅道客户端
	client, err := zentao.NewClient(cfg)
	if err != nil {
		log.Fatal("创建禅道客户端失败: %v", err)
	}

	var storyIDs []int

	switch storyType {
	case "story":
		stories, err := client.Story.ProductsListAll(productID)
		if err != nil {
			log.Fatal("获取产品需求列表失败: %v", err)
		}
		log.Info("产品 %d 下共有 %d 个研发需求", productID, len(stories))
		for _, s := range stories {
			storyIDs = append(storyIDs, s.ID)
		}
	case "requirement":
		requirements, err := client.Requirement.ProductsListAll(productID)
		if err != nil {
			log.Fatal("获取产品用户需求列表失败: %v", err)
		}
		log.Info("产品 %d 下共有 %d 个用户需求", productID, len(requirements))
		for _, r := range requirements {
			storyIDs = append(storyIDs, r.ID)
		}
	case "epic":
		epics, err := client.Epic.ProductsListAll(productID)
		if err != nil {
			log.Fatal("获取产品业务需求列表失败: %v", err)
		}
		log.Info("产品 %d 下共有 %d 个业务需求", productID, len(epics))
		for _, e := range epics {
			storyIDs = append(storyIDs, e.ID)
		}
	}

	return storyIDs
}

// loadConfig 从YAML文件加载配置
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