package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	// 解析命令行参数
	configFile := flag.String("config", "config.yaml", "配置文件路径")
	excelFile := flag.String("file", "", "Excel文件路径")
	flag.Parse()

	// 验证Excel文件参数
	if *excelFile == "" {
		fmt.Println("请使用 -file 参数指定Excel文件路径")
	}

	// 加载配置文件
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 如果命令行指定了Excel文件，覆盖配置文件中的设置
	if *excelFile != "" {
		config.ExcelFile = *excelFile
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		log.Fatal(err)
	}

	// 创建Excel读取器
	reader, err := NewExcelReader(config.ExcelFile)
	if err != nil {
		log.Fatalf("创建Excel读取器失败: %v", err)
	}
	defer reader.Close()

	// 读取需求数据
	stories, err := reader.ReadStories()
	if err != nil {
		log.Fatalf("读取Excel数据失败: %v", err)
	}

	if len(stories) == 0 {
		log.Fatal("Excel文件中没有找到需求数据")
	}

	// 创建导入器
	importer, err := NewImporter(config)
	if err != nil {
		log.Fatalf("创建导入器失败: %v", err)
	}

	// 导入需求
	fmt.Printf("开始导入 %d 个需求...\n", len(stories))
	results := importer.ImportStories(stories)

	// 生成并显示报告
	report := importer.GenerateReport(results)
	fmt.Println(report)

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
