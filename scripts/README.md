# Scripts

This directory contains utility scripts for deployment, maintenance, and automation.

## Directory Structure

```
scripts/
├── node/               # Node agent installation scripts
└── db/                 # Database management scripts
```

## Script Categories

### Node Scripts
- **Location**: `node/`
- **Purpose**: Install and configure monitoring agents on target nodes

Available scripts:
- `install-node-exporter.sh` - Install Prometheus Node Exporter
- `install-dcgm-exporter.sh` - Install NVIDIA DCGM Exporter
- `install-all-smi.sh` - Install all-smi for multi-vendor GPU support
- `bootstrap.sh` - One-line bootstrap script for auto-registration
- `uninstall.sh` - Remove all monitoring agents

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
