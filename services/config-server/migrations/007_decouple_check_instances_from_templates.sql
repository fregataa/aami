-- Migration 007: Decouple check_instances from templates
-- Instances now contain a snapshot of template content at creation time
-- This allows templates to be modified or deleted without affecting existing instances

-- Add template fields (copied from template at creation time)
ALTER TABLE check_instances
    ADD COLUMN name VARCHAR(255),
    ADD COLUMN check_type VARCHAR(100),
    ADD COLUMN script_content TEXT,
    ADD COLUMN language VARCHAR(50),
    ADD COLUMN default_config JSONB DEFAULT '{}',
    ADD COLUMN description TEXT,
    ADD COLUMN version VARCHAR(50),
    ADD COLUMN hash VARCHAR(64);

-- Add metadata fields (for tracking origin template)
ALTER TABLE check_instances
    ADD COLUMN created_from_template_id VARCHAR(255),
    ADD COLUMN created_from_template_name VARCHAR(255),
    ADD COLUMN template_version VARCHAR(50);

-- Drop unique indexes that reference template_id
DROP INDEX IF EXISTS check_instances_unique_global;
DROP INDEX IF EXISTS check_instances_unique_namespace;
DROP INDEX IF EXISTS check_instances_unique_group;

-- Drop template_id index
DROP INDEX IF EXISTS idx_check_instances_template_id;

-- Drop foreign key constraint (PostgreSQL doesn't have named FK for this, need to find it)
DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'check_instances'::regclass
    AND contype = 'f'
    AND confrelid = 'check_templates'::regclass;

    IF constraint_name IS NOT NULL THEN
        EXECUTE 'ALTER TABLE check_instances DROP CONSTRAINT ' || constraint_name;
    END IF;
END $$;

-- Drop template_id column
ALTER TABLE check_instances DROP COLUMN IF EXISTS template_id;

-- Add NOT NULL constraints to required template fields
-- (After data population in production, but since we have no data, we can do it now)
ALTER TABLE check_instances
    ALTER COLUMN name SET NOT NULL,
    ALTER COLUMN check_type SET NOT NULL,
    ALTER COLUMN script_content SET NOT NULL,
    ALTER COLUMN language SET NOT NULL,
    ALTER COLUMN version SET NOT NULL,
    ALTER COLUMN hash SET NOT NULL;

-- Create new indexes for template fields
CREATE INDEX idx_check_instances_name ON check_instances(name);
CREATE INDEX idx_check_instances_check_type ON check_instances(check_type);
CREATE INDEX idx_check_instances_hash ON check_instances(hash);
CREATE INDEX idx_check_instances_created_from_template_id ON check_instances(created_from_template_id);

-- New unique constraints based on name+checktype instead of template_id
-- This allows multiple instances of the same check type at different scopes
-- (Previous constraint was too restrictive - one instance per template per scope)
-- Now we allow multiple instances, differentiated by name+checktype+scope

-- Update table comment
COMMENT ON TABLE check_instances IS 'Check instances containing snapshots of check templates. Independent from templates after creation.';
COMMENT ON COLUMN check_instances.name IS 'Check name (copied from template at creation)';
COMMENT ON COLUMN check_instances.check_type IS 'Check type identifier (copied from template)';
COMMENT ON COLUMN check_instances.script_content IS 'Check script content (copied from template)';
COMMENT ON COLUMN check_instances.language IS 'Script language (bash, python, etc.)';
COMMENT ON COLUMN check_instances.default_config IS 'Default configuration (copied from template)';
COMMENT ON COLUMN check_instances.description IS 'Check description';
COMMENT ON COLUMN check_instances.version IS 'Check version';
COMMENT ON COLUMN check_instances.hash IS 'SHA256 hash of script_content';
COMMENT ON COLUMN check_instances.created_from_template_id IS 'Optional: ID of template this was created from';
COMMENT ON COLUMN check_instances.created_from_template_name IS 'Optional: Name of template at creation time';
COMMENT ON COLUMN check_instances.template_version IS 'Optional: Version of template at creation time';
COMMENT ON COLUMN check_instances.config IS 'Override parameters (merged with default_config from instance)';
