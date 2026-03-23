#!/bin/bash

# 快速生成管理员密码的脚本（无需Python依赖）

PASSWORD="${1:-123456}"

echo "🔐 生成管理员密码哈希"
echo "================================"
echo "密码: $PASSWORD"
echo ""

# 检查是否安装了htpasswd
if command -v htpasswd &> /dev/null; then
    echo "✅ 使用htpasswd生成Bcrypt哈希:"
    htpasswd -nbB admin "$PASSWORD" | cut -d: -f2
    echo ""
elif command -v python3 &> /dev/null; then
    echo "✅ 使用Python生成Bcrypt哈希:"
    python3 -c "import bcrypt; print(bcrypt.hashpw('$PASSWORD'.encode(), bcrypt.gensalt()).decode())" 2>/dev/null || echo "请安装bcrypt: pip3 install bcrypt"
    echo ""
fi

# MD5哈希（前端用）
if command -v md5 &> /dev/null; then
    # macOS
    MD5_HASH=$(echo -n "$PASSWORD" | md5)
elif command -v md5sum &> /dev/null; then
    # Linux
    MD5_HASH=$(echo -n "$PASSWORD" | md5sum | cut -d' ' -f1)
fi

echo "📱 MD5哈希（前端测试用）:"
echo "$MD5_HASH"
echo ""

# Go代码片段
echo "📝 Go代码示例:"
cat << 'EOF'
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "123456"
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    fmt.Println(string(hash))
}
EOF

echo ""
echo "================================"
echo "💡 使用方法:"
echo "  ./generate_admin_password.sh [密码]"
echo "  默认密码: 123456"
