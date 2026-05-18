// Package zentao 封装禅道API客户端
package zentao

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/jan2xue/zentao_import_story/internal/logger"
	"github.com/jan2xue/zentao_import_story/pkg/story"
)

// TypedID 带类型的需求ID
type TypedID struct {
	ID       int
	Type     story.StoryType
	Title    string
	OpenedBy string
}

// DeleteResult 表示删除结果
type DeleteResult struct {
	Success     bool
	StoryID     int
	StoryType   string
	Title       string
	Error       error
	ElapsedTime time.Duration
	ResponseMsg string // 响应消息
}

// DeleteFilter 筛选条件
type DeleteFilter struct {
	ProductID int    // 产品ID（必填）
	Title     string // 标题部分匹配（可选，为空则不按标题筛选）
	OpenedBy  string // 创建者筛选（可选，为空则不按创建者筛选）
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
		StoryID:   storyID,
		StoryType: string(storyType),
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
		result := d.DeleteStory(id.ID, id.Type)
		result.Title = id.Title
		results[idx] = result
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

// DeleteStoriesConcurrent 并发批量删除需求（大批量时提升性能）
func (d *Deleter) DeleteStoriesConcurrent(ids []TypedID, concurrency int) []DeleteResult {
	if concurrency <= 0 {
		concurrency = 3
	}
	if concurrency > 10 {
		concurrency = 10
	}

	results := make([]DeleteResult, len(ids))
	d.logger.Info("开始并发批量删除需求，共 %d 个需求，并发数: %d", len(ids), concurrency)

	// 用信号量控制并发数
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for idx, id := range ids {
		wg.Add(1)
		go func(idx int, id TypedID) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			d.logger.Info("正在删除需求 ID: %d, 类型: %s, 标题: %s", id.ID, id.Type, id.Title)
			result := d.DeleteStory(id.ID, id.Type)
			result.Title = id.Title

			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(idx, id)
	}

	wg.Wait()

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}
	d.logger.Info("并发批量删除完成，成功: %d，失败: %d", successCount, len(ids)-successCount)

	return results
}

// FetchByFilter 按筛选条件获取需求列表
// 支持按产品ID（必填）+ 标题部分匹配 + 创建者筛选
// 去重策略：Epic API会返回关联的Requirement和Story，Requirement API会返回关联的Story
// 因此先获取Story，再获取Requirement（去除Story中已有的ID），最后获取Epic（去除Story和Requirement中已有的ID）
func (d *Deleter) FetchByFilter(filter DeleteFilter) []TypedID {
	var matched []TypedID
	seenIDs := make(map[int]bool) // 已处理的ID集合，用于去重

	// 第一步：获取Story列表（最小集合，不含其他类型的关联数据）
	stories, err := d.storyDeleter.ProductsListAll(filter.ProductID)
	if err != nil {
		d.logger.Error("获取产品研发需求列表失败: %v", err)
	} else {
		for _, s := range stories {
			seenIDs[s.ID] = true
			if d.matchFilter(s.Title, s.OpenedBy, filter) {
				matched = append(matched, TypedID{ID: s.ID, Type: story.StoryTypeStory, Title: s.Title, OpenedBy: s.OpenedBy})
			}
		}
	}

	// 第二步：获取Requirement列表，去除已在Story中出现的ID
	requirements, err := d.reqDeleter.ProductsListAll(filter.ProductID)
	if err != nil {
		d.logger.Error("获取产品用户需求列表失败: %v", err)
	} else {
		for _, r := range requirements {
			if seenIDs[r.ID] {
				continue // 跳过已在Story列表中出现的ID
			}
			seenIDs[r.ID] = true
			if d.matchFilter(r.Title, r.OpenedBy, filter) {
				matched = append(matched, TypedID{ID: r.ID, Type: story.StoryTypeRequirement, Title: r.Title, OpenedBy: r.OpenedBy})
			}
		}
	}

	// 第三步：获取Epic列表，去除已在Story或Requirement中出现的ID
	epics, err := d.epicDeleter.ProductsListAll(filter.ProductID)
	if err != nil {
		d.logger.Error("获取产品业务需求列表失败: %v", err)
	} else {
		for _, e := range epics {
			if seenIDs[e.ID] {
				continue // 跳过已在Story或Requirement列表中出现的ID
			}
			if d.matchFilter(e.Title, e.OpenedBy, filter) {
				matched = append(matched, TypedID{ID: e.ID, Type: story.StoryTypeEpic, Title: e.Title, OpenedBy: e.OpenedBy})
			}
		}
	}

	return matched
}

// matchFilter 检查需求是否匹配筛选条件（标题部分匹配，创建者精确匹配）
func (d *Deleter) matchFilter(title, openedBy string, filter DeleteFilter) bool {
	if filter.Title != "" && !strings.Contains(title, filter.Title) {
		return false
	}
	if filter.OpenedBy != "" && openedBy != filter.OpenedBy {
		return false
	}
	return true
}

// GenerateDeleteReport 生成删除报告
func (d *Deleter) GenerateDeleteReport(results []DeleteResult) string {
	var totalCount, successCount int
	var totalTime time.Duration
	var report string

	report += "\n=== 需求删除报告 ===\n\n"

	// 统计结果
	for idx, result := range results {
		title := result.Title
		if title == "" {
			title = "(未知)"
		}
		typeInfo := result.StoryType
		if typeInfo != "" {
			typeInfo = "[" + typeInfo + "] "
		}
		if result.Success {
			successCount++
			report += fmt.Sprintf("✓ %s需求 #%d (ID: %d, 标题: %s) 删除成功 (耗时: %v)\n",
				typeInfo, idx+1, result.StoryID, title, result.ElapsedTime)
		} else {
			report += fmt.Sprintf("✗ %s需求 #%d (ID: %d, 标题: %s) 删除失败: %v\n",
				typeInfo, idx+1, result.StoryID, title, result.Error)
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

// FormatMatchedList 格式化匹配结果列表用于显示
func FormatMatchedList(items []TypedID) string {
	if len(items) == 0 {
		return "  (无匹配结果)"
	}

	var b strings.Builder
	// 按类型分组统计
	typeCounts := make(map[story.StoryType]int)
	for _, item := range items {
		typeCounts[item.Type]++
	}

	b.WriteString(fmt.Sprintf("  共 %d 条需求:\n", len(items)))
	for _, t := range []story.StoryType{story.StoryTypeEpic, story.StoryTypeRequirement, story.StoryTypeStory} {
		if cnt, ok := typeCounts[t]; ok {
			b.WriteString(fmt.Sprintf("    - %s: %d 条\n", getTypeDisplayName(t), cnt))
		}
	}

	b.WriteString("\n  匹配列表:\n")
	b.WriteString(fmt.Sprintf("  %-6s %-10s %-6s %-12s %s\n", "序号", "类型", "ID", "创建者", "标题"))
	b.WriteString(fmt.Sprintf("  %-6s %-10s %-6s %-12s %s\n", "----", "--------", "----", "----------", "--------------------"))
	for i, item := range items {
		b.WriteString(fmt.Sprintf("  %-6d %-10s %-6d %-12s %s\n", i+1, getTypeDisplayName(item.Type), item.ID, item.OpenedBy, item.Title))
	}

	return b.String()
}

// getTypeDisplayName 获取需求类型的显示名称
func getTypeDisplayName(t story.StoryType) string {
	switch t {
	case story.StoryTypeEpic:
		return "业务需求"
	case story.StoryTypeRequirement:
		return "用户需求"
	case story.StoryTypeStory:
		return "研发需求"
	default:
		return "未知"
	}
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
