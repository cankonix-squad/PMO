-- Migration: 000013_report_snapshots
-- Purpose: Periodic reporting snapshots for weekly/monthly/quarterly reports (P1-007)

CREATE TYPE report_period AS ENUM ('WEEKLY', 'MONTHLY', 'QUARTERLY');
CREATE TYPE report_status AS ENUM ('DRAFT', 'PUBLISHED', 'ARCHIVED');

CREATE TABLE IF NOT EXISTS report_snapshots (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Period
    period_type         report_period NOT NULL,
    period_label        VARCHAR(20) NOT NULL,  -- e.g. "2026-W34", "2026-08", "2026-Q3"
    period_start        DATE        NOT NULL,
    period_end          DATE        NOT NULL,

    -- Scope (optional: scoped to a single project, or org-wide if NULL)
    project_id          UUID        REFERENCES projects(id) ON DELETE SET NULL,

    -- Summary metrics (JSON blob for flexibility)
    metrics             JSONB       NOT NULL DEFAULT '{}',

    -- Executive summary text
    executive_summary   TEXT,

    -- Status lifecycle
    status              report_status NOT NULL DEFAULT 'DRAFT',

    -- Export metadata
    export_format       VARCHAR(20),    -- 'PDF', 'XLSX', etc. (Phase 2)
    export_url          TEXT,           -- object storage URL (Phase 2)

    -- Audit
    created_by          UUID        NOT NULL REFERENCES users(id),
    published_at        TIMESTAMPTZ,
    published_by        UUID        REFERENCES users(id),

    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_report_snapshots_org        ON report_snapshots(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_report_snapshots_project    ON report_snapshots(project_id)      WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_report_snapshots_period     ON report_snapshots(organization_id, period_type, period_start) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_report_snapshots_status     ON report_snapshots(status)           WHERE deleted_at IS NULL;
