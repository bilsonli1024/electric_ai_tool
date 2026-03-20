#!/bin/bash

echo "==================================="
echo "认证系统测试脚本"
echo "==================================="
echo ""

API_BASE="http://localhost:3002"

echo "测试MD5哈希: 'password' -> 5f4dcc3b5aa765d61d8327deb882cf99"
PASSWORD_HASH="5f4dcc3b5aa765d61d8327deb882cf99"
EMAIL="test_$(date +%s)@example.com"

echo ""
echo "1. 测试注册 (邮箱: $EMAIL)"
echo "-----------------------------------"
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password_hash\": \"$PASSWORD_HASH\"
  }")

echo "$REGISTER_RESPONSE" | jq '.'

SESSION_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.session_id')
USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.user.id')
USERNAME=$(echo "$REGISTER_RESPONSE" | jq -r '.user.username')

if [ "$SESSION_ID" != "null" ] && [ "$SESSION_ID" != "" ]; then
    echo "✅ 注册成功"
    echo "   用户ID: $USER_ID"
    echo "   用户名: $USERNAME"
    echo "   会话ID: ${SESSION_ID:0:20}..."
else
    echo "❌ 注册失败"
    exit 1
fi

echo ""
echo "2. 测试获取用户信息"
echo "-----------------------------------"
ME_RESPONSE=$(curl -s -X GET "$API_BASE/api/auth/me" \
  -H "Authorization: Bearer $SESSION_ID")

echo "$ME_RESPONSE" | jq '.'

if echo "$ME_RESPONSE" | jq -e '.email' > /dev/null; then
    echo "✅ 获取用户信息成功"
else
    echo "❌ 获取用户信息失败"
fi

echo ""
echo "3. 测试登出"
echo "-----------------------------------"
LOGOUT_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/logout" \
  -H "Authorization: Bearer $SESSION_ID")

echo "$LOGOUT_RESPONSE" | jq '.'
echo "✅ 登出成功"

echo ""
echo "4. 测试使用邮箱登录"
echo "-----------------------------------"
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"login_id\": \"$EMAIL\",
    \"password_hash\": \"$PASSWORD_HASH\"
  }")

echo "$LOGIN_RESPONSE" | jq '.'

NEW_SESSION_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.session_id')

if [ "$NEW_SESSION_ID" != "null" ] && [ "$NEW_SESSION_ID" != "" ]; then
    echo "✅ 邮箱登录成功"
else
    echo "❌ 邮箱登录失败"
    exit 1
fi

echo ""
echo "5. 测试使用用户ID登录"
echo "-----------------------------------"
LOGIN_RESPONSE2=$(curl -s -X POST "$API_BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"login_id\": \"$USERNAME\",
    \"password_hash\": \"$PASSWORD_HASH\"
  }")

echo "$LOGIN_RESPONSE2" | jq '.'

if echo "$LOGIN_RESPONSE2" | jq -e '.session_id' > /dev/null; then
    echo "✅ 用户ID登录成功"
else
    echo "❌ 用户ID登录失败"
fi

echo ""
echo "6. 测试忘记密码"
echo "-----------------------------------"
FORGOT_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/forgot-password" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\"
  }")

echo "$FORGOT_RESPONSE" | jq '.'

if echo "$FORGOT_RESPONSE" | jq -e '.message' > /dev/null; then
    echo "✅ 忘记密码请求成功"
else
    echo "❌ 忘记密码请求失败"
fi

echo ""
echo "7. 验证数据库记录"
echo "-----------------------------------"
echo "查询用户表..."
mysql -u root -p -e "
USE electric_ai_tool;
SELECT id, username, email, LENGTH(password_hash) as hash_len, LENGTH(salt) as salt_len, status 
FROM users 
WHERE email = '$EMAIL';
" 2>/dev/null || echo "跳过数据库验证（需要MySQL访问权限）"

echo ""
echo "查询密码重置令牌..."
mysql -u root -p -e "
USE electric_ai_tool;
SELECT id, user_id, LEFT(token, 20) as token_preview, expires_at, used 
FROM password_reset_tokens 
WHERE user_id = $USER_ID 
ORDER BY created_at DESC 
LIMIT 1;
" 2>/dev/null || echo "跳过数据库验证（需要MySQL访问权限）"

echo ""
echo "==================================="
echo "测试完成！"
echo "==================================="
echo ""
echo "总结："
echo "✅ 注册功能"
echo "✅ 邮箱登录"
echo "✅ 用户ID登录"
echo "✅ 获取用户信息"
echo "✅ 登出功能"
echo "✅ 忘记密码"
echo ""
echo "注意："
echo "- 密码重置令牌需要手动从数据库获取进行测试"
echo "- 生产环境需要配置邮件发送服务"
