#!/usr/bin/env bash
set -euo pipefail

# AAMI Preflight Check Script
#
# Validates system requirements before AAMI installation to prevent
# mid-install failures. Supports both server and node installation modes.
#
# Usage: ./preflight-check.sh [OPTIONS]
#
# Options:
#   --mode MODE      Check mode: 'server' or 'node' (default: auto-detect)
#   --server URL     Config Server URL (for node mode connectivity check)
#   --fix            Attempt automatic fixes for issues
#   --json           Output results in JSON format
#   --quiet          Only show errors
#   --verbose        Show detailed check information
#   -h, --help       Show this help message

readonly SCRIPT_NAME="$(basename "$0")"
readonly SCRIPT_VERSION="1.0.0"

# Configuration
MODE=""  # auto-detect by default
CONFIG_SERVER_URL=""
FIX_MODE=false
JSON_OUTPUT=false
QUIET_MODE=false
VERBOSE_MODE=false

# Server mode requirements
readonly SERVER_MIN_CPU=2
readonly SERVER_MIN_RAM_GB=4
readonly SERVER_MIN_DISK_GB=20
readonly SERVER_PORTS=(8080 9090 3000 5432 6379)

# Node mode requirements
readonly NODE_MIN_CPU=1
readonly NODE_MIN_RAM_GB=1
readonly NODE_MIN_DISK_GB=5
readonly NODE_PORTS=(9100 9400)

# Colors (disabled in JSON mode)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Check results storage
declare -A CHECK_RESULTS
declare -a ERRORS=()
declare -a WARNINGS=()
declare -a FIXES=()

# Disable colors for JSON output or non-terminal
disable_colors() {
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    CYAN=''
    BOLD=''
    NC=''
}

# Print functions
info() {
    if [[ "$QUIET_MODE" == "false" && "$JSON_OUTPUT" == "false" ]]; then
        echo -e "${GREEN}[INFO]${NC} $1"
    fi
}

warn() {
    if [[ "$JSON_OUTPUT" == "false" ]]; then
        echo -e "${YELLOW}[WARN]${NC} $1"
    fi
    WARNINGS+=("$1")
}

error() {
    if [[ "$JSON_OUTPUT" == "false" ]]; then
        echo -e "${RED}[ERROR]${NC} $1" >&2
    fi
    ERRORS+=("$1")
}

verbose() {
    if [[ "$VERBOSE_MODE" == "true" && "$JSON_OUTPUT" == "false" ]]; then
        echo -e "${CYAN}[DEBUG]${NC} $1"
    fi
}

# Print check result
print_check() {
    local status="$1"
    local message="$2"

    if [[ "$JSON_OUTPUT" == "true" ]]; then
        return
    fi

    if [[ "$status" == "pass" ]]; then
        echo -e "  ${GREEN}[✓]${NC} $message"
    elif [[ "$status" == "fail" ]]; then
        echo -e "  ${RED}[✗]${NC} $message"
    elif [[ "$status" == "warn" ]]; then
        echo -e "  ${YELLOW}[!]${NC} $message"
    elif [[ "$status" == "info" ]]; then
        echo -e "  ${BLUE}[i]${NC} $message"
    fi
}

# Print section header
print_section() {
    if [[ "$JSON_OUTPUT" == "false" && "$QUIET_MODE" == "false" ]]; then
        echo ""
        echo -e "${BOLD}$1${NC}"
    fi
}

# Print usage
usage() {
    cat <<EOF
Usage: $SCRIPT_NAME [OPTIONS]

AAMI Preflight Check - Validates system requirements before installation.

Options:
    --mode MODE      Check mode: 'server' or 'node' (default: auto-detect)
    --server URL     Config Server URL (for node mode connectivity check)
    --fix            Attempt automatic fixes for issues
    --json           Output results in JSON format
    --quiet          Only show errors
    --verbose        Show detailed check information
    -h, --help       Show this help message
    -V, --version    Show version information

Examples:
    # Basic server check
    $SCRIPT_NAME --mode server

    # Node check with connectivity test
    $SCRIPT_NAME --mode node --server https://config.example.com

    # Auto-fix issues
    $SCRIPT_NAME --mode server --fix

    # JSON output for CI/CD
    $SCRIPT_NAME --mode node --json

Exit Codes:
    0 - All checks passed
    1 - One or more checks failed
    2 - Invalid arguments
    3 - Missing dependencies for this script

EOF
    exit 0
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --mode)
                MODE="$2"
                if [[ "$MODE" != "server" && "$MODE" != "node" ]]; then
                    echo "Error: Mode must be 'server' or 'node'" >&2
                    exit 2
                fi
                shift 2
                ;;
            --server)
                CONFIG_SERVER_URL="$2"
                shift 2
                ;;
            --fix)
                FIX_MODE=true
                shift
                ;;
            --json)
                JSON_OUTPUT=true
                disable_colors
                shift
                ;;
            --quiet)
                QUIET_MODE=true
                shift
                ;;
            --verbose)
                VERBOSE_MODE=true
                shift
                ;;
            -h|--help)
                usage
                ;;
            -V|--version)
                echo "$SCRIPT_NAME version $SCRIPT_VERSION"
                exit 0
                ;;
            *)
                echo "Error: Unknown option: $1" >&2
                echo "Use --help for usage information" >&2
                exit 2
                ;;
        esac
    done
}

# Auto-detect mode based on environment
detect_mode() {
    if [[ -n "$MODE" ]]; then
        return
    fi

    # Check if Docker is available (likely server)
    if command -v docker &> /dev/null; then
        # Check if docker-compose exists
        if command -v docker-compose &> /dev/null || docker compose version &> /dev/null 2>&1; then
            MODE="server"
            verbose "Auto-detected mode: server (Docker and Docker Compose available)"
            return
        fi
    fi

    # Default to node mode
    MODE="node"
    verbose "Auto-detected mode: node"
}

# ============================================================================
# System Requirements Checks
# ============================================================================

check_os() {
    verbose "Checking operating system..."

    local os_name=""
    local os_version=""
    local supported=true

    # Check for macOS first
    if [[ "$(uname -s)" == "Darwin" ]]; then
        os_name="macos"
        os_version=$(sw_vers -productVersion 2>/dev/null || echo "unknown")
        CHECK_RESULTS[os_name]="$os_name"
        CHECK_RESULTS[os_version]="$os_version"
        CHECK_RESULTS[os_pretty]="macOS $os_version"
        CHECK_RESULTS[os_supported]="false"

        print_check "warn" "OS: macOS $os_version (development only, not for production)"
        WARNINGS+=("macOS is supported for development/testing only. Production deployment requires Linux.")
        return 0
    fi

    if [[ -f /etc/os-release ]]; then
        # shellcheck source=/dev/null
        source /etc/os-release
        os_name="$ID"
        os_version="$VERSION_ID"

        CHECK_RESULTS[os_name]="$os_name"
        CHECK_RESULTS[os_version]="$os_version"
        CHECK_RESULTS[os_pretty]="${PRETTY_NAME:-$os_name $os_version}"

        # Check supported versions
        case "$os_name" in
            ubuntu)
                if [[ "${os_version%%.*}" -lt 20 ]]; then
                    supported=false
                fi
                ;;
            debian)
                if [[ "${os_version%%.*}" -lt 11 ]]; then
                    supported=false
                fi
                ;;
            centos|rocky|rhel|almalinux)
                if [[ "${os_version%%.*}" -lt 8 ]]; then
                    supported=false
                fi
                ;;
            *)
                # Unknown OS, warn but don't fail
                print_check "warn" "OS: ${CHECK_RESULTS[os_pretty]} (not officially tested)"
                WARNINGS+=("OS $os_name is not officially tested")
                CHECK_RESULTS[os_supported]="true"
                return 0
                ;;
        esac
    else
        print_check "fail" "OS: Unable to detect (missing /etc/os-release)"
        ERRORS+=("Cannot detect operating system")
        CHECK_RESULTS[os_supported]="false"
        return 1
    fi

    if [[ "$supported" == "true" ]]; then
        print_check "pass" "OS: ${CHECK_RESULTS[os_pretty]} (supported)"
        CHECK_RESULTS[os_supported]="true"
        return 0
    else
        print_check "fail" "OS: ${CHECK_RESULTS[os_pretty]} (not supported)"
        ERRORS+=("OS version $os_name $os_version is not supported. Minimum: Ubuntu 20.04, Debian 11, CentOS/Rocky 8")
        CHECK_RESULTS[os_supported]="false"
        return 1
    fi
}

check_cpu() {
    verbose "Checking CPU cores..."

    local cpu_cores

    if [[ "$(uname -s)" == "Darwin" ]]; then
        cpu_cores=$(sysctl -n hw.ncpu 2>/dev/null || echo "0")
    else
        cpu_cores=$(nproc 2>/dev/null || grep -c ^processor /proc/cpuinfo 2>/dev/null || echo "0")
    fi

    local min_cpu
    if [[ "$MODE" == "server" ]]; then
        min_cpu=$SERVER_MIN_CPU
    else
        min_cpu=$NODE_MIN_CPU
    fi

    CHECK_RESULTS[cpu_cores]="$cpu_cores"
    CHECK_RESULTS[cpu_minimum]="$min_cpu"

    if [[ "$cpu_cores" -ge "$min_cpu" ]]; then
        print_check "pass" "CPU: $cpu_cores cores (minimum: $min_cpu)"
        CHECK_RESULTS[cpu_passed]="true"
        return 0
    else
        print_check "fail" "CPU: $cpu_cores cores (minimum: $min_cpu)"
        ERRORS+=("Insufficient CPU cores: $cpu_cores (minimum: $min_cpu)")
        CHECK_RESULTS[cpu_passed]="false"
        return 1
    fi
}

check_ram() {
    verbose "Checking RAM..."

    local ram_gb

    if [[ "$(uname -s)" == "Darwin" ]]; then
        local ram_bytes
        ram_bytes=$(sysctl -n hw.memsize 2>/dev/null || echo "0")
        ram_gb=$((ram_bytes / 1024 / 1024 / 1024))
    else
        local ram_kb
        ram_kb=$(grep MemTotal /proc/meminfo 2>/dev/null | awk '{print $2}' || echo "0")
        ram_gb=$((ram_kb / 1024 / 1024))
    fi

    local min_ram
    if [[ "$MODE" == "server" ]]; then
        min_ram=$SERVER_MIN_RAM_GB
    else
        min_ram=$NODE_MIN_RAM_GB
    fi

    CHECK_RESULTS[ram_gb]="$ram_gb"
    CHECK_RESULTS[ram_minimum]="$min_ram"

    if [[ "$ram_gb" -ge "$min_ram" ]]; then
        print_check "pass" "RAM: ${ram_gb}GB (minimum: ${min_ram}GB)"
        CHECK_RESULTS[ram_passed]="true"
        return 0
    else
        print_check "fail" "RAM: ${ram_gb}GB (minimum: ${min_ram}GB)"
        ERRORS+=("Insufficient RAM: ${ram_gb}GB (minimum: ${min_ram}GB)")
        CHECK_RESULTS[ram_passed]="false"
        return 1
    fi
}

check_disk() {
    verbose "Checking disk space..."

    local disk_available

    if [[ "$(uname -s)" == "Darwin" ]]; then
        # macOS df output is different
        disk_available=$(df -g / 2>/dev/null | awk 'NR==2 {print $4}' || echo "0")
    else
        disk_available=$(df -BG / 2>/dev/null | awk 'NR==2 {gsub(/G/,"",$4); print $4}' || echo "0")
    fi

    local min_disk
    if [[ "$MODE" == "server" ]]; then
        min_disk=$SERVER_MIN_DISK_GB
    else
        min_disk=$NODE_MIN_DISK_GB
    fi

    CHECK_RESULTS[disk_gb]="$disk_available"
    CHECK_RESULTS[disk_minimum]="$min_disk"

    if [[ "$disk_available" -ge "$min_disk" ]]; then
        print_check "pass" "Disk: ${disk_available}GB free (minimum: ${min_disk}GB)"
        CHECK_RESULTS[disk_passed]="true"
        return 0
    else
        print_check "fail" "Disk: ${disk_available}GB free (minimum: ${min_disk}GB)"
        ERRORS+=("Insufficient disk space: ${disk_available}GB (minimum: ${min_disk}GB)")
        CHECK_RESULTS[disk_passed]="false"
        return 1
    fi
}

# ============================================================================
# Software Dependency Checks
# ============================================================================

check_command() {
    local cmd="$1"
    local required="${2:-true}"
    local min_version="${3:-}"

    if command -v "$cmd" &> /dev/null; then
        local version=""
        case "$cmd" in
            docker)
                version=$(docker version --format '{{.Server.Version}}' 2>/dev/null || docker --version 2>/dev/null | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
                ;;
            curl)
                version=$(curl --version 2>/dev/null | head -1 | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
                ;;
            jq)
                version=$(jq --version 2>/dev/null | grep -Eo '[0-9]+\.[0-9]+' | head -1 || echo "unknown")
                ;;
            *)
                version=$($cmd --version 2>/dev/null | head -1 || echo "installed")
                ;;
        esac

        CHECK_RESULTS["sw_${cmd}"]="$version"
        CHECK_RESULTS["sw_${cmd}_installed"]="true"

        if [[ -n "$min_version" ]]; then
            print_check "pass" "$cmd: $version (minimum: $min_version)"
        else
            print_check "pass" "$cmd: $version"
        fi
        return 0
    else
        CHECK_RESULTS["sw_${cmd}_installed"]="false"

        if [[ "$required" == "true" ]]; then
            print_check "fail" "$cmd: not installed (required)"
            ERRORS+=("$cmd is required but not installed")
            FIXES+=("Install $cmd")
            return 1
        else
            print_check "warn" "$cmd: not installed (optional)"
            WARNINGS+=("$cmd is not installed (optional)")
            return 0
        fi
    fi
}

check_docker() {
    verbose "Checking Docker..."

    if ! check_command "docker" "true" "20.10"; then
        if [[ "$FIX_MODE" == "true" ]]; then
            info "Attempting to install Docker..."
            # This is a simplified version - real implementation would be more robust
            if command -v apt-get &> /dev/null; then
                apt-get update && apt-get install -y docker.io
            elif command -v yum &> /dev/null; then
                yum install -y docker
            fi
        fi
        return 1
    fi

    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        print_check "fail" "Docker daemon: not running"
        ERRORS+=("Docker daemon is not running")

        if [[ "$FIX_MODE" == "true" ]]; then
            info "Attempting to start Docker..."
            systemctl start docker 2>/dev/null || service docker start 2>/dev/null || true
        else
            FIXES+=("Start Docker: sudo systemctl start docker")
        fi
        return 1
    else
        print_check "pass" "Docker daemon: running"
    fi

    return 0
}

check_docker_compose() {
    verbose "Checking Docker Compose..."

    # Check for docker compose (v2 plugin)
    if docker compose version &> /dev/null 2>&1; then
        local version
        version=$(docker compose version --short 2>/dev/null || echo "unknown")
        print_check "pass" "Docker Compose (plugin): $version"
        CHECK_RESULTS[docker_compose_version]="$version"
        CHECK_RESULTS[docker_compose_type]="plugin"
        return 0
    fi

    # Check for docker-compose (standalone)
    if command -v docker-compose &> /dev/null; then
        local version
        version=$(docker-compose version --short 2>/dev/null || echo "unknown")
        print_check "pass" "Docker Compose (standalone): $version"
        CHECK_RESULTS[docker_compose_version]="$version"
        CHECK_RESULTS[docker_compose_type]="standalone"
        return 0
    fi

    print_check "fail" "Docker Compose: not installed"
    ERRORS+=("Docker Compose is required but not installed")
    FIXES+=("Install Docker Compose: https://docs.docker.com/compose/install/")
    return 1
}

check_software_server() {
    print_section "Software Dependencies"

    local failed=0

    check_docker || ((failed++))
    check_docker_compose || ((failed++))
    check_command "curl" "true" || ((failed++))
    check_command "jq" "false" || true  # Optional

    return "$failed"
}

check_software_node() {
    print_section "Software Dependencies"

    local failed=0

    check_command "curl" "true" || ((failed++))
    check_command "systemctl" "true" || ((failed++))
    check_command "tar" "true" || ((failed++))
    check_command "jq" "false" || true  # Optional

    return "$failed"
}

# ============================================================================
# Network Connectivity Checks
# ============================================================================

check_dns() {
    verbose "Checking DNS resolution..."

    if host google.com &> /dev/null || nslookup google.com &> /dev/null || ping -c 1 google.com &> /dev/null; then
        print_check "pass" "DNS: working"
        CHECK_RESULTS[dns_working]="true"
        return 0
    else
        print_check "fail" "DNS: not working"
        ERRORS+=("DNS resolution is not working")
        CHECK_RESULTS[dns_working]="false"
        return 1
    fi
}

check_registry() {
    local registry="$1"
    local timeout="${2:-5}"

    verbose "Checking connectivity to $registry..."

    if curl -sSf --connect-timeout "$timeout" "https://$registry" &> /dev/null || \
       curl -sSf --connect-timeout "$timeout" "https://$registry/v2/" &> /dev/null 2>&1; then
        print_check "pass" "$registry: reachable"
        CHECK_RESULTS["net_${registry//\./_}"]="true"
        return 0
    else
        print_check "fail" "$registry: not reachable"
        ERRORS+=("Cannot reach $registry - check internet connectivity or firewall")
        CHECK_RESULTS["net_${registry//\./_}"]="false"
        return 1
    fi
}

check_config_server() {
    if [[ -z "$CONFIG_SERVER_URL" ]]; then
        verbose "Skipping Config Server check (no URL provided)"
        return 0
    fi

    verbose "Checking Config Server connectivity..."

    local health_url="${CONFIG_SERVER_URL}/api/v1/health"

    if curl -sSf --connect-timeout 5 "$health_url" &> /dev/null; then
        print_check "pass" "Config Server: reachable ($CONFIG_SERVER_URL)"
        CHECK_RESULTS[config_server_reachable]="true"
        return 0
    else
        print_check "fail" "Config Server: not reachable ($CONFIG_SERVER_URL)"
        ERRORS+=("Cannot reach Config Server at $CONFIG_SERVER_URL")
        CHECK_RESULTS[config_server_reachable]="false"
        return 1
    fi
}

check_network_server() {
    print_section "Network Connectivity"

    local failed=0

    check_dns || ((failed++))
    check_registry "docker.io" || ((failed++))
    check_registry "ghcr.io" || true  # Optional, don't fail

    return "$failed"
}

check_network_node() {
    print_section "Network Connectivity"

    local failed=0

    check_dns || ((failed++))
    check_config_server || ((failed++))
    check_registry "github.com" || true  # For downloading exporters

    return "$failed"
}

# ============================================================================
# Port Availability Checks
# ============================================================================

check_port() {
    local port="$1"

    verbose "Checking port $port..."

    local in_use=false
    local process_info=""

    # Check if port is in use
    if command -v ss &> /dev/null; then
        if ss -tuln 2>/dev/null | grep -q ":${port} "; then
            in_use=true
            process_info=$(ss -tulnp 2>/dev/null | grep ":${port} " | awk '{print $NF}' | head -1 || echo "unknown")
        fi
    elif command -v netstat &> /dev/null; then
        if netstat -tuln 2>/dev/null | grep -q ":${port} "; then
            in_use=true
            process_info=$(netstat -tulnp 2>/dev/null | grep ":${port} " | awk '{print $NF}' | head -1 || echo "unknown")
        fi
    else
        # Fallback: try to bind to the port
        if ! (echo >/dev/tcp/localhost/"$port") 2>/dev/null; then
            in_use=false
        else
            in_use=true
        fi
    fi

    CHECK_RESULTS["port_${port}_available"]="$([[ "$in_use" == "false" ]] && echo "true" || echo "false")"

    if [[ "$in_use" == "false" ]]; then
        print_check "pass" "Port $port: available"
        return 0
    else
        # Extract process name from process_info
        local process_name
        process_name=$(echo "$process_info" | grep -oP '(?<=\(")[^"]+' || echo "$process_info")

        print_check "fail" "Port $port: in use${process_name:+ by $process_name}"
        ERRORS+=("Port $port is already in use${process_name:+ by $process_name}")
        CHECK_RESULTS["port_${port}_process"]="$process_name"

        if [[ "$FIX_MODE" == "true" && -n "$process_name" ]]; then
            info "To free port $port, stop the service using it"
        fi

        return 1
    fi
}

check_ports_server() {
    print_section "Port Availability"

    local failed=0

    for port in "${SERVER_PORTS[@]}"; do
        check_port "$port" || ((failed++))
    done

    return "$failed"
}

check_ports_node() {
    print_section "Port Availability"

    local failed=0

    for port in "${NODE_PORTS[@]}"; do
        check_port "$port" || ((failed++))
    done

    return "$failed"
}

# ============================================================================
# Permission Checks
# ============================================================================

check_root_or_sudo() {
    verbose "Checking root/sudo access..."

    if [[ $EUID -eq 0 ]]; then
        print_check "pass" "Running as root"
        CHECK_RESULTS[is_root]="true"
        CHECK_RESULTS[has_sudo]="true"
        return 0
    fi

    if sudo -n true 2>/dev/null; then
        print_check "pass" "sudo: available (passwordless)"
        CHECK_RESULTS[is_root]="false"
        CHECK_RESULTS[has_sudo]="true"
        return 0
    fi

    if sudo -v 2>/dev/null; then
        print_check "pass" "sudo: available"
        CHECK_RESULTS[is_root]="false"
        CHECK_RESULTS[has_sudo]="true"
        return 0
    fi

    print_check "fail" "sudo: not available"
    ERRORS+=("Root or sudo access is required for installation")
    CHECK_RESULTS[is_root]="false"
    CHECK_RESULTS[has_sudo]="false"
    return 1
}

check_docker_socket() {
    verbose "Checking Docker socket access..."

    if [[ ! -S /var/run/docker.sock ]]; then
        print_check "warn" "Docker socket: not found"
        WARNINGS+=("Docker socket not found at /var/run/docker.sock")
        return 0
    fi

    if docker info &> /dev/null; then
        print_check "pass" "Docker socket: accessible"
        CHECK_RESULTS[docker_socket_accessible]="true"
        return 0
    else
        print_check "fail" "Docker socket: not accessible"
        ERRORS+=("Cannot access Docker socket. Add user to docker group: sudo usermod -aG docker \$USER")
        CHECK_RESULTS[docker_socket_accessible]="false"

        if [[ "$FIX_MODE" == "true" ]]; then
            info "Attempting to add current user to docker group..."
            sudo usermod -aG docker "$USER" 2>/dev/null || true
            warn "Please log out and back in for group changes to take effect"
        fi

        return 1
    fi
}

check_permissions_server() {
    print_section "Permissions"

    local failed=0

    check_root_or_sudo || ((failed++))
    check_docker_socket || ((failed++))

    return "$failed"
}

check_permissions_node() {
    print_section "Permissions"

    local failed=0

    check_root_or_sudo || ((failed++))

    # Check write access to key directories
    if [[ -w /etc/systemd/system ]] || sudo test -w /etc/systemd/system 2>/dev/null; then
        print_check "pass" "/etc/systemd/system: writable"
    else
        print_check "fail" "/etc/systemd/system: not writable"
        ((failed++))
    fi

    return "$failed"
}

# ============================================================================
# Hardware Detection (Node mode)
# ============================================================================

check_hardware() {
    print_section "Hardware Detection"

    # NVIDIA GPU
    if command -v nvidia-smi &> /dev/null; then
        local gpu_count
        local gpu_model
        gpu_count=$(nvidia-smi --query-gpu=count --format=csv,noheader 2>/dev/null | head -1 || echo "0")
        gpu_model=$(nvidia-smi --query-gpu=name --format=csv,noheader 2>/dev/null | head -1 || echo "unknown")

        print_check "info" "NVIDIA GPU detected: ${gpu_count}x ${gpu_model}"
        CHECK_RESULTS[gpu_vendor]="nvidia"
        CHECK_RESULTS[gpu_count]="$gpu_count"
        CHECK_RESULTS[gpu_model]="$gpu_model"

        # Check NVIDIA driver version
        local driver_version
        driver_version=$(nvidia-smi --query-gpu=driver_version --format=csv,noheader 2>/dev/null | head -1 || echo "unknown")
        print_check "info" "NVIDIA Driver: $driver_version"
        CHECK_RESULTS[nvidia_driver]="$driver_version"
    else
        verbose "No NVIDIA GPU detected"
        CHECK_RESULTS[gpu_vendor]="none"
    fi

    # AMD GPU
    if command -v rocm-smi &> /dev/null; then
        print_check "info" "AMD GPU detected (ROCm available)"
        CHECK_RESULTS[amd_gpu]="true"
    fi

    # InfiniBand
    if command -v ibstat &> /dev/null; then
        local ib_devices
        ib_devices=$(ibstat -l 2>/dev/null | wc -l || echo "0")
        if [[ "$ib_devices" -gt 0 ]]; then
            print_check "info" "InfiniBand: $ib_devices device(s) detected"
            CHECK_RESULTS[infiniband]="$ib_devices"
        fi
    fi

    return 0
}

# ============================================================================
# JSON Output
# ============================================================================

output_json() {
    local passed="$1"

    # Build JSON manually to avoid jq dependency
    cat <<EOF
{
  "version": "$SCRIPT_VERSION",
  "mode": "$MODE",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "passed": $passed,
  "checks": {
    "system": {
      "os": "${CHECK_RESULTS[os_pretty]:-unknown}",
      "os_supported": ${CHECK_RESULTS[os_supported]:-false},
      "cpu_cores": ${CHECK_RESULTS[cpu_cores]:-0},
      "cpu_minimum": ${CHECK_RESULTS[cpu_minimum]:-0},
      "cpu_passed": ${CHECK_RESULTS[cpu_passed]:-false},
      "ram_gb": ${CHECK_RESULTS[ram_gb]:-0},
      "ram_minimum": ${CHECK_RESULTS[ram_minimum]:-0},
      "ram_passed": ${CHECK_RESULTS[ram_passed]:-false},
      "disk_gb": ${CHECK_RESULTS[disk_gb]:-0},
      "disk_minimum": ${CHECK_RESULTS[disk_minimum]:-0},
      "disk_passed": ${CHECK_RESULTS[disk_passed]:-false}
    },
    "software": {
      "docker": "${CHECK_RESULTS[sw_docker]:-not installed}",
      "docker_compose": "${CHECK_RESULTS[docker_compose_version]:-not installed}",
      "curl": "${CHECK_RESULTS[sw_curl]:-not installed}"
    },
    "network": {
      "dns": ${CHECK_RESULTS[dns_working]:-false},
      "docker_io": ${CHECK_RESULTS[net_docker_io]:-false},
      "config_server": ${CHECK_RESULTS[config_server_reachable]:-null}
    },
    "permissions": {
      "root": ${CHECK_RESULTS[is_root]:-false},
      "sudo": ${CHECK_RESULTS[has_sudo]:-false},
      "docker_socket": ${CHECK_RESULTS[docker_socket_accessible]:-null}
    }
  },
  "errors": [$(printf '"%s",' "${ERRORS[@]}" | sed 's/,$//')],
  "warnings": [$(printf '"%s",' "${WARNINGS[@]}" | sed 's/,$//')]
}
EOF
}

# ============================================================================
# Main
# ============================================================================

print_header() {
    if [[ "$JSON_OUTPUT" == "true" || "$QUIET_MODE" == "true" ]]; then
        return
    fi

    echo ""
    echo -e "${BOLD}AAMI Preflight Check v${SCRIPT_VERSION}${NC}"
    echo "========================="
    echo ""
    echo -e "Mode: ${CYAN}${MODE^} Installation${NC}"
}

print_summary() {
    local total_errors=${#ERRORS[@]}
    local total_warnings=${#WARNINGS[@]}

    if [[ "$JSON_OUTPUT" == "true" ]]; then
        return
    fi

    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if [[ $total_errors -eq 0 ]]; then
        echo -e "${GREEN}Result: All checks passed!${NC}"
    else
        echo -e "${RED}Result: $total_errors issue(s) found${NC}"
        echo ""
        echo "ERRORS:"
        for err in "${ERRORS[@]}"; do
            echo -e "  ${RED}[✗]${NC} $err"
        done
    fi

    if [[ $total_warnings -gt 0 ]]; then
        echo ""
        echo "WARNINGS:"
        for warn in "${WARNINGS[@]}"; do
            echo -e "  ${YELLOW}[!]${NC} $warn"
        done
    fi

    if [[ "${#FIXES[@]}" -gt 0 && "$FIX_MODE" == "false" ]]; then
        echo ""
        echo "SUGGESTED FIXES:"
        for fix in "${FIXES[@]:-}"; do
            echo -e "  ${BLUE}→${NC} $fix"
        done
        echo ""
        echo "Run with --fix to attempt automatic fixes."
    fi

    echo ""
}

main() {
    parse_args "$@"

    # Check if running in a terminal
    if [[ ! -t 1 ]]; then
        disable_colors
    fi

    detect_mode
    print_header

    # Temporarily disable exit on error for checks
    set +e

    local total_failures=0

    # Run checks based on mode
    print_section "System Requirements"
    check_os; [[ $? -ne 0 ]] && ((total_failures++)) || true
    check_cpu; [[ $? -ne 0 ]] && ((total_failures++)) || true
    check_ram; [[ $? -ne 0 ]] && ((total_failures++)) || true
    check_disk; [[ $? -ne 0 ]] && ((total_failures++)) || true

    if [[ "$MODE" == "server" ]]; then
        check_software_server; [[ $? -ne 0 ]] && ((total_failures++)) || true
        check_network_server; [[ $? -ne 0 ]] && ((total_failures++)) || true
        check_ports_server; [[ $? -ne 0 ]] && ((total_failures++)) || true
        check_permissions_server; [[ $? -ne 0 ]] && ((total_failures++)) || true
    else
        check_software_node; [[ $? -ne 0 ]] && ((total_failures++)) || true
        check_network_node; [[ $? -ne 0 ]] && ((total_failures++)) || true
        check_ports_node; [[ $? -ne 0 ]] && ((total_failures++)) || true
        check_permissions_node; [[ $? -ne 0 ]] && ((total_failures++)) || true
        check_hardware || true
    fi

    # Re-enable exit on error
    set -e

    # Output results
    if [[ "$JSON_OUTPUT" == "true" ]]; then
        if [[ ${#ERRORS[@]} -eq 0 ]]; then
            output_json "true"
        else
            output_json "false"
        fi
    else
        print_summary
    fi

    # Exit with appropriate code
    if [[ ${#ERRORS[@]} -eq 0 ]]; then
        exit 0
    else
        exit 1
    fi
}

main "$@"
