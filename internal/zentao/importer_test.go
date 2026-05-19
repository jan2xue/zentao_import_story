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

func TestImporter_ImportStory_RequirementModuleFallback(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	// 测试1: Excel未填写模块ID(-1)，配置文件有默认值5，应使用配置值
	t.Run("ExcelEmpty_UseConfigDefault", func(t *testing.T) {
		cfg := &mockConfig{module: 5, reviewer: "tester"}
		importer := NewImporterWithMocks(log, nil, nil, nil, cfg)
		// Module=-1 表示Excel未填写
		got := importer.resolveModule(-1)
		if got != 5 {
			t.Fatalf("期望模块ID=5，实际=%d", got)
		}
	})

	// 测试2: Excel显式填写0，应直接使用0（0是合法值）
	t.Run("ExcelExplicitZero_UseZero", func(t *testing.T) {
		cfg := &mockConfig{module: 5, reviewer: "tester"}
		importer := NewImporterWithMocks(log, nil, nil, nil, cfg)
		got := importer.resolveModule(0)
		if got != 0 {
			t.Fatalf("期望模块ID=0，实际=%d", got)
		}
	})

	// 测试3: Excel填写正数，应直接使用
	t.Run("ExcelPositive_UseExcelValue", func(t *testing.T) {
		cfg := &mockConfig{module: 5, reviewer: "tester"}
		importer := NewImporterWithMocks(log, nil, nil, nil, cfg)
		got := importer.resolveModule(10)
		if got != 10 {
			t.Fatalf("期望模块ID=10，实际=%d", got)
		}
	})

	// 测试4: Excel未填写且配置也为0，应使用0
	t.Run("ExcelEmpty_ConfigZero_UseZero", func(t *testing.T) {
		cfg := &mockConfig{module: 0, reviewer: "tester"}
		importer := NewImporterWithMocks(log, nil, nil, nil, cfg)
		got := importer.resolveModule(-1)
		if got != 0 {
			t.Fatalf("期望模块ID=0，实际=%d", got)
		}
	})
}

func TestImporter_ImportStories_EpicRequirementResolveID(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	// Epic创建API不返回ID（ID=0），需通过列表查询获取
	mockEpic := &mockEpicService{
		createFn: func(req EpicCreateRequest) (*EpicCreateResponse, *req.Response, error) {
			return &EpicCreateResponse{Status: "success", ID: 0}, nil, nil // 模拟API不返回ID
		},
		// Epic列表包含关联的Requirement和Story
		listFn: func(productID int) ([]EpicListItem, error) {
			return []EpicListItem{
				{ID: 601, Title: "用户需求B", Product: 1}, // 实际是Requirement，Epic API也会返回
				{ID: 501, Title: "业务需求A", Product: 1}, // 真正的Epic
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
	}

	// Requirement创建API不返回ID，需通过列表查询获取
	mockReq := &mockReqService{
		createFn: func(req RequirementCreateRequest) (*RequirementCreateResponse, *req.Response, error) {
			// 验证ParentID已正确解析为501（Epic的实际ID）
			if req.Parent != 501 {
				t.Errorf("期望Requirement的Parent=501, 得到 %d", req.Parent)
			}
			return &RequirementCreateResponse{Status: "success", ID: 0}, nil, nil
		},
		// Requirement列表包含关联的Story
		listFn: func(productID int) ([]RequirementListItem, error) {
			return []RequirementListItem{
				{ID: 601, Title: "用户需求B", Product: 1}, // 真正的Requirement
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
	}

	// Story创建API返回ID
	mockStorySvc := &mockStoryService{
		createFn: func(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
			// 验证ParentID已正确解析为601（Requirement的实际ID）
			if req.Parent != 601 {
				t.Errorf("期望Story的Parent=601, 得到 %d", req.Parent)
			}
			return &StoryCreateResponse{Status: "success", ID: 701}, nil, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		listFn: func(productID int) ([]StoryListItem, error) {
			return nil, nil
		},
	}

	cfg := &mockConfig{module: 1, reviewer: "tester"}
	importer := NewImporterWithMocks(log, mockEpic, mockReq, mockStorySvc, cfg)

	stories := []story.Story{
		{Type: story.StoryTypeEpic, Title: "业务需求A", ProductID: 1, Priority: 3, Category: "feature", RowIndex: 2},
		{Type: story.StoryTypeRequirement, Title: "用户需求B", ProductID: 1, Priority: 3, Category: "feature", ParentRef: "@2", RowIndex: 3},
		{Type: story.StoryTypeStory, Title: "研发需求C", ProductID: 1, Priority: 3, Category: "feature", ParentRef: "@3", RowIndex: 4},
	}

	results := importer.ImportStories(stories)

	if len(results) != 3 {
		t.Fatalf("期望3个结果, 得到 %d", len(results))
	}

	// Epic: 创建返回ID=0, 通过列表查询去重后得到ID=501
	// Epic列表中ID=601("用户需求B")在Story列表中未出现，但ID=601标题是"用户需求B"不匹配"业务需求A"
	// 所以匹配到ID=501("业务需求A")
	if !results[0].Success {
		t.Fatal("Epic应导入成功")
	}
	if results[0].StoryID != 501 {
		t.Errorf("Epic的StoryID应为501(通过列表查询去重), 得到 %d", results[0].StoryID)
	}

	// Requirement: 创建返回ID=0, 通过列表查询得到ID=601
	if !results[1].Success {
		t.Fatal("Requirement应导入成功")
	}
	if results[1].StoryID != 601 {
		t.Errorf("Requirement的StoryID应为601(通过列表查询), 得到 %d", results[1].StoryID)
	}

	// Story: 创建返回ID=701
	if !results[2].Success {
		t.Fatal("Story应导入成功")
	}
	if results[2].StoryID != 701 {
		t.Errorf("Story的StoryID应为701, 得到 %d", results[2].StoryID)
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

func TestDeleter_FetchByFilter_WithTitle(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	// Story列表：ID=1,4 是真正的研发需求
	mockEpic := &mockEpicService{
		listFn: func(productID int) ([]EpicListItem, error) {
			return []EpicListItem{
				{ID: 1, Title: "目标Epic测试", Product: 78, OpenedBy: "zhangsan"},  // 实际是Story，Epic API也会返回
				{ID: 2, Title: "目标Epic", Product: 78, OpenedBy: "lisi"},          // 真正的Epic
				{ID: 4, Title: "目标Epic测试数据", Product: 78, OpenedBy: "wangwu"},  // 实际是Story，Epic API也会返回
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req EpicCreateRequest) (*EpicCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	// Requirement列表：ID=1 也在Story列表中（重复），ID=3 是真正的用户需求
	mockReq := &mockReqService{
		listFn: func(productID int) ([]RequirementListItem, error) {
			return []RequirementListItem{
				{ID: 1, Title: "目标Epic测试", Product: 78, OpenedBy: "zhangsan"},  // 实际是Story，Req API也会返回
				{ID: 3, Title: "目标Req测试", Product: 78, OpenedBy: "zhangsan"},   // 真正的Requirement
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req RequirementCreateRequest) (*RequirementCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	mockStorySvc := &mockStoryService{
		listFn: func(productID int) ([]StoryListItem, error) {
			return []StoryListItem{
				{ID: 1, Title: "目标Epic测试", Product: 78, OpenedBy: "zhangsan"},
				{ID: 4, Title: "目标Epic测试数据", Product: 78, OpenedBy: "wangwu"},
				{ID: 5, Title: "无关Story", Product: 78, OpenedBy: "zhangsan"},
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	deleter := NewDeleterWithMocks(log, mockEpic, mockReq, mockStorySvc)

	// 按标题部分匹配"目标Epic"
	items := deleter.FetchByFilter(DeleteFilter{ProductID: 78, Title: "目标Epic"})

	// 去重后应匹配：
	// Story: ID=1("目标Epic测试"), ID=4("目标Epic测试数据")
	// Requirement: ID=3("目标Req测试") - 标题不包含"目标Epic"，不匹配
	// Epic: ID=2("目标Epic") - ID=1和ID=4已在Story中，跳过
	// 最终匹配：ID=1(Story), ID=4(Story), ID=2(Epic)
	if len(items) != 3 {
		t.Fatalf("期望3个匹配项, 得到 %d", len(items))
	}

	typeResult := map[int]story.StoryType{}
	for _, item := range items {
		typeResult[item.ID] = item.Type
	}
	if typeResult[1] != story.StoryTypeStory {
		t.Errorf("ID=1 应为Story类型, 得到 %v", typeResult[1])
	}
	if typeResult[4] != story.StoryTypeStory {
		t.Errorf("ID=4 应为Story类型, 得到 %v", typeResult[4])
	}
	if typeResult[2] != story.StoryTypeEpic {
		t.Errorf("ID=2 应为Epic类型, 得到 %v", typeResult[2])
	}
}

func TestDeleter_FetchByFilter_WithOpenedBy(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	// Epic列表包含所有（含关联的Story和Requirement）
	mockEpic := &mockEpicService{
		listFn: func(productID int) ([]EpicListItem, error) {
			return []EpicListItem{
				{ID: 1, Title: "Epic1", Product: 78, OpenedBy: "zhangsan"},
				{ID: 2, Title: "Epic2", Product: 78, OpenedBy: "lisi"},
				{ID: 3, Title: "Req1", Product: 78, OpenedBy: "zhangsan"}, // 实际是Requirement
				{ID: 4, Title: "Story1", Product: 78, OpenedBy: "wangwu"}, // 实际是Story
				{ID: 5, Title: "Story2", Product: 78, OpenedBy: "zhangsan"}, // 实际是Story
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req EpicCreateRequest) (*EpicCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	// Requirement列表包含关联的Story
	mockReq := &mockReqService{
		listFn: func(productID int) ([]RequirementListItem, error) {
			return []RequirementListItem{
				{ID: 3, Title: "Req1", Product: 78, OpenedBy: "zhangsan"},
				{ID: 4, Title: "Story1", Product: 78, OpenedBy: "wangwu"}, // 实际是Story
				{ID: 5, Title: "Story2", Product: 78, OpenedBy: "zhangsan"}, // 实际是Story
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req RequirementCreateRequest) (*RequirementCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	mockStorySvc := &mockStoryService{
		listFn: func(productID int) ([]StoryListItem, error) {
			return []StoryListItem{
				{ID: 4, Title: "Story1", Product: 78, OpenedBy: "wangwu"},
				{ID: 5, Title: "Story2", Product: 78, OpenedBy: "zhangsan"},
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	deleter := NewDeleterWithMocks(log, mockEpic, mockReq, mockStorySvc)

	// 按创建者"zhangsan"筛选
	items := deleter.FetchByFilter(DeleteFilter{ProductID: 78, OpenedBy: "zhangsan"})

	// 去重后匹配"zhangsan"：
	// Story: ID=5("Story2", zhangsan) ✓; ID=4("Story1", wangwu) ✗
	// Requirement: ID=3("Req1", zhangsan) ✓ (ID=4,5已在Story中)
	// Epic: ID=1("Epic1", zhangsan) ✓ (ID=3,4,5已在前两层); ID=2("Epic2", lisi) ✗
	// 最终：ID=5(Story), ID=3(Requirement), ID=1(Epic)
	if len(items) != 3 {
		t.Fatalf("期望3个匹配项, 得到 %d", len(items))
	}

	typeResult := map[int]story.StoryType{}
	for _, item := range items {
		typeResult[item.ID] = item.Type
	}
	if typeResult[5] != story.StoryTypeStory {
		t.Errorf("ID=5 应为Story类型, 得到 %v", typeResult[5])
	}
	if typeResult[3] != story.StoryTypeRequirement {
		t.Errorf("ID=3 应为Requirement类型, 得到 %v", typeResult[3])
	}
	if typeResult[1] != story.StoryTypeEpic {
		t.Errorf("ID=1 应为Epic类型, 得到 %v", typeResult[1])
	}
}

func TestDeleter_FetchByFilter_WithTitleAndOpenedBy(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	// Epic列表包含所有
	mockEpic := &mockEpicService{
		listFn: func(productID int) ([]EpicListItem, error) {
			return []EpicListItem{
				{ID: 1, Title: "测试需求A", Product: 78, OpenedBy: "zhangsan"},
				{ID: 2, Title: "测试需求B", Product: 78, OpenedBy: "lisi"},
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req EpicCreateRequest) (*EpicCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	mockReq := &mockReqService{
		listFn: func(productID int) ([]RequirementListItem, error) {
			return nil, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req RequirementCreateRequest) (*RequirementCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	mockStorySvc := &mockStoryService{
		listFn: func(productID int) ([]StoryListItem, error) {
			return nil, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	deleter := NewDeleterWithMocks(log, mockEpic, mockReq, mockStorySvc)

	// 按标题部分匹配"测试需求" + 创建者"zhangsan"
	items := deleter.FetchByFilter(DeleteFilter{ProductID: 78, Title: "测试需求", OpenedBy: "zhangsan"})

	// 去重后：Story列表为空，Requirement列表为空
	// Epic: ID=1("测试需求A", zhangsan) ✓; ID=2("测试需求B", lisi) ✗
	if len(items) != 1 {
		t.Fatalf("期望1个匹配项, 得到 %d", len(items))
	}
	if items[0].ID != 1 {
		t.Errorf("期望ID=1, 得到 %d", items[0].ID)
	}
	if items[0].Type != story.StoryTypeEpic {
		t.Errorf("ID=1 应为Epic类型, 得到 %v", items[0].Type)
	}
}

func TestDeleter_FetchByFilter_NoFilter(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	// Epic列表包含所有（含关联的Requirement和Story）
	mockEpic := &mockEpicService{
		listFn: func(productID int) ([]EpicListItem, error) {
			return []EpicListItem{
				{ID: 1, Title: "Epic1", Product: 78, OpenedBy: "user1"},
				{ID: 2, Title: "Epic2", Product: 78, OpenedBy: "user2"},
				{ID: 3, Title: "Story1", Product: 78, OpenedBy: "user1"}, // 实际是Story
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req EpicCreateRequest) (*EpicCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	mockReq := &mockReqService{
		listFn: func(productID int) ([]RequirementListItem, error) {
			return []RequirementListItem{
				{ID: 3, Title: "Story1", Product: 78, OpenedBy: "user1"}, // 实际是Story
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req RequirementCreateRequest) (*RequirementCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	mockStorySvc := &mockStoryService{
		listFn: func(productID int) ([]StoryListItem, error) {
			return []StoryListItem{
				{ID: 3, Title: "Story1", Product: 78, OpenedBy: "user1"},
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	deleter := NewDeleterWithMocks(log, mockEpic, mockReq, mockStorySvc)

	// 不指定标题和创建者，应匹配产品下所有需求
	items := deleter.FetchByFilter(DeleteFilter{ProductID: 78})

	// 去重后：Story: ID=3; Requirement: 空(去重); Epic: ID=1, ID=2 (ID=3已在Story中)
	// 共3条：ID=3(Story), ID=1(Epic), ID=2(Epic)
	if len(items) != 3 {
		t.Fatalf("期望3个匹配项, 得到 %d", len(items))
	}
}

func TestDeleter_FetchByFilter_NoMatch(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	mockEpic := &mockEpicService{
		listFn: func(productID int) ([]EpicListItem, error) {
			return []EpicListItem{
				{ID: 1, Title: "Epic1", Product: 78, OpenedBy: "user1"},
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req EpicCreateRequest) (*EpicCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	mockReq := &mockReqService{
		listFn: func(productID int) ([]RequirementListItem, error) {
			return nil, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req RequirementCreateRequest) (*RequirementCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	mockStorySvc := &mockStoryService{
		listFn: func(productID int) ([]StoryListItem, error) {
			return nil, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	deleter := NewDeleterWithMocks(log, mockEpic, mockReq, mockStorySvc)

	// 标题部分匹配不命中
	items := deleter.FetchByFilter(DeleteFilter{ProductID: 78, Title: "不存在的"})

	if len(items) != 0 {
		t.Fatalf("期望0个匹配项, 得到 %d", len(items))
	}
}

// TestDeleter_FetchByFilter_Deduplication 专门测试去重逻辑
// 模拟真实场景：Epic API返回关联的Story和Requirement，导致同一ID在多个API中出现
func TestDeleter_FetchByFilter_Deduplication(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	// 真实数据：2个Story(ID=100,101), 1个Requirement(ID=200), 1个Epic(ID=300)
	// Epic API会返回所有4条（包含关联的Story和Requirement）
	mockEpic := &mockEpicService{
		listFn: func(productID int) ([]EpicListItem, error) {
			return []EpicListItem{
				{ID: 100, Title: "Story-A", Product: 1, OpenedBy: "dev1"},
				{ID: 101, Title: "Story-B", Product: 1, OpenedBy: "dev2"},
				{ID: 200, Title: "Req-C", Product: 1, OpenedBy: "pm1"},
				{ID: 300, Title: "Epic-D", Product: 1, OpenedBy: "pm2"},
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req EpicCreateRequest) (*EpicCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	// Requirement API返回关联的Story和自身
	mockReq := &mockReqService{
		listFn: func(productID int) ([]RequirementListItem, error) {
			return []RequirementListItem{
				{ID: 100, Title: "Story-A", Product: 1, OpenedBy: "dev1"},
				{ID: 101, Title: "Story-B", Product: 1, OpenedBy: "dev2"},
				{ID: 200, Title: "Req-C", Product: 1, OpenedBy: "pm1"},
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req RequirementCreateRequest) (*RequirementCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	// Story API只返回Story
	mockStorySvc := &mockStoryService{
		listFn: func(productID int) ([]StoryListItem, error) {
			return []StoryListItem{
				{ID: 100, Title: "Story-A", Product: 1, OpenedBy: "dev1"},
				{ID: 101, Title: "Story-B", Product: 1, OpenedBy: "dev2"},
			}, nil
		},
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			return nil, nil, nil
		},
		createFn: func(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
	}

	deleter := NewDeleterWithMocks(log, mockEpic, mockReq, mockStorySvc)

	items := deleter.FetchByFilter(DeleteFilter{ProductID: 1})

	// 去重后应有4条需求：
	// Story: ID=100, ID=101
	// Requirement: ID=200 (ID=100,101已在Story中)
	// Epic: ID=300 (ID=100,101,200已在前两层)
	if len(items) != 4 {
		t.Fatalf("期望4个去重后的需求, 得到 %d", len(items))
	}

	typeResult := map[int]story.StoryType{}
	for _, item := range items {
		typeResult[item.ID] = item.Type
	}
	if typeResult[100] != story.StoryTypeStory {
		t.Errorf("ID=100 应为Story, 得到 %v", typeResult[100])
	}
	if typeResult[101] != story.StoryTypeStory {
		t.Errorf("ID=101 应为Story, 得到 %v", typeResult[101])
	}
	if typeResult[200] != story.StoryTypeRequirement {
		t.Errorf("ID=200 应为Requirement, 得到 %v", typeResult[200])
	}
	if typeResult[300] != story.StoryTypeEpic {
		t.Errorf("ID=300 应为Epic, 得到 %v", typeResult[300])
	}
}

func TestDeleter_DeleteStoriesConcurrent(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	var deleteCount int32
	mockStorySvc := &mockStoryService{
		deleteFn: func(id int) (map[string]interface{}, *req.Response, error) {
			atomic.AddInt32(&deleteCount, 1)
			return map[string]interface{}{"status": "success"}, nil, nil
		},
		createFn: func(req StoryCreateRequest) (*StoryCreateResponse, *req.Response, error) {
			return nil, nil, nil
		},
		listFn: func(productID int) ([]StoryListItem, error) {
			return nil, nil
		},
	}

	deleter := NewDeleterWithMocks(log, nil, nil, mockStorySvc)

	ids := make([]TypedID, 10)
	for i := 0; i < 10; i++ {
		ids[i] = TypedID{ID: i + 1, Type: story.StoryTypeStory, Title: fmt.Sprintf("需求%d", i+1)}
	}

	results := deleter.DeleteStoriesConcurrent(ids, 3)

	if len(results) != 10 {
		t.Fatalf("期望10个结果, 得到 %d", len(results))
	}
	if int(deleteCount) != 10 {
		t.Errorf("期望10次删除调用, 得到 %d", deleteCount)
	}
	for i, r := range results {
		if !r.Success {
			t.Errorf("需求 #%d 应删除成功", i+1)
		}
	}
}

func TestFormatMatchedList(t *testing.T) {
	items := []TypedID{
		{ID: 1, Type: story.StoryTypeEpic, Title: "业务需求A", OpenedBy: "zhangsan"},
		{ID: 2, Type: story.StoryTypeRequirement, Title: "用户需求B", OpenedBy: "lisi"},
		{ID: 3, Type: story.StoryTypeStory, Title: "研发需求C", OpenedBy: "wangwu"},
	}

	result := FormatMatchedList(items)

	if !bytes.Contains([]byte(result), []byte("3 条需求")) {
		t.Error("应显示总数量")
	}
	if !bytes.Contains([]byte(result), []byte("业务需求A")) {
		t.Error("应显示标题")
	}
	if !bytes.Contains([]byte(result), []byte("zhangsan")) {
		t.Error("应显示创建者")
	}
}

func TestFormatMatchedList_Empty(t *testing.T) {
	result := FormatMatchedList(nil)
	if !bytes.Contains([]byte(result), []byte("无匹配结果")) {
		t.Error("空列表应显示无匹配结果")
	}
}
