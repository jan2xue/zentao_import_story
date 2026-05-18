// Package zentao 封装禅道API客户端
package zentao

import (
	"fmt"
	"strconv"
	"strings"
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
	client       *Client
	logger       *logger.Logger
	epicCreator  EpicCreator
	reqCreator   RequirementCreator
	storyCreator StoryCreator
	config       ConfigProvider
}

// NewImporter 创建新的导入器
func NewImporter(client *Client, log *logger.Logger) *Importer {
	return &Importer{
		client:       client,
		logger:       log,
		epicCreator:  client.Epic,
		reqCreator:   client.Requirement,
		storyCreator: client.Story,
		config:       client.config,
	}
}

// NewImporterWithMocks 创建导入器（用于测试，直接注入mock实现）
func NewImporterWithMocks(log *logger.Logger, epic EpicCreator, req RequirementCreator, story StoryCreator, cfg ConfigProvider) *Importer {
	return &Importer{
		logger:       log,
		epicCreator:  epic,
		reqCreator:   req,
		storyCreator: story,
		config:       cfg,
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
		ProductID:  s.ProductID,
		Title:      s.Title,
		Pri:        s.Priority,
		Grade:      1,
		Spec:       s.Spec,
		Category:   s.Category,
		Parent:     s.ParentID,
		Source:     s.Source,
		SourceNote: s.SourceNote,
		Keywords:   s.Keywords,
		Verify:     s.Verify,
		Estimate:   s.Estimate,
	}

	// 设置评审人（如果配置了默认评审人）
	if i.config.GetDefaultReviewer() != "" {
		req.Reviewer = []string{i.config.GetDefaultReviewer()}
	}

	i.logger.Debug("创建业务请求 - 产品ID: %d, 标题: %s", s.ProductID, s.Title)

	resp, rsp, err := i.epicCreator.Create(req)
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
		ProductID:  s.ProductID,
		Title:      s.Title,
		Pri:        s.Priority,
		Grade:      1,
		Spec:       s.Spec,
		Category:   s.Category,
		Parent:     s.ParentID,
		Source:     s.Source,
		SourceNote: s.SourceNote,
		Keywords:   s.Keywords,
		Verify:     s.Verify,
		Estimate:   s.Estimate,
	}

	// 设置模块ID（用户需求必须指定有效的模块ID）
	if i.config.GetDefaultModule() >= 0 {
		req.Module = i.config.GetDefaultModule()
	} else {
		return 0, nil, fmt.Errorf("创建用户需求需要有效的模块ID，请在config.yaml中配置defaultModule参数")
	}

	// 设置评审人（如果配置了默认评审人）
	if i.config.GetDefaultReviewer() != "" {
		req.Reviewer = []string{i.config.GetDefaultReviewer()}
	}

	i.logger.Debug("创建用户需求请求 - 产品ID: %d, 标题: %s, 模块ID: %d", s.ProductID, s.Title, req.Module)

	resp, rsp, err := i.reqCreator.Create(req)
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
		ProductID:  s.ProductID,
		Title:      s.Title,
		Pri:        s.Priority,
		Grade:      1,
		Spec:       s.Spec,
		Category:   s.Category,
		Parent:     s.ParentID,
		Source:     s.Source,
		SourceNote: s.SourceNote,
		Keywords:   s.Keywords,
		Verify:     s.Verify,
		Estimate:   s.Estimate,
	}

	// 设置评审人（如果配置了默认评审人）
	if i.config.GetDefaultReviewer() != "" {
		req.Reviewer = []string{i.config.GetDefaultReviewer()}
	}

	i.logger.Debug("创建研发需求请求 - 产品ID: %d, 标题: %s", s.ProductID, s.Title)

	resp, rsp, err := i.storyCreator.Create(req)
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

// ImportStories 按层级导入需求（Epic → Requirement → Story）
// 解析 "@行号" 格式的父需求引用，自动替换为实际创建的禅道ID
// Epic/Requirement创建API不返回ID，需通过产品列表查询获取实际ID，确保父子关系正确建立
func (i *Importer) ImportStories(stories []story.Story) []ImportResult {
	results := make([]ImportResult, len(stories))
	// 行号到禅道ID的映射（用于解析 @n 引用）
	rowIDMap := make(map[int]int)

	// 按层级分组并保持原始顺序
	var epics, requirements, storiesGroup []int // 存储在stories切片中的索引
	for idx, s := range stories {
		switch s.Type {
		case story.StoryTypeEpic:
			epics = append(epics, idx)
		case story.StoryTypeRequirement:
			requirements = append(requirements, idx)
		case story.StoryTypeStory:
			storiesGroup = append(storiesGroup, idx)
		}
	}

	i.logger.Info("开始层级导入: %d个业务需求 → %d个用户需求 → %d个研发需求",
		len(epics), len(requirements), len(storiesGroup))

	// 第一阶段：导入 Epic
	i.logger.Info("========== 阶段1: 导入业务需求(Epic) ==========")
	for _, idx := range epics {
		s := &stories[idx]
		i.resolveParentRef(s, rowIDMap)
		results[idx] = i.ImportStory(s)
		if results[idx].Success {
			// Epic创建API不返回ID，通过产品列表查询获取实际ID
			actualID := i.resolveCreatedID(s.Type, s.ProductID, s.Title, results[idx].StoryID)
			if actualID > 0 {
				results[idx].StoryID = actualID
				rowIDMap[s.RowIndex] = actualID
			} else {
				rowIDMap[s.RowIndex] = results[idx].StoryID
			}
		}
	}

	// 第二阶段：解析 Requirement 的父引用并导入
	i.logger.Info("========== 阶段2: 导入用户需求(Requirement) ==========")
	for _, idx := range requirements {
		s := &stories[idx]
		i.resolveParentRef(s, rowIDMap)
		results[idx] = i.ImportStory(s)
		if results[idx].Success {
			// Requirement创建API不返回ID，通过产品列表查询获取实际ID
			actualID := i.resolveCreatedID(s.Type, s.ProductID, s.Title, results[idx].StoryID)
			if actualID > 0 {
				results[idx].StoryID = actualID
				rowIDMap[s.RowIndex] = actualID
			} else {
				rowIDMap[s.RowIndex] = results[idx].StoryID
			}
		}
	}

	// 第三阶段：解析 Story 的父引用并导入
	i.logger.Info("========== 阶段3: 导入研发需求(Story) ==========")
	for _, idx := range storiesGroup {
		s := &stories[idx]
		i.resolveParentRef(s, rowIDMap)
		results[idx] = i.ImportStory(s)
		// Story创建API会返回ID，无需额外查询
		if results[idx].Success {
			rowIDMap[s.RowIndex] = results[idx].StoryID
		}
	}

	// 汇总统计
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}
	i.logger.Info("层级导入完成，成功: %d，失败: %d", successCount, len(stories)-successCount)

	return results
}

// resolveCreatedID 查询禅道获取需求创建后的实际ID
// Epic和Requirement的创建API不返回ID，需要通过产品列表按标题匹配查询
// Story类型的创建API会返回ID，直接使用即可
// 去重策略：Epic API会返回关联的Requirement和Story，Requirement API会返回关联的Story
// 因此查询时需要排除低层级中已存在的ID，确保匹配到正确类型的实际ID
func (i *Importer) resolveCreatedID(sType story.StoryType, productID int, title string, createRespID int) int {
	// Story类型创建API会返回ID，直接使用
	if sType == story.StoryTypeStory && createRespID > 0 {
		return createRespID
	}

	// Epic/Requirement创建API不返回ID，通过产品列表按标题匹配查询
	// 需要去重：先获取低层级列表的ID，在高层级列表中排除这些ID
	switch sType {
	case story.StoryTypeEpic:
		// 查询Epic时需排除Story和Requirement中已有的ID
		seenIDs := make(map[int]bool)

		stories, err := i.storyCreator.ProductsListAll(productID)
		if err != nil {
			i.logger.Info("查询产品研发需求列表失败(产品ID=%d): %v", productID, err)
		} else {
			for _, s := range stories {
				seenIDs[s.ID] = true
			}
		}

		requirements, err := i.reqCreator.ProductsListAll(productID)
		if err != nil {
			i.logger.Info("查询产品用户需求列表失败(产品ID=%d): %v", productID, err)
		} else {
			for _, r := range requirements {
				seenIDs[r.ID] = true
			}
		}

		epics, err := i.epicCreator.ProductsListAll(productID)
		if err != nil {
			i.logger.Info("查询产品业务需求列表失败(产品ID=%d): %v", productID, err)
			return createRespID
		}
		for _, e := range epics {
			if seenIDs[e.ID] {
				continue // 跳过已在Story或Requirement中出现的ID
			}
			if e.Title == title {
				i.logger.Info("通过产品列表查询到业务需求实际ID: %d (标题: %s)", e.ID, title)
				return e.ID
			}
		}
		i.logger.Info("在产品业务需求列表中未找到匹配项(产品ID=%d, 标题=%s)", productID, title)

	case story.StoryTypeRequirement:
		// 查询Requirement时需排除Story中已有的ID
		seenIDs := make(map[int]bool)

		stories, err := i.storyCreator.ProductsListAll(productID)
		if err != nil {
			i.logger.Info("查询产品研发需求列表失败(产品ID=%d): %v", productID, err)
		} else {
			for _, s := range stories {
				seenIDs[s.ID] = true
			}
		}

		requirements, err := i.reqCreator.ProductsListAll(productID)
		if err != nil {
			i.logger.Info("查询产品用户需求列表失败(产品ID=%d): %v", productID, err)
			return createRespID
		}
		for _, r := range requirements {
			if seenIDs[r.ID] {
				continue // 跳过已在Story中出现的ID
			}
			if r.Title == title {
				i.logger.Info("通过产品列表查询到用户需求实际ID: %d (标题: %s)", r.ID, title)
				return r.ID
			}
		}
		i.logger.Info("在产品用户需求列表中未找到匹配项(产品ID=%d, 标题=%s)", productID, title)
	}

	// 如果创建API返回了ID（某些版本可能支持），直接使用
	if createRespID > 0 {
		return createRespID
	}

	i.logger.Info("无法获取需求实际ID(标题=%s)，创建返回ID=%d", title, createRespID)
	return 0
}

// resolveParentRef 解析父需求引用，将 "@行号" 格式替换为实际的禅道ID
func (i *Importer) resolveParentRef(s *story.Story, rowIDMap map[int]int) {
	if s.ParentRef == "" {
		return
	}

	// 检查是否为 "@行号" 格式
	ref := strings.TrimPrefix(s.ParentRef, "@")
	if ref == s.ParentRef {
		// 不是 @ 格式，ParentID 已经在解析时设置
		return
	}

	rowNum, err := strconv.Atoi(ref)
	if err != nil {
		i.logger.Error("无效的父需求引用格式: %s（行%d），应为 @行号", s.ParentRef, s.RowIndex)
		return
	}

	parentID, ok := rowIDMap[rowNum]
	if !ok {
		i.logger.Error("父需求引用 @%d 未找到对应的禅道ID（行%d），该行可能导入失败", rowNum, s.RowIndex)
		return
	}

	s.ParentID = parentID
	i.logger.Debug("行%d 父需求引用 @%d 解析为禅道ID: %d", s.RowIndex, rowNum, parentID)
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
