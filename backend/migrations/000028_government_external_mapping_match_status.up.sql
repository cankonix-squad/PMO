-- Migration 000028: government_external_mappings — nullable internal_entity_id + match_status
-- CANKORA-HARDEN-001: Government lineage integrity
--
-- Previously internal_entity_id was NOT NULL and the ingestor wrote a random uuid.New()
-- placeholder, fabricating a fake lineage link.  This migration:
--   1. Makes internal_entity_id nullable — NULL means "not yet resolved".
--   2. Adds match_status VARCHAR(30) NOT NULL DEFAULT 'PENDING_MATCH'.
--   3. Resets any existing rows that have a non-nil internal_entity_id which was
--      written by the fake-UUID code to PENDING_MATCH / NULL, because those UUIDs
--      do not correspond to real CANKORA entities.
--   4. Drops and recreates the entity index to cover the nullable column cleanly.
--
-- MatchStatus values:
--   PENDING_MATCH  — external record received, not yet linked to an internal entity
--   MATCHED        — external record resolved to a real internal entity

-- Step 1: drop the NOT NULL constraint by altering the column to be nullable
ALTER TABLE government_external_mappings
    ALTER COLUMN internal_entity_id DROP NOT NULL;

-- Step 2: add match_status column
ALTER TABLE government_external_mappings
    ADD COLUMN IF NOT EXISTS match_status VARCHAR(30) NOT NULL DEFAULT 'PENDING_MATCH';

-- Step 3: reset rows that had a fake-UUID placeholder to honest PENDING_MATCH / NULL.
--   Any row created before this migration was written with uuid.New() — a random UUID
--   that does not correspond to any real project/budget/location/vendor row.
--   We cannot distinguish "real" matches from fake ones, so we conservatively reset all
--   existing rows.  A future reconciliation step will re-match them properly.
UPDATE government_external_mappings
SET    internal_entity_id = NULL,
       match_status       = 'PENDING_MATCH'
WHERE  match_status = 'PENDING_MATCH'   -- all existing rows (column just added, so all = default)
   OR  match_status IS NULL;

-- Step 4: rebuild the entity index (internal_entity_id is now nullable)
DROP INDEX IF EXISTS idx_gov_map_entity;
CREATE INDEX idx_gov_map_entity ON government_external_mappings(
    organization_id, internal_entity_type, internal_entity_id
) WHERE internal_entity_id IS NOT NULL;

-- Step 5: index on match_status for efficient PENDING_MATCH queries
CREATE INDEX IF NOT EXISTS idx_gov_map_match_status
    ON government_external_mappings(organization_id, match_status);
