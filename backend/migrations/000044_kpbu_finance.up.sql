-- KPBU K2: financing structure, government support, availability payments.
CREATE TABLE IF NOT EXISTS kpbu_financings (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID          NOT NULL REFERENCES organizations(id),
    kpbu_project_id   UUID          NOT NULL REFERENCES kpbu_projects(id),
    source            VARCHAR(30)   NOT NULL,
    provider          VARCHAR(300),
    agreement_number  VARCHAR(200),
    agreement_date    DATE,
    amount            NUMERIC(20,2) NOT NULL DEFAULT 0,
    disbursed_amount  NUMERIC(20,2) NOT NULL DEFAULT 0,
    currency          VARCHAR(10)   NOT NULL DEFAULT 'IDR',
    status            VARCHAR(20)   NOT NULL DEFAULT 'PLANNED',
    notes             TEXT,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_kpbu_financings_project
    ON kpbu_financings (kpbu_project_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS kpbu_government_supports (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID          NOT NULL REFERENCES organizations(id),
    kpbu_project_id   UUID          NOT NULL REFERENCES kpbu_projects(id),
    type              VARCHAR(30)   NOT NULL,
    provider          VARCHAR(300),
    decision_number   VARCHAR(200),
    decision_date     DATE,
    amount            NUMERIC(20,2) NOT NULL DEFAULT 0,
    currency          VARCHAR(10)   NOT NULL DEFAULT 'IDR',
    status            VARCHAR(20)   NOT NULL DEFAULT 'PROPOSED',
    notes             TEXT,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_kpbu_gov_supports_project
    ON kpbu_government_supports (kpbu_project_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS kpbu_availability_payments (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID          NOT NULL REFERENCES organizations(id),
    kpbu_project_id   UUID          NOT NULL REFERENCES kpbu_projects(id),
    period_year       INT           NOT NULL,
    period_month      INT           NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    due_date          DATE,
    scheduled_amount  NUMERIC(20,2) NOT NULL DEFAULT 0,
    deduction_amount  NUMERIC(20,2) NOT NULL DEFAULT 0,
    deduction_reason  TEXT,
    paid_amount       NUMERIC(20,2) NOT NULL DEFAULT 0,
    paid_date         DATE,
    invoice_number    VARCHAR(200),
    currency          VARCHAR(10)   NOT NULL DEFAULT 'IDR',
    notes             TEXT,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_kpbu_ap_period
    ON kpbu_availability_payments (kpbu_project_id, period_year, period_month) WHERE deleted_at IS NULL;
