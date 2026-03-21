# 统一任务管理系统 - 使用指南

## 概述

统一任务管理系统将文案生成任务和图片生成任务合并到一个统一的任务表中，提供一致的API接口和更好的任务管理能力。

## 数据库迁移

### 1. 执行迁移脚本

```bash
# 在服务器上执行
mysql -u root -p electric_ai_tool < migrate_20260321015117_unified_tasks.sql
```

### 2. 验证迁移

```sql
-- 查看新表结构
DESC unified_tasks_tab;

-- 查看迁移的数据
SELECT COUNT(*) FROM unified_tasks_tab;
SELECT task_type, COUNT(*) FROM unified_tasks_tab GROUP BY task_type;
```

### 3. 备份并删除旧表（可选，建议在确认无误后执行）

```sql
-- 重命名旧表作为备份
RENAME TABLE copywriting_tasks_tab TO copywriting_tasks_tab_backup_20260321;
RENAME TABLE tasks_tab TO tasks_tab_backup_20260321;

-- 或者直接删除（谨慎！）
-- DROP TABLE IF EXISTS copywriting_tasks_tab;
-- DROP TABLE IF EXISTS tasks_tab;
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

1. **数据迁移**：执行迁移脚本后，旧数据会被复制到新表，但旧表不会自动删除
2. **向后兼容**：旧的API接口仍然可用，但建议逐步迁移到新接口
3. **时间筛选**：支持按创建时间范围筛选任务
4. **用户筛选**：默认只显示当前用户的任务，可以通过`view_all=true`查看所有任务
5. **排序**：任务列表默认按创建时间倒序排序（最新的在前）

## 后续工作

1. [ ] 修改文案生成handler，同时写入统一任务表
2. [ ] 修改图片生成handler，同时写入统一任务表
3. [ ] 添加任务详情页面
4. [ ] 添加任务筛选UI（按类型、状态、时间）
5. [ ] 废弃旧的任务API接口
6. [ ] 删除旧的任务表（在确认无误后）
