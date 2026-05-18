# 禅道需求导入工具 (ZenTao Story Importer)

**ZenTao Story Importer** 是一个基于 Go 语言的高效工具，旨在通过从 Excel 电子表格直接批量导入需求到禅道（ZenTao）系统，简化项目管理流程。

> **提示**
> 该工具是产品经理和开发人员将本地 Excel 需求快速、准确地迁移到禅道系统的理想选择。

## 🚀 核心功能

*   **层级导入**：支持在一个 Excel 中混合填写不同类型需求，自动按 Epic → Requirement → Story 顺序导入并建立父子层级关系。
*   **智能ID解析**：Epic/Requirement 创建后禅道不返回ID，工具自动通过产品列表查询实际ID（采用分层去重策略避免API返回重复数据），确保父子关系正确建立。
*   **智能引用**：支持 `@行号` 格式引用父需求，无需提前知道禅道 ID，工具自动解析。
*   **条件删除**：删除操作必须指定产品ID，支持标题（部分匹配）和创建者筛选组合条件，带二次确认防误删。
*   **批量删除**：支持按产品ID批量删除需求（自动涵盖所有类型），删除前有确认提示。
*   **产品确认**：导入前显示产品信息和需求类型分布，要求用户确认，防止数据导入错误产品。
*   **自动分页**：删除功能支持自动分页获取，突破API默认20条限制。
*   **智能字段映射**：自动将 Excel 列映射到禅道需求字段（标题、优先级、分类等）。
*   **数据验证**：预检查数据完整性，确保必填字段（标题、产品 ID 等）存在且有效。
*   **详细报告**：生成包含导入/删除结果、耗时统计和成功率的详尽报告。
*   **灵活配置**：支持通过 YAML 文件进行配置，并可以通过命令行参数进行覆盖。

## ⚠️ 系统要求

*   **禅道版本**：V21.7.9 及以上（开源版）/ V21.0+（企业版/旗舰版/IPD版）
*   **操作系统**：Windows 7/10/11、Linux、macOS
*   **Go 版本**：1.16+（仅编译时需要）

## 📁 项目结构

```
.
├── cmd/
│   ├── zentao_tool/          # 主程序入口
│   │   └── main.go
│   └── release/              # 发布打包工具
│       └── main.go
├── internal/                  # 私有代码
│   ├── config/               # 配置管理
│   ├── excel/                # Excel读写操作
│   ├── logger/               # 日志记录
│   └── zentao/               # 禅道API封装
├── pkg/story/                # 需求领域模型（可复用）
├── config.example.yaml       # 配置文件示例
├── requirements.xlsx         # 示例Excel文件
├── changelog.txt             # 版本更新记录
├── 使用说明.txt              # 详细使用说明
└── README.md
```

## 🛠️ 安装

### 方式一：下载发布版本

从 `release/` 目录下载对应版本的 zip 包，解压后即可使用。

### 方式二：从源码编译

确保您的系统中已安装 **Go 1.16+**。

```bash
# 克隆仓库
git clone https://github.com/jan2xue/zentao_import_story.git
cd zentao_import_story

# 安装依赖
go mod download

# 编译可执行文件
go build -o zentao_story_tool.exe ./cmd/zentao_tool
```

### 方式三：生成发布包

```bash
# 运行发布打包工具
go run ./cmd/release -version 1.1.0

# 生成的文件位于 release/ 目录
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
| `defaultModule` | 默认模块ID | 用户需求在Excel未指定时必填 |

> [!IMPORTANT]
> `defaultReviewer` 为必填项，禅道 API 创建需求时要求指定评审人。
>
> `defaultModule` 仅在导入用户需求(requirement)时需要，请先在禅道 Web 界面创建模块并获取模块ID。

## 📖 使用方法

### 导入需求

导入采用层级模式，Excel中"需求类型"列决定每行数据的类型，工具自动按 Epic → Requirement → Story 顺序导入并建立父子关系：

> [!IMPORTANT]
> Epic和Requirement的禅道创建API不会返回ID，工具会在创建后自动通过产品列表查询获取禅道实际生成的ID，确保子需求能正确关联父需求。因此，**同一产品下请避免创建标题完全相同的需求**，以免ID匹配错误。

```powershell
# 导入需求
./zentao_story_tool.exe -action import
```

> [!NOTE]
> 导入前会显示产品信息和需求类型分布确认界面，需要用户确认产品ID和名称无误后才会执行导入，防止数据导入错误产品。

### 父需求引用

**父需求引用格式**（Excel第8列"父需求ID"）：
- `@行号`：引用本 Excel 中第 N 行数据创建后得到的禅道 ID（如 `@1` 引用第 1 行），行号从1开始（不包含标题行）
- 纯数字：直接使用禅道系统中已存在的需求 ID

### 删除需求

删除操作必须指定产品ID，支持标题（部分匹配）和创建者（精确匹配）作为可选过滤条件：

```powershell
# 删除产品78下所有需求（包括Epic/Requirement/Story）
./zentao_story_tool.exe -action delete -product 78

# 删除产品78下标题包含"测试"的需求
./zentao_story_tool.exe -action delete -product 78 -title 测试

# 删除产品78下由zhangsan创建的需求
./zentao_story_tool.exe -action delete -product 78 -openedBy zhangsan

# 组合条件：标题含"测试"且创建者为zhangsan
./zentao_story_tool.exe -action delete -product 78 -title 测试 -openedBy zhangsan
```

> [!IMPORTANT]
> - `-product` 为必填参数
> - `-title` 为部分匹配（包含即匹配）
> - `-openedBy` 为精确匹配（需填写禅道账号名）
> - 执行前会显示匹配结果列表，需输入 `yes` 确认后才删除
> - 大批量删除(>20条)自动切换并发模式提升性能
> - 查询时采用分层去重策略：先获取Story，再获取Requirement（去重），最后获取Epic（去重），避免禅道API返回的重复ID

### 高级用法

指定自定义配置文件或 Excel 文件：

```powershell
# 使用自定义配置文件
./zentao_story_tool.exe -config custom-config.yaml -excel data.xlsx

# 指定Excel文件
./zentao_story_tool.exe -action import -excel requirements.xlsx
```

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-config` | 配置文件路径 | `config.yaml` |
| `-excel` | Excel 文件路径 | 配置文件中的值 |
| `-action` | 操作类型: `import`(导入) 或 `delete`(删除) | `import` |
| `-product` | 产品ID（删除时必填） | - |
| `-title` | 标题筛选，部分匹配（删除时可选） | - |
| `-openedBy` | 创建者筛选，精确匹配账号名（删除时可选） | - |

## 📊 Excel 格式说明

### 列格式（第一行为标题行，共13列）

| 列序号 | 列名 | 必填 | 说明 |
|--------|------|------|------|
| 1 | 需求类型 | 是 | `epic`/`requirement`/`story` |
| 2 | 产品ID | 是 | 数字，禅道中的产品ID |
| 3 | 模块ID | 否 | 数字，不填则使用配置文件默认值 |
| 4 | 标题 | 是 | 需求的标题 |
| 5 | 优先级 | 否 | 1-4的数字，默认3 |
| 6 | 分类 | 是 | feature/interface/performance/safe/experience/improve/other |
| 7 | 需求描述 | 是 | 详细描述 |
| 8 | 父需求ID | 否 | `@行号`引用或纯数字ID |
| 9 | 来源 | 否 | customer/user/po/market/service/operation/support/competitor/partner/dev/tester/bug/forum/other |
| 10 | 来源备注 | 否 | 字符串 |
| 11 | 预计工时 | 否 | 数字 |
| 12 | 关键词 | 否 | 字符串 |
| 13 | 验收标准 | 否 | 字符串 |

### 示例数据

| 需求类型 | 产品ID | 模块ID | 标题 | 优先级 | 分类 | 需求描述 | 父需求ID | 来源 | 来源备注 | 预计工时 | 关键词 | 验收标准 |
|----------|--------|--------|------|--------|------|----------|----------|------|----------|----------|--------|----------|
| epic | 78 | | 智能座舱系统 | 1 | feature | 开发新一代智能座舱系统，集成多媒体、导航、语音控制等功能 | | market | 市场调研 | 100 | 智能座舱,多媒体 | 功能完整，性能稳定 |
| requirement | 78 | 5 | 语音控制功能 | 2 | feature | 用户可以通过语音控制车内设备 | @1 | user | 用户反馈 | 20 | 语音,控制 | 识别率95%以上 |
| story | 78 | 5 | 实现语音识别模块 | 3 | feature | 开发语音识别核心模块，支持中英文识别 | @2 | dev | 技术方案 | 40 | 语音识别,AI | 单元测试覆盖率80% |

> 说明：`@1` 引用第1行（智能座舱系统）创建后的禅道ID，`@2` 引用第2行（语音控制功能）创建后的禅道ID。

### 需求类型说明

| 类型 | Excel值 | 说明 | 配置要求 |
|------|---------|------|----------|
| 业务需求 | `epic` | 高层次的业务需求，通常是产品的大功能模块 | **需配置 `defaultReviewer`**，需有相应权限 |
| 用户需求 | `requirement` | 从用户角度出发的需求描述 | **需配置 `defaultReviewer`** 和 `defaultModule` |
| 研发需求 | `story` | 具体的研发实现需求，可直接分配给开发团队 | **需配置 `defaultReviewer`** |

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

> [!IMPORTANT]
> - 所有类型的需求（epic/requirement/story）均需在 `config.yaml` 中配置 `defaultReviewer`（评审人用户名）
> - 用户需求(requirement)需要有效的模块ID，可通过Excel第3列"模块ID"指定，或在 `config.yaml` 中配置 `defaultModule` 作为默认值

### 模块ID配置

模块ID优先从Excel第3列"模块ID"读取；若该列为空，则使用 `config.yaml` 中的 `defaultModule`。

用户需求(requirement)必须指定有效的模块ID(>0)。获取模块ID的步骤：

1. 登录禅道 Web 界面
2. 进入 **产品** → 选择产品 → **模块**
3. 创建或查看模块
4. 从 URL 或模块列表中获取模块ID（如：`/product-module-edit-1-0-0.html` 中的 `1`）
5. 在Excel模块ID列填写，或在 `config.yaml` 中配置：
   ```yaml
   defaultModule: 1  # 替换为实际的模块ID
   ```

## 📝 错误处理与日志

*   **日志记录**：程序运行详情将保存到当前目录下的 `import.log` 文件中。
*   **执行报告**：每次运行结束后，控制台都会打印一份结果报告。
*   **容错处理**：工具独立处理每条数据，单条失败不会中断整个流程。

## 📋 版本历史

详见 [changelog.txt](changelog.txt)

## 📄 许可证

本项目采用 [MIT License](LICENSE) 许可证。
