-- 000015_monitoring_baseline_snapshot.up.sql
-- Baseline dan snapshot periodik untuk monitoring progress proyek (P1-011).
-- Menyimpan planned/actual physical, financial, schedule, period, source,
-- validation status, dan variance. Unique per project+period+type.

-- Baseline: rencana awal yang dibekukan (approved) sebagai acuan.
CREATE TABLE IF NOT EXISTS project_baselines (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID          NOT NULL REFERENCES organizations(id),
    project_id      UUID          NOT NULL REFERENCES projects(id),

    -- Identitas baseline
    version         SMALLINT      NOT NULL DEFAULT 1,
    label           VARCHAR(200),              -- e.g. "Baseline Rev-0"
    approved_at     TIMESTAMPTZ,
    approved_by     UUID          REFERENCES users(id),

    -- Fisik (physical progress %)
    physical_target DECIMAL(5,2)  NOT NULL DEFAULT 0 CHECK (physical_target BETWEEN 0 AND 100),

    -- Keuangan
    budget_total    DECIMAL(20,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(10)   NOT NULL DEFAULT 'IDR',

    -- Jadwal
    planned_start   DATE          NOT NULL,
    planned_end     DATE          NOT NULL,

    -- Meta
    source          VARCHAR(100),              -- e.g. "Manual", "Primavera P6"
    notes           TEXT,
    is_active       BOOLEAN       NOT NULL DEFAULT TRUE,  -- hanya 1 aktif per project

    created_by      UUID          REFERENCES users(id),
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    UNIQUE (project_id, version)
);

CREATE INDEX IF NOT EXISTS idx_baselines_project_id      ON project_baselines(project_id);
CREATE INDEX IF NOT EXISTS idx_baselines_organization_id ON project_baselines(organization_id);
CREATE INDEX IF NOT EXISTS idx_baselines_deleted_at      ON project_baselines(deleted_at);

-- Snapshot periodik: rekaman realisasi pada satu cut-off period.
-- Status DRAFT → SUBMITTED → VALID | REJECTED → STALE
CREATE TABLE IF NOT EXISTS project_snapshots (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID          NOT NULL REFERENCES organizations(id),
    project_id      UUID          NOT NULL REFERENCES projects(id),
    baseline_id     UUID          REFERENCES project_baselines(id),

    -- Periode pelaporan
    period_year     SMALLINT      NOT NULL,
    period_month    SMALLINT      NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    period_label    VARCHAR(50),               -- e.g. "Jul 2025" — denormalized for display

    -- Fisik
    physical_actual DECIMAL(5,2)  NOT NULL DEFAULT 0 CHECK (physical_actual BETWEEN 0 AND 100),
    physical_target DECIMAL(5,2)  NOT NULL DEFAULT 0 CHECK (physical_target BETWEEN 0 AND 100),
    physical_variance DECIMAL(6,2) GENERATED ALWAYS AS (physical_actual - physical_target) STORED,

    -- Keuangan
    financial_actual  DECIMAL(20,2) NOT NULL DEFAULT 0,
    financial_target  DECIMAL(20,2) NOT NULL DEFAULT 0,
    financial_variance DECIMAL(20,2) GENERATED ALWAYS AS (financial_actual - financial_target) STORED,
    currency          VARCHAR(10)   NOT NULL DEFAULT 'IDR',

    -- Jadwal
    schedule_actual_start DATE,
    schedule_actual_end   DATE,
    schedule_deviation_days INT,  -- positif = terlambat

    -- Validasi
    status          VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                    CHECK (status IN ('DRAFT','SUBMITTED','VALID','REJECTED','STALE')),
    submitted_at    TIMESTAMPTZ,
    submitted_by    UUID          REFERENCES users(id),
    validated_at    TIMESTAMPTZ,
    validated_by    UUID          REFERENCES users(id),
    rejection_reason TEXT,

    -- Sumber data
    source          VARCHAR(100),
    notes           TEXT,

    created_by      UUID          REFERENCES users(id),
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    -- Hanya satu snapshot aktif per project+period
    UNIQUE (project_id, period_year, period_month)
);

CREATE INDEX IF NOT EXISTS idx_snapshots_project_id      ON project_snapshots(project_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_organization_id ON project_snapshots(organization_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_period          ON project_snapshots(period_year, period_month);
CREATE INDEX IF NOT EXISTS idx_snapshots_status          ON project_snapshots(status);
CREATE INDEX IF NOT EXISTS idx_snapshots_deleted_at      ON project_snapshots(deleted_at);
