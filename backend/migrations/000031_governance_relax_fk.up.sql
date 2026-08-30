-- Migration 000031: data governance — relax data_submissions FK/unique constraints
-- CANKORA-P3-003: Official Data Validation & Approval Workflow
--
-- The base `data_submissions` table (migration 000016, dataquality module) was
-- designed for the snapshot validation queue: project_id and snapshot_id are
-- NOT NULL and snapshot_id is UNIQUE (one submission per snapshot).
--
-- The governance module shares this table for STANDALONE official submissions
-- that are not tied to a monitoring snapshot. Those columns must therefore be
-- nullable, and the unique constraint on snapshot_id must be dropped so a
-- NULL snapshot (governance) and multiple snapshot submissions (dataquality)
-- can coexist.
--
-- period_month is also relaxed (NULL allowed) so a governance submission can
-- cover a full-year period without a specific month. The old CHECK forced
-- 1..12 on every row, which conflicts with NULL for full-year locks.

-- 1. Drop the unique index on snapshot_id (kept as a non-unique index)
DROP INDEX IF EXISTS idx_data_submissions_snapshot_unique;
ALTER TABLE data_submissions DROP CONSTRAINT IF EXISTS uq_data_submissions_snapshot;

-- 2. Make project_id / snapshot_id nullable (FKs are kept, SET NULL on delete)
ALTER TABLE data_submissions
    ALTER COLUMN project_id DROP NOT NULL,
    ALTER COLUMN snapshot_id DROP NOT NULL;

-- Recreate a non-unique index for snapshot lookups (the original unique index
-- was unnamed per GORM convention in migration 000016; guard both names).
CREATE INDEX IF NOT EXISTS idx_data_submissions_snapshot ON data_submissions(snapshot_id) WHERE snapshot_id IS NOT NULL;

-- 3. Relax period_month: allow NULL (full-year period), keep 1..12 when set
ALTER TABLE data_submissions
    DROP CONSTRAINT IF EXISTS chk_data_submissions_period_month,
    ADD CONSTRAINT chk_data_submissions_period_month
        CHECK (period_month IS NULL OR (period_month BETWEEN 1 AND 12));

-- 4. Fresh composite index for governance operational queries
CREATE INDEX IF NOT EXISTS idx_data_submissions_governance
    ON data_submissions(organization_id, dataset_type, period_year, status)
    WHERE deleted_at IS NULL;
