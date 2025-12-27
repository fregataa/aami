#!/usr/bin/env bash
set -euo pipefail

# Register a monitoring target via AAMI API
#
# Usage: ./register-target.sh <hostname> <ip> <group_id> [gpu_model] [gpu_count]
#
# Example:
#   ./register-target.sh gpu-node-01.example.com 10.0.1.10 group-id A100 8

# Configuration
AAMI_API_URL="${AAMI_API_URL:-http://localhost:8080/api/v1}"
AAMI_API_KEY="${AAMI_API_KEY:-}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Print usage
usage() {
    cat <<EOF
Usage: $(basename "$0") <hostname> <ip> <group_id> [gpu_model] [gpu_count]

Arguments:
    hostname   - Target hostname (required)
    ip         - Target IP address (required)
    group_id   - Primary group ID (required)
    gpu_model  - GPU model (optional, e.g., A100, H100)
    gpu_count  - Number of GPUs (optional, e.g., 8)

Environment Variables:
    AAMI_API_URL - API base URL (default: http://localhost:8080/api/v1)
    AAMI_API_KEY - API key for authentication (if required)

Example:
    $(basename "$0") gpu-node-01.example.com 10.0.1.10 group-id A100 8

EOF
    exit 1
}

# Check arguments
if [ $# -lt 3 ]; then
    echo -e "${RED}Error: Missing required arguments${NC}" >&2
    usage
fi

HOSTNAME="$1"
IP_ADDRESS="$2"
GROUP_ID="$3"
GPU_MODEL="${4:-}"
GPU_COUNT="${5:-}"

# Build labels JSON
if [ -n "$GPU_MODEL" ] && [ -n "$GPU_COUNT" ]; then
    LABELS=$(cat <<EOF
    "gpu_model": "$GPU_MODEL",
    "gpu_count": "$GPU_COUNT"
EOF
)
else
    LABELS=""
fi

# Build request body
REQUEST_BODY=$(cat <<EOF
{
  "hostname": "$HOSTNAME",
  "ip_address": "$IP_ADDRESS",
  "primary_group_id": "$GROUP_ID",
  "exporters": [
    {
      "type": "node_exporter",
      "port": 9100,
      "enabled": true,
      "scrape_interval": "15s",
      "scrape_timeout": "10s"
    }
EOF
)

# Add DCGM exporter if GPU info is provided
if [ -n "$GPU_MODEL" ]; then
    REQUEST_BODY+=$(cat <<EOF
,
    {
      "type": "dcgm_exporter",
      "port": 9400,
      "enabled": true,
      "scrape_interval": "30s",
      "scrape_timeout": "10s"
    }
EOF
)
fi

REQUEST_BODY+=$(cat <<EOF

  ]
EOF
)

# Add labels if provided
if [ -n "$LABELS" ]; then
    REQUEST_BODY+=$(cat <<EOF
,
  "labels": {
    $LABELS
  }
EOF
)
fi

REQUEST_BODY+="}"

# Build curl command
CURL_CMD="curl -s -w '\n%{http_code}' -X POST"
CURL_CMD="$CURL_CMD -H 'Content-Type: application/json'"

if [ -n "$AAMI_API_KEY" ]; then
    CURL_CMD="$CURL_CMD -H 'Authorization: Bearer $AAMI_API_KEY'"
fi

CURL_CMD="$CURL_CMD -d '$REQUEST_BODY'"
CURL_CMD="$CURL_CMD $AAMI_API_URL/targets"

# Execute request
echo -e "${YELLOW}Registering target $HOSTNAME...${NC}"
RESPONSE=$(eval "$CURL_CMD")

# Parse response
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

# Check HTTP status
if [ "$HTTP_CODE" -eq 200 ] || [ "$HTTP_CODE" -eq 201 ]; then
    echo -e "${GREEN}Success!${NC}"
    echo "$BODY" | jq .

    # Extract and display target ID
    TARGET_ID=$(echo "$BODY" | jq -r .id)
    echo ""
    echo -e "${GREEN}Target ID: $TARGET_ID${NC}"
    echo "Target will appear in Prometheus within 30 seconds."
else
    echo -e "${RED}Failed! HTTP $HTTP_CODE${NC}" >&2
    echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
    exit 1
fi
