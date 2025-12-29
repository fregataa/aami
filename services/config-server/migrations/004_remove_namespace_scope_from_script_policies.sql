-- Migration: Remove namespace scope from script_policies
-- Purpose: Simplify scope hierarchy from (global, namespace, group) to (global, group)
-- Date: 2025-12-30

-- ============================================================
-- PART 1: Drop indexes that reference namespace_id
-- ============================================================

-- Drop direct index on namespace_id
DROP INDEX IF EXISTS idx_script_policies_namespace_id;

-- Drop composite indexes that include namespace_id
DROP INDEX IF EXISTS idx_script_policies_scope_namespace;
DROP INDEX IF EXISTS idx_script_policies_scope_namespace_group;

-- Also drop old check_instances indexes if they exist
DROP INDEX IF EXISTS idx_check_instances_namespace;
DROP INDEX IF EXISTS idx_check_instances_namespace_id;

-- ============================================================
-- PART 2: Drop the namespace_id column
-- ============================================================

-- Remove the namespace_id column (will also drop any foreign key constraint)
ALTER TABLE script_policies DROP COLUMN IF EXISTS namespace_id;

-- ============================================================
-- PART 3: Update scope constraint
-- ============================================================

-- Drop the old scope constraint
ALTER TABLE script_policies DROP CONSTRAINT IF EXISTS check_instances_scope_check;
ALTER TABLE script_policies DROP CONSTRAINT IF EXISTS script_policies_scope_check;

-- Add new constraint that only allows 'global' and 'group'
ALTER TABLE script_policies ADD CONSTRAINT script_policies_scope_check
    CHECK (scope IN ('global', 'group'));

-- Update any existing 'namespace' scope records to 'global'
-- (This is a safety measure in case any namespace-scoped records exist)
UPDATE script_policies SET scope = 'global' WHERE scope = 'namespace';

-- ============================================================
-- PART 4: Create new composite index for effective query
-- ============================================================

-- Create composite index for scope + group_id (used in effective checks query)
CREATE INDEX IF NOT EXISTS idx_script_policies_scope_group
ON script_policies(scope, group_id)
WHERE deleted_at IS NULL AND is_active = true;

-- ============================================================
-- NOTES
-- ============================================================
-- This migration removes the 'namespace' scope level from script_policies:
--   - Drops namespace_id column
--   - Updates scope constraint to only allow 'global' and 'group'
--   - Migrates any existing 'namespace' scoped records to 'global'
--
-- After this migration, the scope hierarchy is simplified to:
--   Group (highest priority) > Global (lowest priority)
--
-- IMPORTANT: This migration does not include a rollback script.
-- If rollback is needed, restore the namespace_id column and constraint manually.
