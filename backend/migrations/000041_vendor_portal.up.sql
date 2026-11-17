-- Migration: 000041_vendor_portal (VENDOR-PORTAL-001)
-- Contractor portal for APBN projects: vendor-linked user accounts plus
-- progress reports and documents submitted by contractors and verified by Balai.

-- A user with vendor_id set is an external contractor account. It can only
-- reach contracts whose contracts.vendor_id equals this value.
ALTER TABLE users ADD COLUMN IF NOT EXISTS vendor_id UUID REFERENCES vendors(id);
CREATE INDEX IF NOT EXISTS idx_users_vendor_id ON users (vendor_id) WHERE vendor_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS vendor_progress_reports (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id       UUID          NOT NULL REFERENCES organizations(id),
    project_id            UUID          NOT NULL REFERENCES projects(id),
    contract_id           UUID          NOT NULL REFERENCES contracts(id),
    vendor_id             UUID          NOT NULL REFERENCES vendors(id),
    period_year           INT           NOT NULL CHECK (period_year BETWEEN 2000 AND 2100),
    period_month          INT           NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    physical_progress_pct NUMERIC(5,2)  NOT NULL CHECK (physical_progress_pct BETWEEN 0 AND 100),
    planned_progress_pct  NUMERIC(5,2)  CHECK (planned_progress_pct BETWEEN 0 AND 100),
    notes                 TEXT,
    status                VARCHAR(20)   NOT NULL DEFAULT 'SUBMITTED',
    submitted_by          UUID          NOT NULL REFERENCES users(id),
    submitted_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    reviewed_by           UUID          REFERENCES users(id),
    reviewed_at           TIMESTAMPTZ,
    review_notes          TEXT,
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

-- One live (submitted or verified) report per contract per month; a rejected
-- report can be resubmitted.
CREATE UNIQUE INDEX IF NOT EXISTS uq_vendor_progress_contract_period
    ON vendor_progress_reports (contract_id, period_year, period_month)
    WHERE deleted_at IS NULL AND status <> 'REJECTED';
CREATE INDEX IF NOT EXISTS idx_vendor_progress_org_status
    ON vendor_progress_reports (organization_id, status) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS vendor_documents (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID          NOT NULL REFERENCES organizations(id),
    project_id       UUID          NOT NULL REFERENCES projects(id),
    contract_id      UUID          NOT NULL REFERENCES contracts(id),
    vendor_id        UUID          NOT NULL REFERENCES vendors(id),
    name             VARCHAR(500)  NOT NULL,
    category         VARCHAR(30)   NOT NULL,
    period_year      INT,
    period_month     INT CHECK (period_month IS NULL OR period_month BETWEEN 1 AND 12),
    file_url         TEXT          NOT NULL,
    mime_type        VARCHAR(200),
    file_size        BIGINT        NOT NULL DEFAULT 0,
    notes            TEXT,
    status           VARCHAR(20)   NOT NULL DEFAULT 'SUBMITTED',
    submitted_by     UUID          NOT NULL REFERENCES users(id),
    submitted_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    reviewed_by      UUID          REFERENCES users(id),
    reviewed_at      TIMESTAMPTZ,
    review_notes     TEXT,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_vendor_documents_contract
    ON vendor_documents (contract_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vendor_documents_org_status
    ON vendor_documents (organization_id, status) WHERE deleted_at IS NULL;
