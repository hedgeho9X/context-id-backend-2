#!/bin/bash

# Docker 和 Docker Compose 快速安装脚本
# 支持 Ubuntu/Debian/CentOS/RHEL/Amazon Linux

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si)
        VER=$(lsb_release -sr)
    elif [ -f /etc/redhat-release ]; then
        OS="CentOS"
    elif [ -f /etc/debian_version ]; then
        OS="Debian"
    else
        log_error "无法检测操作系统"
        exit 1
    fi
    
    log_info "检测到操作系统: $OS $VER"
}

# 检查是否为root用户或有sudo权限
check_privileges() {
    if [[ $EUID -eq 0 ]]; then
        SUDO=""
    elif command -v sudo >/dev/null 2>&1; then
        SUDO="sudo"
        log_info "将使用sudo执行特权命令"
    else
        log_error "需要root权限或sudo权限来安装Docker"
        exit 1
    fi
}

# 安装Docker (Ubuntu/Debian)
install_docker_ubuntu() {
    log_info "在Ubuntu/Debian上安装Docker..."
    
    # 更新包索引
    $SUDO apt-get update
    
    # 安装必要的包
    $SUDO apt-get install -y \
        ca-certificates \
        curl \
        gnupg \
        lsb-release
    
    # 添加Docker官方GPG密钥
    $SUDO mkdir -m 0755 -p /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | $SUDO gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    
    # 设置稳定版仓库
    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
        $(lsb_release -cs) stable" | $SUDO tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    # 更新包索引
    $SUDO apt-get update
    
    # 安装Docker Engine
    $SUDO apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
}

# 安装Docker (CentOS/RHEL/Amazon Linux)
install_docker_centos() {
    log_info "在CentOS/RHEL上安装Docker..."
    
    # 安装必要的包
    $SUDO yum install -y yum-utils
    
    # 添加Docker仓库
    $SUDO yum-config-manager \
        --add-repo \
        https://download.docker.com/linux/centos/docker-ce.repo
    
    # 安装Docker Engine
    $SUDO yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
}

# 安装Docker Compose (独立版本，作为备用)
install_docker_compose_standalone() {
    log_info "安装Docker Compose独立版本..."
    
    # 获取最新版本
    DOCKER_COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep 'tag_name' | cut -d\" -f4)
    
    if [ -z "$DOCKER_COMPOSE_VERSION" ]; then
        DOCKER_COMPOSE_VERSION="v2.24.1"
        log_warn "无法获取最新版本，使用默认版本: $DOCKER_COMPOSE_VERSION"
    fi
    
    # 下载Docker Compose
    $SUDO curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    
    # 添加执行权限
    $SUDO chmod +x /usr/local/bin/docker-compose
    
    # 创建软链接
    $SUDO ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose
}

# 配置Docker
configure_docker() {
    log_info "配置Docker..."
    
    # 启动Docker服务
    $SUDO systemctl start docker
    $SUDO systemctl enable docker
    
    # 将当前用户添加到docker组（避免每次使用sudo）
    if [ "$SUDO" = "sudo" ]; then
        $SUDO usermod -aG docker $USER
        log_info "已将用户 $USER 添加到docker组"
        log_warn "请注销并重新登录以使组权限生效，或运行: newgrp docker"
    fi
}

# 验证安装
verify_installation() {
    log_info "验证Docker安装..."
    
    # 检查Docker版本
    if command -v docker >/dev/null 2>&1; then
        DOCKER_VERSION=$(docker --version)
        log_info "Docker版本: $DOCKER_VERSION"
    else
        log_error "Docker安装失败"
        exit 1
    fi
    
    # 检查Docker Compose版本
    if docker compose version >/dev/null 2>&1; then
        COMPOSE_VERSION=$(docker compose version)
        log_info "Docker Compose版本: $COMPOSE_VERSION"
    elif command -v docker-compose >/dev/null 2>&1; then
        COMPOSE_VERSION=$(docker-compose --version)
        log_info "Docker Compose版本: $COMPOSE_VERSION"
    else
        log_warn "Docker Compose未正确安装，尝试安装独立版本..."
        install_docker_compose_standalone
    fi
    
    # 测试Docker运行
    log_info "测试Docker运行..."
    if $SUDO docker run --rm hello-world >/dev/null 2>&1; then
        log_info "✅ Docker安装成功并正常运行！"
    else
        log_error "❌ Docker运行测试失败"
        exit 1
    fi
}

# 显示使用说明
show_usage() {
    log_info "Docker安装完成！"
    echo
    echo "使用说明:"
    echo "1. 如果您不是root用户，请运行以下命令使docker组权限生效:"
    echo "   newgrp docker"
    echo
    echo "2. 或者注销并重新登录"
    echo
    echo "3. 测试Docker:"
    echo "   docker run hello-world"
    echo
    echo "4. 启动您的项目:"
    echo "   docker-compose up -d"
    echo
}

# 主函数
main() {
    log_info "开始安装Docker和Docker Compose..."
    
    detect_os
    check_privileges
    
    case "$OS" in
        *Ubuntu*|*Debian*)
            install_docker_ubuntu
            ;;
        *CentOS*|*"Red Hat"*|*"Amazon Linux"*)
            install_docker_centos
            ;;
        *)
            log_error "不支持的操作系统: $OS"
            log_info "请手动安装Docker: https://docs.docker.com/engine/install/"
            exit 1
            ;;
    esac
    
    configure_docker
    verify_installation
    show_usage
}

# 运行主函数
main "$@"
