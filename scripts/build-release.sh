#!/bin/bash
#
# VPublish 发版打包脚本
# 用法: ./scripts/build-release.sh [VERSION]
#
# 示例:
#   ./scripts/build-release.sh v2.1.0
#   ./scripts/build-release.sh          # 自动从 git tag 获取版本
#

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# 版本号
VERSION="${1:-}"
if [ -z "$VERSION" ]; then
    VERSION=$(git describe --tags --always 2>/dev/null || echo "v2.0.0")
fi

# 进入项目目录
cd "$PROJECT_DIR"

echo ""
echo "=========================================="
echo " VPublish Release Build"
echo "=========================================="
echo " Version: $VERSION"
echo " Project: $PROJECT_DIR"
echo "=========================================="
echo ""

# 检查 git 状态
check_git_status() {
    log_step "检查 Git 状态..."
    if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
        log_warn "存在未提交的更改:"
        git status --short
        read -p "是否继续? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_error "已取消"
            exit 1
        fi
    fi
}

# 检查依赖
check_dependencies() {
    log_step "检查依赖..."

    if ! command -v go &> /dev/null; then
        log_error "未安装 Go"
        exit 1
    fi

    if ! command -v npm &> /dev/null; then
        log_error "未安装 Node.js/npm"
        exit 1
    fi

    log_info "依赖检查通过"
}

# 执行打包
do_build() {
    log_step "开始打包..."

    # 使用 Makefile 的 package target
    make clean
    VERSION="$VERSION" make package

    if [ $? -eq 0 ]; then
        log_info "打包成功!"
    else
        log_error "打包失败"
        exit 1
    fi
}

# 显示结果
show_result() {
    local PACKAGE_NAME="vpublish-${VERSION}-linux-amd64"
    local PACKAGE_PATH="${PROJECT_DIR}/dist/${PACKAGE_NAME}.tar.gz"

    if [ -f "$PACKAGE_PATH" ]; then
        local SIZE=$(du -h "$PACKAGE_PATH" | cut -f1)
        local MD5=$(md5sum "$PACKAGE_PATH" | cut -d' ' -f1)

        echo ""
        echo "=========================================="
        echo " 打包完成"
        echo "=========================================="
        echo " 文件: dist/${PACKAGE_NAME}.tar.gz"
        echo " 大小: ${SIZE}"
        echo " MD5:  ${MD5}"
        echo ""
        echo " 部署步骤:"
        echo "   1. 上传到目标服务器:"
        echo "      scp dist/${PACKAGE_NAME}.tar.gz user@server:/tmp/"
        echo ""
        echo "   2. 解压并安装:"
        echo "      tar -xzf ${PACKAGE_NAME}.tar.gz"
        echo "      cd ${PACKAGE_NAME}"
        echo "      sudo ./deploy/install.sh"
        echo ""
        echo "   3. 修改配置并启动:"
        echo "      vi /opt/vpublish/configs/config.yaml"
        echo "      systemctl start vpublish"
        echo "=========================================="
    fi
}

main() {
    check_git_status
    check_dependencies
    do_build
    show_result
}

main "$@"