package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewLoggerWithWriter(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(&buf)

	logger.Info("测试信息")
	output := buf.String()

	if !strings.Contains(output, "[INFO]") {
		t.Errorf("日志输出应包含 [INFO] 标签，但得到: %s", output)
	}
	if !strings.Contains(output, "测试信息") {
		t.Errorf("日志输出应包含消息内容，但得到: %s", output)
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(&buf)

	logger.Info("这是一条信息: %s", "test")
	output := buf.String()

	if !strings.Contains(output, "[INFO]") {
		t.Error("Info 日志应包含 [INFO] 标签")
	}
	if !strings.Contains(output, "这是一条信息: test") {
		t.Error("Info 日志内容不匹配")
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(&buf)

	logger.Error("这是一个错误: %d", 404)
	output := buf.String()

	if !strings.Contains(output, "[ERROR]") {
		t.Error("Error 日志应包含 [ERROR] 标签")
	}
	if !strings.Contains(output, "这是一个错误: 404") {
		t.Error("Error 日志内容不匹配")
	}
}

func TestLogger_Success(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(&buf)

	logger.Success("操作成功: %s", "导入完成")
	output := buf.String()

	if !strings.Contains(output, "[SUCCESS]") {
		t.Error("Success 日志应包含 [SUCCESS] 标签")
	}
	if !strings.Contains(output, "操作成功: 导入完成") {
		t.Error("Success 日志内容不匹配")
	}
}

func TestLogger_MultipleWriters(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	logger := NewLoggerWithWriter(&buf1, &buf2)

	logger.Info("多writer测试")

	if !strings.Contains(buf1.String(), "多writer测试") {
		t.Error("第一个writer应包含日志内容")
	}
	if !strings.Contains(buf2.String(), "多writer测试") {
		t.Error("第二个writer应包含日志内容")
	}
}
