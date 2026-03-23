#!/bin/bash

# 数据库初始化脚本
# 用于重建electric_ai_tool数据库

# 数据库配置
DB_HOST="${DB_HOST:-localhost}"
DB_USER="${DB_USER:-root}"
DB_PASSWORD="${DB_PASSWORD:-}"
DB_NAME="electric_ai_tool"

echo "======================================"
echo "Electric AI Tool 数据库初始化"
echo "======================================"
echo "Host: $DB_HOST"
echo "User: $DB_USER"
echo "Database: $DB_NAME"
echo "======================================"
echo ""

# 检查MySQL是否可用
if ! command -v mysql &> /dev/null; then
    echo "❌ MySQL客户端未安装"
    exit 1
fi

# 确认操作
read -p "⚠️  警告：此操作将删除并重建数据库 $DB_NAME，所有数据将丢失！是否继续？(yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "❌ 操作已取消"
    exit 0
fi

echo ""
echo "📦 开始初始化数据库..."
echo ""

# 执行SQL文件
if [ -z "$DB_PASSWORD" ]; then
    mysql -h "$DB_HOST" -u "$DB_USER" < migrations/schema.sql
else
    mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASSWORD" < migrations/schema.sql
fi

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ 数据库初始化成功！"
    echo ""
    echo "📋 已创建的表："
    echo "  - users_tab                  (用户表)"
    echo "  - roles_tab                  (角色表)"
    echo "  - permissions_tab            (权限表)"
    echo "  - user_roles_tab             (用户角色关系表)"
    echo "  - role_permissions_tab       (角色权限关系表)"
    echo "  - task_center_tab            (任务中心底表)"
    echo "  - copywriting_tasks_tab      (文案生成任务表)"
    echo "  - tasks_tab                  (图片生成任务表)"
    echo ""
    echo "👤 默认管理员账号："
    echo "  Email: admin@gmail.com"
    echo "  Password: 123456"
    echo "  (请登录后立即修改密码)"
    echo ""
else
    echo ""
    echo "❌ 数据库初始化失败！"
    echo "请检查MySQL连接和权限"
    exit 1
fi
