#!/bin/bash

# Context-ID 项目部署脚本
# 支持快速部署到生产环境

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "Context-ID 项目部署脚本"
    echo ""
    echo "使用方法:"
    echo "  $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  -i, --install-docker    安装Docker和Docker Compose"
    echo "  -s, --setup-env         交互式配置环境变量"
    echo "  -d, --deploy            部署项目"
    echo "  -u, --update            更新项目"
    echo "  --ip IP_ADDRESS         指定服务器IP地址"
    echo "  --domain DOMAIN         指定域名"
    echo ""
    echo "示例:"
    echo "  $0 --install-docker     # 安装Docker"
    echo "  $0 --ip 192.168.1.100   # 使用IP部署"
    echo "  $0 --domain example.com # 使用域名部署"
    echo ""
}

# 安装Docker
install_docker() {
    log_step "安装Docker和Docker Compose..."
    
    if command -v docker >/dev/null 2>&1; then
        log_info "Docker已安装，版本: $(docker --version)"
    else
        log_info "开始安装Docker..."
        chmod +x ./scripts/install-docker.sh
        ./scripts/install-docker.sh
    fi
}

# 获取服务器IP
get_server_ip() {
    if [ -n "$SERVER_IP" ]; then
        echo "$SERVER_IP"
        return
    fi
    
    # 尝试多种方法获取IP
    local ip=""
    
    # 方法1: ip route (Linux)
    if command -v ip >/dev/null 2>&1; then
        ip=$(ip route get 8.8.8.8 2>/dev/null | grep -oP 'src \K\S+' | head -1)
    fi
    
    # 方法2: ifconfig (通用)
    if [ -z "$ip" ] && command -v ifconfig >/dev/null 2>&1; then
        ip=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}' | head -1)
    fi
    
    # 方法3: hostname -I (Linux)
    if [ -z "$ip" ] && command -v hostname >/dev/null 2>&1; then
        ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    fi
    
    echo "$ip"
}

# 交互式配置环境
setup_environment() {
    log_step "配置部署环境..."
    
    # 获取当前IP
    local current_ip=$(get_server_ip)
    
    echo ""
    echo "请选择部署方式:"
    echo "1) 使用IP地址部署 (推荐用于开发/测试)"
    echo "2) 使用域名部署 (推荐用于生产环境)"
    echo "3) 使用localhost (仅限本地开发)"
    echo ""
    
    read -p "请选择 [1-3]: " deploy_type
    
    case $deploy_type in
        1)
            if [ -n "$current_ip" ]; then
                log_info "检测到服务器IP: $current_ip"
                read -p "使用此IP? [Y/n]: " use_detected_ip
                if [[ $use_detected_ip =~ ^[Nn]$ ]]; then
                    read -p "请输入服务器IP地址: " SERVER_IP
                else
                    SERVER_IP="$current_ip"
                fi
            else
                read -p "请输入服务器IP地址: " SERVER_IP
            fi
            
            export APP_EXTERNAL_URL="http://$SERVER_IP:8080"
            export APP_CASDOOR_EXTERNAL_URL="http://$SERVER_IP:8000"
            ;;
        2)
            read -p "请输入主域名 (如: example.com): " DOMAIN
            read -p "是否使用HTTPS? [Y/n]: " use_https
            
            if [[ $use_https =~ ^[Nn]$ ]]; then
                protocol="http"
                backend_port=":8080"
                casdoor_port=":8000"
            else
                protocol="https"
                backend_port=""
                casdoor_port=""
            fi
            
            export APP_EXTERNAL_URL="$protocol://$DOMAIN$backend_port"
            export APP_CASDOOR_EXTERNAL_URL="$protocol://auth.$DOMAIN$casdoor_port"
            ;;
        3)
            export APP_EXTERNAL_URL="http://localhost:8080"
            export APP_CASDOOR_EXTERNAL_URL="http://localhost:8000"
            ;;
        *)
            log_error "无效选择"
            exit 1
            ;;
    esac
    
    log_info "配置完成:"
    log_info "  应用地址: $APP_EXTERNAL_URL"
    log_info "  认证地址: $APP_CASDOOR_EXTERNAL_URL"
}

# 部署项目
deploy_project() {
    log_step "部署Context-ID项目..."
    
    # 检查Docker
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker未安装，请先运行: $0 --install-docker"
        exit 1
    fi
    
    # 停止现有服务
    log_info "停止现有服务..."
    docker-compose down 2>/dev/null || true
    
    # 构建并启动服务
    log_info "构建并启动服务..."
    docker-compose up -d --build
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 20
    
    # 检查服务状态
    log_info "检查服务状态..."
    docker-compose ps
    
    # 测试API
    log_info "测试API连接..."
    if curl -s "http://localhost:8080/api/v1/auth/login-url" >/dev/null; then
        log_info "✅ API服务正常"
    else
        log_warn "⚠️ API服务可能未完全启动，请稍后测试"
    fi
    
    log_info "🎉 部署完成！"
    echo ""
    echo "访问地址:"
    echo "  后端API: $APP_EXTERNAL_URL"
    echo "  认证服务: $APP_CASDOOR_EXTERNAL_URL"
    echo ""
    echo "测试命令:"
    echo "  curl $APP_EXTERNAL_URL/api/v1/auth/login-url"
}

# 更新项目
update_project() {
    log_step "更新Context-ID项目..."
    
    # 拉取最新代码
    if [ -d ".git" ]; then
        log_info "拉取最新代码..."
        git pull
    fi
    
    # 重新构建并部署
    deploy_project
}

# 主函数
main() {
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -i|--install-docker)
                install_docker
                exit 0
                ;;
            -s|--setup-env)
                setup_environment
                exit 0
                ;;
            -d|--deploy)
                DEPLOY_ONLY=true
                shift
                ;;
            -u|--update)
                update_project
                exit 0
                ;;
            --ip)
                SERVER_IP="$2"
                export APP_EXTERNAL_URL="http://$SERVER_IP:8080"
                export APP_CASDOOR_EXTERNAL_URL="http://$SERVER_IP:8000"
                shift 2
                ;;
            --domain)
                DOMAIN="$2"
                export APP_EXTERNAL_URL="https://$DOMAIN"
                export APP_CASDOOR_EXTERNAL_URL="https://auth.$DOMAIN"
                shift 2
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 如果没有指定环境变量，进行交互式配置
    if [ -z "$APP_EXTERNAL_URL" ] && [ "$DEPLOY_ONLY" != "true" ]; then
        setup_environment
    fi
    
    # 部署项目
    deploy_project
}

# 运行主函数
main "$@"
