#!/bin/bash

# 检查 Go 版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.24"

version_ge() {
    printf '%s\n%s' "$2" "$1" | sort -V -C
}

echo "检测到 Go 版本: $GO_VERSION"
echo "需要 Go 版本: >= $REQUIRED_VERSION"

if ! version_ge "$GO_VERSION" "$REQUIRED_VERSION"; then
    echo ""
    echo "❌ Go 版本过低！"
    echo ""
    echo "请升级 Go 到 1.24 或更高版本："
    echo ""
    echo "  macOS (Homebrew):"
    echo "    brew update && brew upgrade go"
    echo ""
    echo "  手动安装:"
    echo "    访问 https://go.dev/dl/ 下载最新版本"
    echo ""
    exit 1
fi

echo "✅ Go 版本符合要求"
echo ""

# 检查 .env 文件
if [ ! -f ".env" ]; then
    echo "⚠️  未找到 .env 文件"
    if [ -f ".env.example" ]; then
        echo "从 .env.example 创建 .env..."
        cp .env.example .env
        echo "✅ 已创建 .env 文件，请编辑并填入你的 GEMINI_API_KEY"
        exit 0
    else
        echo "❌ 未找到 .env.example 文件"
        exit 1
    fi
fi

# 检查依赖
if [ ! -f "go.sum" ]; then
    echo "安装依赖..."
    go mod download
    go mod tidy
    echo "✅ 依赖安装完成"
    echo ""
fi

# 启动服务器
echo "启动 Go 服务器..."
go run main.go
