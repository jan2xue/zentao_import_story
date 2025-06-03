package main

import (
	"fmt"
	"time"

	"github.com/easysoft/go-zentao/v21/zentao"
)

// 全局日志记录器变量，在main.go中初始化
var logger *Logger

// ImportResult 表示导入结果
type ImportResult struct {
	Success     bool
	StoryID     int
	Error       error
	ElapsedTime time.Duration
}

// Importer 处理需求导入到禅道
type Importer struct {
	client *zentao.Client
	config *Config
}

// NewImporter 创建新的导入器
func NewImporter(config *Config) (*Importer, error) {
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
	return &Importer{
		client: client,
		config: config,
	}, nil
}

// ImportStory 导入单个需求
func (i *Importer) ImportStory(story *ExcelStory) ImportResult {
	start := time.Now()
	result := ImportResult{}

	// 转换为禅道需求格式
	zentaoStory := story.ToZentaoStory()

	logger.Info("正在导入需求: %s", story.Title)

	// 创建需求
	createdStory, rsp, err := i.client.Stories.Create(*zentaoStory)
	if (err != nil) || (rsp != nil && rsp.StatusCode != 200) {
		errMsg := fmt.Sprintf("创建需求失败: %v", err)
		logger.Error(errMsg)
		result.Error = fmt.Errorf(errMsg)
		result.Success = false
	} else {
		logger.Success("需求创建成功，ID: %d", createdStory.ID)
		result.StoryID = createdStory.ID
		result.Success = true
	}

	result.ElapsedTime = time.Since(start)
	return result
}

// ImportStories 批量导入需求
func (i *Importer) ImportStories(stories []ExcelStory) []ImportResult {
	results := make([]ImportResult, len(stories))

	logger.Info("开始批量导入需求，共 %d 个需求", len(stories))

	for idx, story := range stories {
		logger.Info("正在导入第 %d/%d 个需求", idx+1, len(stories))
		results[idx] = i.ImportStory(&story)
	}

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}
	logger.Info("需求导入完成，成功: %d，失败: %d", successCount, len(stories)-successCount)

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
		}
		totalTime += result.ElapsedTime
	}
	totalCount = len(results)

	// 添加统计信息
	report += fmt.Sprintf("\n总计统计:\n")
	report += fmt.Sprintf("- 总需求数: %d\n", totalCount)
	report += fmt.Sprintf("- 成功导入: %d\n", successCount)
	report += fmt.Sprintf("- 失败数量: %d\n", totalCount-successCount)
	report += fmt.Sprintf("- 总耗时: %v\n", totalTime)
	report += fmt.Sprintf("- 平均耗时: %v\n", totalTime/time.Duration(totalCount))
	report += fmt.Sprintf("- 成功率: %.1f%%\n", float64(successCount)/float64(totalCount)*100)

	return report
}
