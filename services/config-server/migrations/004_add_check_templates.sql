-- Migration 004: Add check_templates table for reusable check definitions
-- Template/Instance pattern (consistent with Alert system)

CREATE TABLE check_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    check_type VARCHAR(255) NOT NULL,
    script_content TEXT NOT NULL,
    language VARCHAR(50) NOT NULL DEFAULT 'bash',
    default_config JSONB NOT NULL DEFAULT '{}',
    description TEXT,
    version VARCHAR(50) NOT NULL,
    hash VARCHAR(64) NOT NULL,
    deleted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Unique constraint: name must be unique (ignoring soft-deleted records)
CREATE UNIQUE INDEX check_templates_name_deleted_at_key
    ON check_templates(name)
    WHERE deleted_at IS NULL;

-- Indexes for efficient querying
CREATE INDEX idx_check_templates_check_type ON check_templates(check_type);
CREATE INDEX idx_check_templates_name ON check_templates(name);
CREATE INDEX idx_check_templates_deleted_at ON check_templates(deleted_at);
CREATE INDEX idx_check_templates_hash ON check_templates(hash);

-- Comments explaining the table design
COMMENT ON TABLE check_templates IS 'Reusable check script definitions (consistent with AlertTemplate pattern)';
COMMENT ON COLUMN check_templates.name IS 'Unique template name (e.g., "disk-usage-check")';
COMMENT ON COLUMN check_templates.check_type IS 'Type of check (e.g., disk, mount, network)';
COMMENT ON COLUMN check_templates.script_content IS 'The actual script code (bash, python, etc.)';
COMMENT ON COLUMN check_templates.language IS 'Script language: bash, python, shell';
COMMENT ON COLUMN check_templates.default_config IS 'Default parameters for this check (can be overridden by instances)';
COMMENT ON COLUMN check_templates.version IS 'Template version string for tracking updates';
COMMENT ON COLUMN check_templates.hash IS 'SHA256 hash of script_content for version detection and auto-update';
