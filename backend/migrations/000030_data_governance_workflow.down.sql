-- Down migration 000030: remove data governance workflow additions
-- Reverts migration 000030_data_governance_workflow.up.sql

DROP TABLE IF EXISTS data_lock_periods;
DROP TABLE IF EXISTS data_submission_items;

ALTER TABLE data_submissions DROP CONSTRAINT IF EXISTS chk_data_submissions_status;
ALTER TABLE data_submissions ADD CONSTRAINT chk_data_submissions_status CHECK (
    status IN ('DRAFT','SUBMITTED','VALID','REJECTED','STALE')
);

DROP INDEX IF EXISTS idx_data_submissions_dataset;
DROP INDEX IF EXISTS idx_data_submissions_period;
DROP INDEX IF EXISTS idx_data_submissions_source;

ALTER TABLE data_submissions
    DROP COLUMN IF EXISTS dataset_type,
    DROP COLUMN IF EXISTS source_type,
    DROP COLUMN IF EXISTS source_entity_type,
    DROP COLUMN IF EXISTS source_entity_id,
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS review_notes,
    DROP COLUMN IF EXISTS approved_by,
    DROP COLUMN IF EXISTS approved_at,
    DROP COLUMN IF EXISTS locked_by,
    DROP COLUMN IF EXISTS locked_at;
