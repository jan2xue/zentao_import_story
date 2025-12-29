package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	var err error
	// 初始化日志记录器
	logger, err = NewLogger()
	if err != nil {
		// 由于日志记录器初始化失败，只能使用标准日志
		fmt.Printf("初始化日志记录器失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	excelPath := flag.String("excel", "requirements.xlsx", "Excel文件路径")
	flag.Parse()

	// 加载配置文件
	config, err := loadConfig(*configPath)
	if err != nil {
		logger.Fatal("加载配置文件失败: %v", err)
	}

	// 如果命令行指定了Excel文件，覆盖配置文件中的设置
	if *excelPath != "" {
		config.ExcelFile = *excelPath
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		logger.Fatal("%v", err)
	}

	// 创建Excel读取器
	reader, err := NewExcelReader(config.ExcelFile)
	if err != nil {
		logger.Fatal("创建Excel读取器失败: %v", err)
	}
	defer reader.Close()

	// 读取需求数据
	stories, err := reader.ReadStories(config.DefaultPriority)
	if err != nil {
		logger.Fatal("读取Excel数据失败: %v", err)
	}

	logger.Info("从Excel中读取到 %d 个需求", len(stories))

	// 创建导入器
	importer, err := NewImporter(config)
	if err != nil {
		logger.Fatal("创建导入器失败: %v", err)
	}

	// 导入需求
	results := importer.ImportStories(stories)

	// 生成并打印报告
	report := importer.GenerateReport(results)
	logger.Info("\n%s", report)

	logger.Info("日志文件已保存至: %s", logger.GetLogFilePath())

	// 如果有失败的导入，使用非零状态码退出
	for _, result := range results {
		if !result.Success {
			os.Exit(1)
		}
	}
}

// loadConfig 从YAML文件加载配置
func loadConfig(configFile string) (*Config, error) {
	// 首先创建默认配置
	config := NewDefaultConfig()

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return config, nil
}

// validateConfig 验证配置是否完整
func validateConfig(config *Config) error {
	if config.ZentaoURL == "" {
		return fmt.Errorf("禅道URL不能为空")
	}
	if config.ZentaoUsername == "" {
		return fmt.Errorf("禅道用户名不能为空")
	}
	if config.ZentaoPassword == "" {
		return fmt.Errorf("禅道密码不能为空")
	}
	if config.ExcelFile == "" {
		return fmt.Errorf("Excel文件路径不能为空")
	}
	if !filepath.IsAbs(config.ExcelFile) {
		absPath, err := filepath.Abs(config.ExcelFile)
		if err != nil {
			return fmt.Errorf("转换Excel文件路径为绝对路径失败: %w", err)
		}
		config.ExcelFile = absPath
	}
	return nil
}
