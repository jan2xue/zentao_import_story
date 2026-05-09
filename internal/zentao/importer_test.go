package zentao

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/imroc/req/v3"
	"github.com/jan2xue/zentao_import_story/internal/logger"
	"github.com/jan2xue/zentao_import_story/pkg/story"
)

func TestImporter_ImportStory_Success(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	mockStory := &mockStoryService{
		createFn: func(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
			if req.Title != "测试需求" {
				t.Errorf("期望标题='测试需求', 得到 '%s'", req.Title)
			}
			if req.Pri != 2 {
				t.Errorf("期望优先级=2, 得到 %d", req.Pri)
			}
			return &StoryCreateResponse{Status: "success", ID: 42}, nil, nil
		},
	}

	cfg := &mockConfig{reviewer: "tester"}
	importer := NewImporterWithMocks(log, nil, nil, mockStory, cfg)

	s := &story.Story{
		Type:      story.StoryTypeStory,
		Title:     "测试需求",
		ProductID: 1,
		Priority:  2,
		Category:  "feature",
		Spec:      "desc",
	}

	result := importer.ImportStory(s)

	if !result.Success {
		t.Fatal("希望导入成功")
	}
	if result.StoryID != 42 {
		t.Errorf("期望 StoryID=42, 得到 %d", result.StoryID)
	}
}

func TestImporter_ImportStory_APIFailure(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	mockStory := &mockStoryService{
		createFn: func(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
			return &StoryCreateResponse{Status: "fail"}, nil, nil
		},
	}

	cfg := &mockConfig{reviewer: "tester"}
	importer := NewImporterWithMocks(log, nil, nil, mockStory, cfg)

	s := &story.Story{
		Type:      story.StoryTypeStory,
		Title:     "失败需求",
		ProductID: 1,
		Priority:  3,
		Category:  "feature",
		Spec:      "desc",
	}

	result := importer.ImportStory(s)

	if result.Success {
		t.Fatal("希望导入失败")
	}
	if result.Error == nil {
		t.Fatal("期望有错误信息")
	}
}

func TestImporter_ImportStory_RequirementModuleMissing(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	cfg := &mockConfig{module: 0, reviewer: ""} // 模块ID=0，应失败
	importer := NewImporterWithMocks(log, nil, nil, nil, cfg)

	s := &story.Story{
		Type:      story.StoryTypeRequirement,
		Title:     "用户需求",
		ProductID: 1,
		Priority:  3,
		Category:  "feature",
		Spec:      "desc",
	}

	result := importer.ImportStory(s)

	if result.Success {
		t.Fatal("模块ID缺失时导入应失败")
	}
	if result.Error == nil {
		t.Fatal("期望有错误信息")
	}
}

func TestImporter_ImportStories_Batch(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	var callCount int32
	mockStory := &mockStoryService{
		createFn: func(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
			atomic.AddInt32(&callCount, 1)
			return &StoryCreateResponse{Status: "success", ID: 100 + int(req.Pri)}, nil, nil
		},
	}

	cfg := &mockConfig{reviewer: "tester"}
	importer := NewImporterWithMocks(log, nil, nil, mockStory, cfg)

	stories := []story.Story{
		{Type: story.StoryTypeStory, Title: "需求A", ProductID: 1, Priority: 1, Category: "feature", Spec: "a"},
		{Type: story.StoryTypeStory, Title: "需求B", ProductID: 1, Priority: 2, Category: "feature", Spec: "b"},
		{Type: story.StoryTypeStory, Title: "需求C", ProductID: 1, Priority: 3, Category: "feature", Spec: "c"},
	}

	results := importer.ImportStories(stories)

	if len(results) != 3 {
		t.Fatalf("期望3个结果, 得到 %d", len(results))
	}
	if int(callCount) != 3 {
		t.Errorf("期望3次create调用, 得到 %d", callCount)
	}
	for i, r := range results {
		if !r.Success {
			t.Errorf("需求 #%d 应成功", i+1)
		}
	}
}

func TestImporter_GenerateReport(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)
	importer := &Importer{logger: log}

	results := []ImportResult{
		{Success: true, StoryID: 1, ElapsedTime: time.Second},
		{Success: true, StoryID: 2, ElapsedTime: 2 * time.Second},
		{Success: false, Error: nil, ElapsedTime: time.Second},
	}

	report := importer.GenerateReport(results)

	if report == "" {
		t.Error("报告不应为空")
	}
	if !bytes.Contains([]byte(report), []byte("需求导入报告")) {
		t.Error("报告应包含标题")
	}
	if !bytes.Contains([]byte(report), []byte("总需求数: 3")) {
		t.Error("报告应显示正确的总需求数")
	}
	if !bytes.Contains([]byte(report), []byte("成功导入: 2")) {
		t.Error("报告应显示正确的成功数量")
	}
	if !bytes.Contains([]byte(report), []byte("失败数量: 1")) {
		t.Error("报告应显示正确的失败数量")
	}
}

func TestImporter_GenerateReport_Empty(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)
	importer := &Importer{logger: log}

	results := []ImportResult{}

	report := importer.GenerateReport(results)

	if report == "" {
		t.Error("空结果报告不应为空")
	}
}

func TestDeleter_DeleteStory_Success(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	mockStory := &mockStoryService{
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			if id != 42 {
				t.Errorf("期望删除ID=42, 得到 %d", id)
			}
			return map[string]interface{}{"status": "success"}, nil, nil
		},
	}

	deleter := NewDeleterWithMocks(log, nil, nil, mockStory)

	result := deleter.DeleteStory(42, story.StoryTypeStory)

	if !result.Success {
		t.Fatal("希望删除成功")
	}
}

func TestDeleter_DeleteStory_Failure(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	mockStory := &mockStoryService{
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, fmt.Errorf("模拟删除失败")
		},
	}

	deleter := NewDeleterWithMocks(log, nil, nil, mockStory)

	result := deleter.DeleteStory(999, story.StoryTypeStory)

	if result.Success {
		t.Fatal("希望删除失败")
	}
}

func TestDeleter_GenerateDeleteReport(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)
	deleter := &Deleter{logger: log}

	results := []DeleteResult{
		{Success: true, StoryID: 1, ElapsedTime: time.Second},
		{Success: false, Error: nil, ElapsedTime: time.Second},
	}

	report := deleter.GenerateDeleteReport(results)

	if report == "" {
		t.Error("报告不应为空")
	}
	if !bytes.Contains([]byte(report), []byte("需求删除报告")) {
		t.Error("报告应包含标题")
	}
}
