#!/bin/bash
# Process Monitor Script
# Monitor specific processes and optionally restart them

set -e

# Configuration (can be overridden via environment variables)
PROCESSES=${PROCESSES:-"nginx,postgres"}
RESTART_ON_FAILURE=${RESTART_ON_FAILURE:-false}

# Parse PROCESSES into array
IFS=',' read -ra PROCESS_ARRAY <<< "$PROCESSES"

# Initialize output
output="{\"timestamp\": \"$(date -Iseconds)\", \"processes\": [], \"actions\": []}"
all_running=true

# Check each process
for process in "${PROCESS_ARRAY[@]}"; do
    process=$(echo "$process" | tr -d ' ')

    if [ -z "$process" ]; then
        continue
    fi

    # Check if process is running
    pid=$(pgrep -x "$process" 2>/dev/null | head -1 || true)

    if [ -n "$pid" ]; then
        # Process is running, get details
        cpu=$(ps -p "$pid" -o %cpu= 2>/dev/null | tr -d ' ' || echo "0")
        mem=$(ps -p "$pid" -o %mem= 2>/dev/null | tr -d ' ' || echo "0")
        uptime=$(ps -p "$pid" -o etime= 2>/dev/null | tr -d ' ' || echo "unknown")

        proc_json="{\"name\": \"$process\", \"status\": \"running\", \"pid\": $pid, \"cpu\": \"$cpu%\", \"memory\": \"$mem%\", \"uptime\": \"$uptime\"}"
    else
        # Process is not running
        all_running=false
        proc_json="{\"name\": \"$process\", \"status\": \"stopped\", \"pid\": null}"

        # Optionally restart
        if [ "$RESTART_ON_FAILURE" = "true" ]; then
            # Try to restart using systemctl if available
            if command -v systemctl &> /dev/null; then
                if systemctl restart "$process" 2>/dev/null; then
                    action="{\"process\": \"$process\", \"action\": \"restarted\", \"success\": true}"
                else
                    action="{\"process\": \"$process\", \"action\": \"restart_failed\", \"success\": false}"
                fi
                output=$(echo "$output" | jq --argjson a "$action" '.actions += [$a]')
            fi
        fi
    fi

    output=$(echo "$output" | jq --argjson p "$proc_json" '.processes += [$p]')
done

# Set overall status
if [ "$all_running" = true ]; then
    output=$(echo "$output" | jq '.status = "ok"')
else
    output=$(echo "$output" | jq '.status = "warning"')
fi

echo "$output"
