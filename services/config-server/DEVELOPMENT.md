# Development Guide - AAMI Config Server

This guide explains how to run and develop the AAMI Config Server in local and cloud environments.

## 📑 Table of Contents

1. [Local Environment - Docker Compose (Recommended)](#1-local-environment---docker-compose-recommended)
2. [Local Environment - Direct Execution](#2-local-environment---direct-execution)
3. [Development Environment - Hot Reload](#3-development-environment---hot-reload)
4. [Cloud Environment - Kubernetes](#4-cloud-environment---kubernetes)
5. [API Testing](#5-api-testing)
6. [Sample Data Generation](#6-sample-data-generation)

---

## 1. Local Environment - Docker Compose (Recommended)

The fastest and simplest way. PostgreSQL and Config Server start automatically.

### Prerequisites

- Docker Desktop or Docker Engine
- Docker Compose v2+

### Starting

```bash
# 1. Navigate to project directory
cd /path/to/aami/services/config-server

# 2. Start all services with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f config-server

# Check status
docker-compose ps
```

### Connection Information

- **Config Server API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **PostgreSQL**: localhost:5432 (postgres/postgres)

### Testing

```bash
# Health check
curl http://localhost:8080/health

# List namespaces
curl http://localhost:8080/api/v1/namespaces

# Service Discovery
curl http://localhost:8080/api/v1/sd/prometheus
```

### Stopping

```bash
# Stop containers
docker-compose down

# Remove data volumes as well
docker-compose down -v
```

---

## 2. Local Environment - Direct Execution

Run directly in a Go development environment.

### Prerequisites

- Go 1.25 or higher
- PostgreSQL 15 or higher (running)

### Step 1: PostgreSQL Setup

```bash
# Verify PostgreSQL is running
psql -U postgres -c "SELECT version();"

# Create database
psql -U postgres -c "CREATE DATABASE aami_config;"
```

### Step 2: Environment Variables

```bash
# Create .env file
cat > .env << 'EOF'
# Server
PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=aami_config
DB_SSLMODE=disable
EOF
```

### Step 3: Dependencies and Build

```bash
# Download Go modules
go mod download

# Build
go build -o config-server ./cmd/config-server
```

### Step 4: Run

```bash
# Load environment variables and run
export $(cat .env | xargs) && ./config-server

# Or set environment variables directly
DB_HOST=localhost \
DB_PORT=5432 \
DB_USER=postgres \
DB_PASSWORD=postgres \
DB_NAME=aami_config \
PORT=8080 \
./config-server
```

### Step 5: Verification

```bash
# Health check
curl http://localhost:8080/health

# Expected output:
{
  "status": "healthy",
  "timestamp": "2025-12-28T19:00:00Z",
  "version": "v1.0.0",
  "components": {
    "database": {
      "name": "database",
      "status": "healthy",
      "message": "Database connection is healthy",
      ...
    }
  }
}
```

---

## 3. Development Environment - Hot Reload

Development environment with automatic rebuild on code changes.

### Prerequisites

- Docker Desktop
- Docker Compose v2+

### Starting

```bash
# Start development environment (Hot reload + pgAdmin included)
docker-compose -f docker-compose.dev.yml up -d

# View logs (watch for automatic rebuilds)
docker-compose -f docker-compose.dev.yml logs -f config-server
```

### Additional Services

- **pgAdmin**: http://localhost:5050
  - Email: admin@aami.local
  - Password: admin

### Code Editing

```bash
# Modify source code - automatic rebuild triggers
vim internal/api/handler/target.go

# Watch rebuild in logs
# [air] building...
# [air] build succeeded
# [air] restarting process...
```

---

## 4. Cloud Environment - Kubernetes

Deploy to a Kubernetes cluster.

### Prerequisites

- Kubernetes cluster (Minikube, GKE, EKS, AKS, etc.)
- kubectl installed and configured
- (Optional) Ingress Controller (NGINX)

### Option A: Deploy with kubectl

```bash
# 1. Create namespace
kubectl apply -f k8s/namespace.yaml

# 2. Create ConfigMap and Secret
kubectl apply -f k8s/configmap.yaml

# ⚠️ IMPORTANT: Update Secret for production!
vim k8s/secret.yaml
# Change DB_PASSWORD
kubectl apply -f k8s/secret.yaml

# 3. Deploy Deployment and Service
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# 4. (Optional) Deploy Ingress
# Modify domain in k8s/ingress.yaml
vim k8s/ingress.yaml
kubectl apply -f k8s/ingress.yaml

# 5. (Optional) Deploy HPA
kubectl apply -f k8s/hpa.yaml
```

### Option B: Deploy with Kustomize

```bash
# Deploy all resources at once
kubectl apply -k k8s/

# Preview changes
kubectl kustomize k8s/
```

### Deployment Verification

```bash
# Check pod status
kubectl get pods -n aami

# Expected output:
NAME                            READY   STATUS    RESTARTS   AGE
config-server-7d8f9c5b6-abc12   1/1     Running   0          2m
config-server-7d8f9c5b6-def34   1/1     Running   0          2m
config-server-7d8f9c5b6-ghi56   1/1     Running   0          2m

# Check service
kubectl get svc -n aami

# Check logs
kubectl logs -f deployment/config-server -n aami

# Health check
kubectl port-forward svc/config-server 8080:80 -n aami
curl http://localhost:8080/health
```

### Cloud-Specific Guides

#### Google Cloud (GKE)

```bash
# Create GKE cluster
gcloud container clusters create aami-cluster \
  --zone=us-central1-a \
  --num-nodes=3 \
  --machine-type=n1-standard-2

# Configure credentials
gcloud container clusters get-credentials aami-cluster --zone=us-central1-a

# Deploy
kubectl apply -k k8s/
```

#### AWS (EKS)

```bash
# Create EKS cluster
eksctl create cluster \
  --name aami-cluster \
  --region us-west-2 \
  --nodes 3 \
  --node-type t3.medium

# Deploy
kubectl apply -k k8s/
```

#### Azure (AKS)

```bash
# Create AKS cluster
az aks create \
  --resource-group aami-rg \
  --name aami-cluster \
  --node-count 3 \
  --node-vm-size Standard_DS2_v2

# Configure credentials
az aks get-credentials --resource-group aami-rg --name aami-cluster

# Deploy
kubectl apply -k k8s/
```

---

## 5. API Testing

Once the server is running, test the following APIs.

### Health Check

```bash
curl http://localhost:8080/health

curl http://localhost:8080/health/ready

curl http://localhost:8080/health/live
```

### Namespace Management

```bash
# Create namespace
curl -X POST http://localhost:8080/api/v1/namespaces \
  -H "Content-Type: application/json" \
  -d '{
    "name": "infrastructure",
    "policy_priority": 10,
    "description": "Infrastructure namespace"
  }'

# List namespaces
curl http://localhost:8080/api/v1/namespaces

# Get specific namespace
curl http://localhost:8080/api/v1/namespaces/name/infrastructure
```

### Group Management

```bash
# Create group
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production-servers",
    "namespace_id": "<namespace-id>",
    "description": "Production servers group",
    "priority": 100
  }'

# List groups
curl http://localhost:8080/api/v1/groups
```

### Target Registration

```bash
# Create target
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "server-01",
    "ip_address": "192.168.1.10",
    "group_ids": ["<group-id>"],
    "status": "active",
    "labels": {
      "env": "production",
      "region": "us-west"
    }
  }'

# List targets
curl http://localhost:8080/api/v1/targets
```

### Exporter Registration

```bash
# Create exporter
curl -X POST http://localhost:8080/api/v1/exporters \
  -H "Content-Type: application/json" \
  -d '{
    "target_id": "<target-id>",
    "type": "node_exporter",
    "port": 9100,
    "enabled": true,
    "metrics_path": "/metrics",
    "scrape_interval": "15s"
  }'
```

### Service Discovery

```bash
# Prometheus HTTP SD (all targets)
curl http://localhost:8080/api/v1/sd/prometheus

# Active targets only
curl http://localhost:8080/api/v1/sd/prometheus/active

# With filters
curl "http://localhost:8080/api/v1/sd/prometheus?status=active&exporter_type=node_exporter"

# Generate File SD
curl -X POST http://localhost:8080/api/v1/sd/prometheus/file/active \
  -H "Content-Type: application/json" \
  -d '{
    "output_path": "/tmp/prometheus-targets.json"
  }'
```

---

## 6. Sample Data Generation

Script to generate sample data for testing.

### Create Script

```bash
cat > create-sample-data.sh << 'EOF'
#!/bin/bash

API_BASE="http://localhost:8080/api/v1"

echo "=== Creating sample data for AAMI Config Server ==="

# 1. Create Namespaces
echo "1. Creating namespaces..."
NS_INFRA=$(curl -s -X POST "$API_BASE/namespaces" \
  -H "Content-Type: application/json" \
  -d '{"name":"infrastructure","policy_priority":10,"description":"Infrastructure namespace"}' \
  | jq -r '.id')
echo "  - Infrastructure namespace: $NS_INFRA"

NS_LOGICAL=$(curl -s -X POST "$API_BASE/namespaces" \
  -H "Content-Type: application/json" \
  -d '{"name":"logical","policy_priority":50,"description":"Logical grouping namespace"}' \
  | jq -r '.id')
echo "  - Logical namespace: $NS_LOGICAL"

NS_ENV=$(curl -s -X POST "$API_BASE/namespaces" \
  -H "Content-Type: application/json" \
  -d '{"name":"environment","policy_priority":100,"description":"Environment namespace"}' \
  | jq -r '.id')
echo "  - Environment namespace: $NS_ENV"

# 2. Create Groups
echo "2. Creating groups..."
GROUP_PROD=$(curl -s -X POST "$API_BASE/groups" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"production\",\"namespace_id\":\"$NS_ENV\",\"priority\":100,\"description\":\"Production environment\"}" \
  | jq -r '.id')
echo "  - Production group: $GROUP_PROD"

GROUP_WEB=$(curl -s -X POST "$API_BASE/groups" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"web-servers\",\"namespace_id\":\"$NS_LOGICAL\",\"priority\":50,\"description\":\"Web servers\"}" \
  | jq -r '.id')
echo "  - Web servers group: $GROUP_WEB"

# 3. Create Targets
echo "3. Creating targets..."
TARGET_1=$(curl -s -X POST "$API_BASE/targets" \
  -H "Content-Type: application/json" \
  -d "{\"hostname\":\"web-01\",\"ip_address\":\"192.168.1.10\",\"group_ids\":[\"$GROUP_PROD\",\"$GROUP_WEB\"],\"status\":\"active\",\"labels\":{\"env\":\"production\",\"region\":\"us-west\"}}" \
  | jq -r '.id')
echo "  - Target web-01: $TARGET_1"

TARGET_2=$(curl -s -X POST "$API_BASE/targets" \
  -H "Content-Type: application/json" \
  -d "{\"hostname\":\"web-02\",\"ip_address\":\"192.168.1.11\",\"group_ids\":[\"$GROUP_PROD\",\"$GROUP_WEB\"],\"status\":\"active\",\"labels\":{\"env\":\"production\",\"region\":\"us-west\"}}" \
  | jq -r '.id')
echo "  - Target web-02: $TARGET_2"

# 4. Create Exporters
echo "4. Creating exporters..."
curl -s -X POST "$API_BASE/exporters" \
  -H "Content-Type: application/json" \
  -d "{\"target_id\":\"$TARGET_1\",\"type\":\"node_exporter\",\"port\":9100,\"enabled\":true,\"metrics_path\":\"/metrics\",\"scrape_interval\":\"15s\"}" > /dev/null
echo "  - Node exporter for web-01"

curl -s -X POST "$API_BASE/exporters" \
  -H "Content-Type: application/json" \
  -d "{\"target_id\":\"$TARGET_2\",\"type\":\"node_exporter\",\"port\":9100,\"enabled\":true,\"metrics_path\":\"/metrics\",\"scrape_interval\":\"15s\"}" > /dev/null
echo "  - Node exporter for web-02"

echo ""
echo "=== Sample data created successfully! ==="
echo ""
echo "Test with:"
echo "  curl http://localhost:8080/api/v1/targets"
echo "  curl http://localhost:8080/api/v1/sd/prometheus/active"
EOF

chmod +x create-sample-data.sh
```

### Execute

```bash
# Install jq (for JSON parsing)
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# Generate sample data
./create-sample-data.sh
```

### Verification

```bash
# List targets
curl http://localhost:8080/api/v1/targets | jq

# Check Service Discovery
curl http://localhost:8080/api/v1/sd/prometheus/active | jq
```

---

## Troubleshooting

### Docker Compose Errors

**Issue**: `Cannot connect to the Docker daemon`

```bash
# Verify Docker Desktop is running
docker ps

# Start Docker service (Linux)
sudo systemctl start docker
```

**Issue**: `port is already allocated`

```bash
# Check which process is using the port
lsof -i :8080
lsof -i :5432

# Stop conflicting process or change port
docker-compose down
```

### Database Connection Errors

**Issue**: `connection refused`

```bash
# Verify PostgreSQL is running
pg_isready -h localhost -p 5432

# Check PostgreSQL logs
docker-compose logs postgres
```

**Issue**: `database does not exist`

```bash
# Create database
docker-compose exec postgres psql -U postgres -c "CREATE DATABASE aami_config;"
```

### Migration Issues

**Note**: Database migrations are handled automatically by the migration init container when using Docker Compose. For detailed migration procedures, troubleshooting, and production deployment, see [docs/MIGRATION.md](./docs/MIGRATION.md).

**Issue**: Migration container fails or schema validation error

```bash
# Check migration container status
docker-compose ps migration

# View migration logs
docker-compose logs migration

# Verify database tables
docker-compose exec postgres psql -U admin -d config_server -c "\dt"

# Re-run migration manually if needed
docker-compose exec postgres psql -U admin -d config_server -f /migrations/001_initial_schema.sql

# Restart config-server after fixing
docker-compose restart config-server
```

**Issue**: Config server fails with "missing required tables" error

```bash
# This means migrations didn't run successfully
# Check migration container logs
docker-compose logs migration

# Run migration manually
docker-compose exec postgres psql -U admin -d config_server -f /migrations/001_initial_schema.sql
```

---

## Next Steps

1. **API Documentation**: View all APIs in Swagger UI (coming soon)
2. **Prometheus Integration**: Connect Service Discovery to Prometheus
3. **Monitoring**: Set up Grafana dashboards
4. **Production Deployment**: Deploy to Kubernetes and operate

For more details, see:
- [README.md](./README.md) - Project overview
- [AGENT.md](./.agent/docs/AGENT.md) - Architecture guide
- [MIGRATION.md](./docs/MIGRATION.md) - Database migration guide
- [k8s/README.md](./k8s/README.md) - Kubernetes deployment guide
