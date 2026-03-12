#!/bin/bash
#
# VPublish 一键安装脚本
# 用法: ./install.sh
#

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 root 权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 用户或 sudo 执行此脚本"
        exit 1
    fi
}

# 检查 systemd
check_systemd() {
    if ! command -v systemctl &> /dev/null; then
        log_error "当前系统不支持 systemd"
        exit 1
    fi
    log_info "systemd 检测通过"
}

# 创建目录结构
create_directories() {
    log_info "创建目录结构..."
    mkdir -p /opt/vpublish
    mkdir -p /opt/vpublish/uploads
    mkdir -p /opt/vpublish/configs
    mkdir -p /opt/vpublish/web/dist
    mkdir -p /opt/vpublish/logs
}

# 复制文件
copy_files() {
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local base_dir="$(dirname "$script_dir")"

    log_info "复制程序文件..."

    # 复制二进制
    if [ -f "$base_dir/vpublish-server" ]; then
        cp "$base_dir/vpublish-server" /opt/vpublish/
        chmod +x /opt/vpublish/vpublish-server
        log_info "后端服务已复制"
    else
        log_error "找不到 vpublish-server 二进制文件"
        exit 1
    fi

    # 复制配置文件模板
    if [ -f "$base_dir/configs/config.yaml.example" ]; then
        cp "$base_dir/configs/config.yaml.example" /opt/vpublish/configs/
        if [ ! -f /opt/vpublish/configs/config.yaml ]; then
            cp "$base_dir/configs/config.yaml.example" /opt/vpublish/configs/config.yaml
            log_info "配置文件已创建 (请修改 /opt/vpublish/configs/config.yaml)"
        fi
    fi

    # 复制前端文件
    if [ -d "$base_dir/web/dist" ]; then
        cp -r "$base_dir/web/dist/"* /opt/vpublish/web/dist/
        log_info "前端文件已复制"
    else
        log_warn "未找到前端构建文件，请手动构建前端"
    fi

    # 复制 systemd 服务文件
    if [ -f "$script_dir/vpublish.service" ]; then
        cp "$script_dir/vpublish.service" /etc/systemd/system/
        log_info "systemd 服务文件已安装"
    fi

    # 复制 Nginx 配置示例
    if [ -f "$script_dir/nginx.conf.example" ]; then
        mkdir -p /opt/vpublish/deploy
        cp "$script_dir/nginx.conf.example" /opt/vpublish/deploy/
        log_info "Nginx 配置示例已复制到 /opt/vpublish/deploy/"
    fi
}

# 安装 systemd 服务
install_service() {
    log_info "安装 systemd 服务..."
    systemctl daemon-reload
    systemctl enable vpublish
    log_info "服务已启用 (vpublish)"
}

# 打印后续步骤
print_next_steps() {
    echo ""
    echo "========================================"
    echo "安装完成!"
    echo "========================================"
    echo ""
    echo "后续步骤:"
    echo ""
    echo "1. 修改配置文件:"
    echo "   vi /opt/vpublish/configs/config.yaml"
    echo ""
    echo "2. 创建数据库 (如未创建):"
    echo "   mysql -u root -p -e \"CREATE DATABASE vpublish CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;\""
    echo ""
    echo "3. 启动服务:"
    echo "   systemctl start vpublish"
    echo ""
    echo "4. 检查服务状态:"
    echo "   systemctl status vpublish"
    echo ""
    echo "5. 查看日志:"
    echo "   journalctl -u vpublish -f"
    echo ""
    echo "6. 配置 Nginx 反向代理 (可选):"
    echo "   cp /opt/vpublish/deploy/nginx.conf.example /etc/nginx/sites-available/vpublish"
    echo "   ln -s /etc/nginx/sites-available/vpublish /etc/nginx/sites-enabled/"
    echo "   nginx -t && systemctl reload nginx"
    echo ""
}

main() {
    echo "========================================"
    echo "VPublish 安装脚本"
    echo "========================================"
    echo ""

    check_root
    check_systemd
    create_directories
    copy_files
    install_service
    print_next_steps
}

main "$@"