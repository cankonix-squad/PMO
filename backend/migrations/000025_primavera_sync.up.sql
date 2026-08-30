-- Migration 000025: primavera_sync_runs + primavera_activity_mappings
-- P2-011 Primavera P6 Adapter — idempotent sync, lineage, conflict/error report

-- ---------------------------------------------------------------------------
-- primavera_sync_runs — one record per sync attempt (push / pull)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS primavera_sync_runs (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID            NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id          UUID            REFERENCES projects(id) ON DELETE SET NULL,
    -- null = org-wide batch; non-null = single-project sync
    direction           VARCHAR(10)     NOT NULL DEFAULT 'IMPORT',
                                        -- IMPORT | EXPORT (only IMPORT in P2-011 scope)
    source_file_name    VARCHAR(500)    NOT NULL DEFAULT '',
    source_file_size    BIGINT          NOT NULL DEFAULT 0,
    source_mime_type    VARCHAR(100)    NOT NULL DEFAULT 'text/xml',
    -- XER (text) or PMXML (xml) from Primavera P6
    format              VARCHAR(20)     NOT NULL DEFAULT 'XER',
                                        -- XER | PMXML | JSON
    status              VARCHAR(20)     NOT NULL DEFAULT 'PENDING',
                                        -- PENDING | RUNNING | DONE | FAILED | CANCELLED
    total_activities    INT             NOT NULL DEFAULT 0,
    imported_activities INT             NOT NULL DEFAULT 0,
    skipped_activities  INT             NOT NULL DEFAULT 0,
    failed_activities   INT             NOT NULL DEFAULT 0,
    conflict_count      INT             NOT NULL DEFAULT 0,
    error_summary       JSONB           NOT NULL DEFAULT '[]',
    -- [{"code":"E001","message":"...","activity_id":"A1010","row":12}, ...]
    conflict_report     JSONB           NOT NULL DEFAULT '[]',
    -- [{"activity_id":"A1010","field":"planned_end","existing":"2024-12-01","incoming":"2025-01-01"}, ...]
    lineage             JSONB           NOT NULL DEFAULT '{}',
    -- {"source_project_id":"P001","exported_at":"2024-01-01","p6_version":"22.12","operator":"admin"}
    started_at          TIMESTAMPTZ,
    finished_at         TIMESTAMPTZ,
    triggered_by        UUID            NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT now()
);

CREATE INDEX idx_p6sync_org            ON primavera_sync_runs(organization_id);
CREATE INDEX idx_p6sync_project        ON primavera_sync_runs(project_id);
CREATE INDEX idx_p6sync_status         ON primavera_sync_runs(organization_id, status);
CREATE INDEX idx_p6sync_triggered_by   ON primavera_sync_runs(triggered_by);
CREATE INDEX idx_p6sync_created_at     ON primavera_sync_runs(organization_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- primavera_activity_mappings — links a P6 activity_id → CANKORA entity
-- Provides idempotency: re-import of the same activity_id updates rather
-- than duplicates the linked entity.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS primavera_activity_mappings (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID            NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id          UUID            NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sync_run_id         UUID            NOT NULL REFERENCES primavera_sync_runs(id) ON DELETE CASCADE,
    -- P6 identifiers
    p6_activity_id      VARCHAR(100)    NOT NULL,   -- e.g. "A1010"
    p6_wbs_code         VARCHAR(200)    NOT NULL DEFAULT '',
    p6_activity_name    VARCHAR(500)    NOT NULL DEFAULT '',
    -- Mapped CANKORA entity
    entity_type         VARCHAR(50)     NOT NULL,   -- task | milestone | baseline | snapshot
    entity_id           UUID            NOT NULL,
    -- Sync metadata
    action              VARCHAR(20)     NOT NULL DEFAULT 'CREATE',
                                        -- CREATE | UPDATE | SKIP | CONFLICT
    baseline_physical   NUMERIC(5,2)    NOT NULL DEFAULT 0,
    actual_physical     NUMERIC(5,2)    NOT NULL DEFAULT 0,
    planned_start       DATE,
    planned_end         DATE,
    actual_start        DATE,
    actual_end          DATE,
    raw_payload         JSONB           NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT now(),
    -- Uniqueness: one mapping per org+project+p6_activity per entity_type
    CONSTRAINT uq_p6_activity_mapping UNIQUE (organization_id, project_id, p6_activity_id, entity_type)
);

CREATE INDEX idx_p6map_org             ON primavera_activity_mappings(organization_id);
CREATE INDEX idx_p6map_project         ON primavera_activity_mappings(organization_id, project_id);
CREATE INDEX idx_p6map_run             ON primavera_activity_mappings(sync_run_id);
CREATE INDEX idx_p6map_entity          ON primavera_activity_mappings(entity_type, entity_id);
CREATE INDEX idx_p6map_activity_id     ON primavera_activity_mappings(organization_id, p6_activity_id);
