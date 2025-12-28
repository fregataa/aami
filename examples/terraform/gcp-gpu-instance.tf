# GCP GPU Instance with AAMI Monitoring
#
# This Terraform configuration creates a GPU instance (a2-highgpu-8g)
# on Google Cloud Compute Engine with automatic AAMI monitoring bootstrap.
#
# Prerequisites:
# - GCP credentials configured
# - VPC network created
# - AAMI Config Server deployed and accessible
# - Bootstrap token created in AAMI Config Server

terraform {
  required_version = ">= 1.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

# Variables
variable "gcp_project" {
  description = "GCP project ID"
  type        = string
}

variable "gcp_region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "gcp_zone" {
  description = "GCP zone"
  type        = string
  default     = "us-central1-a"
}

variable "instance_count" {
  description = "Number of instances to create"
  type        = number
  default     = 1
}

variable "machine_type" {
  description = "Machine type"
  type        = string
  default     = "a2-highgpu-8g" # 8x NVIDIA A100 GPUs
}

variable "network" {
  description = "VPC network name"
  type        = string
  default     = "default"
}

variable "subnetwork" {
  description = "Subnetwork name"
  type        = string
  default     = ""
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
  default     = "infrastructure:gcp/us-central1"
}

# Service Account
resource "google_service_account" "aami_gpu" {
  account_id   = "aami-gpu-instance"
  display_name = "AAMI GPU Instance Service Account"
  description  = "Service account for AAMI GPU instances"
}

resource "google_project_iam_member" "aami_gpu_logging" {
  project = var.gcp_project
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.aami_gpu.email}"
}

resource "google_project_iam_member" "aami_gpu_monitoring" {
  project = var.gcp_project
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.aami_gpu.email}"
}

# Firewall Rules
resource "google_compute_firewall" "aami_gpu_ssh" {
  name    = "aami-gpu-allow-ssh-internal"
  network = var.network

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["10.0.0.0/8"]
  target_tags   = ["aami-gpu-node"]

  description = "Allow SSH from internal network"
}

resource "google_compute_firewall" "aami_gpu_node_exporter" {
  name    = "aami-gpu-allow-node-exporter-internal"
  network = var.network

  allow {
    protocol = "tcp"
    ports    = ["9100"]
  }

  source_ranges = ["10.0.0.0/8"]
  target_tags   = ["aami-gpu-node"]

  description = "Allow Node Exporter metrics from internal network"
}

resource "google_compute_firewall" "aami_gpu_dcgm_exporter" {
  name    = "aami-gpu-allow-dcgm-exporter-internal"
  network = var.network

  allow {
    protocol = "tcp"
    ports    = ["9400"]
  }

  source_ranges = ["10.0.0.0/8"]
  target_tags   = ["aami-gpu-node"]

  description = "Allow DCGM Exporter metrics from internal network"
}

# GPU Instance
resource "google_compute_instance" "gpu_node" {
  count = var.instance_count

  name         = "aami-gpu-node-${count.index}"
  machine_type = var.machine_type
  zone         = var.gcp_zone

  tags = ["aami-gpu-node"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 100
      type  = "pd-ssd"
    }
  }

  # Additional persistent disk for data
  attached_disk {
    source      = google_compute_disk.data_disk[count.index].id
    device_name = "data"
  }

  # GPU accelerators
  guest_accelerator {
    type  = "nvidia-tesla-a100"
    count = 8
  }

  # GPU instances require on-host maintenance
  scheduling {
    on_host_maintenance = "TERMINATE"
    automatic_restart   = true
  }

  network_interface {
    network    = var.network
    subnetwork = var.subnetwork

    # Assign external IP
    access_config {
      # Ephemeral IP
    }
  }

  metadata = {
    ssh-keys = "ubuntu:${file("~/.ssh/id_rsa.pub")}"
  }

  metadata_startup_script = templatefile("${path.module}/../cloud-init/gcp-startup-script.sh", {
    bootstrap_token    = var.aami_bootstrap_token
    config_server_url  = var.aami_config_server_url
    primary_group      = var.aami_primary_group
  })

  service_account {
    email  = google_service_account.aami_gpu.email
    scopes = ["cloud-platform"]
  }

  labels = {
    environment = "production"
    managed_by  = "terraform"
    purpose     = "gpu-compute"
    aami_group  = replace(var.aami_primary_group, ":", "-")
  }

  lifecycle {
    create_before_destroy = false
  }
}

# Data Disk
resource "google_compute_disk" "data_disk" {
  count = var.instance_count

  name = "aami-gpu-node-${count.index}-data"
  type = "pd-ssd"
  zone = var.gcp_zone
  size = 1000

  labels = {
    managed_by = "terraform"
    purpose    = "gpu-data"
  }
}

# Outputs
output "instance_names" {
  description = "Instance names"
  value       = google_compute_instance.gpu_node[*].name
}

output "instance_ids" {
  description = "Instance IDs"
  value       = google_compute_instance.gpu_node[*].instance_id
}

output "instance_internal_ips" {
  description = "Internal IP addresses"
  value       = google_compute_instance.gpu_node[*].network_interface[0].network_ip
}

output "instance_external_ips" {
  description = "External IP addresses"
  value = [
    for instance in google_compute_instance.gpu_node :
    instance.network_interface[0].access_config[0].nat_ip
  ]
}

output "node_exporter_urls" {
  description = "Node Exporter metrics endpoints"
  value = [
    for instance in google_compute_instance.gpu_node :
    "http://${instance.network_interface[0].network_ip}:9100/metrics"
  ]
}

output "dcgm_exporter_urls" {
  description = "DCGM Exporter metrics endpoints"
  value = [
    for instance in google_compute_instance.gpu_node :
    "http://${instance.network_interface[0].network_ip}:9400/metrics"
  ]
}
