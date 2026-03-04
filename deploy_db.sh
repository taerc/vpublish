#!/bin/bash
# ============================================
# vpublish 数据库部署脚本 (Linux/Mac)
# ============================================

set -e

# 默认配置
DB_HOST="localhost"
DB_PORT="3306"
DB_USER="root"
DB_PASS=""
DB_NAME="vpublish"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 显示帮助
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --host      数据库主机 (默认: localhost)"
    echo "  -P, --port      数据库端口 (默认: 3306)"
    echo "  -u, --user      数据库用户 (默认: root)"
    echo "  -p, --password  数据库密码 (默认: 空)"
    echo "  --help          显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0"
    echo "  $0 -h 192.168.1.100 -u root -p mypassword"
    exit 0
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--host)
            DB_HOST="$2"
            shift 2
            ;;
        -P|--port)
            DB_PORT="$2"
            shift 2
            ;;
        -u|--user)
            DB_USER="$2"
            shift 2
            ;;
        -p|--password)
            DB_PASS="$2"
            shift 2
            ;;
        --help)
            show_help
            ;;
        *)
            echo -e "${RED}未知参数: $1${NC}"
            show_help
            ;;
    esac
done

echo "============================================"
echo "  vpublish 数据库部署脚本"
echo "============================================"
echo ""
echo "数据库配置:"
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo "  用户: $DB_USER"
echo "  数据库: $DB_NAME"
echo ""

# 检查 mysql 命令
if ! command -v mysql &> /dev/null; then
    echo -e "${RED}[错误] 未找到 mysql 命令，请确保 MySQL 客户端已安装${NC}"
    exit 1
fi

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_FILE="$SCRIPT_DIR/migrations/init.sql"

# 检查 SQL 文件
if [ ! -f "$SQL_FILE" ]; then
    echo -e "${RED}[错误] 未找到 SQL 文件: $SQL_FILE${NC}"
    exit 1
fi

echo -e "${YELLOW}正在部署数据库...${NC}"
echo ""

# 执行 SQL 脚本
if [ -z "$DB_PASS" ]; then
    mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" < "$SQL_FILE"
else
    mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" < "$SQL_FILE"
fi

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}============================================"
    echo "  数据库部署成功!"
    echo "============================================${NC}"
    echo ""
    echo "默认管理员账户:"
    echo "  用户名: admin"
    echo "  密码: admin123"
    echo ""
    echo "测试APP密钥:"
    echo "  AppKey: test_app_key_12345678"
    echo "  AppSecret: test_app_secret_abcdefgh"
    echo ""
else
    echo ""
    echo -e "${RED}[错误] 数据库部署失败，请检查配置和权限${NC}"
    exit 1
fi