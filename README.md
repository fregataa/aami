# AAMI (AI Accelerator Monitoring Infrastructure)

**Hybrid Infrastructure Monitoring System for AI Accelerator Clusters**

AAMI is an integrated monitoring solution designed to efficiently monitor and manage large-scale AI accelerator infrastructure (GPUs, NPUs, etc.). Built on the Prometheus ecosystem, it supports group-based hierarchical structure, dynamic target management, and customizable alert rules.

[![CI](https://github.com/your-org/aami/workflows/CI/badge.svg)](https://github.com/your-org/aami/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

## Key Features

### Flexible Infrastructure Management
- **3 Independent Namespaces**: Infrastructure, Logical, Environment
- **Hierarchical Group Structure**: Unlimited depth group tree with multiple membership support
- **Namespace Priority System**: Intelligent policy inheritance and override

### Dynamic Target Management
- **Automatic Service Discovery**: Prometheus file-based SD integration
- **Real-time Configuration**: Instant reflection of target and check updates
- **Customizable Monitoring**: Per-group check intervals and metric paths
- **Individual Overrides**: Fine-grained control for specific servers

### AI Accelerator Monitoring
- **GPU Support**: NVIDIA (DCGM), AMD, Intel
- **NPU Support**: Various vendors (Gaudi, TPU, etc.)
- **High-Speed Networks**: InfiniBand, RoCE monitoring
- **Parallel Filesystems**: Lustre, BeeGFS, GPFS health checks
- **Storage Performance**: NVMe over Fabrics monitoring

### Intelligent Alert Management
- **Rule Templates**: Predefined alert rules with customizable thresholds
- **Group-Based Customization**: Environment-specific alert thresholds
- **Policy Inheritance**: Smart merging with priority-based override
- **Policy Tracing**: Debug which group settings are applied

### Enterprise Ready
- **Microservice Architecture**: Independent scaling of components
- **High Availability**: PostgreSQL + Redis for reliability
- **Air-Gapped Support**: Complete offline installation packages
- **Hybrid Deployment**: Kubernetes, Docker Compose, or bare metal

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Config Server API                          │
│  (Go) - Group/Target Management, Dynamic SD, Alert Customization│
└──────────────┬───────────────────────────────┬──────────────────┘
               │                               │
       ┌───────▼────────┐            ┌────────▼─────────┐
       │   PostgreSQL   │            │      Redis       │
       │   (Metadata)   │            │   (SD Cache)     │
       └────────────────┘            └──────────────────┘
               │
       ┌───────▼──────────────────────────────────────────────────┐
       │              Prometheus (TSDB)                           │
       │  - Automatic target discovery via file-based SD          │
       │  - Auto-generated group-based alert rules                │
       └──────────────────┬───────────────────────────────────────┘
                          │
       ┌──────────────────┼──────────────────┐
       │                  │                  │
   ┌───▼────┐      ┌─────▼──────┐     ┌────▼─────┐
   │ Node   │      │  GPU/NPU   │     │ Custom   │
   │Exporter│      │ Exporters  │     │Exporters │
   └────────┘      └────────────┘     └──────────┘
       │                  │                  │
   ┌───▼──────────────────▼──────────────────▼──────┐
   │         AI Accelerator Cluster                 │
   │  (GPU Servers, NPU Nodes, Storage, Network)    │
   └────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Docker 20.10+ and Docker Compose v2.0+
- Go 1.21+ (for development)
- PostgreSQL 15+ and Redis 7+ (or use Docker)

### Installation

#### Option 1: Docker Compose (Recommended for Quick Start)

```bash
git clone https://github.com/your-org/aami.git
cd aami/deploy/docker-compose

cp .env.example .env
# Edit .env with your configuration

docker-compose up -d
```

Access the services:
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- Config Server API: http://localhost:8080

#### Option 2: Kubernetes

```bash
cd deploy/kubernetes
kubectl apply -k .
```

#### Option 3: Air-Gapped Installation

```bash
cd deploy/offline
./create-bundle.sh    # On internet-connected machine
./install-offline.sh  # On air-gapped machine
```

For detailed installation instructions, see [Deployment Guide](deploy/README.md).

### Agent Deployment

Deploy monitoring agents to your target nodes:

```bash
# One-line bootstrap installation
curl -fsSL https://your-server/bootstrap.sh | bash -s -- --token YOUR_TOKEN

# Manual installation
curl -fsSL https://raw.githubusercontent.com/your-org/aami/main/scripts/node/install-node-exporter.sh | bash
```

See [examples/](examples/) for cloud-init and Terraform configurations.

## Development

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/your-org/aami.git
cd aami

# Start dependencies
cd deploy/docker-compose
docker-compose up -d postgres redis

# Run Config Server
cd ../../services/config-server
go mod download
go run cmd/server/main.go
```

### Running Tests

```bash
cd services/config-server
go test ./...
go test -cover ./...
```

### Code Quality

```bash
# Linting
golangci-lint run

# Formatting
go fmt ./...
```

For detailed development setup, see [Development Guide](docs/en/DEVELOPMENT.md) ([한국어](docs/kr/DEVELOPMENT.md)).

## Documentation

- **[Quick Start Guide](docs/en/QUICKSTART.md)** - Step-by-step tutorial for getting started
- **[API Reference](docs/en/API.md)** - Complete REST API documentation
- **[Development Guide](docs/en/DEVELOPMENT.md)** - Environment setup and development workflow ([한국어](docs/kr/DEVELOPMENT.md))
- **[Deployment Guide](deploy/README.md)** - Deployment options and strategies
- **[API Usage Examples](examples/api-usage/)** - Code examples for common operations
- **[Configuration Examples](examples/)** - Example configurations and templates

## Project Structure

```
aami/
├── config/                   # Configuration files (Prometheus, Grafana, Alertmanager)
├── deploy/                   # Deployment configurations
│   ├── docker-compose/       # Docker Compose setup
│   ├── kubernetes/           # Kubernetes manifests
│   ├── ansible/              # Ansible playbooks
│   └── offline/              # Air-gapped installation
├── docs/                     # Documentation
│   ├── en/                   # English documentation
│   └── kr/                   # Korean documentation
├── examples/                 # Configuration examples
├── scripts/                  # Utility scripts
├── services/                 # Microservices
│   ├── config-server/        # Config Server (Go)
│   ├── config-server-ui/     # Web UI (Next.js, optional)
│   └── exporters/            # Custom exporters
├── tests/                    # Integration tests
└── tools/                    # Development tools
```

## Technology Stack

**Backend**: Go 1.21+, Gin, PostgreSQL 15, Redis 7, GORM v2
**Monitoring**: Prometheus 2.45+, Grafana 10.0+, Alertmanager 0.26+
**Deployment**: Docker, Kubernetes, Ansible
**CI/CD**: GitHub Actions

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards

- All code comments must be in English
- Follow Go [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Write tests for new features
- Ensure CI passes before submitting PR

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/your-org/aami/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/aami/discussions)
- **Documentation**: [Wiki](https://github.com/your-org/aami/wiki)

---

**Built with ❤️ for AI Infrastructure Teams**
