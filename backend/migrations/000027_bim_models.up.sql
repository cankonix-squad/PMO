-- Migration 000027: bim_models + bim_model_versions + bim_project_mappings
-- P3-001 BIM/Digital Twin Integration Foundation
-- Stores external BIM model references, versions, discipline, and project mapping.
-- Does NOT store binary files in rows — only metadata and external URIs.

-- ---------------------------------------------------------------------------
-- bim_models — one record per BIM model registered in the system
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS bim_models (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID            NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                VARCHAR(255)    NOT NULL,
    description         TEXT            NOT NULL DEFAULT '',
    -- Discipline: ARCHITECTURAL | STRUCTURAL | MEP | CIVIL | LANDSCAPE | OTHER
    discipline          VARCHAR(50)     NOT NULL DEFAULT 'OTHER',
    -- Provider: autodesk_bim360 | trimble_connect | bentley_projectwise | local | other
    provider            VARCHAR(100)    NOT NULL DEFAULT 'other',
    -- External model identifier in the BIM provider system (no binary data stored here)
    external_model_id   VARCHAR(500)    NOT NULL DEFAULT '',
    -- Human-readable external reference URL (viewer link, not download)
    viewer_url          TEXT            NOT NULL DEFAULT '',
    -- Status: ACTIVE | ARCHIVED | DRAFT
    status              VARCHAR(20)     NOT NULL DEFAULT 'DRAFT',
    -- Metadata: file format, coordinate system, units, etc.
    metadata            JSONB           NOT NULL DEFAULT '{}',
    created_by          UUID            NOT NULL REFERENCES users(id),
    deleted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT now()
);

CREATE INDEX idx_bim_model_org        ON bim_models(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_bim_model_status     ON bim_models(organization_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_bim_model_discipline ON bim_models(organization_id, discipline) WHERE deleted_at IS NULL;
CREATE INDEX idx_bim_model_created_at ON bim_models(organization_id, created_at DESC) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- bim_model_versions — immutable version history per BIM model
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS bim_model_versions (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    bim_model_id        UUID            NOT NULL REFERENCES bim_models(id) ON DELETE CASCADE,
    organization_id     UUID            NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    version_label       VARCHAR(100)    NOT NULL,
    -- e.g. "v1.0", "Rev A", "2026-08-28"
    external_version_id VARCHAR(500)    NOT NULL DEFAULT '',
    -- Version ID in the BIM provider system
    change_summary      TEXT            NOT NULL DEFAULT '',
    -- What changed in this version
    file_size_bytes     BIGINT          NOT NULL DEFAULT 0,
    -- Size of source file (metadata only, file not stored here)
    checksum            VARCHAR(128)    NOT NULL DEFAULT '',
    -- SHA-256 or provider checksum for integrity verification
    published_at        TIMESTAMPTZ,
    -- When this version was published/approved; NULL = draft
    created_by          UUID            NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT now()
);

CREATE INDEX idx_bim_version_model    ON bim_model_versions(bim_model_id);
CREATE INDEX idx_bim_version_org      ON bim_model_versions(organization_id);
CREATE INDEX idx_bim_version_created  ON bim_model_versions(bim_model_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- bim_project_mappings — many-to-many: bim_model <-> project
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS bim_project_mappings (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID            NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    bim_model_id        UUID            NOT NULL REFERENCES bim_models(id) ON DELETE CASCADE,
    project_id          UUID            NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    -- Role of this model in the project: PRIMARY | REFERENCE | ASBUILT | OTHER
    model_role          VARCHAR(50)     NOT NULL DEFAULT 'REFERENCE',
    -- Notes about why this model is linked to this project
    notes               TEXT            NOT NULL DEFAULT '',
    linked_by           UUID            NOT NULL REFERENCES users(id),
    linked_at           TIMESTAMPTZ     NOT NULL DEFAULT now(),
    UNIQUE (bim_model_id, project_id)
);

CREATE INDEX idx_bim_mapping_org      ON bim_project_mappings(organization_id);
CREATE INDEX idx_bim_mapping_project  ON bim_project_mappings(project_id);
CREATE INDEX idx_bim_mapping_model    ON bim_project_mappings(bim_model_id);
