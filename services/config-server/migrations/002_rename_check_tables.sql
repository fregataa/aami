-- Migration: Rename check system tables and columns
-- Purpose: Rename CheckTemplate -> MonitoringScript and CheckInstance -> ScriptPolicy
-- Date: 2025-12-29

-- ============================================================
-- PART 1: Rename check_templates to monitoring_scripts
-- ============================================================

-- Rename the table
ALTER TABLE check_templates RENAME TO monitoring_scripts;

-- Rename the check_type column to script_type
ALTER TABLE monitoring_scripts RENAME COLUMN check_type TO script_type;

-- Add version and hash columns if they don't exist (for script integrity tracking)
ALTER TABLE monitoring_scripts ADD COLUMN IF NOT EXISTS version VARCHAR(50) NOT NULL DEFAULT '1.0.0';
ALTER TABLE monitoring_scripts ADD COLUMN IF NOT EXISTS hash VARCHAR(64) NOT NULL DEFAULT '';

-- Remove timeout_seconds column if it exists (moved to config)
ALTER TABLE monitoring_scripts DROP COLUMN IF EXISTS timeout_seconds;

-- Drop old indexes (if any exist)
DROP INDEX IF EXISTS idx_check_templates_name;
DROP INDEX IF EXISTS idx_check_templates_check_type;
DROP INDEX IF EXISTS idx_check_templates_deleted_at;

-- Create new indexes with updated names
CREATE INDEX IF NOT EXISTS idx_monitoring_scripts_name ON monitoring_scripts(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_monitoring_scripts_script_type ON monitoring_scripts(script_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_monitoring_scripts_deleted_at ON monitoring_scripts(deleted_at);

-- ============================================================
-- PART 2: Rename check_instances to script_policies
-- ============================================================

-- Rename the table
ALTER TABLE check_instances RENAME TO script_policies;

-- Rename the check_type column to script_type
ALTER TABLE script_policies RENAME COLUMN check_type TO script_type;

-- Add version and hash columns if they don't exist (copied from template)
ALTER TABLE script_policies ADD COLUMN IF NOT EXISTS version VARCHAR(50) NOT NULL DEFAULT '1.0.0';
ALTER TABLE script_policies ADD COLUMN IF NOT EXISTS hash VARCHAR(64) NOT NULL DEFAULT '';

-- Add default_config column if it doesn't exist
ALTER TABLE script_policies ADD COLUMN IF NOT EXISTS default_config JSONB NOT NULL DEFAULT '{}';

-- Add description column if it doesn't exist
ALTER TABLE script_policies ADD COLUMN IF NOT EXISTS description TEXT;

-- Add template tracking columns if they don't exist
ALTER TABLE script_policies ADD COLUMN IF NOT EXISTS created_from_template_id VARCHAR(255);
ALTER TABLE script_policies ADD COLUMN IF NOT EXISTS created_from_template_name VARCHAR(255);
ALTER TABLE script_policies ADD COLUMN IF NOT EXISTS template_version VARCHAR(50);

-- Rename enabled to is_active if enabled exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'script_policies' AND column_name = 'enabled') THEN
        ALTER TABLE script_policies RENAME COLUMN enabled TO is_active;
    END IF;
END $$;

-- Add is_active column if it doesn't exist
ALTER TABLE script_policies ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;

-- Remove obsolete columns if they exist
ALTER TABLE script_policies DROP COLUMN IF EXISTS merge_strategy;
ALTER TABLE script_policies DROP COLUMN IF EXISTS timeout_seconds;

-- Rename the template_id foreign key constraint
DO $$
BEGIN
    -- Drop old foreign key if it exists
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints
               WHERE constraint_name = 'check_instances_template_id_fkey'
               AND table_name = 'script_policies') THEN
        ALTER TABLE script_policies DROP CONSTRAINT check_instances_template_id_fkey;
    END IF;

    -- Add new foreign key with updated reference
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints
                   WHERE constraint_name = 'script_policies_template_id_fkey'
                   AND table_name = 'script_policies') THEN
        ALTER TABLE script_policies
        ADD CONSTRAINT script_policies_template_id_fkey
        FOREIGN KEY (template_id) REFERENCES monitoring_scripts(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Drop old indexes (if any exist)
DROP INDEX IF EXISTS idx_check_instances_template_id;
DROP INDEX IF EXISTS idx_check_instances_name;
DROP INDEX IF EXISTS idx_check_instances_check_type;
DROP INDEX IF EXISTS idx_check_instances_scope;
DROP INDEX IF EXISTS idx_check_instances_namespace_id;
DROP INDEX IF EXISTS idx_check_instances_group_id;
DROP INDEX IF EXISTS idx_check_instances_priority;
DROP INDEX IF EXISTS idx_check_instances_enabled;
DROP INDEX IF EXISTS idx_check_instances_deleted_at;

-- Create new indexes with updated names
CREATE INDEX IF NOT EXISTS idx_script_policies_template_id ON script_policies(template_id);
CREATE INDEX IF NOT EXISTS idx_script_policies_name ON script_policies(name);
CREATE INDEX IF NOT EXISTS idx_script_policies_script_type ON script_policies(script_type);
CREATE INDEX IF NOT EXISTS idx_script_policies_scope ON script_policies(scope);
CREATE INDEX IF NOT EXISTS idx_script_policies_namespace_id ON script_policies(namespace_id) WHERE namespace_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_script_policies_group_id ON script_policies(group_id) WHERE group_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_script_policies_priority ON script_policies(priority);
CREATE INDEX IF NOT EXISTS idx_script_policies_is_active ON script_policies(is_active);
CREATE INDEX IF NOT EXISTS idx_script_policies_deleted_at ON script_policies(deleted_at);

-- ============================================================
-- PART 3: Update composite indexes for effective query
-- ============================================================

-- Drop old composite indexes (if any exist)
DROP INDEX IF EXISTS idx_check_instances_scope_namespace;
DROP INDEX IF EXISTS idx_check_instances_scope_namespace_group;

-- Create new composite indexes
CREATE INDEX IF NOT EXISTS idx_script_policies_scope_namespace
ON script_policies(scope, namespace_id)
WHERE deleted_at IS NULL AND is_active = true;

CREATE INDEX IF NOT EXISTS idx_script_policies_scope_namespace_group
ON script_policies(scope, namespace_id, group_id)
WHERE deleted_at IS NULL AND is_active = true;

-- ============================================================
-- NOTES
-- ============================================================
-- This migration renames:
--   - check_templates -> monitoring_scripts
--   - check_instances -> script_policies
--   - check_type -> script_type (in both tables)
--
-- The foreign key relationship is updated to reference the new table name.
-- All indexes are recreated with new names following the updated naming convention.
--
-- IMPORTANT: This migration does not include a rollback script.
-- If rollback is needed, reverse the renames manually or restore from backup.
