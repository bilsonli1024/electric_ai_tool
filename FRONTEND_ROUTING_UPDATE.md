# 前端路由和任务中心更新说明

## 更新时间
2026-03-21

## 更新概述
本次更新实现了基于React Router的完整路由系统，修复了任务中心显示问题，并添加了任务详情页面跳转和状态恢复功能。

## 主要改动

### 1. 安装React Router
```json
// package.json
"dependencies": {
  "react-router-dom": "^6.28.0"
}
```

**操作步骤：**
```bash
cd web
npm install
```

### 2. 路由系统架构

#### 2.1 MainApp.tsx重构
- 引入`BrowserRouter`, `Routes`, `Route`, `useNavigate`, `useLocation`
- 实现路由映射：
  - `/` → 文案生成页（默认）
  - `/copywriting` → 文案生成页
  - `/image-generation` → 图片生成页
  - `/tasks` → 任务中心
  - `/user` → 用户管理
  - `/modeltest` → 联通性测试

#### 2.2 侧边栏导航
- 点击侧边栏按钮使用`navigate()`跳转
- 根据当前路径高亮对应菜单项

### 3. 任务中心（TaskCenter）改进

#### 3.1 显示修复
**问题：** 后端返回新字段（`task_id`, `operator`, `ctime`），但前端仍使用旧字段显示

**解决方案：**
- 更新`types/index.ts`，添加新字段类型定义
- 修改TaskCenter显示逻辑：
  ```typescript
  // 任务ID：优先显示task_id，回退到#id
  {task.task_id || `#${task.id}`}
  
  // 操作者：优先显示operator，回退到username
  {task.operator || task.username || '-'}
  
  // 创建时间：格式化秒级时间戳
  formatTimestamp(task.ctime || task.created_at)
  ```

#### 3.2 时间格式化
```typescript
const formatTimestamp = (timestamp: number | string | undefined) => {
  if (!timestamp) return '-';
  
  // 处理ISO字符串格式（旧格式）
  if (typeof timestamp === 'string') {
    const date = new Date(timestamp);
    if (!isNaN(date.getTime())) {
      return date.toLocaleString('zh-CN');
    }
    return timestamp;
  }
  
  // 处理秒级时间戳（新格式）
  const date = new Date(timestamp * 1000);
  return date.toLocaleString('zh-CN');
};
```

#### 3.3 任务详情跳转
- 添加"查看详情"按钮到每行
- 点击后根据任务类型跳转：
  ```typescript
  const handleViewDetail = (task: any) => {
    const taskType = task.task_type || task.type;
    const taskId = task.task_id || task.id;
    
    if (taskType === 'copywriting') {
      navigate(`/copywriting?task_id=${taskId}`);
    } else if (taskType === 'image') {
      navigate(`/image-generation?task_id=${taskId}`);
    }
  };
  ```

### 4. 文案生成页面（CopywritingGenerator）任务恢复

#### 4.1 URL参数监听
```typescript
useEffect(() => {
  const searchParams = new URLSearchParams(location.search);
  const taskIdParam = searchParams.get('task_id');
  
  if (taskIdParam) {
    loadTaskDetail(taskIdParam);
  }
}, [location.search]);
```

#### 4.2 任务详情加载
`loadTaskDetail`函数实现：
- 调用`apiClient.getTaskCenterDetail(taskId)`
- 验证任务类型
- 解析并恢复：
  - 竞品链接（JSON数组）
  - AI分析结果（keywords, sellingPoints, reviewInsights, imageInsights）
  - 用户选择数据（如果存在）
  - 产品详情
  - 生成的文案
- 根据任务状态设置当前步骤：
  - `pending` → competitors页
  - `ongoing` → configuration或result页
  - `completed` → result页

#### 4.3 加载状态UI
```typescript
{isLoadingTask && (
  <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center">
    <div className="bg-white rounded-2xl p-8 flex flex-col items-center gap-4">
      <Loader2 className="animate-spin text-orange-500" size={48} />
      <p className="text-lg font-medium">正在加载任务数据...</p>
    </div>
  </div>
)}
```

### 5. 类型定义更新

#### 5.1 扩展Task类型
```typescript
export interface Task {
  // 新字段
  task_id?: string;
  operator?: string;
  ctime?: number;
  mtime?: number;
  task_status?: string;
  
  // 旧字段（保持兼容）
  id: number;
  user_id?: number;
  username?: string;
  status?: string;
  created_at?: string;
  // ...
}
```

#### 5.2 新增类型
```typescript
export interface TaskCenterBase {
  id: number;
  task_id: string;
  task_type: string;
  task_status: string;
  operator: string;
  ctime: number;
  mtime: number;
}

export interface CopywritingTaskDetail { /* ... */ }
export interface ImageTaskDetail { /* ... */ }
export interface TaskCenterDetail { /* ... */ }
```

## 待实现功能

### 1. 图片生成页面任务恢复
类似文案生成页面，需要：
- 监听URL `task_id`参数
- 调用`getTaskCenterDetail` API
- 恢复表单状态

### 2. 任务状态只能往前
- 在已完成竞品分析后，禁用竞品链接输入
- 在文案生成完成后，配置页面改为只读
- 提示用户："任务已进入下一阶段，不可编辑之前的内容"

### 3. 用户选择保存
在文案配置页面：
- 用户每次toggle关键词/卖点时，自动调用API保存`user_selected_data`
- 确保跳转回任务时，显示用户的选择状态（已选和未选的都要显示）

### 4. Tab页URL路径
考虑为每个tab页分配子路径：
- `/copywriting` → 竞品分析输入
- `/copywriting/configuration?task_id=xxx` → 文案配置
- `/copywriting/result?task_id=xxx` → 生成结果

这样可以：
- 支持浏览器前进/后退
- 更好的书签支持
- 更清晰的页面隔离

## 测试清单

### 基础功能
- [ ] 运行`npm install`安装依赖
- [ ] 启动前端：`npm run dev`
- [ ] 访问 http://localhost:3000

### 任务中心测试
- [ ] 查看任务列表，确认显示：task_id, operator, 格式化的时间
- [ ] 测试"我的任务" / "全部任务"切换
- [ ] 点击"查看详情"按钮

### 任务跳转测试
- [ ] 文案生成任务 → 跳转到 `/copywriting?task_id=xxx`
- [ ] 图片生成任务 → 跳转到 `/image-generation?task_id=xxx`
- [ ] URL变化正确
- [ ] 页面加载任务数据

### 任务恢复测试（文案生成）
- [ ] pending状态：显示竞品分析页
- [ ] ongoing状态（已分析）：显示配置页，关键词和卖点正确显示
- [ ] completed状态：显示结果页
- [ ] 用户之前的选择被正确恢复

### 路由测试
- [ ] 点击侧边栏菜单，URL正确变化
- [ ] 浏览器前进/后退按钮工作正常
- [ ] 刷新页面后保持在当前路由

## 注意事项

1. **向后兼容**：代码同时支持新旧字段，确保与旧数据兼容
2. **错误处理**：所有JSON.parse都有try-catch包裹
3. **加载状态**：使用全屏遮罩显示加载状态，避免用户困惑
4. **类型安全**：使用TypeScript类型确保数据结构正确

## 相关文件

### 修改的文件
- `web/package.json` - 添加react-router-dom依赖
- `web/src/MainApp.tsx` - 路由系统实现
- `web/src/components/TaskCenter.tsx` - 显示修复和详情跳转
- `web/src/components/CopywritingGenerator.tsx` - 任务恢复实现
- `web/src/types/index.ts` - 类型定义更新

### 未修改但需要后续更新的文件
- `web/src/components/ImageGenerationPage.tsx` - 需要添加任务恢复
- `web/src/services/api.ts` - API接口已就绪

## 下一步计划

1. 完成图片生成页面的任务恢复
2. 实现任务状态只能往前的控制
3. 添加用户选择的实时保存
4. 考虑为Tab页添加子路由
5. 添加更多的错误提示和用户引导
