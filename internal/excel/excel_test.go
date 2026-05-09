package excel

import (
	"testing"
)

func TestReader_parseRow(t *testing.T) {
	reader := &Reader{}

	tests := []struct {
		name            string
		row             []string
		defaultPriority int
		wantErr         bool
	}{
		{
			name:            "完整数据行-story",
			row:             []string{"story", "标题", "1", "2", "feature", "描述"},
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "完整数据行-epic",
			row:             []string{"epic", "业务需求标题", "1", "1", "feature", "描述"},
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "完整数据行-requirement",
			row:             []string{"requirement", "用户需求标题", "1", "2", "improve", "描述"},
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "缺少必填字段",
			row:             []string{"story", "标题", "1"},
			defaultPriority: 3,
			wantErr:         true,
		},
		{
			name:            "无效的产品ID",
			row:             []string{"story", "标题", "invalid", "2", "feature", "描述"},
			defaultPriority: 3,
			wantErr:         true,
		},
		{
			name:            "无效的需求类型",
			row:             []string{"invalid_type", "标题", "1", "2", "feature", "描述"},
			defaultPriority: 3,
			wantErr:         true,
		},
		{
			name:            "父需求@行号引用",
			row:             []string{"story", "子需求", "1", "2", "feature", "描述", "@1"},
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "父需求直接ID",
			row:             []string{"story", "子需求", "1", "2", "feature", "描述", "100"},
			defaultPriority: 3,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := reader.parseRow(tt.row, tt.defaultPriority, 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
