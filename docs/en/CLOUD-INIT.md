# Cloud-Init Integration Guide

This guide explains how to integrate AAMI monitoring with cloud instance initialization using cloud-init across different cloud providers.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Bootstrap Process](#bootstrap-process)
4. [Cloud Provider Integration](#cloud-provider-integration)
   - [AWS EC2](#aws-ec2)
   - [Google Cloud Compute Engine](#google-cloud-compute-engine)
   - [Azure Virtual Machines](#azure-virtual-machines)
5. [Terraform Integration](#terraform-integration)
6. [Customization](#customization)
7. [Troubleshooting](#troubleshooting)
8. [Best Practices](#best-practices)

## Overview

Cloud-init enables automatic configuration of cloud instances during first boot. AAMI provides cloud-init scripts that:

- **Register the instance** in AAMI Config Server
- **Install monitoring agents** (Node Exporter, DCGM Exporter)
- **Deploy dynamic checks** for hardware health monitoring
- **Configure automatic updates** via systemd timers

**Benefits**:
- Zero-touch monitoring deployment
- Consistent configuration across environments
- Automatic enrollment in group-based policies
- Infrastructure as Code compatible

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Cloud Provider Console / Terraform / CLI                   │
│  - Launches VM with cloud-init script                       │
│  - Passes bootstrap token and config                        │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ VM Boot
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  VM Instance - Cloud-Init Execution                         │
│  1. Update system packages                                  │
│  2. Install dependencies (jq, curl, smartmontools, etc.)    │
│  3. Call AAMI bootstrap API                                 │
│     POST /api/v1/bootstrap/register                         │
│  4. Install Node Exporter with textfile collector           │
│  5. Install dynamic-check.sh script                         │
│  6. Configure systemd timer (1 minute interval)             │
│  7. Verify installation                                     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ Register
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  AAMI Config Server                                         │
│  - Creates target record                                    │
│  - Assigns to primary group + secondary groups              │
│  - Returns effective checks configuration                   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ Poll (every 1 min)
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  dynamic-check.sh                                           │
│  - Fetch effective checks                                   │
│  - Execute check scripts                                    │
│  - Write metrics to textfile collector                      │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ Scrape
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  Prometheus                                                 │
│  - Collects all metrics (system + custom)                   │
└─────────────────────────────────────────────────────────────┘
```

## Bootstrap Process

### 1. Bootstrap Token Creation

First, create a bootstrap token in AAMI Config Server:

```bash
curl -X POST http://config-server:8080/api/v1/bootstrap/tokens \
  -H "Content-Type: application/json" \
  -d '{
    "token": "my-secure-bootstrap-token-12345",
    "description": "Token for AWS production GPU nodes",
    "default_group_id": 123,
    "expires_at": "2024-12-31T23:59:59Z",
    "max_uses": 100
  }'
```

**Token Properties**:
- **Token**: Secret string used for authentication
- **Default Group**: Primary group for registered nodes
- **Expiration**: Token validity period
- **Max Uses**: Limit number of registrations

### 2. Instance Bootstrap

When an instance boots with cloud-init, it:

1. **Collects metadata** from cloud provider (instance ID, type, region, etc.)
2. **Calls bootstrap API** with token and metadata
3. **Receives configuration** (node ID, assigned groups, checks)
4. **Installs components** (Node Exporter, check scripts)
5. **Starts monitoring** (systemd timer begins check execution)

### 3. Verification

After bootstrap completes:

```bash
# SSH to instance
ssh ubuntu@<instance-ip>

# Check bootstrap log
sudo cat /var/log/aami-bootstrap.log

# Verify Node Exporter
systemctl status node_exporter
curl http://localhost:9100/metrics

# Verify dynamic checks
systemctl status aami-dynamic-check.timer
sudo tail -f /var/log/aami/dynamic-check.log

# Check metrics
curl http://localhost:9100/metrics | grep -E 'aami_|mount_check|disk_smart'
```

## Cloud Provider Integration

### AWS EC2

#### User Data Script

AWS EC2 uses **user data** for cloud-init. The script runs as root during first boot.

**Template**: `examples/cloud-init/aws-ec2-userdata.sh`

#### Key Features
- Fetches EC2 metadata (instance ID, type, AZ, IPs)
- Uses `ec2-metadata` command for metadata access
- Supports EBS volume attachments
- Configures UFW firewall rules

#### Manual Deployment

1. **Create instance with AWS Console**:
   - Launch instance wizard
   - Advanced Details → User data
   - Copy `aws-ec2-userdata.sh` content
   - Replace `${bootstrap_token}`, `${config_server_url}`, `${primary_group}`

2. **Using AWS CLI**:

```bash
aws ec2 run-instances \
  --image-id ami-0c7217cdde317cfec \
  --instance-type p4d.24xlarge \
  --key-name my-ssh-key \
  --subnet-id subnet-xxxxx \
  --security-group-ids sg-xxxxx \
  --user-data file://aws-ec2-userdata.sh \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=aami-gpu-node}]'
```

#### Terraform Integration

See [examples/terraform/aws-gpu-instance.tf](../examples/terraform/aws-gpu-instance.tf)

```hcl
resource "aws_instance" "gpu_node" {
  ami           = "ami-0c7217cdde317cfec"
  instance_type = "p4d.24xlarge"

  user_data = templatefile("${path.module}/../cloud-init/aws-ec2-userdata.sh", {
    bootstrap_token    = var.aami_bootstrap_token
    config_server_url  = var.aami_config_server_url
    primary_group      = "infrastructure:aws/us-east-1"
  })
}
```

### Google Cloud Compute Engine

#### Startup Script

GCP uses **metadata startup-script** for initialization.

**Template**: `examples/cloud-init/gcp-startup-script.sh`

#### Key Features
- Accesses GCP metadata API (`http://metadata.google.internal`)
- Retrieves project ID, zone, machine type
- Configures firewall rules via gcloud
- Supports both internal and external IPs

#### Manual Deployment

1. **Create instance with GCP Console**:
   - Create instance
   - Management → Automation → Startup script
   - Copy `gcp-startup-script.sh` content
   - Replace variables

2. **Using gcloud CLI**:

```bash
gcloud compute instances create aami-gpu-node-0 \
  --zone=us-central1-a \
  --machine-type=a2-highgpu-8g \
  --image-family=ubuntu-2204-lts \
  --image-project=ubuntu-os-cloud \
  --boot-disk-size=100GB \
  --boot-disk-type=pd-ssd \
  --metadata-from-file startup-script=gcp-startup-script.sh \
  --tags=aami-gpu-node
```

#### Terraform Integration

See [examples/terraform/gcp-gpu-instance.tf](../examples/terraform/gcp-gpu-instance.tf)

```hcl
resource "google_compute_instance" "gpu_node" {
  name         = "aami-gpu-node-0"
  machine_type = "a2-highgpu-8g"

  metadata_startup_script = templatefile("${path.module}/../cloud-init/gcp-startup-script.sh", {
    bootstrap_token    = var.aami_bootstrap_token
    config_server_url  = var.aami_config_server_url
    primary_group      = "infrastructure:gcp/us-central1"
  })
}
```

### Azure Virtual Machines

#### Custom Data

Azure uses **custom data** for cloud-init configuration.

**Template**: `examples/cloud-init/azure-custom-data.yaml`

#### Key Features
- Uses cloud-config YAML format
- Accesses Azure Instance Metadata Service (IMDS)
- Retrieves VM ID, resource group, subscription
- Supports cloud-init directives (packages, write_files, runcmd)

#### Manual Deployment

1. **Create VM with Azure Portal**:
   - Create virtual machine
   - Advanced → Custom data
   - Copy `azure-custom-data.yaml` content
   - Replace variables

2. **Using Azure CLI**:

```bash
az vm create \
  --resource-group my-resource-group \
  --name aami-gpu-node-0 \
  --location eastus \
  --size Standard_NC24ads_A100_v4 \
  --image Canonical:0001-com-ubuntu-server-jammy:22_04-lts-gen2:latest \
  --custom-data azure-custom-data.yaml \
  --admin-username azureuser \
  --ssh-key-values @~/.ssh/id_rsa.pub
```

#### Terraform Integration

See [examples/terraform/azure-gpu-instance.tf](../examples/terraform/azure-gpu-instance.tf)

```hcl
resource "azurerm_linux_virtual_machine" "gpu_node" {
  name                = "aami-gpu-node-0"
  size                = "Standard_NC24ads_A100_v4"

  custom_data = base64encode(templatefile("${path.module}/../cloud-init/azure-custom-data.yaml", {
    bootstrap_token    = var.aami_bootstrap_token
    config_server_url  = var.aami_config_server_url
    primary_group      = "infrastructure:azure/eastus"
  }))
}
```

## Terraform Integration

### Complete Workflow

1. **Create Terraform variables**:

```hcl
# variables.tf
variable "aami_config_server_url" {
  description = "AAMI Config Server URL"
  type        = string
}

variable "aami_bootstrap_token" {
  description = "Bootstrap token for automatic registration"
  type        = string
  sensitive   = true
}

variable "aami_primary_group" {
  description = "Primary group for monitoring"
  type        = string
  default     = "infrastructure:cloud/region"
}
```

2. **Use in instance definition**:

```hcl
# main.tf
resource "aws_instance" "gpu" {
  # ... instance configuration ...

  user_data = templatefile("${path.module}/cloud-init.sh", {
    bootstrap_token    = var.aami_bootstrap_token
    config_server_url  = var.aami_config_server_url
    primary_group      = var.aami_primary_group
  })
}
```

3. **Deploy**:

```bash
terraform init
terraform plan
terraform apply
```

### Multi-Cloud Deployment

Deploy identical monitoring across clouds:

```hcl
# main.tf
module "aws_gpu" {
  source = "./modules/aws-gpu"

  aami_config_server_url = var.aami_config_server_url
  aami_bootstrap_token   = var.aami_bootstrap_token
  aami_primary_group     = "infrastructure:aws/us-east-1"
}

module "gcp_gpu" {
  source = "./modules/gcp-gpu"

  aami_config_server_url = var.aami_config_server_url
  aami_bootstrap_token   = var.aami_bootstrap_token
  aami_primary_group     = "infrastructure:gcp/us-central1"
}

module "azure_gpu" {
  source = "./modules/azure-gpu"

  aami_config_server_url = var.aami_config_server_url
  aami_bootstrap_token   = var.aami_bootstrap_token
  aami_primary_group     = "infrastructure:azure/eastus"
}
```

## Customization

### Adding Custom Initialization

Extend cloud-init scripts with custom steps:

```bash
#!/bin/bash
# ... existing AAMI bootstrap code ...

# Custom initialization
echo "[CUSTOM] Installing additional packages..."
apt-get install -y \
    nvidia-driver-535 \
    cuda-toolkit-12-2

# Configure GPU settings
echo "[CUSTOM] Configuring GPU persistence mode..."
nvidia-smi -pm 1

# Mount additional storage
echo "[CUSTOM] Mounting NFS share..."
mkdir -p /mnt/shared
mount -t nfs nfs-server.example.com:/export/shared /mnt/shared

# Install custom monitoring checks
echo "[CUSTOM] Installing custom checks..."
curl -o /usr/local/lib/aami/checks/my-custom-check.sh \
    https://my-repo.com/checks/my-custom-check.sh
chmod +x /usr/local/lib/aami/checks/my-custom-check.sh
```

### Environment-Specific Configuration

Use different groups based on environment:

```hcl
# Terraform
locals {
  environment = terraform.workspace

  aami_group_mapping = {
    production  = "infrastructure:aws/prod/us-east-1"
    staging     = "infrastructure:aws/staging/us-east-1"
    development = "infrastructure:aws/dev/us-east-1"
  }
}

resource "aws_instance" "node" {
  user_data = templatefile("${path.module}/cloud-init.sh", {
    primary_group = local.aami_group_mapping[local.environment]
  })
}
```

### Conditional Check Installation

Install checks only for specific instance types:

```bash
#!/bin/bash
INSTANCE_TYPE=$(ec2-metadata --instance-type | cut -d ' ' -f 2)

# Install GPU checks only on GPU instances
if [[ "$INSTANCE_TYPE" =~ ^(p3|p4|p5|g4|g5)\. ]]; then
    echo "Installing DCGM Exporter for GPU monitoring..."
    curl -L https://github.com/NVIDIA/dcgm-exporter/releases/download/3.1.7-3.1.4/dcgm-exporter_3.1.7-3.1.4_amd64.deb \
        -o /tmp/dcgm-exporter.deb
    dpkg -i /tmp/dcgm-exporter.deb
fi

# Install InfiniBand checks only on instances with IB
if [ -d "/sys/class/infiniband" ]; then
    echo "Installing InfiniBand monitoring..."
    apt-get install -y infiniband-diags
fi
```

## Troubleshooting

### Cloud-Init Not Running

```bash
# Check cloud-init status
cloud-init status

# View cloud-init logs
sudo cat /var/log/cloud-init.log
sudo cat /var/log/cloud-init-output.log

# Re-run cloud-init (testing only)
sudo cloud-init clean --logs
sudo cloud-init init
sudo cloud-init modules --mode=config
sudo cloud-init modules --mode=final
```

### Bootstrap API Failure

```bash
# Check bootstrap log
sudo cat /var/log/aami-bootstrap.log

# Test Config Server connectivity
curl -v http://config-server:8080/api/v1/health

# Test bootstrap API manually
curl -X POST http://config-server:8080/api/v1/bootstrap/register \
  -H "Content-Type: application/json" \
  -d '{
    "token": "your-bootstrap-token",
    "hostname": "test-node",
    "primary_group": "infrastructure:test"
  }'
```

### Component Installation Failures

```bash
# Check for errors in installation
sudo journalctl -xe

# Test individual components
sudo systemctl status node_exporter
sudo systemctl status aami-dynamic-check.timer

# Manually install components
sudo /path/to/install-node-exporter.sh
sudo /usr/local/bin/dynamic-check.sh --debug
```

### Network Issues

```bash
# Test DNS resolution
nslookup config-server.example.com

# Test HTTP connectivity
curl -v http://config-server:8080/api/v1/health

# Check security group/firewall
# AWS
aws ec2 describe-security-groups --group-ids sg-xxxxx

# GCP
gcloud compute firewall-rules list --filter="name~aami"

# Azure
az network nsg show --resource-group xxx --name xxx
```

## Best Practices

### Security

1. **Protect Bootstrap Tokens**:
   - Store in secure secret management (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault)
   - Use short expiration periods
   - Rotate regularly
   - Limit max uses per token

2. **Use Private Networks**:
   - Deploy Config Server in private subnet
   - Use VPC peering or VPN for connectivity
   - Restrict security groups to internal traffic only

3. **Enable Encryption**:
   - Use HTTPS for Config Server
   - Encrypt instance disks
   - Use secure boot when available

### Idempotency

Ensure scripts can run multiple times safely:

```bash
# Check if already installed
if [ -f "/usr/local/bin/node_exporter" ]; then
    echo "Node Exporter already installed, skipping..."
    exit 0
fi

# Use conditional checks
if ! systemctl is-active --quiet node_exporter; then
    echo "Starting Node Exporter..."
    systemctl start node_exporter
fi
```

### Error Handling

Add robust error handling:

```bash
set -euo pipefail  # Exit on error, undefined variables, pipe failures

# Trap errors
trap 'echo "Error on line $LINENO"' ERR

# Retry logic for network operations
retry_count=0
max_retries=5
while [ $retry_count -lt $max_retries ]; do
    if curl -f http://config-server:8080/api/v1/health; then
        break
    fi
    retry_count=$((retry_count + 1))
    sleep 10
done
```

### Logging

Comprehensive logging for debugging:

```bash
# Log to file and stdout
exec > >(tee /var/log/aami-bootstrap.log)
exec 2>&1

# Timestamp all output
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

log "Starting AAMI bootstrap..."
```

### Testing

Test cloud-init scripts before production:

```bash
# Local testing with Docker
docker run -it ubuntu:22.04 bash
# Paste and run script manually

# Test with Vagrant
vagrant up
vagrant ssh
sudo cat /var/log/aami-bootstrap.log

# Test cloud-init syntax
cloud-init schema --config-file azure-custom-data.yaml
```

## References

- [Cloud-Init Documentation](https://cloudinit.readthedocs.io/)
- [AWS EC2 User Data](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/user-data.html)
- [GCP Startup Scripts](https://cloud.google.com/compute/docs/instances/startup-scripts)
- [Azure Custom Data](https://docs.microsoft.com/en-us/azure/virtual-machines/custom-data)
- [AAMI Bootstrap API](/docs/en/API.md#bootstrap-api)
