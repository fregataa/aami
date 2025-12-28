#!/usr/bin/env bash
#
# Dynamic Check Runner for AAMI Monitoring
#
# This script fetches effective checks from the AAMI Config Server,
# executes them, and outputs results to Node Exporter's textfile collector.
#
# Usage: ./dynamic-check.sh [OPTIONS]
#
# Options:
#   -c, --config-server URL  Config Server URL (default: from /etc/aami/config)
#   -h, --hostname NAME      Override hostname (default: $(hostname))
#   -d, --debug              Enable debug logging
#   --help                   Show this help message
#
# Environment Variables:
#   AAMI_CONFIG_SERVER_URL   - Config Server URL
#   AAMI_HOSTNAME            - Override hostname
#   AAMI_DEBUG               - Enable debug logging (1=on, 0=off)

set -euo pipefail

# Configuration
CONFIG_SERVER_URL="${AAMI_CONFIG_SERVER_URL:-}"
HOSTNAME="${AAMI_HOSTNAME:-$(hostname)}"
DEBUG="${AAMI_DEBUG:-0}"

# Paths
TEXTFILE_DIR="${TEXTFILE_DIR:-/var/lib/node_exporter/textfile_collector}"
CHECK_SCRIPTS_DIR="${CHECK_SCRIPTS_DIR:-/usr/local/lib/aami/checks}"
CONFIG_FILE="${CONFIG_FILE:-/etc/aami/config}"
LOG_FILE="${LOG_FILE:-/var/log/aami/dynamic-check.log}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"

    if [ "$DEBUG" = "1" ]; then
        case "$level" in
            ERROR) echo -e "${RED}[ERROR]${NC} $message" >&2 ;;
            WARN)  echo -e "${YELLOW}[WARN]${NC} $message" ;;
            INFO)  echo -e "${GREEN}[INFO]${NC} $message" ;;
            DEBUG) echo -e "${BLUE}[DEBUG]${NC} $message" ;;
        esac
    fi
}

error() { log ERROR "$@"; }
warn() { log WARN "$@"; }
info() { log INFO "$@"; }
debug() { log DEBUG "$@"; }

# Print usage
usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Dynamic Check Runner for AAMI Monitoring

Options:
    -c, --config-server URL  Config Server URL (default: from $CONFIG_FILE)
    -h, --hostname NAME      Override hostname (default: $(hostname))
    -d, --debug              Enable debug logging
    --help                   Show this help message

Environment Variables:
    AAMI_CONFIG_SERVER_URL   - Config Server URL
    AAMI_HOSTNAME            - Override hostname
    AAMI_DEBUG               - Enable debug logging (1=on, 0=off)

Example:
    $(basename "$0") --config-server http://config-server:8080
    $(basename "$0") --debug

EOF
    exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--config-server)
            CONFIG_SERVER_URL="$2"
            shift 2
            ;;
        --hostname)
            HOSTNAME="$2"
            shift 2
            ;;
        -d|--debug)
            DEBUG=1
            shift
            ;;
        --help)
            usage
            ;;
        *)
            error "Unknown option: $1"
            usage
            ;;
    esac
done

# Load config from file if not set via env/args
if [ -z "$CONFIG_SERVER_URL" ] && [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
fi

# Validate required configuration
if [ -z "$CONFIG_SERVER_URL" ]; then
    error "Config Server URL not configured. Set via:"
    error "  1. Environment variable: AAMI_CONFIG_SERVER_URL"
    error "  2. Config file: $CONFIG_FILE"
    error "  3. Command-line flag: --config-server"
    exit 1
fi

# Create directories if they don't exist
mkdir -p "$TEXTFILE_DIR" "$CHECK_SCRIPTS_DIR" "$(dirname "$LOG_FILE")"

info "Starting dynamic check run for hostname: $HOSTNAME"
debug "Config Server: $CONFIG_SERVER_URL"
debug "Textfile Directory: $TEXTFILE_DIR"
debug "Check Scripts Directory: $CHECK_SCRIPTS_DIR"

# Fetch effective checks from Config Server
fetch_effective_checks() {
    local url="${CONFIG_SERVER_URL}/api/v1/checks/node/${HOSTNAME}"
    local temp_file=$(mktemp)

    debug "Fetching effective checks from: $url"

    if ! curl -sf -H "Accept: application/json" "$url" -o "$temp_file"; then
        error "Failed to fetch effective checks from Config Server"
        rm -f "$temp_file"
        return 1
    fi

    cat "$temp_file"
    rm -f "$temp_file"
    return 0
}

# Save check script with hash-based versioning
save_check_script() {
    local check_name="$1"
    local script_content="$2"
    local script_hash="$3"

    local script_dir="$CHECK_SCRIPTS_DIR/$check_name"
    local script_file="$script_dir/${check_name}_${script_hash}.sh"
    local current_link="$script_dir/current.sh"

    # Create check directory
    mkdir -p "$script_dir"

    # Check if script already exists with this hash
    if [ -f "$script_file" ]; then
        debug "Check script already exists: $script_file"
    else
        info "Saving new check script: $check_name (hash: ${script_hash:0:8})"
        echo "$script_content" > "$script_file"
        chmod +x "$script_file"
    fi

    # Update symlink to current version
    ln -sf "$script_file" "$current_link"

    echo "$current_link"
}

# Execute a single check
execute_check() {
    local check_name="$1"
    local script_path="$2"
    local config="$3"

    local output_file="${TEXTFILE_DIR}/${check_name}.prom.$$"
    local final_file="${TEXTFILE_DIR}/${check_name}.prom"

    debug "Executing check: $check_name"
    debug "Script: $script_path"
    debug "Config: $config"

    # Execute check script with config as input
    if echo "$config" | timeout 30s "$script_path" > "$output_file" 2>&1; then
        # Move to final location atomically
        mv "$output_file" "$final_file"
        info "Check completed successfully: $check_name"
        return 0
    else
        local exit_code=$?
        error "Check failed: $check_name (exit code: $exit_code)"

        # Output error metric
        cat > "$output_file" <<EOF
# HELP aami_check_error Check execution error (1=error)
# TYPE aami_check_error gauge
aami_check_error{check=\"$check_name\"} 1
EOF
        mv "$output_file" "$final_file"

        # Log the error output
        if [ -f "$output_file" ]; then
            error "Check output: $(cat "$output_file")"
        fi

        return 1
    fi
}

# Main execution
main() {
    local start_time=$(date +%s)
    local checks_total=0
    local checks_success=0
    local checks_failed=0

    # Fetch effective checks
    info "Fetching effective checks from Config Server"
    local checks_json=$(fetch_effective_checks)

    if [ $? -ne 0 ] || [ -z "$checks_json" ]; then
        error "Failed to fetch checks or no checks configured"

        # Output error metric
        cat > "${TEXTFILE_DIR}/aami_status.prom" <<EOF
# HELP aami_check_fetch_status Check fetch status (1=success, 0=failed)
# TYPE aami_check_fetch_status gauge
aami_check_fetch_status 0

# HELP aami_check_fetch_timestamp_seconds Last check fetch timestamp
# TYPE aami_check_fetch_timestamp_seconds gauge
aami_check_fetch_timestamp_seconds $(date +%s)
EOF
        exit 1
    fi

    debug "Received checks JSON: $checks_json"

    # Process each check
    echo "$checks_json" | jq -c '.[]' | while read -r check; do
        checks_total=$((checks_total + 1))

        # Extract check details
        local check_name=$(echo "$check" | jq -r '.name')
        local script_content=$(echo "$check" | jq -r '.script_content')
        local script_hash=$(echo "$check" | jq -r '.script_hash')
        local check_config=$(echo "$check" | jq -r '.config // "{}"')

        info "Processing check: $check_name"

        # Save script
        local script_path=$(save_check_script "$check_name" "$script_content" "$script_hash")

        # Execute check
        if execute_check "$check_name" "$script_path" "$check_config"; then
            checks_success=$((checks_success + 1))
        else
            checks_failed=$((checks_failed + 1))
        fi
    done

    # Calculate duration
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    # Output summary metrics
    cat > "${TEXTFILE_DIR}/aami_status.prom" <<EOF
# HELP aami_check_fetch_status Check fetch status (1=success, 0=failed)
# TYPE aami_check_fetch_status gauge
aami_check_fetch_status 1

# HELP aami_check_fetch_timestamp_seconds Last check fetch timestamp
# TYPE aami_check_fetch_timestamp_seconds gauge
aami_check_fetch_timestamp_seconds $end_time

# HELP aami_check_execution_duration_seconds Check execution duration
# TYPE aami_check_execution_duration_seconds gauge
aami_check_execution_duration_seconds $duration

# HELP aami_checks_total Total number of checks configured
# TYPE aami_checks_total gauge
aami_checks_total $checks_total

# HELP aami_checks_success Number of successful checks
# TYPE aami_checks_success gauge
aami_checks_success $checks_success

# HELP aami_checks_failed Number of failed checks
# TYPE aami_checks_failed gauge
aami_checks_failed $checks_failed
EOF

    info "Check run completed: total=$checks_total, success=$checks_success, failed=$checks_failed, duration=${duration}s"
}

# Run main function
main

exit 0
