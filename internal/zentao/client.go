// Package zentao 封装禅道API客户端
package zentao

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/jan2xue/zentao_import_story/internal/config"
)

const apiVersionPath = "/api.php/v2"

// loginRequest v2.0 登录请求参数
type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

// loginResponse v2.0 登录响应
type loginResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

// Client 封装禅道客户端
type Client struct {
	httpClient  *req.Client
	config      *config.Config
	token       string
	baseURL     *url.URL
	Epic        *EpicService
	Requirement *RequirementService
	Story       *StoryService
	Product     *ProductService
}

// NewClient 创建新的禅道客户端
func NewClient(cfg *config.Config) (*Client, error) {
	// 解析baseURL
	baseURLStr := cfg.ZentaoURL
	if strings.HasSuffix(baseURLStr, "/") {
		baseURLStr = strings.TrimSuffix(baseURLStr, "/")
	}
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return nil, fmt.Errorf("解析禅道URL失败: %w", err)
	}
	if !strings.HasSuffix(baseURL.Path, apiVersionPath) {
		baseURL.Path += apiVersionPath
	}

	// 创建HTTP客户端（不绑定固定token，改用动态注入）
	httpClient := req.C().SetLogger(nil)

	c := &Client{
		httpClient: httpClient,
		config:     cfg,
		baseURL:    baseURL,
	}

	// 注册OnBeforeRequest中间件：每次请求动态注入当前token
	c.httpClient.OnBeforeRequest(func(client *req.Client, req *req.Request) error {
		req.SetHeader("Token", c.token)
		return nil
	})

	// 注册401自动重试：token过期时刷新并重试一次
	c.httpClient.
		SetCommonRetryCount(1).
		SetCommonRetryCondition(func(resp *req.Response, err error) bool {
			return err == nil && resp != nil && resp.StatusCode == 401
		}).
		SetCommonRetryHook(func(resp *req.Response, err error) {
			token, loginErr := c.login()
			if loginErr == nil {
				c.token = token
			}
		})

	// 使用 v2.0 API 获取 token
	token, err := c.login()
	if err != nil {
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}
	c.token = token

	// 初始化服务
	c.Epic = NewEpicService(c)
	c.Requirement = NewRequirementService(c)
	c.Story = NewStoryService(c)
	c.Product = NewProductService(c)

	return c, nil
}

// login 使用 v2.0 API 登录获取 token
func (c *Client) login() (string, error) {
	loginURL := c.baseURL.String() + "/users/login"

	var resp loginResponse
	_, err := c.httpClient.R().
		SetBody(loginRequest{
			Account:  c.config.ZentaoUsername,
			Password: c.config.ZentaoPassword,
		}).
		SetSuccessResult(&resp).
		Post(loginURL)

	if err != nil {
		return "", fmt.Errorf("登录请求失败: %w", err)
	}

	if resp.Status != "success" {
		return "", fmt.Errorf("登录失败: status=%s", resp.Status)
	}

	if resp.Token == "" {
		return "", fmt.Errorf("登录成功但未返回 token")
	}

	return resp.Token, nil
}

// GetToken 获取访问令牌
func (c *Client) GetToken() string {
	return c.token
}

// RequestURL 构建API请求URL
func (c *Client) RequestURL(path string) string {
	u := *c.baseURL
	u.Path = c.baseURL.Path + path
	return u.String()
}

// R 获取HTTP请求构建器
// 注意：Token已由OnBeforeRequest中间件自动注入，无需在此设置
func (c *Client) R() *req.Request {
	return c.httpClient.R()
}
