-- Migration 000032: data governance — fix period_month NULL + drop snapshot unique
-- CANKORA-P3-003: Official Data Validation & Approval Workflow
--
-- Migration 000031 relaxed project_id/snapshot_id NOT NULL but left:
--   * period_month NOT NULL — governance full-year submissions need NULL month
--   * UNIQUE constraint on snapshot_id (data_submissions_snapshot_id_key) —
--     harmless for NULL governance rows, but unnecessary restriction; dropped
--     so snapshot linkage remains purely referential.

-- 1. period_month nullable (full-year governance submissions)
ALTER TABLE data_submissions ALTER COLUMN period_month DROP NOT NULL;

-- 2. Drop the GORM-created unique constraint on snapshot_id (keep plain index)
ALTER TABLE data_submissions DROP CONSTRAINT IF EXISTS data_submissions_snapshot_id_key;
CREATE INDEX IF NOT EXISTS idx_data_submissions_snapshot_id
    ON data_submissions(snapshot_id) WHERE snapshot_id IS NOT NULL;

-- 3. Guard the relaxed CHECK (already applied by 000031; idempotent safety)
ALTER TABLE data_submissions
    DROP CONSTRAINT IF EXISTS chk_data_submissions_period_month,
    ADD CONSTRAINT chk_data_submissions_period_month
        CHECK (period_month IS NULL OR (period_month BETWEEN 1 AND 12));
