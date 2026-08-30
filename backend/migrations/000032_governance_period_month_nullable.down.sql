-- Down migration 000032: revert governance period_month NULL relaxation
-- Reverts migration 000032_governance_period_month_nullable.up.sql

-- 3. Restore original CHECK (period_month NOT NULL with 1..12)
ALTER TABLE data_submissions
    DROP CONSTRAINT IF EXISTS chk_data_submissions_period_month,
    ADD CONSTRAINT chk_data_submissions_period_month
        CHECK (period_month BETWEEN 1 AND 12);

-- 2. Restore unique constraint on snapshot_id
DROP INDEX IF EXISTS idx_data_submissions_snapshot_id;
ALTER TABLE data_submissions
    ADD CONSTRAINT data_submissions_snapshot_id_key UNIQUE (snapshot_id);

-- 1. Restore period_month NOT NULL
ALTER TABLE data_submissions ALTER COLUMN period_month SET NOT NULL;
