#!/bin/bash

# Context-ID Backend 启动脚本

set -e

echo "🚀 启动 Context-ID Backend 服务..."

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装，请先安装Docker"
    exit 1
fi

# 检查Docker Compose是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose未安装，请先安装Docker Compose"
    exit 1
fi

# 创建必要的目录
mkdir -p logs

echo "📦 拉取最新镜像..."
docker-compose pull

echo "🔨 构建服务..."
docker-compose build

echo "🚀 启动服务..."
docker-compose up -d

echo "⏳ 等待服务启动..."
sleep 10

echo "🔍 检查服务状态..."
docker-compose ps

echo ""
echo "✅ 服务启动完成！"
echo ""
echo "📋 服务信息:"
echo "  - GoFrame API: http://localhost:8080"
echo "  - Casdoor管理后台: http://localhost:8000"
echo "  - 默认管理员账号: built-in/admin"
echo "  - 默认管理员密码: 123"
echo ""
echo "📝 查看日志:"
echo "  docker-compose logs -f"
echo ""
echo "🛑 停止服务:"
echo "  docker-compose down"
echo ""
