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

// ReadStories 读取所有需求数据
func (r *Reader) ReadStories(defaultPriority int) ([]story.Story, error) {
	// 获取第一个工作表
	sheets := r.file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件中没有工作表")
	}

	// 读取所有行
	rows, err := r.file.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}

	if len(rows) < 2 { // 至少需要标题行和一行数据
		return nil, fmt.Errorf("Excel文件中没有数据")
	}

	// 跳过标题行，处理数据行
	var stories []story.Story
	for i, row := range rows[1:] {
		s, err := r.parseRow(row, defaultPriority)
		if err != nil {
			return nil, fmt.Errorf("第%d行数据解析失败: %w", i+2, err)
		}
		stories = append(stories, s)
	}

	return stories, nil
}

// parseRow 解析Excel行数据为需求结构
func (r *Reader) parseRow(row []string, defaultPriority int) (story.Story, error) {
	if len(row) < 4 { // 检查必填字段
		return story.Story{}, fmt.Errorf("行数据不完整，缺少必填字段")
	}

	// 解析产品ID
	productID, err := strconv.Atoi(strings.TrimSpace(row[1]))
	if err != nil {
		return story.Story{}, fmt.Errorf("产品ID必须是数字: %w", err)
	}

	// 解析优先级
	priority := defaultPriority
	if len(row) > 2 && strings.TrimSpace(row[2]) != "" {
		p, err := strconv.Atoi(strings.TrimSpace(row[2]))
		if err != nil || p < 1 || p > 4 {
			return story.Story{}, fmt.Errorf("优先级必须是1-4之间的数字")
		}
		priority = p
	}

	// 创建需求对象
	s := story.Story{
		Title:     strings.TrimSpace(row[0]),
		ProductID: productID,
		Priority:  priority,
		Category:  strings.TrimSpace(row[3]),
	}

	// 检查必填字段
	if s.Title == "" {
		return story.Story{}, fmt.Errorf("标题不能为空")
	}
	if s.Category == "" {
		return story.Story{}, fmt.Errorf("分类不能为空")
	}
	if len(row) > 4 {
		s.Spec = strings.TrimSpace(row[4])
	}
	if s.Spec == "" {
		return story.Story{}, fmt.Errorf("需求描述不能为空")
	}

	// 设置可选字段
	if len(row) > 5 {
		if parentID := strings.TrimSpace(row[5]); parentID != "" {
			if id, err := strconv.Atoi(parentID); err == nil {
				s.ParentID = id
			}
		}
	}
	if len(row) > 6 {
		s.Source = strings.TrimSpace(row[6])
	}
	if len(row) > 7 {
		s.SourceNote = strings.TrimSpace(row[7])
	}
	if len(row) > 8 {
		if estimate := strings.TrimSpace(row[8]); estimate != "" {
			if est, err := strconv.ParseFloat(estimate, 64); err == nil {
				s.Estimate = est
			}
		}
	}
	if len(row) > 9 {
		s.Keywords = strings.TrimSpace(row[9])
	}
	if len(row) > 10 {
		s.Verify = strings.TrimSpace(row[10])
	}

	return s, nil
}
