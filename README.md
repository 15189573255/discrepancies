# Discrepancies

基于 Wails v2 的桌面应用程序，用于比较 ZIP 压缩包与工作目录的文件差异。

## 功能特性

- 可视化展示新增、修改、删除的文件
- 文本文件差异预览，支持快速跳转到差异位置
- 选择性导出差异文件或直接打包为 ZIP
- 可配置的文件/目录排除规则
- 记忆上次使用的路径

## 项目结构

```
Discrepancies/
├── main.go                 # Wails 应用入口
├── app.go                  # 后端 API（暴露给前端的方法）
├── internal/
│   ├── compare/
│   │   ├── compare.go      # 核心比较逻辑、导出功能
│   │   ├── archive.go      # ZIP 文件读取
│   │   └── diff.go         # 文本差异对比
│   ├── config/
│   │   └── config.go       # 配置管理（存储在 ~/.discrepancies/）
│   └── models/
│       └── types.go        # 数据结构定义
├── frontend/
│   ├── src/
│   │   ├── App.svelte      # 主界面组件
│   │   └── style.css       # 全局样式
│   ├── wailsjs/            # 自动生成的 Go 绑定
│   └── index.html
├── build/                  # 构建配置和资源
├── wails.json              # Wails 项目配置
└── go.mod
```

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- Wails CLI v2

### 安装 Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 开发模式

```bash
# 克隆项目
git clone https://gitee.com/ciddwd/discrepancies.git
cd discrepancies

# 启动开发服务器（支持热重载）
wails dev
```

### 生产构建

```bash
wails build
```

构建产物在 `build/bin/` 目录下。

## 使用说明

1. **选择原始 ZIP** - 选择作为基准的 ZIP 压缩包
2. **选择工作目录** - 选择当前工作的项目目录
3. **点击比较** - 分析两者之间的文件差异
4. **查看差异** - 点击左侧文件列表查看详细差异，使用上/下按钮或 `Ctrl+↑/↓` 快速跳转
5. **导出** - 勾选需要的文件后：
   - **导出选中项**: 导出为文件夹
   - **导出为 ZIP**: 直接打包成 ZIP 文件

## 排除规则

默认排除以下文件/目录：

| 类型 | 模式 | 说明 |
|------|------|------|
| 目录 | `obj`, `bin` | .NET 编译输出 |
| 目录 | `.idea`, `.vs`, `.vscode` | IDE 配置 |
| 目录 | `node_modules` | Node.js 依赖 |
| 文件 | `*.vbproj`, `*.csproj` | 项目文件 |
| 文件 | `*.suo`, `*.user` | 用户配置 |

可在应用内的「排除规则」设置中自定义。

## 技术栈

- **后端**: Go + Wails v2
- **前端**: Svelte + TypeScript + Vite + Tailwind CSS
- **差异对比**: github.com/sergi/go-diff

## License

[Apache License 2.0](LICENSE)
