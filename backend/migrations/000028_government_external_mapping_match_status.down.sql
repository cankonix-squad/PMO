-- Down migration 000028: revert match_status + internal_entity_id nullable changes
-- WARNING: this restores internal_entity_id to NOT NULL, which will fail if any
-- row has internal_entity_id = NULL.  Run this only in controlled environments.

-- Step 1: remove match_status index
DROP INDEX IF EXISTS idx_gov_map_match_status;

-- Step 2: restore entity index (non-partial)
DROP INDEX IF EXISTS idx_gov_map_entity;
CREATE INDEX idx_gov_map_entity ON government_external_mappings(
    organization_id, internal_entity_type, internal_entity_id
);

-- Step 3: drop match_status column
ALTER TABLE government_external_mappings
    DROP COLUMN IF EXISTS match_status;

-- Step 4: restore internal_entity_id NOT NULL constraint
-- This will fail if any NULL rows exist — intentional: caller must resolve them first.
ALTER TABLE government_external_mappings
    ALTER COLUMN internal_entity_id SET NOT NULL;
