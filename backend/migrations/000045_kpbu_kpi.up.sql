-- KPBU K3: service KPIs and monthly measurements (internal or via BUP portal).
CREATE TABLE IF NOT EXISTS kpbu_kpis (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID          NOT NULL REFERENCES organizations(id),
    kpbu_project_id  UUID          NOT NULL REFERENCES kpbu_projects(id),
    code             VARCHAR(30)   NOT NULL,
    name             VARCHAR(300)  NOT NULL,
    description      TEXT,
    unit             VARCHAR(50),
    target_value     NUMERIC(20,4) NOT NULL,
    direction        VARCHAR(20)   NOT NULL DEFAULT 'HIGHER_BETTER',
    deduction_pct    NUMERIC(6,2)  NOT NULL DEFAULT 0 CHECK (deduction_pct BETWEEN 0 AND 100),
    is_active        BOOLEAN       NOT NULL DEFAULT TRUE,
    sort_order       INT           NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_kpbu_kpis_code
    ON kpbu_kpis (kpbu_project_id, code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS kpbu_kpi_measurements (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID          NOT NULL REFERENCES organizations(id),
    kpbu_project_id  UUID          NOT NULL REFERENCES kpbu_projects(id),
    kpi_id           UUID          NOT NULL REFERENCES kpbu_kpis(id),
    period_year      INT           NOT NULL,
    period_month     INT           NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    actual_value     NUMERIC(20,4) NOT NULL,
    notes            TEXT,
    source           VARCHAR(20)   NOT NULL DEFAULT 'INTERNAL',
    status           VARCHAR(20)   NOT NULL DEFAULT 'SUBMITTED',
    submitted_by     UUID          REFERENCES users(id),
    submitted_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    reviewed_by      UUID          REFERENCES users(id),
    reviewed_at      TIMESTAMPTZ,
    review_notes     TEXT,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);
-- One live measurement per KPI and period; a rejected one can be resubmitted.
CREATE UNIQUE INDEX IF NOT EXISTS uq_kpbu_kpi_measure_period
    ON kpbu_kpi_measurements (kpi_id, period_year, period_month)
    WHERE deleted_at IS NULL AND status <> 'REJECTED';
CREATE INDEX IF NOT EXISTS idx_kpbu_kpi_measure_project
    ON kpbu_kpi_measurements (kpbu_project_id, status) WHERE deleted_at IS NULL;
