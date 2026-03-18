// Package zentao 封装禅道API客户端
package zentao

import (
	"fmt"
	"time"

	"github.com/imroc/req/v3"
	"github.com/jan2xue/zentao_import_story/internal/logger"
	"github.com/jan2xue/zentao_import_story/pkg/story"
)

// ImportResult 表示导入结果
type ImportResult struct {
	Success     bool
	StoryID     int
	StoryType   string
	Error       error
	ElapsedTime time.Duration
	RequestData string // 请求数据（用于调试）
	ResponseMsg string // 响应消息
}

// Importer 处理需求导入到禅道
type Importer struct {
	client *Client
	logger *logger.Logger
}

// NewImporter 创建新的导入器
func NewImporter(client *Client, log *logger.Logger) *Importer {
	return &Importer{
		client: client,
		logger: log,
	}
}

// ImportStory 导入单个需求
func (i *Importer) ImportStory(s *story.Story) ImportResult {
	start := time.Now()
	result := ImportResult{
		StoryType: string(s.Type),
	}

	i.logger.Info("正在导入%s: %s", s.GetTypeString(), s.Title)

	var createdID int
	var err error
	var rsp *req.Response

	// 根据需求类型选择不同的API创建
	switch s.Type {
	case story.StoryTypeEpic:
		createdID, rsp, err = i.createEpic(s)
	case story.StoryTypeRequirement:
		createdID, rsp, err = i.createRequirement(s)
	case story.StoryTypeStory:
		createdID, rsp, err = i.createStory(s)
	default:
		createdID, rsp, err = i.createStory(s)
	}

	if err != nil {
		errMsg := fmt.Sprintf("创建%s失败: %v", s.GetTypeString(), err)
		i.logger.ErrorWithDetail("需求导入失败", err, map[string]interface{}{
			"需求类型":    s.GetTypeString(),
			"需求标题":    s.Title,
			"产品ID":    s.ProductID,
			"HTTP状态码": i.getStatusCode(rsp),
			"响应内容":    i.getResponseBody(rsp),
		})
		result.Error = fmt.Errorf("%s", errMsg)
		result.Success = false
		result.ResponseMsg = i.getResponseBody(rsp)
	} else {
		i.logger.Success("%s创建成功，ID: %d", s.GetTypeString(), createdID)
		result.StoryID = createdID
		result.Success = true
	}

	result.ElapsedTime = time.Since(start)
	return result
}

// getStatusCode 安全获取HTTP状态码
func (i *Importer) getStatusCode(rsp *req.Response) int {
	if rsp == nil {
		return 0
	}
	return rsp.StatusCode
}

// getResponseBody 安全获取响应体
func (i *Importer) getResponseBody(rsp *req.Response) string {
	if rsp == nil {
		return ""
	}
	return rsp.String()
}

// createEpic 创建业务需求
func (i *Importer) createEpic(s *story.Story) (int, *req.Response, error) {
	req := EpicCreateRequest{
		ProductID: s.ProductID,
		Title:     s.Title,
		Pri:       s.Priority,
		Spec:      s.Spec,
		Category:  s.Category,
		Parent:    s.ParentID,
		Source:    s.Source,
		Verify:    s.Verify,
		Estimate:  s.Estimate,
	}

	i.logger.Debug("创建业务请求 - 产品ID: %d, 标题: %s", s.ProductID, s.Title)

	resp, rsp, err := i.client.Epic.Create(req)
	if err != nil {
		return 0, rsp, i.wrapAPIError("创建业务需求", err, rsp)
	}

	if resp.Status != "success" {
		return 0, rsp, fmt.Errorf("API返回失败状态: %s", resp.Status)
	}

	return resp.ID, rsp, nil
}

// createRequirement 创建用户需求
func (i *Importer) createRequirement(s *story.Story) (int, *req.Response, error) {
	req := RequirementCreateRequest{
		ProductID: s.ProductID,
		Title:     s.Title,
		Pri:       s.Priority,
		Spec:      s.Spec,
		Category:  s.Category,
		Parent:    s.ParentID,
		Source:    s.Source,
		Verify:    s.Verify,
		Estimate:  s.Estimate,
	}

	// 设置模块ID（用户需求必须指定有效的模块ID）
	if i.client.config.DefaultModule >= 0 {
		req.Module = i.client.config.DefaultModule
	} else {
		return 0, nil, fmt.Errorf("创建用户需求需要有效的模块ID，请在config.yaml中配置defaultModule参数")
	}

	// 设置评审人（如果配置了默认评审人）
	if i.client.config.DefaultReviewer != "" {
		req.Reviewer = []string{i.client.config.DefaultReviewer}
	}

	i.logger.Debug("创建用户需求请求 - 产品ID: %d, 标题: %s, 模块ID: %d", s.ProductID, s.Title, req.Module)

	resp, rsp, err := i.client.Requirement.Create(req)
	if err != nil {
		return 0, rsp, i.wrapAPIError("创建用户需求", err, rsp)
	}

	if resp.Status != "success" {
		return 0, rsp, fmt.Errorf("API返回失败状态: %s", resp.Status)
	}

	return resp.ID, rsp, nil
}

// createStory 创建研发需求
func (i *Importer) createStory(s *story.Story) (int, *req.Response, error) {
	req := StoryCreateRequest{
		ProductID: s.ProductID,
		Title:     s.Title,
		Pri:       s.Priority,
		Spec:      s.Spec,
		Category:  s.Category,
		Parent:    s.ParentID,
		Source:    s.Source,
		Verify:    s.Verify,
		Estimate:  s.Estimate,
	}

	// 设置评审人（如果配置了默认评审人）
	if i.client.config.DefaultReviewer != "" {
		req.Reviewer = []string{i.client.config.DefaultReviewer}
	}

	i.logger.Debug("创建研发需求请求 - 产品ID: %d, 标题: %s", s.ProductID, s.Title)

	resp, rsp, err := i.client.Story.Create(req)
	if err != nil {
		return 0, rsp, i.wrapAPIError("创建研发需求", err, rsp)
	}

	if resp.Status != "success" {
		return 0, rsp, fmt.Errorf("API返回失败状态: %s", resp.Status)
	}

	return resp.ID, rsp, nil
}

// wrapAPIError 包装API错误，添加详细信息
func (i *Importer) wrapAPIError(operation string, err error, rsp *req.Response) error {
	if rsp == nil {
		return fmt.Errorf("%s失败: %w (无HTTP响应)", operation, err)
	}
	return fmt.Errorf("%s失败: %w (HTTP %d: %s)", operation, err, rsp.StatusCode, rsp.String())
}

// ImportStories 批量导入需求
func (i *Importer) ImportStories(stories []story.Story) []ImportResult {
	results := make([]ImportResult, len(stories))

	i.logger.Info("开始批量导入需求，共 %d 个需求", len(stories))

	for idx, s := range stories {
		i.logger.Info("正在导入第 %d/%d 个需求", idx+1, len(stories))
		results[idx] = i.ImportStory(&s)
	}

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}
	i.logger.Info("需求导入完成，成功: %d，失败: %d", successCount, len(stories)-successCount)

	return results
}

// GenerateReport 生成导入报告
func (i *Importer) GenerateReport(results []ImportResult) string {
	var totalCount, successCount int
	var totalTime time.Duration
	var report string

	report += "\n=== 需求导入报告 ===\n\n"

	// 统计结果
	for idx, result := range results {
		if result.Success {
			successCount++
			report += fmt.Sprintf("✓ 需求 #%d 导入成功 (ID: %d, 耗时: %v)\n",
				idx+1, result.StoryID, result.ElapsedTime)
		} else {
			report += fmt.Sprintf("✗ 需求 #%d 导入失败: %v\n",
				idx+1, result.Error)
			if result.ResponseMsg != "" {
				report += fmt.Sprintf("    响应内容: %s\n", i.truncateResponse(result.ResponseMsg))
			}
		}
		totalTime += result.ElapsedTime
	}
	totalCount = len(results)

	// 添加统计信息
	report += "\n总计统计:\n"
	report += fmt.Sprintf("- 总需求数: %d\n", totalCount)
	report += fmt.Sprintf("- 成功导入: %d\n", successCount)
	report += fmt.Sprintf("- 失败数量: %d\n", totalCount-successCount)
	report += fmt.Sprintf("- 总耗时: %v\n", totalTime)
	if totalCount > 0 {
		report += fmt.Sprintf("- 平均耗时: %v\n", totalTime/time.Duration(totalCount))
		report += fmt.Sprintf("- 成功率: %.1f%%\n", float64(successCount)/float64(totalCount)*100)
	} else {
		report += "- 平均耗时: N/A\n"
		report += "- 成功率: N/A\n"
	}

	return report
}

// truncateResponse 截断响应内容，避免日志过长
func (i *Importer) truncateResponse(s string) string {
	const maxLen = 200
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
