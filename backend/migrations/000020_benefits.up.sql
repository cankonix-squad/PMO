CREATE TABLE IF NOT EXISTS benefit_indicators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), organization_id UUID NOT NULL REFERENCES organizations(id),
    project_id UUID REFERENCES projects(id), name VARCHAR(200) NOT NULL, unit VARCHAR(50) NOT NULL,
    aggregation_method VARCHAR(20) NOT NULL CHECK (aggregation_method IN ('SUM','AVERAGE','LATEST')),
    owner_id UUID REFERENCES users(id), source VARCHAR(100), description TEXT,
    created_by UUID REFERENCES users(id), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_benefit_indicators_scope ON benefit_indicators(organization_id, project_id, deleted_at);
CREATE TABLE IF NOT EXISTS benefit_measurements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), organization_id UUID NOT NULL REFERENCES organizations(id), indicator_id UUID NOT NULL REFERENCES benefit_indicators(id),
    period_year SMALLINT NOT NULL, period_month SMALLINT NOT NULL CHECK (period_month BETWEEN 1 AND 12), baseline NUMERIC(20,4) NOT NULL DEFAULT 0,
    target NUMERIC(20,4) NOT NULL DEFAULT 0, actual NUMERIC(20,4) NOT NULL DEFAULT 0, source VARCHAR(100), validation_status VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (validation_status IN ('DRAFT','SUBMITTED','VALID','REJECTED','STALE')),
    created_by UUID REFERENCES users(id), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), deleted_at TIMESTAMPTZ,
    UNIQUE(indicator_id, period_year, period_month)
);
CREATE INDEX IF NOT EXISTS idx_benefit_measurements_scope ON benefit_measurements(organization_id, indicator_id, validation_status);