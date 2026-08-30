-- Down migration 000031: revert data governance FK relaxation
-- Reverts migration 000031_governance_relax_fk.up.sql

-- 4. Drop governance composite index
DROP INDEX IF EXISTS idx_data_submissions_governance;

-- 3. Restore period_month NOT NULL with original CHECK
ALTER TABLE data_submissions
    DROP CONSTRAINT IF EXISTS chk_data_submissions_period_month,
    ADD CONSTRAINT chk_data_submissions_period_month
        CHECK (period_month BETWEEN 1 AND 12);
ALTER TABLE data_submissions ALTER COLUMN period_month SET NOT NULL;

-- 2. Restore project_id / snapshot_id NOT NULL
ALTER TABLE data_submissions
    ALTER COLUMN project_id SET NOT NULL,
    ALTER COLUMN snapshot_id SET NOT NULL;

-- 1. Restore unique index on snapshot_id
DROP INDEX IF EXISTS idx_data_submissions_snapshot;
CREATE UNIQUE INDEX IF NOT EXISTS idx_data_submissions_snapshot_unique
    ON data_submissions(snapshot_id);
