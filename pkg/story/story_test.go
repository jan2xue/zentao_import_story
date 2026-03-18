package story

import (
	"testing"
)

func TestStory_GetTypeString(t *testing.T) {
	tests := []struct {
		name     string
		story    *Story
		expected string
	}{
		{
			name: "Epic类型",
			story: &Story{Type: StoryTypeEpic},
			expected: "业务需求(Epic)",
		},
		{
			name: "Requirement类型",
			story: &Story{Type: StoryTypeRequirement},
			expected: "用户需求(Requirement)",
		},
		{
			name: "Story类型",
			story: &Story{Type: StoryTypeStory},
			expected: "研发需求(Story)",
		},
		{
			name: "默认类型",
			story: &Story{Type: "unknown"},
			expected: "研发需求(Story)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.story.GetTypeString(); got != tt.expected {
				t.Errorf("GetTypeString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStory_Fields(t *testing.T) {
	s := &Story{
		Title:      "测试需求",
		ProductID:  1,
		Priority:   2,
		Category:   "feature",
		Spec:       "需求描述",
		ParentID:   0,
		Source:     "customer",
		SourceNote: "客户反馈",
		Estimate:   8.5,
		Keywords:   "测试,需求",
		Verify:     "验收标准",
	}

	if s.Title != "测试需求" {
		t.Errorf("标题不匹配: got %v, want %v", s.Title, "测试需求")
	}
	if s.ProductID != 1 {
		t.Errorf("产品ID不匹配: got %v, want %v", s.ProductID, 1)
	}
	if s.Priority != 2 {
		t.Errorf("优先级不匹配: got %v, want %v", s.Priority, 2)
	}
}
