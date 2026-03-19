# Electric AI Tool - 部署指南

## 系统要求

- Go 1.24+
- Node.js 18+
- MySQL 8.0+
- Gemini API Key

## 快速开始

### 1. 数据库配置

#### 方式一：使用初始化脚本（推荐）

```bash
cd go_server
./init_db.sh
```

按提示输入数据库信息，脚本会自动创建数据库并导入表结构。

#### 方式二：手动配置

```bash
# 登录MySQL
mysql -u root -p

# 创建数据库
CREATE DATABASE electric_ai_tool CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 导入表结构
mysql -u root -p electric_ai_tool < go_server/schema.sql
```

### 2. 环境配置

复制并编辑环境配置文件：

```bash
cd go_server
cp .env.example .env
```

编辑 `.env` 文件，填入实际配置：

```env
# Gemini API配置
GEMINI_API_KEY=your_actual_gemini_api_key

# 服务器端口
PORT=3002

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_mysql_password
DB_NAME=electric_ai_tool

# CDN配置（可选，如不配置则使用本地存储）
CDN_ENDPOINT=
CDN_BUCKET=
CDN_ACCESS_KEY=
CDN_SECRET_KEY=
```

### 3. 安装依赖

#### 后端

```bash
cd go_server
go mod tidy
```

#### 前端

```bash
cd web
npm install
```

### 4. 运行项目

#### 开发模式

在两个终端中分别运行：

```bash
# 终端1：启动后端
cd go_server
go run main.go

# 终端2：启动前端
cd web
npm run dev
```

访问：http://localhost:5173

#### 生产模式

```bash
# 构建前端
cd web
npm run build

# 启动后端（会自动服务前端静态文件）
cd ../go_server
go build -o server
./server
```

访问：http://localhost:3002

## 功能模块

### 1. 用户管理
- 用户注册
- 用户登录
- 会话管理
- 用户信息查看

### 2. 一键生图
- 产品分析
- 图片生成
- 图片编辑
- A+内容生成
- 所有操作自动创建任务记录

### 3. 任务中心
- 查看我的任务
- 查看全部任务
- 任务状态跟踪
- 任务历史版本管理

### 4. CDN图片管理
- 产品白底图自动上传CDN
- 生成图片自动保存CDN
- 图片链接持久化存储
- 支持本地存储模式（无需配置CDN）

## 数据库表说明

- `users`: 用户表
- `sessions`: 会话表
- `tasks`: 任务表
- `task_history`: 任务历史记录表（保存每次生成的版本）
- `cdn_images`: CDN图片记录表

## API接口

### 认证相关
- POST `/api/auth/register` - 用户注册
- POST `/api/auth/login` - 用户登录
- POST `/api/auth/logout` - 用户登出
- GET `/api/auth/me` - 获取当前用户信息

### 任务相关（需要登录）
- POST `/api/tasks/analyze` - 创建产品分析任务
- POST `/api/tasks/generate-image` - 创建图片生成任务
- GET `/api/tasks` - 获取我的任务列表
- GET `/api/tasks/all` - 获取全部任务列表
- GET `/api/tasks/history?task_id=xxx` - 获取任务历史记录

### 兼容接口（无需登录）
- POST `/api/analyze` - 产品分析（旧接口，兼容）
- POST `/api/generate-image` - 图片生成（旧接口，兼容）
- POST `/api/edit-image` - 图片编辑
- POST `/api/aplus-content` - A+内容生成

## 注意事项

1. 首次使用需要先注册账号
2. CDN配置为可选项，如不配置则图片保存在本地 `/tmp/cdn_uploads/` 目录
3. 数据库连接使用连接池模式，性能优化
4. 所有密码使用 bcrypt 加密存储
5. 会话有效期为24小时

## 故障排查

### 数据库连接失败
- 检查MySQL服务是否启动
- 检查 `.env` 文件中的数据库配置是否正确
- 检查数据库用户权限

### API调用失败
- 检查 Gemini API Key 是否有效
- 检查网络连接
- 查看后端日志输出

### 前端无法访问
- 确认前端已构建（生产模式）
- 确认后端服务已启动
- 检查端口是否被占用

## 技术栈

### 后端
- Go 1.24
- MySQL 8.0
- Gemini API
- bcrypt 密码加密

### 前端
- React 19
- TypeScript
- Vite
- TailwindCSS
- Motion (动画)

## 许可证

Apache-2.0
