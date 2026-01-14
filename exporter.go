package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/easysoft/go-zentao/v21/zentao"
)

// ExportResult 表示导出结果
type ExportResult struct {
	Success     bool
	StoryCount  int
	Error       error
	ElapsedTime int64 // 以毫秒为单位
}

// Exporter 处理从禅道导出需求
type Exporter struct {
	client *zentao.Client
	config *Config
}

// NewExporter 创建新的导出器
func NewExporter(config *Config) (*Exporter, error) {
	client, err := zentao.NewBasicAuthClient(
		config.ZentaoUsername,
		config.ZentaoPassword,
		zentao.WithBaseURL(config.ZentaoURL),
		zentao.WithDevMode(),
		zentao.WithDumpAll(),
		zentao.WithoutProxy(),
	)
	if err != nil {
		return nil, fmt.Errorf("创建禅道客户端失败: %w", err)
	}
	return &Exporter{
		client: client,
		config: config,
	}, nil
}

// ExportStories 从禅道服务器导出需求（支持分页）
func (e *Exporter) ExportStories(productID int) ([]ExcelStory, ExportResult) {
	result := ExportResult{}
	startTime := time.Now()

	logger.Info("正在从禅道服务器导出需求，产品ID: %d", productID)

	var allExcelStories []ExcelStory
	page := 1
	totalFetched := 0

	// 循环获取所有分页数据
	for {
		logger.Info("正在获取第 %d 页需求...", page)

		// 使用ZenTao API获取产品需求列表
		storiesList, rsp, err := e.client.Stories.ProductsList(productID)
		if err != nil || (rsp != nil && rsp.StatusCode >= 400) {
			errMsg := fmt.Sprintf("获取产品需求列表失败(第%d页): %v", page, err)
			logger.Error("%s", errMsg)
			result.Error = fmt.Errorf("%s", errMsg)
			result.Success = false
			result.ElapsedTime = time.Since(startTime).Milliseconds()
			return nil, result
		}

		// 获取当前页的需求数量
		currentPageCount := len(storiesList.Stories)
		logger.Info("第 %d 页获取到 %d 个需求", page, currentPageCount)

		// 如果当前页没有数据，退出循环
		if currentPageCount == 0 {
			break
		}

		// 将禅道需求转换为Excel格式
		for _, zentaoStory := range storiesList.Stories {
			excelStory := e.zentaoStoryToExcelStory(zentaoStory)
			if excelStory != nil {
				allExcelStories = append(allExcelStories, *excelStory)
			}
		}

		totalFetched += currentPageCount

		// 检查是否还有更多页
		// 如果当前页的数量小于limit，说明已经是最后一页
		if storiesList.Limit > 0 && currentPageCount < storiesList.Limit {
			logger.Info("已到达最后一页")
			break
		}

		// 如果有total字段，检查是否已获取所有数据
		if storiesList.Total > 0 && totalFetched >= storiesList.Total {
			logger.Info("已获取所有 %d 个需求", storiesList.Total)
			break
		}

		page++
	}

	result.Success = true
	result.StoryCount = len(allExcelStories)
	result.ElapsedTime = time.Since(startTime).Milliseconds()
	logger.Info("成功转换 %d 个需求为Excel格式，耗时: %dms", len(allExcelStories), result.ElapsedTime)

	return allExcelStories, result
}

// zentaoStoryToExcelStory 将禅道需求转换为Excel需求格式
func (e *Exporter) zentaoStoryToExcelStory(zentaoStory zentao.StoriesBody) *ExcelStory {
	// 从StoriesBody获取基本信息
	priority := zentaoStory.Pri
	if priority < 1 || priority > 4 {
		priority = 3 // 默认优先级
	}

	// 处理Parent字段，由于其类型为any，需要特别处理
	parentID := 0
	if zentaoStory.Parent != nil {
		// 尝试转换为整数
		if id, ok := zentaoStory.Parent.(float64); ok {
			parentID = int(id)
		} else if id, ok := zentaoStory.Parent.(int); ok {
			parentID = id
		} else if idStr, ok := zentaoStory.Parent.(string); ok {
			// 如果是字符串，尝试转换为整数
			if id, err := strconv.Atoi(idStr); err == nil {
				parentID = id
			}
		}
	}

	// 获取完整的需求信息，包括Spec和Verify字段
	fullStory, rsp, err := e.client.Stories.GetByID(zentaoStory.ID)
	if err != nil || (rsp != nil && rsp.StatusCode >= 400) {
		logger.Error("获取需求详细信息失败，ID: %d, 错误: %v", zentaoStory.ID, err)
		// 如果获取详细信息失败，使用列表中的基本信息
		return &ExcelStory{
			Title:     zentaoStory.Title,
			ProductID: zentaoStory.Product,
			Priority:  priority,
			Category:  string(zentaoStory.Category),
			ParentID:  parentID,
			Source:    string(zentaoStory.Source),
		}
	}

	return &ExcelStory{
		Title:      fullStory.Title,
		ProductID:  fullStory.Product,
		Priority:   fullStory.Pri,
		Category:   string(fullStory.Category),
		Spec:       fullStory.Spec,
		ParentID:   parentID, // 使用从列表获取的parentID，因为GetByID返回的可能格式不同
		Source:     string(fullStory.Source),
		SourceNote: fullStory.SourceNote,
		Estimate:   fullStory.Estimate,
		Keywords:   fullStory.Keywords,
		Verify:     fullStory.Verify,
	}
}
