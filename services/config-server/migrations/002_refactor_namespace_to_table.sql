-- Refactor namespace from string constant to database table
-- This enables dynamic namespace management and future extensibility

BEGIN;

-- Step 1: Create namespaces table
CREATE TABLE namespaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    policy_priority INTEGER NOT NULL, -- Lower value = higher priority
    merge_strategy VARCHAR(20) NOT NULL DEFAULT 'merge' CHECK (merge_strategy IN ('override', 'merge', 'append')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Step 2: Insert default namespaces
INSERT INTO namespaces (name, description, policy_priority, merge_strategy) VALUES
    ('infrastructure', 'Physical/cloud infrastructure organization (e.g., aws/us-east-1/zone-a, datacenter/rack-01)', 100, 'merge'),
    ('logical', 'Logical organization (e.g., project/team, cluster/workload)', 50, 'merge'),
    ('environment', 'Deployment environments (e.g., production, staging, development) - Highest priority', 10, 'override');

-- Step 3: Add namespace_id column to groups table (nullable initially for migration)
ALTER TABLE groups ADD COLUMN namespace_id UUID;

-- Step 4: Migrate existing namespace strings to namespace_id references
UPDATE groups SET namespace_id = (SELECT id FROM namespaces WHERE name = 'infrastructure') WHERE namespace = 'infrastructure';
UPDATE groups SET namespace_id = (SELECT id FROM namespaces WHERE name = 'logical') WHERE namespace = 'logical';
UPDATE groups SET namespace_id = (SELECT id FROM namespaces WHERE name = 'environment') WHERE namespace = 'environment';

-- Step 5: Make namespace_id NOT NULL and add foreign key constraint
-- Note: As per design decision, we don't use FK constraints, but add index for performance
ALTER TABLE groups ALTER COLUMN namespace_id SET NOT NULL;

-- Step 6: Drop old namespace column
ALTER TABLE groups DROP COLUMN namespace;

-- Step 7: Create indexes for performance
CREATE INDEX idx_groups_namespace_id ON groups(namespace_id);
CREATE INDEX idx_namespaces_name ON namespaces(name);
CREATE INDEX idx_namespaces_policy_priority ON namespaces(policy_priority);

-- Step 8: Create updated_at trigger for namespaces
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_namespaces_updated_at BEFORE UPDATE ON namespaces
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;
