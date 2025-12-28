-- Migration 008: Decouple alert_rules from templates
-- Rules now contain a snapshot of template content at creation time
-- This allows templates to be modified or deleted without affecting existing rules

-- Add template fields (copied from template at creation time)
ALTER TABLE alert_rules
    ADD COLUMN name VARCHAR(255),
    ADD COLUMN description TEXT,
    ADD COLUMN severity VARCHAR(20),
    ADD COLUMN query_template TEXT,
    ADD COLUMN default_config JSONB DEFAULT '{}';

-- Add metadata fields (for tracking origin template)
ALTER TABLE alert_rules
    ADD COLUMN created_from_template_id VARCHAR(255),
    ADD COLUMN created_from_template_name VARCHAR(255);

-- Add soft delete support (consistent with check_instances)
ALTER TABLE alert_rules
    ADD COLUMN deleted_at TIMESTAMP;

-- Drop template_id index
DROP INDEX IF EXISTS idx_alert_rules_template;

-- Drop foreign key constraint
DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'alert_rules'::regclass
    AND contype = 'f'
    AND confrelid = 'alert_templates'::regclass;

    IF constraint_name IS NOT NULL THEN
        EXECUTE 'ALTER TABLE alert_rules DROP CONSTRAINT ' || constraint_name;
    END IF;
END $$;

-- Drop template_id column
ALTER TABLE alert_rules DROP COLUMN IF EXISTS template_id;

-- Add NOT NULL constraints to required template fields
-- (After data population in production, but since we have no data, we can do it now)
ALTER TABLE alert_rules
    ALTER COLUMN name SET NOT NULL,
    ALTER COLUMN severity SET NOT NULL,
    ALTER COLUMN query_template SET NOT NULL;

-- Add check constraint for severity
ALTER TABLE alert_rules
    ADD CONSTRAINT check_alert_rules_severity
    CHECK (severity IN ('critical', 'warning', 'info'));

-- Create new indexes for template fields
CREATE INDEX idx_alert_rules_name ON alert_rules(name);
CREATE INDEX idx_alert_rules_severity ON alert_rules(severity);
CREATE INDEX idx_alert_rules_created_from_template_id ON alert_rules(created_from_template_id);
CREATE INDEX idx_alert_rules_deleted_at ON alert_rules(deleted_at);

-- Update table comment
COMMENT ON TABLE alert_rules IS 'Alert rules containing snapshots of alert templates. Independent from templates after creation.';
COMMENT ON COLUMN alert_rules.name IS 'Alert name (copied from template at creation)';
COMMENT ON COLUMN alert_rules.description IS 'Alert description (copied from template)';
COMMENT ON COLUMN alert_rules.severity IS 'Alert severity: critical, warning, or info';
COMMENT ON COLUMN alert_rules.query_template IS 'PromQL query template (copied from template)';
COMMENT ON COLUMN alert_rules.default_config IS 'Default configuration (copied from template)';
COMMENT ON COLUMN alert_rules.created_from_template_id IS 'Optional: ID of template this was created from';
COMMENT ON COLUMN alert_rules.created_from_template_name IS 'Optional: Name of template at creation time';
COMMENT ON COLUMN alert_rules.deleted_at IS 'Soft delete timestamp (NULL = active)';
COMMENT ON COLUMN alert_rules.config IS 'Override parameters (merged with default_config based on merge_strategy)';
