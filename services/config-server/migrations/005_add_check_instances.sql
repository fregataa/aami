-- Migration 005: Add check_instances table for applying templates at specific scopes
-- Template/Instance pattern (consistent with Alert system: AlertTemplate/AlertRule)

CREATE TABLE check_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES check_templates(id) ON DELETE CASCADE,
    scope VARCHAR(20) NOT NULL CHECK (scope IN ('global', 'namespace', 'group')),
    namespace_id UUID REFERENCES namespaces(id) ON DELETE CASCADE,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    config JSONB NOT NULL DEFAULT '{}',
    priority INTEGER NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT true,
    deleted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Scope validation: ensure proper hierarchy
    CHECK (
        (scope = 'global' AND namespace_id IS NULL AND group_id IS NULL) OR
        (scope = 'namespace' AND namespace_id IS NOT NULL AND group_id IS NULL) OR
        (scope = 'group' AND namespace_id IS NOT NULL AND group_id IS NOT NULL)
    )
);

-- Unique constraint: one instance per template per scope (ignoring soft-deleted records)
CREATE UNIQUE INDEX check_instances_unique_global
    ON check_instances(template_id)
    WHERE scope = 'global' AND deleted_at IS NULL;

CREATE UNIQUE INDEX check_instances_unique_namespace
    ON check_instances(template_id, namespace_id)
    WHERE scope = 'namespace' AND deleted_at IS NULL;

CREATE UNIQUE INDEX check_instances_unique_group
    ON check_instances(template_id, group_id)
    WHERE scope = 'group' AND deleted_at IS NULL;

-- Indexes for efficient querying
CREATE INDEX idx_check_instances_template_id ON check_instances(template_id);
CREATE INDEX idx_check_instances_scope ON check_instances(scope);
CREATE INDEX idx_check_instances_namespace_id ON check_instances(namespace_id);
CREATE INDEX idx_check_instances_group_id ON check_instances(group_id);
CREATE INDEX idx_check_instances_is_active ON check_instances(is_active);
CREATE INDEX idx_check_instances_deleted_at ON check_instances(deleted_at);
CREATE INDEX idx_check_instances_priority ON check_instances(priority);

-- Comments explaining the table design
COMMENT ON TABLE check_instances IS 'Template applications at specific scopes: Global/Namespace/Group (consistent with AlertRule pattern)';
COMMENT ON COLUMN check_instances.template_id IS 'References check_templates.id';
COMMENT ON COLUMN check_instances.scope IS 'Application scope: global (all nodes), namespace (namespace nodes), group (group nodes)';
COMMENT ON COLUMN check_instances.namespace_id IS 'NULL for global, references namespace for namespace/group level';
COMMENT ON COLUMN check_instances.group_id IS 'NULL for global/namespace, references group for group level';
COMMENT ON COLUMN check_instances.config IS 'Override parameters (merged with template default_config)';
COMMENT ON COLUMN check_instances.priority IS 'Priority for conflict resolution (lower = higher priority)';
COMMENT ON COLUMN check_instances.is_active IS 'Whether the instance is currently active';
