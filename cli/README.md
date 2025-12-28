# AAMI CLI

Command-line interface for managing AAMI monitoring infrastructure.

## Installation

### From Source

```bash
# Build
go build -o aami cmd/aami/main.go

# Install to system path (optional)
sudo cp aami /usr/local/bin/

# Or move to bin directory
mkdir -p bin
mv aami bin/
```

### Verify Installation

```bash
aami version
# Output: aami version 0.1.0
```

## Quick Start

### 1. Initialize Configuration

```bash
aami config init
aami config set server http://localhost:8080
```

### 2. Create Resources

```bash
# Create namespace
aami namespace create --name=production --priority=100

# Create group
aami group create --name=web-tier --namespace=<namespace-id>

# Register target
aami target create --hostname=web-01 --ip=10.0.1.100 --group=<group-id>
```

### 3. Bootstrap Token

```bash
# Create bootstrap token
aami bootstrap-token create --name=prod-token --max-uses=50 --expires=30d

# Register node with token (on target node)
aami bootstrap-token register \
  --token=<token> \
  --hostname=$(hostname) \
  --ip=$(hostname -I | awk '{print $1}')
```

## Documentation

- **[CLI User Guide](./CLI.md)** - Complete command reference and workflows
- **[CLI Quick Reference](./CLI-QUICK-REFERENCE.md)** - Cheat sheet for common commands

## Configuration

Configuration file location: `~/.aami/config.yaml`

```yaml
server: http://localhost:8080
default:
  namespace: ""
  output: table
output:
  noheaders: false
  color: true
```

## Available Commands

### Resource Management
- `aami namespace` - Manage namespaces
- `aami group` - Manage groups
- `aami target` - Manage targets
- `aami bootstrap-token` - Manage bootstrap tokens

### Utility Commands
- `aami config` - Manage CLI configuration
- `aami version` - Show version information

## Output Formats

```bash
# Table format (default)
aami namespace list

# JSON format
aami namespace list -o json

# YAML format
aami namespace list -o yaml
```

## Global Flags

```bash
-s, --server string   Server URL (overrides config)
-o, --output string   Output format: table|json|yaml
-h, --help           Show help
```

## Environment Variables

```bash
export AAMI_SERVER=http://localhost:8080
```

## Development

### Build

```bash
go build -o aami cmd/aami/main.go
```

### Run Tests

```bash
go test ./...
```

### Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing

## Project Structure

```
cli/
├── cmd/aami/          # Main entry point
├── internal/
│   ├── client/        # API client
│   ├── cmd/           # Command implementations
│   ├── config/        # Configuration management
│   └── output/        # Output formatters
├── bin/               # Built binaries
├── CLI.md             # Full user guide
├── CLI-QUICK-REFERENCE.md  # Quick reference
└── README.md          # This file
```

## License

MIT License - see [LICENSE](../LICENSE) for details
