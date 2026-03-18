// Package zentao 封装禅道API客户端 - Story研发需求服务 (API V2.0)
package zentao

import (
	"fmt"

	"github.com/imroc/req/v3"
)

// StoryCreateRequest 创建研发需求的请求体 (API V2.0)
type StoryCreateRequest struct {
	ProductID  int      `json:"productID"`            // 产品ID (必填)
	Title      string   `json:"title"`                // 标题 (必填)
	Pri        int      `json:"pri,omitempty"`        // 优先级，默认是3
	Module     int      `json:"module,omitempty"`     // 所属模块
	Parent     int      `json:"parent,omitempty"`     // 父需求
	Estimate   float64  `json:"estimate,omitempty"`   // 预计工时
	Spec       string   `json:"spec,omitempty"`       // 需求描述
	Category   string   `json:"category,omitempty"`   // 类别
	Source     string   `json:"source,omitempty"`     // 来源
	Verify     string   `json:"verify,omitempty"`     // 验收标准
	AssignedTo string   `json:"assignedTo,omitempty"` // 指派给
	Reviewer   []string `json:"reviewer,omitempty"`   // 评审人
	Project    int      `json:"project,omitempty"`    // 所属项目
	Execution  int      `json:"execution,omitempty"`  // 所属执行
}

// StoryCreateResponse 创建研发需求的响应
type StoryCreateResponse struct {
	Status string `json:"status"` // 状态(success 成功 | fail 失败)
	ID     int    `json:"id"`     // 创建的需求ID
}

// StoryService 处理研发需求的API操作
type StoryService struct {
	client *Client
}

// NewStoryService 创建新的Story服务
func NewStoryService(client *Client) *StoryService {
	return &StoryService{client: client}
}

// Create 创建研发需求
// POST /api.php/v2/stories
func (s *StoryService) Create(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
	var resp StoryCreateResponse
	rsp, err := s.client.R().
		SetBody(&req).
		SetSuccessResult(&resp).
		Post(s.client.RequestURL("/stories"))
	if err != nil {
		return nil, rsp, err
	}
	return &resp, rsp, nil
}

// GetByID 获取研发需求详情
// GET /api.php/v2/stories/{id}
func (s *StoryService) GetByID(id int) (map[string]interface{}, *req.Response, error) {
	var resp map[string]interface{}
	rsp, err := s.client.R().
		SetSuccessResult(&resp).
		Get(s.client.RequestURL(fmt.Sprintf("/stories/%d", id)))
	if err != nil {
		return nil, rsp, err
	}
	return resp, rsp, nil
}

// UpdateByID 修改研发需求
// PUT /api.php/v2/stories/{id}
func (s *StoryService) UpdateByID(id int, req StoryCreateRequest) (map[string]interface{}, *req.Response, error) {
	var resp map[string]interface{}
	rsp, err := s.client.R().
		SetBody(&req).
		SetSuccessResult(&resp).
		Put(s.client.RequestURL(fmt.Sprintf("/stories/%d", id)))
	if err != nil {
		return nil, rsp, err
	}
	return resp, rsp, nil
}

// DeleteByID 删除研发需求
// DELETE /api.php/v2/stories/{id}
func (s *StoryService) DeleteByID(id int) (map[string]interface{}, *req.Response, error) {
	var resp map[string]interface{}
	rsp, err := s.client.R().
		SetSuccessResult(&resp).
		Delete(s.client.RequestURL(fmt.Sprintf("/stories/%d", id)))
	if err != nil {
		return nil, rsp, err
	}
	return resp, rsp, nil
}

// StoryListItem 需求列表项
type StoryListItem struct {
	ID          int         `json:"id"`
	Parent      interface{} `json:"parent"`
	Product     int         `json:"product"`
	Branch      int         `json:"branch"`
	Module      int         `json:"module"`
	Plan        interface{} `json:"plan"`
	Source      string      `json:"source"`
	SourceNote  string      `json:"sourceNote"`
	Title       string      `json:"title"`
	Keywords    string      `json:"keywords"`
	Type        string      `json:"type"`
	Category    string      `json:"category"`
	Pri         int         `json:"pri"`
	Estimate    interface{} `json:"estimate"`
	Status      string      `json:"status"`
	Stage       string      `json:"stage"`
	OpenedBy    string      `json:"openedBy"`
	OpenedDate  string      `json:"openedDate"`
	AssignedTo  string      `json:"assignedTo"`
	Spec        string      `json:"spec"`
	Verify      string      `json:"verify"`
	Version     int         `json:"version"`
}

// StoryListResponse 需求列表响应
type StoryListResponse struct {
	Status string         `json:"status"`
	Stories []StoryListItem `json:"stories"`
	Total  int            `json:"total,omitempty"`
	Limit  int            `json:"limit,omitempty"`
}

// StoryDetailResponse 需求详情响应
type StoryDetailResponse struct {
	Status string        `json:"status"`
	Story  StoryListItem `json:"story"`
}

// ProductsList 获取产品研发需求列表
// GET /api.php/v2/products/{id}/stories
func (s *StoryService) ProductsList(productID int) (*StoryListResponse, *req.Response, error) {
	var resp StoryListResponse
	rsp, err := s.client.R().
		SetSuccessResult(&resp).
		Get(s.client.RequestURL(fmt.Sprintf("/products/%d/stories", productID)))
	if err != nil {
		return nil, rsp, err
	}
	return &resp, rsp, nil
}

// GetStoryDetail 获取需求详情（返回结构化数据）
// GET /api.php/v2/stories/{id}
func (s *StoryService) GetStoryDetail(id int) (*StoryDetailResponse, *req.Response, error) {
	var resp StoryDetailResponse
	rsp, err := s.client.R().
		SetSuccessResult(&resp).
		Get(s.client.RequestURL(fmt.Sprintf("/stories/%d", id)))
	if err != nil {
		return nil, rsp, err
	}
	return &resp, rsp, nil
}

// ProjectsList 获取项目需求列表
// GET /api.php/v2/projects/{id}/stories
func (s *StoryService) ProjectsList(projectID int) (map[string]interface{}, *req.Response, error) {
	var resp map[string]interface{}
	rsp, err := s.client.R().
		SetSuccessResult(&resp).
		Get(s.client.RequestURL(fmt.Sprintf("/projects/%d/stories", projectID)))
	if err != nil {
		return nil, rsp, err
	}
	return resp, rsp, nil
}

// ExecutionsList 获取执行需求列表
// GET /api.php/v2/executions/{id}/stories
func (s *StoryService) ExecutionsList(executionID int) (map[string]interface{}, *req.Response, error) {
	var resp map[string]interface{}
	rsp, err := s.client.R().
		SetSuccessResult(&resp).
		Get(s.client.RequestURL(fmt.Sprintf("/executions/%d/stories", executionID)))
	if err != nil {
		return nil, rsp, err
	}
	return resp, rsp, nil
}
