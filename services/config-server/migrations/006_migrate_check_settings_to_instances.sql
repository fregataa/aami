-- Migration 006: Optional data migration from check_settings to check_instances
-- This is an OPTIONAL migration script that converts existing check_settings to the new Template/Instance pattern
--
-- WARNING: This script assumes you have already created check_templates for each check_type
-- Run this ONLY after manually creating the appropriate check_templates
--
-- Steps:
-- 1. Create check_templates manually for each check_type used in check_settings
-- 2. Run this script to convert check_settings to check_instances
-- 3. Verify the migration
-- 4. Optionally drop the old check_settings table

-- IMPORTANT: Uncomment the sections below to execute the migration

-- Step 1: Verify that check_templates exist for all check_types in check_settings
-- Run this query first to see what templates you need to create:
/*
SELECT DISTINCT check_type
FROM check_settings
WHERE NOT EXISTS (
    SELECT 1 FROM check_templates
    WHERE check_templates.check_type = check_settings.check_type
);
*/

-- Step 2: Convert check_settings to check_instances
-- This assumes:
-- - Each check_setting will become a group-level check_instance
-- - The config in check_settings will override the template's default_config
-- - Priority from check_settings will be preserved
/*
INSERT INTO check_instances (
    template_id,
    scope,
    namespace_id,
    group_id,
    config,
    priority,
    is_active,
    created_at,
    updated_at
)
SELECT
    t.id AS template_id,
    'group' AS scope,
    g.namespace_id AS namespace_id,
    cs.group_id AS group_id,
    cs.config AS config,
    cs.priority AS priority,
    true AS is_active,
    cs.created_at,
    cs.updated_at
FROM check_settings cs
JOIN groups g ON cs.group_id = g.id
JOIN check_templates t ON t.check_type = cs.check_type
WHERE NOT EXISTS (
    -- Prevent duplicate inserts if script is run multiple times
    SELECT 1 FROM check_instances ci
    WHERE ci.template_id = t.id
    AND ci.group_id = cs.group_id
);
*/

-- Step 3: Verify the migration
-- Check that all check_settings were converted:
/*
SELECT
    cs.id as old_setting_id,
    cs.check_type,
    cs.group_id,
    ci.id as new_instance_id,
    ci.template_id
FROM check_settings cs
LEFT JOIN groups g ON cs.group_id = g.id
LEFT JOIN check_templates t ON t.check_type = cs.check_type
LEFT JOIN check_instances ci ON ci.template_id = t.id AND ci.group_id = cs.group_id
ORDER BY cs.check_type, cs.group_id;
*/

-- Step 4: (OPTIONAL) Drop the old check_settings table
-- WARNING: Only do this after verifying the migration is successful
-- and you are confident you no longer need the old data
/*
DROP TABLE IF EXISTS check_settings CASCADE;
*/

-- Notes:
-- 1. The check_settings.merge_strategy field is not migrated because the new system
--    uses a different approach (config override at instance level)
-- 2. If you have custom merge strategies, you'll need to handle them differently
-- 3. Consider keeping check_settings for a while as a backup before dropping it
