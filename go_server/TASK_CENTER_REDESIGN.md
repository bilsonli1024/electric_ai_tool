# 任务中心重新设计方案

## 设计思路

采用**统一底表 + 详细表**的架构：

### 1. 统一底表（task_center_tab）
存储所有任务的基本信息，字段包括：
- `id`: 自增主键
- `task_id`: 全局唯一任务ID（格式：taskType_timestamp_randomString）
- `task_type`: 任务类型（copywriting/image）
- `task_status`: 统一状态（pending/ongoing/completed/failed）
- `operator`: 创建者邮箱
- `ctime`: 创建时间（秒级时间戳）
- `mtime`: 更新时间（秒级时间戳）

### 2. 详细表
每种任务类型有自己的详细表：

#### copywriting_tasks_tab（文案生成详细表）
- `task_id`: 关联底表的task_id（不是外键）
- `competitor_urls`: 竞品链接JSON数组
- `analysis_result`: AI初始分析结果
- `user_selected_data`: 用户选择后的数据
- `product_details`: 产品详情
- `generated_copy`: 生成的文案
- `analyze_model/generate_model`: 使用的AI模型
- `error_message`: 错误信息

#### tasks_tab（图片生成详细表）
- `task_id`: 关联底表的task_id（不是外键）
- `sku/keywords/selling_points`: 产品信息
- `competitor_link`: 竞品链接
- `copywriting_task_id`: 关联的文案生成任务（如果有）
- `generate_model/aspect_ratio`: 生成配置
- `result_data/generated_image_urls`: 生成结果
- `error_message`: 错误信息

## 数据库迁移

已创建迁移文件：`migrations/migrate_20260321112806_redesign_task_center.sql`

执行步骤：
```bash
cd go_server
mysql -u root -p electric_ai_tool < migrations/migrate_20260321112806_redesign_task_center.sql
```

注意：
- 旧表会被重命名为 `*_old`
- 由于项目未上线，如无重要数据可直接执行
- 如需迁移旧数据，需手动编写迁移脚本

## 代码实现

### 已完成：

1. **数据模型** (`models/task_center.go`)
   - `TaskCenterBase`: 底表模型
   - `CopywritingTaskDetail`: 文案详细模型
   - `ImageTaskDetail`: 图片详细模型
   - `TaskCenterDetail`: 联合详情模型

2. **Service层**
   - `TaskCenterService`: 统一底表操作
     - `GenerateTaskID()`: 生成全局唯一task_id
     - `CreateBaseTask()`: 创建底表记录
     - `UpdateTaskStatus()`: 更新任务状态
     - `GetTasks()`: 获取任务列表（支持筛选）
     - `GetTaskDetail()`: 获取完整任务详情
   
   - `CopywritingTaskService`: 文案详细表操作
     - `CreateTask()`: 创建详细记录
     - `SaveAnalysisResult()`: 保存分析结果
     - `SaveUserSelectedData()`: 保存用户选择
     - `SaveGeneratedCopy()`: 保存生成文案
   
   - `ImageTaskService`: 图片详细表操作
     - `CreateTask()`: 创建详细记录
     - `SaveResultData()`: 保存生成结果

### 待完成：

1. **修改Handler层**
   需要修改：
   - `CopywritingHandler.AnalyzeCompetitors()`
   - `CopywritingHandler.GenerateCopy()`
   - `TaskHandler.GenerateImageWithTask()`
   
   修改要点：
   ```go
   // 1. 生成task_id
   taskID := taskCenterService.GenerateTaskID("copywriting")
   
   // 2. 获取operator（用户邮箱）
   user, _ := authService.GetUserByID(userID)
   operator := user.Email
   
   // 3. 创建底表记录
   taskCenterService.CreateBaseTask(taskID, "copywriting", operator)
   
   // 4. 创建详细表记录
   copywritingTaskService.CreateTask(taskID, competitorURLs, model)
   
   // 5. 执行业务逻辑（分析/生成）
   // ...
   
   // 6. 更新状态
   taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusOngoing)
   
   // 7. 保存结果
   copywritingTaskService.SaveAnalysisResult(taskID, result)
   taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusCompleted)
   ```

2. **创建TaskCenterHandler**
   提供任务中心API：
   - `GET /api/task-center/list`: 获取任务列表
   - `GET /api/task-center/detail?task_id=xxx`: 获取任务详情

3. **前端TaskCenter组件**
   - 从新API获取数据
   - 添加详情按钮
   - 实现跳转逻辑

4. **前端页面参数恢复**
   - 文案生成页面：从task_id恢复数据
   - 图片生成页面：从task_id恢复数据

## 任务状态映射

旧状态（copywriting_tasks） -> 新状态：
- 0(分析中) -> ongoing
- 1(分析完成) -> ongoing
- 2(生成中) -> ongoing
- 3(已完成) -> completed
- 10(分析失败) -> failed
- 11(生成失败) -> failed

## API示例

### 创建文案任务
```
POST /api/copywriting/analyze
{
  "urls": ["..."],
  "model": "gemini"
}

Response:
{
  "task_id": "copywriting_1732112806_a3f2b1c4",
  "data": {...}
}
```

### 获取任务列表
```
GET /api/task-center/list?operator=user@example.com&limit=20&offset=0

Response:
{
  "data": [
    {
      "id": 1,
      "task_id": "copywriting_1732112806_a3f2b1c4",
      "task_type": "copywriting",
      "task_status": "completed",
      "operator": "user@example.com",
      "ctime": 1732112806,
      "mtime": 1732112900
    }
  ],
  "total": 50
}
```

### 获取任务详情
```
GET /api/task-center/detail?task_id=copywriting_1732112806_a3f2b1c4

Response:
{
  "id": 1,
  "task_id": "copywriting_1732112806_a3f2b1c4",
  "task_type": "copywriting",
  "task_status": "completed",
  "operator": "user@example.com",
  "ctime": 1732112806,
  "mtime": 1732112900,
  "detail_data": {
    "competitor_urls": "[...]",
    "analysis_result": "{...}",
    "user_selected_data": "{...}",
    "generated_copy": "{...}",
    ...
  }
}
```

## 下一步行动

1. 执行数据库迁移
2. 在main.go中初始化新的Service
3. 修改Handler层的任务创建逻辑
4. 创建TaskCenterHandler和Domain
5. 更新前端TaskCenter组件
6. 实现前端页面的参数恢复逻辑
