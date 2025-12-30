#!/bin/bash
# Health Check Script
# Basic system health check script

set -e

# Configuration (can be overridden via environment variables)
CHECK_DISK=${CHECK_DISK:-true}
CHECK_MEMORY=${CHECK_MEMORY:-true}
CHECK_LOAD=${CHECK_LOAD:-true}
TIMEOUT=${TIMEOUT:-30}

# Output format
output_json() {
    echo "{\"status\": \"$1\", \"message\": \"$2\", \"checks\": $3}"
}

# Initialize checks array
checks="[]"
status="healthy"
messages=()

# Check disk usage
if [ "$CHECK_DISK" = "true" ]; then
    disk_usage=$(df -h / | awk 'NR==2 {print $5}' | tr -d '%')
    if [ "$disk_usage" -gt 90 ]; then
        status="unhealthy"
        messages+=("Disk usage critical: ${disk_usage}%")
    elif [ "$disk_usage" -gt 80 ]; then
        if [ "$status" != "unhealthy" ]; then
            status="degraded"
        fi
        messages+=("Disk usage warning: ${disk_usage}%")
    fi
    checks=$(echo "$checks" | jq --arg name "disk" --arg value "${disk_usage}%" '. + [{"name": $name, "value": $value}]')
fi

# Check memory usage
if [ "$CHECK_MEMORY" = "true" ]; then
    mem_info=$(free | awk '/Mem:/ {printf "%.0f", $3/$2 * 100}')
    if [ "$mem_info" -gt 95 ]; then
        status="unhealthy"
        messages+=("Memory usage critical: ${mem_info}%")
    elif [ "$mem_info" -gt 85 ]; then
        if [ "$status" != "unhealthy" ]; then
            status="degraded"
        fi
        messages+=("Memory usage warning: ${mem_info}%")
    fi
    checks=$(echo "$checks" | jq --arg name "memory" --arg value "${mem_info}%" '. + [{"name": $name, "value": $value}]')
fi

# Check load average
if [ "$CHECK_LOAD" = "true" ]; then
    cpu_count=$(nproc)
    load_1m=$(cat /proc/loadavg | awk '{print $1}')
    load_ratio=$(echo "$load_1m $cpu_count" | awk '{printf "%.2f", $1/$2}')

    if [ "$(echo "$load_ratio > 2" | bc -l)" -eq 1 ]; then
        status="unhealthy"
        messages+=("Load average critical: ${load_1m} (ratio: ${load_ratio})")
    elif [ "$(echo "$load_ratio > 1" | bc -l)" -eq 1 ]; then
        if [ "$status" != "unhealthy" ]; then
            status="degraded"
        fi
        messages+=("Load average warning: ${load_1m} (ratio: ${load_ratio})")
    fi
    checks=$(echo "$checks" | jq --arg name "load" --arg value "$load_1m" '. + [{"name": $name, "value": $value}]')
fi

# Combine messages
if [ ${#messages[@]} -eq 0 ]; then
    message="All checks passed"
else
    message=$(IFS="; "; echo "${messages[*]}")
fi

# Output result
output_json "$status" "$message" "$checks"
