#!/usr/bin/env bash
# Snip one-line installer
# Usage: curl -L https://snip.dev/install.sh | bash
# Or with options: curl -L ... | bash -s -- --port 8080 --data /var/lib/snip
set -euo pipefail

REPO="${SNIP_REPO:-Vesper36/snip}"
INSTALL_DIR="${SNIP_INSTALL_DIR:-/usr/local/bin}"
DATA_DIR="${SNIP_DATA_DIR:-/var/lib/snip}"
PORT="${SNIP_PORT:-53524}"

# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        --port) PORT="$2"; shift 2 ;;
        --data-dir) DATA_DIR="$2"; shift 2 ;;
        --install-dir) INSTALL_DIR="$2"; shift 2 ;;
        --repo) REPO="$2"; shift 2 ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    linux) ;;
    darwin) OS="darwin" ;;
    *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported arch: $ARCH" >&2; exit 1 ;;
esac

echo "==> Detected: ${OS}/${ARCH}"
echo "==> Installing to: ${INSTALL_DIR}/snip"
echo "==> Data directory: ${DATA_DIR}"
echo "==> Port: ${PORT}"

# Get latest release
LATEST_URL=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep "browser_download_url" \
    | grep "${OS}-${ARCH}\"" \
    | head -1 \
    | cut -d '"' -f 4)

if [ -z "$LATEST_URL" ]; then
    echo "No release found for ${OS}/${ARCH}" >&2
    echo "Try building from source: go install github.com/${REPO}/cmd/server@latest" >&2
    exit 1
fi

# Download
TMP=$(mktemp)
echo "==> Downloading from ${LATEST_URL}"
curl -sL "$LATEST_URL" -o "$TMP"
chmod +x "$TMP"

# Install
mkdir -p "$INSTALL_DIR" "$DATA_DIR"
mv "$TMP" "${INSTALL_DIR}/snip"

# Optionally install systemd service (Linux only)
if [ "$OS" = "linux" ] && [ -d /etc/systemd/system ] && command -v systemctl >/dev/null 2>&1; then
    cat > /etc/systemd/system/snip.service <<EOF
[Unit]
Description=Snip - code & file sharing
After=network.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/snip
Environment=SNIP_PORT=${PORT}
Environment=SNIP_DB_PATH=${DATA_DIR}/snip.db
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    if systemctl enable --now snip; then
        echo "==> Installed and started systemd service"
        echo "    Check status: systemctl status snip"
        echo "    View logs:    journalctl -u snip -f"
    fi
elif [ "$OS" = "darwin" ] && command -v brew >/dev/null 2>&1; then
    echo "==> Installed to ${INSTALL_DIR}/snip"
    echo "    Run with: SNIP_PORT=${PORT} ${INSTALL_DIR}/snip"
else
    echo "==> Installed to ${INSTALL_DIR}/snip"
    echo "    Run with: SNIP_PORT=${PORT} ${INSTALL_DIR}/snip"
fi

echo ""
echo "Snip is now available. Open http://localhost:${PORT}"
