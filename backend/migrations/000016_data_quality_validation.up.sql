CREATE TABLE IF NOT EXISTS data_submissions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID NOT NULL REFERENCES organizations(id),
    project_id        UUID NOT NULL REFERENCES projects(id),
    snapshot_id       UUID NOT NULL UNIQUE REFERENCES project_snapshots(id),
    source            VARCHAR(100),
    source_reference  VARCHAR(255),
    period_year       SMALLINT NOT NULL,
    period_month      SMALLINT NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    status            VARCHAR(20) NOT NULL DEFAULT 'SUBMITTED'
                      CHECK (status IN ('DRAFT','SUBMITTED','VALID','REJECTED','STALE')),
    completeness_pct  DECIMAL(5,2) NOT NULL DEFAULT 0 CHECK (completeness_pct BETWEEN 0 AND 100),
    freshness_at      TIMESTAMPTZ,
    freshness_days    INT,
    sla_due_at        TIMESTAMPTZ,
    submitted_by      UUID REFERENCES users(id),
    submitted_at      TIMESTAMPTZ,
    validator_id      UUID REFERENCES users(id),
    validated_at      TIMESTAMPTZ,
    rejection_reason  TEXT,
    lineage           JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_data_submissions_queue
    ON data_submissions(organization_id, status, submitted_at);
CREATE INDEX IF NOT EXISTS idx_data_submissions_project
    ON data_submissions(organization_id, project_id, period_year, period_month);
CREATE INDEX IF NOT EXISTS idx_data_submissions_sla
    ON data_submissions(organization_id, sla_due_at);