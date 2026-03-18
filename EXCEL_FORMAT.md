# Excel导入格式说明

## 命令行参数指定需求类型

通过 `-type` 参数指定导入的需求类型：

```bash
# 导入业务需求
./zentao_tool -type epic -excel requirements.xlsx

# 导入用户需求
./zentao_tool -type requirement -excel requirements.xlsx

# 导入研发需求（默认）
./zentao_tool -type story -excel requirements.xlsx
```

## 列格式（第一行为标题行）

| 列序号 | 列名 | 必填 | 说明 |
|--------|------|------|------|
| 1 | 标题 | 是 | 需求的标题 |
| 2 | 产品ID | 是 | 数字，禅道中的产品ID |
| 3 | 优先级 | 否 | 1-4的数字，默认3 |
| 4 | 分类 | 是 | feature/interface/performance/safe/experience/improve/other |
| 5 | 需求描述 | 是 | 详细描述 |
| 6 | 父需求ID | 否 | 数字，父需求的ID |
| 7 | 来源 | 否 | customer/user/po/market/service/operation/support/competitor/partner/dev/tester/bug/forum/other |
| 8 | 来源备注 | 否 | 字符串 |
| 9 | 预计工时 | 否 | 数字 |
| 10 | 关键词 | 否 | 字符串 |
| 11 | 验收标准 | 否 | 字符串 |

## 示例数据

```
标题                    | 产品ID | 优先级 | 分类    | 需求描述                                              | 父需求ID | 来源   | 来源备注 | 预计工时 | 关键词          | 验收标准
智能座舱系统            | 1      | 1      | feature | 开发新一代智能座舱系统，集成多媒体、导航、语音控制等功能 | 0        | market | 市场调研 | 100      | 智能座舱,多媒体 | 功能完整，性能稳定
语音控制功能            | 1      | 2      | feature | 用户可以通过语音控制车内设备                           | 1        | user   | 用户反馈 | 20       | 语音,控制       | 识别率95%以上
实现语音识别模块        | 1      | 3      | feature | 开发语音识别核心模块，支持中英文识别                   | 2        | dev    | 技术方案 | 40       | 语音识别,AI     | 单元测试覆盖率80%
```

## 需求类型说明

- **epic/业务需求**: 高层次的业务需求，通常是产品的大功能模块
- **requirement/用户需求**: 从用户角度出发的需求描述
- **story/研发需求**: 具体的研发实现需求，可直接分配给开发团队

## 分类选项

- `feature` - 功能
- `interface` - 接口
- `performance` - 性能
- `safe` - 安全
- `experience` - 体验
- `improve` - 改进
- `other` - 其他

## 来源选项

- `customer` - 客户
- `user` - 用户
- `po` - 产品经理
- `market` - 市场
- `service` - 客服
- `operation` - 运营
- `support` - 技术支持
- `competitor` - 竞争对手
- `partner` - 合作伙伴
- `dev` - 开发人员
- `tester` - 测试人员
- `bug` - Bug
- `forum` - 论坛
- `other` - 其他
