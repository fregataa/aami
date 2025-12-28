#!/usr/bin/env bash
#
# Mount Point Accessibility Check
#
# This script checks if specified mount points are accessible and writable.
# It reads configuration from stdin in JSON format and outputs Prometheus metrics.
#
# Expected JSON input format:
# {
#   "mount_points": ["/mnt/data", "/mnt/backup"]
# }
#
# Output format (Prometheus):
# mount_check{path="/mnt/data"} 1
# mount_check{path="/mnt/backup"} 0

set -euo pipefail

# Read JSON config from stdin (passed by dynamic-check.sh)
if [ $# -eq 0 ]; then
    # No arguments, read from stdin
    config=$(cat)
else
    # Config passed as argument
    config="$1"
fi

# Extract mount points from JSON
mount_points=$(echo "$config" | jq -r '.mount_points[]' 2>/dev/null || echo "")

# Output Prometheus metrics header
echo "# HELP mount_check Mount point accessibility (1=ok, 0=fail)"
echo "# TYPE mount_check gauge"

# Check each mount point
if [ -z "$mount_points" ]; then
    # No mount points configured, output a metric indicating no checks
    echo "mount_check{path=\"none\"} 1"
else
    while IFS= read -r mount_point; do
        if [ -z "$mount_point" ]; then
            continue
        fi

        # Check if mount point exists, is mounted, and is writable
        if mountpoint -q "$mount_point" 2>/dev/null && [ -w "$mount_point" ]; then
            echo "mount_check{path=\"$mount_point\"} 1"
        else
            echo "mount_check{path=\"$mount_point\"} 0"
        fi
    done <<< "$mount_points"
fi
