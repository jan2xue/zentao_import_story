// Package zentao 封装禅道API客户端 - Product产品服务
package zentao

import (
	"fmt"
)

// Product 产品信息
type Product struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	Code        string      `json:"code"`
	Type        string      `json:"type"`
	Status      string      `json:"status"`
	Description interface{} `json:"desc"`
}

// ProductDetailResponse 产品详情响应
type ProductDetailResponse struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

// ProductService 产品服务
type ProductService struct {
	client *Client
}

// NewProductService 创建新的产品服务
func NewProductService(client *Client) *ProductService {
	return &ProductService{client: client}
}

// GetByID 获取产品详情
// GET /api.php/v2/products/{id}
func (s *ProductService) GetByID(id int) (*Product, error) {
	var resp ProductDetailResponse
	rsp, err := s.client.R().
		SetSuccessResult(&resp).
		Get(s.client.RequestURL(fmt.Sprintf("/products/%d", id)))
	if err != nil {
		return nil, fmt.Errorf("获取产品信息失败: %w", err)
	}
	
	if rsp.StatusCode >= 400 {
		return nil, fmt.Errorf("获取产品信息失败，HTTP状态码: %d", rsp.StatusCode)
	}
	
	if resp.Status != "success" {
		return nil, fmt.Errorf("获取产品信息失败: status=%s", resp.Status)
	}
	
	// 从 map 中提取产品信息
	product := &Product{ID: id}
	if resp.Data != nil {
		if name, ok := resp.Data["name"]; ok {
			if nameStr, ok := name.(string); ok {
				product.Name = nameStr
			}
		}
		if code, ok := resp.Data["code"]; ok {
			if codeStr, ok := code.(string); ok {
				product.Code = codeStr
			}
		}
	}
	
	return product, nil
}

// GetProductInfo 批量获取多个产品的ID和名称映射
func (s *ProductService) GetProductInfo(productIDs []int) (map[int]string, error) {
	result := make(map[int]string)
	
	for _, id := range productIDs {
		product, err := s.GetByID(id)
		if err != nil {
			result[id] = "[产品不存在或无权限]"
		} else if product.Name == "" {
			result[id] = "[产品名称为空]"
		} else {
			result[id] = product.Name
		}
	}
	
	return result, nil
}
