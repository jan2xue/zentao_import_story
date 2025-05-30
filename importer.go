package main

import (
	"fmt"
	"time"

	"github.com/easysoft/go-zentao/v21/zentao"
)

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

	// 创建需求
	createdStory, rsp, err := i.client.Stories.Create(*zentaoStory)
	if (err != nil) || (rsp != nil && rsp.StatusCode != 200) {
		result.Error = fmt.Errorf("创建需求失败: %w", err)
		result.Success = false
	} else {
		result.StoryID = createdStory.ID
		result.Success = true
	}

	result.ElapsedTime = time.Since(start)
	return result
}

// ImportStories 批量导入需求
func (i *Importer) ImportStories(stories []ExcelStory) []ImportResult {
	results := make([]ImportResult, len(stories))

	for idx, story := range stories {
		// 导入需求
		results[idx] = i.ImportStory(&story)
	}

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
