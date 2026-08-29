-- Down migration 000034: restore old lock-period unique index
-- Reverts migration 000034_governance_lock_period_fullyear_unique.up.sql

-- 1. Drop the expression unique index
DROP INDEX IF EXISTS uq_data_lock_periods_org_dataset_month;

-- 2. Drop the month sanity check (re-added by up migration)
ALTER TABLE data_lock_periods DROP CONSTRAINT IF EXISTS chk_data_lock_periods_period_month;

-- 3. Restore the legacy unique index (NULL-distinct behaviour, as before)
CREATE UNIQUE INDEX IF NOT EXISTS uq_data_lock_periods_org_dataset
    ON data_lock_periods(organization_id, dataset_type, period_year, period_month)
    WHERE deleted_at IS NULL;
