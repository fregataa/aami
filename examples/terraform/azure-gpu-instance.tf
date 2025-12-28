# Azure GPU Instance with AAMI Monitoring
#
# This Terraform configuration creates a GPU instance (Standard_NC24ads_A100_v4)
# on Azure with automatic AAMI monitoring bootstrap.
#
# Prerequisites:
# - Azure credentials configured
# - Resource group created
# - Virtual network and subnet created
# - AAMI Config Server deployed and accessible
# - Bootstrap token created in AAMI Config Server

terraform {
  required_version = ">= 1.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

# Variables
variable "resource_group_name" {
  description = "Resource group name"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "eastus"
}

variable "instance_count" {
  description = "Number of instances to create"
  type        = number
  default     = 1
}

variable "vm_size" {
  description = "VM size"
  type        = string
  default     = "Standard_NC24ads_A100_v4" # 1x NVIDIA A100 GPU
}

variable "vnet_name" {
  description = "Virtual network name"
  type        = string
}

variable "subnet_name" {
  description = "Subnet name"
  type        = string
}

variable "admin_username" {
  description = "Admin username"
  type        = string
  default     = "azureuser"
}

variable "ssh_public_key_path" {
  description = "Path to SSH public key"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
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
  default     = "infrastructure:azure/eastus"
}

# Data sources
data "azurerm_resource_group" "main" {
  name = var.resource_group_name
}

data "azurerm_virtual_network" "main" {
  name                = var.vnet_name
  resource_group_name = data.azurerm_resource_group.main.name
}

data "azurerm_subnet" "main" {
  name                 = var.subnet_name
  virtual_network_name = data.azurerm_virtual_network.main.name
  resource_group_name  = data.azurerm_resource_group.main.name
}

# Network Security Group
resource "azurerm_network_security_group" "aami_gpu" {
  name                = "aami-gpu-nsg"
  location            = var.location
  resource_group_name = data.azurerm_resource_group.main.name

  # SSH
  security_rule {
    name                       = "AllowSSHInternal"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "10.0.0.0/8"
    destination_address_prefix = "*"
  }

  # Node Exporter
  security_rule {
    name                       = "AllowNodeExporterInternal"
    priority                   = 110
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "9100"
    source_address_prefix      = "10.0.0.0/8"
    destination_address_prefix = "*"
  }

  # DCGM Exporter
  security_rule {
    name                       = "AllowDCGMExporterInternal"
    priority                   = 120
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "9400"
    source_address_prefix      = "10.0.0.0/8"
    destination_address_prefix = "*"
  }

  tags = {
    ManagedBy = "terraform"
    Purpose   = "aami-monitoring"
  }
}

# Network Interface
resource "azurerm_network_interface" "aami_gpu" {
  count = var.instance_count

  name                = "aami-gpu-node-${count.index}-nic"
  location            = var.location
  resource_group_name = data.azurerm_resource_group.main.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = data.azurerm_subnet.main.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.aami_gpu[count.index].id
  }

  tags = {
    ManagedBy = "terraform"
  }
}

# Associate NSG with NIC
resource "azurerm_network_interface_security_group_association" "aami_gpu" {
  count = var.instance_count

  network_interface_id      = azurerm_network_interface.aami_gpu[count.index].id
  network_security_group_id = azurerm_network_security_group.aami_gpu.id
}

# Public IP
resource "azurerm_public_ip" "aami_gpu" {
  count = var.instance_count

  name                = "aami-gpu-node-${count.index}-pip"
  location            = var.location
  resource_group_name = data.azurerm_resource_group.main.name
  allocation_method   = "Static"
  sku                 = "Standard"

  tags = {
    ManagedBy = "terraform"
  }
}

# Managed Disk for data
resource "azurerm_managed_disk" "data" {
  count = var.instance_count

  name                 = "aami-gpu-node-${count.index}-data"
  location             = var.location
  resource_group_name  = data.azurerm_resource_group.main.name
  storage_account_type = "Premium_LRS"
  create_option        = "Empty"
  disk_size_gb         = 1024

  tags = {
    ManagedBy = "terraform"
    Purpose   = "data"
  }
}

# Virtual Machine
resource "azurerm_linux_virtual_machine" "gpu_node" {
  count = var.instance_count

  name                = "aami-gpu-node-${count.index}"
  location            = var.location
  resource_group_name = data.azurerm_resource_group.main.name
  size                = var.vm_size

  admin_username = var.admin_username

  network_interface_ids = [
    azurerm_network_interface.aami_gpu[count.index].id
  ]

  admin_ssh_key {
    username   = var.admin_username
    public_key = file(var.ssh_public_key_path)
  }

  os_disk {
    name                 = "aami-gpu-node-${count.index}-osdisk"
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
    disk_size_gb         = 100
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }

  # Custom data (cloud-init)
  custom_data = base64encode(templatefile("${path.module}/../cloud-init/azure-custom-data.yaml", {
    bootstrap_token    = var.aami_bootstrap_token
    config_server_url  = var.aami_config_server_url
    primary_group      = var.aami_primary_group
  }))

  # Enable boot diagnostics
  boot_diagnostics {
    storage_account_uri = null # Use managed storage
  }

  tags = {
    Name        = "aami-gpu-node-${count.index}"
    Environment = "production"
    ManagedBy   = "terraform"
    Purpose     = "gpu-compute"
    AAMIGroup   = var.aami_primary_group
  }
}

# Attach data disk
resource "azurerm_virtual_machine_data_disk_attachment" "data" {
  count = var.instance_count

  managed_disk_id    = azurerm_managed_disk.data[count.index].id
  virtual_machine_id = azurerm_linux_virtual_machine.gpu_node[count.index].id
  lun                = 0
  caching            = "ReadWrite"
}

# Outputs
output "vm_ids" {
  description = "Virtual machine IDs"
  value       = azurerm_linux_virtual_machine.gpu_node[*].id
}

output "vm_names" {
  description = "Virtual machine names"
  value       = azurerm_linux_virtual_machine.gpu_node[*].name
}

output "public_ips" {
  description = "Public IP addresses"
  value       = azurerm_public_ip.aami_gpu[*].ip_address
}

output "private_ips" {
  description = "Private IP addresses"
  value       = azurerm_network_interface.aami_gpu[*].private_ip_address
}

output "node_exporter_urls" {
  description = "Node Exporter metrics endpoints"
  value = [
    for nic in azurerm_network_interface.aami_gpu :
    "http://${nic.private_ip_address}:9100/metrics"
  ]
}

output "dcgm_exporter_urls" {
  description = "DCGM Exporter metrics endpoints"
  value = [
    for nic in azurerm_network_interface.aami_gpu :
    "http://${nic.private_ip_address}:9400/metrics"
  ]
}
