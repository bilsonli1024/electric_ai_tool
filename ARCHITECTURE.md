# Electric AI Tool - 系统架构说明

## 项目概述

这是一个基于 Gemini AI 的电商产品图片生成和管理系统，集成了用户管理、任务追踪、CDN存储等完整功能。

## 新增功能

### 1. 用户认证系统
- ✅ 用户注册和登录
- ✅ 基于会话的身份认证
- ✅ 密码bcrypt加密
- ✅ 会话自动过期（24小时）
- ✅ 用户信息管理

### 2. 任务管理系统
- ✅ 每次图片生成自动创建任务记录
- ✅ 任务状态跟踪（pending、processing、completed、failed）
- ✅ 任务历史版本管理
- ✅ 任务创建者记录
- ✅ 我的任务/全部任务视图切换

### 3. CDN图片管理
- ✅ 产品白底图自动上传到后端
- ✅ 后端统一上传到CDN
- ✅ 图片CDN链接持久化存储
- ✅ 支持本地存储模式（开发/无CDN环境）
- ✅ 图片类型分类（product、style_ref、generated）

### 4. 数据持久化
- ✅ 用户数据存储
- ✅ 任务数据存储
- ✅ 任务历史版本存储
- ✅ 产品分析结果存储（TEXT字段）
- ✅ 会话管理

### 5. 前端界面优化
- ✅ 新增导航栏
- ✅ 用户管理页面
- ✅ 任务中心页面
- ✅ 一键生图保留原有功能
- ✅ 响应式设计

## 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                     前端 (React + TS)                    │
├─────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐ │
│  │ 登录/注册 │  │ 一键生图 │  │ 任务中心 │  │用户管理 │ │
│  └──────────┘  └──────────┘  └──────────┘  └─────────┘ │
└────────────────────┬────────────────────────────────────┘
                     │ HTTP/JSON
┌────────────────────┴────────────────────────────────────┐
│                   后端 (Go + HTTP)                       │
├─────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐ │
│  │ 认证服务 │  │ 任务服务 │  │ AI服务   │  │CDN服务  │ │
│  └──────────┘  └──────────┘  └──────────┘  └─────────┘ │
└────┬──────────────────┬────────────────────────┬────────┘
     │                  │                        │
┌────┴────┐      ┌──────┴──────┐        ┌───────┴────────┐
│ MySQL   │      │ Gemini API  │        │  CDN / Local   │
└─────────┘      └─────────────┘        └────────────────┘
```

## 数据库设计

### users - 用户表
- id: 用户ID
- username: 用户名（唯一）
- email: 邮箱（唯一）
- password_hash: 密码哈希
- status: 账号状态
- created_at, updated_at, last_login_at

### sessions - 会话表
- id: 会话ID（UUID）
- user_id: 关联用户
- expires_at: 过期时间
- created_at: 创建时间

### tasks - 任务表
- id: 任务ID
- user_id: 创建者
- task_type: 任务类型（analyze、generate_image、edit_image、aplus_content）
- sku, keywords, selling_points: 任务参数
- status: 任务状态
- result_data: 结果数据（TEXT类型，存储JSON）
- error_message: 错误信息
- created_at, updated_at

### task_history - 任务历史表
- id: 历史记录ID
- task_id: 关联任务
- user_id: 创建者
- version: 版本号
- prompt: 生成提示词
- aspect_ratio: 宽高比
- product_images_urls: 产品图CDN链接（JSON数组）
- style_ref_image_url: 风格参考图CDN链接
- generated_image_url: 生成图片CDN链接
- edit_instruction: 编辑指令
- status: 状态
- created_at

### cdn_images - CDN图片表
- id: 图片记录ID
- user_id: 上传用户
- original_filename: 原始文件名
- cdn_url: CDN访问链接
- cdn_key: CDN存储key
- file_size: 文件大小
- mime_type: MIME类型
- image_type: 图片类型（product、style_ref、generated）
- created_at

## API设计

### 认证API
```
POST /api/auth/register
{
  "username": "user1",
  "email": "user@example.com",
  "password": "password123"
}

POST /api/auth/login
{
  "username": "user1",
  "password": "password123"
}

POST /api/auth/logout
Headers: Authorization: Bearer {session_id}

GET /api/auth/me
Headers: Authorization: Bearer {session_id}
```

### 任务API
```
POST /api/tasks/analyze
Headers: Authorization: Bearer {session_id}
{
  "sku": "ABC123",
  "keywords": "关键词",
  "sellingPoints": "卖点",
  "competitorLink": "竞品链接"
}

POST /api/tasks/generate-image
Headers: Authorization: Bearer {session_id}
{
  "prompt": "生成描述",
  "aspectRatio": "1:1",
  "productImages": ["data:image/png;base64,..."],
  "styleRefImage": "data:image/png;base64,..."
}

GET /api/tasks?limit=20&offset=0&type=generate_image
Headers: Authorization: Bearer {session_id}

GET /api/tasks/all?limit=20&offset=0
Headers: Authorization: Bearer {session_id}

GET /api/tasks/history?task_id=123&limit=20&offset=0
Headers: Authorization: Bearer {session_id}
```

## 技术特性

### 后端
1. **连接池模式**：数据库使用连接池，最大50个连接
2. **中间件架构**：CORS、认证中间件
3. **错误处理**：统一错误响应格式
4. **安全性**：bcrypt密码加密、会话验证
5. **灵活存储**：支持CDN或本地存储

### 前端
1. **类型安全**：TypeScript全覆盖
2. **状态管理**：React Hooks
3. **UI组件**：TailwindCSS + 自定义组件
4. **动画效果**：Motion动画库
5. **响应式设计**：移动端适配

## 部署流程

1. **准备环境**
   - MySQL 8.0+
   - Go 1.24+
   - Node.js 18+

2. **初始化数据库**
   ```bash
   cd go_server
   ./init_db.sh
   ```

3. **配置环境**
   - 复制 `.env.example` 为 `.env`
   - 填入Gemini API Key和数据库配置

4. **快速启动**
   ```bash
   ./start.sh
   ```

5. **访问系统**
   - 开发模式：http://localhost:5173
   - 生产模式：http://localhost:3002

## 安全建议

1. 定期更换数据库密码
2. 使用HTTPS部署到生产环境
3. 配置CDN访问权限
4. 定期备份数据库
5. 监控API调用频率

## 扩展建议

1. 添加管理员角色和权限管理
2. 实现任务队列和异步处理
3. 添加图片审核功能
4. 集成更多CDN服务商
5. 添加任务导出功能
6. 实现WebSocket实时通知

## 维护指南

### 日志查看
```bash
# 后端日志
tail -f go_server/logs/app.log

# 数据库慢查询
mysql -u root -p -e "SHOW FULL PROCESSLIST;"
```

### 数据备份
```bash
mysqldump -u root -p electric_ai_tool > backup_$(date +%Y%m%d).sql
```

### 性能监控
- 监控数据库连接数
- 监控API响应时间
- 监控CDN使用量
- 监控磁盘空间

## 联系方式

如有问题，请查看 DEPLOYMENT.md 或提交 Issue。
