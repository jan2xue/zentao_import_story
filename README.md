# 禅道需求导入工具 (ZenTao Story Importer)

**ZenTao Story Importer**是一个基于 Go 语言的高效工具，旨在通过从 Excel 电子表格直接批量导入用户需求（Story）到禅道（ZenTao）系统，简化项目管理流程。

> **提示**
> 该工具是产品经理和开发人员将本地 Excel 需求快速、准确地迁移到禅道系统的理想选择。

## 🚀 核心功能

*   **批量导入**：数秒内导入成百上千条需求。
*   **批量导出**：从禅道服务器导出指定产品的所有需求到Excel文件。
*   **智能字段映射**：自动将 Excel 列映射到禅道需求字段（标题、优先级、分类等）。
*   **数据验证**：预检查数据完整性，确保必填字段（标题、产品 ID 等）存在且有效。
*   **详细报告**：生成包含导入结果、耗时统计和成功率的详尽报告。
*   **灵活配置**：支持通过 YAML 文件进行配置，并可以通过命令行参数进行覆盖。

## 🛠️ 安装

确保您的系统中已安装 **Go 1.16+**。

```bash
# 克隆仓库
git clone https://github.com/jan2xue/zentao_import_story.git
cd zentao_import_story

# 编译可执行文件
go build -o zentao_story_tool.exe
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

defaultPriority: 3                      # 默认优先级 1-4

# 文件配置
excelFile: "requirements.xlsx"          # 默认 Excel 文件路径

# 默认值配置
defaultPriority: 3                      # 如果 Excel 中未填写优先级，则使用此默认值 (1-4)
```

## 📊 Excel 模板结构

Excel 文件应遵循以下列结构（第一行为表头）：

| 标题* | 产品 ID* | 优先级 | 分类* | 需求描述* | 父需求 ID | 来源 | 来源备注 | 预计工时 | 关键词 | 验收标准 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |

> [!IMPORTANT]
> \* 标记的字段为 **必填项**。

### 字段详细说明

*   **标题**：需求的简短摘要。
*   **产品 ID**：禅道系统中对应的产品数字 ID。
*   **优先级**：1 到 4（1 为最高）。如果为空，则使用 `defaultPriority`。
*   **分类**：必须是以下之一：`feature` (功能), `interface` (接口), `performance` (性能), `safe` (安全), `experience` (体验), `improve` (改进)。
*   **需求描述**：需求的详细说明，支持文本或 HTML。
*   **预计工时**：预计的人工耗时（数字）。
*   **验收标准**：需求的验收标准。

## 📖 使用方法

### 基础用法
使用 `config.yaml` 中的默认设置：

```powershell
./zentao_story_tool.exe -action import
```

### 导出功能
从禅道服务器导出需求到Excel文件：

```powershell
./zentao_story_tool.exe -action export -product 1
```

### 高级用法
指定自定义配置文件或 Excel 文件：

```powershell
# 导入操作
./zentao_story_tool.exe -action import -config custom-config.yaml -excel data.xlsx

# 导出操作
./zentao_story_tool.exe -action export -config custom-config.yaml -product 1 -excel exported_stories.xlsx
```

### 命令行参数
*   `-config`: 配置文件路径 (默认: `config.yaml`)。
*   `-excel`: Excel 文件路径 (会覆盖配置文件中的设置)。
*   `-action`: 操作类型: `import`(导入) 或 `export`(导出) (默认: `import`)。
*   `-product`: 产品ID(导出时使用)。

## 📝 错误处理与日志

*   **日志记录**：程序运行详情将保存到当前目录下的 `.log` 文件中。
*   **执行报告**：每次运行结束后，控制台都会打印一份结导入结果报告。
*   **容错处理**：工具独立处理每一行数据，单行失败不会中断整个导入流程。

## 📄 许可证

本项目采用 [MIT License](LICENSE) 许可证。