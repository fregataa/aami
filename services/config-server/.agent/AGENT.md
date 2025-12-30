# Config Server Agent Guide

This document provides essential context for AI coding agents working specifically on the Config Server.

## Quick Context

**Service**: Config Server
**Language**: Go 1.21+
**Architecture**: Clean Architecture
**Current Sprint**: Sprint 6 - Unified Error Handling (Planned)
**Location**: `/Users/sh/aami/services/config-server/`

## Purpose & Scope

Config Server is a REST API service that manages configuration for the AAMI monitoring infrastructure. It provides centralized management of:
- Targets (monitored servers/nodes)
- Groups (hierarchical organization)
- Namespaces (logical grouping)
- Exporters (metric collectors)
- Alerts (alert rules)
- Check Settings (monitoring checks)
- Bootstrap Tokens (auto-registration)

## Architecture Pattern

### Clean Architecture Overview

```
┌─────────────────────────────────────┐
│  API Layer (HTTP)                   │  ← Gin handlers, DTOs, middleware
├─────────────────────────────────────┤
│  Service Layer                      │  ← Business logic, orchestration
├─────────────────────────────────────┤
│  Repository Layer                   │  ← GORM models, data access
├─────────────────────────────────────┤
│  Domain Layer                       │  ← Pure Go structs, no dependencies
└─────────────────────────────────────┘
         ↓
┌─────────────────────────────────────┐
│  PostgreSQL Database                │
└─────────────────────────────────────┘
```

### Core Principles

1. **Domain Independence**: Domain models have NO framework dependencies, no GORM tags
2. **Dependency Inversion**: Dependencies point inward (API → Service → Repository → Domain)
3. **ORM Separation**: GORM models in `repository/models/`, convert to/from domain models
4. **Layer Responsibilities**: Domain (entities), Repository (data), Service (logic), API (HTTP)

See: `.agent/ARCHITECTURE.md` for detailed patterns and examples.

## Current Sprint: Sprint 6 - Unified Error Handling

**Status**: Planned (not started)
**Goal**: Consolidate error handling in single `internal/errors` package
**Reference**: `../../.agent/planning/sprints/current/sprint-06-error-handling.md`

## Critical Rules

1. ❌ **Never** add GORM tags to domain models
2. ❌ **Never** import Gin in service layer  
3. ✅ **Always** convert between ORM and domain models at repository boundary
4. ✅ **Always** use context for database operations

## Quick Reference

- **Architecture Details**: `.agent/ARCHITECTURE.md`
- **Sprint Info**: `../../.agent/planning/TRACKER.md`
- **Domain Models**: `internal/domain/`
- **GORM Models**: `internal/repository/models/`

Last Updated: 2024-12-29
