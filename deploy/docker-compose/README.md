# AAMI Docker Compose Deployment

This directory contains Docker Compose configuration for deploying the complete AAMI monitoring stack.

## Services

The stack includes the following services:

| Service | Port | Description |
|---------|------|-------------|
| **Prometheus** | 9090 | Time-series database and monitoring system |
| **Grafana** | 3000 | Visualization and dashboarding |
| **Alertmanager** | 9093 | Alert management and routing |
| **PostgreSQL** | 5432 | Metadata storage for Config Server |
| **Redis** | 6379 | Caching layer for Config Server |

## Prerequisites

- Docker 20.10+
- Docker Compose v2.0+
- At least 4GB RAM
- 20GB free disk space

## Quick Start

### 1. Configure Environment

```bash
# Copy example environment file
cp .env.example .env

# Edit environment variables
vim .env
```

### 2. Start the Stack

```bash
# Start all services
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f
```

### 3. Access Services

Once started, access the services at:

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000
  - Default credentials: `admin` / `admin`
  - You'll be prompted to change the password on first login
- **Alertmanager**: http://localhost:9093

### 4. Verify Installation

```bash
# Check Prometheus health
curl http://localhost:9090/-/healthy

# Check Grafana health
curl http://localhost:3000/api/health

# Check Alertmanager health
curl http://localhost:9093/-/healthy
```

## Configuration

### Environment Variables

Edit `.env` file to customize your deployment:

```env
# Ports
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
ALERTMANAGER_PORT=9093
POSTGRES_PORT=5432
REDIS_PORT=6379

# Grafana credentials
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=changeme

# PostgreSQL credentials
POSTGRES_DB=config_server
POSTGRES_USER=admin
POSTGRES_PASSWORD=changeme
```

### Prometheus Configuration

Prometheus configuration is located at:
- **Main config**: `../../config/prometheus/prometheus.yml`
- **Alert rules**: `../../config/prometheus/rules/system-alerts.yml`
- **Service discovery**: `../../config/prometheus/sd/*.json`

To reload Prometheus configuration:
```bash
docker-compose restart prometheus
# Or send SIGHUP signal
docker-compose kill -s SIGHUP prometheus
```

### Grafana Configuration

Grafana is pre-configured with:
- Prometheus datasource (auto-provisioned)
- System metrics dashboard (auto-imported)

Additional dashboards can be added to:
- `../../config/grafana/dashboards/`

### Alertmanager Configuration

Edit `../../config/alertmanager/alertmanager.yml` to configure:
- Email notifications (SMTP settings)
- Webhook integrations (Slack, PagerDuty, etc.)
- Alert routing rules
- Inhibition rules

## Managing the Stack

### Start Services

```bash
# Start all services
docker-compose up -d

# Start specific service
docker-compose up -d prometheus

# Start and follow logs
docker-compose up
```

### Stop Services

```bash
# Stop all services
docker-compose stop

# Stop specific service
docker-compose stop prometheus
```

### Restart Services

```bash
# Restart all services
docker-compose restart

# Restart specific service
docker-compose restart prometheus
```

### View Logs

```bash
# View all logs
docker-compose logs

# Follow logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f prometheus

# View last N lines
docker-compose logs --tail=100 prometheus
```

### Remove Stack

```bash
# Stop and remove containers
docker-compose down

# Remove containers and volumes (WARNING: data loss!)
docker-compose down -v

# Remove everything including images
docker-compose down -v --rmi all
```

## Data Persistence

Data is persisted in Docker volumes:

- `prometheus-data` - Prometheus time-series database
- `grafana-data` - Grafana dashboards and settings
- `alertmanager-data` - Alertmanager state and silences
- `postgres-data` - PostgreSQL database files
- `redis-data` - Redis persistence files

### Backup Volumes

```bash
# Backup Prometheus data
docker run --rm -v aami_prometheus-data:/data -v $(pwd):/backup \
  alpine tar czf /backup/prometheus-backup.tar.gz /data

# Backup PostgreSQL data
docker-compose exec postgres pg_dump -U admin config_server > backup.sql
```

### Restore Volumes

```bash
# Restore Prometheus data
docker run --rm -v aami_prometheus-data:/data -v $(pwd):/backup \
  alpine tar xzf /backup/prometheus-backup.tar.gz -C /data

# Restore PostgreSQL data
docker-compose exec -T postgres psql -U admin config_server < backup.sql
```

## Monitoring Targets

### Add Monitoring Targets

To monitor servers with Node Exporter:

1. Install Node Exporter on target server:
```bash
curl -fsSL https://raw.githubusercontent.com/your-org/aami/main/scripts/node/install-node-exporter.sh | sudo bash
```

2. Add target to service discovery file:
```bash
# Edit config/prometheus/sd/node-exporter.json
[
  {
    "targets": ["10.0.1.10:9100", "10.0.1.11:9100"],
    "labels": {
      "job": "node-exporter",
      "env": "production",
      "datacenter": "dc1"
    }
  }
]
```

3. Prometheus will automatically discover new targets within 30 seconds.

### Verify Targets

Check targets in Prometheus:
- UI: http://localhost:9090/targets
- API: `curl http://localhost:9090/api/v1/targets`

## Troubleshooting

### Services Won't Start

Check logs for errors:
```bash
docker-compose logs SERVICE_NAME
```

Common issues:
- **Port conflict**: Another service is using the port. Change port in `.env`
- **Permission denied**: Check file permissions on config directories
- **Out of memory**: Increase Docker memory limit

### Cannot Access Services

1. Check if containers are running:
```bash
docker-compose ps
```

2. Check port bindings:
```bash
docker-compose port prometheus 9090
```

3. Test connectivity:
```bash
curl http://localhost:9090/-/healthy
```

### Prometheus Not Scraping Targets

1. Check Prometheus logs:
```bash
docker-compose logs -f prometheus
```

2. Verify service discovery files exist and are valid JSON:
```bash
cat ../../config/prometheus/sd/node-exporter.json | jq .
```

3. Check target connectivity from Prometheus container:
```bash
docker-compose exec prometheus wget -O- http://TARGET_IP:9100/metrics
```

### Grafana Dashboards Not Loading

1. Check Grafana logs:
```bash
docker-compose logs -f grafana
```

2. Verify datasource connection:
   - Go to Configuration → Data Sources
   - Click "Test" on Prometheus datasource

3. Re-import dashboard manually:
   - Go to Dashboards → Import
   - Upload `../../config/grafana/dashboards/system-metrics.json`

### Alertmanager Not Sending Alerts

1. Check Alertmanager configuration:
```bash
docker-compose exec alertmanager amtool check-config /etc/alertmanager/alertmanager.yml
```

2. Send test alert:
```bash
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{"labels":{"alertname":"TestAlert","severity":"warning"},"annotations":{"summary":"Test alert"}}]'
```

3. Check SMTP settings in `../../config/alertmanager/alertmanager.yml`

## Upgrading

### Upgrade Services

1. Update image versions in `docker-compose.yml`
2. Pull new images:
```bash
docker-compose pull
```
3. Recreate containers:
```bash
docker-compose up -d
```

### Upgrade Prometheus

Prometheus upgrades are typically backward compatible:
```bash
# Change image version in docker-compose.yml
image: prom/prometheus:v2.46.0

# Pull and restart
docker-compose pull prometheus
docker-compose up -d prometheus
```

## Performance Tuning

### Prometheus

Adjust retention and storage:
```yaml
# In docker-compose.yml
command:
  - '--storage.tsdb.retention.time=30d'
  - '--storage.tsdb.retention.size=50GB'
```

### PostgreSQL

Increase shared buffers for better performance:
```yaml
# In docker-compose.yml
environment:
  - POSTGRES_SHARED_BUFFERS=256MB
  - POSTGRES_MAX_CONNECTIONS=200
```

### Redis

Enable persistence and increase memory:
```yaml
# In docker-compose.yml
command: redis-server --maxmemory 512mb --maxmemory-policy allkeys-lru
```

## Security Considerations

1. **Change default passwords** in `.env`
2. **Enable TLS/SSL** for production deployments
3. **Restrict network access** using Docker networks
4. **Use secrets management** for sensitive data
5. **Regular backups** of persistent volumes
6. **Keep images updated** with security patches

## Next Steps

- [Configure targets](../../docs/en/QUICKSTART.md#registering-targets)
- [Create custom dashboards](../../docs/en/QUICKSTART.md#viewing-metrics)
- [Set up alert rules](../../docs/en/QUICKSTART.md#setting-up-alerts)
- [Deploy Config Server](../../services/config-server/README.md)

## Additional Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Alertmanager Documentation](https://prometheus.io/docs/alerting/latest/alertmanager/)

---

For issues or questions, see the [main README](../../README.md) or open an issue on GitHub.
