# Scripts

This directory contains utility scripts for deployment, maintenance, and automation.

## Directory Structure

```
scripts/
├── preflight-check.sh  # Pre-installation system validation
├── install-server.sh   # One-command Config Server installation
├── node/               # Node agent installation scripts
└── db/                 # Database management scripts
```

## Script Categories

### Preflight Check Script

- **Location**: `preflight-check.sh`
- **Purpose**: Validate system requirements before AAMI installation

The preflight check script verifies:
- System requirements (CPU, RAM, disk space, OS version)
- Software dependencies (Docker, curl, etc.)
- Network connectivity (registries, Config Server)
- Port availability (8080, 9090, 3000, etc.)
- Permissions (root/sudo access)
- Hardware detection (GPU for node mode)

Example usage:
```bash
# Server mode check
./preflight-check.sh --mode server

# Node mode check with Config Server connectivity test
./preflight-check.sh --mode node --server https://config.example.com

# Auto-fix common issues
./preflight-check.sh --mode server --fix

# JSON output for CI/CD pipelines
./preflight-check.sh --mode server --json
```

Options:
- `--mode [server|node]` - Check mode (default: auto-detect)
- `--server URL` - Config Server URL for connectivity test
- `--fix` - Attempt to automatically fix issues
- `--json` - Output results in JSON format
- `--quiet` - Only show errors
- `--verbose` - Show detailed information

### Server Installation Script

- **Location**: `install-server.sh`
- **Purpose**: One-command installation of the complete AAMI monitoring stack

The install script provides fully automated deployment of:
- PostgreSQL database
- Redis cache
- Config Server API
- Prometheus
- Grafana
- Alertmanager

Example usage:
```bash
# Interactive installation
./install-server.sh

# Unattended installation with custom settings
./install-server.sh --unattended --domain config.example.com

# One-line installation from GitHub
curl -fsSL https://raw.githubusercontent.com/fregataa/aami/main/scripts/install-server.sh | bash
```

Options:
- `--version VERSION` - AAMI version to install (default: latest)
- `--install-dir PATH` - Installation directory (default: /opt/aami)
- `--data-dir PATH` - Data directory (default: /var/lib/aami)
- `--domain DOMAIN` - Config Server domain (default: localhost)
- `--port PORT` - Config Server port (default: 8080)
- `--postgres-password PW` - PostgreSQL password (auto-generated if not set)
- `--grafana-password PW` - Grafana admin password (auto-generated if not set)
- `--skip-preflight` - Skip preflight checks (not recommended)
- `--unattended` - Non-interactive mode
- `--verbose` - Show detailed output

Installation steps performed:
1. Run preflight checks (system requirements, dependencies, ports)
2. Download AAMI from GitHub (clone or tarball)
3. Generate secure credentials
4. Configure environment (.env file)
5. Start Docker Compose stack
6. Wait for all services to become healthy
7. Create initial bootstrap token for node registration

After successful installation:
- Credentials are saved to `/opt/aami/credentials.txt`
- Services are accessible at localhost (configurable ports)
- A bootstrap token is created for registering monitoring nodes

### Node Scripts
- **Location**: `node/`
- **Purpose**: Install and configure monitoring agents on target nodes

Available scripts:
- `install-node-exporter.sh` - Install Prometheus Node Exporter
- `dynamic-check.sh` - Execute dynamic checks from Config Server
- `install-dcgm-exporter.sh` - Install NVIDIA DCGM Exporter (planned)
- `install-all-smi.sh` - Install all-smi for multi-vendor GPU support (planned)
- `bootstrap.sh` - One-line bootstrap script for auto-registration (planned)
- `uninstall.sh` - Remove all monitoring agents (planned)

Example usage:
```bash
# Install Node Exporter
curl -fsSL https://raw.githubusercontent.com/your-org/aami/main/scripts/node/install-node-exporter.sh | bash

# Or with bootstrap token
curl -fsSL https://raw.githubusercontent.com/your-org/aami/main/scripts/node/bootstrap.sh | \
  bash -s -- --token YOUR_BOOTSTRAP_TOKEN --server https://config-server.example.com
```

### Database Scripts
- **Location**: `db/`
- **Purpose**: Database migrations and maintenance

Available scripts:
- `migrate-up.sh` - Apply database migrations
- `migrate-down.sh` - Rollback migrations
- `backup.sh` - Backup PostgreSQL database
- `restore.sh` - Restore from backup
- `seed.sh` - Seed test data (development only)

Example usage:
```bash
# Run migrations
./db/migrate-up.sh

# Backup database
./db/backup.sh /path/to/backup/dir
```

## Script Conventions

All scripts follow these conventions:

1. **Shebang**: Use `#!/usr/bin/env bash` for portability
2. **Error Handling**: Set `set -euo pipefail` for safety
3. **Comments**: All comments must be in English
4. **Exit Codes**:
   - 0: Success
   - 1: General error
   - 2: Invalid arguments
   - 3: Missing dependencies
5. **Logging**: Use `echo` with prefixes:
   - `[INFO]` - Informational messages
   - `[WARN]` - Warnings
   - `[ERROR]` - Errors

Example template:
```bash
#!/usr/bin/env bash
set -euo pipefail

# Script description
readonly SCRIPT_NAME="$(basename "$0")"

# Print usage information
usage() {
    cat <<EOF
Usage: $SCRIPT_NAME [OPTIONS]

Description of what this script does.

Options:
    -h, --help     Show this help message
    -v, --verbose  Enable verbose output
EOF
}

# Main function
main() {
    echo "[INFO] Starting $SCRIPT_NAME..."
    # Script logic here
    echo "[INFO] Done!"
}

main "$@"
```

## Testing Scripts

Test scripts before deployment:

```bash
# Syntax check
shellcheck scripts/node/*.sh

# Dry run (if supported)
./scripts/node/install-node-exporter.sh --dry-run

# Test in Docker container
docker run --rm -v $(pwd)/scripts:/scripts ubuntu:22.04 bash /scripts/node/install-node-exporter.sh
```

## Security Considerations

- Always verify script sources before execution
- Use HTTPS for downloading scripts
- Validate checksums when available
- Review bootstrap tokens before distribution
- Use environment variables for sensitive data (never hardcode)

## Quick Links

- [Development Guide](../docs/DEVELOPMENT.md)
- [Deployment Guide](../docs/DEPLOYMENT.md)
- [Bootstrap Script Guide](../docs/bootstrap-and-deployment.md)
