# 快速参考

## 🚀 快速启动

```bash
# 完整部署（首次）
cd go_server && ./init_db.sh    # 初始化数据库
cp .env.example .env             # 配置环境变量
cd .. && ./start.sh              # 启动服务

# 访问系统
http://localhost:3002
```

## 📋 核心命令

```bash
# 开发模式
cd go_server && go run main.go   # 后端: localhost:3002
cd web && npm run dev            # 前端: localhost:5173

# 生产模式
cd web && npm run build          # 构建前端
cd ../go_server && go run main.go # 启动后端（含静态文件）

# 数据库
./go_server/init_db.sh          # 初始化数据库
mysql -u root -p electric_ai_tool < schema.sql  # 导入表结构
```

## 🗂️ 项目结构

```
electric_ai_tool/
├── go_server/          # 后端服务
│   ├── config/         # 数据库配置
│   ├── models/         # 数据模型
│   ├── services/       # 业务服务
│   ├── handlers/       # API处理器
│   ├── middleware/     # 中间件
│   ├── utils/          # 工具函数
│   ├── schema.sql      # 数据库表结构
│   ├── init_db.sh      # 数据库初始化
│   └── .env.example    # 环境配置示例
├── web/                # 前端应用
│   └── src/
│       ├── components/ # React组件
│       ├── services/   # API服务
│       └── types/      # 类型定义
└── docs/               # 文档
    ├── DEPLOYMENT.md
    ├── ARCHITECTURE.md
    └── PROJECT_SUMMARY.md
```

## 🔌 API端点

### 认证 (Public)
```
POST /api/auth/register    注册
POST /api/auth/login       登录
```

### 认证 (需登录)
```
POST /api/auth/logout      登出
GET  /api/auth/me          获取用户信息
```

### 任务 (需登录)
```
POST /api/tasks/analyze           产品分析
POST /api/tasks/generate-image    图片生成
GET  /api/tasks                   我的任务
GET  /api/tasks/all               全部任务
GET  /api/tasks/history           任务历史
```

### 兼容接口 (Public)
```
POST /api/analyze            产品分析
POST /api/generate-image     图片生成
POST /api/edit-image         图片编辑
POST /api/aplus-content      A+内容
```

## 🗄️ 数据库表

| 表名 | 说明 |
|------|------|
| users | 用户信息 |
| sessions | 用户会话 |
| tasks | 任务记录 |
| task_history | 任务历史版本 |
| cdn_images | CDN图片记录 |

## ⚙️ 环境变量

```env
# 必需配置
GEMINI_API_KEY=your_key
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=electric_ai_tool

# 可选配置
PORT=3002
DB_HOST=localhost
DB_PORT=3306
CDN_ENDPOINT=
CDN_BUCKET=
CDN_ACCESS_KEY=
CDN_SECRET_KEY=
```

## 🎯 功能模块

| 模块 | 路径 | 功能 |
|------|------|------|
| 登录/注册 | Auth.tsx | 用户认证 |
| 一键生图 | App.tsx | 产品分析、图片生成 |
| 任务中心 | TaskCenter.tsx | 任务列表、历史版本 |
| 用户管理 | UserManagement.tsx | 个人信息 |

## 🔐 认证流程

```
注册 → 生成会话 → 返回session_id
登录 → 验证密码 → 生成会话 → 返回session_id
请求 → 携带session_id → 验证会话 → 处理请求
登出 → 删除会话
```

## 📦 依赖管理

```bash
# Go依赖
go mod tidy

# Node依赖
npm install
```

## 🐛 故障排查

| 问题 | 解决方案 |
|------|----------|
| 数据库连接失败 | 检查MySQL服务、.env配置 |
| 端口被占用 | 修改PORT环境变量 |
| 前端构建失败 | 删除node_modules重新安装 |
| Go编译失败 | 运行go mod tidy |

## 📚 文档索引

- **DEPLOYMENT.md** - 详细部署指南
- **ARCHITECTURE.md** - 系统架构说明
- **PROJECT_SUMMARY.md** - 项目完成总结
- **README.md** - 项目概述（如有）

## 🎨 技术栈

**后端**: Go 1.24 + MySQL 8.0 + Gemini API  
**前端**: React 19 + TypeScript + Vite + TailwindCSS  
**安全**: bcrypt + 会话管理  
**存储**: CDN + 本地备份

## 📝 注意事项

1. ⚠️ 首次使用必须初始化数据库
2. ⚠️ GEMINI_API_KEY必须配置
3. ⚠️ CDN配置为可选（可用本地存储）
4. ⚠️ 会话有效期24小时
5. ⚠️ 密码使用bcrypt加密存储

---

**更多详情请查看完整文档**
