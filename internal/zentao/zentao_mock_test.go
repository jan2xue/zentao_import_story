package zentao

import (
	"github.com/imroc/req/v3"
)

// mockEpicService 实现 EpicCreator 接口
type mockEpicService struct {
	createFn func(req EpicCreateRequest) (*EpicCreateResponse, *req.Response, error)
	deleteFn func(id int) (map[string]interface{}, *req.Response, error)
	listFn   func(productID int) ([]EpicListItem, error)
}

func (m *mockEpicService) Create(req EpicCreateRequest) (*EpicCreateResponse, *req.Response, error) {
	return m.createFn(req)
}

func (m *mockEpicService) DeleteByID(id int) (map[string]interface{}, *req.Response, error) {
	return m.deleteFn(id)
}

func (m *mockEpicService) ProductsListAll(productID int) ([]EpicListItem, error) {
	return m.listFn(productID)
}

// mockReqService 实现 RequirementCreator 接口
type mockReqService struct {
	createFn func(req RequirementCreateRequest) (*RequirementCreateResponse, *req.Response, error)
	deleteFn func(id int) (map[string]interface{}, *req.Response, error)
	listFn   func(productID int) ([]RequirementListItem, error)
}

func (m *mockReqService) Create(req RequirementCreateRequest) (*RequirementCreateResponse, *req.Response, error) {
	return m.createFn(req)
}

func (m *mockReqService) DeleteByID(id int) (map[string]interface{}, *req.Response, error) {
	return m.deleteFn(id)
}

func (m *mockReqService) ProductsListAll(productID int) ([]RequirementListItem, error) {
	return m.listFn(productID)
}

// mockStoryService 实现 StoryCreator 接口
type mockStoryService struct {
	createFn func(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error)
	deleteFn func(id int) (map[string]interface{}, *req.Response, error)
	listFn   func(productID int) ([]StoryListItem, error)
}

func (m *mockStoryService) Create(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
	return m.createFn(req)
}

func (m *mockStoryService) DeleteByID(id int) (map[string]interface{}, *req.Response, error) {
	return m.deleteFn(id)
}

func (m *mockStoryService) ProductsListAll(productID int) ([]StoryListItem, error) {
	return m.listFn(productID)
}

// mockConfig 实现 ConfigProvider 接口
type mockConfig struct {
	module   int
	reviewer string
}

func (m *mockConfig) GetDefaultModule() int      { return m.module }
func (m *mockConfig) GetDefaultReviewer() string { return m.reviewer }
