#!/bin/bash
set -e

# Fazt.sh Installer / Upgrader
# Usage: curl -s https://fazt-sh.github.io/fazt/install.sh | sudo bash

# Colors
GREEN='\033[1;32m'
BLUE='\033[1;34m'
YELLOW='\033[1;33m'
RED='\033[1;31m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${BLUE}⚡ Fazt.sh Installer${NC}"

# Check Root
if [ "$EUID" -ne 0 ]; then 
  echo -e "${RED}Please run as root (sudo)${NC}"
  exit 1
fi

# Detect Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi
if [ "$ARCH" = "aarch64" ]; then ARCH="arm64"; fi

echo -e "Target: ${BOLD}$OS/$ARCH${NC}"

# Detect Existing Installation
IS_UPGRADE=0
if command -v fazt >/dev/null 2>&1; then
    IS_UPGRADE=1
    CURRENT_VER=$(fazt --version 2>/dev/null | grep -o "v[0-9.]*" || echo "unknown")
    echo -e "Current: ${YELLOW}$CURRENT_VER${NC}"
fi

# Find Latest Release
RELEASE_URL="https://github.com/fazt-sh/fazt/releases/latest"
LATEST_TAG=$(curl -sL -I -o /dev/null -w '%{url_effective}' "$RELEASE_URL" | grep -o "v[0-9.]*$")

if [ -z "$LATEST_TAG" ]; then
    echo -e "${RED}Failed to find latest release.${NC}"
    exit 1
fi

if [ "$IS_UPGRADE" -eq 1 ] && [ "$CURRENT_VER" == "$LATEST_TAG" ]; then
    echo -e "${GREEN}✓ You are already on the latest version ($LATEST_TAG)${NC}"
    # Optional: Force reinstall flag? For now just exit or continue?
    # Let's continue to allow repairing/reinstalling.
    echo "Reinstalling..."
else
    echo -e "Latest:  ${GREEN}$LATEST_TAG${NC}"
fi

# Download
echo "Downloading..."
FILE="fazt-${LATEST_TAG}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/fazt-sh/fazt/releases/download/${LATEST_TAG}/${FILE}"
TMP=$(mktemp -d)

if ! curl -sL "$URL" -o "$TMP/$FILE"; then
    echo -e "${RED}Download failed.${NC}"
    exit 1
fi
tar -xzf "$TMP/$FILE" -C "$TMP"

# Install / Upgrade
echo "Installing binary to /usr/local/bin/fazt..."
mv "$TMP/fazt" /usr/local/bin/fazt
chmod +x /usr/local/bin/fazt

# Cleanup
rm -rf "$TMP"

# Handle Service Restart
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet fazt; then
        echo "Restarting fazt service..."
        systemctl restart fazt
        echo -e "${GREEN}✓ Service restarted.${NC}"
    fi
fi

echo ""
if [ "$IS_UPGRADE" -eq 1 ]; then
    echo -e "${GREEN}✓ Upgrade complete!${NC} ($LATEST_TAG)"
else
    echo -e "${GREEN}✓ Installation complete!${NC} ($LATEST_TAG)"
    echo ""
    echo -e "To configure and start the service, run:"
    echo -e "${BOLD}fazt service install --domain example.com --email you@example.com --https${NC}"
fi
echo ""
