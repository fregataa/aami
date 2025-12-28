# AAMI CLI User Guide

The AAMI CLI (`aami`) is a command-line tool for managing the AAMI monitoring infrastructure. It provides an intuitive interface to create, manage, and monitor namespaces, groups, targets, and bootstrap tokens.

## Table of Contents

- [Installation](#installation)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Command Reference](#command-reference)
  - [Namespace Commands](#namespace-commands)
  - [Group Commands](#group-commands)
  - [Target Commands](#target-commands)
  - [Bootstrap Token Commands](#bootstrap-token-commands)
  - [Config Commands](#config-commands)
- [Common Workflows](#common-workflows)
- [Output Formats](#output-formats)
- [Troubleshooting](#troubleshooting)

## Installation

### From Source

```bash
cd services/config-server
go build -o aami cmd/aami/main.go

# Install to system path (optional)
sudo cp aami /usr/local/bin/
```

### Verify Installation

```bash
aami version
# Output: aami version 0.1.0
```

## Getting Started

### 1. Initialize Configuration

Create the CLI configuration file:

```bash
aami config init
# ✓ Initialized configuration at: /Users/username/.aami/config.yaml
```

### 2. Configure Server URL

Set the config server URL:

```bash
aami config set server http://localhost:8080
# ✓ Set server = http://localhost:8080
```

Or use environment variable:

```bash
export AAMI_SERVER=http://localhost:8080
```

### 3. Verify Connection

```bash
aami namespace list
```

## Configuration

### Configuration File

Location: `~/.aami/config.yaml`

```yaml
server: http://localhost:8080

default:
  namespace: ""        # Default namespace (empty = none)
  output: table       # Default output format (table|json|yaml)

output:
  noheaders: false    # Hide table headers
  color: true         # Enable colored output
```

### Configuration Commands

```bash
# View current configuration
aami config view

# Set a value
aami config set server http://config-server:8080
aami config set output json
aami config set default-namespace production

# Get a value
aami config get server

# Show config file path
aami config path
```

### Configuration Precedence

Configuration values are resolved in the following order (highest to lowest):

1. Command-line flags (`--server`, `--output`)
2. Environment variables (`AAMI_SERVER`)
3. Configuration file (`~/.aami/config.yaml`)
4. Built-in defaults

## Command Reference

### Global Flags

Available for all commands:

```bash
-s, --server string   Server URL (overrides config)
-o, --output string   Output format: table|json|yaml (overrides config)
-h, --help           Show help
```

### Namespace Commands

Manage logical namespaces for organizing monitoring infrastructure.

#### Create a Namespace

```bash
aami namespace create \
  --name=production \
  --description="Production environment" \
  --priority=100 \
  --merge-strategy=merge
```

**Flags:**
- `--name, -n` (required): Namespace name
- `--description, -d`: Description
- `--priority, -p`: Policy priority (default: 100)
- `--merge-strategy`: Merge strategy (merge|override, default: merge)

#### List Namespaces

```bash
# List all namespaces
aami namespace list

# List in JSON format
aami namespace list -o json
```

#### Get a Namespace

```bash
# Get by ID
aami namespace get 550e8400-e29b-41d4-a716-446655440000

# Get by name
aami namespace get --name=production
```

#### Update a Namespace

```bash
aami namespace update <id> \
  --description="Updated description" \
  --priority=200
```

**Flags:**
- `--name, -n`: New name
- `--description, -d`: New description
- `--priority, -p`: New priority
- `--merge-strategy`: New merge strategy

#### Delete a Namespace

```bash
aami namespace delete <id>
```

**Aliases:** `ns` (e.g., `aami ns list`)

### Group Commands

Manage hierarchical groups within namespaces.

#### Create a Group

```bash
# Create a top-level group
aami group create \
  --name=web-tier \
  --namespace=<namespace-id> \
  --description="Web servers" \
  --priority=100

# Create a child group
aami group create \
  --name=web-production \
  --namespace=<namespace-id> \
  --parent=<parent-group-id>
```

**Flags:**
- `--name, -n` (required): Group name
- `--namespace` (required): Namespace ID
- `--description, -d`: Description
- `--priority, -p`: Priority (default: 100)
- `--parent`: Parent group ID (for hierarchical groups)

#### List Groups

```bash
# List all groups
aami group list

# List groups in a namespace
aami group list --namespace=<namespace-id>
```

#### Get a Group

```bash
aami group get <id>
```

#### Update a Group

```bash
aami group update <id> \
  --description="Updated description" \
  --parent=<new-parent-id>
```

#### Delete a Group

```bash
aami group delete <id>
```

#### View Group Hierarchy

```bash
# List child groups
aami group children <id>

# List ancestor groups
aami group ancestors <id>
```

### Target Commands

Manage monitored targets (servers/nodes).

#### Create a Target

```bash
# Create with single group
aami target create \
  --hostname=web-01 \
  --ip=10.0.1.100 \
  --group=<group-id>

# Create with multiple groups
aami target create \
  --hostname=web-01 \
  --ip=10.0.1.100 \
  --groups=<group-id-1>,<group-id-2>

# Create without group (auto-creates own group)
aami target create \
  --hostname=web-01 \
  --ip=10.0.1.100
```

**Flags:**
- `--hostname` (required): Target hostname
- `--ip` (required): IP address
- `--group`: Single group ID
- `--groups`: Comma-separated group IDs

#### List Targets

```bash
# List all targets
aami target list

# List targets in a group
aami target list --group=<group-id>
```

#### Get a Target

```bash
# Get by ID
aami target get <id>

# Get by hostname
aami target get --hostname=web-01
```

#### Update a Target

```bash
aami target update <id> \
  --hostname=web-01-new \
  --groups=<group-id-1>,<group-id-2>
```

**Flags:**
- `--hostname`: New hostname
- `--ip`: New IP address
- `--groups`: New comma-separated group IDs

#### Delete a Target

```bash
aami target delete <id>
```

#### Target Operations

```bash
# Update target status
aami target status <id> --status=active
aami target status <id> --status=inactive

# Send heartbeat
aami target heartbeat <id>
```

### Bootstrap Token Commands

Manage bootstrap tokens for automated node registration.

#### Create a Bootstrap Token

```bash
# Create with 7 days expiry
aami bootstrap-token create \
  --name=production-token \
  --max-uses=10 \
  --expires=7d

# Create with specific expiry date
aami bootstrap-token create \
  --name=staging-token \
  --max-uses=50 \
  --expires=2025-12-31

# Create with default expiry (30 days)
aami bootstrap-token create \
  --name=dev-token \
  --max-uses=100
```

**Flags:**
- `--name, -n` (required): Token name
- `--max-uses`: Maximum number of uses (default: 10)
- `--expires`: Expiry (format: `7d`, `30d`, or `YYYY-MM-DD`, default: 30d)

**⚠️ Important:** Save the token value immediately! It cannot be retrieved later.

#### List Bootstrap Tokens

```bash
aami bootstrap-token list
```

#### Get a Bootstrap Token

```bash
aami bootstrap-token get <id>
```

#### Update a Bootstrap Token

```bash
# Increase max uses
aami bootstrap-token update <id> --max-uses=20

# Extend expiry by 30 days
aami bootstrap-token update <id> --expires=30d
```

#### Delete a Bootstrap Token

```bash
aami bootstrap-token delete <id>
```

#### Validate a Bootstrap Token

```bash
aami bootstrap-token validate <token-string>
```

#### Register a Node with Bootstrap Token

```bash
# Register with auto-detected hostname
aami bootstrap-token register \
  --token=<token-string> \
  --ip=10.0.1.100

# Register with specific hostname and group
aami bootstrap-token register \
  --token=<token-string> \
  --hostname=web-01 \
  --ip=10.0.1.100 \
  --group=<group-id>

# Use in automation (e.g., cloud-init)
aami bootstrap-token register \
  --token=$(cat /etc/aami/bootstrap-token) \
  --hostname=$(hostname) \
  --ip=$(hostname -I | awk '{print $1}')
```

**Flags:**
- `--token` (required): Bootstrap token string
- `--ip` (required): IP address
- `--hostname`: Hostname (default: OS hostname)
- `--group`: Group ID (optional, creates own group if omitted)

**Aliases:** `bt`, `token` (e.g., `aami bt list`)

### Config Commands

Manage CLI configuration.

```bash
# View configuration
aami config view

# Set values
aami config set server http://localhost:8080
aami config set output json
aami config set default-namespace production

# Get values
aami config get server
aami config get output

# Initialize config file
aami config init

# Show config file path
aami config path
```

## Common Workflows

### Workflow 1: Initial Setup

```bash
# 1. Initialize CLI
aami config init
aami config set server http://config-server:8080

# 2. Create namespace
NAMESPACE_ID=$(aami namespace create \
  --name=production \
  --priority=100 \
  -o json | jq -r '.id')

# 3. Create group
GROUP_ID=$(aami group create \
  --name=web-tier \
  --namespace=$NAMESPACE_ID \
  -o json | jq -r '.id')

# 4. Create bootstrap token
aami bootstrap-token create \
  --name=web-token \
  --max-uses=50 \
  --expires=30d \
  -o json | tee web-token.json

# Save token for later use
TOKEN=$(jq -r '.token' web-token.json)
```

### Workflow 2: Node Registration (On Target Node)

```bash
# Using bootstrap token from server
aami bootstrap-token register \
  --token=$TOKEN \
  --hostname=$(hostname) \
  --ip=$(hostname -I | awk '{print $1}') \
  --group=$GROUP_ID
```

### Workflow 3: Monitoring Operations

```bash
# List all active targets
aami target list

# Check specific target
aami target get --hostname=web-01

# Send heartbeat
aami target heartbeat <target-id>

# List targets in a group
aami target list --group=<group-id>
```

### Workflow 4: Hierarchical Group Management

```bash
# Create parent group
PARENT_ID=$(aami group create \
  --name=web-tier \
  --namespace=$NS_ID \
  -o json | jq -r '.id')

# Create child groups
aami group create --name=web-prod --namespace=$NS_ID --parent=$PARENT_ID
aami group create --name=web-staging --namespace=$NS_ID --parent=$PARENT_ID

# View hierarchy
aami group children $PARENT_ID
```

### Workflow 5: Bulk Operations with Scripts

```bash
# Create multiple targets from CSV
while IFS=, read -r hostname ip; do
  aami target create \
    --hostname=$hostname \
    --ip=$ip \
    --group=$GROUP_ID
done < targets.csv

# List all targets and filter
aami target list -o json | \
  jq -r '.[] | select(.status=="active") | .hostname'

# Update multiple targets
for target_id in $(aami target list -o json | jq -r '.[].id'); do
  aami target heartbeat $target_id
done
```

## Output Formats

### Table Format (Default)

Human-readable table output:

```bash
aami namespace list
```

```
ID                                   NAME        POLICY PRIORITY  MERGE STRATEGY  CREATED AT
550e8400-e29b-41d4-a716-446655440000 production  100               merge           2025-12-29 12:00:00
```

### JSON Format

Machine-readable JSON output:

```bash
aami namespace list -o json
```

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "production",
    "description": "Production environment",
    "policy_priority": 100,
    "merge_strategy": "merge",
    "created_at": "2025-12-29T12:00:00Z",
    "updated_at": "2025-12-29T12:00:00Z"
  }
]
```

**Use with jq:**

```bash
# Get all namespace names
aami namespace list -o json | jq -r '.[].name'

# Filter by priority
aami namespace list -o json | jq '.[] | select(.policy_priority > 100)'
```

### YAML Format

Configuration-friendly YAML output:

```bash
aami namespace get <id> -o yaml
```

```yaml
id: 550e8400-e29b-41d4-a716-446655440000
name: production
description: Production environment
policy_priority: 100
merge_strategy: merge
created_at: 2025-12-29T12:00:00Z
updated_at: 2025-12-29T12:00:00Z
```

## Troubleshooting

### Connection Issues

**Problem:** `failed to make GET request: connection refused`

**Solution:**
1. Verify server is running:
   ```bash
   curl http://localhost:8080/health
   ```

2. Check server URL in config:
   ```bash
   aami config get server
   ```

3. Set correct server URL:
   ```bash
   aami config set server http://localhost:8080
   ```

### Authentication Issues

**Problem:** `request failed with status 401: Unauthorized`

**Solution:** Currently, the CLI does not support authentication. This will be added in a future release. Ensure your server is configured to allow unauthenticated access or use a reverse proxy for authentication.

### Config File Issues

**Problem:** `failed to load config: no such file or directory`

**Solution:**
```bash
aami config init
```

### Invalid Input

**Problem:** `failed to create namespace: validation error`

**Solution:** Check the error message for specific validation failures:
- Ensure required flags are provided
- Verify data formats (e.g., IP addresses, UUIDs)
- Check field lengths and constraints

### Command Not Found

**Problem:** `aami: command not found`

**Solution:**
1. Verify installation:
   ```bash
   which aami
   ```

2. Install to system path:
   ```bash
   sudo cp bin/aami /usr/local/bin/
   ```

3. Or add to PATH:
   ```bash
   export PATH=$PATH:/path/to/aami/bin
   ```

## Advanced Usage

### Using with Shell Scripts

```bash
#!/bin/bash
set -e

# Set server
export AAMI_SERVER=http://config-server:8080

# Create resources
NS_ID=$(aami namespace create --name=prod -o json | jq -r '.id')
GROUP_ID=$(aami group create --name=web --namespace=$NS_ID -o json | jq -r '.id')

# Register multiple nodes
for i in {1..5}; do
  aami target create \
    --hostname=web-$i \
    --ip=10.0.1.$((100+i)) \
    --group=$GROUP_ID
done

echo "Created namespace: $NS_ID"
echo "Created group: $GROUP_ID"
echo "Registered 5 targets"
```

### Using with Cloud-Init

```yaml
#cloud-config
packages:
  - curl
  - jq

write_files:
  - path: /usr/local/bin/aami
    permissions: '0755'
    content: |
      # Download aami binary
      # (binary content)

  - path: /etc/aami/bootstrap-token
    permissions: '0600'
    content: |
      YOUR_BOOTSTRAP_TOKEN_HERE

runcmd:
  - |
    aami config set server http://config-server:8080
    aami bootstrap-token register \
      --token=$(cat /etc/aami/bootstrap-token) \
      --hostname=$(hostname) \
      --ip=$(hostname -I | awk '{print $1}')
```

### Using with Ansible

```yaml
- name: Register node with AAMI
  shell: |
    aami bootstrap-token register \
      --token={{ aami_bootstrap_token }} \
      --hostname={{ inventory_hostname }} \
      --ip={{ ansible_default_ipv4.address }} \
      --group={{ aami_group_id }}
  environment:
    AAMI_SERVER: "{{ aami_server_url }}"
```

## Getting Help

```bash
# General help
aami --help

# Command-specific help
aami namespace --help
aami namespace create --help

# Version information
aami version
```

## Next Steps

- Learn about the [Config Server API](../README.md)
- Explore [Node Registration](./NODE-REGISTRATION.md)
- Read about [Architecture](./ARCHITECTURE.md)
