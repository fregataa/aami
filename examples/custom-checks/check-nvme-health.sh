#!/usr/bin/env bash
#
# NVMe Health Check
#
# This script checks the health status of NVMe devices using nvme-cli.
# Requires nvme-cli package to be installed.
#
# Expected JSON input format:
# {
#   "devices": ["/dev/nvme0n1", "/dev/nvme1n1"]  # Optional, defaults to all /dev/nvme*n1 devices
# }
#
# Output format (Prometheus):
# nvme_health{device="nvme0n1"} 1
# nvme_temperature_celsius{device="nvme0n1"} 45
# nvme_available_spare_percent{device="nvme0n1"} 100
# nvme_percentage_used{device="nvme0n1"} 5

set -euo pipefail

# Check if nvme-cli is available
if ! command -v nvme &> /dev/null; then
    echo "# HELP nvme_health NVMe health status (1=healthy, 0=failing, -1=unavailable)"
    echo "# TYPE nvme_health gauge"
    echo "nvme_health{device=\"none\"} -1"
    echo "# nvme-cli not installed"
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
    # Default: check all /dev/nvme*n1 devices (namespace 1 of each controller)
    devices=$(ls /dev/nvme*n1 2>/dev/null || echo "")
fi

# Output Prometheus metrics header
echo "# HELP nvme_health NVMe health status (1=healthy, 0=critical_warning, -1=not_available)"
echo "# TYPE nvme_health gauge"
echo "# HELP nvme_temperature_celsius NVMe temperature in Celsius"
echo "# TYPE nvme_temperature_celsius gauge"
echo "# HELP nvme_available_spare_percent NVMe available spare percentage"
echo "# TYPE nvme_available_spare_percent gauge"
echo "# HELP nvme_percentage_used NVMe percentage used"
echo "# TYPE nvme_percentage_used gauge"
echo "# HELP nvme_data_units_read NVMe data units read (512-byte units)"
echo "# TYPE nvme_data_units_read counter"
echo "# HELP nvme_data_units_written NVMe data units written (512-byte units)"
echo "# TYPE nvme_data_units_written counter"
echo "# HELP nvme_power_cycles NVMe power cycles"
echo "# TYPE nvme_power_cycles counter"
echo "# HELP nvme_power_on_hours NVMe power on hours"
echo "# TYPE nvme_power_on_hours counter"

# Check each NVMe device
if [ -z "$devices" ]; then
    echo "nvme_health{device=\"none\"} -1"
else
    for device in $devices; do
        if [ ! -e "$device" ]; then
            continue
        fi

        device_name=$(basename "$device")

        # Get SMART log
        smart_output=$(nvme smart-log "$device" 2>/dev/null || echo "")

        if [ -z "$smart_output" ]; then
            echo "nvme_health{device=\"$device_name\"} -1"
            continue
        fi

        # Critical Warning (0 = healthy, >0 = warning/critical)
        critical_warning=$(echo "$smart_output" | awk '/^critical_warning/ {print $3}')
        if [ -n "$critical_warning" ]; then
            if [ "$critical_warning" = "0" ] || [ "$critical_warning" = "0x0" ]; then
                health_value=1
            else
                health_value=0
            fi
            echo "nvme_health{device=\"$device_name\"} $health_value"
        fi

        # Temperature (Kelvin to Celsius: K - 273)
        temperature_k=$(echo "$smart_output" | awk '/^temperature/ {print $3}')
        if [ -n "$temperature_k" ] && [ "$temperature_k" -gt 0 ] 2>/dev/null; then
            temperature_c=$((temperature_k - 273))
            echo "nvme_temperature_celsius{device=\"$device_name\"} $temperature_c"
        fi

        # Available Spare
        available_spare=$(echo "$smart_output" | awk '/^available_spare[^_]/ {print $3}' | tr -d '%')
        if [ -n "$available_spare" ] && [ "$available_spare" -ge 0 ] 2>/dev/null; then
            echo "nvme_available_spare_percent{device=\"$device_name\"} $available_spare"
        fi

        # Percentage Used
        percentage_used=$(echo "$smart_output" | awk '/^percentage_used/ {print $3}' | tr -d '%')
        if [ -n "$percentage_used" ] && [ "$percentage_used" -ge 0 ] 2>/dev/null; then
            echo "nvme_percentage_used{device=\"$device_name\"} $percentage_used"
        fi

        # Data Units Read (each unit is 1000 512-byte blocks = 512KB)
        data_units_read=$(echo "$smart_output" | awk '/^data_units_read/ {print $3}' | tr -d ',')
        if [ -n "$data_units_read" ] && [ "$data_units_read" -ge 0 ] 2>/dev/null; then
            echo "nvme_data_units_read{device=\"$device_name\"} $data_units_read"
        fi

        # Data Units Written
        data_units_written=$(echo "$smart_output" | awk '/^data_units_written/ {print $3}' | tr -d ',')
        if [ -n "$data_units_written" ] && [ "$data_units_written" -ge 0 ] 2>/dev/null; then
            echo "nvme_data_units_written{device=\"$device_name\"} $data_units_written"
        fi

        # Power Cycles
        power_cycles=$(echo "$smart_output" | awk '/^power_cycles/ {print $3}' | tr -d ',')
        if [ -n "$power_cycles" ] && [ "$power_cycles" -ge 0 ] 2>/dev/null; then
            echo "nvme_power_cycles{device=\"$device_name\"} $power_cycles"
        fi

        # Power On Hours
        power_on_hours=$(echo "$smart_output" | awk '/^power_on_hours/ {print $3}' | tr -d ',')
        if [ -n "$power_on_hours" ] && [ "$power_on_hours" -ge 0 ] 2>/dev/null; then
            echo "nvme_power_on_hours{device=\"$device_name\"} $power_on_hours"
        fi
    done
fi
