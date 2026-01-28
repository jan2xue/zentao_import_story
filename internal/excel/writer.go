// Package excel 处理Excel文件的读写操作
package excel

import (
	"fmt"

	"github.com/xuri/excelize/v2"
	"github.com/jan2xue/zentao_import_story/pkg/story"
)

// Writer 处理Excel文件的写入
type Writer struct {
	file *excelize.File
}

// NewWriter 创建新的Excel写入器
func NewWriter(filePath string) (*Writer, error) {
	// 创建新的Excel文件
	f := excelize.NewFile()
	// 删除默认工作表
	f.DeleteSheet("Sheet1")
	// 创建新的工作表
	f.NewSheet("Stories")
	// 设置标题行
	titleRow := []string{
		"标题",    // A列
		"产品ID",  // B列
		"优先级",   // C列
		"分类",    // D列
		"需求描述",  // E列
		"父需求ID", // F列
		"来源",    // G列
		"来源备注",  // H列
		"预计工时",  // I列
		"关键词",   // J列
		"验收标准",  // K列
	}
	for colIdx, title := range titleRow {
		colName, _ := excelize.CoordinatesToCellName(colIdx+1, 1) // Excel列名 (A, B, C...)
		f.SetCellValue("Stories", colName, title)
	}
	return &Writer{file: f}, nil
}

// Close 关闭Excel文件
func (w *Writer) Close() error {
	return w.file.Close()
}

// Save 保存Excel文件
func (w *Writer) Save(filePath string) error {
	return w.file.SaveAs(filePath)
}

// WriteStories 将需求数据写入Excel文件
func (w *Writer) WriteStories(stories []story.Story) error {
	for rowIndex, s := range stories {
		rowNum := rowIndex + 2 // 从第2行开始写入数据
		// A列: 标题
		w.file.SetCellValue("Stories", fmt.Sprintf("A%d", rowNum), s.Title)
		// B列: 产品ID
		w.file.SetCellValue("Stories", fmt.Sprintf("B%d", rowNum), s.ProductID)
		// C列: 优先级
		w.file.SetCellValue("Stories", fmt.Sprintf("C%d", rowNum), s.Priority)
		// D列: 分类
		w.file.SetCellValue("Stories", fmt.Sprintf("D%d", rowNum), s.Category)
		// E列: 需求描述
		w.file.SetCellValue("Stories", fmt.Sprintf("E%d", rowNum), s.Spec)
		// F列: 父需求ID
		if s.ParentID != 0 {
			w.file.SetCellValue("Stories", fmt.Sprintf("F%d", rowNum), s.ParentID)
		}
		// G列: 来源
		if s.Source != "" {
			w.file.SetCellValue("Stories", fmt.Sprintf("G%d", rowNum), s.Source)
		}
		// H列: 来源备注
		if s.SourceNote != "" {
			w.file.SetCellValue("Stories", fmt.Sprintf("H%d", rowNum), s.SourceNote)
		}
		// I列: 预计工时
		if s.Estimate != 0 {
			w.file.SetCellValue("Stories", fmt.Sprintf("I%d", rowNum), s.Estimate)
		}
		// J列: 关键词
		if s.Keywords != "" {
			w.file.SetCellValue("Stories", fmt.Sprintf("J%d", rowNum), s.Keywords)
		}
		// K列: 验收标准
		if s.Verify != "" {
			w.file.SetCellValue("Stories", fmt.Sprintf("K%d", rowNum), s.Verify)
		}
	}
	return nil
}
