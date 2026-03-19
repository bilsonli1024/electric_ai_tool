# Electric AI Tool

智能亚马逊产品图生成工具，使用 Google Gemini AI 生成专业的产品营销图片。

## 📦 项目结构

```
electric_ai_tool/
├── web/              # 前端应用（React + TypeScript）
├── server/           # Node.js 后端服务（TypeScript + Express）
└── go_server/        # Go 后端服务（Go 1.24 + net/http）
```

## 🚀 功能特性

- **产品卖点分析**: 基于关键词和竞品分析，AI 生成专业产品卖点
- **智能图片生成**: 使用 Gemini 多模态能力生成产品营销图
- **图片编辑**: 智能去背景、扩图等图片编辑功能
- **A+ 页面方案**: 生成符合亚马逊规范的 A+ 页面内容方案
- **风格参考**: 支持上传参考图，生成指定风格的产品图

## 💻 技术栈

### 前端 (web/)
- React 18
- TypeScript
- Vite
- TailwindCSS

### 后端选项

#### Node.js 版本 (server/)
- Node.js + Express
- TypeScript
- @google/genai SDK
- **端口**: 3001

#### Go 版本 (go_server/) ⚡ 新增
- Go 1.24
- 标准库 net/http
- google.golang.org/genai SDK
- **端口**: 3002

> **提示**: 两个后端版本功能完全一致，API 兼容。Go 版本性能更优，Node.js 版本开发更快。

## 🏁 快速开始

### 方式一：使用 Node.js 后端

```bash
# 1. 配置环境变量
cd server
cp .env.example .env
# 编辑 .env，填入 GEMINI_API_KEY

# 2. 安装依赖并启动
npm install
npm run dev

# 服务运行在 http://localhost:3001
```

### 方式二：使用 Go 后端（推荐）

```bash
# 1. 配置环境变量
cd go_server
cp .env.example .env
# 编辑 .env，填入 GEMINI_API_KEY

# 2. 启动服务
./start.sh

# 或手动启动
go mod download
go run main.go

# 服务运行在 http://localhost:3002
```

### 启动前端

```bash
cd web
npm install
npm run dev

# 前端运行在 http://localhost:5173
```

## 📚 详细文档

### Node.js 后端
- 查看 [server/](./server/) 目录

### Go 后端
- [README.md](./go_server/README.md) - 完整使用文档
- [COMPARISON.md](./go_server/COMPARISON.md) - Go vs Node.js 对比
- [PROJECT_OVERVIEW.md](./go_server/PROJECT_OVERVIEW.md) - 项目架构总览
- [API_EXAMPLES.md](./go_server/API_EXAMPLES.md) - API 测试示例

## 🔧 环境要求

### Node.js 版本
- Node.js 18+
- npm 或 yarn

### Go 版本
- Go 1.24+
- 参考 [升级指南](./go_server/README.md#环境要求)

### 通用要求
- Gemini API Key（从 [Google AI Studio](https://makersuite.google.com/app/apikey) 获取）

## 🆚 后端选择指南

| 特性 | Node.js | Go |
|------|---------|-----|
| 性能 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 开发速度 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| 部署简易度 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 资源占用 | ~50-80MB | ~15-30MB |
| 并发能力 | 异步 I/O | Goroutines |
| 启动速度 | ~2-3s | ~1s |

### 选择 Node.js，如果：
- 团队主要是前端开发者
- 需要快速原型开发
- 需要使用 npm 生态的特定包

### 选择 Go，如果：
- 追求最佳性能和资源利用率
- 需要处理高并发请求
- 简化部署流程（单文件部署）
- 在资源受限的环境运行

## 🎨 API 端点

两个后端版本提供完全相同的 API：

| 端点 | 方法 | 功能 |
|------|------|------|
| `/api/health` | GET | 健康检查 |
| `/api/analyze` | POST | 分析产品卖点 |
| `/api/generate-image` | POST | 生成产品图片 |
| `/api/edit-image` | POST | 编辑图片 |
| `/api/aplus-content` | POST | 生成 A+ 页面方案 |

详细 API 文档：
- [Node.js API](./server/)
- [Go API 示例](./go_server/API_EXAMPLES.md)

## 🚢 部署

### Node.js 部署

```bash
# 构建前端
cd web
npm run build

# 启动后端（生产模式）
cd ../server
npm run build
npm run start:prod
```

### Go 部署

```bash
# 构建前端
cd web
npm run build

# 编译 Go 服务器
cd ../go_server
go build -o go_server main.go

# 运行
./go_server
```

### Docker 部署

```bash
# Go 版本
cd go_server
docker build -t electric-ai-go .
docker run -p 3002:3002 --env-file .env electric-ai-go

# Node.js 版本
cd server
docker build -t electric-ai-node .
docker run -p 3001:3001 --env-file .env electric-ai-node
```

## 🧪 测试

### 测试后端 API

```bash
# Node.js
curl http://localhost:3001/api/health

# Go
curl http://localhost:3002/api/health
# 或使用测试脚本
cd go_server
./test_api.sh
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🔗 相关链接

- [Google Gemini API](https://ai.google.dev/gemini-api)
- [Google Gemini Go SDK](https://pkg.go.dev/google.golang.org/genai)
- [Google Gemini Node.js SDK](https://www.npmjs.com/package/@google/genai)

---

**Made with ❤️ using Google Gemini AI**
