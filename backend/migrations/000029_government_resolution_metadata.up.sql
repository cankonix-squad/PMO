-- Migration 000029: government_external_mappings — entity resolution metadata
-- CANKORA-P3-002: Government Entity Resolution (PENDING_MATCH → MATCHED)
--
-- Adds resolution tracking fields to government_external_mappings:
--   match_confidence  — numeric confidence score (0-100) from candidate matching
--   match_reason      — human-readable reason code (EXACT_CODE, EXACT_NAME, etc.)
--   matched_by        — UUID of the user who confirmed the match (nullable)
--   matched_at        — timestamp when match was confirmed (nullable)
--   rejected_by       — UUID of the user who rejected the mapping (nullable)
--   rejected_at       — timestamp of rejection (nullable)
--   reject_reason     — free-text reason for rejection (nullable)
--
-- match_status was already added in migration 000028 (HARDEN-001).
-- internal_entity_id was already made nullable in migration 000028.

ALTER TABLE government_external_mappings
    ADD COLUMN IF NOT EXISTS match_confidence  SMALLINT,
    ADD COLUMN IF NOT EXISTS match_reason      VARCHAR(50),
    ADD COLUMN IF NOT EXISTS matched_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS matched_at        TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS rejected_by       UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS rejected_at       TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reject_reason     TEXT;

-- Index: filter MATCHED records by internal entity (analytics / reporting use)
CREATE INDEX IF NOT EXISTS idx_gov_map_matched
    ON government_external_mappings(organization_id, internal_entity_type, internal_entity_id)
    WHERE match_status = 'MATCHED' AND internal_entity_id IS NOT NULL;

-- Index: resolution queue — find pending mappings quickly per dataset type
CREATE INDEX IF NOT EXISTS idx_gov_map_pending_dataset
    ON government_external_mappings(organization_id, dataset_type, match_status)
    WHERE match_status = 'PENDING_MATCH';
