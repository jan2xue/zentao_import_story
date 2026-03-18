// Package story 定义需求领域模型
package story

// StoryType 需求类型
type StoryType string

const (
	StoryTypeEpic        StoryType = "epic"        // 业务需求
	StoryTypeRequirement StoryType = "requirement" // 用户需求
	StoryTypeStory       StoryType = "story"       // 研发需求
)

// Story 表示需求数据模型
type Story struct {
	Type       StoryType // 需求类型：epic/requirement/story
	Title      string    // 标题*
	ProductID  int       // 产品ID*
	Priority   int       // 优先级* (1-4)
	Category   string    // 分类*
	Spec       string    // 需求描述
	ParentID   int       // 父需求ID
	Source     string    // 来源
	SourceNote string    // 来源备注
	Estimate   float64   // 预计工时
	Keywords   string    // 关键词
	Verify     string    // 验收标准
}

// GetTypeString 获取需求类型的字符串表示
func (s *Story) GetTypeString() string {
	switch s.Type {
	case StoryTypeEpic:
		return "业务需求(Epic)"
	case StoryTypeRequirement:
		return "用户需求(Requirement)"
	case StoryTypeStory:
		return "研发需求(Story)"
	default:
		return "研发需求(Story)"
	}
}
