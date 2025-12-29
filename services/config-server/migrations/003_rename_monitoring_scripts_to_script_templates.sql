-- Migration: Rename monitoring_scripts to script_templates
-- Purpose: Rename to better reflect the template nature of the entity
-- Date: 2025-12-30

-- ============================================================
-- PART 1: Drop existing foreign key constraint
-- ============================================================

-- Drop the foreign key from script_policies that references monitoring_scripts
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints
               WHERE constraint_name = 'script_policies_template_id_fkey'
               AND table_name = 'script_policies') THEN
        ALTER TABLE script_policies DROP CONSTRAINT script_policies_template_id_fkey;
    END IF;
END $$;

-- ============================================================
-- PART 2: Rename the monitoring_scripts table
-- ============================================================

ALTER TABLE monitoring_scripts RENAME TO script_templates;

-- ============================================================
-- PART 3: Rename indexes
-- ============================================================

-- Drop old indexes
DROP INDEX IF EXISTS idx_monitoring_scripts_name;
DROP INDEX IF EXISTS idx_monitoring_scripts_script_type;
DROP INDEX IF EXISTS idx_monitoring_scripts_deleted_at;

-- Create new indexes with updated names
CREATE INDEX IF NOT EXISTS idx_script_templates_name ON script_templates(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_script_templates_script_type ON script_templates(script_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_script_templates_deleted_at ON script_templates(deleted_at);

-- ============================================================
-- PART 4: Re-create foreign key constraint with new table name
-- ============================================================

ALTER TABLE script_policies
ADD CONSTRAINT script_policies_template_id_fkey
FOREIGN KEY (template_id) REFERENCES script_templates(id) ON DELETE SET NULL;

-- ============================================================
-- NOTES
-- ============================================================
-- This migration renames:
--   - monitoring_scripts -> script_templates
--
-- The foreign key relationship in script_policies is updated to reference
-- the new table name (script_templates).
--
-- All indexes are recreated with new names following the updated naming convention.
--
-- IMPORTANT: This migration does not include a rollback script.
-- If rollback is needed, reverse the renames manually or restore from backup.
