#!/bin/bash
set -e

# Fazt.sh Installer
# Usage: curl -s https://fazt-sh.github.io/fazt/install.sh | sudo bash

# Colors
GREEN='\033[1;32m'
BLUE='\033[1;34m'
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

# Find Latest Release
RELEASE_URL="https://github.com/fazt-sh/fazt/releases/latest"
LATEST_TAG=$(curl -sL -I -o /dev/null -w '%{url_effective}' "$RELEASE_URL" | grep -o "v[0-9.]*$")

if [ -z "$LATEST_TAG" ]; then
    echo -e "${RED}Failed to find latest release.${NC}"
    exit 1
fi

echo -e "Version: ${GREEN}$LATEST_TAG${NC}"

# Download
echo "Downloading..."
FILE="fazt-${LATEST_TAG}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/fazt-sh/fazt/releases/download/${LATEST_TAG}/${FILE}"
TMP=$(mktemp -d)

curl -sL "$URL" -o "$TMP/$FILE"
tar -xzf "$TMP/$FILE" -C "$TMP"

# Install
echo "Installing to /usr/local/bin/fazt..."
mv "$TMP/fazt" /usr/local/bin/fazt
chmod +x /usr/local/bin/fazt

# Cleanup
rm -rf "$TMP"

echo -e "${GREEN}✓ Installed successfully!${NC}"
echo ""
echo -e "To configure and start the service, run:"
echo -e "${BOLD}fazt service install --domain example.com --email you@example.com --https${NC}"
