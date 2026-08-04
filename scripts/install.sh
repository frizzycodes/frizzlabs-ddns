#!/usr/bin/env bash
set -euo pipefail

# Automated installation script for frizzlabs-ddns on Linux systems

BINARY_NAME="frizzlabs-ddns"
INSTALL_BIN_DIR="/usr/local/bin"
CONFIG_DIR="/etc/frizzlabs-ddns"
STATE_DIR="/var/lib/frizzlabs-ddns"
SYSTEMD_DIR="/etc/systemd/system"

# Automatically resolve repository root directory if run from scripts/ subdirectory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" 2>/dev/null && pwd || echo "")"
if [[ -n "${SCRIPT_DIR}" && -d "${SCRIPT_DIR}/../cmd" ]]; then
    cd "${SCRIPT_DIR}/.."
fi

echo "=== Installing ${BINARY_NAME} ==="

if [[ $EUID -ne 0 ]]; then
   echo "Error: This script must be run as root (or via sudo)." >&2
   exit 1
fi

if ! command -v go >/dev/null 2>&1; then
    echo "Error: Go compiler is required for build. Please install Go 1.21+." >&2
    exit 1
fi

echo "--> Building ${BINARY_NAME}..."
go build -ldflags="-s -w" -o "bin/${BINARY_NAME}" ./cmd/frizzlabs-ddns

echo "--> Installing binary to ${INSTALL_BIN_DIR}/${BINARY_NAME}..."
install -d "${INSTALL_BIN_DIR}"
install -m 0755 "bin/${BINARY_NAME}" "${INSTALL_BIN_DIR}/${BINARY_NAME}"

echo "--> Setting up configuration and state directories..."
install -d -m 0755 "${CONFIG_DIR}"
install -d -m 0755 "${STATE_DIR}"

if [[ ! -f "${CONFIG_DIR}/config.json" ]]; then
    echo "--> Installing template configuration to ${CONFIG_DIR}/config.json..."
    install -m 0600 "configs/config.example.json" "${CONFIG_DIR}/config.json"
else
    echo "--> Configuration file ${CONFIG_DIR}/config.json already exists. Skipping."
fi

if [[ -d "/run/systemd/system" ]]; then
    echo "--> Installing Systemd service and timer..."
    install -m 0644 "systemd/${BINARY_NAME}.service" "${SYSTEMD_DIR}/"
    install -m 0644 "systemd/${BINARY_NAME}.timer" "${SYSTEMD_DIR}/"
    
    echo "--> Reloading systemd manager configuration..."
    systemctl daemon-reload
    
    echo "=== Installation Successful ==="
    echo "Next steps:"
    echo "1. Edit your token and domain in ${CONFIG_DIR}/config.json"
    echo "2. Enable and start the timer:"
    echo "   sudo systemctl enable --now ${BINARY_NAME}.timer"
else
    echo "=== Installation Successful (Non-Systemd environment) ==="
    echo "Binary installed at ${INSTALL_BIN_DIR}/${BINARY_NAME}"
fi
