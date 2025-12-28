#!/usr/bin/env bash
set -euo pipefail

# Install Prometheus Node Exporter
#
# This script downloads and installs Node Exporter as a systemd service
# on Linux systems.
#
# Usage: ./install-node-exporter.sh [OPTIONS]
#
# Options:
#   -v, --version VERSION    Node Exporter version (default: 1.6.1)
#   -p, --port PORT          Listen port (default: 9100)
#   -h, --help              Show this help message

# Configuration
NODE_EXPORTER_VERSION="${NODE_EXPORTER_VERSION:-1.6.1}"
NODE_EXPORTER_PORT="${NODE_EXPORTER_PORT:-9100}"
INSTALL_DIR="/usr/local/bin"
SERVICE_USER="node_exporter"
SERVICE_FILE="/etc/systemd/system/node_exporter.service"
TEXTFILE_DIR="/var/lib/node_exporter/textfile_collector"

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

Install Prometheus Node Exporter as a systemd service.

Options:
    -v, --version VERSION    Node Exporter version (default: 1.6.1)
    -p, --port PORT          Listen port (default: 9100)
    -h, --help              Show this help message

Environment Variables:
    NODE_EXPORTER_VERSION   - Version to install
    NODE_EXPORTER_PORT      - Port to listen on

Example:
    $(basename "$0")
    $(basename "$0") --version 1.7.0 --port 9100

EOF
    exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            NODE_EXPORTER_VERSION="$2"
            shift 2
            ;;
        -p|--port)
            NODE_EXPORTER_PORT="$2"
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
    armv7l)
        ARCH="armv7"
        ;;
    *)
        error "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

info "Installing Node Exporter v${NODE_EXPORTER_VERSION} for ${ARCH}"

# Download Node Exporter
DOWNLOAD_URL="https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/node_exporter-${NODE_EXPORTER_VERSION}.linux-${ARCH}.tar.gz"
TEMP_DIR=$(mktemp -d)

info "Downloading from: $DOWNLOAD_URL"
if ! curl -L -o "${TEMP_DIR}/node_exporter.tar.gz" "$DOWNLOAD_URL"; then
    error "Failed to download Node Exporter"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Extract archive
info "Extracting archive..."
tar xzf "${TEMP_DIR}/node_exporter.tar.gz" -C "$TEMP_DIR"

# Install binary
info "Installing binary to $INSTALL_DIR..."
cp "${TEMP_DIR}/node_exporter-${NODE_EXPORTER_VERSION}.linux-${ARCH}/node_exporter" "$INSTALL_DIR/"
chmod +x "${INSTALL_DIR}/node_exporter"

# Create service user
if ! id "$SERVICE_USER" &>/dev/null; then
    info "Creating service user: $SERVICE_USER"
    useradd --no-create-home --shell /bin/false "$SERVICE_USER"
fi

# Create textfile collector directory
info "Creating textfile collector directory: $TEXTFILE_DIR"
mkdir -p "$TEXTFILE_DIR"
chown "$SERVICE_USER:$SERVICE_USER" "$TEXTFILE_DIR"
chmod 755 "$TEXTFILE_DIR"

# Create systemd service file
info "Creating systemd service..."
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Prometheus Node Exporter
Wants=network-online.target
After=network-online.target

[Service]
User=$SERVICE_USER
Group=$SERVICE_USER
Type=simple
ExecStart=$INSTALL_DIR/node_exporter \\
    --web.listen-address=:$NODE_EXPORTER_PORT \\
    --collector.textfile.directory=$TEXTFILE_DIR \\
    --collector.filesystem.mount-points-exclude=^/(dev|proc|sys|var/lib/docker/.+|var/lib/kubelet/.+)(\$|/) \\
    --collector.filesystem.fs-types-exclude=^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)\$

Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and start service
info "Reloading systemd daemon..."
systemctl daemon-reload

info "Enabling Node Exporter service..."
systemctl enable node_exporter

info "Starting Node Exporter service..."
systemctl start node_exporter

# Cleanup
rm -rf "$TEMP_DIR"

# Verify installation
sleep 2
if systemctl is-active --quiet node_exporter; then
    info "Node Exporter installed and running successfully!"
    info "Metrics available at: http://localhost:${NODE_EXPORTER_PORT}/metrics"

    # Test metrics endpoint
    if curl -s "http://localhost:${NODE_EXPORTER_PORT}/metrics" > /dev/null; then
        info "Metrics endpoint is responding"
    else
        warn "Metrics endpoint is not responding. Check the service status."
    fi
else
    error "Node Exporter service failed to start"
    error "Check logs with: journalctl -u node_exporter -f"
    exit 1
fi

# Print next steps
cat <<EOF

${GREEN}Installation complete!${NC}

Textfile collector enabled at: $TEXTFILE_DIR

Next steps:
1. Register this node in AAMI Config Server
2. Install dynamic check scripts from AAMI repository
3. Add firewall rule if needed:
   sudo ufw allow ${NODE_EXPORTER_PORT}/tcp
4. Verify metrics:
   curl http://localhost:${NODE_EXPORTER_PORT}/metrics

Service commands:
  Status:  sudo systemctl status node_exporter
  Stop:    sudo systemctl stop node_exporter
  Start:   sudo systemctl start node_exporter
  Restart: sudo systemctl restart node_exporter
  Logs:    sudo journalctl -u node_exporter -f

EOF
