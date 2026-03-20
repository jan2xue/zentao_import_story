// Package zentao 封装禅道API客户端 - Requirement用户需求服务
package zentao

import (
	"fmt"

	"github.com/imroc/req/v3"
)

// RequirementCreateRequest 创建用户需求的请求体 (API V2.0)
type RequirementCreateRequest struct {
	ProductID  int      `json:"productID"`            // 产品ID (必填)
	Title      string   `json:"title"`                // 标题 (必填)
	Pri        int      `json:"pri,omitempty"`        // 优先级，默认是3
	Module     int      `json:"module"`               // 所属模块
	Parent     int      `json:"parent,omitempty"`     // 父用户需求
	Estimate   float64  `json:"estimate,omitempty"`   // 预计工时
	Spec       string   `json:"spec,omitempty"`       // 用户需求描述
	Category   string   `json:"category,omitempty"`   // 类别
	Source     string   `json:"source,omitempty"`     // 来源
	Verify     string   `json:"verify,omitempty"`     // 验收标准
	AssignedTo string   `json:"assignedTo,omitempty"` // 指派给
	Reviewer   []string `json:"reviewer,omitempty"`   // 评审人
}

// RequirementCreateResponse 创建用户需求的响应
type RequirementCreateResponse struct {
	Status string `json:"status"` // 状态(success 成功 | fail 失败)
	ID     int    `json:"id"`     // 创建的用户需求ID
}

// RequirementService 处理用户需求的API操作
type RequirementService struct {
	client *Client
}

// NewRequirementService 创建新的Requirement服务
func NewRequirementService(client *Client) *RequirementService {
	return &RequirementService{client: client}
}

// Create 创建用户需求
// POST /api.php/v2/requirements
func (s *RequirementService) Create(req RequirementCreateRequest) (*RequirementCreateResponse, *req.Response, error) {
	var resp RequirementCreateResponse
	rsp, err := s.client.R().
		SetBody(&req).
		SetSuccessResult(&resp).
		Post(s.client.RequestURL("/requirements"))
	if err != nil {
		return nil, rsp, err
	}
	return &resp, rsp, nil
}

// GetByID 获取用户需求详情
// GET /api.php/v2/requirements/{id}
func (s *RequirementService) GetByID(id int) (map[string]interface{}, *req.Response, error) {
	var resp map[string]interface{}
	rsp, err := s.client.R().
		SetSuccessResult(&resp).
		Get(s.client.RequestURL(fmt.Sprintf("/requirements/%d", id)))
	if err != nil {
		return nil, rsp, err
	}
	return resp, rsp, nil
}

// UpdateByID 修改用户需求
// PUT /api.php/v2/requirements/{id}
func (s *RequirementService) UpdateByID(id int, req RequirementCreateRequest) (map[string]interface{}, *req.Response, error) {
	var resp map[string]interface{}
	rsp, err := s.client.R().
		SetBody(&req).
		SetSuccessResult(&resp).
		Put(s.client.RequestURL(fmt.Sprintf("/requirements/%d", id)))
	if err != nil {
		return nil, rsp, err
	}
	return resp, rsp, nil
}

// DeleteByID 删除用户需求
// DELETE /api.php/v2/requirements/{id}
func (s *RequirementService) DeleteByID(id int) (map[string]interface{}, *req.Response, error) {
	var resp map[string]interface{}
	rsp, err := s.client.R().
		SetSuccessResult(&resp).
		Delete(s.client.RequestURL(fmt.Sprintf("/requirements/%d", id)))
	if err != nil {
		return nil, rsp, err
	}
	return resp, rsp, nil
}

// RequirementListItem 用户需求列表项
type RequirementListItem struct {
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
}

// RequirementListWithPagerResponse 带分页信息的用户需求列表响应
type RequirementListWithPagerResponse struct {
	Status       string                `json:"status"`
	Requirements []RequirementListItem `json:"requirements"`
	Pager        Pager                 `json:"pager"`
}

// ProductsList 获取产品用户需求列表（单页）
// GET /api.php/v2/products/{id}/requirements
func (s *RequirementService) ProductsList(productID int, opts *ListOptions) (*RequirementListWithPagerResponse, *req.Response, error) {
	var resp RequirementListWithPagerResponse
	req := s.client.R().SetSuccessResult(&resp)
	
	if opts != nil {
		if opts.RecPerPage > 0 {
			req.SetQueryParam("recPerPage", fmt.Sprintf("%d", opts.RecPerPage))
		}
		if opts.PageID > 0 {
			req.SetQueryParam("pageID", fmt.Sprintf("%d", opts.PageID))
		}
	}
	
	rsp, err := req.Get(s.client.RequestURL(fmt.Sprintf("/products/%d/requirements", productID)))
	if err != nil {
		return nil, rsp, err
	}
	return &resp, rsp, nil
}

// ProductsListAll 获取产品所有用户需求（自动分页）
// GET /api.php/v2/products/{id}/requirements
func (s *RequirementService) ProductsListAll(productID int) ([]RequirementListItem, error) {
	var allItems []RequirementListItem
	pageID := 1
	pageSize := 100
	
	for {
		resp, _, err := s.ProductsList(productID, &ListOptions{
			PageID:     pageID,
			RecPerPage: pageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("获取用户需求列表失败(页%d): %w", pageID, err)
		}
		
		allItems = append(allItems, resp.Requirements...)
		
		if resp.Pager.PageTotal == 0 || pageID >= resp.Pager.PageTotal {
			break
		}
		pageID++
	}
	
	return allItems, nil
}
