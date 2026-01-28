// Package story 定义需求领域模型
package story

import (
	"github.com/easysoft/go-zentao/v21/zentao"
)

// Story 表示需求数据模型
type Story struct {
	Title      string  // 标题*
	ProductID  int     // 产品ID*
	Priority   int     // 优先级* (1-4)
	Category   string  // 分类*
	Spec       string  // 需求描述
	ParentID   int     // 父需求ID
	Source     string  // 来源
	SourceNote string  // 来源备注
	Estimate   float64 // 预计工时
	Keywords   string  // 关键词
	Verify     string  // 验收标准
}

// ToZentaoStory 将Story转换为禅道需求创建元数据
func (s *Story) ToZentaoStory() *zentao.StoriesCreateMeta {
	return &zentao.StoriesCreateMeta{
		Product: s.ProductID,
		StoriesMeta: zentao.StoriesMeta{
			Title:  s.Title,
			Spec:   s.Spec,
			Verify: s.Verify,
		},
		StoriesExtMeta: zentao.StoriesExtMeta{
			Pri:        s.Priority,
			Category:   zentao.StoriesCategory(s.Category),
			Parent:     s.ParentID,
			Source:     zentao.StoriesSource(s.Source),
			SourceNote: s.SourceNote,
			Estimate:   s.Estimate,
			Keywords:   s.Keywords,
		},
	}
}
