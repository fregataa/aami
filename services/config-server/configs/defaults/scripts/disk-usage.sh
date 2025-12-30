#!/bin/bash
# Disk Usage Report Script
# Detailed disk usage report with threshold alerting

set -e

# Configuration (can be overridden via environment variables)
THRESHOLD_PERCENT=${THRESHOLD_PERCENT:-80}
INCLUDE_MOUNTS=${INCLUDE_MOUNTS:-"/,/home,/var"}

# Parse INCLUDE_MOUNTS into array
IFS=',' read -ra MOUNT_ARRAY <<< "$INCLUDE_MOUNTS"

# Initialize output
output="{\"timestamp\": \"$(date -Iseconds)\", \"filesystems\": [], \"warnings\": []}"
warnings=()

# Check each mount point
for mount in "${MOUNT_ARRAY[@]}"; do
    mount=$(echo "$mount" | tr -d ' ')

    if [ ! -d "$mount" ]; then
        continue
    fi

    # Get disk info
    disk_info=$(df -P "$mount" 2>/dev/null | tail -1)
    if [ -z "$disk_info" ]; then
        continue
    fi

    filesystem=$(echo "$disk_info" | awk '{print $1}')
    size=$(echo "$disk_info" | awk '{print $2}')
    used=$(echo "$disk_info" | awk '{print $3}')
    avail=$(echo "$disk_info" | awk '{print $4}')
    use_percent=$(echo "$disk_info" | awk '{print $5}' | tr -d '%')

    # Convert to human readable
    size_hr=$(numfmt --to=iec --suffix=B $((size * 1024)) 2>/dev/null || echo "${size}K")
    used_hr=$(numfmt --to=iec --suffix=B $((used * 1024)) 2>/dev/null || echo "${used}K")
    avail_hr=$(numfmt --to=iec --suffix=B $((avail * 1024)) 2>/dev/null || echo "${avail}K")

    # Add to output
    fs_json="{\"mount\": \"$mount\", \"filesystem\": \"$filesystem\", \"size\": \"$size_hr\", \"used\": \"$used_hr\", \"available\": \"$avail_hr\", \"use_percent\": $use_percent}"
    output=$(echo "$output" | jq --argjson fs "$fs_json" '.filesystems += [$fs]')

    # Check threshold
    if [ "$use_percent" -ge "$THRESHOLD_PERCENT" ]; then
        warning="Mount $mount is ${use_percent}% full (threshold: ${THRESHOLD_PERCENT}%)"
        warnings+=("$warning")
        output=$(echo "$output" | jq --arg w "$warning" '.warnings += [$w]')
    fi
done

# Set status based on warnings
if [ ${#warnings[@]} -gt 0 ]; then
    output=$(echo "$output" | jq '.status = "warning"')
else
    output=$(echo "$output" | jq '.status = "ok"')
fi

echo "$output"
