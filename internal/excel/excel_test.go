package excel

import (
	"reflect"
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
			row:             []string{"story", "1", "10", "标题", "2", "feature", "描述"},
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "完整数据行-epic",
			row:             []string{"epic", "1", "", "业务需求标题", "1", "feature", "描述"},
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "完整数据行-requirement",
			row:             []string{"requirement", "1", "5", "用户需求标题", "2", "improve", "描述"},
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "缺少必填字段",
			row:             []string{"story", "1", "", "标题"},
			defaultPriority: 3,
			wantErr:         true,
		},
		{
			name:            "无效的产品ID",
			row:             []string{"story", "invalid", "", "标题", "2", "feature", "描述"},
			defaultPriority: 3,
			wantErr:         true,
		},
		{
			name:            "无效的需求类型",
			row:             []string{"invalid_type", "1", "", "标题", "2", "feature", "描述"},
			defaultPriority: 3,
			wantErr:         true,
		},
		{
			name:            "父需求@行号引用",
			row:             []string{"story", "1", "", "子需求", "2", "feature", "描述", "@1"},
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "父需求直接ID",
			row:             []string{"story", "1", "", "子需求", "2", "feature", "描述", "100"},
			defaultPriority: 3,
			wantErr:         false,
		},
		{
			name:            "模块ID非数字",
			row:             []string{"story", "1", "abc", "标题", "2", "feature", "描述"},
			defaultPriority: 3,
			wantErr:         true,
		},
		{
			name:            "模块ID为负数",
			row:             []string{"story", "1", "-1", "标题", "2", "feature", "描述"},
			defaultPriority: 3,
			wantErr:         true,
		},
		{
			name:            "模块ID显式为0-合法值",
			row:             []string{"story", "1", "0", "标题", "2", "feature", "描述"},
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

func TestReader_parseRow_Reviewers(t *testing.T) {
	reader := &Reader{}

	tests := []struct {
		name string
		row  []string
		want []string
	}{
		{
			name: "评审人为空-使用配置默认值",
			row:  []string{"story", "1", "", "标题", "2", "feature", "描述", "", "", "", "", "", ""},
			want: nil,
		},
		{
			name: "评审人列缺失-使用配置默认值",
			row:  []string{"story", "1", "", "标题", "2", "feature", "描述"},
			want: nil,
		},
		{
			name: "评审人为空白字符-使用配置默认值",
			row:  []string{"story", "1", "", "标题", "2", "feature", "描述", "", "", "", "", "", "", "   "},
			want: nil,
		},
		{
			name: "单个评审人",
			row:  []string{"story", "1", "", "标题", "2", "feature", "描述", "", "", "", "", "", "", "zhangsan"},
			want: []string{"zhangsan"},
		},
		{
			name: "多个评审人-英文逗号分隔",
			row:  []string{"story", "1", "", "标题", "2", "feature", "描述", "", "", "", "", "", "", "zhangsan,lisi"},
			want: []string{"zhangsan", "lisi"},
		},
		{
			name: "多个评审人-混合分隔符与空格",
			row:  []string{"story", "1", "", "标题", "2", "feature", "描述", "", "", "", "", "", "", " zhangsan、lisi; wangwu "},
			want: []string{"zhangsan", "lisi", "wangwu"},
		},
		{
			name: "评审人重复-自动去重",
			row:  []string{"story", "1", "", "标题", "2", "feature", "描述", "", "", "", "", "", "", "zhangsan,zhangsan"},
			want: []string{"zhangsan"},
		},
		{
			name: "epic类型评审人",
			row:  []string{"epic", "1", "", "标题", "1", "feature", "描述", "", "", "", "", "", "", "admin"},
			want: []string{"admin"},
		},
		{
			name: "requirement类型评审人",
			row:  []string{"requirement", "1", "", "标题", "1", "feature", "描述", "", "", "", "", "", "", "admin,po"},
			want: []string{"admin", "po"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := reader.parseRow(tt.row, 3, 1)
			if err != nil {
				t.Fatalf("parseRow() 返回错误: %v", err)
			}
			if !reflect.DeepEqual(s.Reviewers, tt.want) {
				t.Errorf("Reviewers = %v, 期望 %v", s.Reviewers, tt.want)
			}
		})
	}
}

func TestSplitReviewers(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "空字符串", raw: "", want: nil},
		{name: "仅空格", raw: "   ", want: nil},
		{name: "仅分隔符", raw: ",,；;", want: nil},
		{name: "单值", raw: "admin", want: []string{"admin"}},
		{name: "中文逗号", raw: "admin，tester", want: []string{"admin", "tester"}},
		{name: "全角空格分隔", raw: "admin　tester", want: []string{"admin", "tester"}},
		{name: "制表符与换行分隔", raw: "admin\ttester\nlisi", want: []string{"admin", "tester", "lisi"}},
		{name: "全角空格首尾", raw: "　admin　", want: []string{"admin"}},
		{name: "斜杠不作为分隔符", raw: "corp/user", want: []string{"corp/user"}},
		{name: "竖线不作为分隔符", raw: "a|b", want: []string{"a|b"}},
		{name: "末尾分隔符", raw: "admin,", want: []string{"admin"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitReviewers(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitReviewers(%q) = %v, 期望 %v", tt.raw, got, tt.want)
			}
		})
	}
}
