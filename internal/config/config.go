// Package config 处理应用程序配置
package config

// Config 存储程序配置信息
type Config struct {
	// 禅道系统配置
	ZentaoURL      string `yaml:"zentaoUrl"`
	ZentaoUsername string `yaml:"zentaoUsername"`
	ZentaoPassword string `yaml:"zentaoPassword"`

	// Excel文件配置
	ExcelFile string `yaml:"excelFile"`

	// 默认值配置
	DefaultPriority int `yaml:"defaultPriority"` // 默认优先级 1-4
}

// NewDefaultConfig 返回默认配置
func NewDefaultConfig() *Config {
	return &Config{
		DefaultPriority: 3, // 默认优先级为3
	}
}
