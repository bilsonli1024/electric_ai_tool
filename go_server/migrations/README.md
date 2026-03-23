# Database Migrations

## 当前架构

本项目使用单一的 `schema.sql` 文件管理数据库结构。

### 文件说明

- **`schema.sql`** - 完整的数据库结构定义
  - 包含所有表的DDL
  - 包含初始数据（管理员用户、默认角色权限）
  - 使用INT枚举和INT时间戳

### 初始化数据库

使用初始化脚本：

```bash
cd go_server
chmod +x init_db.sh
./init_db.sh
```

⚠️ **警告**: 此操作会删除并重建整个数据库，所有数据将丢失！

### 手动初始化

```bash
mysql -u root -p electric_ai_tool < migrations/schema.sql
```

### 数据库架构特点

1. **INT枚举**: 所有状态字段使用INT类型
2. **INT时间戳**: 所有时间字段使用UNIX时间戳
3. **统一命名**: 创建时间=ctime, 更新时间=mtime
4. **RBAC**: 完整的基于角色的权限管理
5. **任务中心**: 统一的任务管理架构

### 默认数据

- 管理员用户: `admin@gmail.com` / `123456`
- 默认角色: 超级管理员、普通用户
- 默认权限: 完整的权限树

### 枚举值说明

所有枚举值定义在 `go_server/models/enums.go`

主要枚举：
- 用户类型: 0=普通用户, 99=管理员
- 用户状态: 0=待审批, 1=正常, 2=已删除
- 任务类型: 1=文案生成, 2=图片生成
- 任务状态: 0=待处理, 1=进行中, 2=已完成, 3=失败
- AI模型: 1=Gemini, 2=GPT, 3=DeepSeek

### 注意事项

1. 这是一个破坏性更新，会删除所有现有数据
2. 确保在执行前备份重要数据
3. 初始化后务必修改管理员密码
4. 生产环境请谨慎操作

### 相关文档

- `../IMPLEMENTATION_COMPLETE.md` - 完整实现说明
- `../QUICK_START.md` - 快速开始指南
- `../REFACTOR_SUMMARY.md` - 重构总结
