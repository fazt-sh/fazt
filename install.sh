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
echo -e "${DIM}  Sovereign Compute for Individuals${NC}"
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

# Paths
SYSTEM_BINARY="/usr/local/bin/fazt"
USER_BINARY="$HOME/.local/bin/fazt"
SYSTEM_SERVICE="/etc/systemd/system/fazt.service"
USER_SERVICE="$HOME/.config/systemd/user/fazt-local.service"
USER_DB="$HOME/.config/fazt/data.db"

# Detect local IP (prefer 192.168.x.x or 10.x.x.x)
detect_local_ip() {
    local ip
    # Try to get primary local IP
    ip=$(ip route get 1 2>/dev/null | awk '{print $7; exit}')
    if [ -z "$ip" ]; then
        # Fallback: get first non-loopback IP
        ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    fi
    if [ -z "$ip" ]; then
        ip="127.0.0.1"
    fi
    echo "$ip"
}

# Check for existing installations
check_existing() {
    SYSTEM_EXISTS=false
    USER_EXISTS=false
    SYSTEM_RUNNING=false
    USER_RUNNING=false
    CURRENT_VERSION="none"

    # Check system installation
    if [ -f "$SYSTEM_BINARY" ]; then
        SYSTEM_EXISTS=true
        CURRENT_VERSION=$($SYSTEM_BINARY --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
        if command -v systemctl >/dev/null 2>&1; then
            if systemctl is-active --quiet fazt 2>/dev/null; then
                SYSTEM_RUNNING=true
            fi
        fi
    fi

    # Check user installation
    if [ -f "$USER_BINARY" ]; then
        USER_EXISTS=true
        if [ "$CURRENT_VERSION" = "none" ]; then
            CURRENT_VERSION=$($USER_BINARY --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
        fi
        if command -v systemctl >/dev/null 2>&1; then
            if systemctl --user is-active --quiet fazt-local 2>/dev/null; then
                USER_RUNNING=true
            fi
        fi
    fi
}

# Get latest release
get_latest_release() {
    LATEST_URL="$(curl -sL -I -o /dev/null -w '%{url_effective}' https://github.com/fazt-sh/fazt/releases/latest)"
    TAG="$(basename "$LATEST_URL")"

    if [ -z "$TAG" ]; then
        echo -e "${RED}Failed to find latest release.${NC}"
        exit 1
    fi

    NEW_VERSION="${TAG#v}"
}

# Download binary
download_binary() {
    FILE_NAME="fazt-${TAG}-${OS}-${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/fazt-sh/fazt/releases/download/${TAG}/${FILE_NAME}"

    echo -e "${BLUE}Downloading ${BOLD}${TAG}${NC}${BLUE} for ${BOLD}${OS}/${ARCH}${NC}..."

    TMP_DIR="$(mktemp -d)"
    trap 'rm -rf "$TMP_DIR"' EXIT

    if ! curl -sL --fail "$DOWNLOAD_URL" -o "$TMP_DIR/$FILE_NAME"; then
        echo -e "${RED}Download failed!${NC}"
        exit 1
    fi

    tar -xzf "$TMP_DIR/$FILE_NAME" -C "$TMP_DIR"

    # Find the binary in extracted files
    if [ -f "$TMP_DIR/fazt" ]; then
        DOWNLOADED_BINARY="$TMP_DIR/fazt"
    elif [ -f "$TMP_DIR/fazt-${OS}-${ARCH}" ]; then
        DOWNLOADED_BINARY="$TMP_DIR/fazt-${OS}-${ARCH}"
    else
        DOWNLOADED_BINARY="$(find "$TMP_DIR" -type f -perm -u+x | head -n 1)"
    fi

    if [ -z "$DOWNLOADED_BINARY" ] || [ ! -f "$DOWNLOADED_BINARY" ]; then
        echo -e "${RED}Could not find binary in archive.${NC}"
        exit 1
    fi

    chmod +x "$DOWNLOADED_BINARY"
    echo -e "${GREEN}Downloaded ${TAG}${NC}"
}

# Install for local development
install_local_dev() {
    echo ""
    echo -e "${BOLD}Local Development Setup${NC}"
    echo ""

    # Detect local IP
    LOCAL_IP=$(detect_local_ip)
    echo -e "${BLUE}Detected local IP: ${BOLD}${LOCAL_IP}${NC}"
    read -p "Use this IP? [Y/n]: " CONFIRM < /dev/tty

    if [[ "$CONFIRM" =~ ^[Nn]$ ]]; then
        read -p "Enter IP address: " LOCAL_IP < /dev/tty
    fi

    # Port selection
    read -p "Port [8080]: " PORT < /dev/tty
    PORT=${PORT:-8080}

    # Create directories
    mkdir -p "$HOME/.local/bin"
    mkdir -p "$HOME/.config/fazt"
    mkdir -p "$HOME/.config/systemd/user"

    # Install binary
    echo -e "${BLUE}Installing binary to ${USER_BINARY}...${NC}"
    cp "$DOWNLOADED_BINARY" "$USER_BINARY"
    chmod +x "$USER_BINARY"

    # Ensure ~/.local/bin is in PATH
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        echo -e "${YELLOW}Note: Add ~/.local/bin to your PATH:${NC}"
        echo '  export PATH="$HOME/.local/bin:$PATH"'
        echo ""
    fi

    # Initialize database if not exists
    if [ ! -f "$USER_DB" ]; then
        echo -e "${BLUE}Initializing database...${NC}"
        "$USER_BINARY" server init \
            --username dev \
            --password dev \
            --domain "$LOCAL_IP" \
            --db "$USER_DB"
    else
        echo -e "${BLUE}Database already exists at ${USER_DB}${NC}"
    fi

    # Create user systemd service
    echo -e "${BLUE}Creating systemd user service...${NC}"
    cat > "$USER_SERVICE" << EOF
[Unit]
Description=Fazt Local Development Server
After=network.target

[Service]
Type=simple
ExecStart=${USER_BINARY} server start --port ${PORT} --domain ${LOCAL_IP} --db ${USER_DB}
Restart=always
RestartSec=5
WorkingDirectory=${HOME}/.config/fazt
Environment=FAZT_ENV=development

[Install]
WantedBy=default.target
EOF

    # Enable and start service
    echo -e "${BLUE}Enabling service...${NC}"
    systemctl --user daemon-reload
    systemctl --user enable fazt-local
    systemctl --user start fazt-local

    # Enable linger for persistence across reboots
    if command -v loginctl >/dev/null 2>&1; then
        loginctl enable-linger "$(whoami)" 2>/dev/null || true
    fi

    # Wait for service to start
    sleep 2

    # Verify
    if systemctl --user is-active --quiet fazt-local; then
        echo ""
        echo -e "${GREEN}Local development server installed and running!${NC}"
        echo ""
        echo -e "  ${BOLD}Dashboard:${NC}  http://admin.${LOCAL_IP}.nip.io:${PORT}"
        echo -e "  ${BOLD}Apps:${NC}       http://<app>.${LOCAL_IP}.nip.io:${PORT}"
        echo -e "  ${BOLD}Database:${NC}   ${USER_DB}"
        echo ""
        echo -e "  ${DIM}Credentials: dev / dev${NC}"
        echo ""
        echo -e "${BOLD}Commands:${NC}"
        echo "  systemctl --user status fazt-local   # Check status"
        echo "  systemctl --user restart fazt-local  # Restart"
        echo "  journalctl --user -u fazt-local -f   # View logs"
    else
        echo -e "${RED}Service failed to start. Check logs:${NC}"
        echo "  journalctl --user -u fazt-local -n 50"
        exit 1
    fi
}

# Install for production server
install_production() {
    echo ""
    echo -e "${BOLD}Production Server Setup${NC}"
    echo -e "${DIM}(Requires sudo)${NC}"
    echo ""

    read -p "Domain (e.g. my-paas.com): " DOMAIN < /dev/tty
    read -p "Email (Enter to skip HTTPS): " EMAIL < /dev/tty
    echo ""

    # Install binary to system path
    echo -e "${BLUE}Installing binary to ${SYSTEM_BINARY}...${NC}"
    if [ "$EUID" -ne 0 ]; then
        sudo cp "$DOWNLOADED_BINARY" "$SYSTEM_BINARY"
        sudo chmod +x "$SYSTEM_BINARY"
        if [ "$OS" = "linux" ]; then
            sudo setcap CAP_NET_BIND_SERVICE=+eip "$SYSTEM_BINARY" 2>/dev/null || true
        fi
    else
        cp "$DOWNLOADED_BINARY" "$SYSTEM_BINARY"
        chmod +x "$SYSTEM_BINARY"
        if [ "$OS" = "linux" ]; then
            setcap CAP_NET_BIND_SERVICE=+eip "$SYSTEM_BINARY" 2>/dev/null || true
        fi
    fi

    # Run service install command
    CMD="$SYSTEM_BINARY service install --domain $DOMAIN"
    if [ -n "$EMAIL" ]; then
        CMD="$CMD --email $EMAIL --https"
    fi

    if [ "$EUID" -ne 0 ]; then
        sudo $CMD
    else
        $CMD
    fi
}

# Install CLI only
install_cli_only() {
    echo ""
    echo -e "${BOLD}CLI Tool Installation${NC}"
    echo ""

    # Prefer user directory, fall back to system
    mkdir -p "$HOME/.local/bin"
    echo -e "${BLUE}Installing binary to ${USER_BINARY}...${NC}"
    cp "$DOWNLOADED_BINARY" "$USER_BINARY"
    chmod +x "$USER_BINARY"

    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        echo -e "${YELLOW}Note: Add ~/.local/bin to your PATH:${NC}"
        echo '  export PATH="$HOME/.local/bin:$PATH"'
        echo ""
    fi

    echo -e "${GREEN}Binary installed!${NC}"
    echo ""

    read -p "Connect to a remote server? [y/N]: " CONFIRM < /dev/tty

    if [[ "$CONFIRM" =~ ^[Yy]$ ]]; then
        echo ""
        read -p "Server URL (e.g. https://my-paas.com): " URL < /dev/tty
        read -p "API Token: " TOKEN < /dev/tty
        echo ""
        "$USER_BINARY" remote add default --url "$URL" --token "$TOKEN"
        echo -e "${GREEN}Remote configured!${NC}"
    else
        echo ""
        echo "Run 'fazt --help' to get started."
    fi
}

# Upgrade existing installation
upgrade_existing() {
    local target_binary="$1"
    local service_name="$2"
    local is_user_service="$3"

    echo -e "${YELLOW}Upgrading from v${CURRENT_VERSION} to v${NEW_VERSION}${NC}"
    echo ""

    # Stop service if running
    if [ "$is_user_service" = "true" ] && [ "$USER_RUNNING" = "true" ]; then
        echo -e "${BLUE}Stopping user service...${NC}"
        systemctl --user stop fazt-local
    elif [ "$SYSTEM_RUNNING" = "true" ]; then
        echo -e "${BLUE}Stopping system service...${NC}"
        if [ "$EUID" -ne 0 ]; then
            sudo systemctl stop fazt
        else
            systemctl stop fazt
        fi
    fi

    # Backup and replace binary
    if [ -f "$target_binary" ]; then
        if [ "$is_user_service" = "true" ]; then
            cp "$target_binary" "${target_binary}.old"
            cp "$DOWNLOADED_BINARY" "$target_binary"
            chmod +x "$target_binary"
        else
            if [ "$EUID" -ne 0 ]; then
                sudo cp "$target_binary" "${target_binary}.old"
                sudo cp "$DOWNLOADED_BINARY" "$target_binary"
                sudo chmod +x "$target_binary"
            else
                cp "$target_binary" "${target_binary}.old"
                cp "$DOWNLOADED_BINARY" "$target_binary"
                chmod +x "$target_binary"
            fi
        fi
    fi

    echo -e "${GREEN}Binary updated${NC}"

    # Restart service
    if [ "$is_user_service" = "true" ] && [ "$USER_RUNNING" = "true" ]; then
        echo -e "${BLUE}Restarting user service...${NC}"
        systemctl --user start fazt-local
        sleep 2
        if systemctl --user is-active --quiet fazt-local; then
            echo -e "${GREEN}Service restarted successfully${NC}"
        else
            echo -e "${RED}Service failed to start${NC}"
        fi
    elif [ "$SYSTEM_RUNNING" = "true" ]; then
        echo -e "${BLUE}Restarting system service...${NC}"
        if [ "$EUID" -ne 0 ]; then
            sudo systemctl start fazt
        else
            systemctl start fazt
        fi
        sleep 2
        if systemctl is-active --quiet fazt; then
            echo -e "${GREEN}Service restarted successfully${NC}"
        else
            echo -e "${RED}Service failed to start${NC}"
        fi
    fi

    echo ""
    echo -e "${GREEN}Upgrade complete! v${CURRENT_VERSION} → v${NEW_VERSION}${NC}"
}

# Main flow
check_existing
get_latest_release

# Handle existing installation
if [ "$SYSTEM_EXISTS" = "true" ] || [ "$USER_EXISTS" = "true" ]; then
    echo -e "${BLUE}Existing installation detected (v${CURRENT_VERSION})${NC}"

    if [ "$CURRENT_VERSION" = "$NEW_VERSION" ]; then
        echo -e "${GREEN}Already running latest version (v${NEW_VERSION})${NC}"
        echo ""
        echo -e "${BOLD}What would you like to do?${NC}"
        echo "1. Reinstall/reconfigure"
        echo "2. Exit"
        read -p "> Select [1/2]: " CHOICE < /dev/tty

        if [ "$CHOICE" != "1" ]; then
            echo "Nothing to do."
            exit 0
        fi
    else
        echo -e "${YELLOW}New version available: v${NEW_VERSION}${NC}"
        echo ""
        echo -e "${BOLD}What would you like to do?${NC}"
        echo "1. Upgrade (keep configuration)"
        echo "2. Fresh install (reconfigure)"
        echo "3. Exit"
        read -p "> Select [1/2/3]: " CHOICE < /dev/tty

        if [ "$CHOICE" = "3" ]; then
            exit 0
        fi

        if [ "$CHOICE" = "1" ]; then
            download_binary
            if [ "$USER_EXISTS" = "true" ]; then
                upgrade_existing "$USER_BINARY" "fazt-local" "true"
            else
                upgrade_existing "$SYSTEM_BINARY" "fazt" "false"
            fi
            exit 0
        fi
    fi
fi

# Fresh installation
download_binary

echo ""
echo -e "${BOLD}Select Installation Type:${NC}"
echo ""
echo "1. ${BOLD}Production Server${NC}"
echo "   For VPS/cloud. System service, real domain, HTTPS."
echo ""
echo "2. ${BOLD}Local Development${NC}"
echo "   For dev machines. User service, auto-start, IP-based."
echo ""
echo "3. ${BOLD}CLI Only${NC}"
echo "   Just the binary. Connect to remote servers."
echo ""
read -p "> Select [1/2/3]: " MODE < /dev/tty

case "$MODE" in
    1) install_production ;;
    2) install_local_dev ;;
    3) install_cli_only ;;
    *) echo "Invalid selection."; exit 1 ;;
esac
