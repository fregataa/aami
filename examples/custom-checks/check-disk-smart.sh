#!/usr/bin/env bash
#
# Disk SMART Health Check
#
# This script checks the SMART health status of all available disks.
# Requires smartmontools package to be installed.
#
# Expected JSON input format:
# {
#   "devices": ["/dev/sda", "/dev/sdb"]  # Optional, defaults to all /dev/sd* devices
# }
#
# Output format (Prometheus):
# disk_smart_health{device="sda"} 1
# disk_smart_health{device="sdb"} 0
# disk_smart_health_info{device="sda",status="PASSED",temperature="35"} 1

set -euo pipefail

# Check if smartctl is available
if ! command -v smartctl &> /dev/null; then
    echo "# HELP disk_smart_health Disk SMART health status (1=healthy, 0=failing, -1=unavailable)"
    echo "# TYPE disk_smart_health gauge"
    echo "disk_smart_health{device=\"none\"} -1"
    echo "# smartctl not installed"
    exit 0
fi

# Read JSON config from stdin (passed by dynamic-check.sh)
if [ $# -eq 0 ]; then
    config=$(cat)
else
    config="$1"
fi

# Extract devices from JSON, or use default
devices=$(echo "$config" | jq -r '.devices[]?' 2>/dev/null)
if [ -z "$devices" ]; then
    # Default: check all /dev/sd* devices
    devices=$(ls /dev/sd[a-z] 2>/dev/null || echo "")
fi

# Output Prometheus metrics header
echo "# HELP disk_smart_health Disk SMART health status (1=healthy, 0=failing, -1=not_supported)"
echo "# TYPE disk_smart_health gauge"
echo "# HELP disk_smart_temperature_celsius Disk temperature in Celsius"
echo "# TYPE disk_smart_temperature_celsius gauge"
echo "# HELP disk_smart_reallocated_sectors Reallocated sectors count"
echo "# TYPE disk_smart_reallocated_sectors gauge"

# Check each disk
if [ -z "$devices" ]; then
    echo "disk_smart_health{device=\"none\"} -1"
else
    for disk in $devices; do
        if [ ! -e "$disk" ]; then
            continue
        fi

        device_name=$(basename "$disk")

        # Check SMART support
        if ! smartctl -i "$disk" &>/dev/null; then
            echo "disk_smart_health{device=\"$device_name\"} -1"
            continue
        fi

        # Get SMART health status
        health_output=$(smartctl -H "$disk" 2>/dev/null || echo "")

        if echo "$health_output" | grep -q "PASSED"; then
            health_value=1
        elif echo "$health_output" | grep -q "FAILED"; then
            health_value=0
        else
            health_value=-1
        fi

        echo "disk_smart_health{device=\"$device_name\"} $health_value"

        # Get additional SMART attributes
        smart_data=$(smartctl -A "$disk" 2>/dev/null || echo "")

        # Temperature
        temp=$(echo "$smart_data" | awk '/Temperature_Celsius/ {print $10}' | head -1)
        if [ -n "$temp" ] && [ "$temp" -ge 0 ] 2>/dev/null; then
            echo "disk_smart_temperature_celsius{device=\"$device_name\"} $temp"
        fi

        # Reallocated Sectors
        reallocated=$(echo "$smart_data" | awk '/Reallocated_Sector/ {print $10}' | head -1)
        if [ -n "$reallocated" ] && [ "$reallocated" -ge 0 ] 2>/dev/null; then
            echo "disk_smart_reallocated_sectors{device=\"$device_name\"} $reallocated"
        fi
    done
fi
