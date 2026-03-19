# Electric AI Tool

智能亚马逊产品图生成工具，使用 Google Gemini AI 生成专业的产品营销图片。

**🎉 新版本已全面升级！集成用户管理、任务系统、CDN存储等完整功能。**

## ✨ 核心功能

### 🖼️ 图片生成（原有功能）
- **产品卖点分析**: 基于关键词和竞品分析，AI生成专业产品卖点
- **智能图片生成**: 使用Gemini多模态能力生成产品营销图
- **图片编辑**: 智能去背景、扩图等图片编辑功能
- **A+页面方案**: 生成符合亚马逊规范的A+页面内容方案
- **风格参考**: 支持上传参考图，生成指定风格的产品图

### 🆕 新增功能
- **用户系统**: 完整的注册、登录、会话管理
- **任务中心**: 自动记录所有操作，支持任务历史查看
- **CDN集成**: 产品图片自动上传CDN，持久化存储
- **版本管理**: 每次生成保留历史版本，可随时查看
- **导航系统**: 统一导航栏，清晰的功能模块划分

## 📦 项目结构

```
electric_ai_tool/
├── go_server/          # Go后端服务（推荐）
│   ├── config/         # 数据库配置
│   ├── models/         # 数据模型
│   ├── services/       # 业务服务（AI、认证、任务、CDN）
│   ├── handlers/       # API处理器
│   ├── middleware/     # 中间件（CORS、认证）
│   ├── schema.sql      # 数据库表结构
│   ├── init_db.sh      # 数据库初始化脚本
│   └── .env.example    # 环境配置示例
├── web/                # 前端应用
│   └── src/
│       ├── components/ # React组件（Auth、Navbar、TaskCenter等）
│       ├── services/   # API服务
│       └── types/      # TypeScript类型定义
├── docs/               # 完整文档
│   ├── DEPLOYMENT.md      # 部署指南
│   ├── ARCHITECTURE.md    # 架构说明
│   ├── PROJECT_SUMMARY.md # 项目总结
│   └── QUICKSTART.md      # 快速参考
└── start.sh            # 一键启动脚本
```

## 🚀 快速开始

### 一键启动（推荐）

```bash
# 1. 初始化数据库
cd go_server
./init_db.sh          # 按提示输入数据库信息

# 2. 配置环境
cp .env.example .env  # 编辑.env文件，填入API Key和数据库配置

# 3. 启动服务
cd ..
./start.sh            # 自动构建前端并启动后端

# 4. 访问系统
# 打开浏览器访问: http://localhost:3002
```

### 开发模式

```bash
# 终端1 - 启动后端
cd go_server
go run main.go        # 后端: http://localhost:3002

# 终端2 - 启动前端
cd web
npm run dev           # 前端: http://localhost:5173
```

## 💻 技术栈

### 后端
- **语言**: Go 1.24
- **数据库**: MySQL 8.0 + 连接池
- **AI服务**: Google Gemini API
- **安全**: bcrypt密码加密
- **存储**: CDN + 本地备份

### 前端
- **框架**: React 19
- **语言**: TypeScript 5.8
- **构建**: Vite 6.2
- **样式**: TailwindCSS 4.1
- **动画**: Motion 12.36

### 数据库
- **引擎**: InnoDB
- **字符集**: utf8mb4_unicode_ci
- **表结构**: 5张核心表（用户、会话、任务、任务历史、CDN图片）

## 🔌 API接口

### 认证接口（公开）
```
POST /api/auth/register    # 用户注册
POST /api/auth/login       # 用户登录
```

### 认证接口（需登录）
```
POST /api/auth/logout      # 用户登出
GET  /api/auth/me          # 获取用户信息
```

### 任务接口（需登录）
```
POST /api/tasks/analyze           # 产品分析（创建任务）
POST /api/tasks/generate-image    # 图片生成（创建任务）
GET  /api/tasks                   # 我的任务列表
GET  /api/tasks/all               # 全部任务列表
GET  /api/tasks/history           # 任务历史记录
```

### 兼容接口（无需登录，保留原有功能）
```
POST /api/analyze          # 产品分析
POST /api/generate-image   # 图片生成
POST /api/edit-image       # 图片编辑
POST /api/aplus-content    # A+内容生成
```

## 🗄️ 数据库表

| 表名 | 说明 | 字段数 |
|------|------|--------|
| users | 用户信息表 | 8 |
| sessions | 用户会话表 | 4 |
| tasks | 任务记录表 | 12 |
| task_history | 任务历史表 | 13 |
| cdn_images | CDN图片表 | 9 |

## 📚 完整文档

| 文档 | 说明 |
|------|------|
| [QUICKSTART.md](./QUICKSTART.md) | 快速参考（命令、API、故障排查） |
| [DEPLOYMENT.md](./DEPLOYMENT.md) | 详细部署指南 |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | 系统架构说明 |
| [PROJECT_SUMMARY.md](./PROJECT_SUMMARY.md) | 项目完成总结 |

## ⚙️ 环境配置

```env
# Gemini API配置（必需）
GEMINI_API_KEY=your_gemini_api_key

# 服务器端口（可选）
PORT=3002

# 数据库配置（必需）
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=electric_ai_tool

# CDN配置（可选，不配置时使用本地存储）
CDN_ENDPOINT=
CDN_BUCKET=
CDN_ACCESS_KEY=
CDN_SECRET_KEY=
```

## 🎯 使用流程

1. **注册/登录** → 创建账号并登录系统
2. **一键生图** → 上传产品图，生成营销图，自动创建任务
3. **任务中心** → 查看所有任务记录和历史版本
4. **用户管理** → 查看个人信息和账号状态

## 🔐 安全特性

- ✅ 密码bcrypt加密存储
- ✅ 会话Token随机生成（64字节）
- ✅ 会话自动过期（24小时）
- ✅ SQL参数化查询（防注入）
- ✅ CORS跨域保护
- ✅ 认证中间件统一管理

## 📊 系统特性

### 性能优化
- 数据库连接池（50个最大连接）
- 索引优化（用户、任务、会话）
- 分页查询支持
- CDN加速图片访问

### 可维护性
- 模块化设计
- 服务分层架构
- 统一错误处理
- 完整类型定义
- 详细文档

### 可扩展性
- 中间件架构
- 服务接口化
- CDN可配置切换
- 任务类型可扩展
- 前端组件化

## 🐛 故障排查

| 问题 | 解决方案 |
|------|----------|
| 数据库连接失败 | 1. 检查MySQL是否运行<br>2. 检查.env配置<br>3. 运行init_db.sh |
| 端口被占用 | 修改.env中的PORT值 |
| API调用失败 | 1. 检查GEMINI_API_KEY<br>2. 查看后端日志 |
| 前端构建失败 | 删除node_modules，重新npm install |

## 🚢 部署建议

### 开发环境
```bash
./start.sh           # 使用一键启动脚本
```

### 生产环境
```bash
# 1. 构建前端
cd web && npm run build

# 2. 编译后端
cd ../go_server && go build -o server

# 3. 配置systemd或supervisor
./server
```

### Docker部署
```bash
# 待添加Dockerfile
```

## 🔄 版本历史

### v2.0.0（当前版本）
- ✅ 新增用户认证系统
- ✅ 新增任务管理系统
- ✅ 新增CDN图片存储
- ✅ 新增任务历史版本
- ✅ 新增导航栏和模块化界面
- ✅ 数据库持久化
- ✅ 完整文档

### v1.0.0
- 基础图片生成功能
- 产品分析
- A+内容生成

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📄 许可证

Apache-2.0 License

## 🔗 相关链接

- [Google Gemini API](https://ai.google.dev/gemini-api)
- [Google Gemini Go SDK](https://pkg.go.dev/google.golang.org/genai)
- [MySQL 8.0文档](https://dev.mysql.com/doc/refman/8.0/en/)
- [React 19文档](https://react.dev/)

## 💡 技术亮点

- 🚀 **高性能**: Go语言 + 连接池，支持高并发
- 🔒 **安全可靠**: bcrypt加密 + 会话管理 + SQL防注入
- 📦 **开箱即用**: 一键启动脚本，5分钟完成部署
- 🎨 **现代化UI**: React 19 + TailwindCSS，响应式设计
- 📊 **完整追踪**: 任务系统 + 历史版本，操作可追溯
- ☁️ **灵活存储**: CDN + 本地存储双模式
- 📚 **文档完善**: 4份详细文档，快速上手

---

**Made with ❤️ using Google Gemini AI**

**🎉 升级到v2.0，体验完整的企业级功能！**
