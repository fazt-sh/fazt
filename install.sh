#!/bin/bash
set -e

# Colors
GREEN='\033[1;32m'
BLUE='\033[1;34m'
YELLOW='\033[1;33m'
RED='\033[1;31m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m' # No Color

# Banner
echo -e "\033[38;5;196m________\033[0m            \033[38;5;202m_____","\033[0m"
echo -e "\033[38;5;196m___  __/","\033[38;5;208m_____ \033[38;5;214m________\033[38;5;220m  _/","\033[0m"
echo -e "\033[38;5;208m__  /_ \033[38;5;214m_  __ ","\`\033[38;5;220m/__  /\033[38;5;226m_  __/","\033[0m"
echo -e "\033[38;5;214m_  __/ \033[38;5;220m/ /_/ /\033[38;5;226m__  /_","\033[38;5;228m/ /_","\033[0m"
echo -e "\033[38;5;220m/_/    \033[38;5;226m\__,_/ \033[38;5;228m_____/\033[38;5;231m\__/","\033[0m"
echo -e "${DIM}  Single Binary PaaS & Analytics${NC}"
echo ""

# Detect OS and Arch
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux)     OS="linux" ;;
    Darwin)    OS="darwin" ;;
    *)         echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
    x86_64)    ARCH="amd64" ;;
    aarch64)   ARCH="arm64" ;;
    arm64)     ARCH="arm64" ;;
    *)         echo "Unsupported Architecture: $ARCH"; exit 1 ;;
esac

BINARY_NAME="fazt"

# 1. Get the latest release tag URL
LATEST_URL="$(curl -sL -I -o /dev/null -w '%{url_effective}' https://github.com/fazt-sh/fazt/releases/latest)"
TAG="$(basename "$LATEST_URL")"

if [ -z "$TAG" ]; then
    echo -e "${RED}✗ Failed to find latest release tag.${NC}"
    exit 1
fi

# 2. Construct download URL
FILE_NAME="fazt-${TAG}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/fazt-sh/fazt/releases/download/${TAG}/${FILE_NAME}"

echo -e "${BLUE}ℹ${NC} Downloading ${BOLD}${TAG}${NC} for ${BOLD}${OS}/${ARCH}${NC}..."

# 3. Download and extract
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

if curl -sL --fail "$DOWNLOAD_URL" -o "$TMP_DIR/$FILE_NAME"; then
    tar -xzf "$TMP_DIR/$FILE_NAME" -C "$TMP_DIR"
    
    if [ -f "$TMP_DIR/$BINARY_NAME" ]; then
        mv "$TMP_DIR/$BINARY_NAME" .
    elif [ -f "$TMP_DIR/fazt-${OS}-${ARCH}" ]; then
        mv "$TMP_DIR/fazt-${OS}-${ARCH}" ./$BINARY_NAME
    else
        FOUND="$(find "$TMP_DIR" -type f -perm -u+x | head -n 1)"
        if [ -n "$FOUND" ]; then
            mv "$FOUND" ./$BINARY_NAME
        else
            echo -e "${RED}✗ Could not find binary in archive.${NC}"
            exit 1
        fi
    fi
else
    echo -e "${RED}✗ Download failed!${NC}"
    exit 1
fi

chmod +x "$BINARY_NAME"
echo -e "${GREEN}✓ Downloaded ./fazt${NC}"
echo ""

# Move to path
echo -e "${BLUE}ℹ${NC} Installing to /usr/local/bin/fazt..."
BINARY_PATH="./fazt"

if [ "$EUID" -ne 0 ]; then
    if command -v sudo >/dev/null 2>&1; then
        sudo mv "$BINARY_NAME" /usr/local/bin/fazt
        BINARY_PATH="fazt"
    else
        echo -e "${YELLOW}sudo not found. Staying in current directory.${NC}"
    fi
else
    mv "$BINARY_NAME" /usr/local/bin/fazt
    BINARY_PATH="fazt"
fi

echo ""
echo -e "${BOLD}Select Installation Type:${NC}"
echo "1. Headless Server (Daemon)"
echo "   Best for VPS. Installs Systemd Service. Starts on boot."
echo ""
echo "2. Command Line Tool (Portable)"
echo "   Best for Laptops. Just installs the binary."
echo ""
read -p "> Select [1/2]: " MODE

if [ "$MODE" = "1" ]; then
    echo ""
    read -p "Domain or IP (e.g. my-paas.com): " DOMAIN
    read -p "Email (Enter to skip for HTTP): " EMAIL
    echo ""
    
    # Check sudo for service install
    CMD="$BINARY_PATH service install --domain $DOMAIN"
    if [ -n "$EMAIL" ]; then
        CMD="$CMD --email $EMAIL --https"
    fi
    
    if [ "$EUID" -ne 0 ]; then
        sudo $CMD
    else
        $CMD
    fi

elif [ "$MODE" = "2" ]; then
    echo ""
    read -p "Do you want to connect to a remote server? [y/N] " CONFIRM
    
    if [[ "$CONFIRM" =~ ^[Yy]$ ]]; then
        echo ""
        read -p "Server URL (e.g. https://my-paas.com): " URL
        read -p "API Token: " TOKEN
        echo ""
        $BINARY_PATH client set-auth-token --token "$TOKEN" --server "$URL"
        echo ""
        echo -e "${DIM}Note: Config saved to ./data.db (Use FAZT_DB_PATH env for global)${NC}"
    else
        echo ""
        echo -e "${GREEN}✓ Setup Complete.${NC}"
        echo ""
        echo "To run a local server:"
        echo "  fazt server init"
        echo "  fazt server start"
    fi

else
    echo ""
    echo "You can run 'fazt' manually."
fi
