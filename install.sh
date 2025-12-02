#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
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
RELEASE_TAG="latest"

echo -e "${GREEN}Detected $OS/$ARCH...${NC}"

# Fetch latest release URL
# We cheat a bit and use the GitHub releases structure
DOWNLOAD_URL="https://github.com/fazt-sh/fazt/releases/latest/download/fazt-${OS}-${ARCH}.tar.gz"

echo -e "${GREEN}Downloading from $DOWNLOAD_URL...${NC}"

# Download and extract
curl -sL "$DOWNLOAD_URL" | tar xz

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
