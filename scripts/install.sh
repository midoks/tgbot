#!/bin/bash
# ============================================
# tgbot 一键安装脚本
# 支持系统: CentOS 7+, Ubuntu 16.04+, Debian 9+
# 支持架构: x86_64, i386, arm64
# ============================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

VERSION="v0.1"
REPO="midoks/tgbot"
INSTALL_DIR="/opt/tgbot"
DATA_DIR="${INSTALL_DIR}/custom/data"
CONF_FILE="${INSTALL_DIR}/custom/conf"
SERVICE_NAME="tgbot"

detect_arch() {
    case "$(uname -m)" in
        x86_64) ARCH="amd64" ;;
        i386|i686) ARCH="386" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)
            echo -e "${RED}不支持的架构: $(uname -m)${NC}"
            exit 1
            ;;
    esac
    echo -e "${GREEN}检测到架构: ${ARCH}${NC}"
}

detect_os() {
    if [ -f /etc/centos-release ] || [ -f /etc/redhat-release ]; then
        OS="centos"
    elif [ -f /etc/debian_version ] || [ -f /etc/lsb-release ]; then
        OS="debian"
    else
        echo -e "${YELLOW}无法检测操作系统，将尝试通用安装方式${NC}"
        OS="unknown"
    fi
    echo -e "${GREEN}检测到操作系统: ${OS}${NC}"
}

install_deps() {
    echo -e "${YELLOW}正在安装依赖...${NC}"
    if [ "$OS" = "centos" ]; then
        yum install -y wget tar
    elif [ "$OS" = "debian" ]; then
        apt-get install -y wget tar
    else
        echo -e "${YELLOW}跳过依赖安装，请确保已安装 wget 和 tar${NC}"
    fi
}

create_dirs() {
    echo -e "${YELLOW}创建目录结构...${NC}"
    mkdir -p "$INSTALL_DIR"
}

get_file_md5() {
    local file="$1"
    if [ -f "$file" ]; then
        if command -v md5sum &> /dev/null; then
            md5sum "$file" | awk '{print $1}'
        elif command -v md5 &> /dev/null; then
            md5 -r "$file" | awk '{print $1}'
        else
            echo ""
        fi
    else
        echo ""
    fi
}

download_and_extract() {
    echo -e "${YELLOW}正在下载 tgbot ${VERSION}...${NC}"
    URL="https://github.com/${REPO}/releases/download/${VERSION}/tgbot_${VERSION}_linux_${ARCH}.tar.gz"
    TMP_FILE=$(mktemp)

    if ! wget -q -O "$TMP_FILE" "$URL"; then
        echo -e "${RED}下载失败，请检查网络连接${NC}"
        rm -f "$TMP_FILE"
        exit 1
    fi

    NEW_MD5=$(get_file_md5 "$TMP_FILE")
    echo -e "${GREEN}下载文件 MD5: ${NEW_MD5}${NC}"

    BINARY_FILE="${INSTALL_DIR}/tgbot"
    if [ -f "$BINARY_FILE" ]; then
        OLD_MD5=$(get_file_md5 "$BINARY_FILE")
        echo -e "${YELLOW}已安装文件 MD5: ${OLD_MD5}${NC}"

        if [ "$NEW_MD5" = "$OLD_MD5" ] && [ -n "$OLD_MD5" ]; then
            echo -e "${GREEN}文件已存在且 MD5 相同，跳过安装${NC}"
            rm -f "$TMP_FILE"
            return
        else
            echo -e "${YELLOW}文件已存在但 MD5 不同，将覆盖安装${NC}"
        fi
    else
        echo -e "${YELLOW}文件不存在，开始安装...${NC}"
    fi

    echo -e "${YELLOW}正在解压...${NC}"
    tar -xzf "$TMP_FILE" -C "$INSTALL_DIR"
    rm -f "$TMP_FILE"

    chmod +x "$BINARY_FILE"
    echo -e "${GREEN}解压完成${NC}"
}

create_service() {
    echo -e "${YELLOW}创建系统服务...${NC}"

    cd $INSTALL_DIR && ./tgbot install

    echo -e "${GREEN}服务创建完成${NC}"
}

show_info() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}       tgbot 安装完成！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${YELLOW}服务信息:${NC}"
    echo -e "  服务名称: ${SERVICE_NAME}"
    echo -e "  安装目录: ${INSTALL_DIR}"
    echo -e "  数据目录: ${DATA_DIR}"
    echo -e "  配置文件: ${CONF_FILE}"
    echo ""
    echo -e "${YELLOW}访问地址:${NC}"
    echo -e "  http://$(hostname -I | awk '{print $1}'):9393"
    echo ""
    echo -e "${YELLOW}服务管理:${NC}"
    echo -e "  启动: systemctl start ${SERVICE_NAME}"
    echo -e "  停止: systemctl stop ${SERVICE_NAME}"
    echo -e "  重启: systemctl restart ${SERVICE_NAME}"
    echo -e "  状态: systemctl status ${SERVICE_NAME}"
    echo -e "  日志: journalctl -u ${SERVICE_NAME} -f"
    echo ""
    echo -e "${GREEN}========================================${NC}"
}

main() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}        tgbot 一键安装脚本${NC}"
    echo -e "${GREEN}        Version: ${VERSION}${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""

    if [ "$(id -u)" != "0" ]; then
        echo -e "${RED}错误: 请使用 root 用户运行此脚本${NC}"
        exit 1
    fi

    detect_arch
    detect_os
    install_deps
    create_dirs
    download_and_extract
    create_service
    systemctl start tgbot
    show_info
}

main "$@"