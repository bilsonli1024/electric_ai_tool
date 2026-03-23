# 快速开始指南 - 重构后第一次运行

## 当前状态

✅ **已完成的核心改造**:
1. 数据库schema完全重新设计（INT枚举 + INT时间戳）
2. 用户认证系统改造（注册不自动登录，等待审批）
3. 枚举API实现（`/api/domain/enums`）
4. 时间工具（东八区）
5. CDN配置选项

⚠️ **警告**: 这是一个破坏性更新，会删除所有现有数据！

---

## 第一步：初始化数据库

### 选项A：使用初始化脚本（推荐）

```bash
cd go_server
chmod +x init_db.sh
./init_db.sh
```

脚本会：
- 删除旧数据库 `electric_ai_tool`
- 创建新数据库和所有表
- 初始化管理员账号: `admin@gmail.com` / `123456`
- 初始化默认角色和权限

### 选项B：手动执行

```bash
cd go_server
mysql -u root -p electric_ai_tool < migrations/schema.sql
```

---

## 第二步：配置环境变量

```bash
cd go_server
cp .env.example .env
# 编辑 .env 文件，配置数据库和API密钥
```

重要配置：
```env
# 数据库
DB_HOST=localhost
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=electric_ai_tool

# AI模型
GEMINI_API_KEY=your_key_here

# 图片存储（默认本地）
IMAGE_STORAGE_TYPE=local
```

---

## 第三步：编译并运行后端

```bash
cd go_server
go mod tidy
go build
./electric_ai_tool_server
```

应该看到：
```
✅ 后端服务已启动，监听端口: 3002
```

---

## 第四步：测试基础功能

### 1. 测试健康检查

```bash
curl http://localhost:3002/api/health
```

### 2. 测试管理员登录

```bash
curl -X POST http://localhost:3002/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@gmail.com",
    "password": "123456"
  }'
```

应该返回：
```json
{
  "user": {
    "id": 1,
    "email": "admin@gmail.com",
    "username": "Administrator",
    "user_type": 99,
    "user_status": 1,
    ...
  },
  "session_id": "..."
}
```

### 3. 测试枚举API

```bash
curl -H "Authorization: Bearer <session_id>" \
  http://localhost:3002/api/domain/enums
```

应该返回所有枚举定义。

---

## 第五步：前端适配（重要！）

### 当前前端还未适配新的数据格式！

需要修改：

#### 1. 注册页面
- 添加"申请管理员"复选框
- 注册成功后跳转到登录页
- 显示"等待审批"提示

#### 2. 枚举管理
创建 `web/src/utils/enums.ts`:
```typescript
// 在登录后调用
fetch('/api/domain/enums')
  .then(res => res.json())
  .then(data => localStorage.setItem('enums', JSON.stringify(data)))

// 使用枚举
export function getEnumLabel(enumType: string, value: number): string {
  const enums = JSON.parse(localStorage.getItem('enums') || '{}')
  const items = enums[enumType] || []
  const item = items.find((i: any) => i.value === value)
  return item?.label || '未知'
}
```

#### 3. 时间格式化
创建 `web/src/utils/time.ts`:
```typescript
export function formatTimestamp(timestamp: number): string {
  return new Date(timestamp * 1000).toLocaleString('zh-CN', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}
```

#### 4. API调用适配
所有传递状态/模型的地方改用数字：
```typescript
// ❌ 旧代码
{ model: 'gemini' }

// ✅ 新代码
{ model: 1 }  // 1=Gemini
```

---

## 已知问题和待完成功能

### 🔴 高优先级（必须完成）

1. **Service层大量改造**
   - 所有SQL查询需要适配新字段
   - 需要修改的文件：
     - `copywriting_service.go`
     - `copywriting_task_service.go`
     - `image_task_service.go`
     - `task_center_service.go`
     - `unified_task_service.go`

2. **Handler层适配**
   - `copywriting_handler.go`
   - `auth_task_handlers.go`（图片生成部分）
   - `task_center_handler.go`

### 🟡 中优先级（功能增强）

3. **RBAC完整实现**
   - 权限检查中间件
   - 用户管理Service/Handler/前端
   - 角色管理Service/Handler/前端

4. **前端完整适配**
   - 枚举系统集成
   - 时间显示格式化
   - 所有组件状态字段改用数字

---

## 测试计划

### 后端测试
- [x] 数据库初始化
- [x] 管理员登录
- [x] 枚举API
- [ ] 文案生成流程
- [ ] 图片生成流程
- [ ] 任务中心查询

### 前端测试
- [ ] 注册流程（新）
- [ ] 登录流程
- [ ] 枚举加载
- [ ] 时间显示
- [ ] 文案生成
- [ ] 图片生成
- [ ] 任务中心

---

## 回滚方案

如果遇到问题需要回滚：

1. **备份现有数据库**（如果有）
   ```bash
   mysqldump -u root -p electric_ai_tool > backup_$(date +%Y%m%d).sql
   ```

2. **恢复数据库**
   ```bash
   mysql -u root -p electric_ai_tool < backup_YYYYMMDD.sql
   ```

3. **切换到旧代码分支**
   ```bash
   git checkout <previous_branch>
   ```

---

## 下一步工作

按优先级推荐顺序：

1. **测试当前可用功能**
   - 登录/登出
   - 枚举API
   - 健康检查

2. **改造一个完整模块**（建议先改造文案生成）
   - Service层
   - Handler层
   - 前端页面

3. **扩展到其他模块**
   - 照着第一个模块的模式改造
   - 图片生成
   - 任务中心

4. **实现RBAC功能**
   - 用户管理
   - 角色管理
   - 权限控制

---

## 获取帮助

遇到问题时：

1. 查看日志：`go_server/logs/`
2. 检查数据库：`mysql -u root -p electric_ai_tool`
3. 参考 `REFACTOR_SUMMARY.md` 了解完整架构
4. 查看 `migrations/schema.sql` 了解数据库结构

---

**重要提示**: 当前系统处于部分可用状态，只有用户认证、枚举API等基础功能可用。业务功能（文案生成、图片生成）需要继续改造Service层才能正常工作。
