// Package main 是应用程序入口
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
	action := flag.String("action", "import", "操作类型: import(导入) 或 export(导出)")
	productID := flag.Int("product", 0, "产品ID(导出时使用)")
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
		log.Info("执行导入操作")
		handleImport(cfg, log)
	case "export":
		log.Info("执行导出操作，产品ID: %d", *productID)
		if *productID == 0 {
			log.Fatal("导出操作需要指定产品ID (-product 参数)")
		}
		handleExport(cfg, log, *productID)
	default:
		log.Fatal("不支持的操作类型: %s，仅支持 import 或 export", *action)
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

// handleExport 处理导出操作
func handleExport(cfg *config.Config, log *logger.Logger, productID int) {
	// 创建禅道客户端
	client, err := zentao.NewClient(cfg)
	if err != nil {
		log.Fatal("创建禅道客户端失败: %v", err)
	}

	// 创建导出器
	exporter := zentao.NewExporter(client, log)

	// 从禅道服务器导出需求
	stories, result := exporter.ExportStories(productID)
	if !result.Success {
		log.Fatal("导出需求失败: %v", result.Error)
	}

	log.Info("从禅道服务器导出 %d 个需求，耗时: %dms", result.StoryCount, result.ElapsedTime)

	if len(stories) == 0 {
		log.Info("没有找到需要导出的需求")
		os.Exit(0)
	}

	// 创建Excel写入器
	writer, err := excel.NewWriter(cfg.ExcelFile)
	if err != nil {
		log.Fatal("创建Excel写入器失败: %v", err)
	}
	defer writer.Close()

	// 将需求写入Excel文件
	err = writer.WriteStories(stories)
	if err != nil {
		log.Fatal("写入Excel文件失败: %v", err)
	}

	// 保存Excel文件
	err = writer.Save(cfg.ExcelFile)
	if err != nil {
		log.Fatal("保存Excel文件失败: %v", err)
	}

	log.Info("成功导出 %d 个需求到Excel文件: %s", len(stories), cfg.ExcelFile)
	log.Info("日志文件已保存至: %s", log.GetLogFilePath())
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
