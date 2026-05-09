// Package zentao 封装禅道API客户端
package zentao

import (
	"fmt"
	"time"

	"github.com/imroc/req/v3"
	"github.com/jan2xue/zentao_import_story/internal/logger"
	"github.com/jan2xue/zentao_import_story/pkg/story"
)

// TypedID 带类型的需求ID
type TypedID struct {
	ID   int
	Type story.StoryType
}

// DeleteResult 表示删除结果
type DeleteResult struct {
	Success     bool
	StoryID     int
	Error       error
	ElapsedTime time.Duration
	ResponseMsg string // 响应消息
}

// Deleter 处理需求删除操作
type Deleter struct {
	client       *Client
	logger       *logger.Logger
	epicDeleter  EpicCreator
	reqDeleter   RequirementCreator
	storyDeleter StoryCreator
}

// NewDeleter 创建新的删除器
func NewDeleter(client *Client, log *logger.Logger) *Deleter {
	return &Deleter{
		client:       client,
		logger:       log,
		epicDeleter:  client.Epic,
		reqDeleter:   client.Requirement,
		storyDeleter: client.Story,
	}
}

// NewDeleterWithMocks 创建删除器（用于测试）
func NewDeleterWithMocks(log *logger.Logger, epic EpicCreator, req RequirementCreator, story StoryCreator) *Deleter {
	return &Deleter{
		logger:       log,
		epicDeleter:  epic,
		reqDeleter:   req,
		storyDeleter: story,
	}
}

// DeleteStory 删除单个需求（按ID删除，需指定类型以选择正确的API端点）
func (d *Deleter) DeleteStory(storyID int, storyType story.StoryType) DeleteResult {
	start := time.Now()
	result := DeleteResult{
		StoryID: storyID,
	}

	d.logger.Info("正在删除需求 ID: %d, 类型: %s", storyID, storyType)

	var rsp *req.Response
	var err error

	switch storyType {
	case story.StoryTypeEpic:
		_, rsp, err = d.epicDeleter.DeleteByID(storyID)
	case story.StoryTypeRequirement:
		_, rsp, err = d.reqDeleter.DeleteByID(storyID)
	case story.StoryTypeStory:
		_, rsp, err = d.storyDeleter.DeleteByID(storyID)
	default:
		_, rsp, err = d.storyDeleter.DeleteByID(storyID)
	}

	if err != nil {
		d.logger.ErrorWithDetail("需求删除失败", err, map[string]interface{}{
			"需求ID":    storyID,
			"需求类型":   string(storyType),
			"HTTP状态码": d.getStatusCode(rsp),
			"响应内容":   d.getResponseBody(rsp),
		})
		result.Error = fmt.Errorf("删除需求失败: %w", err)
		result.Success = false
		result.ResponseMsg = d.getResponseBody(rsp)
	} else if rsp != nil && rsp.StatusCode >= 400 {
		d.logger.ErrorWithDetail("需求删除失败（HTTP错误）", fmt.Errorf("HTTP状态码: %d", rsp.StatusCode), map[string]interface{}{
			"需求ID":    storyID,
			"需求类型":   string(storyType),
			"HTTP状态码": rsp.StatusCode,
			"响应内容":   d.getResponseBody(rsp),
		})
		result.Error = fmt.Errorf("删除需求失败，HTTP状态码: %d", rsp.StatusCode)
		result.Success = false
		result.ResponseMsg = d.getResponseBody(rsp)
	} else {
		d.logger.Success("需求删除成功，ID: %d", storyID)
		result.Success = true
	}

	result.ElapsedTime = time.Since(start)
	return result
}

// DeleteStories 批量删除需求
func (d *Deleter) DeleteStories(ids []TypedID) []DeleteResult {
	results := make([]DeleteResult, len(ids))

	d.logger.Info("开始批量删除需求，共 %d 个需求", len(ids))

	for idx, id := range ids {
		d.logger.Info("正在删除第 %d/%d 个需求", idx+1, len(ids))
		results[idx] = d.DeleteStory(id.ID, id.Type)
	}

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}
	d.logger.Info("需求删除完成，成功: %d，失败: %d", successCount, len(ids)-successCount)

	return results
}

// GenerateDeleteReport 生成删除报告
func (d *Deleter) GenerateDeleteReport(results []DeleteResult) string {
	var totalCount, successCount int
	var totalTime time.Duration
	var report string

	report += "\n=== 需求删除报告 ===\n\n"

	// 统计结果
	for idx, result := range results {
		if result.Success {
			successCount++
			report += fmt.Sprintf("✓ 需求 #%d (ID: %d) 删除成功 (耗时: %v)\n",
				idx+1, result.StoryID, result.ElapsedTime)
		} else {
			report += fmt.Sprintf("✗ 需求 #%d (ID: %d) 删除失败: %v\n",
				idx+1, result.StoryID, result.Error)
			if result.ResponseMsg != "" {
				report += fmt.Sprintf("    响应内容: %s\n", d.truncateResponse(result.ResponseMsg))
			}
		}
		totalTime += result.ElapsedTime
	}
	totalCount = len(results)

	// 添加统计信息
	report += "\n总计统计:\n"
	report += fmt.Sprintf("- 总需求数: %d\n", totalCount)
	report += fmt.Sprintf("- 成功删除: %d\n", successCount)
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

// getStatusCode 安全获取HTTP状态码
func (d *Deleter) getStatusCode(rsp *req.Response) int {
	if rsp == nil {
		return 0
	}
	return rsp.StatusCode
}

// getResponseBody 安全获取响应体
func (d *Deleter) getResponseBody(rsp *req.Response) string {
	if rsp == nil {
		return ""
	}
	return rsp.String()
}

// truncateResponse 截断响应内容，避免日志过长
func (d *Deleter) truncateResponse(s string) string {
	const maxLen = 200
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
