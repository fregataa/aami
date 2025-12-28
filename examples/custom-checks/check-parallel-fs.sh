#!/usr/bin/env bash
#
# Parallel Filesystem Latency Check
#
# This script checks the latency and accessibility of parallel filesystems
# (Lustre, GPFS, BeeGFS, etc.) by performing simple I/O operations.
#
# Expected JSON input format:
# {
#   "filesystems": [
#     {"path": "/mnt/lustre", "type": "lustre"},
#     {"path": "/mnt/gpfs", "type": "gpfs"},
#     {"path": "/mnt/beegfs", "type": "beegfs"}
#   ],
#   "timeout_seconds": 5,
#   "test_size_mb": 1
# }
#
# Output format (Prometheus):
# parallel_fs_latency_seconds{path="/mnt/lustre",type="lustre",operation="read"} 0.123
# parallel_fs_accessible{path="/mnt/lustre",type="lustre"} 1

set -euo pipefail

# Read JSON config from stdin (passed by dynamic-check.sh)
if [ $# -eq 0 ]; then
    config=$(cat)
else
    config="$1"
fi

# Extract configuration
filesystems=$(echo "$config" | jq -c '.filesystems[]?' 2>/dev/null)
timeout_seconds=$(echo "$config" | jq -r '.timeout_seconds // 5' 2>/dev/null)
test_size_mb=$(echo "$config" | jq -r '.test_size_mb // 1' 2>/dev/null)

# Output Prometheus metrics header
echo "# HELP parallel_fs_accessible Parallel filesystem accessibility (1=accessible, 0=not_accessible)"
echo "# TYPE parallel_fs_accessible gauge"
echo "# HELP parallel_fs_latency_seconds Parallel filesystem latency in seconds"
echo "# TYPE parallel_fs_latency_seconds gauge"
echo "# HELP parallel_fs_throughput_mbps Parallel filesystem throughput in MB/s"
echo "# TYPE parallel_fs_throughput_mbps gauge"

# Check each filesystem
if [ -z "$filesystems" ]; then
    echo "parallel_fs_accessible{path=\"none\",type=\"none\"} -1"
else
    while IFS= read -r fs_config; do
        if [ -z "$fs_config" ]; then
            continue
        fi

        fs_path=$(echo "$fs_config" | jq -r '.path')
        fs_type=$(echo "$fs_config" | jq -r '.type // "unknown"')

        # Check if filesystem is mounted and accessible
        if [ ! -d "$fs_path" ]; then
            echo "parallel_fs_accessible{path=\"$fs_path\",type=\"$fs_type\"} 0"
            continue
        fi

        if ! mountpoint -q "$fs_path" 2>/dev/null; then
            echo "parallel_fs_accessible{path=\"$fs_path\",type=\"$fs_type\"} 0"
            continue
        fi

        # Filesystem is accessible
        echo "parallel_fs_accessible{path=\"$fs_path\",type=\"$fs_type\"} 1"

        # Perform I/O latency test (write operation)
        test_file="$fs_path/.aami_health_check_$$"

        # Write test with timeout
        write_start=$(date +%s.%N)
        if timeout "${timeout_seconds}s" dd if=/dev/zero of="$test_file" bs=1M count="$test_size_mb" conv=fsync 2>/dev/null; then
            write_end=$(date +%s.%N)
            write_latency=$(echo "$write_end - $write_start" | bc -l)
            write_throughput=$(echo "scale=2; $test_size_mb / $write_latency" | bc -l)

            echo "parallel_fs_latency_seconds{path=\"$fs_path\",type=\"$fs_type\",operation=\"write\"} $write_latency"
            echo "parallel_fs_throughput_mbps{path=\"$fs_path\",type=\"$fs_type\",operation=\"write\"} $write_throughput"
        else
            # Write timeout or error
            echo "parallel_fs_latency_seconds{path=\"$fs_path\",type=\"$fs_type\",operation=\"write\"} -1"
        fi

        # Read test with timeout (if write succeeded)
        if [ -f "$test_file" ]; then
            read_start=$(date +%s.%N)
            if timeout "${timeout_seconds}s" dd if="$test_file" of=/dev/null bs=1M 2>/dev/null; then
                read_end=$(date +%s.%N)
                read_latency=$(echo "$read_end - $read_start" | bc -l)
                read_throughput=$(echo "scale=2; $test_size_mb / $read_latency" | bc -l)

                echo "parallel_fs_latency_seconds{path=\"$fs_path\",type=\"$fs_type\",operation=\"read\"} $read_latency"
                echo "parallel_fs_throughput_mbps{path=\"$fs_path\",type=\"$fs_type\",operation=\"read\"} $read_throughput"
            else
                # Read timeout or error
                echo "parallel_fs_latency_seconds{path=\"$fs_path\",type=\"$fs_type\",operation=\"read\"} -1"
            fi

            # Cleanup test file
            rm -f "$test_file"
        fi

        # Check for Lustre-specific metrics
        if [ "$fs_type" = "lustre" ] && command -v lfs &> /dev/null; then
            # Get Lustre filesystem stats
            df_output=$(lfs df "$fs_path" 2>/dev/null || echo "")
            if [ -n "$df_output" ]; then
                # You can extend this to parse Lustre-specific metrics
                # For now, we just mark that Lustre tools are available
                echo "# Lustre filesystem detected at $fs_path"
            fi
        fi

        # Check for GPFS-specific metrics
        if [ "$fs_type" = "gpfs" ] && command -v mmlsfs &> /dev/null; then
            # Get GPFS filesystem stats
            gpfs_output=$(mmlsfs "$fs_path" 2>/dev/null || echo "")
            if [ -n "$gpfs_output" ]; then
                # You can extend this to parse GPFS-specific metrics
                echo "# GPFS filesystem detected at $fs_path"
            fi
        fi

        # Check for BeeGFS-specific metrics
        if [ "$fs_type" = "beegfs" ] && command -v beegfs-ctl &> /dev/null; then
            # Get BeeGFS filesystem stats
            beegfs_output=$(beegfs-ctl --getentryinfo "$fs_path" 2>/dev/null || echo "")
            if [ -n "$beegfs_output" ]; then
                # You can extend this to parse BeeGFS-specific metrics
                echo "# BeeGFS filesystem detected at $fs_path"
            fi
        fi

    done <<< "$filesystems"
fi
