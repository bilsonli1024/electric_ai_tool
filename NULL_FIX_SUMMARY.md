# NULL值扫描错误修复总结

## 错误信息
```
加载任务失败: sql: Scan error on column index 5, name "user_selected_data": converting NULL to string is unsupported
```

## 根本原因
Go的`database/sql`包在扫描数据时，无法将数据库中的NULL值直接转换为string类型。当字段值为NULL时，Scan操作会失败。

## 解决方案
在所有可能包含NULL值的字段上使用SQL的`COALESCE`函数，将NULL转换为空字符串：

```sql
COALESCE(column_name, '') 
```

## 修复的文件和方法

### 1. services/copywriting_task_service.go
**方法**: `GetTaskByID`

**修改前**:
```sql
SELECT id, task_id, competitor_urls, analysis_result, analyze_model,
       user_selected_data, product_details, generated_copy, generate_model,
       error_message, created_at, updated_at
FROM copywriting_tasks_tab
WHERE task_id = ?
```

**修改后**:
```sql
SELECT id, task_id, competitor_urls, 
       COALESCE(analysis_result, ''), COALESCE(analyze_model, ''),
       COALESCE(user_selected_data, ''), COALESCE(product_details, ''), 
       COALESCE(generated_copy, ''), COALESCE(generate_model, ''),
       COALESCE(error_message, ''), created_at, updated_at
FROM copywriting_tasks_tab
WHERE task_id = ?
```

### 2. services/image_task_service.go
**方法**: `GetTaskByID`

**修改前**:
```sql
SELECT id, task_id, sku, keywords, selling_points, competitor_link,
       copywriting_task_id, generate_model, aspect_ratio,
       result_data, generated_image_urls, error_message,
       created_at, updated_at
FROM tasks_tab
WHERE task_id = ?
```

**修改后**:
```sql
SELECT id, task_id, 
       COALESCE(sku, ''), COALESCE(keywords, ''), 
       COALESCE(selling_points, ''), COALESCE(competitor_link, ''),
       COALESCE(copywriting_task_id, ''), COALESCE(generate_model, ''), 
       COALESCE(aspect_ratio, ''),
       COALESCE(result_data, ''), COALESCE(generated_image_urls, ''), 
       COALESCE(error_message, ''),
       created_at, updated_at
FROM tasks_tab
WHERE task_id = ?
```

### 3. services/task_center_service.go
**方法**: `getCopywritingDetail`

**修改后**:
```sql
SELECT id, task_id, 
       COALESCE(task_name, '') as task_name,
       competitor_urls, 
       COALESCE(analysis_result, ''), COALESCE(analyze_model, ''),
       COALESCE(user_selected_data, ''), COALESCE(product_details, ''), 
       COALESCE(generated_copy, ''), COALESCE(generate_model, ''),
       COALESCE(error_message, ''), created_at, updated_at
FROM copywriting_tasks_tab
WHERE task_id = ?
```

**方法**: `getImageDetail`

**修改后**:
```sql
SELECT id, task_id, 
       COALESCE(sku, ''), COALESCE(keywords, ''), 
       COALESCE(selling_points, ''), COALESCE(competitor_link, ''),
       COALESCE(copywriting_task_id, ''), COALESCE(generate_model, ''), 
       COALESCE(aspect_ratio, ''),
       COALESCE(result_data, ''), COALESCE(generated_image_urls, ''), 
       COALESCE(error_message, ''),
       created_at, updated_at
FROM tasks_tab
WHERE task_id = ?
```

## 为什么某些字段不需要COALESCE？

以下字段不使用COALESCE的原因：

1. **主键字段** (`id`): 永远不会是NULL
2. **NOT NULL字段** (`task_id`, `created_at`, `updated_at`): 数据库定义为NOT NULL
3. **必填字段** (`competitor_urls`): 创建任务时必须提供

## 其他解决方案（未采用）

### 方案A: 使用sql.NullString
```go
var detail struct {
    UserSelectedData sql.NullString
    // ...
}

// 扫描后转换
if detail.UserSelectedData.Valid {
    task.UserSelectedData = detail.UserSelectedData.String
} else {
    task.UserSelectedData = ""
}
```

**缺点**: 需要定义中间结构体，代码更复杂

### 方案B: 修改数据库表定义
```sql
ALTER TABLE copywriting_tasks_tab 
MODIFY COLUMN user_selected_data TEXT DEFAULT '';
```

**缺点**: 
- 需要修改现有数据
- 影响所有历史记录
- 可能破坏现有逻辑（NULL和空字符串语义不同）

## 最佳实践建议

1. **新建表时**:
   - 对于可选字符串字段，设置`DEFAULT ''`
   - 避免字符串字段使用NULL（除非NULL有特殊业务含义）

2. **查询时**:
   - 对所有可能为NULL的字符串字段使用COALESCE
   - 对数值字段使用`COALESCE(field, 0)`
   - 对时间字段考虑是否需要默认值

3. **Go代码**:
   - 对于频繁查询的字段，考虑使用sql.NullXxx类型
   - 对于偶尔查询的字段，使用COALESCE更简单

## 测试验证

修复后，以下操作应该正常工作：

1. ✅ 查看新创建的任务详情（字段为NULL）
2. ✅ 查看进行中的任务详情（部分字段为NULL）
3. ✅ 查看已完成的任务详情（所有字段有值）
4. ✅ 任务中心列表页正常显示

## 编译和部署

```bash
# 1. 编译
cd /Users/bilson.li/work/personal/code/electric_ai_tool/go_server
go build

# 2. 重启服务
./start.sh

# 3. 测试
# 访问任务中心，点击任意任务的"查看详情"
```

## 相关问题

如果仍然遇到类似错误，检查：

1. 是否有其他Service方法也在查询这些表？
2. 是否有直接的SQL查询没有使用COALESCE？
3. 数据库表结构是否与model定义一致？

使用以下命令查找所有SELECT语句：
```bash
cd go_server
grep -r "SELECT.*FROM copywriting_tasks_tab" services/
grep -r "SELECT.*FROM tasks_tab" services/
```
