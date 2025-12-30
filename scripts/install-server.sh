#!/usr/bin/env bash
#
# AAMI Config Server Installation Script
# One-command installation of the complete AAMI monitoring stack.
#
# Usage:
#   curl -fsSL https://aami.io/install.sh | bash
#   ./install-server.sh [OPTIONS]
#
# Options:
#   --version VERSION      AAMI version to install (default: latest)
#   --install-dir PATH     Installation directory (default: /opt/aami)
#   --data-dir PATH        Data directory (default: /var/lib/aami)
#   --domain DOMAIN        Domain for Config Server (default: localhost)
#   --port PORT            Config Server port (default: 8080)
#   --postgres-password PW PostgreSQL password (auto-generated if not set)
#   --grafana-password PW  Grafana admin password (auto-generated if not set)
#   --skip-preflight       Skip preflight checks (not recommended)
#   --unattended           Non-interactive mode
#   --verbose              Show detailed output
#   --help                 Show this help message
#
# Examples:
#   # Interactive installation
#   ./install-server.sh
#
#   # Unattended installation with custom settings
#   ./install-server.sh --unattended --domain config.example.com --install-dir /opt/aami
#
#   # Specify version
#   ./install-server.sh --version v1.0.0
#

set -uo pipefail

# ==============================================================================
# Constants
# ==============================================================================

readonly SCRIPT_NAME="AAMI Server Installer"
readonly SCRIPT_VERSION="1.0.0"
readonly GITHUB_REPO="fregataa/aami"
readonly GITHUB_URL="https://github.com/${GITHUB_REPO}"
readonly GITHUB_RAW="https://raw.githubusercontent.com/${GITHUB_REPO}"

# Default values
DEFAULT_INSTALL_DIR="/opt/aami"
DEFAULT_DATA_DIR="/var/lib/aami"
DEFAULT_DOMAIN="localhost"
DEFAULT_PORT="8080"
DEFAULT_VERSION="latest"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Installation settings (will be set by arguments or prompts)
INSTALL_DIR=""
DATA_DIR=""
DOMAIN=""
PORT=""
VERSION=""
POSTGRES_PASSWORD=""
GRAFANA_PASSWORD=""
SKIP_PREFLIGHT=false
UNATTENDED=false
VERBOSE=false

# Generated values
BOOTSTRAP_TOKEN=""

# ==============================================================================
# Utility Functions
# ==============================================================================

print_header() {
    echo -e "${BOLD}${BLUE}"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                 AAMI Server Installation                     │"
    echo "│                     Version ${SCRIPT_VERSION}                              │"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo -e "${NC}"
}

print_step() {
    local step=$1
    local total=$2
    local message=$3
    echo -e "\n${BOLD}${CYAN}[${step}/${total}]${NC} ${message}"
}

print_substep() {
    local status=$1
    local message=$2
    if [[ "$status" == "ok" ]]; then
        echo -e "       ${GREEN}✓${NC} ${message}"
    elif [[ "$status" == "warn" ]]; then
        echo -e "       ${YELLOW}⚠${NC} ${message}"
    elif [[ "$status" == "fail" ]]; then
        echo -e "       ${RED}✗${NC} ${message}"
    elif [[ "$status" == "info" ]]; then
        echo -e "       ${BLUE}→${NC} ${message}"
    else
        echo -e "         ${message}"
    fi
}

print_error() {
    echo -e "${RED}ERROR:${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}WARNING:${NC} $1"
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_verbose() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# Generate a secure random password
generate_password() {
    local length=${1:-32}
    if command -v openssl &>/dev/null; then
        openssl rand -base64 48 | tr -dc 'a-zA-Z0-9' | head -c "$length"
    elif [[ -f /dev/urandom ]]; then
        tr -dc 'a-zA-Z0-9' < /dev/urandom | head -c "$length"
    else
        # Fallback: use date and process ID (less secure)
        echo "$(date +%s%N)$$" | sha256sum | head -c "$length"
    fi
}

# Generate bootstrap token
generate_bootstrap_token() {
    local random_part
    random_part=$(generate_password 32)
    echo "aami_bootstrap_${random_part}"
}

# Check if running as root or with sudo
check_root() {
    if [[ $EUID -ne 0 ]]; then
        if command -v sudo &>/dev/null; then
            return 0  # Can use sudo
        else
            return 1  # No root access
        fi
    fi
    return 0  # Running as root
}

# Run command with sudo if needed
run_sudo() {
    if [[ $EUID -ne 0 ]]; then
        sudo "$@"
    else
        "$@"
    fi
}

# Prompt for yes/no with default
prompt_yn() {
    local prompt=$1
    local default=${2:-y}
    local response

    if [[ "$UNATTENDED" == true ]]; then
        [[ "$default" == "y" ]] && return 0 || return 1
    fi

    if [[ "$default" == "y" ]]; then
        read -r -p "$prompt [Y/n]: " response
        [[ -z "$response" || "$response" =~ ^[Yy] ]] && return 0 || return 1
    else
        read -r -p "$prompt [y/N]: " response
        [[ "$response" =~ ^[Yy] ]] && return 0 || return 1
    fi
}

# Prompt for input with default
prompt_input() {
    local prompt=$1
    local default=$2
    local response

    if [[ "$UNATTENDED" == true ]]; then
        echo "$default"
        return
    fi

    read -r -p "$prompt [$default]: " response
    echo "${response:-$default}"
}

# ==============================================================================
# Argument Parsing
# ==============================================================================

show_help() {
    cat << EOF
${SCRIPT_NAME} v${SCRIPT_VERSION}

One-command installation of the complete AAMI monitoring stack.

Usage:
  $0 [OPTIONS]

Options:
  --version VERSION      AAMI version to install (default: ${DEFAULT_VERSION})
  --install-dir PATH     Installation directory (default: ${DEFAULT_INSTALL_DIR})
  --data-dir PATH        Data directory (default: ${DEFAULT_DATA_DIR})
  --domain DOMAIN        Domain for Config Server (default: ${DEFAULT_DOMAIN})
  --port PORT            Config Server port (default: ${DEFAULT_PORT})
  --postgres-password PW PostgreSQL password (auto-generated if not set)
  --grafana-password PW  Grafana admin password (auto-generated if not set)
  --skip-preflight       Skip preflight checks (not recommended)
  --unattended           Non-interactive mode (use defaults)
  --verbose              Show detailed output
  --help                 Show this help message

Examples:
  # Interactive installation
  $0

  # Unattended installation with custom settings
  $0 --unattended --domain config.example.com

  # Specify version
  $0 --version v1.0.0

Environment Variables:
  AAMI_INSTALL_DIR       Override default installation directory
  AAMI_DATA_DIR          Override default data directory
  AAMI_POSTGRES_PASSWORD PostgreSQL password
  AAMI_GRAFANA_PASSWORD  Grafana admin password

EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --install-dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --data-dir)
                DATA_DIR="$2"
                shift 2
                ;;
            --domain)
                DOMAIN="$2"
                shift 2
                ;;
            --port)
                PORT="$2"
                shift 2
                ;;
            --postgres-password)
                POSTGRES_PASSWORD="$2"
                shift 2
                ;;
            --grafana-password)
                GRAFANA_PASSWORD="$2"
                shift 2
                ;;
            --skip-preflight)
                SKIP_PREFLIGHT=true
                shift
                ;;
            --unattended)
                UNATTENDED=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Use --help for usage information."
                exit 1
                ;;
        esac
    done

    # Apply environment variable overrides
    INSTALL_DIR="${INSTALL_DIR:-${AAMI_INSTALL_DIR:-$DEFAULT_INSTALL_DIR}}"
    DATA_DIR="${DATA_DIR:-${AAMI_DATA_DIR:-$DEFAULT_DATA_DIR}}"
    DOMAIN="${DOMAIN:-$DEFAULT_DOMAIN}"
    PORT="${PORT:-$DEFAULT_PORT}"
    VERSION="${VERSION:-$DEFAULT_VERSION}"
    POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-${AAMI_POSTGRES_PASSWORD:-}}"
    GRAFANA_PASSWORD="${GRAFANA_PASSWORD:-${AAMI_GRAFANA_PASSWORD:-}}"
}

# ==============================================================================
# Installation Steps
# ==============================================================================

# Step 1: Run preflight checks
run_preflight_checks() {
    print_step 1 7 "Running preflight checks..."

    if [[ "$SKIP_PREFLIGHT" == true ]]; then
        print_substep "warn" "Preflight checks skipped (--skip-preflight)"
        return 0
    fi

    # Check if preflight script exists locally or download it
    local preflight_script=""
    if [[ -f "${INSTALL_DIR}/scripts/preflight-check.sh" ]]; then
        preflight_script="${INSTALL_DIR}/scripts/preflight-check.sh"
    elif [[ -f "$(dirname "$0")/preflight-check.sh" ]]; then
        preflight_script="$(dirname "$0")/preflight-check.sh"
    else
        # Download preflight script
        print_substep "info" "Downloading preflight check script..."
        local temp_script
        temp_script=$(mktemp)
        if ! curl -fsSL "${GITHUB_RAW}/main/scripts/preflight-check.sh" -o "$temp_script" 2>/dev/null; then
            print_substep "warn" "Could not download preflight script, running basic checks"
            run_basic_checks
            return $?
        fi
        chmod +x "$temp_script"
        preflight_script="$temp_script"
    fi

    # Run preflight checks
    local preflight_args="--mode server"
    if [[ "$VERBOSE" == true ]]; then
        preflight_args="$preflight_args --verbose"
    fi

    print_verbose "Running: $preflight_script $preflight_args"

    if bash "$preflight_script" $preflight_args; then
        print_substep "ok" "All preflight checks passed"
        return 0
    else
        local exit_code=$?
        print_substep "fail" "Preflight checks failed"

        if [[ "$UNATTENDED" == true ]]; then
            print_error "Preflight checks failed. Run with --skip-preflight to ignore (not recommended)."
            return 1
        fi

        if prompt_yn "Continue anyway?" "n"; then
            print_substep "warn" "Continuing despite preflight failures"
            return 0
        else
            return 1
        fi
    fi
}

# Basic checks if preflight script is unavailable
run_basic_checks() {
    local failed=0

    # Check Docker
    if command -v docker &>/dev/null; then
        print_substep "ok" "Docker: $(docker --version | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+')"
    else
        print_substep "fail" "Docker not installed"
        ((failed++))
    fi

    # Check Docker Compose
    if command -v docker-compose &>/dev/null || docker compose version &>/dev/null 2>&1; then
        local compose_version
        compose_version=$(docker compose version 2>/dev/null | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+' || docker-compose --version | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+')
        print_substep "ok" "Docker Compose: ${compose_version}"
    else
        print_substep "fail" "Docker Compose not installed"
        ((failed++))
    fi

    # Check curl
    if command -v curl &>/dev/null; then
        print_substep "ok" "curl: available"
    else
        print_substep "fail" "curl not installed"
        ((failed++))
    fi

    # Check port availability
    for port in "$PORT" 9090 3000 5432 6379; do
        if ! (echo >/dev/tcp/localhost/$port) 2>/dev/null; then
            print_substep "ok" "Port $port: available"
        else
            print_substep "fail" "Port $port: in use"
            ((failed++))
        fi
    done

    return $failed
}

# Step 2: Download AAMI
download_aami() {
    print_step 2 7 "Downloading AAMI..."

    # Check if already exists
    if [[ -d "$INSTALL_DIR" && -f "${INSTALL_DIR}/deploy/docker-compose/docker-compose.yaml" ]]; then
        print_substep "info" "AAMI already exists at ${INSTALL_DIR}"

        if [[ "$UNATTENDED" == true ]]; then
            print_substep "info" "Using existing installation"
            return 0
        fi

        if prompt_yn "Use existing installation?" "y"; then
            return 0
        fi

        print_substep "info" "Removing existing installation..."
        run_sudo rm -rf "$INSTALL_DIR"
    fi

    # Create installation directory
    print_substep "info" "Creating installation directory: ${INSTALL_DIR}"
    run_sudo mkdir -p "$INSTALL_DIR"

    # Determine download method
    if [[ "$VERSION" == "latest" ]]; then
        # Clone main branch
        print_substep "info" "Cloning latest version from GitHub..."
        if command -v git &>/dev/null; then
            if ! run_sudo git clone --depth 1 "${GITHUB_URL}.git" "$INSTALL_DIR" 2>/dev/null; then
                print_substep "fail" "Failed to clone repository"
                return 1
            fi
        else
            # Download tarball
            print_substep "info" "git not available, downloading tarball..."
            local temp_file
            temp_file=$(mktemp)
            if ! curl -fsSL "${GITHUB_URL}/archive/refs/heads/main.tar.gz" -o "$temp_file"; then
                print_substep "fail" "Failed to download AAMI"
                return 1
            fi
            run_sudo tar -xzf "$temp_file" -C "$(dirname "$INSTALL_DIR")"
            run_sudo mv "$(dirname "$INSTALL_DIR")/aami-main" "$INSTALL_DIR"
            rm -f "$temp_file"
        fi
    else
        # Download specific version
        print_substep "info" "Downloading version ${VERSION}..."
        local temp_file
        temp_file=$(mktemp)
        if ! curl -fsSL "${GITHUB_URL}/archive/refs/tags/${VERSION}.tar.gz" -o "$temp_file" 2>/dev/null; then
            print_substep "fail" "Failed to download version ${VERSION}"
            return 1
        fi
        run_sudo tar -xzf "$temp_file" -C "$(dirname "$INSTALL_DIR")"
        run_sudo mv "$(dirname "$INSTALL_DIR")/aami-${VERSION#v}" "$INSTALL_DIR"
        rm -f "$temp_file"
    fi

    print_substep "ok" "Downloaded to ${INSTALL_DIR}"
    return 0
}

# Step 3: Configure environment
configure_environment() {
    print_step 3 7 "Configuring environment..."

    local compose_dir="${INSTALL_DIR}/deploy/docker-compose"
    local env_file="${compose_dir}/.env"
    local env_example="${compose_dir}/.env.example"

    # Create data directory
    print_substep "info" "Creating data directory: ${DATA_DIR}"
    run_sudo mkdir -p "${DATA_DIR}"/{postgres,prometheus,grafana}

    # Generate passwords if not provided
    if [[ -z "$POSTGRES_PASSWORD" ]]; then
        POSTGRES_PASSWORD=$(generate_password 24)
        print_substep "ok" "Generated PostgreSQL password"
    else
        print_substep "info" "Using provided PostgreSQL password"
    fi

    if [[ -z "$GRAFANA_PASSWORD" ]]; then
        GRAFANA_PASSWORD=$(generate_password 16)
        print_substep "ok" "Generated Grafana password"
    else
        print_substep "info" "Using provided Grafana password"
    fi

    # Interactive configuration (if not unattended)
    if [[ "$UNATTENDED" != true ]]; then
        echo ""
        echo "Configuration (press Enter to accept defaults):"
        DOMAIN=$(prompt_input "  Config Server domain" "$DOMAIN")
        PORT=$(prompt_input "  Config Server port" "$PORT")
        echo ""
    fi

    # Create .env file
    print_substep "info" "Creating .env file..."

    if [[ -f "$env_example" ]]; then
        run_sudo cp "$env_example" "$env_file"
    fi

    # Write configuration to .env file
    run_sudo tee "$env_file" > /dev/null << EOF
# AAMI Configuration
# Generated by install-server.sh at $(date -Iseconds)

# PostgreSQL Configuration
POSTGRES_USER=aami
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
POSTGRES_DB=aami

# Redis Configuration
REDIS_PASSWORD=

# Config Server Configuration
CONFIG_SERVER_PORT=${PORT}
CONFIG_SERVER_DOMAIN=${DOMAIN}

# Grafana Configuration
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=${GRAFANA_PASSWORD}

# Data Directories
DATA_DIR=${DATA_DIR}

# Installation Info
AAMI_INSTALL_DIR=${INSTALL_DIR}
AAMI_VERSION=${VERSION}
EOF

    print_substep "ok" "Environment configured"
    return 0
}

# Step 4: Start services
start_services() {
    print_step 4 7 "Starting services..."

    local compose_dir="${INSTALL_DIR}/deploy/docker-compose"
    cd "$compose_dir" || return 1

    # Determine docker compose command
    local compose_cmd="docker compose"
    if ! docker compose version &>/dev/null 2>&1; then
        compose_cmd="docker-compose"
    fi

    # Pull images first
    print_substep "info" "Pulling Docker images..."
    if ! run_sudo $compose_cmd pull 2>/dev/null; then
        print_substep "warn" "Some images may need to be built"
    fi

    # Start services
    print_substep "info" "Starting containers..."
    if ! run_sudo $compose_cmd up -d 2>&1 | while read -r line; do
        print_verbose "$line"
    done; then
        print_substep "fail" "Failed to start services"
        return 1
    fi

    # List started services
    local services
    services=$(run_sudo $compose_cmd ps --format '{{.Service}}' 2>/dev/null | tr '\n' ' ')
    print_substep "ok" "Started services: ${services:-all}"

    return 0
}

# Step 5: Wait for health checks
wait_for_health() {
    print_step 5 7 "Waiting for services to be healthy..."

    local max_wait=120
    local interval=5
    local elapsed=0

    # Wait for Config Server
    print_substep "info" "Waiting for Config Server..."
    while [[ $elapsed -lt $max_wait ]]; do
        if curl -sf "http://localhost:${PORT}/api/v1/health" &>/dev/null; then
            print_substep "ok" "Config Server is healthy"
            break
        fi
        sleep $interval
        ((elapsed += interval))
        print_verbose "Waiting... ${elapsed}s/${max_wait}s"
    done

    if [[ $elapsed -ge $max_wait ]]; then
        print_substep "fail" "Config Server did not become healthy within ${max_wait}s"
        return 1
    fi

    # Wait for Prometheus
    print_substep "info" "Waiting for Prometheus..."
    elapsed=0
    while [[ $elapsed -lt $max_wait ]]; do
        if curl -sf "http://localhost:9090/-/healthy" &>/dev/null; then
            print_substep "ok" "Prometheus is healthy"
            break
        fi
        sleep $interval
        ((elapsed += interval))
    done

    # Wait for Grafana
    print_substep "info" "Waiting for Grafana..."
    elapsed=0
    while [[ $elapsed -lt $max_wait ]]; do
        if curl -sf "http://localhost:3000/api/health" &>/dev/null; then
            print_substep "ok" "Grafana is healthy"
            break
        fi
        sleep $interval
        ((elapsed += interval))
    done

    print_substep "ok" "All services healthy"
    return 0
}

# Step 6: Create bootstrap token
create_bootstrap_token() {
    print_step 6 7 "Creating bootstrap token..."

    local token_name="initial-bootstrap-$(date +%Y%m%d)"
    local expires_at
    expires_at=$(date -d "+30 days" -Iseconds 2>/dev/null || date -v+30d -Iseconds 2>/dev/null || echo "")

    # Build token payload
    local token_payload
    if [[ -n "$expires_at" ]]; then
        token_payload='{
            "name": "'"${token_name}"'",
            "max_uses": 100,
            "expires_at": "'"${expires_at}"'"
        }'
    else
        # Fallback: 30 days from now in a simpler format
        token_payload='{
            "name": "'"${token_name}"'",
            "max_uses": 100
        }'
    fi

    print_substep "info" "Creating bootstrap token..."

    local token_response
    token_response=$(curl -sf -X POST "http://localhost:${PORT}/api/v1/bootstrap-tokens" \
        -H "Content-Type: application/json" \
        -d "$token_payload" 2>/dev/null)

    BOOTSTRAP_TOKEN=$(echo "$token_response" | grep -o '"token":"[^"]*"' | head -1 | cut -d'"' -f4)

    if [[ -n "$BOOTSTRAP_TOKEN" ]]; then
        print_substep "ok" "Bootstrap token created"
        print_substep "info" "Nodes registered with this token will have self-groups auto-created"
    else
        # Generate a placeholder if API fails
        BOOTSTRAP_TOKEN=$(generate_bootstrap_token)
        print_substep "warn" "API not available, generated placeholder token"
        print_substep "info" "Create a real token via the API after services are fully ready"
    fi

    return 0
}

# Step 7: Finalize and show summary
finalize_installation() {
    print_step 7 7 "Finalizing installation..."

    # Save credentials to a file
    local creds_file="${INSTALL_DIR}/credentials.txt"
    run_sudo tee "$creds_file" > /dev/null << EOF
# AAMI Credentials
# Generated at $(date -Iseconds)
# IMPORTANT: Keep this file secure!

PostgreSQL:
  Host: localhost:5432
  Database: aami
  User: aami
  Password: ${POSTGRES_PASSWORD}

Grafana:
  URL: http://localhost:3000
  User: admin
  Password: ${GRAFANA_PASSWORD}

Bootstrap Token:
  ${BOOTSTRAP_TOKEN}

Config Server:
  URL: http://${DOMAIN}:${PORT}
  Health: http://localhost:${PORT}/api/v1/health

Prometheus:
  URL: http://localhost:9090
EOF

    run_sudo chmod 600 "$creds_file"
    print_substep "ok" "Credentials saved to ${creds_file}"

    print_substep "ok" "Installation complete!"

    # Print success summary
    echo ""
    echo -e "${BOLD}${GREEN}"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                                                              │"
    echo "│  ✅ AAMI installed successfully!                            │"
    echo "│                                                              │"
    echo "├─────────────────────────────────────────────────────────────┤"
    echo -e "${NC}"
    echo -e "${BOLD}  Access URLs:${NC}"
    echo -e "    Config Server: ${CYAN}http://${DOMAIN}:${PORT}${NC}"
    echo -e "    Grafana:       ${CYAN}http://localhost:3000${NC} (admin/${GRAFANA_PASSWORD:0:8}...)"
    echo -e "    Prometheus:    ${CYAN}http://localhost:9090${NC}"
    echo ""
    echo -e "${BOLD}  Bootstrap Token (save this!):${NC}"
    echo -e "    ${YELLOW}${BOOTSTRAP_TOKEN}${NC}"
    echo ""
    echo -e "${BOLD}  Credentials File:${NC}"
    echo -e "    ${creds_file}"
    echo ""
    echo -e "${BOLD}  Next Steps:${NC}"
    echo "    1. Register nodes using the bootstrap token:"
    echo -e "       ${CYAN}curl -fsSL http://${DOMAIN}:${PORT}/bootstrap.sh | \\"
    echo -e "         sudo bash -s -- --token ${BOOTSTRAP_TOKEN:0:20}... --server http://${DOMAIN}:${PORT}${NC}"
    echo ""
    echo "    2. Access Grafana to view dashboards"
    echo "    3. Configure alert rules in Config Server"
    echo ""
    echo -e "${BOLD}  Documentation:${NC}"
    echo "    ${GITHUB_URL}#readme"
    echo ""
    echo -e "${GREEN}"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo -e "${NC}"

    return 0
}

# ==============================================================================
# Main
# ==============================================================================

main() {
    # Parse arguments
    parse_args "$@"

    # Print header
    print_header

    # Check for root/sudo
    if ! check_root; then
        print_error "This script requires root or sudo access."
        exit 1
    fi

    # Run installation steps
    if ! run_preflight_checks; then
        print_error "Preflight checks failed. Fix the issues and try again."
        exit 1
    fi

    if ! download_aami; then
        print_error "Failed to download AAMI."
        exit 1
    fi

    if ! configure_environment; then
        print_error "Failed to configure environment."
        exit 1
    fi

    if ! start_services; then
        print_error "Failed to start services."
        exit 1
    fi

    if ! wait_for_health; then
        print_error "Services did not become healthy."
        print_error "Check logs with: docker compose -f ${INSTALL_DIR}/deploy/docker-compose/docker-compose.yaml logs"
        exit 1
    fi

    if ! create_bootstrap_token; then
        print_warning "Failed to create bootstrap token. You can create one manually via the API."
    fi

    finalize_installation

    exit 0
}

# Run main function
main "$@"
