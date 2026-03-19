#!/bin/bash

# API 测试脚本
# 用于验证 Go 服务器的 API 端点

BASE_URL="http://localhost:3002"

echo "🧪 测试 Go 服务器 API"
echo "===================="
echo ""

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# 测试健康检查
echo "1️⃣ 测试健康检查 (GET /api/health)"
echo "-----------------------------------"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/health")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✅ 成功${NC} (HTTP $HTTP_CODE)"
    echo "响应: $BODY"
else
    echo -e "${RED}❌ 失败${NC} (HTTP $HTTP_CODE)"
    echo "响应: $BODY"
fi
echo ""

# 测试分析端点（需要 API Key）
echo "2️⃣ 测试分析端点 (POST /api/analyze)"
echo "-----------------------------------"
echo "⚠️  此测试需要有效的 GEMINI_API_KEY"

ANALYZE_DATA='{
  "keywords": "wireless headphones",
  "sellingPoints": "noise cancellation, long battery life",
  "sku": "TEST-001"
}'

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/analyze" \
  -H "Content-Type: application/json" \
  -d "$ANALYZE_DATA")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✅ 成功${NC} (HTTP $HTTP_CODE)"
    echo "响应: $(echo $BODY | head -c 200)..." # 只显示前200个字符
else
    echo -e "${RED}❌ 失败${NC} (HTTP $HTTP_CODE)"
    echo "响应: $BODY"
fi
echo ""

# 测试 CORS
echo "3️⃣ 测试 CORS (OPTIONS /api/health)"
echo "-----------------------------------"
RESPONSE=$(curl -s -w "\n%{http_code}" -X OPTIONS "$BASE_URL/api/health" \
  -H "Origin: http://example.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✅ 成功${NC} (HTTP $HTTP_CODE)"
    echo "CORS 配置正确"
else
    echo -e "${RED}❌ 失败${NC} (HTTP $HTTP_CODE)"
fi
echo ""

echo "===================="
echo "✨ 测试完成"
echo ""
echo "注意："
echo "- 完整功能测试需要配置有效的 GEMINI_API_KEY"
echo "- 图片生成/编辑测试需要提供 base64 图片数据"
