package config

import (
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	if cfg.DefaultPriority != 3 {
		t.Errorf("默认优先级应该是3，但得到 %d", cfg.DefaultPriority)
	}
}

func TestConfig_Values(t *testing.T) {
	cfg := &Config{
		ZentaoURL:       "http://test.zentao.com",
		ZentaoUsername:  "admin",
		ZentaoPassword:  "password",
		ExcelFile:       "test.xlsx",
		DefaultPriority: 2,
	}

	if cfg.ZentaoURL != "http://test.zentao.com" {
		t.Errorf("禅道URL不匹配")
	}
	if cfg.DefaultPriority != 2 {
		t.Errorf("优先级不匹配")
	}
}
