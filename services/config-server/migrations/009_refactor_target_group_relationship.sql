-- Migration 009: Refactor Target-Group Relationship
-- Replace Primary/Secondary pattern with unified junction table + Default Own Group

-- Step 1: Add is_default_own column to groups table
ALTER TABLE groups
    ADD COLUMN is_default_own BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX idx_groups_is_default_own ON groups(is_default_own);

COMMENT ON COLUMN groups.is_default_own IS 'True if this is an auto-created default group for a target';

-- Step 2: Create unified target_groups junction table (if not exists)
CREATE TABLE IF NOT EXISTS target_groups (
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    is_default_own BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (target_id, group_id)
);

CREATE INDEX idx_target_groups_target_id ON target_groups(target_id);
CREATE INDEX idx_target_groups_group_id ON target_groups(group_id);
CREATE INDEX idx_target_groups_is_default_own ON target_groups(is_default_own);

COMMENT ON TABLE target_groups IS 'Junction table for many-to-many relationship between targets and groups';
COMMENT ON COLUMN target_groups.target_id IS 'Reference to target';
COMMENT ON COLUMN target_groups.group_id IS 'Reference to group';
COMMENT ON COLUMN target_groups.is_default_own IS 'True if this is the auto-created default group mapping';

-- Step 3: Create database trigger to prevent orphan targets
CREATE OR REPLACE FUNCTION prevent_target_orphan()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT COUNT(*) FROM target_groups WHERE target_id = OLD.target_id) <= 1 THEN
        RAISE EXCEPTION 'Cannot remove last group mapping for target (target_id: %)', OLD.target_id;
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_prevent_target_orphan
    BEFORE DELETE ON target_groups
    FOR EACH ROW
    EXECUTE FUNCTION prevent_target_orphan();

-- Step 4: Migrate data from primary_group_id to target_groups
-- Note: This assumes no data exists yet. If data exists, uncomment and modify:
/*
INSERT INTO target_groups (target_id, group_id, is_default_own, created_at)
SELECT
    id as target_id,
    primary_group_id as group_id,
    false as is_default_own,
    created_at
FROM targets
WHERE primary_group_id IS NOT NULL
ON CONFLICT (target_id, group_id) DO NOTHING;

-- Migrate secondary groups if target_secondary_groups table exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'target_secondary_groups') THEN
        INSERT INTO target_groups (target_id, group_id, is_default_own, created_at)
        SELECT
            target_id,
            group_id,
            false as is_default_own,
            CURRENT_TIMESTAMP
        FROM target_secondary_groups
        ON CONFLICT (target_id, group_id) DO NOTHING;
    END IF;
END $$;
*/

-- Step 5: Drop old tables and columns
DROP TABLE IF EXISTS target_secondary_groups;

-- Drop foreign key constraint on primary_group_id
DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'targets'::regclass
    AND contype = 'f'
    AND confrelid = 'groups'::regclass
    AND conkey = ARRAY[(SELECT attnum FROM pg_attribute
                        WHERE attrelid = 'targets'::regclass
                        AND attname = 'primary_group_id')];

    IF constraint_name IS NOT NULL THEN
        EXECUTE 'ALTER TABLE targets DROP CONSTRAINT ' || constraint_name;
    END IF;
END $$;

-- Drop primary_group_id column
ALTER TABLE targets DROP COLUMN IF EXISTS primary_group_id;

-- Step 6: Update table comments
COMMENT ON TABLE targets IS 'Monitoring targets (servers, services). Groups are managed via target_groups junction table.';
