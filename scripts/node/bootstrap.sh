#!/usr/bin/env bash
#
# AAMI Node Bootstrap Script
# One-command registration and setup of monitoring agents on target nodes.
#
# Usage:
#   curl -fsSL http://config-server:8080/bootstrap.sh | \
#     sudo bash -s -- --token TOKEN --server URL
#
# Options:
#   --token TOKEN          Bootstrap token from Config Server (required)
#   --server URL           Config Server URL (required)
#   --port PORT            Node Exporter port (default: 9100)
#   --labels KEY=VALUE     Additional labels (can be repeated)
#   --group-id ID          Group ID to assign target to
#   --dry-run              Show what would be done without executing
#   --skip-preflight       Skip preflight checks (not recommended)
#   --skip-gpu             Skip GPU detection and exporter installation
#   --install-all-smi      Install all-smi multi-vendor GPU exporter
#   --unattended           Non-interactive mode
#   --verbose              Enable verbose output
#   --help                 Show this help message
#

set -uo pipefail

# ==============================================================================
# Constants
# ==============================================================================

readonly SCRIPT_NAME="AAMI Node Bootstrap"
readonly SCRIPT_VERSION="1.0.0"

# Default values
DEFAULT_PORT="9100"
DEFAULT_DCGM_PORT="9400"
DEFAULT_ALL_SMI_PORT="9401"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Configuration (set by arguments)
BOOTSTRAP_TOKEN=""
CONFIG_SERVER=""
NODE_EXPORTER_PORT="$DEFAULT_PORT"
DCGM_EXPORTER_PORT="$DEFAULT_DCGM_PORT"
ALL_SMI_PORT="$DEFAULT_ALL_SMI_PORT"
declare -a LABELS=()
GROUP_ID=""
DRY_RUN=false
SKIP_PREFLIGHT=false
SKIP_GPU=false
INSTALL_ALL_SMI=false
UNATTENDED=false
VERBOSE=false

# Detected system info
DETECTED_HOSTNAME=""
DETECTED_IP=""
DETECTED_OS=""
DETECTED_OS_VERSION=""
DETECTED_ARCH=""
DETECTED_GPU_VENDOR=""
DETECTED_GPU_MODEL=""
DETECTED_GPU_COUNT=0

# Result
REGISTERED_TARGET_ID=""

# ==============================================================================
# Utility Functions
# ==============================================================================

print_header() {
    echo -e "${BOLD}${BLUE}"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                   AAMI Node Bootstrap                        │"
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

print_verbose() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# ==============================================================================
# Argument Parsing
# ==============================================================================

show_help() {
    cat << EOF
${SCRIPT_NAME} v${SCRIPT_VERSION}

One-command registration and setup of monitoring agents on target nodes.

Usage:
  $0 --token TOKEN --server URL [OPTIONS]

Required:
  --token TOKEN          Bootstrap token from Config Server
  --server URL           Config Server URL (e.g., http://config-server:8080)

Optional:
  --port PORT            Node Exporter port (default: ${DEFAULT_PORT})
  --labels KEY=VALUE     Additional labels (can be repeated)
  --group-id ID          Group ID to assign target to (default: self group)
  --dry-run              Show what would be done without executing
  --skip-preflight       Skip preflight checks (not recommended)
  --skip-gpu             Skip GPU detection and exporter installation
  --install-all-smi      Install all-smi multi-vendor GPU exporter (port: ${DEFAULT_ALL_SMI_PORT})
  --unattended           Non-interactive mode
  --verbose              Enable verbose output
  --help                 Show this help message

Examples:
  # Basic usage
  $0 --token aami_bootstrap_xxx --server http://config-server:8080

  # With custom labels
  $0 --token aami_xxx --server http://config-server:8080 \\
    --labels env=production --labels rack=A1

  # Dry run to preview
  $0 --token aami_xxx --server http://config-server:8080 --dry-run

  # One-liner from Config Server
  curl -fsSL http://config-server:8080/bootstrap.sh | \\
    sudo bash -s -- --token aami_xxx --server http://config-server:8080

EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --token)
                BOOTSTRAP_TOKEN="$2"
                shift 2
                ;;
            --server)
                CONFIG_SERVER="$2"
                shift 2
                ;;
            --port)
                NODE_EXPORTER_PORT="$2"
                shift 2
                ;;
            --labels)
                LABELS+=("$2")
                shift 2
                ;;
            --group-id)
                GROUP_ID="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --skip-preflight)
                SKIP_PREFLIGHT=true
                shift
                ;;
            --skip-gpu)
                SKIP_GPU=true
                shift
                ;;
            --install-all-smi)
                INSTALL_ALL_SMI=true
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

    # Validate required arguments
    if [[ -z "$BOOTSTRAP_TOKEN" ]]; then
        print_error "Missing required argument: --token"
        echo "Use --help for usage information."
        exit 1
    fi

    if [[ -z "$CONFIG_SERVER" ]]; then
        print_error "Missing required argument: --server"
        echo "Use --help for usage information."
        exit 1
    fi

    # Remove trailing slash from server URL
    CONFIG_SERVER="${CONFIG_SERVER%/}"
}

# ==============================================================================
# Bootstrap Steps
# ==============================================================================

# Step 1: Run preflight checks
run_preflight_checks() {
    print_step 1 8 "Running preflight checks..."

    if [[ "$SKIP_PREFLIGHT" == true ]]; then
        print_substep "warn" "Preflight checks skipped (--skip-preflight)"
        return 0
    fi

    # Check basic requirements
    local failed=0

    # Check curl
    if command -v curl &>/dev/null; then
        print_substep "ok" "curl: available"
    else
        print_substep "fail" "curl: not installed"
        ((failed++))
    fi

    # Check systemctl
    if command -v systemctl &>/dev/null; then
        print_substep "ok" "systemctl: available"
    else
        print_substep "fail" "systemctl: not available (systemd required)"
        ((failed++))
    fi

    # Check port availability
    if ! (echo >/dev/tcp/localhost/$NODE_EXPORTER_PORT) 2>/dev/null; then
        print_substep "ok" "Port ${NODE_EXPORTER_PORT}: available"
    else
        print_substep "fail" "Port ${NODE_EXPORTER_PORT}: in use"
        ((failed++))
    fi

    # Check Config Server connectivity
    print_substep "info" "Testing Config Server connectivity..."
    if curl -sf "${CONFIG_SERVER}/api/v1/health" &>/dev/null; then
        print_substep "ok" "Config Server: reachable"
    else
        print_substep "fail" "Config Server: not reachable at ${CONFIG_SERVER}"
        ((failed++))
    fi

    if [[ $failed -gt 0 ]]; then
        print_error "Preflight checks failed with $failed error(s)"
        return 1
    fi

    print_substep "ok" "All preflight checks passed"
    return 0
}

# Step 2: Validate bootstrap token
validate_token() {
    print_step 2 8 "Validating bootstrap token..."

    if [[ "$DRY_RUN" == true ]]; then
        print_substep "info" "[DRY-RUN] Would validate token: ${BOOTSTRAP_TOKEN:0:20}..."
        return 0
    fi

    # Call token validation endpoint
    local response
    local http_code

    response=$(curl -sf -w "\n%{http_code}" \
        "${CONFIG_SERVER}/api/v1/bootstrap-tokens/validate" \
        -H "Content-Type: application/json" \
        -d "{\"token\": \"${BOOTSTRAP_TOKEN}\"}" 2>/dev/null) || true

    http_code=$(echo "$response" | tail -n1)
    local body
    body=$(echo "$response" | sed '$d')

    print_verbose "Token validation response: $http_code"
    print_verbose "Response body: $body"

    if [[ "$http_code" == "200" ]]; then
        local remaining
        remaining=$(echo "$body" | grep -o '"remaining_uses":[0-9]*' | cut -d: -f2 || echo "unknown")
        print_substep "ok" "Token valid (${remaining} uses remaining)"
        return 0
    elif [[ "$http_code" == "404" ]] || [[ "$http_code" == "401" ]]; then
        print_substep "fail" "Token invalid or expired"
        return 1
    else
        # Try simple GET validation as fallback
        print_substep "warn" "Token validation endpoint not available, proceeding..."
        return 0
    fi
}

# Step 3: Detect system information
detect_system_info() {
    print_step 3 8 "Detecting system information..."

    # Hostname
    DETECTED_HOSTNAME=$(hostname -f 2>/dev/null || hostname)
    print_substep "ok" "Hostname: ${DETECTED_HOSTNAME}"

    # IP Address
    DETECTED_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || ip route get 1 | awk '{print $7; exit}' 2>/dev/null || echo "127.0.0.1")
    print_substep "ok" "IP Address: ${DETECTED_IP}"

    # OS Info
    if [[ -f /etc/os-release ]]; then
        # shellcheck disable=SC1091
        source /etc/os-release
        DETECTED_OS="$ID"
        DETECTED_OS_VERSION="$VERSION_ID"
    else
        DETECTED_OS=$(uname -s)
        DETECTED_OS_VERSION=$(uname -r)
    fi
    print_substep "ok" "OS: ${DETECTED_OS} ${DETECTED_OS_VERSION}"

    # Architecture
    DETECTED_ARCH=$(uname -m)
    case $DETECTED_ARCH in
        x86_64) DETECTED_ARCH="amd64" ;;
        aarch64) DETECTED_ARCH="arm64" ;;
    esac
    print_substep "ok" "Architecture: ${DETECTED_ARCH}"

    return 0
}

# Step 4: Detect hardware (GPU)
detect_hardware() {
    print_step 4 8 "Detecting hardware..."

    if [[ "$SKIP_GPU" == true ]]; then
        print_substep "info" "GPU detection skipped (--skip-gpu)"
        return 0
    fi

    # NVIDIA GPU detection
    if command -v nvidia-smi &>/dev/null; then
        DETECTED_GPU_VENDOR="nvidia"
        DETECTED_GPU_COUNT=$(nvidia-smi --query-gpu=count --format=csv,noheader,nounits 2>/dev/null | head -1 || echo "0")
        DETECTED_GPU_MODEL=$(nvidia-smi --query-gpu=name --format=csv,noheader 2>/dev/null | head -1 || echo "unknown")
        print_substep "ok" "NVIDIA GPU detected: ${DETECTED_GPU_COUNT}x ${DETECTED_GPU_MODEL}"
    # AMD GPU detection
    elif command -v rocm-smi &>/dev/null; then
        DETECTED_GPU_VENDOR="amd"
        DETECTED_GPU_COUNT=$(rocm-smi --showproductname 2>/dev/null | grep -c "GPU" || echo "0")
        DETECTED_GPU_MODEL=$(rocm-smi --showproductname 2>/dev/null | grep "GPU" | head -1 | awk '{print $NF}' || echo "unknown")
        print_substep "ok" "AMD GPU detected: ${DETECTED_GPU_COUNT}x ${DETECTED_GPU_MODEL}"
    else
        print_substep "info" "No GPU detected"
    fi

    return 0
}

# Step 5: Install Node Exporter
install_node_exporter() {
    print_step 5 8 "Installing Node Exporter..."

    if [[ "$DRY_RUN" == true ]]; then
        print_substep "info" "[DRY-RUN] Would install Node Exporter on port ${NODE_EXPORTER_PORT}"
        return 0
    fi

    # Check if already installed
    if systemctl is-active --quiet node_exporter 2>/dev/null; then
        print_substep "info" "Node Exporter already running"
        return 0
    fi

    # Get script directory
    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

    # Try to use local install script first
    if [[ -f "${script_dir}/install-node-exporter.sh" ]]; then
        print_substep "info" "Using local install script..."
        bash "${script_dir}/install-node-exporter.sh" --port "$NODE_EXPORTER_PORT"
    else
        # Download and run install script
        print_substep "info" "Downloading install script..."
        local temp_script
        temp_script=$(mktemp)
        if curl -fsSL "${CONFIG_SERVER}/scripts/node/install-node-exporter.sh" -o "$temp_script" 2>/dev/null; then
            chmod +x "$temp_script"
            bash "$temp_script" --port "$NODE_EXPORTER_PORT"
            rm -f "$temp_script"
        else
            # Try GitHub directly
            if curl -fsSL "https://raw.githubusercontent.com/fregataa/aami/main/scripts/node/install-node-exporter.sh" -o "$temp_script" 2>/dev/null; then
                chmod +x "$temp_script"
                bash "$temp_script" --port "$NODE_EXPORTER_PORT"
                rm -f "$temp_script"
            else
                print_substep "fail" "Failed to download install script"
                return 1
            fi
        fi
    fi

    # Verify installation
    sleep 2
    if systemctl is-active --quiet node_exporter; then
        print_substep "ok" "Node Exporter installed and running"
    else
        print_substep "fail" "Node Exporter installation failed"
        return 1
    fi

    return 0
}

# Step 5.5: Install all-smi (optional)
install_all_smi() {
    if [[ "$INSTALL_ALL_SMI" != true ]]; then
        return 0
    fi

    print_step "5.5" 8 "Installing all-smi multi-vendor GPU exporter..."

    if [[ "$DRY_RUN" == true ]]; then
        print_substep "info" "[DRY-RUN] Would install all-smi on port ${ALL_SMI_PORT}"
        return 0
    fi

    # Check if already installed
    if systemctl is-active --quiet all-smi 2>/dev/null; then
        print_substep "info" "all-smi already running"
        return 0
    fi

    # Get script directory
    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

    # Try Python script first (recommended), fall back to bash script
    if command -v python3 &>/dev/null && [[ -f "${script_dir}/install_all_smi.py" ]]; then
        print_substep "info" "Using Python install script (recommended)..."
        python3 "${script_dir}/install_all_smi.py" --port "$ALL_SMI_PORT"
    elif [[ -f "${script_dir}/install-all-smi.sh" ]]; then
        print_substep "info" "Using bash install script (legacy)..."
        bash "${script_dir}/install-all-smi.sh" --port "$ALL_SMI_PORT"
    else
        # Download and run install script
        print_substep "info" "Downloading install script..."
        local temp_script
        temp_script=$(mktemp)

        # Try Python script from Config Server
        if command -v python3 &>/dev/null && curl -fsSL "${CONFIG_SERVER}/scripts/node/install_all_smi.py" -o "$temp_script" 2>/dev/null; then
            chmod +x "$temp_script"
            python3 "$temp_script" --port "$ALL_SMI_PORT"
            rm -f "$temp_script"
        # Try bash script from Config Server
        elif curl -fsSL "${CONFIG_SERVER}/scripts/node/install-all-smi.sh" -o "$temp_script" 2>/dev/null; then
            chmod +x "$temp_script"
            bash "$temp_script" --port "$ALL_SMI_PORT"
            rm -f "$temp_script"
        # Try GitHub directly (Python)
        elif command -v python3 &>/dev/null && curl -fsSL "https://raw.githubusercontent.com/fregataa/aami/main/scripts/node/install_all_smi.py" -o "$temp_script" 2>/dev/null; then
            chmod +x "$temp_script"
            python3 "$temp_script" --port "$ALL_SMI_PORT"
            rm -f "$temp_script"
        # Try GitHub directly (bash)
        elif curl -fsSL "https://raw.githubusercontent.com/fregataa/aami/main/scripts/node/install-all-smi.sh" -o "$temp_script" 2>/dev/null; then
            chmod +x "$temp_script"
            bash "$temp_script" --port "$ALL_SMI_PORT"
            rm -f "$temp_script"
        else
            print_substep "fail" "Failed to download install script"
            rm -f "$temp_script"
            return 1
        fi
    fi

    # Verify installation
    sleep 2
    if systemctl is-active --quiet all-smi; then
        print_substep "ok" "all-smi installed and running on port ${ALL_SMI_PORT}"
    else
        print_substep "fail" "all-smi installation failed"
        return 1
    fi

    return 0
}

# Step 6: Install dynamic check
install_dynamic_check() {
    print_step 6 8 "Installing dynamic check..."

    if [[ "$DRY_RUN" == true ]]; then
        print_substep "info" "[DRY-RUN] Would install dynamic check script and cron job"
        return 0
    fi

    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local install_dir="/opt/aami/scripts"
    local textfile_dir="/var/lib/node_exporter/textfile_collector"

    # Create directories
    mkdir -p "$install_dir"
    mkdir -p "$textfile_dir"

    # Prefer Python version if Python 3 is available
    local use_python=false
    local script_name=""
    local cron_command=""

    if command -v python3 &> /dev/null; then
        # Try to install Python version
        if [[ -f "${script_dir}/dynamic_check.py" ]]; then
            cp "${script_dir}/dynamic_check.py" "${install_dir}/"
            use_python=true
        else
            # Download from Config Server or GitHub
            if curl -fsSL "${CONFIG_SERVER}/scripts/node/dynamic_check.py" -o "${install_dir}/dynamic_check.py" 2>/dev/null; then
                use_python=true
            elif curl -fsSL "https://raw.githubusercontent.com/fregataa/aami/main/scripts/node/dynamic_check.py" -o "${install_dir}/dynamic_check.py" 2>/dev/null; then
                use_python=true
            fi
        fi

        if [[ "$use_python" == true ]]; then
            chmod +x "${install_dir}/dynamic_check.py"
            script_name="dynamic_check.py"
            cron_command="python3 ${install_dir}/dynamic_check.py --config-server ${CONFIG_SERVER} --textfile-dir ${textfile_dir}"
            print_substep "ok" "dynamic_check.py installed (Python)"
        fi
    fi

    # Fallback to bash version if Python not available
    if [[ "$use_python" == false ]]; then
        if [[ -f "${script_dir}/dynamic-check.sh" ]]; then
            cp "${script_dir}/dynamic-check.sh" "${install_dir}/"
        else
            # Download from Config Server or GitHub
            if ! curl -fsSL "${CONFIG_SERVER}/scripts/node/dynamic-check.sh" -o "${install_dir}/dynamic-check.sh" 2>/dev/null; then
                if ! curl -fsSL "https://raw.githubusercontent.com/fregataa/aami/main/scripts/node/dynamic-check.sh" -o "${install_dir}/dynamic-check.sh" 2>/dev/null; then
                    print_substep "warn" "Could not download dynamic check script, skipping"
                    return 0
                fi
            fi
        fi
        chmod +x "${install_dir}/dynamic-check.sh"
        script_name="dynamic-check.sh"
        cron_command="${install_dir}/dynamic-check.sh --server ${CONFIG_SERVER} --output-dir ${textfile_dir}"
        print_substep "ok" "dynamic-check.sh installed (Bash)"
    fi

    # Create cron job
    local cron_file="/etc/cron.d/aami-dynamic-check"
    cat > "$cron_file" << EOF
# AAMI Dynamic Check - runs every minute
* * * * * root ${cron_command} >> /var/log/aami-dynamic-check.log 2>&1
EOF

    chmod 644 "$cron_file"
    print_substep "ok" "Cron job registered (1-minute interval)"

    # Run first check
    print_substep "info" "Running initial check..."
    if [[ "$use_python" == true ]]; then
        python3 "${install_dir}/dynamic_check.py" --config-server "${CONFIG_SERVER}" --textfile-dir "${textfile_dir}" 2>/dev/null || true
    else
        "${install_dir}/dynamic-check.sh" --server "${CONFIG_SERVER}" --output-dir "${textfile_dir}" 2>/dev/null || true
    fi

    return 0
}

# Step 7: Register with Config Server
register_with_server() {
    print_step 7 8 "Registering with Config Server..."

    if [[ "$DRY_RUN" == true ]]; then
        print_substep "info" "[DRY-RUN] Would register node with Config Server"
        print_substep "info" "  Hostname: ${DETECTED_HOSTNAME}"
        print_substep "info" "  IP: ${DETECTED_IP}"
        print_substep "info" "  Labels: ${LABELS[*]:-none}"
        return 0
    fi

    # Build labels JSON
    local labels_json="{"
    labels_json+="\"os\": \"${DETECTED_OS}\","
    labels_json+="\"os_version\": \"${DETECTED_OS_VERSION}\","
    labels_json+="\"arch\": \"${DETECTED_ARCH}\""

    if [[ -n "$DETECTED_GPU_VENDOR" ]]; then
        labels_json+=",\"gpu_vendor\": \"${DETECTED_GPU_VENDOR}\""
        labels_json+=",\"gpu_model\": \"${DETECTED_GPU_MODEL}\""
        labels_json+=",\"gpu_count\": \"${DETECTED_GPU_COUNT}\""
    fi

    # Add custom labels
    for label in "${LABELS[@]}"; do
        local key="${label%%=*}"
        local value="${label#*=}"
        labels_json+=",\"${key}\": \"${value}\""
    done
    labels_json+="}"

    # Build request payload
    local payload="{
        \"token\": \"${BOOTSTRAP_TOKEN}\",
        \"hostname\": \"${DETECTED_HOSTNAME}\",
        \"ip_address\": \"${DETECTED_IP}\",
        \"labels\": ${labels_json}"

    if [[ -n "$GROUP_ID" ]]; then
        payload+=",\"group_id\": \"${GROUP_ID}\""
    fi

    payload+="}"

    print_verbose "Registration payload: $payload"

    # Call registration API
    local response
    local http_code

    response=$(curl -sf -w "\n%{http_code}" \
        -X POST "${CONFIG_SERVER}/api/v1/bootstrap/register" \
        -H "Content-Type: application/json" \
        -d "$payload" 2>/dev/null) || true

    http_code=$(echo "$response" | tail -n1)
    local body
    body=$(echo "$response" | sed '$d')

    print_verbose "Registration response: $http_code"
    print_verbose "Response body: $body"

    if [[ "$http_code" == "200" ]] || [[ "$http_code" == "201" ]]; then
        REGISTERED_TARGET_ID=$(echo "$body" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4 || echo "")
        print_substep "ok" "Node registered successfully"
        if [[ -n "$REGISTERED_TARGET_ID" ]]; then
            print_substep "info" "Target ID: ${REGISTERED_TARGET_ID}"
        fi
        return 0
    else
        print_substep "fail" "Registration failed (HTTP ${http_code})"
        print_substep "info" "Response: ${body}"
        return 1
    fi
}

# Step 7.5: Register exporters with Config Server
register_exporters() {
    print_step "7.5" 8 "Registering exporters..."

    if [[ "$DRY_RUN" == true ]]; then
        print_substep "info" "[DRY-RUN] Would register exporters"
        return 0
    fi

    if [[ -z "$REGISTERED_TARGET_ID" ]]; then
        print_substep "fail" "No target ID available"
        return 1
    fi

    # Register Node Exporter
    local node_payload="{
        \"target_id\": \"${REGISTERED_TARGET_ID}\",
        \"type\": \"node_exporter\",
        \"port\": ${NODE_EXPORTER_PORT},
        \"enabled\": true
    }"

    local response
    response=$(curl -sf -X POST "${CONFIG_SERVER}/api/v1/exporters" \
        -H "Content-Type: application/json" \
        -d "$node_payload" 2>/dev/null) || true

    if [[ -n "$response" ]]; then
        print_substep "ok" "Node Exporter registered (port ${NODE_EXPORTER_PORT})"
    else
        print_substep "warn" "Failed to register Node Exporter"
    fi

    # Register all-smi if installed
    if [[ "$INSTALL_ALL_SMI" == true ]]; then
        local allsmi_payload="{
            \"target_id\": \"${REGISTERED_TARGET_ID}\",
            \"type\": \"all_smi\",
            \"port\": ${ALL_SMI_PORT},
            \"enabled\": true
        }"

        response=$(curl -sf -X POST "${CONFIG_SERVER}/api/v1/exporters" \
            -H "Content-Type: application/json" \
            -d "$allsmi_payload" 2>/dev/null) || true

        if [[ -n "$response" ]]; then
            print_substep "ok" "all-smi registered (port ${ALL_SMI_PORT})"
        else
            print_substep "warn" "Failed to register all-smi"
        fi
    fi

    return 0
}

# Step 8: Verify registration
verify_registration() {
    print_step 8 8 "Verifying registration..."

    if [[ "$DRY_RUN" == true ]]; then
        print_substep "info" "[DRY-RUN] Would verify registration"
        return 0
    fi

    # Check Node Exporter metrics
    if curl -sf "http://localhost:${NODE_EXPORTER_PORT}/metrics" &>/dev/null; then
        print_substep "ok" "Node Exporter metrics accessible"
    else
        print_substep "warn" "Node Exporter metrics not accessible"
    fi

    # Check DCGM Exporter if GPU detected
    if [[ -n "$DETECTED_GPU_VENDOR" ]] && [[ "$DETECTED_GPU_VENDOR" == "nvidia" ]]; then
        if curl -sf "http://localhost:${DCGM_EXPORTER_PORT}/metrics" &>/dev/null; then
            print_substep "ok" "DCGM Exporter metrics accessible"
        else
            print_substep "info" "DCGM Exporter not installed (optional)"
        fi
    fi

    # Check all-smi if installed
    if [[ "$INSTALL_ALL_SMI" == true ]]; then
        if curl -sf "http://localhost:${ALL_SMI_PORT}/metrics" &>/dev/null; then
            print_substep "ok" "all-smi metrics accessible on port ${ALL_SMI_PORT}"
        else
            print_substep "warn" "all-smi metrics not accessible"
        fi
    fi

    # Verify target in Config Server
    if [[ -n "$REGISTERED_TARGET_ID" ]]; then
        if curl -sf "${CONFIG_SERVER}/api/v1/targets/${REGISTERED_TARGET_ID}" &>/dev/null; then
            print_substep "ok" "Target confirmed in Config Server"
        else
            print_substep "warn" "Could not verify target in Config Server"
        fi
    fi

    print_substep "ok" "Verification complete"
    return 0
}

# Print success summary
print_summary() {
    echo ""
    echo -e "${BOLD}${GREEN}"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                                                              │"
    echo "│  ✅ Node bootstrap complete!                                │"
    echo "│                                                              │"
    echo "├─────────────────────────────────────────────────────────────┤"
    echo -e "${NC}"
    echo -e "${BOLD}  Node Information:${NC}"
    echo -e "    Hostname:  ${CYAN}${DETECTED_HOSTNAME}${NC}"
    echo -e "    IP:        ${CYAN}${DETECTED_IP}${NC}"
    if [[ -n "$REGISTERED_TARGET_ID" ]]; then
        echo -e "    Target ID: ${CYAN}${REGISTERED_TARGET_ID}${NC}"
    fi
    echo ""
    echo -e "${BOLD}  Installed Components:${NC}"
    echo -e "    - Node Exporter:    ${CYAN}http://localhost:${NODE_EXPORTER_PORT}/metrics${NC}"
    if [[ "$INSTALL_ALL_SMI" == true ]]; then
        echo -e "    - all-smi:          ${CYAN}http://localhost:${ALL_SMI_PORT}/metrics${NC}"
    fi
    if [[ -n "$DETECTED_GPU_VENDOR" ]]; then
        echo -e "    - GPU:              ${CYAN}${DETECTED_GPU_COUNT}x ${DETECTED_GPU_MODEL}${NC}"
    fi
    echo -e "    - Dynamic Checks:   ${CYAN}/var/lib/node_exporter/textfile_collector/${NC}"
    echo ""
    echo -e "  The node will appear in Prometheus within 30 seconds."
    echo ""
    echo -e "${GREEN}"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo -e "${NC}"
}

# ==============================================================================
# Main
# ==============================================================================

main() {
    # Parse arguments
    parse_args "$@"

    # Print header
    print_header

    # Check root
    check_root

    # Run bootstrap steps
    if ! run_preflight_checks; then
        print_error "Preflight checks failed. Fix the issues and try again."
        exit 1
    fi

    if ! validate_token; then
        print_error "Token validation failed."
        exit 1
    fi

    detect_system_info
    detect_hardware

    if ! install_node_exporter; then
        print_error "Failed to install Node Exporter."
        exit 1
    fi

    if ! install_all_smi; then
        print_error "Failed to install all-smi."
        exit 1
    fi

    install_dynamic_check

    if ! register_with_server; then
        print_error "Failed to register with Config Server."
        exit 1
    fi

    register_exporters

    verify_registration

    # Print summary
    if [[ "$DRY_RUN" == true ]]; then
        echo ""
        echo -e "${YELLOW}[DRY-RUN] No changes were made.${NC}"
    else
        print_summary
    fi

    exit 0
}

# Run main function
main "$@"
