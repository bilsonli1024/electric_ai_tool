#!/bin/bash

echo "🚀 Electric AI Tool - 快速启动"
echo "=============================="
echo ""

if [ ! -f "go_server/.env" ]; then
    echo "⚠️  未找到环境配置文件"
    echo "请先配置 go_server/.env 文件"
    echo "可以从 go_server/.env.example 复制并修改"
    exit 1
fi

if ! command -v mysql &> /dev/null; then
    echo "⚠️  未检测到MySQL，请确保MySQL已安装并运行"
fi

echo "1. 检查数据库连接..."
cd go_server

DB_HOST=$(grep DB_HOST .env | cut -d '=' -f2)
DB_PORT=$(grep DB_PORT .env | cut -d '=' -f2)
DB_USER=$(grep DB_USER .env | cut -d '=' -f2)
DB_PASSWORD=$(grep DB_PASSWORD .env | cut -d '=' -f2)
DB_NAME=$(grep DB_NAME .env | cut -d '=' -f2)

mysql -h "${DB_HOST:-localhost}" -P "${DB_PORT:-3306}" -u "${DB_USER}" -p"${DB_PASSWORD}" -e "USE ${DB_NAME};" 2>/dev/null

if [ $? -ne 0 ]; then
    echo "❌ 数据库连接失败"
    echo "请运行 ./go_server/init_db.sh 初始化数据库"
    exit 1
fi

echo "✅ 数据库连接成功"
echo ""

echo "2. 安装后端依赖..."
go mod tidy
if [ $? -ne 0 ]; then
    echo "❌ 后端依赖安装失败"
    exit 1
fi
echo "✅ 后端依赖安装成功"
echo ""

echo "3. 安装前端依赖..."
cd ../web
if [ ! -d "node_modules" ]; then
    npm install
    if [ $? -ne 0 ]; then
        echo "❌ 前端依赖安装失败"
        exit 1
    fi
fi
echo "✅ 前端依赖安装成功"
echo ""

echo "4. 构建前端..."
npm run build
if [ $? -ne 0 ]; then
    echo "❌ 前端构建失败"
    exit 1
fi
echo "✅ 前端构建成功"
echo ""

echo "5. 启动服务..."
cd ../go_server
echo ""
echo "=========================================="
echo "🎉 服务启动中..."
echo "访问地址: http://localhost:3002"
echo "=========================================="
echo ""
go run main.go
