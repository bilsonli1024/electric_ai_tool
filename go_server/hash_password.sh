#!/bin/bash
# 生成用户密码哈希的脚本
# 使用方式: ./hash_password.sh <password> <salt>

set -e

if [ "$#" -lt 2 ]; then
    echo "使用方式:"
    echo "  ./hash_password.sh <password> <salt>"
    echo "  ./hash_password.sh 123456 admin2026electric"
    exit 1
fi

PASSWORD="$1"
SALT="$2"

# 计算密码的MD5
PASSWORD_MD5=$(echo -n "$PASSWORD" | md5)
echo "密码: $PASSWORD"
echo "MD5: $PASSWORD_MD5"
echo "Salt: $SALT"

# 组合MD5和salt，计算SHA256
COMBINED="${PASSWORD_MD5}${SALT}"
FINAL_HASH=$(echo -n "$COMBINED" | shasum -a 256 | awk '{print $1}')

echo ""
echo "最终SHA256哈希: $FINAL_HASH"
echo ""
echo "SQL语句:"
echo "UPDATE users_tab SET password='$FINAL_HASH', salt='$SALT' WHERE email='admin@gmail.com';"
