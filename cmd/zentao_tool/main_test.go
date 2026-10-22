package main

import (
	"reflect"
	"testing"

	"github.com/jan2xue/zentao_import_story/pkg/story"
)

func TestSummarizeReviewers(t *testing.T) {
	tests := []struct {
		name             string
		stories          []story.Story
		wantReviewers    []string
		wantFallbackRows int
	}{
		{
			name:             "空列表",
			stories:          nil,
			wantReviewers:    []string{},
			wantFallbackRows: 0,
		},
		{
			name: "全部回退配置默认值",
			stories: []story.Story{
				{Title: "a"},
				{Title: "b"},
			},
			wantReviewers:    []string{},
			wantFallbackRows: 2,
		},
		{
			name: "去重并按名称排序",
			stories: []story.Story{
				{Title: "a", Reviewers: []string{"zhangsan"}},
				{Title: "b", Reviewers: []string{"admin", "zhangsan"}},
				{Title: "c"},
			},
			wantReviewers:    []string{"admin", "zhangsan"},
			wantFallbackRows: 1,
		},
		{
			name: "单行多个评审人仅计一次回退",
			stories: []story.Story{
				{Title: "a", Reviewers: []string{"lisi", "wangwu"}},
			},
			wantReviewers:    []string{"lisi", "wangwu"},
			wantFallbackRows: 0,
		},
		{
			name: "混合来源",
			stories: []story.Story{
				{Title: "a", Reviewers: []string{"zhangsan"}},
				{Title: "b"},
				{Title: "c", Reviewers: []string{"admin"}},
				{Title: "d"},
			},
			wantReviewers:    []string{"admin", "zhangsan"},
			wantFallbackRows: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReviewers, gotFallbackRows := summarizeReviewers(tt.stories)
			if !reflect.DeepEqual(gotReviewers, tt.wantReviewers) {
				t.Errorf("评审人清单 = %v, 期望 %v", gotReviewers, tt.wantReviewers)
			}
			if gotFallbackRows != tt.wantFallbackRows {
				t.Errorf("回退行数 = %d, 期望 %d", gotFallbackRows, tt.wantFallbackRows)
			}
		})
	}
}
