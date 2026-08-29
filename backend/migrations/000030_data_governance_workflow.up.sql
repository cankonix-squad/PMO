-- Migration 000030: data governance — official validation & approval workflow
-- CANKORA-P3-003: Official Data Validation & Approval Workflow
--
-- Adds:
--   1. data_submission_items — line items of a submission (entity + payload + validation status)
--   2. data_lock_periods     — period-level lock so approved data cannot be changed
--
-- NOTE: the parent table `data_submissions` already exists (migration P1-012 /
-- dataquality module). This migration ADDS columns to support the official
-- approval FSM (IN_REVIEW / APPROVED / LOCKED states) without breaking the
-- existing snapshot-validation queue.

-- ---------------------------------------------------------------------------
-- 1. Extend data_submissions with official-approval fields
-- ---------------------------------------------------------------------------
ALTER TABLE data_submissions
    ADD COLUMN IF NOT EXISTS dataset_type      VARCHAR(50) NOT NULL DEFAULT 'OTHER',
    ADD COLUMN IF NOT EXISTS source_type       VARCHAR(50) NOT NULL DEFAULT 'MANUAL',
    ADD COLUMN IF NOT EXISTS source_entity_type VARCHAR(100),
    ADD COLUMN IF NOT EXISTS source_entity_id   UUID,
    ADD COLUMN IF NOT EXISTS reviewed_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS reviewed_at        TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS review_notes       TEXT,
    ADD COLUMN IF NOT EXISTS approved_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS approved_at        TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS locked_by          UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS locked_at          TIMESTAMPTZ;

-- Status check constraint: extend allowed values with IN_REVIEW / APPROVED / LOCKED / CANCELLED
ALTER TABLE data_submissions DROP CONSTRAINT IF EXISTS chk_data_submissions_status;
ALTER TABLE data_submissions ADD CONSTRAINT chk_data_submissions_status CHECK (
    status IN ('DRAFT','SUBMITTED','IN_REVIEW','APPROVED','REJECTED','LOCKED','CANCELLED','VALID','STALE')
);

-- Indexes for operational queries
CREATE INDEX IF NOT EXISTS idx_data_submissions_dataset
    ON data_submissions(organization_id, dataset_type, status);
CREATE INDEX IF NOT EXISTS idx_data_submissions_period
    ON data_submissions(organization_id, period_year, period_month);
CREATE INDEX IF NOT EXISTS idx_data_submissions_source
    ON data_submissions(organization_id, source_type);

-- ---------------------------------------------------------------------------
-- 2. data_submission_items
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS data_submission_items (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id      UUID NOT NULL REFERENCES data_submissions(id) ON DELETE CASCADE,
    entity_type        VARCHAR(100) NOT NULL,
    entity_id          UUID,
    action             VARCHAR(20)  NOT NULL DEFAULT 'CREATE', -- CREATE | UPDATE | DELETE | UPSERT | VALIDATE_ONLY
    payload_before     JSONB,
    payload_after      JSONB NOT NULL DEFAULT '{}'::jsonb,
    validation_status  VARCHAR(20)  NOT NULL DEFAULT 'PENDING', -- PENDING | VALID | INVALID
    validation_errors  JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_data_submission_items_submission
    ON data_submission_items(submission_id);
CREATE INDEX IF NOT EXISTS idx_data_submission_items_entity
    ON data_submission_items(entity_type, entity_id);

-- ---------------------------------------------------------------------------
-- 3. data_lock_periods
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS data_lock_periods (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    dataset_type    VARCHAR(50) NOT NULL,
    period_year     INT NOT NULL,
    period_month    INT, -- NULL = full-year lock
    status          VARCHAR(20) NOT NULL DEFAULT 'OPEN', -- OPEN | LOCKED
    locked_by       UUID REFERENCES users(id) ON DELETE SET NULL,
    locked_at       TIMESTAMPTZ,
    lock_reason     TEXT,
    created_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- A period is uniquely identified by (org, dataset, year, month) — but only one
-- lock row per period is allowed on non-deleted rows.
CREATE UNIQUE INDEX IF NOT EXISTS uq_data_lock_periods_org_dataset
    ON data_lock_periods(organization_id, dataset_type, period_year, period_month)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_data_lock_periods_status
    ON data_lock_periods(organization_id, status);
