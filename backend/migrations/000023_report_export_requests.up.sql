-- Migration 000023: report_export_requests + report_definitions
-- P2-009 Reporting Integration — export request queue and report catalog registry

-- ---------------------------------------------------------------------------
-- report_definitions — catalog of available report definitions
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS report_definitions (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name             VARCHAR(200) NOT NULL,
    description      TEXT,
    category         VARCHAR(50) NOT NULL DEFAULT 'GENERAL',   -- EXECUTIVE / PORTFOLIO / PROJECT / RISK / BUDGET / BENEFIT / PRIORITY
    dataset_key      VARCHAR(100) NOT NULL,                    -- executive-summary / project-performance / risk-issue / budget / benefits / priority
    visualization_type VARCHAR(50) DEFAULT 'TABLE',            -- TABLE / BAR_CHART / LINE_CHART / PIE_CHART / MAP / KPI_CARD
    available        BOOLEAN     NOT NULL DEFAULT true,
    requires_powerbi BOOLEAN     NOT NULL DEFAULT false,
    embed_configured BOOLEAN     NOT NULL DEFAULT false,       -- true jika Power BI embed URL sudah dikonfigurasi
    sort_order       INT         NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_report_definitions_org      ON report_definitions(organization_id);
CREATE INDEX idx_report_definitions_category ON report_definitions(organization_id, category);
CREATE INDEX idx_report_definitions_key      ON report_definitions(organization_id, dataset_key);

-- ---------------------------------------------------------------------------
-- export_status enum
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'export_status') THEN
        CREATE TYPE export_status AS ENUM ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED');
    END IF;
END$$;

-- ---------------------------------------------------------------------------
-- export_format enum
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'export_format_type') THEN
        CREATE TYPE export_format_type AS ENUM ('PDF', 'XLSX', 'CSV');
    END IF;
END$$;

-- ---------------------------------------------------------------------------
-- report_export_requests — async export job queue
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS report_export_requests (
    id               UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID              NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    report_id        UUID              REFERENCES report_definitions(id) ON DELETE SET NULL,
    dataset_key      VARCHAR(100)      NOT NULL,
    format           export_format_type NOT NULL DEFAULT 'XLSX',
    status           export_status     NOT NULL DEFAULT 'PENDING',
    parameters       JSONB             NOT NULL DEFAULT '{}',  -- filter params snapshot
    file_url         TEXT,                                      -- filled when COMPLETED
    error_message    TEXT,                                      -- filled when FAILED
    requested_by     UUID              NOT NULL REFERENCES users(id),
    started_at       TIMESTAMPTZ,
    completed_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ       NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ       NOT NULL DEFAULT now()
);

CREATE INDEX idx_report_export_org     ON report_export_requests(organization_id);
CREATE INDEX idx_report_export_status  ON report_export_requests(organization_id, status);
CREATE INDEX idx_report_export_user    ON report_export_requests(requested_by);
CREATE INDEX idx_report_export_created ON report_export_requests(organization_id, created_at DESC);
