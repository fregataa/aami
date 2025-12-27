-- Migration 003: Add soft delete support to all entities
-- Adds deleted_at column to all tables and updates unique constraints

-- Add deleted_at columns to all tables
ALTER TABLE namespaces ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_namespaces_deleted_at ON namespaces(deleted_at);

ALTER TABLE groups ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_groups_deleted_at ON groups(deleted_at);

ALTER TABLE targets ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_targets_deleted_at ON targets(deleted_at);

ALTER TABLE exporters ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_exporters_deleted_at ON exporters(deleted_at);

ALTER TABLE alert_templates ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_alert_templates_deleted_at ON alert_templates(deleted_at);

ALTER TABLE alert_rules ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_alert_rules_deleted_at ON alert_rules(deleted_at);

ALTER TABLE check_settings ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_check_settings_deleted_at ON check_settings(deleted_at);

ALTER TABLE bootstrap_tokens ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_bootstrap_tokens_deleted_at ON bootstrap_tokens(deleted_at);

-- Update unique constraints to allow name reuse after deletion
-- Namespace: name must be unique only among non-deleted records
DROP INDEX IF EXISTS namespaces_name_key;
CREATE UNIQUE INDEX namespaces_name_deleted_at_key ON namespaces(name) WHERE deleted_at IS NULL;

-- Target: hostname must be unique only among non-deleted records
DROP INDEX IF EXISTS targets_hostname_key;
CREATE UNIQUE INDEX targets_hostname_deleted_at_key ON targets(hostname) WHERE deleted_at IS NULL;

-- BootstrapToken: token must be unique only among non-deleted records
DROP INDEX IF EXISTS bootstrap_tokens_token_key;
CREATE UNIQUE INDEX bootstrap_tokens_token_deleted_at_key ON bootstrap_tokens(token) WHERE deleted_at IS NULL;

-- Update search indexes to exclude soft-deleted records by default
-- This ensures queries automatically filter out deleted records

-- Comments explaining soft delete behavior:
-- 1. When deleted_at IS NULL: Record is active (normal operation)
-- 2. When deleted_at IS NOT NULL: Record is soft-deleted (hidden from normal queries)
-- 3. GORM automatically adds "WHERE deleted_at IS NULL" to all queries
-- 4. Use Unscoped() in GORM to query deleted records for restore or purge operations
