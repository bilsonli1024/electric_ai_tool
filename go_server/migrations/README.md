# 数据库迁移文件说明

## 文件结构

```
migrations/
├── schema.sql                          # 完整的数据库结构（用于新安装）
└── migrate_YYYYMMDDHHMMSS_*.sql       # 增量迁移文件（按时间戳排序）
```

## 使用方法

### 1. 新项目初始化

对于全新的项目，直接执行 schema.sql：

```bash
mysql -u root -p < migrations/schema.sql
```

### 2. 已有项目迁移

对于已有数据的项目，按时间戳顺序执行迁移文件：

```bash
# 查看当前需要执行的迁移文件
ls -1 migrations/migrate_*.sql

# 逐个执行迁移（或使用脚本批量执行）
mysql -u root -p electric_ai_tool < migrations/migrate_20260321103751_fix_competitor_link_length.sql
```

### 3. 自动执行所有迁移

使用提供的迁移脚本：

```bash
cd /path/to/go_server
./run_migrations.sh
```

## 迁移文件命名规范

所有迁移文件必须遵循以下命名规范：

```
migrate_YYYYMMDDHHMMSS_description.sql
```

- `YYYYMMDDHHMMSS`: 时间戳（年月日时分秒）
- `description`: 简短的英文描述，使用下划线分隔单词

示例：
- `migrate_20260321103751_fix_competitor_link_length.sql`
- `migrate_20260322120000_add_user_avatar_field.sql`

## 迁移文件内容规范

每个迁移文件应该包含：

1. **注释块**：说明迁移目的、原因和日期
2. **USE 语句**：确保在正确的数据库中执行
3. **DDL 语句**：实际的数据库变更操作

示例：

```sql
-- 迁移: 添加用户头像字段
-- 原因: 支持用户个性化头像功能
-- 日期: 2026-03-22 12:00:00

USE electric_ai_tool;

ALTER TABLE users_tab 
ADD COLUMN avatar_url VARCHAR(512) COMMENT '用户头像URL';
```

## 注意事项

1. **不要修改已执行的迁移文件**：迁移文件一旦执行到生产环境，就不应该再修改
2. **保持顺序性**：迁移文件必须按时间戳顺序执行
3. **测试迁移**：在生产环境执行前，先在测试环境验证
4. **备份数据**：执行迁移前务必备份数据库
5. **向下兼容**：尽量保持数据结构的向下兼容性

## 已执行的迁移

### migrate_20260321103751_fix_competitor_link_length.sql
- **日期**: 2026-03-21 10:37:51
- **描述**: 修改 tasks_tab 表的 competitor_link 字段类型从 VARCHAR(512) 改为 TEXT
- **原因**: VARCHAR(512) 不足以存储长URL
