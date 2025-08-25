#!/bin/bash

# Context-ID Backend 停止脚本

set -e

echo "🛑 停止 Context-ID Backend 服务..."

# 停止并删除容器
docker-compose down

echo "✅ 服务已停止"

# 询问是否删除数据卷
read -p "是否删除所有数据卷？这将删除所有数据 (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🗑️ 删除数据卷..."
    docker-compose down -v
    echo "✅ 数据卷已删除"
fi

echo ""
echo "📋 重新启动服务:"
echo "  ./scripts/start.sh"
echo ""
