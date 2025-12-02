#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

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

echo -e "${GREEN}Detected $OS/$ARCH...${NC}"

# 1. Get the latest release tag URL
LATEST_URL=$(curl -sL -I -o /dev/null -w '%{url_effective}' https://github.com/fazt-sh/fazt/releases/latest)

# 2. Extract tag from URL (e.g., https://.../tag/v0.5.2 -> v0.5.2)
TAG=$(basename "$LATEST_URL")

if [ -z "$TAG" ]; then
    echo -e "${RED}Failed to find latest release tag.${NC}"
    exit 1
fi

echo -e "Latest version: ${GREEN}$TAG${NC}"

# 3. Construct download URL
# Format: fazt-<TAG>-<OS>-<ARCH>.tar.gz
FILE_NAME="fazt-${TAG}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/fazt-sh/fazt/releases/download/${TAG}/${FILE_NAME}"

echo -e "${GREEN}Downloading $DOWNLOAD_URL...${NC}"

# 4. Download and extract
# We use a temporary directory to handle extraction cleanly
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

if curl -sL --fail "$DOWNLOAD_URL" -o "$TMP_DIR/$FILE_NAME"; then
    tar -xzf "$TMP_DIR/$FILE_NAME" -C "$TMP_DIR"
    
    # Move binary to current directory
    # The tarball might contain the binary directly or inside a folder
    # We look for the binary named 'fazt' or 'fazt-linux-amd64' etc
    if [ -f "$TMP_DIR/$BINARY_NAME" ]; then
        mv "$TMP_DIR/$BINARY_NAME" .
    elif [ -f "$TMP_DIR/fazt-${OS}-${ARCH}" ]; then
        mv "$TMP_DIR/fazt-${OS}-${ARCH}" ./$BINARY_NAME
    else
        # Fallback: find any executable file
        FOUND=$(find "$TMP_DIR" -type f -perm -u+x | head -n 1)
        if [ -n "$FOUND" ]; then
            mv "$FOUND" ./$BINARY_NAME
        else
            echo -e "${RED}Could not find binary in archive.${NC}"
            exit 1
        fi
    fi
else
    echo -e "${RED}Download failed! (404 Not Found or other error)${NC}"
    echo "This usually means the release asset for $OS/$ARCH is missing."
    exit 1
fi

# Make executable
chmod +x "$BINARY_NAME"

echo -e "${GREEN}Download complete!${NC}"
echo ""
echo "To install as a system service (Production):"
echo "  sudo ./fazt service install --domain example.com --email admin@example.com --https"
echo ""
echo "To run manually (Development):"
echo "  ./fazt server init --username admin --password secret"
echo "  ./fazt server start"