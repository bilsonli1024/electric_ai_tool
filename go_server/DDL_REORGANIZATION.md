# DDL 文件重组说明

## 变更内容

### 1. 文件结构调整

**之前：**
```
go_server/
├── schema.sql          # 完整数据库结构
└── migrate.sql         # 零散的迁移文件
```

**现在：**
```
go_server/
├── migrations/
│   ├── README.md                                          # 迁移文件使用说明
│   ├── schema.sql                                         # 完整数据库结构（用于新安装）
│   └── migrate_20260321103751_fix_competitor_link_length.sql  # 增量迁移文件
├── run_migrations.sh   # 自动执行迁移脚本
└── init_db.sh          # 数据库初始化脚本（已更新）
```

### 2. 数据库字段修复

**问题：** `tasks_tab` 表的 `competitor_link` 字段为 `VARCHAR(512)`，无法存储长URL

**解决方案：** 修改为 `TEXT` 类型

**迁移文件：** `migrate_20260321103751_fix_competitor_link_length.sql`

```sql
ALTER TABLE tasks_tab 
MODIFY COLUMN competitor_link TEXT COMMENT '竞品链接，支持长URL';
```

## 使用方法

### 新项目初始化

```bash
cd go_server
./init_db.sh
```

或者手动执行：

```bash
mysql -u root -p < migrations/schema.sql
```

### 已有项目迁移

**方式一：使用自动迁移脚本（推荐）**

```bash
cd go_server
./run_migrations.sh
```

脚本会：
1. 检查数据库连接
2. 列出所有待执行的迁移文件
3. 按时间戳顺序执行迁移
4. 显示执行结果

**方式二：手动执行**

```bash
mysql -u root -p electric_ai_tool < migrations/migrate_20260321103751_fix_competitor_link_length.sql
```

## 迁移文件命名规范

后续所有 DDL 变更必须遵循以下规范：

```
migrate_YYYYMMDDHHMMSS_description.sql
```

**示例：**
- `migrate_20260321103751_fix_competitor_link_length.sql`
- `migrate_20260322120000_add_user_avatar_field.sql`
- `migrate_20260323153000_create_notifications_table.sql`

## 注意事项

1. **不要修改已执行的迁移文件**
2. **保持迁移文件的顺序性**（按时间戳排序）
3. **执行前务必备份数据库**
4. **先在测试环境验证**
5. **迁移文件应该包含完整的注释说明**

## 已执行的迁移

| 文件名 | 执行日期 | 描述 | 状态 |
|--------|---------|------|------|
| migrate_20260321103751_fix_competitor_link_length.sql | 2026-03-21 | 修复 competitor_link 字段长度限制 | ✅ 待执行 |

## 后续工作

如需添加新的数据库变更：

1. 生成时间戳：
   ```bash
   date +%Y%m%d%H%M%S
   ```

2. 创建迁移文件：
   ```bash
   touch migrations/migrate_YYYYMMDDHHMMSS_your_description.sql
   ```

3. 编写迁移内容（参考 README.md 中的格式规范）

4. 执行迁移：
   ```bash
   ./run_migrations.sh
   ```
