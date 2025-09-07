#!/bin/bash

# 测试不同IP配置的脚本

echo "=== 测试不同IP配置 ==="

# 获取本机IP地址
LOCAL_IP=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}' | head -1)
echo "检测到本机IP: $LOCAL_IP"

echo ""
echo "1. 当前配置测试:"
curl -s "http://localhost:8080/api/v1/auth/login-url" | jq -r '.data.login_url'

echo ""
echo "2. 使用本机IP重新配置并测试:"
echo "正在重新配置环境变量..."

# 停止服务
docker-compose down > /dev/null 2>&1

# 设置环境变量并启动
APP_EXTERNAL_URL="http://$LOCAL_IP:8080" \
APP_CASDOOR_EXTERNAL_URL="http://$LOCAL_IP:8000" \
docker-compose up -d > /dev/null 2>&1

echo "等待服务启动..."
sleep 15

echo ""
echo "3. 新配置测试结果:"
curl -s "http://localhost:8080/api/v1/auth/login-url" | jq -r '.data.login_url'

echo ""
echo "4. 验证环境变量:"
docker exec contextid-backend env | grep -E "APP_EXTERNAL_URL|APP_CASDOOR_EXTERNAL_URL"

echo ""
echo "=== 测试完成 ==="
echo "如果URL中包含您的IP地址 ($LOCAL_IP)，说明配置成功！"

