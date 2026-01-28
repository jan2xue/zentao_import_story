// Package zentao 封装禅道API客户端
package zentao

import (
	"fmt"

	"github.com/easysoft/go-zentao/v21/zentao"
	"github.com/jan2xue/zentao_import_story/internal/config"
)

// Client 封装禅道客户端
type Client struct {
	client *zentao.Client
	config *config.Config
}

// NewClient 创建新的禅道客户端
func NewClient(cfg *config.Config) (*Client, error) {
	client, err := zentao.NewBasicAuthClient(
		cfg.ZentaoUsername,
		cfg.ZentaoPassword,
		zentao.WithBaseURL(cfg.ZentaoURL),
		zentao.WithDevMode(),
		zentao.WithDumpAll(),
		zentao.WithoutProxy(),
	)
	if err != nil {
		return nil, fmt.Errorf("创建禅道客户端失败: %w", err)
	}
	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// GetClient 获取底层禅道客户端
func (c *Client) GetClient() *zentao.Client {
	return c.client
}
