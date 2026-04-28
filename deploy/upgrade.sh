#!/bin/bash
# ============================================
# vpublish 生产环境升级脚本
# ============================================
# 用法: sudo ./upgrade.sh [选项]
#
# 选项:
#   -d, --deploy-dir    安装包解压目录 (默认: 当前目录)
#   -b, --backup-dir    备份目录 (默认: /opt/vpublish/backups)
#   --dry-run           仅打印操作步骤，不实际执行
#   --rollback          执行回滚
#   --help              显示帮助信息
#
# 前置条件:
#   1. 已将 make package 生成的 tar.gz 上传到目标服务器并解压
#   2. 当前脚本位于解压后的目录中
#   3. 以 root 或 sudo 执行
# ============================================

set -euo pipefail

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step()  { echo -e "${BLUE}[STEP]${NC} $1"; }

# 默认配置
DEPLOY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="/opt/vpublish/backups"
DRY_RUN=false
ROLLBACK=false
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# 安装路径
INSTALL_DIR="/opt/vpublish"
SERVICE_NAME="vpublish"

# 数据库配置（从配置文件读取）
DB_HOST=""
DB_PORT=""
DB_USER=""
DB_PASS=""
DB_NAME=""

# ============================================
# 帮助信息
# ============================================
show_help() {
    echo "VPublish 生产环境升级脚本"
    echo ""
    echo "用法: sudo $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -d, --deploy-dir DIR    安装包解压目录 (默认: 当前目录)"
    echo "  -b, --backup-dir DIR    备份目录 (默认: /opt/vpublish/backups)"
    echo "  --dry-run               仅打印操作步骤"
    echo "  --rollback              执行回滚到上次备份"
    echo "  --help                  显示帮助"
    echo ""
    echo "升级流程:"
    echo "  1. 备份当前二进制和前端文件"
    echo "  2. 备份数据库（可选，需已安装 mysqldump）"
    echo "  3. 停止服务"
    echo "  4. 替换二进制和前端文件"
    echo "  5. 启动服务（自动执行数据库迁移）"
    echo "  6. 验证服务健康状态"
    echo ""
    echo "示例:"
    echo "  # 正常升级"
    echo "  sudo ./upgrade.sh"
    echo ""
    echo "  # 指定安装包目录和备份目录"
    echo "  sudo ./upgrade.sh -d /tmp/vpublish-v2.1.0-linux-amd64 -b /mnt/backup/vpublish"
    echo ""
    echo "  # 仅预览操作步骤"
    echo "  sudo ./upgrade.sh --dry-run"
    echo ""
    echo "  # 回滚到上次备份"
    echo "  sudo ./upgrade.sh --rollback"
    exit 0
}

# ============================================
# 解析参数
# ============================================
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--deploy-dir)
            DEPLOY_DIR="$2"
            shift 2
            ;;
        -b|--backup-dir)
            BACKUP_DIR="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --rollback)
            ROLLBACK=true
            shift
            ;;
        --help)
            show_help
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            ;;
    esac
done

# ============================================
# 打印配置
# ============================================
echo "=========================================="
echo " VPublish 升级脚本"
echo "=========================================="
echo " 部署包目录: $DEPLOY_DIR"
echo " 备份目录:   $BACKUP_DIR"
echo " 安装目录:   $INSTALL_DIR"
echo " 服务模式:   $([ "$DRY_RUN" = true ] && echo 'DRY RUN (仅预览)' || echo '实际执行')"
echo "=========================================="
echo ""

# ============================================
# 前置检查
# ============================================
pre_checks() {
    log_step "前置检查..."

    # 检查 root 权限
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 sudo 或 root 执行此脚本"
        exit 1
    fi

    # 检查 systemd
    if ! command -v systemctl &> /dev/null; then
        log_error "当前系统不支持 systemd"
        exit 1
    fi

    # 检查部署目录
    if [ ! -d "$DEPLOY_DIR" ]; then
        log_error "部署包目录不存在: $DEPLOY_DIR"
        exit 1
    fi

    # 检查二进制文件
    if [ ! -f "$DEPLOY_DIR/vpublish-server" ]; then
        log_error "找不到 vpublish-server 二进制文件"
        exit 1
    fi

    # 检查前端文件（可选）
    if [ ! -d "$DEPLOY_DIR/web/dist" ]; then
        log_warn "未找到前端构建文件 (web/dist)，将跳过前端更新"
    fi

    # 检查服务状态
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log_info "服务当前正在运行"
    else
        log_warn "服务当前未运行"
    fi

    log_info "前置检查通过"
}

# ============================================
# 读取数据库配置
# ============================================
read_db_config() {
    local config_file="$INSTALL_DIR/configs/config.yaml"
    if [ ! -f "$config_file" ]; then
        log_warn "未找到配置文件 $config_file，将跳过数据库备份"
        return 1
    fi

    DB_HOST=$(grep -E '^\s+host:' "$config_file" | head -1 | awk '{print $2}')
    DB_PORT=$(grep -E '^\s+port:' "$config_file" | head -1 | awk '{print $2}')
    DB_USER=$(grep -E '^\s+user:' "$config_file" | head -1 | awk '{print $2}')
    DB_PASS=$(grep -E '^\s+password:' "$config_file" | head -1 | awk '{print $2}' | tr -d '"')
    DB_NAME=$(grep -E '^\s+dbname:' "$config_file" | head -1 | awk '{print $2}')

    if [ -z "$DB_HOST" ] || [ -z "$DB_NAME" ]; then
        log_warn "无法从配置文件读取数据库信息，将跳过数据库备份"
        return 1
    fi

    log_info "数据库: $DB_NAME @ $DB_HOST:$DB_PORT"
    return 0
}

# ============================================
# 备份
# ============================================
do_backup() {
    log_step "执行备份..."

    local backup_path="$BACKUP_DIR/backup_${TIMESTAMP}"
    mkdir -p "$backup_path"

    # 1. 备份二进制文件
    if [ -f "$INSTALL_DIR/vpublish-server" ]; then
        cp "$INSTALL_DIR/vpublish-server" "$backup_path/"
        log_info "二进制文件已备份: $backup_path/vpublish-server"
    fi

    if [ -f "$INSTALL_DIR/vpublish-mcp" ]; then
        cp "$INSTALL_DIR/vpublish-mcp" "$backup_path/"
        log_info "MCP 服务已备份: $backup_path/vpublish-mcp"
    fi

    # 2. 备份前端文件
    if [ -d "$INSTALL_DIR/web/dist" ]; then
        cp -r "$INSTALL_DIR/web/dist" "$backup_path/dist"
        log_info "前端文件已备份: $backup_path/dist"
    fi

    # 3. 备份配置文件
    if [ -d "$INSTALL_DIR/configs" ]; then
        cp -r "$INSTALL_DIR/configs" "$backup_path/configs"
        log_info "配置文件已备份: $backup_path/configs"
    fi

    # 4. 备份数据库（如果 mysqldump 可用且配置正确）
    if command -v mysqldump &> /dev/null && read_db_config 2>/dev/null; then
        log_step "备份数据库..."
        mkdir -p "$backup_path/database"

        local db_backup="$backup_path/database/${DB_NAME}_${TIMESTAMP}.sql"
        if [ -n "$DB_PASS" ]; then
            mysqldump -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" > "$db_backup" 2>/dev/null
        else
            mysqldump -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" "$DB_NAME" > "$db_backup" 2>/dev/null
        fi

        if [ $? -eq 0 ]; then
            log_info "数据库备份成功: $db_backup"
        else
            log_warn "数据库备份失败（非致命，继续升级）"
        fi
    fi

    # 5. 记录备份信息
    echo "backup_timestamp=$TIMESTAMP" > "$backup_path/.backup_info"
    echo "deploy_dir=$DEPLOY_DIR" >> "$backup_path/.backup_info"
    echo "commit=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')" >> "$backup_path/.backup_info"

    log_info "备份完成: $backup_path"
}

# ============================================
# 升级
# ============================================
do_upgrade() {
    log_step "开始升级..."

    # 1. 停止服务
    log_step "停止服务..."
    systemctl stop "$SERVICE_NAME" || true
    sleep 2

    # 2. 替换二进制
    log_step "替换二进制文件..."
    cp "$DEPLOY_DIR/vpublish-server" "$INSTALL_DIR/vpublish-server"
    chmod +x "$INSTALL_DIR/vpublish-server"
    log_info "vpublish-server 已替换"

    if [ -f "$DEPLOY_DIR/vpublish-mcp" ]; then
        cp "$DEPLOY_DIR/vpublish-mcp" "$INSTALL_DIR/vpublish-mcp"
        chmod +x "$INSTALL_DIR/vpublish-mcp"
        log_info "vpublish-mcp 已替换"
    fi

    # 3. 替换前端文件
    if [ -d "$DEPLOY_DIR/web/dist" ]; then
        log_step "替换前端文件..."
        # 先备份旧前端
        if [ -d "$INSTALL_DIR/web/dist" ]; then
            rm -rf "$INSTALL_DIR/web/dist.old"
            cp -r "$INSTALL_DIR/web/dist" "$INSTALL_DIR/web/dist.old"
        fi
        rm -rf "$INSTALL_DIR/web/dist"
        cp -r "$DEPLOY_DIR/web/dist" "$INSTALL_DIR/web/dist"
        log_info "前端文件已更新"
    fi

    # 4. 更新部署文件（systemd service 等，仅在新增/变更时）
    if [ -d "$DEPLOY_DIR/deploy" ]; then
        log_step "更新部署文件..."
        if [ -f "$DEPLOY_DIR/deploy/vpublish.service" ]; then
            # 仅在 service 文件有变更时替换
            if [ ! -f "/etc/systemd/system/vpublish.service" ] || \
               ! diff -q "$DEPLOY_DIR/deploy/vpublish.service" /etc/systemd/system/vpublish.service &>/dev/null; then
                cp "$DEPLOY_DIR/deploy/vpublish.service" /etc/systemd/system/
                systemctl daemon-reload
                log_info "systemd 服务文件已更新"
            fi
        fi
    fi

    # 5. 启动服务（GORM AutoMigrate 会自动执行数据库迁移）
    log_step "启动服务..."
    systemctl start "$SERVICE_NAME"
    sleep 3

    # 6. 验证服务健康
    log_step "验证服务健康..."
    local health_response
    health_response=$(curl -s --max-time 5 http://127.0.0.1:8080/health 2>/dev/null || echo "")

    if echo "$health_response" | grep -q '"status":"ok"'; then
        log_info "服务健康检查通过"
        echo "$health_response" | python3 -m json.tool 2>/dev/null || echo "$health_response"
    else
        log_warn "健康检查未返回预期结果（服务可能仍在启动中）"
        log_warn "请执行以下命令检查服务状态:"
        echo "  systemctl status vpublish"
        echo "  journalctl -u vpublish -n 50 --no-pager"
    fi

    # 7. 验证数据库迁移（feature_type 列）
    log_step "验证数据库迁移..."
    if command -v mysql &> /dev/null && read_db_config 2>/dev/null; then
        local col_check
        if [ -n "$DB_PASS" ]; then
            col_check=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" \
                -N -e "SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='$DB_NAME' AND TABLE_NAME='versions' AND COLUMN_NAME='feature_type'" 2>/dev/null || echo "0")
        else
            col_check=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" "$DB_NAME" \
                -N -e "SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='$DB_NAME' AND TABLE_NAME='versions' AND COLUMN_NAME='feature_type'" 2>/dev/null || echo "0")
        fi

        if [ "$col_check" = "1" ]; then
            log_info "feature_type 列已存在"

            # 检查空值记录
            local null_count
            if [ -n "$DB_PASS" ]; then
                null_count=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" \
                    -N -e "SELECT COUNT(*) FROM versions WHERE feature_type IS NULL OR feature_type = ''" 2>/dev/null || echo "0")
            else
                null_count=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" "$DB_NAME" \
                    -N -e "SELECT COUNT(*) FROM versions WHERE feature_type IS NULL OR feature_type = ''" 2>/dev/null || echo "0")
            fi

            if [ "$null_count" = "0" ]; then
                log_info "所有版本记录 feature_type 均为非空 (正常)"
            else
                log_warn "发现 $null_count 条 feature_type 为空的记录"
                log_warn "建议手动执行: UPDATE versions SET feature_type = 'release' WHERE feature_type IS NULL OR feature_type = ''"
            fi
        else
            log_warn "feature_type 列不存在（可能数据库迁移未执行成功）"
            log_warn "请检查日志: journalctl -u vpublish -n 50 --no-pager"
        fi
    fi
}

# ============================================
# 回滚
# ============================================
do_rollback() {
    log_step "执行回滚..."

    # 查找最新备份
    local latest_backup
    latest_backup=$(ls -dt "$BACKUP_DIR"/backup_* 2>/dev/null | head -1)

    if [ -z "$latest_backup" ]; then
        log_error "未找到备份目录: $BACKUP_DIR/backup_*"
        exit 1
    fi

    log_info "使用备份: $latest_backup"

    # 确认回滚
    if [ "$DRY_RUN" = false ]; then
        read -p "确认回滚到 $(basename "$latest_backup")？[y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_error "已取消回滚"
            exit 0
        fi
    fi

    # 1. 停止服务
    log_step "停止服务..."
    systemctl stop "$SERVICE_NAME" || true
    sleep 2

    # 2. 恢复二进制
    if [ -f "$latest_backup/vpublish-server" ]; then
        cp "$latest_backup/vpublish-server" "$INSTALL_DIR/vpublish-server"
        chmod +x "$INSTALL_DIR/vpublish-server"
        log_info "二进制已恢复"
    fi

    if [ -f "$latest_backup/vpublish-mcp" ]; then
        cp "$latest_backup/vpublish-mcp" "$INSTALL_DIR/vpublish-mcp"
        chmod +x "$INSTALL_DIR/vpublish-mcp"
        log_info "MCP 服务已恢复"
    fi

    # 3. 恢复前端
    if [ -d "$latest_backup/dist" ]; then
        rm -rf "$INSTALL_DIR/web/dist"
        cp -r "$latest_backup/dist" "$INSTALL_DIR/web/dist"
        log_info "前端文件已恢复"
    fi

    # 4. 恢复配置
    if [ -d "$latest_backup/configs" ]; then
        rm -rf "$INSTALL_DIR/configs"
        cp -r "$latest_backup/configs" "$INSTALL_DIR/configs"
        log_info "配置文件已恢复"
    fi

    # 5. 恢复数据库（如果存在数据库备份）
    local db_backup=$(find "$latest_backup/database" -name "*.sql" 2>/dev/null | head -1)
    if [ -n "$db_backup" ] && command -v mysql &> /dev/null && read_db_config 2>/dev/null; then
        log_step "恢复数据库..."
        if [ -n "$DB_PASS" ]; then
            mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$db_backup" 2>/dev/null
        else
            mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" "$DB_NAME" < "$db_backup" 2>/dev/null
        fi
        log_info "数据库已恢复: $db_backup"
    fi

    # 6. 启动服务
    log_step "启动服务..."
    systemctl start "$SERVICE_NAME"
    sleep 3

    # 7. 验证
    local health_response
    health_response=$(curl -s --max-time 5 http://127.0.0.1:8080/health 2>/dev/null || echo "")
    if echo "$health_response" | grep -q '"status":"ok"'; then
        log_info "回滚成功，服务运行正常"
    else
        log_warn "回滚完成，但健康检查未通过"
        log_warn "请检查: systemctl status vpublish"
    fi
}

# ============================================
# 主流程
# ============================================
main() {
    if [ "$ROLLBACK" = true ]; then
        do_rollback
        exit 0
    fi

    if [ "$DRY_RUN" = true ]; then
        log_step "=== DRY RUN 模式 ==="
        pre_checks
        echo ""
        log_info "将执行以下操作:"
        echo "  1. 备份当前版本到: $BACKUP_DIR/backup_${TIMESTAMP}"
        echo "     - 二进制文件 (vpublish-server, vpublish-mcp)"
        echo "     - 前端文件 (web/dist)"
        echo "     - 配置文件 (configs/)"
        echo "     - 数据库 (通过 mysqldump)"
        echo "  2. 停止 vpublish 服务"
        echo "  3. 替换二进制和前端文件"
        echo "  4. 启动服务（自动执行数据库迁移）"
        echo "  5. 验证服务健康状态"
        echo ""
        log_info "实际执行请移除 --dry-run 参数"
        exit 0
    fi

    pre_checks
    do_backup
    do_upgrade

    echo ""
    echo "=========================================="
    echo -e "${GREEN}升级完成!${NC}"
    echo "=========================================="
    echo ""
    echo "备份位置: $BACKUP_DIR/backup_${TIMESTAMP}"
    echo ""
    echo "验证命令:"
    echo "  systemctl status vpublish"
    echo "  journalctl -u vpublish -f"
    echo "  curl http://127.0.0.1:8080/health"
    echo ""
    echo "如需回滚:"
    echo "  sudo ./upgrade.sh --rollback"
    echo "  或 sudo bash "$0" --rollback"
    echo "=========================================="
}

main
