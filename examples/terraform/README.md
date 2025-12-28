# Terraform Examples for AAMI

This directory contains Terraform configurations for deploying GPU instances with AAMI monitoring across different cloud providers.

## Available Examples

- **AWS**: `aws-gpu-instance.tf` - GPU instances on EC2 (p4d.24xlarge)
- **GCP**: `gcp-gpu-instance.tf` - GPU instances on Compute Engine (a2-highgpu-8g)
- **Azure**: `azure-gpu-instance.tf` - GPU instances on Azure (Standard_NC24ads_A100_v4)

## Prerequisites

### Common Requirements

1. **AAMI Config Server**: Deploy and configure the Config Server
2. **Bootstrap Token**: Create a bootstrap token in Config Server
3. **Network Setup**: VPC/VNet and subnets configured
4. **Terraform**: Version 1.0 or higher

### Cloud-Specific Requirements

#### AWS
```bash
# Install AWS CLI
aws configure

# Verify credentials
aws sts get-caller-identity
```

#### GCP
```bash
# Install gcloud CLI
gcloud auth application-default login

# Set project
gcloud config set project YOUR_PROJECT_ID
```

#### Azure
```bash
# Install Azure CLI
az login

# Set subscription
az account set --subscription YOUR_SUBSCRIPTION_ID
```

## Usage

### 1. Create Variables File

Create a `terraform.tfvars` file:

#### AWS Example
```hcl
# terraform.tfvars
aws_region                = "us-east-1"
instance_count            = 2
instance_type             = "p4d.24xlarge"
vpc_id                    = "vpc-xxxxx"
subnet_id                 = "subnet-xxxxx"
key_name                  = "my-ssh-key"
aami_config_server_url    = "http://config.example.com:8080"
aami_bootstrap_token      = "your-bootstrap-token"
aami_primary_group        = "infrastructure:aws/us-east-1"
```

#### GCP Example
```hcl
# terraform.tfvars
gcp_project               = "my-project-id"
gcp_region                = "us-central1"
gcp_zone                  = "us-central1-a"
instance_count            = 2
machine_type              = "a2-highgpu-8g"
network                   = "default"
aami_config_server_url    = "http://config.example.com:8080"
aami_bootstrap_token      = "your-bootstrap-token"
aami_primary_group        = "infrastructure:gcp/us-central1"
```

#### Azure Example
```hcl
# terraform.tfvars
resource_group_name       = "my-resource-group"
location                  = "eastus"
instance_count            = 2
vm_size                   = "Standard_NC24ads_A100_v4"
vnet_name                 = "my-vnet"
subnet_name               = "my-subnet"
admin_username            = "azureuser"
aami_config_server_url    = "http://config.example.com:8080"
aami_bootstrap_token      = "your-bootstrap-token"
aami_primary_group        = "infrastructure:azure/eastus"
```

### 2. Initialize Terraform

```bash
terraform init
```

### 3. Plan Deployment

```bash
terraform plan
```

### 4. Apply Configuration

```bash
terraform apply
```

### 5. Verify Deployment

After deployment completes, verify the instances:

```bash
# Show outputs
terraform output

# SSH to instance (AWS example)
ssh -i ~/.ssh/my-key.pem ubuntu@<public_ip>

# Check Node Exporter
curl http://<private_ip>:9100/metrics | grep node_

# Check AAMI metrics
curl http://<private_ip>:9100/metrics | grep aami_

# Check logs
sudo tail -f /var/log/aami/dynamic-check.log
```

## Configuration Options

### Common Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `instance_count` | Number of instances | 1 |
| `aami_config_server_url` | Config Server URL | Required |
| `aami_bootstrap_token` | Bootstrap token | Required |
| `aami_primary_group` | Primary monitoring group | cloud:region |

### Cloud-Specific Variables

#### AWS

| Variable | Description | Default |
|----------|-------------|---------|
| `instance_type` | EC2 instance type | p4d.24xlarge |
| `ami_id` | AMI ID | Ubuntu 22.04 |
| `vpc_id` | VPC ID | Required |
| `subnet_id` | Subnet ID | Required |
| `key_name` | SSH key pair name | Required |

#### GCP

| Variable | Description | Default |
|----------|-------------|---------|
| `machine_type` | Machine type | a2-highgpu-8g |
| `gcp_project` | Project ID | Required |
| `network` | VPC network | default |
| `subnetwork` | Subnetwork name | "" |

#### Azure

| Variable | Description | Default |
|----------|-------------|---------|
| `vm_size` | VM size | Standard_NC24ads_A100_v4 |
| `resource_group_name` | Resource group | Required |
| `vnet_name` | Virtual network | Required |
| `subnet_name` | Subnet name | Required |

## Customization

### Modifying Instance Type

Change GPU instance types by setting the appropriate variable:

```hcl
# AWS - p3, p4, p5 instances
instance_type = "p3.8xlarge"  # 4x V100 GPUs

# GCP - a2, g2 instances
machine_type = "a2-ultragpu-8g"  # 8x A100 GPUs

# Azure - NC, ND series
vm_size = "Standard_ND96asr_v4"  # 8x A100 GPUs
```

### Adding Tags/Labels

Add custom tags by modifying the resource tags:

```hcl
# In the resource definition
tags = {
  Name        = "my-gpu-node"
  Project     = "ml-training"
  Owner       = "data-science-team"
  CostCenter  = "research"
}
```

### Modifying Disk Configuration

Adjust disk sizes and types:

```hcl
# AWS
root_block_device {
  volume_size = 200  # Increase to 200GB
  volume_type = "gp3"
  iops        = 16000
}

# GCP
boot_disk {
  initialize_params {
    size = 200  # 200GB
    type = "pd-ssd"
  }
}

# Azure
os_disk {
  disk_size_gb         = 200  # 200GB
  storage_account_type = "Premium_LRS"
}
```

### Custom Bootstrap Scripts

Modify the user_data/startup_script to add custom initialization:

```hcl
user_data = templatefile("${path.module}/my-custom-userdata.sh", {
  bootstrap_token    = var.aami_bootstrap_token
  config_server_url  = var.aami_config_server_url
  primary_group      = var.aami_primary_group
  custom_param       = "my_value"
})
```

## Monitoring Integration

### Prometheus Service Discovery

After deployment, update Prometheus configuration to scrape the new instances:

```bash
# Get Node Exporter URLs
terraform output node_exporter_urls

# Update Prometheus service discovery
curl http://config-server:8080/api/v1/sd/prometheus/file > /etc/prometheus/sd/aami.json
```

### Grafana Dashboards

Import AAMI dashboards to visualize metrics:

1. Navigate to Grafana
2. Import dashboard from `config/grafana/dashboards/`
3. Select Prometheus datasource

## Troubleshooting

### Bootstrap Failed

```bash
# SSH to instance
ssh -i ~/.ssh/key.pem ubuntu@<ip>

# Check bootstrap log
sudo cat /var/log/aami-bootstrap.log

# Check cloud-init log
sudo cat /var/log/cloud-init-output.log

# Manually re-run bootstrap
sudo /usr/local/bin/dynamic-check.sh --debug
```

### Metrics Not Appearing

```bash
# Check Node Exporter
systemctl status node_exporter
curl http://localhost:9100/metrics

# Check dynamic check timer
systemctl status aami-dynamic-check.timer
systemctl list-timers aami-dynamic-check.timer

# Check logs
sudo tail -f /var/log/aami/dynamic-check.log
```

### Network Connectivity Issues

```bash
# Test Config Server connectivity
curl -v http://config-server:8080/api/v1/checks/node/$(hostname)

# Check security group/firewall rules
# AWS
aws ec2 describe-security-groups --group-ids sg-xxxxx

# GCP
gcloud compute firewall-rules list

# Azure
az network nsg rule list --resource-group xxx --nsg-name xxx
```

## Cleanup

To destroy all resources:

```bash
terraform destroy
```

**Warning**: This will permanently delete all instances and associated resources.

## Best Practices

1. **Use Remote State**: Store Terraform state in S3/GCS/Azure Storage
2. **Enable State Locking**: Use DynamoDB/GCS/Azure Blob for state locking
3. **Use Modules**: Create reusable modules for common patterns
4. **Tag Everything**: Add comprehensive tags for cost tracking
5. **Secure Secrets**: Use AWS Secrets Manager/GCP Secret Manager/Azure Key Vault
6. **Enable Encryption**: Encrypt disks and use secure boot
7. **Regular Backups**: Configure automated snapshots/backups
8. **Cost Monitoring**: Set up budget alerts

## References

- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Terraform GCP Provider](https://registry.terraform.io/providers/hashicorp/google/latest/docs)
- [Terraform Azure Provider](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)
- [AAMI Documentation](/docs/en/)
