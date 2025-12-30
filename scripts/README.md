# Scripts

This directory contains utility scripts for deployment, maintenance, and automation.

## Directory Structure

```
scripts/
├── preflight-check.sh  # Pre-installation system validation
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
