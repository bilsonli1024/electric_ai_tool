#!/bin/bash

# 数据库迁移脚本
# 自动执行所有未执行的迁移文件

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="$SCRIPT_DIR/migrations"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "数据库迁移工具"
echo "========================================="
echo ""

# 检查 migrations 目录
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo -e "${RED}❌ 错误: migrations 目录不存在${NC}"
    exit 1
fi

# 读取数据库配置
read -p "数据库用户名 [root]: " DB_USER
DB_USER=${DB_USER:-root}

read -sp "数据库密码: " DB_PASS
echo ""

read -p "数据库名 [electric_ai_tool]: " DB_NAME
DB_NAME=${DB_NAME:-electric_ai_tool}

read -p "数据库主机 [localhost]: " DB_HOST
DB_HOST=${DB_HOST:-localhost}

echo ""
echo "连接信息:"
echo "  用户: $DB_USER"
echo "  数据库: $DB_NAME"
echo "  主机: $DB_HOST"
echo ""

# 测试数据库连接
echo "测试数据库连接..."
if ! mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" -e "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}❌ 数据库连接失败${NC}"
    exit 1
fi
echo -e "${GREEN}✅ 数据库连接成功${NC}"
echo ""

# 检查数据库是否存在
DB_EXISTS=$(mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" -e "SHOW DATABASES LIKE '$DB_NAME';" | grep -c "$DB_NAME" || true)

if [ "$DB_EXISTS" -eq 0 ]; then
    echo -e "${YELLOW}⚠️  数据库 $DB_NAME 不存在${NC}"
    read -p "是否执行完整的 schema.sql 初始化数据库? (y/n): " INIT_DB
    
    if [ "$INIT_DB" = "y" ] || [ "$INIT_DB" = "Y" ]; then
        echo "执行 schema.sql..."
        mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" < "$MIGRATIONS_DIR/schema.sql"
        echo -e "${GREEN}✅ 数据库初始化完成${NC}"
        exit 0
    else
        echo -e "${RED}❌ 操作已取消${NC}"
        exit 1
    fi
fi

# 查找所有迁移文件
MIGRATE_FILES=$(ls -1 "$MIGRATIONS_DIR"/migrate_*.sql 2>/dev/null | sort)

if [ -z "$MIGRATE_FILES" ]; then
    echo -e "${YELLOW}⚠️  没有找到迁移文件${NC}"
    exit 0
fi

echo "找到以下迁移文件:"
echo "$MIGRATE_FILES" | while read -r file; do
    basename "$file"
done
echo ""

read -p "是否执行这些迁移? (y/n): " CONFIRM

if [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
    echo -e "${YELLOW}操作已取消${NC}"
    exit 0
fi

echo ""
echo "开始执行迁移..."
echo ""

# 执行每个迁移文件
MIGRATION_COUNT=0
while IFS= read -r file; do
    filename=$(basename "$file")
    echo "执行: $filename"
    
    if mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$file"; then
        echo -e "${GREEN}✅ $filename 执行成功${NC}"
        MIGRATION_COUNT=$((MIGRATION_COUNT + 1))
    else
        echo -e "${RED}❌ $filename 执行失败${NC}"
        exit 1
    fi
    echo ""
done <<< "$MIGRATE_FILES"

echo "========================================="
echo -e "${GREEN}✅ 迁移完成！共执行 $MIGRATION_COUNT 个迁移文件${NC}"
echo "========================================="
