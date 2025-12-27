#!/usr/bin/env bash
set -euo pipefail

# Create a monitoring group via AAMI API
#
# Usage: ./create-group.sh <name> <namespace> <description> [parent_id]
#
# Example:
#   ./create-group.sh production environment "Production environment"
#   ./create-group.sh gpu-cluster infrastructure "GPU cluster" parent-group-id

# Configuration
AAMI_API_URL="${AAMI_API_URL:-http://localhost:8080/api/v1}"
AAMI_API_KEY="${AAMI_API_KEY:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print usage
usage() {
    cat <<EOF
Usage: $(basename "$0") <name> <namespace> <description> [parent_id]

Arguments:
    name        - Group name (required)
    namespace   - Namespace: infrastructure, logical, or environment (required)
    description - Group description (required)
    parent_id   - Parent group ID (optional)

Environment Variables:
    AAMI_API_URL - API base URL (default: http://localhost:8080/api/v1)
    AAMI_API_KEY - API key for authentication (if required)

Example:
    $(basename "$0") production environment "Production environment"
    $(basename "$0") gpu-cluster infrastructure "GPU cluster" parent-group-id

EOF
    exit 1
}

# Check arguments
if [ $# -lt 3 ]; then
    echo -e "${RED}Error: Missing required arguments${NC}" >&2
    usage
fi

NAME="$1"
NAMESPACE="$2"
DESCRIPTION="$3"
PARENT_ID="${4:-null}"

# Validate namespace
if [[ ! "$NAMESPACE" =~ ^(infrastructure|logical|environment)$ ]]; then
    echo -e "${RED}Error: Invalid namespace. Must be: infrastructure, logical, or environment${NC}" >&2
    exit 1
fi

# Build request body
if [ "$PARENT_ID" = "null" ]; then
    REQUEST_BODY=$(cat <<EOF
{
  "name": "$NAME",
  "namespace": "$NAMESPACE",
  "description": "$DESCRIPTION",
  "parent_id": null
}
EOF
)
else
    REQUEST_BODY=$(cat <<EOF
{
  "name": "$NAME",
  "namespace": "$NAMESPACE",
  "description": "$DESCRIPTION",
  "parent_id": "$PARENT_ID"
}
EOF
)
fi

# Build curl command
CURL_CMD="curl -s -w '\n%{http_code}' -X POST"
CURL_CMD="$CURL_CMD -H 'Content-Type: application/json'"

# Add authentication header if API key is set
if [ -n "$AAMI_API_KEY" ]; then
    CURL_CMD="$CURL_CMD -H 'Authorization: Bearer $AAMI_API_KEY'"
fi

CURL_CMD="$CURL_CMD -d '$REQUEST_BODY'"
CURL_CMD="$CURL_CMD $AAMI_API_URL/groups"

# Execute request
echo -e "${YELLOW}Creating group...${NC}"
RESPONSE=$(eval "$CURL_CMD")

# Parse response
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

# Check HTTP status
if [ "$HTTP_CODE" -eq 200 ] || [ "$HTTP_CODE" -eq 201 ]; then
    echo -e "${GREEN}Success!${NC}"
    echo "$BODY" | jq .

    # Extract and display group ID
    GROUP_ID=$(echo "$BODY" | jq -r .id)
    echo ""
    echo -e "${GREEN}Group ID: $GROUP_ID${NC}"
    echo "Save this ID for registering targets to this group."
else
    echo -e "${RED}Failed! HTTP $HTTP_CODE${NC}" >&2
    echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
    exit 1
fi
