# Deployment

This directory contains deployment configurations and scripts for various environments.

## Directory Structure

```
deploy/
├── docker-compose/      # Docker Compose deployments
├── kubernetes/          # Kubernetes manifests
├── ansible/            # Ansible playbooks
└── offline/            # Offline/air-gapped installation
```

## Deployment Options

### Docker Compose (Recommended for Development)
- **Location**: `docker-compose/`
- **Use Case**: Local development, small-scale deployments
- **Documentation**: See [docker-compose/README.md](docker-compose/README.md)

Quick start:
```bash
cd docker-compose
cp .env.example .env
# Edit .env file
docker-compose up -d
```

### Kubernetes (Recommended for Production)
- **Location**: `kubernetes/`
- **Use Case**: Production environments, large-scale deployments
- **Documentation**: See [kubernetes/README.md](kubernetes/README.md)

Quick start:
```bash
cd kubernetes
kubectl apply -f namespace.yaml
kubectl apply -k .
```

### Ansible Automation
- **Location**: `ansible/`
- **Use Case**: Automated deployment to multiple servers
- **Documentation**: See [ansible/README.md](ansible/README.md)

Quick start:
```bash
cd ansible
ansible-playbook -i inventory/hosts.yml playbooks/deploy-all.yml
```

### Offline Installation
- **Location**: `offline/`
- **Use Case**: Air-gapped environments, closed networks
- **Documentation**: See [offline/README.md](offline/README.md)

Quick start:
```bash
cd offline
./create-bundle.sh        # On internet-connected machine
./install-offline.sh      # On air-gapped machine
```

## Environment Support

- **Development**: Docker Compose
- **Staging**: Docker Compose or Kubernetes
- **Production**: Kubernetes (recommended) or Ansible-managed VMs
- **Air-gapped**: Offline installation packages

## Prerequisites

- Docker 20.10+ & Docker Compose v2.0+
- Kubernetes 1.24+ (for k8s deployment)
- Ansible 2.9+ (for Ansible deployment)
- Go 1.21+ (for building from source)

## Quick Links

- [Main README](../README.md)
- [Installation Guide](../docs/installation.md)
- [Configuration Guide](../docs/configuration.md)
