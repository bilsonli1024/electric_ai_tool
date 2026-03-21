#!/bin/bash

echo "Electric AI Tool - 数据库初始化"
echo "================================"
echo ""

read -p "请输入MySQL主机地址 [localhost]: " DB_HOST
DB_HOST=${DB_HOST:-localhost}

read -p "请输入MySQL端口 [3306]: " DB_PORT
DB_PORT=${DB_PORT:-3306}

read -p "请输入MySQL用户名 [root]: " DB_USER
DB_USER=${DB_USER:-root}

read -sp "请输入MySQL密码: " DB_PASSWORD
echo ""

read -p "请输入数据库名称 [electric_ai_tool]: " DB_NAME
DB_NAME=${DB_NAME:-electric_ai_tool}

echo ""
echo "正在创建数据库..."

mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

if [ $? -eq 0 ]; then
    echo "✅ 数据库创建成功"
    echo ""
    echo "正在导入表结构..."
    
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < migrations/schema.sql
    
    if [ $? -eq 0 ]; then
        echo "✅ 表结构导入成功"
        echo ""
        echo "数据库初始化完成！"
        echo ""
        echo "请更新 .env 文件中的数据库配置："
        echo "DB_HOST=$DB_HOST"
        echo "DB_PORT=$DB_PORT"
        echo "DB_USER=$DB_USER"
        echo "DB_PASSWORD=your_password"
        echo "DB_NAME=$DB_NAME"
    else
        echo "❌ 表结构导入失败"
        exit 1
    fi
else
    echo "❌ 数据库创建失败"
    exit 1
fi
