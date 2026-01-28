// Package zentao 封装禅道API客户端
package zentao

import (
	"fmt"
	"time"

	"github.com/jan2xue/zentao_import_story/internal/logger"
	"github.com/jan2xue/zentao_import_story/pkg/story"
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
	result := ImportResult{}

	// 转换为禅道需求格式
	zentaoStory := s.ToZentaoStory()

	i.logger.Info("正在导入需求: %s", s.Title)

	// 创建需求
	createdStory, rsp, err := i.client.GetClient().Stories.Create(*zentaoStory)
	if (err != nil) || (rsp != nil && rsp.StatusCode != 200) {
		errMsg := fmt.Sprintf("创建需求失败: %v", err)
		i.logger.Error("%s", errMsg)
		result.Error = fmt.Errorf("%s", errMsg)
		result.Success = false
	} else {
		i.logger.Success("需求创建成功，ID: %d", createdStory.ID)
		result.StoryID = createdStory.ID
		result.Success = true
	}

	result.ElapsedTime = time.Since(start)
	return result
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
