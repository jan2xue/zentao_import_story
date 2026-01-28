package story

import (
	"testing"
)

func TestStory_ToZentaoStory(t *testing.T) {
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

	zentaoStory := s.ToZentaoStory()

	if zentaoStory.Title != s.Title {
		t.Errorf("标题不匹配: got %v, want %v", zentaoStory.Title, s.Title)
	}
	if zentaoStory.Product != s.ProductID {
		t.Errorf("产品ID不匹配: got %v, want %v", zentaoStory.Product, s.ProductID)
	}
	if zentaoStory.Pri != s.Priority {
		t.Errorf("优先级不匹配: got %v, want %v", zentaoStory.Pri, s.Priority)
	}
}
