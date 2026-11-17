-- Migration: 000040_kpbu_foundation (KPBU-K1)
-- KPBU (PPP) project registry, lifecycle stage targets, and parties.

CREATE TABLE IF NOT EXISTS kpbu_projects (
    id                       UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id          UUID          NOT NULL REFERENCES organizations(id),
    project_id               UUID          REFERENCES projects(id),
    code                     VARCHAR(50)   NOT NULL,
    name                     VARCHAR(300)  NOT NULL,
    description              TEXT,
    sector                   VARCHAR(50)   NOT NULL,
    stage                    VARCHAR(30)   NOT NULL DEFAULT 'IDENTIFIED',
    previous_stage           VARCHAR(30),
    initiative               VARCHAR(20)   NOT NULL DEFAULT 'SOLICITED',
    return_scheme            VARCHAR(30)   NOT NULL DEFAULT 'AVAILABILITY_PAYMENT',
    pjpk_name                VARCHAR(200),
    pjpk_unit                VARCHAR(200),
    investment_value         NUMERIC(20,2) NOT NULL DEFAULT 0,
    currency                 VARCHAR(10)   NOT NULL DEFAULT 'IDR',
    concession_years         INT,
    concession_start         DATE,
    concession_end           DATE,
    financial_close_target   DATE,
    financial_close_actual   DATE,
    cod_target               DATE,
    cod_actual               DATE,
    org_unit_id              UUID          REFERENCES org_units(id),
    region_id                UUID,
    river_basin_id           UUID,
    province                 VARCHAR(100),
    latitude                 DOUBLE PRECISION,
    longitude                DOUBLE PRECISION,
    created_by               UUID,
    created_at               TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_kpbu_projects_org_code
    ON kpbu_projects (organization_id, code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_kpbu_projects_org_stage
    ON kpbu_projects (organization_id, stage) WHERE deleted_at IS NULL;

-- One target/actual row per lifecycle stage per KPBU project.
CREATE TABLE IF NOT EXISTS kpbu_stage_milestones (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID         NOT NULL REFERENCES organizations(id),
    kpbu_project_id  UUID         NOT NULL REFERENCES kpbu_projects(id),
    stage            VARCHAR(30)  NOT NULL,
    planned_start    DATE,
    planned_end      DATE,
    actual_start     DATE,
    actual_end       DATE,
    notes            TEXT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_kpbu_stage_milestones_project_stage
    ON kpbu_stage_milestones (kpbu_project_id, stage) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS kpbu_parties (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID          NOT NULL REFERENCES organizations(id),
    kpbu_project_id  UUID          NOT NULL REFERENCES kpbu_projects(id),
    role             VARCHAR(30)   NOT NULL,
    name             VARCHAR(300)  NOT NULL,
    vendor_id        UUID          REFERENCES vendors(id),
    share_pct        NUMERIC(5,2),
    contact_person   VARCHAR(200),
    email            VARCHAR(200),
    phone            VARCHAR(50),
    notes            TEXT,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_kpbu_parties_project
    ON kpbu_parties (kpbu_project_id) WHERE deleted_at IS NULL;
