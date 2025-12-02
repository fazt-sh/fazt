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
echo -e "${BLUE}"
echo "  __           _       _     "
echo " / _| __ _ ___| |_ ___| |__  "
echo "| |_ / _\` |_  / __/ __| '_ \ "
echo "|  _| (_| |/ /| |_\__ \ | | |"
echo "|_|  \__,_/___|\__|___/_| |_|"
echo -e "${NC}"
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

echo -e "${BLUE}ℹ${NC} Detected System: ${BOLD}$OS/$ARCH${NC}"

# 1. Get the latest release tag URL
LATEST_URL=$(curl -sL -I -o /dev/null -w '%{url_effective}' https://github.com/fazt-sh/fazt/releases/latest)
TAG=$(basename "$LATEST_URL")

if [ -z "$TAG" ]; then
    echo -e "${RED}✗ Failed to find latest release tag.${NC}"
    exit 1
fi

echo -e "${BLUE}ℹ${NC} Latest Version:  ${GREEN}$TAG${NC}"

# 3. Construct download URL
FILE_NAME="fazt-${TAG}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/fazt-sh/fazt/releases/download/${TAG}/${FILE_NAME}"

echo -e "${BLUE}ℹ${NC} Downloading...   ${DIM}$DOWNLOAD_URL${NC}"

# 4. Download and extract
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

if curl -sL --fail "$DOWNLOAD_URL" -o "$TMP_DIR/$FILE_NAME"; then
    tar -xzf "$TMP_DIR/$FILE_NAME" -C "$TMP_DIR"
    
    if [ -f "$TMP_DIR/$BINARY_NAME" ]; then
        mv "$TMP_DIR/$BINARY_NAME" .
    elif [ -f "$TMP_DIR/fazt-${OS}-${ARCH}" ]; then
        mv "$TMP_DIR/fazt-${OS}-${ARCH}" ./$BINARY_NAME
    else
        FOUND=$(find "$TMP_DIR" -type f -perm -u+x | head -n 1)
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

echo -e "${GREEN}✓ Download complete!${NC}"
echo ""

echo -e "${BOLD}Next Steps:${NC}"
echo -e "${DIM}──────────────────────────────────────────────${NC}"
echo -e "${BOLD}1. Production Install (Recommended)${NC}"
echo -e "   sudo ./fazt service install --domain example.com --email me@example.com --https"
echo ""
echo -e "${BOLD}2. Manual Run (Development)${NC}"
echo -e "   ./fazt server init --username admin --password secret --domain localhost"
echo -e "   ./fazt server start"
echo ""