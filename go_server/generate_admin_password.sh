#!/bin/bash

# 密码哈希生成脚本
# 用于生成管理员密码的bcrypt哈希

PASSWORD="${1:-123456}"

echo "生成密码哈希..."
echo "密码: $PASSWORD"
echo ""

# 使用Python生成bcrypt哈希（大多数系统都有Python）
python3 << EOF
import hashlib
import base64

# 使用固定的盐（用于admin账户）
SALT = "electric_ai_tool_2026"
password = "$PASSWORD"

# 方案1: 使用bcrypt（推荐）
try:
    import bcrypt
    # 固定的salt（从环境变量的盐生成bcrypt salt）
    salt = bcrypt.gensalt(rounds=10)
    hashed = bcrypt.hashpw(password.encode('utf-8'), salt)
    print("✅ Bcrypt哈希（用于Go后端）:")
    print(hashed.decode('utf-8'))
    print()
except ImportError:
    print("⚠️  bcrypt模块未安装，请运行: pip3 install bcrypt")
    print()

# 方案2: MD5（前端使用）
md5_hash = hashlib.md5(password.encode('utf-8')).hexdigest()
print("✅ MD5哈希（前端使用）:")
print(md5_hash)
print()

# 方案3: 带盐的SHA256（备用）
salted = (SALT + password).encode('utf-8')
sha256_hash = hashlib.sha256(salted).hexdigest()
print("📋 SHA256+盐（备用）:")
print(sha256_hash)
print()

EOF

echo ""
echo "💡 提示:"
echo "1. 将Bcrypt哈希复制到 schema.sql 的 INSERT 语句中"
echo "2. 固定盐 'electric_ai_tool_2026' 已设置"
echo "3. 前端使用MD5，后端使用bcrypt验证"
