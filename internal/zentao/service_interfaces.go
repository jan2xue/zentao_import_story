// Package zentao 封装禅道API客户端 — 服务接口定义
package zentao

import "github.com/imroc/req/v3"

// StoryCreator 研发需求创建/查询接口（用于Importer/Deleter依赖注入）
type StoryCreator interface {
	Create(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error)
	ProductsListAll(productID int) ([]StoryListItem, error)
	DeleteByID(id int) (map[string]interface{}, *req.Response, error)
}

// EpicCreator 业务需求创建/查询接口
type EpicCreator interface {
	Create(req EpicCreateRequest) (*EpicCreateResponse, *req.Response, error)
	ProductsListAll(productID int) ([]EpicListItem, error)
	DeleteByID(id int) (map[string]interface{}, *req.Response, error)
}

// RequirementCreator 用户需求创建/查询接口
type RequirementCreator interface {
	Create(req RequirementCreateRequest) (*RequirementCreateResponse, *req.Response, error)
	ProductsListAll(productID int) ([]RequirementListItem, error)
	DeleteByID(id int) (map[string]interface{}, *req.Response, error)
}

// ConfigProvider 配置访问接口（用于测试隔离）
type ConfigProvider interface {
	GetDefaultModule() int
	GetDefaultReviewer() string
}
