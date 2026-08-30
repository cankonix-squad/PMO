-- Migration 000038: project_periodic_reports
-- Purpose: Official periodic progress & financial input for dashboard trend (CANKORA-DASH-002)
-- Data source classification: OPERATIONAL (not yet official-governed)

CREATE TABLE IF NOT EXISTS project_periodic_reports (
    id               UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id  UUID        NOT NULL REFERENCES organizations(id),
    project_id       UUID        NOT NULL REFERENCES projects(id),
    period_year      SMALLINT    NOT NULL CHECK (period_year >= 2000 AND period_year <= 2100),
    period_month     SMALLINT    NOT NULL CHECK (period_month >= 1 AND period_month <= 12),
    physical_progress_pct NUMERIC(5,2) NOT NULL DEFAULT 0
        CHECK (physical_progress_pct >= 0 AND physical_progress_pct <= 100),
    financial_planned NUMERIC(20,2) NOT NULL DEFAULT 0
        CHECK (financial_planned >= 0),
    financial_actual  NUMERIC(20,2) NOT NULL DEFAULT 0
        CHECK (financial_actual >= 0),
    -- financial_pct is backend-computed: (financial_actual / financial_planned) * 100
    -- stored for query performance; 0 if financial_planned = 0
    financial_pct     NUMERIC(8,4)  NOT NULL DEFAULT 0,
    notes             TEXT,
    reported_by       UUID        REFERENCES users(id),
    reported_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

-- Unique active row per org + project + year + month (soft-delete aware)
CREATE UNIQUE INDEX IF NOT EXISTS uq_periodic_report_active
    ON project_periodic_reports (organization_id, project_id, period_year, period_month)
    WHERE deleted_at IS NULL;

-- Index for dashboard trend queries
CREATE INDEX IF NOT EXISTS idx_periodic_report_org_period
    ON project_periodic_reports (organization_id, period_year, period_month)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_periodic_report_project
    ON project_periodic_reports (project_id, period_year, period_month)
    WHERE deleted_at IS NULL;
