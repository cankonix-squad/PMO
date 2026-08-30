-- Down migration 000035: drop created_by from data_submissions
-- Reverts migration 000035_governance_add_created_by.up.sql

DROP INDEX IF EXISTS idx_data_submissions_created_by;
ALTER TABLE data_submissions DROP COLUMN IF EXISTS created_by;
