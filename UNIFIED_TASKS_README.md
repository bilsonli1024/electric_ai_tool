# 统一任务管理系统 - 使用指南

## 概述

统一任务管理系统将文案生成任务和图片生成任务合并到一个统一的任务表中，提供一致的API接口和更好的任务管理能力。

**注意**：由于项目尚未上线，所有表结构已合并到 `schema.sql` 中，直接执行 `schema.sql` 即可创建完整的数据库结构。

## 数据库初始化

### 方式一：全新安装（推荐）

```bash
# 直接执行schema.sql创建所有表
mysql -u root -p < go_server/schema.sql
```

### 方式二：已有旧数据需要迁移

如果你已经有 `copywriting_tasks_tab` 或 `tasks_tab` 的数据，需要迁移到 `unified_tasks_tab`：

```sql
-- 1. 从文案任务迁移
INSERT INTO unified_tasks_tab (
    user_id, task_name, task_type, status,
    task_config, analysis_result, generation_result,
    analyze_model, generate_model, error_message,
    created_at, updated_at
)
SELECT 
    user_id,
    task_name,
    'copywriting' as task_type,
    status,
    JSON_OBJECT(
        'competitor_urls', competitor_urls,
        'product_details', product_details
    ) as task_config,
    analysis_result,
    generated_copy as generation_result,
    analyze_model,
    generate_model,
    error_message,
    created_at,
    updated_at
FROM copywriting_tasks_tab;

-- 2. 从图片任务迁移
INSERT INTO unified_tasks_tab (
    user_id, task_name, task_type, status,
    task_config, analysis_result,
    analyze_model, generate_model, error_message,
    created_at, updated_at
)
SELECT 
    user_id,
    CONCAT('图片任务_', id) as task_name,
    'image' as task_type,
    status,
    JSON_OBJECT(
        'sku', sku,
        'keywords', keywords,
        'selling_points', selling_points,
        'competitor_link', competitor_link
    ) as task_config,
    result_data as analysis_result,
    analyze_model,
    generate_model,
    error_message,
    created_at,
    updated_at
FROM tasks_tab;

-- 3. 验证迁移结果
SELECT task_type, COUNT(*) FROM unified_tasks_tab GROUP BY task_type;
```

## API接口

### 统一任务接口

#### 1. 获取任务列表

```
GET /api/unified-tasks
```

**查询参数：**
- `limit`: 每页数量（默认20）
- `offset`: 偏移量（默认0）
- `task_type`: 任务类型筛选（copywriting, image）
- `status`: 状态筛选（0-11）
- `start_time`: 开始时间（RFC3339格式）
- `end_time`: 结束时间（RFC3339格式）
- `view_all`: 是否查看所有用户任务（true/false）

**响应示例：**
```json
{
  "data": [
    {
      "id": 1,
      "user_id": 2,
      "username": "bilsonli1024",
      "task_name": "春季新品文案",
      "task_type": "copywriting",
      "status": 3,
      "analyze_model": "gemini",
      "generate_model": "gemini",
      "created_at": "2026-03-21T01:00:00Z",
      "updated_at": "2026-03-21T01:05:00Z"
    }
  ],
  "total": 10
}
```

#### 2. 获取任务详情

```
GET /api/unified-tasks/detail?id=123
```

#### 3. 获取任务统计

```
GET /api/unified-tasks/statistics?view_all=false
```

**响应示例：**
```json
{
  "data": {
    "total_tasks": 50,
    "completed_tasks": 30,
    "processing_tasks": 15,
    "failed_tasks": 5,
    "copywriting_tasks": 25,
    "image_tasks": 25
  }
}
```

#### 4. 创建任务

```
POST /api/unified-tasks/create
```

**请求体：**
```json
{
  "task_name": "春季新品文案",
  "task_type": "copywriting",
  "task_config": {
    "competitor_urls": ["url1", "url2"],
    "product_details": {}
  },
  "analyze_model": "gemini",
  "generate_model": "gemini"
}
```

## 任务状态说明

| 状态码 | 说明 |
|--------|------|
| 0 | 分析中 |
| 1 | 分析完成/待生成 |
| 2 | 生成中 |
| 3 | 已完成 |
| 10 | 分析失败 |
| 11 | 生成失败 |

## 任务类型

- `copywriting`: 文案生成任务
- `image`: 图片生成任务

## 前端集成

前端已更新为使用新的统一API：

```typescript
// 获取任务列表
const response = await apiClient.getUnifiedTasks({
  limit: 20,
  offset: 0,
  view_all: false
});

// 获取统计信息
const stats = await apiClient.getUnifiedTaskStatistics(false);
```

## 注意事项

1. **数据库结构**：所有表结构已合并到 `schema.sql`，新项目直接执行即可
2. **旧表兼容**：`tasks_tab` 和 `copywriting_tasks_tab` 在 schema.sql 中标记为已废弃，但仍保留定义用于向后兼容
3. **推荐使用**：新功能请直接使用 `unified_tasks_tab` 表
4. **时间筛选**：支持按创建时间范围筛选任务
5. **用户筛选**：默认只显示当前用户的任务，可以通过`view_all=true`查看所有任务
6. **排序**：任务列表默认按创建时间倒序排序（最新的在前）

## 数据库表说明

### unified_tasks_tab (推荐使用)
- **用途**：统一的任务管理表，支持文案和图片两种任务类型
- **特点**：
  - 使用 JSON 字段存储灵活的任务配置
  - 完整的索引支持，查询性能优秀
  - 按创建时间倒序排序
  - 支持多维度筛选

### tasks_tab (已废弃)
- **状态**：保留用于向后兼容，新项目不应使用
- **说明**：原图片生成任务表

### copywriting_tasks_tab (已废弃)
- **状态**：保留用于向后兼容，新项目不应使用
- **说明**：原文案生成任务表

## 后续工作

1. [ ] 修改文案生成handler，同时写入统一任务表
2. [ ] 修改图片生成handler，同时写入统一任务表
3. [ ] 添加任务详情页面
4. [ ] 添加任务筛选UI（按类型、状态、时间）
5. [ ] 废弃旧的任务API接口
6. [ ] 删除旧的任务表（在确认无误后）
