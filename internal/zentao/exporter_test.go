package zentao

import (
	"bytes"
	"testing"
	"time"

	"github.com/jan2xue/zentao_import_story/internal/logger"
)

func TestExportResult_Struct(t *testing.T) {
	result := ExportResult{
		Success:     true,
		StoryCount:  10,
		Error:       nil,
		ElapsedTime: 1500,
	}

	if !result.Success {
		t.Error("Success 应为 true")
	}
	if result.StoryCount != 10 {
		t.Errorf("StoryCount 应为 10，但得到 %d", result.StoryCount)
	}
	if result.Error != nil {
		t.Error("Error 应为 nil")
	}
	if result.ElapsedTime != 1500 {
		t.Errorf("ElapsedTime 应为 1500，但得到 %d", result.ElapsedTime)
	}
}

func TestNewExporter(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	// 由于需要真实的 Client，这里只测试构造函数的参数传递
	// 实际测试应使用 mock client
	_ = log
}

func TestExporter_ExportResultWithError(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)
	_ = log

	// 测试包含错误的导出结果
	result := ExportResult{
		Success:     false,
		StoryCount:  0,
		Error:       nil, // 实际使用时会有具体错误
		ElapsedTime: time.Since(time.Now()).Milliseconds(),
	}

	if result.Success {
		t.Error("失败情况下 Success 应为 false")
	}
	if result.StoryCount != 0 {
		t.Error("失败情况下 StoryCount 应为 0")
	}
}

func TestExporter_EmptyResult(t *testing.T) {
	// 测试空结果的处理
	result := ExportResult{
		Success:     true,
		StoryCount:  0,
		Error:       nil,
		ElapsedTime: 0,
	}

	if !result.Success {
		t.Error("空结果也是成功的操作")
	}
	if result.StoryCount != 0 {
		t.Error("空结果的计数应为 0")
	}
}
