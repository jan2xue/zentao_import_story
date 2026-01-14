package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/easysoft/go-zentao/v21/zentao"
	"github.com/xuri/excelize/v2"
)

// ExcelStory 表示从Excel读取的需求数据
type ExcelStory struct {
	Title      string  // 标题*
	ProductID  int     // 产品ID*
	Priority   int     // 优先级* (1-4)
	Category   string  // 分类*
	Spec       string  // 需求描述
	ParentID   int     // 父需求ID
	Source     string  // 来源
	SourceNote string  // 来源备注
	Estimate   float64 // 预计工时
	Keywords   string  // 关键词
	Verify     string  // 验收标准
}

// ExcelReader 处理Excel文件的读取和验证
type ExcelReader struct {
	file *excelize.File
}

// ExcelWriter 处理Excel文件的写入
type ExcelWriter struct {
	file *excelize.File
}

// NewExcelWriter 创建新的Excel写入器
func NewExcelWriter(filePath string) (*ExcelWriter, error) {
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
	return &ExcelWriter{file: f}, nil
}

// Close 关闭Excel文件
func (w *ExcelWriter) Close() error {
	return w.file.Close()
}

// Save 保存Excel文件
func (w *ExcelWriter) Save(filePath string) error {
	return w.file.SaveAs(filePath)
}

// WriteStories 将需求数据写入Excel文件
func (w *ExcelWriter) WriteStories(stories []ExcelStory) error {
	for rowIndex, story := range stories {
		rowNum := rowIndex + 2 // 从第2行开始写入数据
		// A列: 标题
		w.file.SetCellValue("Stories", fmt.Sprintf("A%d", rowNum), story.Title)
		// B列: 产品ID
		w.file.SetCellValue("Stories", fmt.Sprintf("B%d", rowNum), story.ProductID)
		// C列: 优先级
		w.file.SetCellValue("Stories", fmt.Sprintf("C%d", rowNum), story.Priority)
		// D列: 分类
		w.file.SetCellValue("Stories", fmt.Sprintf("D%d", rowNum), story.Category)
		// E列: 需求描述
		w.file.SetCellValue("Stories", fmt.Sprintf("E%d", rowNum), story.Spec)
		// F列: 父需求ID
		if story.ParentID != 0 {
			w.file.SetCellValue("Stories", fmt.Sprintf("F%d", rowNum), story.ParentID)
		}
		// G列: 来源
		if story.Source != "" {
			w.file.SetCellValue("Stories", fmt.Sprintf("G%d", rowNum), story.Source)
		}
		// H列: 来源备注
		if story.SourceNote != "" {
			w.file.SetCellValue("Stories", fmt.Sprintf("H%d", rowNum), story.SourceNote)
		}
		// I列: 预计工时
		if story.Estimate != 0 {
			w.file.SetCellValue("Stories", fmt.Sprintf("I%d", rowNum), story.Estimate)
		}
		// J列: 关键词
		if story.Keywords != "" {
			w.file.SetCellValue("Stories", fmt.Sprintf("J%d", rowNum), story.Keywords)
		}
		// K列: 验收标准
		if story.Verify != "" {
			w.file.SetCellValue("Stories", fmt.Sprintf("K%d", rowNum), story.Verify)
		}
	}
	return nil
}

// NewExcelReader 创建新的Excel读取器
func NewExcelReader(filePath string) (*ExcelReader, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %w", err)
	}
	return &ExcelReader{file: f}, nil
}

// Close 关闭Excel文件
func (r *ExcelReader) Close() error {
	return r.file.Close()
}

// ReadStories 读取所有需求数据
func (r *ExcelReader) ReadStories(defaultPriority int) ([]ExcelStory, error) {
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
	var stories []ExcelStory
	for i, row := range rows[1:] {
		story, err := r.parseRow(row, defaultPriority)
		if err != nil {
			return nil, fmt.Errorf("第%d行数据解析失败: %w", i+2, err)
		}
		stories = append(stories, story)
	}

	return stories, nil
}

// parseRow 解析Excel行数据为需求结构
func (r *ExcelReader) parseRow(row []string, defaultPriority int) (ExcelStory, error) {
	if len(row) < 4 { // 检查必填字段
		return ExcelStory{}, fmt.Errorf("行数据不完整，缺少必填字段")
	}

	// 解析产品ID
	productID, err := strconv.Atoi(strings.TrimSpace(row[1]))
	if err != nil {
		return ExcelStory{}, fmt.Errorf("产品ID必须是数字: %w", err)
	}

	// 解析优先级
	priority := defaultPriority
	if len(row) > 2 && strings.TrimSpace(row[2]) != "" {
		p, err := strconv.Atoi(strings.TrimSpace(row[2]))
		if err != nil || p < 1 || p > 4 {
			return ExcelStory{}, fmt.Errorf("优先级必须是1-4之间的数字")
		}
		priority = p
	}

	// 创建需求对象
	story := ExcelStory{
		Title:     strings.TrimSpace(row[0]),
		ProductID: productID,
		Priority:  priority,
		Category:  strings.TrimSpace(row[3]),
	}

	// 检查必填字段
	if story.Title == "" {
		return ExcelStory{}, fmt.Errorf("标题不能为空")
	}
	if story.Category == "" {
		return ExcelStory{}, fmt.Errorf("分类不能为空")
	}
	if len(row) > 4 {
		story.Spec = strings.TrimSpace(row[4])
	}
	if story.Spec == "" {
		return ExcelStory{}, fmt.Errorf("需求描述不能为空")
	}

	// 设置可选字段

	if len(row) > 5 {
		if parentID := strings.TrimSpace(row[5]); parentID != "" {
			if id, err := strconv.Atoi(parentID); err == nil {
				story.ParentID = id
			}
		}
	}
	if len(row) > 6 {
		story.Source = strings.TrimSpace(row[6])
	}
	if len(row) > 7 {
		story.SourceNote = strings.TrimSpace(row[7])
	}
	if len(row) > 8 {
		if estimate := strings.TrimSpace(row[8]); estimate != "" {
			if est, err := strconv.ParseFloat(estimate, 64); err == nil {
				story.Estimate = est
			}
		}
	}
	if len(row) > 9 {
		story.Keywords = strings.TrimSpace(row[9])
	}
	if len(row) > 10 {
		story.Verify = strings.TrimSpace(row[10])
	}

	return story, nil
}

// ToZentaoStory 将Excel需求转换为禅道需求
func (s *ExcelStory) ToZentaoStory() *zentao.StoriesCreateMeta {
	return &zentao.StoriesCreateMeta{
		Product: s.ProductID,
		StoriesMeta: zentao.StoriesMeta{
			Title:  s.Title,  // 需求标题	必填
			Spec:   s.Spec,   // 需求描述
			Verify: s.Verify, // 验证方法
		},
		StoriesExtMeta: zentao.StoriesExtMeta{
			Pri:        s.Priority,                         // 优先级 1-4	必填
			Category:   zentao.StoriesCategory(s.Category), // 分类	必填
			Parent:     s.ParentID,                         // 父需求ID
			Source:     zentao.StoriesSource(s.Source),     // 来源
			SourceNote: s.SourceNote,                       // 来源备注
			Estimate:   s.Estimate,                         // 预计工时
			Keywords:   s.Keywords,                         // 关键词
		},
	}
}


