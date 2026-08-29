-- Migration 000034: data governance — enforce lock-period uniqueness incl. full-year
-- CANKORA-P3-003-HARDEN: Official Data Validation Hardening
--
-- Problem: the existing unique index uq_data_lock_periods_org_dataset is
--   (organization_id, dataset_type, period_year, period_month) WHERE deleted_at IS NULL
-- PostgreSQL treats NULLs as distinct, so multiple full-year lock rows
-- (period_month = NULL) for the same (org, dataset, year) are ALLOWED —
-- the "one lock per period" invariant is broken.
--
-- Fix: an expression index over COALESCE(period_month, 0) so that:
--   * full-year lock  (period_month NULL)  → key month = 0
--   * monthly lock    (period_month = N)   → key month = N
-- No two lock rows may share (org, dataset, year, month-key) on live rows.
--
-- The old unique index is dropped and replaced by the expression index.
-- Monthly uniqueness (month 1..12) is preserved; full-year uniqueness is now
-- also enforced. The chk_data_lock_periods_period_month CHECK keeps month 1..12.

-- 1. Drop the old unique index (NULL-distinct issue)
DROP INDEX IF EXISTS uq_data_lock_periods_org_dataset;

-- 2. Enforce month sanity (idempotent; NULL allowed = full-year)
ALTER TABLE data_lock_periods
    DROP CONSTRAINT IF EXISTS chk_data_lock_periods_period_month,
    ADD CONSTRAINT chk_data_lock_periods_period_month
        CHECK (period_month IS NULL OR (period_month BETWEEN 1 AND 12));

-- 3. Expression unique index: COALESCE(period_month, 0)
CREATE UNIQUE INDEX IF NOT EXISTS uq_data_lock_periods_org_dataset_month
    ON data_lock_periods(organization_id, dataset_type, period_year, COALESCE(period_month, 0))
    WHERE deleted_at IS NULL;
