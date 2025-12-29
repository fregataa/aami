# Documentation

This directory contains comprehensive documentation for the AAMI project.

## Directory Structure

```
docs/
├── en/                            # English documentation
│   ├── API.md                     # API documentation
│   ├── ALERTING-SYSTEM.md         # Alerting system architecture
│   ├── CHECK-MANAGEMENT.md        # Check management guide
│   ├── CLOUD-INIT.md              # Cloud-init integration guide
│   ├── DEVELOPMENT.md             # Development guide
│   ├── MIGRATION-GUIDE-CHECK-SYSTEM.md  # Check system migration
│   ├── NODE-REGISTRATION.md       # Node registration guide
│   └── QUICKSTART.md              # Quick start guide
├── kr/                            # Korean documentation
├── diagrams/                      # Architecture diagrams
└── README.md                      # This file
```

## Available Documentation

- **QUICKSTART.md** - Quick start guide
- **DEVELOPMENT.md** - Development environment setup and guidelines
- **API.md** - API documentation and examples
- **ALERTING-SYSTEM.md** - Alerting system architecture and design
- **CHECK-MANAGEMENT.md** - Check system management
- **NODE-REGISTRATION.md** - Node registration process
- **CLOUD-INIT.md** - Cloud-init integration guide
- **MIGRATION-GUIDE-CHECK-SYSTEM.md** - Check system migration guide

**Note**: Documentation is available in multiple languages. See `en/` for English and `kr/` for Korean versions.

## Planned Documentation

### Architecture & Design
- **architecture.md** - System architecture and design decisions
- **diagrams/** - Architecture diagrams and visual documentation
  - system-overview.png - High-level system overview
  - data-flow.png - Data flow between components
  - deployment-architecture.png - Deployment topology

### Installation & Configuration
- **installation.md** - Detailed installation guide for various environments
- **configuration.md** - Configuration guide for all components
- **DEPLOYMENT.md** - Deployment strategies and best practices

### API & Integration
- **api-reference.md** - Complete REST API documentation
- **API.md** - API specification and examples

### Operations
- **troubleshooting.md** - Common issues and solutions
- **FAQ.md** - Frequently asked questions

## Quick Links

- [Main README](../README.md) - Project overview
- [Agent Documentation](../.agent/README.md) - Documentation for AI coding agents
- [Project Plan](../.agent/planning/PLAN.md) - Project plan and requirements
- [Sprint Plans](../.agent/planning/sprints/) - Sprint execution plans

## Contributing to Documentation

When contributing to documentation:

1. **Language**: All documentation must be written in English
2. **Format**: Use Markdown format with proper headings and structure
3. **Examples**: Include code examples and command-line snippets where applicable
4. **Diagrams**: Store diagrams in the `diagrams/` subdirectory
5. **Links**: Use relative links to reference other documentation files

## Documentation Standards

- Use clear, concise language
- Include practical examples
- Keep documentation up-to-date with code changes
- Add table of contents for long documents
- Use proper Markdown formatting and syntax highlighting
