# Examples

This directory contains example configurations and templates for various use cases.

## Directory Structure

```
examples/
├── cloud-init/          # Cloud-init configurations
├── terraform/           # Terraform infrastructure examples
└── custom-checks/       # Custom check script examples
```

## Example Categories

### Cloud-Init Configurations
- **Location**: `cloud-init/`
- **Purpose**: Automated VM initialization with AAMI agents

Available examples:
- `aws-ec2.yaml` - AWS EC2 user-data with bootstrap
- `gcp-compute.yaml` - GCP Compute Engine startup script
- `azure-vm.yaml` - Azure VM custom data
- `openstack.yaml` - OpenStack cloud-init

Example usage:
```bash
# AWS EC2 with cloud-init
aws ec2 run-instances \
  --image-id ami-xxxxx \
  --instance-type p4d.24xlarge \
  --user-data file://cloud-init/aws-ec2.yaml
```

### Terraform Infrastructure
- **Location**: `terraform/`
- **Purpose**: Infrastructure as Code examples

Available examples:
- `aws-gpu-cluster/` - AWS GPU cluster with auto-monitoring
- `gcp-tpu-pods/` - GCP TPU pods setup
- `hybrid-multi-cloud/` - Multi-cloud deployment
- `on-premise-proxmox/` - On-premise Proxmox cluster

Example usage:
```bash
cd terraform/aws-gpu-cluster
terraform init
terraform plan
terraform apply
```

### Custom Check Scripts
- **Location**: `custom-checks/`
- **Purpose**: Examples of custom monitoring checks

Available examples:
- `check_gpu_utilization.py` - GPU utilization check
- `check_infiniband.sh` - InfiniBand link status
- `check_lustre_mdt.sh` - Lustre metadata server check
- `check_nvme_health.sh` - NVMe SSD health check
- `check_raid_status.sh` - RAID array status check

Example usage:
```bash
# Run custom check
./custom-checks/check_gpu_utilization.py --threshold 90

# Output for textfile collector
./custom-checks/check_nvme_health.sh > /var/lib/node_exporter/textfile_collector/nvme.prom
```

## Using Examples

### 1. Copy and Customize

```bash
# Copy example to your project
cp examples/cloud-init/aws-ec2.yaml my-config.yaml

# Edit with your settings
vim my-config.yaml
```

### 2. Test Before Deployment

```bash
# Validate YAML syntax
yamllint examples/cloud-init/aws-ec2.yaml

# Test Terraform plan
cd examples/terraform/aws-gpu-cluster
terraform plan -out=tfplan
```

### 3. Deploy

```bash
# Deploy with modified configuration
terraform apply tfplan
```

## Configuration Variables

Most examples use placeholders that need to be replaced:

- `${BOOTSTRAP_TOKEN}` - Bootstrap token from Config Server
- `${CONFIG_SERVER_URL}` - Config Server endpoint
- `${AWS_REGION}` - AWS region
- `${SSH_KEY}` - SSH public key
- `${TAGS}` - Resource tags

Example:
```bash
# Replace placeholders
sed -i "s/\${BOOTSTRAP_TOKEN}/your-token-here/g" cloud-init/aws-ec2.yaml
sed -i "s/\${CONFIG_SERVER_URL}/https:\/\/config.example.com/g" cloud-init/aws-ec2.yaml
```

## Best Practices

1. **Version Control**: Keep customized configs in your own repository
2. **Secrets Management**: Use secret managers (AWS Secrets Manager, HashiCorp Vault)
3. **Testing**: Test in non-production environment first
4. **Documentation**: Document customizations and deployment notes
5. **Security**: Review security groups, IAM roles, and network policies

## Contributing Examples

When adding new examples:

1. Create a descriptive subdirectory
2. Include a README.md with:
   - Purpose and use case
   - Prerequisites
   - Step-by-step instructions
   - Expected output
   - Troubleshooting tips
3. Add comments explaining key configurations
4. Test thoroughly before committing
5. Use English for all documentation and comments

## Quick Links

- [Configuration Guide](../docs/configuration.md)
- [Deployment Guide](../docs/DEPLOYMENT.md)
- [Custom Checks Documentation](../docs/custom-checks.md)
