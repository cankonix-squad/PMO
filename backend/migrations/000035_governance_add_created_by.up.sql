-- Migration 000035: data governance — add created_by to data_submissions
-- CANKORA-P3-003-HARDEN: Official Data Validation Hardening
--
-- The DRAFT submission must record who created it independently of
-- submitted_by (which is only populated at Submit time). dataquality rows
-- (existing) keep created_by NULL — compatible with the legacy validation
-- queue, which uses validator_id/submitted_by only.

ALTER TABLE data_submissions
    ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE SET NULL;

-- Index: find submissions created by a user (audit/ownership lookups)
CREATE INDEX IF NOT EXISTS idx_data_submissions_created_by
    ON data_submissions(created_by)
    WHERE created_by IS NOT NULL;
