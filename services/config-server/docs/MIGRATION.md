# Database Migration Guide

## Overview

This guide explains the database migration policy and procedures for the AAMI Config Server.

## Migration Policy

### ⚠️ Critical: No Auto-Migration

**The Config Server DOES NOT run database migrations automatically in any environment.**

- ✅ **Server role**: Schema validation only
- ❌ **Prohibited**: Automatic migration execution on server startup
- ✅ **Required**: Manual migration execution before server start

### Server Startup Behavior

When the Config Server starts, it:
1. Validates that all required database tables exist
2. **Fails immediately** if any tables are missing
3. Provides clear instructions for running migrations

```
2025/12/29 10:00:00 Validating database schema...
2025/12/29 10:00:00 ERROR: Missing required database tables: [namespaces groups targets]
2025/12/29 10:00:00 Please run database migrations before starting the server:
2025/12/29 10:00:00   psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/001_initial_schema.sql
```

## Migration Files

### Current Schema

The project uses a **unified migration approach**:
- **Single file**: `migrations/001_initial_schema.sql`
- **Idempotent**: Uses `IF NOT EXISTS` clauses
- **Complete**: Contains all tables, indexes, and default data

### Archived Migrations

Historical migration files are in `migrations/archive/`:
- `001_initial_schema.sql` (original)
- `002_refactor_namespace_to_table.sql`
- `003_add_soft_delete.sql`
- ... (through 009)

**Note**: These are kept for reference only and should NOT be run.

---

## Migration Procedures

### 1. Docker Compose (Recommended for Development)

The easiest method. Migration runs automatically via init container before Config Server starts.

#### First Time Setup

```bash
cd deploy/docker-compose

# Start all services
docker-compose up -d

# Check migration logs
docker-compose logs migration

# Expected output:
# === Running database migrations ===
# CREATE EXTENSION
# CREATE TABLE
# ...
# ✓ Database migrations completed successfully
```

#### How It Works

```yaml
services:
  migration:
    image: postgres:16-alpine
    command: psql -f /migrations/001_initial_schema.sql
    restart: "no"  # Runs once only
    depends_on:
      postgres:
        condition: service_healthy

  config-server:
    depends_on:
      migration:
        condition: service_completed_successfully
```

**Flow**:
1. PostgreSQL starts and becomes healthy
2. Migration container runs `001_initial_schema.sql`
3. Migration container exits with success
4. Config Server starts and validates schema
5. If validation passes, server starts normally

#### Troubleshooting

**Issue**: Migration container fails

```bash
# View migration logs
docker-compose logs migration

# Check database connectivity
docker-compose exec postgres psql -U admin -d config_server -c "\dt"

# Re-run migration manually
docker-compose exec postgres psql -U admin -d config_server -f /migrations/001_initial_schema.sql
```

**Issue**: Schema validation fails after migration

```bash
# Check which tables exist
docker-compose exec postgres psql -U admin -d config_server -c "\dt"

# Verify all 10 tables:
# - namespaces
# - groups
# - targets
# - target_groups
# - exporters
# - alert_templates
# - alert_rules
# - check_templates
# - check_instances
# - bootstrap_tokens
```

---

### 2. Local Development (Direct psql)

For local development without Docker Compose.

#### Prerequisites

- PostgreSQL 16+ installed and running
- Database created: `aami_config` or custom name
- User with appropriate permissions

#### Step 1: Create Database

```bash
# Using default postgres user
psql -U postgres -c "CREATE DATABASE aami_config;"

# Or with custom user
createdb -U admin aami_config
```

#### Step 2: Run Migration

```bash
# From config-server directory
cd services/config-server

# Run migration
psql -h localhost -U postgres -d aami_config -f migrations/001_initial_schema.sql

# Expected output:
# CREATE EXTENSION
# CREATE TABLE
# CREATE TABLE
# ... (all tables and indexes)
# CREATE INDEX
# INSERT 0 6  (default alert templates)
```

#### Step 3: Verify

```bash
# Check tables
psql -h localhost -U postgres -d aami_config -c "\dt"

# Should show 10 tables:
                List of relations
 Schema |        Name        | Type  |  Owner
--------+--------------------+-------+----------
 public | alert_rules        | table | postgres
 public | alert_templates    | table | postgres
 public | bootstrap_tokens   | table | postgres
 public | check_instances    | table | postgres
 public | check_templates    | table | postgres
 public | exporters          | table | postgres
 public | groups             | table | postgres
 public | namespaces         | table | postgres
 public | target_groups      | table | postgres
 public | targets            | table | postgres
```

#### Step 4: Start Server

```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=aami_config
export PORT=8080

# Build and run
go build -o config-server ./cmd/config-server
./config-server

# Expected output:
# 2025/12/29 10:00:00 Validating database schema...
# 2025/12/29 10:00:00 ✓ Database schema validation successful - all required tables exist
# 2025/12/29 10:00:00 Starting config-server on :8080
```

---

### 3. Production Deployment

**Critical**: Always follow these steps for production migrations.

#### Pre-Deployment Checklist

- [ ] Review migration SQL thoroughly
- [ ] Test migration on staging environment first
- [ ] Back up production database
- [ ] Schedule maintenance window
- [ ] Prepare rollback plan
- [ ] Notify stakeholders

#### Step 1: Backup Database

```bash
# Full database backup
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME -F c -f backup_$(date +%Y%m%d_%H%M%S).dump

# Schema-only backup
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME -s -f schema_backup_$(date +%Y%m%d_%H%M%S).sql

# Verify backup
pg_restore --list backup_20251229_100000.dump | head -20
```

#### Step 2: Run Migration

```bash
# Connect to production database
psql -h $DB_HOST -U $DB_USER -d $DB_NAME

# Start transaction (for safety)
BEGIN;

# Run migration
\i /path/to/migrations/001_initial_schema.sql

# Verify results
\dt

# If everything looks good, commit
COMMIT;

# If issues, rollback
ROLLBACK;
```

#### Step 3: Verify Migration

```bash
# Check table count
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT schemaname, count(*)
FROM pg_tables
WHERE schemaname = 'public'
GROUP BY schemaname;
"

# Check for errors in logs
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\dt" | grep -c "table"
# Should output: 10
```

#### Step 4: Deploy Application

```bash
# Deploy new config-server version
# The server will validate schema on startup

# Monitor logs
kubectl logs -f deployment/config-server -n aami

# Expected output:
# Validating database schema...
# ✓ Database schema validation successful - all required tables exist
```

#### Rollback Procedure

**If migration fails**:

```bash
# Restore from backup
pg_restore -h $DB_HOST -U $DB_USER -d $DB_NAME -c backup_20251229_100000.dump

# Or restore from SQL
psql -h $DB_HOST -U $DB_USER -d $DB_NAME < schema_backup_20251229_100000.sql
```

**If server fails schema validation**:

```bash
# Check which tables are missing
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\dt"

# Re-run migration if needed
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/001_initial_schema.sql
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy Config Server

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run database migration
        env:
          PGHOST: ${{ secrets.DB_HOST }}
          PGUSER: ${{ secrets.DB_USER }}
          PGPASSWORD: ${{ secrets.DB_PASSWORD }}
          PGDATABASE: ${{ secrets.DB_NAME }}
        run: |
          psql -f services/config-server/migrations/001_initial_schema.sql

      - name: Verify migration
        run: |
          TABLE_COUNT=$(psql -t -c "SELECT count(*) FROM pg_tables WHERE schemaname='public'")
          if [ "$TABLE_COUNT" != "10" ]; then
            echo "Migration failed: expected 10 tables, got $TABLE_COUNT"
            exit 1
          fi

  deploy:
    needs: migrate
    runs-on: ubuntu-latest
    steps:
      - name: Deploy application
        run: |
          kubectl apply -f k8s/
```

---

## Future: Migration Tool (Goose)

### Planned Enhancement

The project will adopt [Goose](https://github.com/pressly/goose) for professional migration management.

#### Why Goose?

- ✅ Migration versioning and tracking
- ✅ Up/down migration support
- ✅ SQL and Go-based migrations
- ✅ Migration status tracking
- ✅ Rollback capabilities
- ✅ CLI tool for easy execution

#### Future Structure

```
migrations/
├── 00001_initial_schema.sql
├── 00002_add_new_feature.sql
└── 00003_refactor_indexes.sql
```

#### Future Commands

```bash
# Check migration status
goose -dir migrations status

# Run migrations
goose -dir migrations up

# Rollback last migration
goose -dir migrations down

# Migrate to specific version
goose -dir migrations up-to 2
```

**Timeline**: To be implemented in Sprint 7 (Q1 2025)

---

## FAQ

### Q: Why doesn't the server run migrations automatically?

**A**: Automatic migrations in production are dangerous because:
- No backup before changes
- No rollback mechanism
- No human review
- Multiple server instances may conflict
- Schema changes need careful planning

### Q: What happens if I start the server without running migrations?

**A**: The server will fail to start with a clear error message:
```
ERROR: Missing required database tables: [namespaces groups targets ...]
Please run database migrations before starting the server
```

### Q: Can I run migrations multiple times?

**A**: Yes! The migration file is idempotent using `IF NOT EXISTS` clauses. Running it multiple times is safe.

### Q: How do I reset the database for testing?

```bash
# Docker Compose
docker-compose down -v  # Removes volumes
docker-compose up -d    # Recreates everything

# Local PostgreSQL
psql -U postgres -c "DROP DATABASE aami_config;"
psql -U postgres -c "CREATE DATABASE aami_config;"
psql -U postgres -d aami_config -f migrations/001_initial_schema.sql
```

### Q: Where are the old migration files?

**A**: They're archived in `migrations/archive/` for reference. Don't run them - use only `001_initial_schema.sql`.

### Q: What if schema validation fails in production?

**A**: Follow this procedure:
1. Check server logs for missing tables
2. Verify migration was run successfully
3. If needed, re-run migration manually
4. Restart server after fixing schema

---

## Troubleshooting

### Common Issues

#### 1. "relation already exists" Error

**Cause**: Migration was run but server still tried to create tables (old behavior)

**Fix**: Update to latest code that uses `validateSchema()` instead of `runMigrations()`

#### 2. "missing required tables" Error

**Cause**: Migration wasn't run before server started

**Fix**:
```bash
# Run migration
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/001_initial_schema.sql

# Restart server
```

#### 3. Migration Container Keeps Restarting

**Cause**: Migration script has errors or `restart` policy is wrong

**Fix**:
```bash
# Check migration container logs
docker-compose logs migration

# Ensure restart: "no" in docker-compose.yaml
```

#### 4. Permission Denied

**Cause**: Database user doesn't have CREATE privileges

**Fix**:
```sql
-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE aami_config TO your_user;
GRANT ALL ON SCHEMA public TO your_user;
```

---

## Migration Checklist

### New Developer Setup

- [ ] Install PostgreSQL 16+
- [ ] Create database: `aami_config`
- [ ] Run migration: `psql ... -f migrations/001_initial_schema.sql`
- [ ] Verify tables: `psql ... -c "\dt"`
- [ ] Start server and confirm validation passes

### Production Deployment

- [ ] Review migration SQL
- [ ] Test on staging
- [ ] Backup production database
- [ ] Schedule maintenance window
- [ ] Run migration with transaction
- [ ] Verify table count (should be 10)
- [ ] Deploy application
- [ ] Monitor server startup logs
- [ ] Verify schema validation passes

### Troubleshooting

- [ ] Check migration logs
- [ ] Verify database connectivity
- [ ] Check table existence
- [ ] Review server startup logs
- [ ] Check for permission issues

---

## References

- [PostgreSQL Documentation](https://www.postgresql.org/docs/16/)
- [Goose Migration Tool](https://github.com/pressly/goose)
- [Database Migration Best Practices](https://www.postgresql.org/docs/current/ddl-basics.html)
- [AAMI Development Guide](../DEVELOPMENT.md)
- [AAMI Architecture Guide](../.agent/docs/AGENT.md)

---

## Change Log

### 2024-12-29
- Initial migration policy documentation
- Unified migration approach (001_initial_schema.sql)
- Schema validation implementation
- Docker Compose init container pattern
- Production deployment procedures
