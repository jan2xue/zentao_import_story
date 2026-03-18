# 禅道需求导入工具 (ZenTao Story Importer)

**ZenTao Story Importer** 是一个基于 Go 语言的高效工具，旨在通过从 Excel 电子表格直接批量导入需求到禅道（ZenTao）系统，简化项目管理流程。

> **提示**
> 该工具是产品经理和开发人员将本地 Excel 需求快速、准确地迁移到禅道系统的理想选择。

## 🚀 核心功能

*   **批量导入**：数秒内导入成百上千条需求。
*   **批量删除**：支持按需求ID或产品ID批量删除需求，删除前有确认提示。
*   **多类型支持**：支持业务需求(epic)、用户需求(requirement)、研发需求(story)三种类型。
*   **智能字段映射**：自动将 Excel 列映射到禅道需求字段（标题、优先级、分类等）。
*   **数据验证**：预检查数据完整性，确保必填字段（标题、产品 ID 等）存在且有效。
*   **详细报告**：生成包含导入/删除结果、耗时统计和成功率的详尽报告。
*   **灵活配置**：支持通过 YAML 文件进行配置，并可以通过命令行参数进行覆盖。

## 📁 项目结构

```
.
├── cmd/zentao_tool/          # 应用程序入口
│   └── main.go
├── internal/                  # 私有代码
│   ├── config/               # 配置管理
│   ├── excel/                # Excel读写操作
│   ├── logger/               # 日志记录
│   └── zentao/               # 禅道API封装
├── pkg/story/                # 需求领域模型（可复用）
├── config.yaml               # 配置文件
├── requirements.xlsx         # 示例Excel文件
└── README.md
```

## 🛠️ 安装

确保您的系统中已安装 **Go 1.16+**。

```bash
# 克隆仓库
git clone https://github.com/jan2xue/zentao_import_story.git
cd zentao_import_story

# 运行测试
go test ./...

# 编译可执行文件
go build -o zentao_story_tool.exe ./cmd/zentao_tool
```

## ⚙️ 配置说明

复制提供的示例文件创建 `config.yaml`：

```bash
cp config.example.yaml config.yaml
```

### 配置项详情 (`config.yaml`)

```yaml
# 禅道系统配置
zentaoUrl: "http://your-zentao-url"     # 禅道服务器基础地址
zentaoUsername: "admin"                 # 禅道用户名
zentaoPassword: "password"              # 禅道密码

# 文件配置
excelFile: "requirements.xlsx"          # 默认 Excel 文件路径

# 默认值配置
defaultPriority: 3                      # 默认优先级（1-4），如果Excel中未指定则使用此值
defaultReviewer: "username"             # 默认评审人（用户名），创建需求时必填
defaultModule: 0                        # 默认模块ID，创建用户需求时需要有效的模块ID
```

### 配置项说明

| 配置项 | 说明 | 必填 |
|--------|------|------|
| `zentaoUrl` | 禅道服务器地址 | 是 |
| `zentaoUsername` | 禅道登录用户名 | 是 |
| `zentaoPassword` | 禅道登录密码 | 是 |
| `excelFile` | Excel 文件路径 | 导入时必填 |
| `defaultPriority` | 默认优先级 1-4 | 否，默认 3 |
| `defaultReviewer` | 默认评审人用户名 | **是**，API 要求必填 |
| `defaultModule` | 默认模块ID | 用户需求必填 |

> [!IMPORTANT]
> `defaultReviewer` 为必填项，禅道 API 创建需求时要求指定评审人。
> 
> `defaultModule` 仅在导入用户需求(requirement)时需要，请先在禅道 Web 界面创建模块并获取模块ID。

## 📖 使用方法

### 导入需求

通过 `-type` 参数指定导入的需求类型：

```powershell
# 导入研发需求（默认）
./zentao_story_tool.exe -action import -type story

# 导入用户需求
./zentao_story_tool.exe -action import -type requirement

# 导入业务需求
./zentao_story_tool.exe -action import -type epic
```

### 删除需求

```powershell
# 按需求ID删除（支持多个ID，逗号分隔）
./zentao_story_tool.exe -action delete -type story -ids 123,456,789

# 按产品ID删除该产品下所有需求
./zentao_story_tool.exe -action delete -type story -product 78
```

> [!WARNING]
> 删除操作会显示确认提示，需要输入 `yes` 或 `y` 确认后才会执行。此操作不可撤销！

### 高级用法

指定自定义配置文件或 Excel 文件：

```powershell
# 使用自定义配置文件
./zentao_story_tool.exe -config custom-config.yaml -excel data.xlsx

# 指定产品ID和Excel文件
./zentao_story_tool.exe -action import -type story -product 78 -excel requirements.xlsx
```

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-config` | 配置文件路径 | `config.yaml` |
| `-excel` | Excel 文件路径 | 配置文件中的值 |
| `-action` | 操作类型: `import`(导入) 或 `delete`(删除) | `import` |
| `-type` | 需求类型: `story`(研发需求)、`requirement`(用户需求)、`epic`(业务需求) | `story` |
| `-ids` | 需求ID列表（删除时使用，逗号分隔） | - |
| `-product` | 产品ID | 配置文件中的值 |

## 📊 Excel 格式说明

### 列格式（第一行为标题行）

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

### 示例数据

| 标题 | 产品ID | 优先级 | 分类 | 需求描述 | 父需求ID | 来源 | 来源备注 | 预计工时 | 关键词 | 验收标准 |
|------|--------|--------|------|----------|----------|------|----------|----------|--------|----------|
| 智能座舱系统 | 1 | 1 | feature | 开发新一代智能座舱系统，集成多媒体、导航、语音控制等功能 | 0 | market | 市场调研 | 100 | 智能座舱,多媒体 | 功能完整，性能稳定 |
| 语音控制功能 | 1 | 2 | feature | 用户可以通过语音控制车内设备 | 1 | user | 用户反馈 | 20 | 语音,控制 | 识别率95%以上 |
| 实现语音识别模块 | 1 | 3 | feature | 开发语音识别核心模块，支持中英文识别 | 2 | dev | 技术方案 | 40 | 语音识别,AI | 单元测试覆盖率80% |

### 分类选项

| 值 | 说明 |
|----|------|
| `feature` | 功能 |
| `interface` | 接口 |
| `performance` | 性能 |
| `safe` | 安全 |
| `experience` | 体验 |
| `improve` | 改进 |
| `other` | 其他 |

### 来源选项

| 值 | 说明 |
|----|------|
| `customer` | 客户 |
| `user` | 用户 |
| `po` | 产品经理 |
| `market` | 市场 |
| `service` | 客服 |
| `operation` | 运营 |
| `support` | 技术支持 |
| `competitor` | 竞争对手 |
| `partner` | 合作伙伴 |
| `dev` | 开发人员 |
| `tester` | 测试人员 |
| `bug` | Bug |
| `forum` | 论坛 |
| `other` | 其他 |

## 📝 需求类型说明

| 类型 | 参数值 | 说明 | 配置要求 |
|------|--------|------|----------|
| 业务需求 | `epic` | 高层次的业务需求，通常是产品的大功能模块 | 需要有相应权限 |
| 用户需求 | `requirement` | 从用户角度出发的需求描述 | **需配置 `defaultModule`** |
| 研发需求 | `story` | 具体的研发实现需求，可直接分配给开发团队 | **需配置 `defaultReviewer`** |

> [!IMPORTANT]
> - 研发需求(story)需要在 `config.yaml` 中配置 `defaultReviewer`（评审人用户名）
> - 用户需求(requirement)需要在 `config.yaml` 中配置 `defaultModule`（有效的模块ID）

### 配置模块ID（用户需求）

用户需求(requirement)必须指定有效的模块ID。获取模块ID的步骤：

1. 登录禅道 Web 界面
2. 进入 **产品** → 选择产品 → **模块**
3. 创建或查看模块
4. 从 URL 或模块列表中获取模块ID（如：`/product-module-edit-1-0-0.html` 中的 `1`）
5. 在 `config.yaml` 中配置：
   ```yaml
   defaultModule: 1  # 替换为实际的模块ID
   ```

## 📝 错误处理与日志

*   **日志记录**：程序运行详情将保存到当前目录下的 `import.log` 文件中。
*   **执行报告**：每次运行结束后，控制台都会打印一份结果报告。
*   **容错处理**：工具独立处理每条数据，单条失败不会中断整个流程。

## 📄 许可证

本项目采用 [MIT License](LICENSE) 许可证。
