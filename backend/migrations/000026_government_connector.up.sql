-- Migration 000026: government_sync_runs + government_sync_records + government_external_mappings
-- P2-002 Government Connector Foundation — SIRUP / SIMPONI / OM-SPAN / mock connectors

-- ---------------------------------------------------------------------------
-- government_sync_runs — one record per government data ingestion attempt
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS government_sync_runs (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID            NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    connector_key       VARCHAR(100)    NOT NULL,
                                        -- government_project_registry | government_budget_reference |
                                        -- government_location_reference | government_vendor_reference
    dataset_type        VARCHAR(100)    NOT NULL,
                                        -- e.g. projects, budget_allocations, locations, vendors
    mode                VARCHAR(20)     NOT NULL DEFAULT 'SAMPLE',
                                        -- SAMPLE | DRY_RUN | COMMIT
    status              VARCHAR(20)     NOT NULL DEFAULT 'PENDING',
                                        -- PENDING | RUNNING | SUCCEEDED | FAILED | CANCELLED
    started_by          UUID            NOT NULL REFERENCES users(id),
    total_records       INT             NOT NULL DEFAULT 0,
    accepted_records    INT             NOT NULL DEFAULT 0,
    rejected_records    INT             NOT NULL DEFAULT 0,
    error_summary       JSONB           NOT NULL DEFAULT '[]',
    -- [{"code":"E001","message":"...","external_id":"abc","row":5}, ...]
    source_hash         VARCHAR(128)    NOT NULL DEFAULT '',
    -- SHA-256 of request payload / source URL for idempotency
    idempotency_key     VARCHAR(256)    NOT NULL DEFAULT '',
    -- caller-supplied idempotency key (connector_key + dataset_type + period)
    started_at          TIMESTAMPTZ,
    finished_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT now()
);

CREATE INDEX idx_gov_run_org           ON government_sync_runs(organization_id);
CREATE INDEX idx_gov_run_connector     ON government_sync_runs(organization_id, connector_key);
CREATE INDEX idx_gov_run_status        ON government_sync_runs(organization_id, status);
CREATE INDEX idx_gov_run_idempotency   ON government_sync_runs(organization_id, idempotency_key);
CREATE INDEX idx_gov_run_created_at    ON government_sync_runs(organization_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- government_sync_records — record-level ingestion log per run
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS government_sync_records (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    sync_run_id         UUID            NOT NULL REFERENCES government_sync_runs(id) ON DELETE CASCADE,
    organization_id     UUID            NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    external_id         VARCHAR(500)    NOT NULL DEFAULT '',
    -- identifier in the government source system
    dataset_type        VARCHAR(100)    NOT NULL,
    status              VARCHAR(20)     NOT NULL DEFAULT 'ACCEPTED',
                                        -- ACCEPTED | REJECTED | SKIPPED
    action              VARCHAR(20)     NOT NULL DEFAULT 'CREATE',
                                        -- CREATE | UPDATE | SKIP | CONFLICT
    validation_errors   JSONB           NOT NULL DEFAULT '[]',
    -- [{"field":"name","message":"required"}, ...]
    raw_payload         JSONB           NOT NULL DEFAULT '{}',
    -- sanitised incoming record (no secrets)
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT now()
);

CREATE INDEX idx_gov_record_run        ON government_sync_records(sync_run_id);
CREATE INDEX idx_gov_record_org        ON government_sync_records(organization_id);
CREATE INDEX idx_gov_record_ext_id     ON government_sync_records(organization_id, dataset_type, external_id);

-- ---------------------------------------------------------------------------
-- government_external_mappings — external_id → internal CANKORA entity lineage
-- Provides idempotency: re-ingest of the same external_id updates the linked entity.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS government_external_mappings (
    id                      UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID            NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    connector_key           VARCHAR(100)    NOT NULL,
    dataset_type            VARCHAR(100)    NOT NULL,
    external_id             VARCHAR(500)    NOT NULL,
    internal_entity_type    VARCHAR(100)    NOT NULL,
    -- project | budget | location | vendor | program | org_unit
    internal_entity_id      UUID            NOT NULL,
    source_payload_hash     VARCHAR(128)    NOT NULL DEFAULT '',
    -- SHA-256 of last ingested raw_payload for change detection
    last_seen_at            TIMESTAMPTZ     NOT NULL DEFAULT now(),
    sync_run_id             UUID            REFERENCES government_sync_runs(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT now(),
    UNIQUE (organization_id, connector_key, dataset_type, external_id)
);

CREATE INDEX idx_gov_map_org           ON government_external_mappings(organization_id);
CREATE INDEX idx_gov_map_connector     ON government_external_mappings(organization_id, connector_key);
CREATE INDEX idx_gov_map_entity        ON government_external_mappings(organization_id, internal_entity_type, internal_entity_id);
CREATE INDEX idx_gov_map_ext_id        ON government_external_mappings(organization_id, dataset_type, external_id);
