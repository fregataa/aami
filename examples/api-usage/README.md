# API Usage Examples

This directory contains example scripts demonstrating common AAMI API usage patterns.

## Available Examples

### Basic Operations
- **`create-group.sh`** - Create a monitoring group
- **`register-target.sh`** - Register a monitoring target
- **`list-targets.sh`** - List all registered targets
- **`apply-alert-rule.sh`** - Apply alert rule to a group

### Advanced Usage
- **`bulk-register.sh`** - Bulk register multiple targets from CSV
- **`update-labels.sh`** - Update target labels in batch
- **`policy-trace.sh`** - Trace alert rule policy inheritance

### Python Examples
- **`api_client.py`** - Python API client library
- **`register_targets.py`** - Register targets using Python
- **`export_config.py`** - Export configuration to JSON/YAML

## Prerequisites

All bash scripts require:
- `curl` - HTTP client
- `jq` - JSON processor

Install dependencies:

```bash
# macOS
brew install curl jq

# Ubuntu/Debian
sudo apt-get install curl jq

# RHEL/CentOS
sudo yum install curl jq
```

Python scripts require:
```bash
pip install requests pyyaml
```

## Configuration

Set environment variables:

```bash
export AAMI_API_URL="http://localhost:8080/api/v1"
export AAMI_API_KEY="your-api-key"  # If authentication is enabled
```

Or create a `.env` file:

```env
AAMI_API_URL=http://localhost:8080/api/v1
AAMI_API_KEY=your-api-key
```

## Usage Examples

### Create Group and Register Target

```bash
# 1. Create a group
./create-group.sh production environment "Production environment"

# 2. Save the returned group ID
GROUP_ID="550e8400-e29b-41d4-a716-446655440000"

# 3. Register a target
./register-target.sh \
  "gpu-node-01.example.com" \
  "10.0.1.10" \
  "$GROUP_ID" \
  "A100" \
  "8"
```

### Bulk Registration

```bash
# Create CSV file with target information
cat > targets.csv <<EOF
hostname,ip_address,group_id,gpu_model,gpu_count
gpu-node-01.example.com,10.0.1.10,GROUP_ID,A100,8
gpu-node-02.example.com,10.0.1.11,GROUP_ID,A100,8
gpu-node-03.example.com,10.0.1.12,GROUP_ID,A100,8
EOF

# Bulk register
./bulk-register.sh targets.csv
```

### Apply Alert Rules

```bash
# Apply high CPU alert with 70% threshold
./apply-alert-rule.sh \
  "$GROUP_ID" \
  "HighCPUUsage" \
  70 \
  "5m"
```

### Trace Policy Inheritance

```bash
# See which groups contributed to final alert configuration
./policy-trace.sh "TARGET_ID" "HighCPUUsage"
```

## Python Client Example

```python
from api_client import AAMIClient

# Initialize client
client = AAMIClient(base_url="http://localhost:8080/api/v1")

# Create group
group = client.create_group(
    name="production",
    namespace="environment",
    description="Production environment"
)

# Register target
target = client.create_target(
    hostname="gpu-node-01.example.com",
    ip_address="10.0.1.10",
    primary_group_id=group["id"],
    exporters=[
        {"type": "node_exporter", "port": 9100, "enabled": True},
        {"type": "dcgm_exporter", "port": 9400, "enabled": True}
    ],
    labels={
        "gpu_model": "A100",
        "gpu_count": "8"
    }
)

print(f"Target registered: {target['id']}")
```

## Testing

Test scripts without making actual changes:

```bash
# Dry run mode (add --dry-run flag if supported)
DRY_RUN=1 ./register-target.sh ...

# Test against mock server
export AAMI_API_URL="http://localhost:8888/api/v1"
```

## Troubleshooting

### Connection Refused

```bash
# Check if Config Server is running
curl http://localhost:8080/api/v1/health

# Check Docker container
docker-compose ps config-server
```

### Authentication Error

```bash
# Verify API key
echo $AAMI_API_KEY

# Test without authentication (dev mode)
unset AAMI_API_KEY
```

### JSON Parsing Error

```bash
# Install jq
which jq || brew install jq

# Check JSON response
curl http://localhost:8080/api/v1/groups | jq .
```

## Contributing

When adding new examples:

1. Follow existing script structure
2. Add error handling
3. Include usage documentation
4. Test with different inputs
5. Add comments explaining logic

---

For detailed API documentation, see [API Reference](../../docs/en/API.md).
