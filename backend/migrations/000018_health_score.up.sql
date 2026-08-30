CREATE TABLE IF NOT EXISTS health_formulas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    version INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT','APPROVED','RETIRED')),
    weights JSONB NOT NULL,
    thresholds JSONB NOT NULL,
    missing_data_rule VARCHAR(30) NOT NULL DEFAULT 'PENALIZE',
    effective_from TIMESTAMPTZ,
    effective_to TIMESTAMPTZ,
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, version)
);
CREATE INDEX IF NOT EXISTS idx_health_formulas_active ON health_formulas(organization_id, status);

CREATE TABLE IF NOT EXISTS health_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    project_id UUID NOT NULL REFERENCES projects(id),
    formula_id UUID NOT NULL REFERENCES health_formulas(id),
    period_year SMALLINT NOT NULL,
    period_month SMALLINT NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    score DECIMAL(5,2) NOT NULL,
    health_class VARCHAR(20) NOT NULL CHECK (health_class IN ('GREEN','YELLOW','RED','CRITICAL')),
    components JSONB NOT NULL,
    explanation TEXT NOT NULL,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, project_id, formula_id, period_year, period_month)
);
CREATE INDEX IF NOT EXISTS idx_health_snapshots_project ON health_snapshots(organization_id, project_id, period_year, period_month);