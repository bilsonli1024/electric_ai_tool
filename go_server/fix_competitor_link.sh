#!/bin/bash

# 快速修复 competitor_link 字段长度问题
# 只执行 migrate_20260321103751_fix_competitor_link_length.sql

echo "========================================="
echo "修复 competitor_link 字段长度"
echo "========================================="
echo ""

read -p "数据库用户名 [root]: " DB_USER
DB_USER=${DB_USER:-root}

read -sp "数据库密码: " DB_PASS
echo ""

read -p "数据库名 [electric_ai_tool]: " DB_NAME
DB_NAME=${DB_NAME:-electric_ai_tool}

echo ""
echo "执行迁移..."

mysql -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" <<EOF
ALTER TABLE tasks_tab 
MODIFY COLUMN competitor_link TEXT COMMENT '竞品链接，支持长URL';
EOF

if [ $? -eq 0 ]; then
    echo "✅ 修复成功！competitor_link 字段已改为 TEXT 类型"
else
    echo "❌ 修复失败，请检查数据库连接和权限"
    exit 1
fi
