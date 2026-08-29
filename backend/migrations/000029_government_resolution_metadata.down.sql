-- Down migration 000029: remove entity resolution metadata fields
-- Reverts migration 000029_government_resolution_metadata.up.sql

DROP INDEX IF EXISTS idx_gov_map_matched;
DROP INDEX IF EXISTS idx_gov_map_pending_dataset;

ALTER TABLE government_external_mappings
    DROP COLUMN IF EXISTS match_confidence,
    DROP COLUMN IF EXISTS match_reason,
    DROP COLUMN IF EXISTS matched_by,
    DROP COLUMN IF EXISTS matched_at,
    DROP COLUMN IF EXISTS rejected_by,
    DROP COLUMN IF EXISTS rejected_at,
    DROP COLUMN IF EXISTS reject_reason;
