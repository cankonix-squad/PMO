-- Migration 000024: import_jobs + import_rows
-- P2-001 CSV/Excel Import — import job lifecycle and row-level results

-- ---------------------------------------------------------------------------
-- import_jobs — one record per upload/import attempt
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS import_jobs (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID         NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    dataset_type     VARCHAR(50)  NOT NULL,   -- project_progress | project_budgets | risks | issues | benefit_measurements
    file_name        VARCHAR(500) NOT NULL,
    file_size        BIGINT       NOT NULL DEFAULT 0,
    mime_type        VARCHAR(100) NOT NULL DEFAULT 'text/csv',
    status           VARCHAR(20)  NOT NULL DEFAULT 'UPLOADED',
                                              -- UPLOADED | VALIDATED | COMMITTED | FAILED | CANCELLED
    total_rows       INT          NOT NULL DEFAULT 0,
    valid_rows       INT          NOT NULL DEFAULT 0,
    invalid_rows     INT          NOT NULL DEFAULT 0,
    error_summary    JSONB        NOT NULL DEFAULT '[]',
    uploaded_by      UUID         NOT NULL REFERENCES users(id),
    validated_at     TIMESTAMPTZ,
    committed_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_import_jobs_org        ON import_jobs(organization_id);
CREATE INDEX idx_import_jobs_status     ON import_jobs(organization_id, status);
CREATE INDEX idx_import_jobs_type       ON import_jobs(organization_id, dataset_type);
CREATE INDEX idx_import_jobs_uploaded   ON import_jobs(uploaded_by);

-- ---------------------------------------------------------------------------
-- import_rows — row-level parse/validation result per job
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS import_rows (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id              UUID        NOT NULL REFERENCES import_jobs(id) ON DELETE CASCADE,
    row_number          INT         NOT NULL,
    raw_payload         JSONB       NOT NULL DEFAULT '{}',
    normalized_payload  JSONB       NOT NULL DEFAULT '{}',
    valid               BOOLEAN     NOT NULL DEFAULT false,
    errors              JSONB       NOT NULL DEFAULT '[]',
    action              VARCHAR(20) NOT NULL DEFAULT 'SKIP',  -- CREATE | UPDATE | SKIP
    target_entity_id    UUID,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_import_rows_job        ON import_rows(job_id);
CREATE INDEX idx_import_rows_valid      ON import_rows(job_id, valid);
CREATE INDEX idx_import_rows_row_number ON import_rows(job_id, row_number);
