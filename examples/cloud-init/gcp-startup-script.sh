#!/bin/bash
#
# GCP Compute Engine Startup Script for AAMI Bootstrap
#
# This script is executed during GCP VM initialization.
# It bootstraps the node with AAMI monitoring components.
#
# Usage in Terraform:
#   metadata_startup_script = templatefile("${path.module}/gcp-startup-script.sh", {
#     bootstrap_token    = var.aami_bootstrap_token
#     config_server_url  = var.aami_config_server_url
#     primary_group      = "infrastructure:gcp/us-central1"
#   })
#
# Usage in GCP Console:
#   Copy this script to Metadata > startup-script

set -euo pipefail

# Configuration (these will be replaced by Terraform templatefile or manually)
BOOTSTRAP_TOKEN="${bootstrap_token}"
CONFIG_SERVER_URL="${config_server_url}"
PRIMARY_GROUP="${primary_group}"

# GCP Metadata
METADATA_SERVER="http://metadata.google.internal/computeMetadata/v1"
METADATA_HEADER="Metadata-Flavor: Google"

get_metadata() {
    curl -sf -H "$METADATA_HEADER" "$METADATA_SERVER/$1"
}

INSTANCE_ID=$(get_metadata "instance/id")
INSTANCE_NAME=$(get_metadata "instance/name")
MACHINE_TYPE=$(get_metadata "instance/machine-type" | awk -F'/' '{print $NF}')
ZONE=$(get_metadata "instance/zone" | awk -F'/' '{print $NF}')
PROJECT_ID=$(get_metadata "project/project-id")
INTERNAL_IP=$(get_metadata "instance/network-interfaces/0/ip")
EXTERNAL_IP=$(get_metadata "instance/network-interfaces/0/access-configs/0/external-ip" || echo "none")

# Logging
exec > >(tee /var/log/aami-bootstrap.log)
exec 2>&1

echo "=================================================="
echo "AAMI Bootstrap Script (GCP)"
echo "=================================================="
echo "Instance ID: $INSTANCE_ID"
echo "Instance Name: $INSTANCE_NAME"
echo "Machine Type: $MACHINE_TYPE"
echo "Zone: $ZONE"
echo "Project: $PROJECT_ID"
echo "Internal IP: $INTERNAL_IP"
echo "External IP: $EXTERNAL_IP"
echo "Config Server: $CONFIG_SERVER_URL"
echo "Primary Group: $PRIMARY_GROUP"
echo "=================================================="

# Update system
echo "[1/6] Updating system packages..."
apt-get update -y
apt-get upgrade -y

# Install dependencies
echo "[2/6] Installing dependencies..."
apt-get install -y \
    curl \
    jq \
    wget \
    ca-certificates \
    gnupg \
    lsb-release \
    smartmontools \
    nvme-cli \
    bc

# Download and run AAMI bootstrap script
echo "[3/6] Running AAMI bootstrap..."
curl -fsSL "$${CONFIG_SERVER_URL}/bootstrap.sh" | \
    bash -s -- \
    --token "$BOOTSTRAP_TOKEN" \
    --group "$PRIMARY_GROUP" \
    --hostname "$INSTANCE_NAME" \
    --metadata instance_id="$INSTANCE_ID" \
    --metadata machine_type="$MACHINE_TYPE" \
    --metadata zone="$ZONE" \
    --metadata project_id="$PROJECT_ID" \
    --metadata internal_ip="$INTERNAL_IP" \
    --metadata external_ip="$EXTERNAL_IP"

# Install Node Exporter with textfile collector
echo "[4/6] Installing Node Exporter..."
curl -fsSL https://raw.githubusercontent.com/fregataa/aami/main/scripts/node/install-node-exporter.sh | bash

# Install dynamic check script
echo "[5/6] Installing dynamic check script..."
curl -fsSL https://raw.githubusercontent.com/fregataa/aami/main/scripts/node/dynamic-check.sh \
    -o /usr/local/bin/dynamic-check.sh
chmod +x /usr/local/bin/dynamic-check.sh

# Configure AAMI
mkdir -p /etc/aami
cat > /etc/aami/config <<EOF
AAMI_CONFIG_SERVER_URL="$CONFIG_SERVER_URL"
AAMI_HOSTNAME="$INSTANCE_NAME"
EOF

# Set up systemd timer
curl -fsSL https://raw.githubusercontent.com/fregataa/aami/main/config/node-exporter/systemd/aami-dynamic-check.service \
    -o /etc/systemd/system/aami-dynamic-check.service
curl -fsSL https://raw.githubusercontent.com/fregataa/aami/main/config/node-exporter/systemd/aami-dynamic-check.timer \
    -o /etc/systemd/system/aami-dynamic-check.timer

# Update Config Server URL in service file
sed -i "s|http://config-server:8080|$CONFIG_SERVER_URL|g" /etc/systemd/system/aami-dynamic-check.service

systemctl daemon-reload
systemctl enable aami-dynamic-check.timer
systemctl start aami-dynamic-check.timer

# Configure firewall (GCP-specific)
echo "Configuring firewall..."
# Allow Node Exporter metrics from internal network
gcloud compute firewall-rules create allow-node-exporter-internal \
    --project="$PROJECT_ID" \
    --allow=tcp:9100 \
    --source-ranges=10.0.0.0/8 \
    --description="Allow Node Exporter metrics from internal network" \
    2>/dev/null || echo "Firewall rule already exists or cannot be created"

# Verify installation
echo "[6/6] Verifying installation..."
sleep 5

# Check Node Exporter
if systemctl is-active --quiet node_exporter; then
    echo "✓ Node Exporter is running"
else
    echo "✗ Node Exporter is not running"
fi

# Check dynamic check timer
if systemctl is-active --quiet aami-dynamic-check.timer; then
    echo "✓ Dynamic check timer is running"
else
    echo "✗ Dynamic check timer is not running"
fi

# Test metrics endpoint
if curl -sf http://localhost:9100/metrics > /dev/null; then
    echo "✓ Metrics endpoint is responding"
else
    echo "✗ Metrics endpoint is not responding"
fi

echo "=================================================="
echo "AAMI Bootstrap completed successfully!"
echo "=================================================="
echo "Node Exporter: http://$INTERNAL_IP:9100/metrics"
echo "Logs: /var/log/aami/dynamic-check.log"
echo "Status: systemctl status aami-dynamic-check.timer"
echo "=================================================="
