#!/bin/bash
#
# AWS EC2 User Data Script for AAMI Bootstrap
#
# This script is executed during EC2 instance initialization.
# It bootstraps the node with AAMI monitoring components.
#
# Usage in Terraform:
#   user_data = templatefile("${path.module}/aws-ec2-userdata.sh", {
#     bootstrap_token    = var.aami_bootstrap_token
#     config_server_url  = var.aami_config_server_url
#     primary_group      = "infrastructure:aws/us-east-1"
#   })
#
# Usage in AWS Console:
#   Copy this script and replace the variables ${...} with actual values

set -euo pipefail

# Configuration (these will be replaced by Terraform templatefile or manually)
BOOTSTRAP_TOKEN="${bootstrap_token}"
CONFIG_SERVER_URL="${config_server_url}"
PRIMARY_GROUP="${primary_group}"

# Metadata
INSTANCE_ID=$(ec2-metadata --instance-id | cut -d ' ' -f 2)
INSTANCE_TYPE=$(ec2-metadata --instance-type | cut -d ' ' -f 2)
AVAILABILITY_ZONE=$(ec2-metadata --availability-zone | cut -d ' ' -f 2)
PUBLIC_HOSTNAME=$(ec2-metadata --public-hostname | cut -d ' ' -f 2)
LOCAL_IPV4=$(ec2-metadata --local-ipv4 | cut -d ' ' -f 2)

# Logging
exec > >(tee /var/log/aami-bootstrap.log)
exec 2>&1

echo "=================================================="
echo "AAMI Bootstrap Script"
echo "=================================================="
echo "Instance ID: $INSTANCE_ID"
echo "Instance Type: $INSTANCE_TYPE"
echo "Availability Zone: $AVAILABILITY_ZONE"
echo "Hostname: $PUBLIC_HOSTNAME"
echo "Local IP: $LOCAL_IPV4"
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
    --hostname "$PUBLIC_HOSTNAME" \
    --metadata instance_id="$INSTANCE_ID" \
    --metadata instance_type="$INSTANCE_TYPE" \
    --metadata availability_zone="$AVAILABILITY_ZONE" \
    --metadata local_ipv4="$LOCAL_IPV4"

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
AAMI_HOSTNAME="$PUBLIC_HOSTNAME"
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
echo "Node Exporter: http://$LOCAL_IPV4:9100/metrics"
echo "Logs: /var/log/aami/dynamic-check.log"
echo "Status: systemctl status aami-dynamic-check.timer"
echo "=================================================="
