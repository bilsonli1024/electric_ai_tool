#!/bin/bash

# 简化的编译脚本
export GOROOT="/Users/bilson.li/.gvm/gos/go1.24.6"
export GOPATH="/Users/bilson.li/.gvm/pkgsets/go1.24.6/global"
export PATH="$GOROOT/bin:$PATH"

cd /Users/bilson.li/work/personal/code/electric_ai_tool/go_server

echo "=== Go Version ==="
go version 2>&1

echo ""
echo "=== Building ==="
go build -v -o electric_ai_tool . 2>&1
BUILD_EXIT=$?

echo ""
echo "=== Build Exit Code: $BUILD_EXIT ==="

if [ $BUILD_EXIT -eq 0 ] && [ -f electric_ai_tool ]; then
    echo "✅ 编译成功！"
    ls -lh electric_ai_tool
else
    echo "❌ 编译失败，退出码: $BUILD_EXIT"
    exit 1
fi
