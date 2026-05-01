# Gingest Desktop

Gingest Desktop 是一款基于 [Wails](https://wails.io/) 框架构建的跨平台本地代码提取与分析工具。它可以帮助开发者快速扫描本地项目目录、根据自定义规则过滤文件，并将代码库结构和内容转换为结构化的 XML 格式输出。该工具极大地方便了与大语言模型 (LLM) 交互时的上下文构建、代码审计及项目整体梳理。

## ✨ 主要特性

* **📂 便捷的目录扫描**：支持一键选择目录，同时提供悬浮拖拽支持（Drop Scan Overlay），快速导入本地项目。
* **🕒 历史记录管理**：自动记录并管理最近访问的目录路径，方便随时切换上下文。
* **🌲 可视化文件树**：提供直观的项目文件目录树视图，实时查看扫描结果。
* **⚙️ 灵活的过滤机制**：内置过滤配置管理器，支持读取和自定义忽略规则（精准过滤不需要生成的依赖文件、编译产物等）。
* **📄 结构化提取与预览**：将文件系统和提取的核心代码合并转换为 XML 格式，提供实时预览（Xml Preview Panel），非常适合直接作为 Prompt 喂给各种 AI 编码助手。
* **⚡ 高效轻量**：底层基于 Go 语言处理高并发的文件 I/O 与解析，前端采用 Vue 3 提供流畅的现代化交互体验。

## 🛠️ 技术栈

* **后端 / 核心引擎**：[Go](https://golang.org/) (处理文件读写、树形遍历、AST/XML 生成及配置持久化)
* **前端 / 界面交互**：[Vue 3](https://vuejs.org/) + [TypeScript](https://www.typescriptlang.org/) + [Vite](https://vitejs.dev/)
* **桌面应用框架**：[Wails v2](https://wails.io/)

## 🚀 开发与构建

### 前置依赖

在运行或构建本项目之前，请确保您的计算机上已安装以下依赖：

* [Go](https://go.dev/doc/install) (建议版本 1.21 及以上)
* [Node.js](https://nodejs.org/en/) (建议版本 v18+)
* [npm](https://www.npmjs.com/) 或相关的包管理工具

### 1. 安装 Wails CLI

如果您尚未安装 Wails 命令行工具，请执行以下命令进行安装：

```bash
go install [github.com/wailsapp/wails/v2/cmd/wails@latest](https://github.com/wailsapp/wails/v2/cmd/wails@latest)
```

### 2. 运行开发模式

在项目根目录下执行以下命令启动开发环境。Wails 会自动启动前端的 Vite 开发服务器和后端的 Go 进程，并支持代码的热重载（Hot-Reload）：

```bash
wails dev
```

### 3. 构建生产版本

如需为当前操作系统编译可执行的生产版本安装包/单文件，请运行：

```bash
wails build
```

> **提示**：构建完成后的产物（如 `.exe` 或 `.app`）将统一生成在 `build/bin/` 目录下。您也可以通过 `wails build -platform windows/amd64,darwin/arm64` 来交叉编译不同的目标平台（需满足 Go 的交叉编译环境要求）。

## 📁 核心目录结构

```text
gingest-desktop/
├── build/                # Wails 各平台的构建配置和打包资源 (图标、Windows 安装脚本、Mac plist 等)
├── frontend/             # 前端工程目录
│   ├── src/
│   │   ├── api/          # 自动生成的 Wails 前后端通信接口绑定
│   │   ├── components/   # Vue 核心组件 (扫描控制台、文件树、XML 预览等)
│   │   └── utils/        # 前端工具类 (格式化、XML 处理等)
│   ├── index.html        # 前端入口 HTML
│   └── package.json      # 前端依赖配置
├── internal/             # Go 核心业务逻辑
│   ├── config/           # 全局配置、过滤规则、近期访问历史的持久化处理
│   ├── ingest/           # 文件读取引擎、树形结构解析、XML 组装逻辑
│   ├── model/            # Go 结构体定义 (DTO)
│   └── utils/            # 后端通用工具类 (如文件 Size 计算等)
├── app.go                # Wails 应用生命周期管理及给前端暴露的 Go 方法入口
├── main.go               # Go 主程序启动入口
└── wails.json            # Wails 项目级别配置文件
```

## 📝 许可证

本项目遵循相关开源协议，详情请查看项目根目录下的 [LICENSE](./LICENSE) 文件。
