# 任务中心优化更新

## 更新时间
2026-03-22

## 问题修复清单

### ✅ 问题1：添加创建者列和筛选功能
- **前端TaskCenter组件**：
  - 添加了创建者筛选输入框
  - 添加了开始时间和结束时间筛选
  - 创建者列始终显示（不再只在"全部任务"模式显示）
  - 添加"清空筛选"按钮

- **后端API**：
  - 支持`operator`参数筛选创建者
  - 支持`start_time`和`end_time`参数筛选时间范围

### ✅ 问题2：优化HTTP请求（去重）
- **前端变更**：
  - 将`getUnifiedTasks`替换为`getTaskCenterTasks`
  - 确保只发送一次请求到`/api/task-center/list`

### ✅ 问题3：使用page_size和page_no参数
- **前端**：
  - `pageSize = 20`, `pageNo = 1`（从1开始）
  - API调用使用`page_size`和`page_no`参数

- **后端**：
  - 支持`page_size`和`page_no`参数
  - 同时保留对旧参数`limit`和`offset`的兼容支持
  - 转换逻辑：`offset = (page_no - 1) * page_size`

### ✅ 问题4：任务名称/SKU显示逻辑
- **前端**：
  - 表头改为"任务名称/SKU"
  - 使用`getTaskName()`函数：
    - 文案生成任务 → 显示`task_name`
    - 图片生成任务 → 显示`sku`

- **后端**：
  - 添加`TaskCenterListItem`模型（包含`task_name`和`sku`字段）
  - `GetTasks`方法使用LEFT JOIN获取详细表数据
  - 创建迁移脚本`migrate_20260322012001_add_task_name.sql`
  - 更新`CopywritingTaskService.CreateTask`接受`taskName`参数

### ✅ 问题5：任务ID格式优化
**新格式**：`缩写 + yyyyMMddHHmmss + 5位随机数字`

- **文案生成任务**：`CP20260322015030 12345`
- **图片生成任务**：`IG20260322015030 67890`

**代码更改**：
- 修改`TaskCenterService.GenerateTaskID()`方法
- 使用`time.Now().Format("20060102150405")`生成时间
- 使用`rand.Intn(90000) + 10000`生成5位随机数字
- **注意**：历史数据保持旧格式，不受影响

### ✅ 问题6：修复NULL值扫描错误
**错误信息**：`sql: Scan error on column index 5, name "user_selected_data": converting NULL to string is unsupported`

**解决方案**：
- 在SQL查询中使用`COALESCE`函数将NULL转为空字符串
- 更新的Service方法：
  - `CopywritingTaskService.GetTaskByID()`
  - `ImageTaskService.GetTaskByID()`

## 文件变更清单

### 前端文件
1. **web/src/components/TaskCenter.tsx** - 完全重写
   - 添加筛选器UI（创建者、开始时间、结束时间）
   - 使用`page_size`和`page_no`
   - 修改任务名称/SKU显示逻辑
   - 创建者列始终显示

2. **web/src/services/api.ts**
   - 更新`getTaskCenterTasks`参数：支持`page_size`、`page_no`、`operator`

### 后端文件
1. **go_server/services/task_center_service.go**
   - 修改`GenerateTaskID()`：新的任务ID格式
   - 修改`GetTasks()`：返回`TaskCenterListItem`，包含task_name和sku

2. **go_server/services/copywriting_task_service.go**
   - 修改`CreateTask()`：添加taskName参数
   - 修改`GetTaskByID()`：使用COALESCE处理NULL值

3. **go_server/services/image_task_service.go**
   - 修改`GetTaskByID()`：使用COALESCE处理NULL值

4. **go_server/handlers/task_center_handler.go**
   - 支持`page_size`、`page_no`参数
   - 支持`operator`筛选参数
   - 保持对旧参数的兼容

5. **go_server/handlers/copywriting_handler.go**
   - 更新CreateTask调用，传递taskName

6. **go_server/models/task_center.go**
   - 添加`TaskCenterListItem`结构体

### 数据库迁移
**go_server/migrations/migrate_20260322012001_add_task_name.sql**
```sql
ALTER TABLE copywriting_tasks_tab 
ADD COLUMN task_name VARCHAR(255) DEFAULT '' COMMENT '任务名称' AFTER task_id;
```

## 部署步骤

### 1. 数据库迁移
```bash
cd /Users/bilson.li/work/personal/code/electric_ai_tool/go_server
./run_migrations.sh
```

或手动执行：
```bash
mysql -u <user> -p <database> < migrations/migrate_20260322012001_add_task_name.sql
```

### 2. 编译后端
```bash
cd go_server
bash -c 'source ~/.gvm/scripts/gvm && gvm use go1.24.6 && go build -o electric_ai_tool'
```

### 3. 重启后端服务
```bash
./start.sh
```

### 4. 前端无需重新安装依赖
前端代码已更新，刷新浏览器即可

## 测试清单

### 任务中心功能测试
- [ ] 访问任务中心，确认列表正确加载
- [ ] 确认显示：任务ID（新格式）、类型、任务名称/SKU、创建者、状态、创建时间
- [ ] 测试筛选：
  - [ ] 输入创建者邮箱筛选
  - [ ] 选择开始时间
  - [ ] 选择结束时间
  - [ ] 点击"清空筛选"
- [ ] 确认只发送一次API请求（不重复请求）
- [ ] 测试分页：点击上一页/下一页
- [ ] 点击"查看详情"，确认可以跳转到详情页

### 新任务创建测试
- [ ] 创建文案生成任务，确认task_id格式为`CP20260322...`
- [ ] 创建图片生成任务，确认task_id格式为`IG20260322...`
- [ ] 在任务中心列表查看新任务：
  - [ ] 文案任务显示任务名称
  - [ ] 图片任务显示SKU

### 详情页测试
- [ ] 点击文案生成任务的"查看详情"
- [ ] 确认不再报错NULL值问题
- [ ] 数据正确加载并显示

## API文档更新

### GET /api/task-center/list

**请求参数**：
- `page_size` (number, 可选): 每页数量，默认20
- `page_no` (number, 可选): 页码，从1开始，默认1
- `operator` (string, 可选): 创建者邮箱筛选
- `task_type` (string, 可选): 任务类型（copywriting/image）
- `task_status` (string, 可选): 任务状态（pending/ongoing/completed/failed）
- `start_time` (number, 可选): 开始时间（秒级时间戳）
- `end_time` (number, 可选): 结束时间（秒级时间戳）
- `view_all` (boolean, 可选): 是否查看所有任务，默认false

**兼容旧参数**：
- `limit` → `page_size`
- `offset` → 转换为`page_no`

**响应格式**：
```json
{
  "data": [
    {
      "id": 1,
      "task_id": "CP2026032201503012345",
      "task_type": "copywriting",
      "task_status": "completed",
      "operator": "user@example.com",
      "ctime": 1711072830,
      "mtime": 1711072850,
      "task_name": "春季新品文案",
      "sku": ""
    },
    {
      "id": 2,
      "task_id": "IG2026032201504567890",
      "task_type": "image",
      "task_status": "ongoing",
      "operator": "user@example.com",
      "ctime": 1711072845,
      "mtime": 1711072845,
      "task_name": "",
      "sku": "QL0905"
    }
  ],
  "total": 25
}
```

## 注意事项

1. **向后兼容**：
   - 旧的任务ID格式仍然有效
   - 同时支持`limit/offset`和`page_size/page_no`参数
   - NULL值使用COALESCE处理，确保不会崩溃

2. **任务ID唯一性**：
   - 新格式包含时间（精确到秒）+ 5位随机数
   - 理论上在同一秒内可以创建99999个任务
   - 实际并发很难达到冲突

3. **性能**：
   - `GetTasks`使用LEFT JOIN，对大数据量可能有影响
   - 建议在task_id字段上创建索引
   - 考虑对task_center_tab的ctime字段创建索引

4. **前端筛选**：
   - 时间筛选自动转换为秒级时间戳
   - 结束时间自动设为当天23:59:59

## 下一步计划

1. 完成图片生成页面的任务恢复功能
2. 实现任务状态只能往前的控制
3. 添加用户选择的实时保存功能
4. 考虑为任务状态变更添加日志记录
