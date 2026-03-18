package excel

import (
	"testing"
)

func TestReader_parseRow(t *testing.T) {
	reader := &Reader{}

	tests := []struct {
		name            string
		row             []string
		storyType       string
		defaultPriority int
		wantErr         bool
	}{
		{
			name:            "完整数据行-story",
			row:             []string{"标题", "1", "2", "feature", "描述"},
			storyType:       "story",
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "完整数据行-epic",
			row:             []string{"业务需求标题", "1", "1", "feature", "描述"},
			storyType:       "epic",
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "完整数据行-requirement",
			row:             []string{"用户需求标题", "1", "2", "improve", "描述"},
			storyType:       "requirement",
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "缺少必填字段",
			row:             []string{"标题", "1"},
			storyType:       "story",
			defaultPriority: 3,
			wantErr:         true,
		},
		{
			name:            "无效的产品ID",
			row:             []string{"标题", "invalid", "2", "feature", "描述"},
			storyType:       "story",
			defaultPriority: 3,
			wantErr:         true,
		},
		{
			name:            "无效的需求类型",
			row:             []string{"标题", "1", "2", "feature", "描述"},
			storyType:       "invalid_type",
			defaultPriority: 3,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := reader.parseRow(tt.row, tt.defaultPriority, tt.storyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}