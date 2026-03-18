// Package zentao 封装禅道API客户端 - Epic业务需求服务
package zentao

import (
	"fmt"

	"github.com/imroc/req/v3"
)

// EpicCreateRequest 创建业务需求的请求体 (API V2.0)
type EpicCreateRequest struct {
	ProductID  int      `json:"productID"`            // 产品ID (必填)
	Title      string   `json:"title"`                // 标题 (必填)
	Pri        int      `json:"pri,omitempty"`        // 优先级，默认是3
	Module     int      `json:"module,omitempty"`     // 所属模块
	Parent     int      `json:"parent,omitempty"`     // 父业务需求
	Estimate   float64  `json:"estimate,omitempty"`   // 预计工时
	Spec       string   `json:"spec,omitempty"`       // 业务需求描述
	Category   string   `json:"category,omitempty"`   // 类别
	Source     string   `json:"source,omitempty"`     // 来源
	Verify     string   `json:"verify,omitempty"`     // 验收标准
	AssignedTo string   `json:"assignedTo,omitempty"` // 指派给
	Reviewer   []string `json:"reviewer,omitempty"`   // 评审人
}

// EpicCreateResponse 创建业务需求的响应
type EpicCreateResponse struct {
	Status string `json:"status"` // 状态(success 成功 | fail 失败)
	ID     int    `json:"id"`     // 创建的业务需求ID
}

// EpicService 处理业务需求的API操作
type EpicService struct {
	client *Client
}

// NewEpicService 创建新的Epic服务
func NewEpicService(client *Client) *EpicService {
	return &EpicService{client: client}
}

// Create 创建业务需求
// POST /api.php/v2/epics
func (s *EpicService) Create(req EpicCreateRequest) (*EpicCreateResponse, *req.Response, error) {
	var resp EpicCreateResponse
	rsp, err := s.client.R().
		SetBody(&req).
		SetSuccessResult(&resp).
		Post(s.client.RequestURL("/epics"))
	if err != nil {
		return nil, rsp, err
	}
	return &resp, rsp, nil
}

// GetByID 获取业务需求详情
// GET /api.php/v2/epics/{id}
func (s *EpicService) GetByID(id int) (map[string]interface{}, *req.Response, error) {
	var resp map[string]interface{}
	rsp, err := s.client.R().
		SetSuccessResult(&resp).
		Get(s.client.RequestURL(fmt.Sprintf("/epics/%d", id)))
	if err != nil {
		return nil, rsp, err
	}
	return resp, rsp, nil
}

// UpdateByID 修改业务需求
// PUT /api.php/v2/epics/{id}
func (s *EpicService) UpdateByID(id int, req EpicCreateRequest) (map[string]interface{}, *req.Response, error) {
	var resp map[string]interface{}
	rsp, err := s.client.R().
		SetBody(&req).
		SetSuccessResult(&resp).
		Put(s.client.RequestURL(fmt.Sprintf("/epics/%d", id)))
	if err != nil {
		return nil, rsp, err
	}
	return resp, rsp, nil
}

// DeleteByID 删除业务需求
// DELETE /api.php/v2/epics/{id}
func (s *EpicService) DeleteByID(id int) (map[string]interface{}, *req.Response, error) {
	var resp map[string]interface{}
	rsp, err := s.client.R().
		SetSuccessResult(&resp).
		Delete(s.client.RequestURL(fmt.Sprintf("/epics/%d", id)))
	if err != nil {
		return nil, rsp, err
	}
	return resp, rsp, nil
}

// ProductsList 获取产品业务需求列表
// GET /api.php/v2/products/{id}/epics
func (s *EpicService) ProductsList(productID int) (map[string]interface{}, *req.Response, error) {
	var resp map[string]interface{}
	rsp, err := s.client.R().
		SetSuccessResult(&resp).
		Get(s.client.RequestURL(fmt.Sprintf("/products/%d/epics", productID)))
	if err != nil {
		return nil, rsp, err
	}
	return resp, rsp, nil
}
