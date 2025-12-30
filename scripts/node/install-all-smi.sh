#!/usr/bin/env bash
set -euo pipefail

# Install all-smi Multi-Vendor AI Accelerator Metrics Exporter
#
# This script downloads and installs all-smi as a systemd service
# on Linux systems. all-smi provides unified Prometheus metrics for
# multiple AI accelerator vendors (NVIDIA, AMD, Intel, etc.)
#
# Usage: ./install-all-smi.sh [OPTIONS]
#
# Options:
#   -v, --version VERSION    all-smi version (default: 0.5.0)
#   -p, --port PORT          Listen port (default: 9401)
#   -h, --help              Show this help message
#
# Reference: https://github.com/lablup/all-smi

# Configuration
ALL_SMI_VERSION="${ALL_SMI_VERSION:-0.5.0}"
ALL_SMI_PORT="${ALL_SMI_PORT:-9401}"
INSTALL_DIR="/usr/local/bin"
SERVICE_USER="all_smi"
SERVICE_FILE="/etc/systemd/system/all-smi.service"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Print functions
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Print usage
usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Install all-smi Multi-Vendor AI Accelerator Metrics Exporter as a systemd service.

all-smi supports:
  - NVIDIA GPUs (CUDA)
  - AMD GPUs (ROCm)
  - Intel Gaudi NPUs
  - Google Cloud TPUs
  - Apple Silicon GPUs
  - Tenstorrent, Rebellions, Furiosa NPUs

Options:
    -v, --version VERSION    all-smi version (default: 0.5.0)
    -p, --port PORT          Listen port (default: 9401)
    -h, --help              Show this help message

Environment Variables:
    ALL_SMI_VERSION   - Version to install
    ALL_SMI_PORT      - Port to listen on

Example:
    $(basename "$0")
    $(basename "$0") --version 0.5.0 --port 9401

EOF
    exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            ALL_SMI_VERSION="$2"
            shift 2
            ;;
        -p|--port)
            ALL_SMI_PORT="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            error "Unknown option: $1"
            usage
            ;;
    esac
done

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    error "This script must be run as root"
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    *)
        error "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

info "Installing all-smi v${ALL_SMI_VERSION} for ${ARCH}"

# Check for Python3 (all-smi is a Python package)
if ! command -v python3 &>/dev/null; then
    error "Python3 is required but not installed"
    exit 1
fi

# Check for pip
if ! command -v pip3 &>/dev/null; then
    error "pip3 is required but not installed"
    exit 1
fi

# Install all-smi via pip
info "Installing all-smi via pip..."
if ! pip3 install "all-smi==${ALL_SMI_VERSION}" --quiet 2>/dev/null; then
    warn "Specific version not found, installing latest version..."
    if ! pip3 install all-smi --quiet; then
        error "Failed to install all-smi"
        exit 1
    fi
fi

# Find the all-smi binary location
ALL_SMI_BIN=$(which all-smi 2>/dev/null || echo "")
if [[ -z "$ALL_SMI_BIN" ]]; then
    # Try common pip install locations
    if [[ -f "/usr/local/bin/all-smi" ]]; then
        ALL_SMI_BIN="/usr/local/bin/all-smi"
    elif [[ -f "$HOME/.local/bin/all-smi" ]]; then
        ALL_SMI_BIN="$HOME/.local/bin/all-smi"
    else
        error "all-smi binary not found after installation"
        exit 1
    fi
fi

info "all-smi binary found at: $ALL_SMI_BIN"

# Create service user
if ! id "$SERVICE_USER" &>/dev/null; then
    info "Creating service user: $SERVICE_USER"
    useradd --no-create-home --shell /bin/false --system "$SERVICE_USER"
fi

# Add service user to video group for GPU access
if getent group video &>/dev/null; then
    info "Adding $SERVICE_USER to video group for GPU access"
    usermod -aG video "$SERVICE_USER"
fi

# Add service user to render group for GPU access (AMD/Intel)
if getent group render &>/dev/null; then
    info "Adding $SERVICE_USER to render group for GPU access"
    usermod -aG render "$SERVICE_USER"
fi

# Create systemd service file
info "Creating systemd service..."
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=all-smi Multi-Vendor AI Accelerator Metrics Exporter
Documentation=https://github.com/lablup/all-smi
Wants=network-online.target
After=network-online.target

[Service]
User=$SERVICE_USER
Group=$SERVICE_USER
Type=simple
ExecStart=$ALL_SMI_BIN serve --port $ALL_SMI_PORT
Restart=on-failure
RestartSec=5s

# Security hardening
NoNewPrivileges=true
ProtectHome=true
PrivateTmp=true

# Allow GPU device access
SupplementaryGroups=video render

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and start service
info "Reloading systemd daemon..."
systemctl daemon-reload

info "Enabling all-smi service..."
systemctl enable all-smi

info "Starting all-smi service..."
systemctl start all-smi

# Verify installation
sleep 2
if systemctl is-active --quiet all-smi; then
    info "all-smi installed and running successfully!"
    info "Metrics available at: http://localhost:${ALL_SMI_PORT}/metrics"

    # Test metrics endpoint
    if curl -s "http://localhost:${ALL_SMI_PORT}/metrics" > /dev/null 2>&1; then
        info "Metrics endpoint is responding"
    else
        warn "Metrics endpoint is not responding yet. Service may still be initializing."
    fi
else
    error "all-smi service failed to start"
    error "Check logs with: journalctl -u all-smi -f"
    exit 1
fi

# Detect available accelerators
info "Detecting available AI accelerators..."
DETECTED_ACCELERATORS=""

if command -v nvidia-smi &>/dev/null; then
    NVIDIA_COUNT=$(nvidia-smi --query-gpu=count --format=csv,noheader,nounits 2>/dev/null | head -1 || echo "0")
    if [[ "$NVIDIA_COUNT" -gt 0 ]]; then
        DETECTED_ACCELERATORS="${DETECTED_ACCELERATORS}NVIDIA GPU (${NVIDIA_COUNT}x), "
    fi
fi

if command -v rocm-smi &>/dev/null; then
    AMD_COUNT=$(rocm-smi --showproductname 2>/dev/null | grep -c "GPU" || echo "0")
    if [[ "$AMD_COUNT" -gt 0 ]]; then
        DETECTED_ACCELERATORS="${DETECTED_ACCELERATORS}AMD GPU (${AMD_COUNT}x), "
    fi
fi

if [[ -n "$DETECTED_ACCELERATORS" ]]; then
    info "Detected accelerators: ${DETECTED_ACCELERATORS%, }"
else
    warn "No AI accelerators detected. all-smi will report no metrics until accelerators are available."
fi

# Print next steps
cat <<EOF

${GREEN}Installation complete!${NC}

all-smi is now monitoring AI accelerators on this system.

Port: ${ALL_SMI_PORT}
Binary: ${ALL_SMI_BIN}
Service: all-smi.service

Next steps:
1. Register this exporter in AAMI Config Server
2. Add firewall rule if needed:
   sudo ufw allow ${ALL_SMI_PORT}/tcp
3. Verify metrics:
   curl http://localhost:${ALL_SMI_PORT}/metrics

Service commands:
  Status:  sudo systemctl status all-smi
  Stop:    sudo systemctl stop all-smi
  Start:   sudo systemctl start all-smi
  Restart: sudo systemctl restart all-smi
  Logs:    sudo journalctl -u all-smi -f

EOF
