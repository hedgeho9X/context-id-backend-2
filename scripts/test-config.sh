#!/bin/bash

# 测试配置脚本 - 检查环境变量是否正确传递到容器中

echo "=== 测试环境变量配置 ==="

echo "1. 检查容器内的环境变量:"
docker exec contextid-backend env | grep -E "APP_|CASDOOR_"

echo ""
echo "2. 测试登录URL API:"
curl -s "http://localhost:8080/api/v1/auth/login-url" | jq .

echo ""
echo "3. 如果URL还是localhost，请尝试以下方法："
echo "   方法1: 完全重新构建容器"
echo "   docker-compose down && docker-compose up -d --build"
echo ""
echo "   方法2: 直接设置环境变量测试"
echo "   APP_EXTERNAL_URL=http://YOUR_IP:8080 APP_CASDOOR_EXTERNAL_URL=http://YOUR_IP:8000 docker-compose up -d"

