package excel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jan2xue/zentao_import_story/pkg/story"
)

func TestNewWriter(t *testing.T) {
	writer, err := NewWriter("test.xlsx")
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}
	defer writer.Close()

	// 测试写入数据
	stories := []story.Story{
		{
			Title:     "测试需求1",
			ProductID: 1,
			Priority:  2,
			Category:  "feature",
			Spec:      "描述1",
		},
		{
			Title:     "测试需求2",
			ProductID: 1,
			Priority:  3,
			Category:  "improve",
			Spec:      "描述2",
		},
	}

	err = writer.WriteStories(stories)
	if err != nil {
		t.Fatalf("写入数据失败: %v", err)
	}

	// 保存到临时目录
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_output.xlsx")
	err = writer.Save(tmpFile)
	if err != nil {
		t.Fatalf("保存文件失败: %v", err)
	}

	// 验证文件是否存在
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("文件未创建")
	}
}

func TestReader_parseRow(t *testing.T) {
	reader := &Reader{}

	tests := []struct {
		name            string
		row             []string
		defaultPriority int
		wantErr         bool
	}{
		{
			name:            "完整数据行",
			row:             []string{"标题", "1", "2", "feature", "描述"},
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "缺少必填字段",
			row:             []string{"标题", "1"},
			defaultPriority: 3,
			wantErr:         true,
		},
		{
			name:            "无效的产品ID",
			row:             []string{"标题", "invalid", "2", "feature", "描述"},
			defaultPriority: 3,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := reader.parseRow(tt.row, tt.defaultPriority)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
