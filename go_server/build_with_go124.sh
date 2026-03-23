#!/bin/bash
export GOROOT="/Users/bilson.li/.gvm/gos/go1.24.6"
export GOPATH="/Users/bilson.li/.gvm/pkgsets/go1.24.6/global"
export PATH="$GOROOT/bin:$PATH"

cd /Users/bilson.li/work/personal/code/electric_ai_tool/go_server

echo "Using Go version:"
go version

echo "Cleaning cache..."
go clean -modcache

echo "Tidying modules..."
go mod tidy

echo "Building..."
go build -o electric_ai_tool .

if [ $? -eq 0 ]; then
    echo "✅ 编译成功！"
    ls -lh electric_ai_tool
else
    echo "❌ 编译失败"
    exit 1
fi
