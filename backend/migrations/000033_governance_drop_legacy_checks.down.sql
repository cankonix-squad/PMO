-- Down migration 000033: restore legacy status/period_month checks
-- Reverts migration 000033_governance_drop_legacy_checks.up.sql

-- 2. Restore legacy period_month check
ALTER TABLE data_submissions ADD CONSTRAINT data_submissions_period_month_check
    CHECK (period_month BETWEEN 1 AND 12);

-- 1. Restore legacy status check (original validation-queue statuses only)
ALTER TABLE data_submissions ADD CONSTRAINT data_submissions_status_check
    CHECK (status IN ('DRAFT','SUBMITTED','VALID','REJECTED','STALE'));
