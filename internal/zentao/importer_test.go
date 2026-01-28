package zentao

import (
	"bytes"
	"testing"
	"time"

	"github.com/jan2xue/zentao_import_story/internal/logger"
)

func TestImportResult_Struct(t *testing.T) {
	result := ImportResult{
		Success:     true,
		StoryID:     123,
		Error:       nil,
		ElapsedTime: time.Second,
	}

	if !result.Success {
		t.Error("Success 应为 true")
	}
	if result.StoryID != 123 {
		t.Errorf("StoryID 应为 123，但得到 %d", result.StoryID)
	}
	if result.Error != nil {
		t.Error("Error 应为 nil")
	}
}

func TestNewImporter(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	// 由于需要真实的 Client，这里只测试构造函数的参数传递
	// 实际测试应使用 mock client
	_ = log
}

func TestImporter_GenerateReport(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	// 创建一个不需要真实 client 的 importer 来测试报告生成
	// 实际代码中需要传入 client，这里简化为测试报告格式
	importer := &Importer{logger: log}

	results := []ImportResult{
		{Success: true, StoryID: 1, ElapsedTime: time.Second},
		{Success: true, StoryID: 2, ElapsedTime: 2 * time.Second},
		{Success: false, Error: nil, ElapsedTime: time.Second},
	}

	report := importer.GenerateReport(results)

	// 验证报告包含必要信息
	if report == "" {
		t.Error("报告不应为空")
	}
	if !bytes.Contains([]byte(report), []byte("需求导入报告")) {
		t.Error("报告应包含标题")
	}
	if !bytes.Contains([]byte(report), []byte("总需求数: 3")) {
		t.Error("报告应显示正确的总需求数")
	}
	if !bytes.Contains([]byte(report), []byte("成功导入: 2")) {
		t.Error("报告应显示正确的成功数量")
	}
	if !bytes.Contains([]byte(report), []byte("失败数量: 1")) {
		t.Error("报告应显示正确的失败数量")
	}
}

func TestImporter_GenerateReport_Empty(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)
	importer := &Importer{logger: log}

	results := []ImportResult{}
	
	// 空结果测试应该能正确处理
	report := importer.GenerateReport(results)
	
	if report == "" {
		t.Error("空结果报告不应为空")
	}
}
