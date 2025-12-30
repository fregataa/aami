# Scripts

This directory contains utility scripts for deployment, maintenance, and automation.

## Directory Structure

```
scripts/
├── preflight_check.py  # Pre-installation system validation (Python)
├── preflight-check.sh  # Pre-installation system validation (Bash, legacy)
├── install-server.sh   # One-command Config Server installation
├── node/               # Node agent installation scripts
├── systemd/            # Systemd service files
└── db/                 # Database management scripts
```

## Script Categories

### Preflight Check Script

- **Location**: `preflight_check.py` (recommended) or `preflight-check.sh` (legacy)
- **Purpose**: Validate system requirements before AAMI installation

Two versions are available:
- **Python version** (`preflight_check.py`): Recommended. Better structured, easier to test, cleaner JSON output.
- **Bash version** (`preflight-check.sh`): Legacy. Use when Python 3 is not available.

The preflight check script verifies:
- System requirements (CPU, RAM, disk space, OS version)
- Software dependencies (Docker, curl, etc.)
- Network connectivity (registries, Config Server)
- Port availability (8080, 9090, 3000, etc.)
- Permissions (root/sudo access)
- Hardware detection (GPU for node mode)

Example usage:
```bash
# Server mode check (Python)
python3 preflight_check.py --mode server

# Node mode check with Config Server connectivity test
python3 preflight_check.py --mode node --server https://config.example.com

# Auto-fix common issues
python3 preflight_check.py --mode server --fix

# JSON output for CI/CD pipelines
python3 preflight_check.py --mode node --json

# Legacy bash version (same options)
./preflight-check.sh --mode server
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
- `bootstrap.sh` - One-line bootstrap script for auto-registration
- `install-node-exporter.sh` - Install Prometheus Node Exporter
- `install_all_smi.py` - Install all-smi multi-vendor AI accelerator exporter (Python, recommended)
- `install-all-smi.sh` - Install all-smi multi-vendor AI accelerator exporter (Bash, legacy)
- `dynamic_check.py` - Execute dynamic checks from Config Server (Python, recommended)
- `dynamic-check.sh` - Execute dynamic checks from Config Server (Bash, legacy)
- `install-dcgm-exporter.sh` - Install NVIDIA DCGM Exporter (planned)
- `uninstall.sh` - Remove all monitoring agents (planned)

#### Bootstrap Script (`bootstrap.sh`)

The bootstrap script automates the entire node registration process:
1. Run preflight checks (system requirements, connectivity)
2. Detect system information (hostname, IP, OS, GPU)
3. Validate bootstrap token with Config Server
4. Install Node Exporter
5. Install dynamic-check.sh and register cron job
6. Register node with Config Server API
7. Verify registration success

Example usage:
```bash
# Basic usage (requires sudo)
sudo ./bootstrap.sh --token aami_bootstrap_xxx --server http://config-server:8080

# One-liner from Config Server
curl -fsSL http://config-server:8080/bootstrap.sh | \
  sudo bash -s -- --token aami_bootstrap_xxx --server http://config-server:8080

# With custom labels
sudo ./bootstrap.sh --token aami_xxx --server http://config-server:8080 \
  --labels env=production --labels rack=A1

# Dry run to preview actions
sudo ./bootstrap.sh --token aami_xxx --server http://config-server:8080 --dry-run
```

Options:
- `--token TOKEN` - Bootstrap token (required)
- `--server URL` - Config Server URL (required)
- `--port PORT` - Node Exporter port (default: 9100)
- `--labels KEY=VALUE` - Additional labels (repeatable)
- `--group-id ID` - Assign to specific group (default: self group)
- `--dry-run` - Preview without executing
- `--skip-preflight` - Skip preflight checks
- `--skip-gpu` - Skip GPU detection
- `--install-all-smi` - Install all-smi multi-vendor GPU exporter (port: 9401)
- `--verbose` - Enable verbose output

#### Install Node Exporter (`install-node-exporter.sh`)

Standalone script to install Node Exporter:
```bash
# Install with defaults
sudo ./install-node-exporter.sh

# Custom port and version
sudo ./install-node-exporter.sh --version 1.7.0 --port 9100
```

#### Install all-smi (`install_all_smi.py`)

Standalone script to install all-smi multi-vendor AI accelerator exporter. Two versions available:
- **Python version** (`install_all_smi.py`): Recommended. Better error handling, cleaner code structure.
- **Bash version** (`install-all-smi.sh`): Legacy. Use when Python 3 is not available.

```bash
# Python version (recommended)
sudo python3 ./install_all_smi.py

# Custom port and version
sudo python3 ./install_all_smi.py --version 0.5.0 --port 9401

# Bash version (legacy)
sudo ./install-all-smi.sh --version 0.5.0 --port 9401
```

Supported accelerators:
- NVIDIA GPUs (CUDA)
- AMD GPUs (ROCm)
- Intel Gaudi NPUs
- Google Cloud TPUs
- Apple Silicon GPUs
- Tenstorrent, Rebellions, Furiosa NPUs

Options:
- `-v, --version VERSION` - all-smi version (default: 0.5.0)
- `-p, --port PORT` - Listen port (default: 9401)
- `-h, --help` - Show help message

#### Dynamic Check Script (`dynamic_check.py`)

Fetches and executes dynamic checks from Config Server. Two versions available:
- **Python version** (`dynamic_check.py`): Recommended. No jq dependency, better error handling.
- **Bash version** (`dynamic-check.sh`): Legacy. Requires jq.

```bash
# Manual execution
python3 dynamic_check.py --config-server http://config-server:8080

# Debug mode
python3 dynamic_check.py --config-server http://config-server:8080 --debug

# Typically runs via cron (installed by bootstrap.sh)
* * * * * /usr/local/bin/dynamic_check.py
```

Options:
- `-c, --config-server URL` - Config Server URL
- `--hostname NAME` - Override hostname
- `-d, --debug` - Enable debug logging
- `--textfile-dir PATH` - Textfile collector directory
- `--check-scripts-dir PATH` - Check scripts directory

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
