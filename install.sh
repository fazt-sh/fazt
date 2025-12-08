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

# Check if this is an upgrade scenario
EXISTING_BINARY="/usr/local/bin/fazt"
SERVICE_FILE="/etc/systemd/system/fazt.service"
IS_UPGRADE=false
SERVICE_WAS_ACTIVE=false

if [ -f "$EXISTING_BINARY" ] || [ -f "$SERVICE_FILE" ]; then
    IS_UPGRADE=true

    # Check current version
    if [ -f "$EXISTING_BINARY" ]; then
        CURRENT_VERSION=$($EXISTING_BINARY --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
        echo -e "${BLUE}ℹ${NC} Existing installation detected (v${CURRENT_VERSION})"
    fi

    # Check if service is active
    if command -v systemctl >/dev/null 2>&1; then
        if systemctl is-active --quiet fazt 2>/dev/null; then
            SERVICE_WAS_ACTIVE=true
            echo -e "${BLUE}ℹ${NC} Service is currently running"
        fi
    fi
fi

# 1. Get the latest release tag URL
LATEST_URL="$(curl -sL -I -o /dev/null -w '%{url_effective}' https://github.com/fazt-sh/fazt/releases/latest)"
TAG="$(basename "$LATEST_URL")"

if [ -z "$TAG" ]; then
    echo -e "${RED}✗ Failed to find latest release tag.${NC}"
    exit 1
fi

# Extract version number for comparison
NEW_VERSION="${TAG#v}"

if [ "$IS_UPGRADE" = true ] && [ "$CURRENT_VERSION" = "$NEW_VERSION" ]; then
    echo -e "${GREEN}✓ Already running the latest version (${NEW_VERSION})${NC}"
    echo ""
    echo "Nothing to do!"
    exit 0
fi

# 2. Construct download URL
FILE_NAME="fazt-${TAG}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/fazt-sh/fazt/releases/download/${TAG}/${FILE_NAME}"

if [ "$IS_UPGRADE" = true ]; then
    echo -e "${YELLOW}⚡${NC} ${BOLD}Upgrading to ${TAG}${NC}"
else
    echo -e "${BLUE}ℹ${NC} Downloading ${BOLD}${TAG}${NC} for ${BOLD}${OS}/${ARCH}${NC}..."
fi

# 3. Download and extract
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

if curl -sL --fail "$DOWNLOAD_URL" -o "$TMP_DIR/$FILE_NAME"; then
    tar -xzf "$TMP_DIR/$FILE_NAME" -C "$TMP_DIR"

    if [ -f "$TMP_DIR/$BINARY_NAME" ]; then
        mv "$TMP_DIR/$BINARY_NAME" "$TMP_DIR/fazt-new"
    elif [ -f "$TMP_DIR/fazt-${OS}-${ARCH}" ]; then
        mv "$TMP_DIR/fazt-${OS}-${ARCH}" "$TMP_DIR/fazt-new"
    else
        FOUND="$(find "$TMP_DIR" -type f -perm -u+x | head -n 1)"
        if [ -n "$FOUND" ]; then
            mv "$FOUND" "$TMP_DIR/fazt-new"
        else
            echo -e "${RED}✗ Could not find binary in archive.${NC}"
            exit 1
        fi
    fi
else
    echo -e "${RED}✗ Download failed!${NC}"
    exit 1
fi

chmod +x "$TMP_DIR/fazt-new"
echo -e "${GREEN}✓ Downloaded ${TAG}${NC}"
echo ""

# Handle upgrade path
if [ "$IS_UPGRADE" = true ]; then
    echo -e "${BLUE}ℹ${NC} Performing upgrade..."

    # Stop service if it was running
    if [ "$SERVICE_WAS_ACTIVE" = true ]; then
        echo -e "${BLUE}ℹ${NC} Stopping service..."
        if [ "$EUID" -ne 0 ]; then
            sudo systemctl stop fazt
        else
            systemctl stop fazt
        fi
    fi

    # Backup current binary
    if [ -f "$EXISTING_BINARY" ]; then
        if [ "$EUID" -ne 0 ]; then
            sudo cp "$EXISTING_BINARY" "$EXISTING_BINARY.old"
        else
            cp "$EXISTING_BINARY" "$EXISTING_BINARY.old"
        fi
        echo -e "${BLUE}ℹ${NC} Backed up old binary to $EXISTING_BINARY.old"
    fi

    # Replace binary
    if [ "$EUID" -ne 0 ]; then
        sudo mv "$TMP_DIR/fazt-new" "$EXISTING_BINARY"
        sudo chmod +x "$EXISTING_BINARY"

        # Apply capabilities for port binding (Linux only)
        if [ "$OS" = "linux" ]; then
            sudo setcap CAP_NET_BIND_SERVICE=+eip "$EXISTING_BINARY" 2>/dev/null || true
        fi
    else
        mv "$TMP_DIR/fazt-new" "$EXISTING_BINARY"
        chmod +x "$EXISTING_BINARY"

        if [ "$OS" = "linux" ]; then
            setcap CAP_NET_BIND_SERVICE=+eip "$EXISTING_BINARY" 2>/dev/null || true
        fi
    fi

    echo -e "${GREEN}✓ Binary updated${NC}"

    # Restart service if it was running
    if [ "$SERVICE_WAS_ACTIVE" = true ]; then
        echo -e "${BLUE}ℹ${NC} Restarting service..."
        if [ "$EUID" -ne 0 ]; then
            sudo systemctl start fazt
        else
            systemctl start fazt
        fi

        # Give it a moment to start
        sleep 2

        # Check if service started successfully
        if systemctl is-active --quiet fazt; then
            echo -e "${GREEN}✓ Service restarted successfully${NC}"
        else
            echo -e "${RED}✗ Service failed to start. Check logs with: journalctl -u fazt -n 50${NC}"
            exit 1
        fi
    fi

    echo ""
    echo -e "${GREEN}✓ Upgrade complete!${NC}"
    echo ""
    echo -e "${BOLD}Upgraded from v${CURRENT_VERSION} to ${TAG}${NC}"
    echo ""
    echo "Check status: systemctl status fazt"
    echo "View logs: journalctl -u fazt -f"

    exit 0
fi

# Fresh installation path (not an upgrade)
echo -e "${BLUE}ℹ${NC} Installing to /usr/local/bin/fazt..."

if [ "$EUID" -ne 0 ]; then
    if command -v sudo >/dev/null 2>&1; then
        sudo mv "$TMP_DIR/fazt-new" "$EXISTING_BINARY"
        BINARY_PATH="fazt"
    else
        echo -e "${YELLOW}sudo not found. Staying in current directory.${NC}"
        mv "$TMP_DIR/fazt-new" "./$BINARY_NAME"
        BINARY_PATH="./fazt"
    fi
else
    mv "$TMP_DIR/fazt-new" "$EXISTING_BINARY"
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
read -p "> Select [1/2]: " MODE < /dev/tty

if [ "$MODE" = "1" ]; then
    echo ""
    read -p "Domain or IP (e.g. my-paas.com): " DOMAIN < /dev/tty
    read -p "Email (Enter to skip for HTTP): " EMAIL < /dev/tty
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
    read -p "Do you want to connect to a remote server? [y/N] " CONFIRM < /dev/tty

    if [[ "$CONFIRM" =~ ^[Yy]$ ]]; then
        echo ""
        read -p "Server URL (e.g. https://my-paas.com): " URL < /dev/tty
        read -p "API Token: " TOKEN < /dev/tty
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
