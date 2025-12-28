# AWS GPU Instance with AAMI Monitoring
#
# This Terraform configuration creates a GPU instance (p4d.24xlarge)
# on AWS EC2 with automatic AAMI monitoring bootstrap.
#
# Prerequisites:
# - AWS credentials configured
# - VPC and subnet created
# - AAMI Config Server deployed and accessible
# - Bootstrap token created in AAMI Config Server

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Variables
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "instance_count" {
  description = "Number of instances to create"
  type        = number
  default     = 1
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "p4d.24xlarge" # 8x NVIDIA A100 GPUs
}

variable "ami_id" {
  description = "AMI ID (Ubuntu 22.04 LTS recommended)"
  type        = string
  # Ubuntu 22.04 LTS in us-east-1
  default     = "ami-0c7217cdde317cfec"
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "subnet_id" {
  description = "Subnet ID"
  type        = string
}

variable "key_name" {
  description = "SSH key pair name"
  type        = string
}

variable "aami_config_server_url" {
  description = "AAMI Config Server URL"
  type        = string
}

variable "aami_bootstrap_token" {
  description = "AAMI bootstrap token"
  type        = string
  sensitive   = true
}

variable "aami_primary_group" {
  description = "Primary group for AAMI monitoring"
  type        = string
  default     = "infrastructure:aws/us-east-1"
}

# Security Group
resource "aws_security_group" "aami_gpu" {
  name        = "aami-gpu-instance"
  description = "Security group for AAMI GPU instances"
  vpc_id      = var.vpc_id

  # SSH access
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"] # Adjust to your internal network
  }

  # Node Exporter metrics (internal only)
  ingress {
    description = "Node Exporter"
    from_port   = 9100
    to_port     = 9100
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  # DCGM Exporter metrics (internal only)
  ingress {
    description = "DCGM Exporter"
    from_port   = 9400
    to_port     = 9400
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  # Outbound internet access
  egress {
    description = "All outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "aami-gpu-instance"
    ManagedBy   = "terraform"
    Purpose     = "aami-monitoring"
  }
}

# IAM Role for EC2 (optional, for CloudWatch logs)
resource "aws_iam_role" "aami_gpu" {
  name = "aami-gpu-instance-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name      = "aami-gpu-instance-role"
    ManagedBy = "terraform"
  }
}

resource "aws_iam_role_policy_attachment" "aami_gpu_cloudwatch" {
  role       = aws_iam_role.aami_gpu.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"
}

resource "aws_iam_instance_profile" "aami_gpu" {
  name = "aami-gpu-instance-profile"
  role = aws_iam_role.aami_gpu.name
}

# GPU Instance
resource "aws_instance" "gpu_node" {
  count = var.instance_count

  ami           = var.ami_id
  instance_type = var.instance_type
  key_name      = var.key_name
  subnet_id     = var.subnet_id

  vpc_security_group_ids = [aws_security_group.aami_gpu.id]
  iam_instance_profile   = aws_iam_instance_profile.aami_gpu.name

  # Root volume (boot disk)
  root_block_device {
    volume_type = "gp3"
    volume_size = 100
    iops        = 3000
    throughput  = 125
    encrypted   = true

    tags = {
      Name = "aami-gpu-node-${count.index}-root"
    }
  }

  # Additional EBS volume for data
  ebs_block_device {
    device_name = "/dev/sdf"
    volume_type = "gp3"
    volume_size = 1000
    iops        = 16000
    throughput  = 1000
    encrypted   = true

    tags = {
      Name = "aami-gpu-node-${count.index}-data"
    }
  }

  # User data for bootstrap
  user_data = templatefile("${path.module}/../cloud-init/aws-ec2-userdata.sh", {
    bootstrap_token    = var.aami_bootstrap_token
    config_server_url  = var.aami_config_server_url
    primary_group      = var.aami_primary_group
  })

  # Enable detailed monitoring
  monitoring = true

  tags = {
    Name        = "aami-gpu-node-${count.index}"
    Environment = "production"
    ManagedBy   = "terraform"
    Purpose     = "gpu-compute"
    AAMIGroup   = var.aami_primary_group
  }

  volume_tags = {
    ManagedBy = "terraform"
  }

  # Wait for instance to be ready
  lifecycle {
    create_before_destroy = false
  }
}

# Outputs
output "instance_ids" {
  description = "EC2 instance IDs"
  value       = aws_instance.gpu_node[*].id
}

output "instance_public_ips" {
  description = "Public IP addresses"
  value       = aws_instance.gpu_node[*].public_ip
}

output "instance_private_ips" {
  description = "Private IP addresses"
  value       = aws_instance.gpu_node[*].private_ip
}

output "instance_public_dns" {
  description = "Public DNS names"
  value       = aws_instance.gpu_node[*].public_dns
}

output "node_exporter_urls" {
  description = "Node Exporter metrics endpoints"
  value = [
    for instance in aws_instance.gpu_node :
    "http://${instance.private_ip}:9100/metrics"
  ]
}

output "dcgm_exporter_urls" {
  description = "DCGM Exporter metrics endpoints"
  value = [
    for instance in aws_instance.gpu_node :
    "http://${instance.private_ip}:9400/metrics"
  ]
}
