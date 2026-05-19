// Package excel 处理Excel文件的读写操作
package excel

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
	"github.com/jan2xue/zentao_import_story/pkg/story"
)

// Reader 处理Excel文件的读取和验证
type Reader struct {
	file *excelize.File
}

// NewReader 创建新的Excel读取器
func NewReader(filePath string) (*Reader, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %w", err)
	}
	return &Reader{file: f}, nil
}

// Close 关闭Excel文件
func (r *Reader) Close() error {
	return r.file.Close()
}

// ReadStories 读取层级需求数据
// Excel列定义: 需求类型 | 产品ID | 模块ID | 标题 | 优先级 | 分类 | 需求描述 | 父需求ID | 来源 | 来源备注 | 预计工时 | 关键词 | 验收标准
// 父需求ID支持格式: "@n"引用第n行数据的禅道ID，或纯数字作为实际禅道ID
func (r *Reader) ReadStories(defaultPriority int) ([]story.Story, error) {
	sheets := r.file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件中没有工作表")
	}

	rows, err := r.file.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel文件中没有数据")
	}

	var stories []story.Story
	for i, row := range rows[1:] {
		s, err := r.parseRow(row, defaultPriority, i+1)
		if err != nil {
			return nil, fmt.Errorf("第%d行数据解析失败: %w", i+2, err)
		}
		stories = append(stories, s)
	}

	return stories, nil
}

// parseRow 解析Excel行数据
// Excel列定义: 需求类型 | 产品ID | 模块ID | 标题 | 优先级 | 分类 | 需求描述 | 父需求ID | 来源 | 来源备注 | 预计工时 | 关键词 | 验收标准
func (r *Reader) parseRow(row []string, defaultPriority int, rowIndex int) (story.Story, error) {
	if len(row) < 7 {
		return story.Story{}, fmt.Errorf("行数据不完整，缺少必填字段，当前列数: %d", len(row))
	}

	// 解析需求类型 (第1列)
	storyType := story.StoryTypeStory
	switch strings.ToLower(strings.TrimSpace(row[0])) {
	case "epic":
		storyType = story.StoryTypeEpic
	case "requirement":
		storyType = story.StoryTypeRequirement
	case "story":
		storyType = story.StoryTypeStory
	default:
		return story.Story{}, fmt.Errorf("无效的需求类型: %s，支持: epic/requirement/story", strings.TrimSpace(row[0]))
	}

	// 解析产品ID (第2列)
	productID, err := strconv.Atoi(strings.TrimSpace(row[1]))
	if err != nil {
		return story.Story{}, fmt.Errorf("产品ID必须是数字: %w", err)
	}

	// 解析模块ID (第3列) - 可选，空表示未指定将使用配置文件默认值，0为合法值表示不归属具体模块
	moduleID := -1 // -1 表示Excel未填写
	if len(row) > 2 && strings.TrimSpace(row[2]) != "" {
		moduleID, err = strconv.Atoi(strings.TrimSpace(row[2]))
		if err != nil {
			return story.Story{}, fmt.Errorf("模块ID必须是数字: %w", err)
		}
		if moduleID < 0 {
			return story.Story{}, fmt.Errorf("模块ID不能为负数: %d", moduleID)
		}
		// moduleID >= 0 表示Excel显式指定了模块ID（0也是合法值）
	}

	// 解析标题 (第4列)
	title := strings.TrimSpace(row[3])
	if title == "" {
		return story.Story{}, fmt.Errorf("标题不能为空")
	}

	// 解析优先级 (第5列)
	priority := defaultPriority
	if len(row) > 4 && strings.TrimSpace(row[4]) != "" {
		p, err := strconv.Atoi(strings.TrimSpace(row[4]))
		if err != nil || p < 1 || p > 4 {
			return story.Story{}, fmt.Errorf("优先级必须是1-4之间的数字")
		}
		priority = p
	}

	// 解析分类 (第6列)
	category := strings.TrimSpace(row[5])
	if category == "" {
		return story.Story{}, fmt.Errorf("分类不能为空")
	}

	s := story.Story{
		Type:      storyType,
		Title:     title,
		ProductID: productID,
		Priority:  priority,
		Category:  category,
		Module:    moduleID,
		RowIndex:  rowIndex,
	}

	// 解析需求描述 (第7列)
	if len(row) > 6 {
		s.Spec = strings.TrimSpace(row[6])
	}
	if s.Spec == "" {
		return story.Story{}, fmt.Errorf("需求描述不能为空")
	}

	// 解析父需求ID (第8列) - 支持 "@n" 引用格式
	if len(row) > 7 {
		if parentRef := strings.TrimSpace(row[7]); parentRef != "" {
			s.ParentRef = parentRef
			// 如果是纯数字，直接解析为禅道ID
			if id, err := strconv.Atoi(parentRef); err == nil {
				s.ParentID = id
			}
			// "@n" 格式将在导入时解析，ParentID 暂为0
		}
	}

	// 解析来源 (第9列)
	if len(row) > 8 {
		s.Source = strings.TrimSpace(row[8])
	}
	// 解析来源备注 (第10列)
	if len(row) > 9 {
		s.SourceNote = strings.TrimSpace(row[9])
	}
	// 解析预计工时 (第11列)
	if len(row) > 10 {
		if estimate := strings.TrimSpace(row[10]); estimate != "" {
			if est, err := strconv.ParseFloat(estimate, 64); err == nil {
				s.Estimate = est
			}
		}
	}
	// 解析关键词 (第12列)
	if len(row) > 11 {
		s.Keywords = strings.TrimSpace(row[11])
	}
	// 解析验收标准 (第13列)
	if len(row) > 12 {
		s.Verify = strings.TrimSpace(row[12])
	}

	return s, nil
}
