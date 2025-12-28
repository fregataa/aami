# AAMI CLI Quick Reference

## Installation

```bash
go build -o aami cmd/aami/main.go
sudo cp aami /usr/local/bin/
aami config init
```

## Configuration

```bash
aami config set server http://localhost:8080
aami config set output json
aami config view
```

## Namespaces

```bash
# Create
aami namespace create --name=prod --priority=100

# List
aami namespace list

# Get
aami namespace get <id>
aami namespace get --name=prod

# Update
aami namespace update <id> --description="Updated"

# Delete
aami namespace delete <id>
```

## Groups

```bash
# Create
aami group create --name=web --namespace=<ns-id>
aami group create --name=web-prod --namespace=<ns-id> --parent=<parent-id>

# List
aami group list
aami group list --namespace=<ns-id>

# Get
aami group get <id>

# Hierarchy
aami group children <id>
aami group ancestors <id>

# Update
aami group update <id> --description="Updated"

# Delete
aami group delete <id>
```

## Targets

```bash
# Create
aami target create --hostname=web-01 --ip=10.0.1.100 --group=<group-id>
aami target create --hostname=web-01 --ip=10.0.1.100 --groups=<id1>,<id2>

# List
aami target list
aami target list --group=<group-id>

# Get
aami target get <id>
aami target get --hostname=web-01

# Update
aami target update <id> --hostname=web-02

# Operations
aami target status <id> --status=active
aami target heartbeat <id>

# Delete
aami target delete <id>
```

## Bootstrap Tokens

```bash
# Create
aami bootstrap-token create --name=prod-token --max-uses=10 --expires=7d
aami bootstrap-token create --name=staging --max-uses=50 --expires=2025-12-31

# List
aami bootstrap-token list

# Get
aami bootstrap-token get <id>

# Update
aami bootstrap-token update <id> --max-uses=20 --expires=30d

# Validate
aami bootstrap-token validate <token-string>

# Register node
aami bootstrap-token register \
  --token=<token> \
  --hostname=web-01 \
  --ip=10.0.1.100 \
  --group=<group-id>

# Register with auto-hostname
aami bootstrap-token register \
  --token=<token> \
  --ip=$(hostname -I | awk '{print $1}')

# Delete
aami bootstrap-token delete <id>
```

## Output Formats

```bash
# Table (default)
aami namespace list

# JSON
aami namespace list -o json

# YAML
aami namespace list -o yaml

# With jq
aami namespace list -o json | jq -r '.[].name'
```

## Global Flags

```bash
-s, --server string   Server URL
-o, --output string   Output format (table|json|yaml)
-h, --help           Help
```

## Common Workflows

### Initial Setup

```bash
# 1. Configure CLI
aami config init
aami config set server http://localhost:8080

# 2. Create namespace
NS_ID=$(aami namespace create --name=prod -o json | jq -r '.id')

# 3. Create group
GROUP_ID=$(aami group create --name=web --namespace=$NS_ID -o json | jq -r '.id')

# 4. Create bootstrap token
aami bootstrap-token create --name=web-token --max-uses=50 --expires=30d -o json | tee token.json
TOKEN=$(jq -r '.token' token.json)
```

### Node Registration (on target node)

```bash
export AAMI_SERVER=http://config-server:8080
aami bootstrap-token register \
  --token=$TOKEN \
  --hostname=$(hostname) \
  --ip=$(hostname -I | awk '{print $1}') \
  --group=$GROUP_ID
```

### Bulk Operations

```bash
# Create multiple targets
while IFS=, read -r hostname ip; do
  aami target create --hostname=$hostname --ip=$ip --group=$GROUP_ID
done < targets.csv

# List all active targets
aami target list -o json | jq -r '.[] | select(.status=="active") | .hostname'

# Send heartbeats
for id in $(aami target list -o json | jq -r '.[].id'); do
  aami target heartbeat $id
done
```

## Aliases

```bash
aami ns list          # namespace list
aami bt list          # bootstrap-token list
aami token list       # bootstrap-token list
```

## Environment Variables

```bash
export AAMI_SERVER=http://localhost:8080
```

## Configuration File

Location: `~/.aami/config.yaml`

```yaml
server: http://localhost:8080
default:
  namespace: ""
  output: table
output:
  noheaders: false
  color: true
```

## Troubleshooting

```bash
# Check server connection
curl http://localhost:8080/health

# Verify config
aami config view

# Check version
aami version

# Get help
aami --help
aami namespace --help
```
