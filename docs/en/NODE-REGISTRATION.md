# Node Registration Guide

## Table of Contents

1. [Overview](#overview)
2. [Registration Method Comparison](#registration-method-comparison)
3. [Prerequisites](#prerequisites)
4. [On-site Server Registration](#on-site-server-registration)
5. [Cloud Server Registration](#cloud-server-registration)
6. [Post-Registration Verification](#post-registration-verification)
7. [Troubleshooting](#troubleshooting)

## Overview

AAMI supports monitoring both on-site (on-premises) and cloud servers. While the registration methods differ in automation levels, both environments receive the same monitoring capabilities.

### Key Differences

| Aspect | On-site Servers | Cloud Servers |
|--------|----------------|---------------|
| **Registration** | Manual or semi-automated | Fully automated |
| **Initial Setup** | SSH login and script execution | Cloud-init / User Data |
| **Deployment Speed** | Individual server setup | Easy bulk deployment |
| **Network** | Existing infrastructure | VPC/Security Group configuration |
| **Use Cases** | GPU clusters, HPC, existing infrastructure | Dynamic scaling, Auto Scaling |

## Registration Method Comparison

### Method 1: Bootstrap Token (Recommended)

**Characteristic**: Server self-registers to Config Server

**Advantages**:
- ✅ Fully automated
- ✅ Automatic hardware detection
- ✅ Suitable for bulk deployment
- ✅ Minimizes human error

**Applicable to**:
- Cloud VMs (AWS, GCP, Azure, etc.)
- Newly provisioned on-site servers
- Automated deployment pipelines

### Method 2: Direct API Call (Manual)

**Characteristic**: Admin manually calls API

**Advantages**:
- ✅ Fine-grained control
- ✅ Suitable for existing servers
- ✅ Custom configuration possible

**Applicable to**:
- Already operational on-site servers
- Servers requiring special configuration
- Small-scale environments

## Prerequisites

### Config Server Setup

#### 1. Create Groups

Design group structure before registering nodes.

```bash
# Example: Environment-based groups
curl -X POST http://config-server:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production",
    "namespace": "environment",
    "description": "Production environment"
  }'

# Example: Function-based groups
curl -X POST http://config-server:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ml-training",
    "namespace": "logical",
    "description": "Machine learning training GPU cluster"
  }'
```

#### 2. Create Bootstrap Token (for auto-registration)

```bash
curl -X POST http://config-server:8080/api/v1/bootstrap-tokens \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ml-cluster-token",
    "default_group_id": "GROUP_ID",
    "max_uses": 100,
    "expires_at": "2024-12-31T23:59:59Z",
    "labels": {
      "environment": "production",
      "cluster": "ml-training"
    }
  }'

# Save the token from response
# Response: {"token": "aami_bootstrap_xxxxx..."}
```

### Network Requirements

**Config Server Access**:
- Port 8080 (HTTP API)
- Node → Config Server communication required

**Prometheus Scraping**:
- Port 9100 (Node Exporter)
- Port 9400 (DCGM Exporter, if GPU present)
- Prometheus → Node communication required

### Pre-flight Validation (Recommended)

You can validate system requirements before node registration:

```bash
# Download and run the script from AAMI repository
curl -fsSL https://raw.githubusercontent.com/fregataa/aami/main/scripts/preflight-check.sh -o preflight-check.sh
chmod +x preflight-check.sh

# Run in node mode with Config Server connectivity test
./preflight-check.sh --mode node --server http://config-server:8080
```

This script checks:
- System requirements (CPU, RAM, disk space)
- Software dependencies (curl, systemctl, tar)
- Config Server connectivity
- Port availability (9100, 9400)
- GPU detection (NVIDIA, AMD)

## On-site Server Registration

### Scenario 1: Bootstrap Script (Semi-automated)

#### Step 1: Connect to Server

```bash
ssh user@onsite-server-01
```

#### Step 2: Execute Bootstrap Script

```bash
# Prepare bootstrap token
BOOTSTRAP_TOKEN="aami_bootstrap_xxxxx..."
CONFIG_SERVER_URL="http://config-server.internal:8080"

# Run bootstrap
curl -fsSL ${CONFIG_SERVER_URL}/bootstrap.sh | \
  bash -s -- \
    --token ${BOOTSTRAP_TOKEN} \
    --server ${CONFIG_SERVER_URL}
```

#### Script Operation Process

1. **Collect System Information**
   - Hostname, IP address
   - CPU cores, memory capacity
   - GPU detection (nvidia-smi)
   - Network interfaces

2. **Install Exporters**
   - Node Exporter (system metrics)
   - DCGM Exporter (if GPU present)

3. **Register to Config Server**
   - Authenticate with bootstrap token
   - Send collected information
   - Auto-assign group (token's default_group_id)

4. **Setup Dynamic Checks**
   - Install dynamic-check.sh
   - Register cron job (1-minute interval)
   - Execute first check

#### Step 3: Verify Registration

```bash
# Check Node Exporter
curl http://localhost:9100/metrics

# Verify registration on Config Server
curl http://config-server:8080/api/v1/targets?hostname=$(hostname)
```

### Scenario 2: Manual API Registration

For servers with Node Exporter already installed:

```bash
curl -X POST http://config-server:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "onsite-gpu-01",
    "ip_address": "192.168.1.100",
    "primary_group_id": "GROUP_ID",
    "exporters": [
      {
        "type": "node_exporter",
        "port": 9100,
        "enabled": true
      },
      {
        "type": "dcgm_exporter",
        "port": 9400,
        "enabled": true
      }
    ],
    "labels": {
      "datacenter": "seoul",
      "rack": "r1",
      "gpu_model": "A100",
      "gpu_count": "8"
    }
  }'
```

## Cloud Server Registration

### AWS EC2 Example

#### Terraform Code

```hcl
# variables.tf
variable "config_server_url" {
  default = "http://config-server.internal:8080"
}

variable "bootstrap_token" {
  description = "AAMI Bootstrap Token"
  sensitive   = true
}

# main.tf
resource "aws_instance" "gpu_node" {
  ami           = "ami-xxxxx"  # Ubuntu 22.04 with GPU drivers
  instance_type = "p4d.24xlarge"

  vpc_security_group_ids = [aws_security_group.gpu_nodes.id]
  subnet_id              = aws_subnet.private.id

  user_data = templatefile("${path.module}/userdata.sh.tpl", {
    config_server_url = var.config_server_url
    bootstrap_token   = var.bootstrap_token
  })

  tags = {
    Name        = "ml-training-node-${count.index + 1}"
    Environment = "production"
    ManagedBy   = "terraform"
  }

  count = 10  # Create 10 GPU nodes
}

# security_group.tf
resource "aws_security_group" "gpu_nodes" {
  name        = "aami-gpu-nodes"
  description = "AAMI monitored GPU nodes"

  # Node Exporter (Prometheus → Node)
  ingress {
    from_port   = 9100
    to_port     = 9100
    protocol    = "tcp"
    cidr_blocks = [var.prometheus_cidr]
  }

  # DCGM Exporter (Prometheus → Node)
  ingress {
    from_port   = 9400
    to_port     = 9400
    protocol    = "tcp"
    cidr_blocks = [var.prometheus_cidr]
  }

  # Outbound to Config Server
  egress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = [var.config_server_cidr]
  }

  # Outbound to internet (for package installation)
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

#### User Data Script

```bash
# userdata.sh.tpl
#!/bin/bash
set -e

# Log file
exec > >(tee /var/log/aami-bootstrap.log)
exec 2>&1

echo "Starting AAMI bootstrap at $(date)"

# Config Server settings
CONFIG_SERVER_URL="${config_server_url}"
BOOTSTRAP_TOKEN="${bootstrap_token}"

# System update
apt-get update
apt-get install -y curl jq

# Run bootstrap
curl -fsSL $${CONFIG_SERVER_URL}/bootstrap.sh | \
  bash -s -- \
    --token $${BOOTSTRAP_TOKEN} \
    --server $${CONFIG_SERVER_URL}

echo "Bootstrap completed at $(date)"
```

### GCP Compute Engine Example

```hcl
resource "google_compute_instance" "gpu_node" {
  name         = "ml-training-node-${count.index + 1}"
  machine_type = "a2-highgpu-8g"  # A100 x8
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 200
    }
  }

  network_interface {
    network = "default"
    access_config {
      # Ephemeral public IP
    }
  }

  metadata_startup_script = templatefile("${path.module}/startup.sh.tpl", {
    config_server_url = var.config_server_url
    bootstrap_token   = var.bootstrap_token
  })

  service_account {
    scopes = ["compute-ro", "storage-ro"]
  }

  count = 10
}
```

### Azure VM Example

```hcl
resource "azurerm_linux_virtual_machine" "gpu_node" {
  name                = "ml-training-node-${count.index + 1}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  size                = "Standard_NC24ads_A100_v4"  # A100 x1

  admin_username = "azureuser"

  admin_ssh_key {
    username   = "azureuser"
    public_key = file("~/.ssh/id_rsa.pub")
  }

  network_interface_ids = [
    azurerm_network_interface.gpu_node[count.index].id,
  ]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }

  custom_data = base64encode(templatefile("${path.module}/cloud-init.yaml.tpl", {
    config_server_url = var.config_server_url
    bootstrap_token   = var.bootstrap_token
  }))

  count = 10
}
```

## Post-Registration Verification

### 1. Check Config Server

```bash
# List all registered targets
curl http://config-server:8080/api/v1/targets

# Check specific target status
curl http://config-server:8080/api/v1/targets/TARGET_ID
```

### 2. Check Prometheus

**Access Web UI**: http://prometheus:9090/targets

Verify:
- Target is "UP" status
- Last Scrape time is recent
- No error messages

### 3. Check Grafana

**Access Dashboard**: http://grafana:3000

Verify:
- Node appears in list
- Metrics are being collected (CPU, memory, disk)
- GPU metrics visible (for GPU nodes)

### 4. Direct Node Verification

```bash
# Node Exporter metrics
curl http://localhost:9100/metrics | grep node_

# DCGM Exporter metrics (if GPU present)
curl http://localhost:9400/metrics | grep DCGM_

# Dynamic check results
ls -la /var/lib/node_exporter/textfile/
cat /var/lib/node_exporter/textfile/check_mount.prom
```

## Troubleshooting

### Issue 1: Bootstrap Script Fails

**Symptom**: Error when executing curl command

**Causes**:
- Cannot reach Config Server
- Bootstrap token expired or max uses exceeded

**Solutions**:

```bash
# Test Config Server connectivity
curl -I http://config-server:8080/api/v1/health

# Check bootstrap token status (admin)
curl http://config-server:8080/api/v1/bootstrap-tokens/TOKEN_ID

# Create new token
curl -X POST http://config-server:8080/api/v1/bootstrap-tokens \
  -H "Content-Type: application/json" \
  -d '{ ... }'
```

### Issue 2: Target Not Appearing in Prometheus

**Symptom**: Registered in Config Server but not visible in Prometheus

**Causes**:
- Service Discovery file not updated
- Prometheus not reading SD file

**Solutions**:

```bash
# Check SD file
curl http://config-server:8080/api/v1/sd/prometheus

# Restart Prometheus
docker-compose restart prometheus

# Check Prometheus logs
docker-compose logs -f prometheus
```

### Issue 3: No Metrics from Node

**Symptom**: Target is UP but no metric values

**Causes**:
- Node Exporter not running
- Firewall blocking ports

**Solutions**:

```bash
# Check on node
systemctl status node_exporter
systemctl status dcgm-exporter  # GPU nodes

# Check and open firewall
sudo ufw status
sudo ufw allow 9100/tcp
sudo ufw allow 9400/tcp

# Test locally
curl http://localhost:9100/metrics
```

### Issue 4: Dynamic Checks Not Running

**Symptom**: Check metrics not appearing

**Causes**:
- Cron not running
- Script download failed

**Solutions**:

```bash
# Check cron job
crontab -l
cat /etc/cron.d/aami-dynamic-check

# Execute manually to see errors
/opt/aami/scripts/dynamic-check.sh

# Check logs
grep aami /var/log/syslog
```

## Bulk Deployment Guide

### Deployment Methods by Environment

| Environment | Tool | Admin Direct SSH |
|-------------|------|------------------|
| Small (1-10 nodes) | Direct SSH | Yes |
| Medium (10-100 nodes) | Ansible | No |
| Large (100+ nodes) | Ansible/Puppet/PXE | No |
| Cloud | Cloud-init/Terraform | No |

### Scenario: Deploy 100 GPU Nodes

#### 1. Preparation

```bash
# 1. Create group
curl -X POST http://config-server:8080/api/v1/groups \
  -d '{"name": "ml-cluster-batch-01"}'

# 2. Create bootstrap token (max_uses=100)
curl -X POST http://config-server:8080/api/v1/bootstrap-tokens \
  -d '{
    "name": "batch-01-token",
    "max_uses": 100,
    "expires_at": "2024-12-31T23:59:59Z"
  }'
```

#### 2a. Cloud Deployment (Terraform)

```bash
# Set Terraform variables
export TF_VAR_bootstrap_token="aami_bootstrap_xxxxx..."
export TF_VAR_node_count=100

# Deploy
terraform init
terraform plan
terraform apply
```

#### 2b. On-premises Deployment (Ansible)

For large on-premises environments, use configuration management tools instead of direct SSH access.

**Inventory file (inventory.ini)**:
```ini
[gpu_nodes]
gpu-node-[001:100].example.com
```

**Playbook (aami-bootstrap.yml)**:
```yaml
- hosts: gpu_nodes
  become: yes
  vars:
    bootstrap_token: "aami_bootstrap_xxxxx..."
    config_server: "http://config-server:8080"
  tasks:
    - name: Run AAMI bootstrap
      shell: |
        curl -fsSL {{ config_server }}/bootstrap.sh | \
          bash -s -- --token {{ bootstrap_token }} --server {{ config_server }}
      args:
        creates: /etc/systemd/system/node_exporter.service
```

**Execute**:
```bash
# Deploy to all 100 nodes in parallel
ansible-playbook -i inventory.ini aami-bootstrap.yml -f 50
```

#### 2c. On-premises Deployment (PXE/Kickstart)

For new server provisioning, include bootstrap in the post-installation script:

**Kickstart snippet**:
```bash
%post
# AAMI Bootstrap
curl -fsSL http://config-server:8080/bootstrap.sh | \
  bash -s -- --token aami_bootstrap_xxxxx --server http://config-server:8080
%end
```

#### 3. Monitor Deployment

```bash
# Check registered node count
watch -n 5 'curl -s http://config-server:8080/api/v1/targets | jq length'

# Check Prometheus target count
curl http://prometheus:9090/api/v1/targets | jq '.data.activeTargets | length'
```

#### 4. Verify Deployment

```bash
# Check if all nodes are UP
curl http://prometheus:9090/api/v1/targets | \
  jq '.data.activeTargets[] | select(.health != "up") | .labels.instance'

# Verify GPU metrics collection
curl -G http://prometheus:9090/api/v1/query \
  --data-urlencode 'query=count(DCGM_FI_DEV_GPU_TEMP)' | \
  jq '.data.result[0].value[1]'
```

## Best Practices

### Bootstrap Token Management

**DO**:
- ✅ Create separate tokens per purpose (dev/staging/prod)
- ✅ Set expiration dates
- ✅ Set usage limits
- ✅ Deactivate after use

**DON'T**:
- ❌ Reuse one token across all environments
- ❌ Create without expiration
- ❌ Commit tokens to public repositories
- ❌ Send tokens via Slack or email

### Group Design

**Recommended Structure**:

```
environment (namespace)
├── production
│   ├── critical
│   └── standard
├── staging
└── development

infrastructure (namespace)
├── datacenter-seoul
│   ├── zone-a
│   └── zone-b
└── datacenter-tokyo

logical (namespace)
├── ml-training
├── ml-inference
├── api-servers
└── databases
```

### Label Strategy

**Useful Labels**:
- `environment`: production, staging, development
- `cluster`: Cluster name
- `datacenter`: Datacenter location
- `rack`: Rack number
- `gpu_model`: GPU model name
- `gpu_count`: GPU count
- `owner`: Owning team
- `cost_center`: Cost center

## Next Steps

After node registration:

1. **Configure Alert Rules**: See [Alert Rules Guide](./ALERT-RULES.md)
2. **Create Dashboards**: See [Dashboard Guide](./DASHBOARDS.md)
3. **Add Dynamic Checks**: See [Check Script Management](./CHECK-SCRIPT-MANAGEMENT.md)
4. **Capacity Planning**: Configure metric retention policies

## References

- [Quick Start Guide](./QUICKSTART.md)
- [API Documentation](./API.md)
- [Check Script Management](./CHECK-SCRIPT-MANAGEMENT.md)
- [Troubleshooting Guide](./TROUBLESHOOTING.md)
