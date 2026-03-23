# Electric AI Tool - 架构重构总结

## 重构概览

本次重构是一个**重大架构升级**，主要目标：
1. 标准化数据类型（状态用INT枚举，时间用UNIX时间戳）
2. 引入完整的RBAC权限管理系统
3. 统一枚举管理，前后端分离
4. 规范化时间处理（统一使用东八区）

---

## 已完成的工作 ✅

### 1. 数据库架构重新设计

**文件**: `go_server/migrations/schema.sql`

#### 核心变更：
- ✅ 所有状态字段改为 `INT` 类型（枚举值）
- ✅ 所有时间字段改为 `INT` 类型（UNIX时间戳）
- ✅ 时间字段统一命名为 `ctime` (创建时间) 和 `mtime` (更新时间)

#### 新增表结构：

**用户管理表**:
- `users_tab`: 用户表
  - 新增: `user_type` (0=普通用户, 99=管理员)
  - 新增: `user_status` (0=待审批, 1=正常, 2=已删除)
  - 密码字段简化为 `password`
  
- `roles_tab`: 角色表
- `permissions_tab`: 权限表
- `user_roles_tab`: 用户角色关系表
- `role_permissions_tab`: 角色权限关系表

**任务管理表**:
- `task_center_tab`: 任务中心底表
  - `task_type`: INT (1=文案生成, 2=图片生成)
  - `task_status`: INT (0=待处理, 1=进行中, 2=已完成, 3=失败)
  
- `copywriting_tasks_tab`: 文案生成任务表
  - `detail_status`: INT (0=待处理, 1=分析中, 2=分析完成, 3=生成中, 4=已完成, 5=失败)
  - `analyze_model`: INT (1=Gemini, 2=GPT, 3=DeepSeek)
  - `generate_model`: INT (1=Gemini, 2=GPT, 3=DeepSeek)
  
- `tasks_tab`: 图片生成任务表
  - `detail_status`: INT (0=待处理, 1=生成中, 2=已完成, 3=失败)
  - `generate_model`: INT (1=Gemini, 2=GPT, 3=DeepSeek)

#### 初始数据：
- ✅ 默认管理员用户: `admin@gmail.com` / `123456`
- ✅ 默认角色: 超级管理员、普通用户
- ✅ 默认权限树: 包含所有功能模块的权限配置

---

### 2. Go后端Models层

#### 新增文件：

**`models/enums.go`** - 枚举定义
```go
// 用户枚举
const UserTypeNormal = 0
const UserTypeAdmin = 99
const UserStatusPendingApproval = 0
const UserStatusNormal = 1
const UserStatusDeleted = 2

// 任务枚举
const TaskTypeCopywriting = 1
const TaskTypeImage = 2
const TaskStatusPending = 0
const TaskStatusOngoing = 1
const TaskStatusCompleted = 2
const TaskStatusFailed = 3

// AI模型枚举
const ModelGemini = 1
const ModelGPT = 2
const ModelDeepSeek = 3

// ... 以及对应的转字符串函数
```

**`models/rbac.go`** - RBAC数据结构
```go
type Role struct { ... }
type Permission struct { ... }
type UserRole struct { ... }
type RolePermission struct { ... }
type PermissionTreeNode struct { ... }
```

#### 更新文件：

**`models/user.go`** - 用户模型
- 使用INT类型的 `user_type`, `user_status`
- 使用INT类型的 `ctime`, `mtime`
- 简化密码字段为 `password`

**`models/task_center.go`** - 任务中心模型
- 所有状态字段改为INT
- 所有时间字段改为INT (ctime, mtime)
- 所有模型字段改为INT

---

### 3. 工具函数

**`utils/time.go`** - 时间处理工具
```go
// 东八区时间转换
func GetCurrentTimestamp() int64
func TimestampToString(timestamp int64) string  // -> "2026-03-23 12:00:00"
func StringToTimestamp(timeStr string) (int64, error)
func TimestampToTime(timestamp int64) time.Time
```

---

### 4. 枚举API

**`handlers/enum_handler.go`** + **`domain/enum_domain.go`**

提供统一枚举接口：`GET /api/domain/enums`

返回格式：
```json
{
  "user_types": [{"value": 0, "label": "普通用户"}, {"value": 99, "label": "管理员"}],
  "user_statuses": [...],
  "task_types": [...],
  "models": [...],
  ...
}
```

---

### 5. 配置文件

**`.env.example`** - 新增图片存储配置
```env
IMAGE_STORAGE_TYPE=local    # local或cdn
CDN_ENDPOINT=...
CDN_BUCKET=...
```

---

### 6. 数据库初始化脚本

**`init_db.sh`** - 一键初始化数据库
```bash
chmod +x init_db.sh
./init_db.sh
```

---

## 还需要完成的工作 🚧

### 关键改造（优先级：高）

#### 1. **AuthService 改造**
- [ ] 修改注册逻辑：不自动登录，返回"等待审批"提示
- [ ] 修改用户状态检查逻辑
- [ ] 适配新的User模型（使用INT和时间戳）
- [ ] 密码加密逻辑调整

#### 2. **所有Service层改造**
需要更新所有数据库查询，使用新的INT字段：
- [ ] `copywriting_service.go`
- [ ] `copywriting_task_service.go` 
- [ ] `image_task_service.go`
- [ ] `task_center_service.go`
- [ ] 所有涉及时间的字段改用 `ctime`, `mtime`
- [ ] 所有涉及状态的字段改用INT
- [ ] 所有涉及模型的字段改用INT

#### 3. **所有Handler层改造**
- [ ] `auth_handlers.go` - 用户认证相关
- [ ] `copywriting_handler.go` - 文案生成相关
- [ ] `auth_task_handlers.go` - 图片生成相关
- [ ] `task_center_handler.go` - 任务中心相关
- [ ] 所有接口的请求/响应参数适配

#### 4. **权限中间件实现**
- [ ] 创建 `middleware/permission.go`
- [ ] 实现基于RBAC的权限检查
- [ ] 权限缓存机制
- [ ] 应用到需要权限控制的路由

#### 5. **RBAC Service层**
- [ ] `services/role_service.go` - 角色管理
- [ ] `services/permission_service.go` - 权限管理
- [ ] `services/user_management_service.go` - 用户管理（含审批）

#### 6. **RBAC Handler层**
- [ ] `handlers/role_handler.go` - 角色管理API
- [ ] `handlers/permission_handler.go` - 权限管理API
- [ ] `handlers/user_management_handler.go` - 用户管理API

---

### 前端改造（优先级：高）

#### 1. **枚举管理**
- [ ] 创建 `src/utils/enums.ts` - 枚举管理工具
- [ ] 在用户登录/刷新时调用 `/api/domain/enums`
- [ ] 缓存枚举到localStorage或Context
- [ ] 提供枚举转换函数 `getEnumLabel(type, value)`

#### 2. **时间格式化**
- [ ] 创建 `src/utils/time.ts` - 时间工具
- [ ] 统一显示格式: `2026-03-23 12:00:00`
- [ ] 前后端传输使用UNIX时间戳

#### 3. **注册流程修改**
- [ ] 注册页面添加"申请管理员"选项
- [ ] 注册成功后跳转到登录页，提示"等待审批"
- [ ] 不再自动登录

#### 4. **新增页面**
- [ ] **用户管理页面** (仅管理员可见)
  - 用户列表
  - 审批功能
  - 编辑/删除用户
  - 分配角色
  
- [ ] **角色管理页面** (仅管理员可见)
  - 角色列表
  - 创建/编辑/删除角色
  - 权限分配（树形选择）

#### 5. **权限控制**
- [ ] 路由守卫：检查用户权限
- [ ] 菜单显示：根据权限过滤
- [ ] 按钮权限：根据权限显示/隐藏

#### 6. **所有组件适配**
- [ ] 所有传递状态的地方改用INT
- [ ] 所有显示状态的地方使用枚举转换
- [ ] 所有显示时间的地方格式化为东八区
- [ ] 所有传递模型参数的地方改用INT

---

### Main.go 路由注册

**`main.go`** - 需要注册新的域：
```go
// 添加
enumDomain := domain.NewEnumDomain()
enumDomain.RegisterRoutes(authMiddleware)

// 后续添加
userMgmtDomain := domain.NewUserManagementDomain(...)
userMgmtDomain.RegisterRoutes(authMiddleware)

roleMgmtDomain := domain.NewRoleManagementDomain(...)
roleMgmtDomain.RegisterRoutes(authMiddleware)
```

---

## 数据迁移注意事项 ⚠️

### 这是一个破坏性变更！

如果你有现有数据：
1. **备份现有数据库**
2. 运行 `init_db.sh` 将**删除并重建**所有表
3. 需要编写数据迁移脚本，将旧数据转换为新格式：
   - 字符串状态 -> INT枚举值
   - `created_at`/`updated_at` (time.Time) -> `ctime`/`mtime` (INT)
   - 字符串模型 -> INT枚举值

### 如果是新项目：
直接运行初始化脚本即可：
```bash
cd go_server
chmod +x init_db.sh
./init_db.sh
```

---

## 测试计划

### 后端测试（优先）
1. [ ] 运行数据库初始化
2. [ ] 测试枚举API: `curl http://localhost:3002/api/domain/enums`
3. [ ] 测试管理员登录
4. [ ] 逐个测试改造后的API

### 前端测试
1. [ ] 枚举加载
2. [ ] 时间显示
3. [ ] 注册流程
4. [ ] 权限显示

---

## 预估工作量

- **后端Service/Handler改造**: 2-3天
- **RBAC完整实现**: 1-2天
- **前端枚举&时间适配**: 1天
- **前端RBAC页面**: 2-3天
- **测试&修复**: 1-2天

**总计**: 7-11天全职开发时间

---

## 下一步建议

### 方案A：逐步推进（推荐）
1. 先运行数据库初始化
2. 改造AuthService（用户登录/注册）
3. 改造一个完整模块（如文案生成）端到端测试
4. 依次改造其他模块
5. 最后实现RBAC界面

### 方案B：并行开发
1. 数据库先行
2. 后端和前端分别独立推进
3. 最后集成测试

---

## 当前项目状态

📊 **进度**: 约 30% 完成

✅ **已完成**:
- 数据库架构设计
- 枚举系统基础
- 时间工具
- RBAC数据模型

🚧 **进行中**:
- 需要开始Service/Handler改造

⏳ **待开始**:
- RBAC业务逻辑实现
- 前端完整适配
- 用户/角色管理界面

---

**注意**: 这是一个大型重构，建议在独立分支进行，充分测试后再合并到主分支。
