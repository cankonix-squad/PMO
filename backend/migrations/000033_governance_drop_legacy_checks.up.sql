-- Migration 000033: data governance — drop legacy status/period_month checks
-- CANKORA-P3-003: Official Data Validation & Approval Workflow
--
-- Migration 000030 added chk_data_submissions_status with the extended FSM
-- statuses but the ORIGINAL GORM auto-named constraint
-- data_submissions_status_check (migration 000016, statuses limited to
-- DRAFT/SUBMITTED/VALID/REJECTED/STALE) was never dropped. Both now coexist,
-- so transitioning to IN_REVIEW/APPROVED/LOCKED/CANCELLED violates the old
-- check (SQLSTATE 23514).
--
-- The legacy data_submissions_period_month_check (period_month 1..12 NOT NULL)
-- is also superseded by chk_data_submissions_period_month (NULL allowed for
-- full-year governance periods).

-- 1. Drop legacy status check (superseded by chk_data_submissions_status)
ALTER TABLE data_submissions DROP CONSTRAINT IF EXISTS data_submissions_status_check;

-- 2. Drop legacy period_month check (superseded by chk_data_submissions_period_month)
ALTER TABLE data_submissions DROP CONSTRAINT IF EXISTS data_submissions_period_month_check;
