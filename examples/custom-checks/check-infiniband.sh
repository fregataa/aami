#!/usr/bin/env bash
#
# InfiniBand Link State Check
#
# This script checks the state of InfiniBand ports on the system.
# Reads port information from /sys/class/infiniband.
#
# Expected JSON input format:
# {
#   "devices": ["mlx5_0", "mlx5_1"]  # Optional, defaults to all devices
# }
#
# Output format (Prometheus):
# infiniband_link_state{device="mlx5_0",port="1"} 4
# infiniband_link_rate_gbps{device="mlx5_0",port="1"} 100

set -euo pipefail

# Check if InfiniBand devices exist
if [ ! -d "/sys/class/infiniband" ]; then
    echo "# HELP infiniband_link_state InfiniBand link state (4=active, 0=down, -1=not_present)"
    echo "# TYPE infiniband_link_state gauge"
    echo "infiniband_link_state{device=\"none\",port=\"0\"} -1"
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
    # Default: check all InfiniBand devices
    devices=$(ls /sys/class/infiniband/ 2>/dev/null || echo "")
fi

# Output Prometheus metrics header
echo "# HELP infiniband_link_state InfiniBand link state (4=active, 3=armed, 2=init, 1=down, 0=polling)"
echo "# TYPE infiniband_link_state gauge"
echo "# HELP infiniband_link_rate_gbps InfiniBand link rate in Gbps"
echo "# TYPE infiniband_link_rate_gbps gauge"
echo "# HELP infiniband_port_rcv_data_bytes InfiniBand port receive data bytes"
echo "# TYPE infiniband_port_rcv_data_bytes counter"
echo "# HELP infiniband_port_xmit_data_bytes InfiniBand port transmit data bytes"
echo "# TYPE infiniband_port_xmit_data_bytes counter"

# Check each device
if [ -z "$devices" ]; then
    echo "infiniband_link_state{device=\"none\",port=\"0\"} -1"
else
    for device in $devices; do
        device_path="/sys/class/infiniband/$device"
        if [ ! -d "$device_path" ]; then
            continue
        fi

        # Check all ports for this device
        for port_path in "$device_path"/ports/*; do
            if [ ! -d "$port_path" ]; then
                continue
            fi

            port_num=$(basename "$port_path")

            # Get link state
            if [ -f "$port_path/state" ]; then
                state_text=$(cat "$port_path/state")

                # Map state text to numeric value
                case "$state_text" in
                    *ACTIVE*|*4:*) state_value=4 ;;
                    *ARMED*|*3:*) state_value=3 ;;
                    *INIT*|*2:*) state_value=2 ;;
                    *DOWN*|*1:*) state_value=1 ;;
                    *POLLING*|*0:*) state_value=0 ;;
                    *) state_value=-1 ;;
                esac

                echo "infiniband_link_state{device=\"$device\",port=\"$port_num\"} $state_value"
            fi

            # Get link rate (in Gbps)
            if [ -f "$port_path/rate" ]; then
                rate_text=$(cat "$port_path/rate" 2>/dev/null || echo "")
                # Extract numeric value (format: "100 Gb/sec" or "56 Gb/sec")
                rate_value=$(echo "$rate_text" | grep -oP '\d+' | head -1)
                if [ -n "$rate_value" ]; then
                    echo "infiniband_link_rate_gbps{device=\"$device\",port=\"$port_num\"} $rate_value"
                fi
            fi

            # Get port counters
            counters_path="$port_path/counters"
            if [ -d "$counters_path" ]; then
                # Receive data
                if [ -f "$counters_path/port_rcv_data" ]; then
                    rcv_data=$(cat "$counters_path/port_rcv_data" 2>/dev/null || echo "0")
                    # Convert to bytes (data is in 4-byte words)
                    rcv_bytes=$((rcv_data * 4))
                    echo "infiniband_port_rcv_data_bytes{device=\"$device\",port=\"$port_num\"} $rcv_bytes"
                fi

                # Transmit data
                if [ -f "$counters_path/port_xmit_data" ]; then
                    xmit_data=$(cat "$counters_path/port_xmit_data" 2>/dev/null || echo "0")
                    # Convert to bytes (data is in 4-byte words)
                    xmit_bytes=$((xmit_data * 4))
                    echo "infiniband_port_xmit_data_bytes{device=\"$device\",port=\"$port_num\"} $xmit_bytes"
                fi
            fi
        done
    done
fi
