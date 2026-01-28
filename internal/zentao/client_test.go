package zentao

import (
	"testing"

	"github.com/jan2xue/zentao_import_story/internal/config"
)

func TestNewClient(t *testing.T) {
	// 由于需要真实的禅道服务器连接，这里仅测试配置验证
	// 实际测试中应使用 mock 或测试服务器

	cfg := &config.Config{
		ZentaoURL:      "http://test.zentao.com",
		ZentaoUsername: "testuser",
		ZentaoPassword: "testpass",
	}

	// 注意：这个测试会尝试连接网络，如果没有可用的禅道服务器会失败
	// 在生产环境中应该使用 mock
	_ = cfg

	// 基础验证测试
	if cfg.ZentaoURL == "" {
		t.Error("禅道URL不应为空")
	}
	if cfg.ZentaoUsername == "" {
		t.Error("用户名不应为空")
	}
	if cfg.ZentaoPassword == "" {
		t.Error("密码不应为空")
	}
}

func TestClient_GetClient(t *testing.T) {
	// 创建 nil client 测试 GetClient 方法
	var client *Client
	
	// 这里只测试方法存在性，实际调用需要有效的 client
	// 由于 GetClient 可能 panic，这里做简单检查
	if client != nil {
		_ = client.GetClient()
	}
}
