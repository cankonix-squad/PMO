-- Migration: 000039_create_project_categories
-- Tabel master kategori proyek (multi-tenant, soft-delete)

CREATE TABLE IF NOT EXISTS project_categories (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID        NOT NULL REFERENCES organizations(id),
    code            VARCHAR(100) NOT NULL,
    name            VARCHAR(300) NOT NULL,
    description     TEXT,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    sort_order      INT         NOT NULL DEFAULT 0,
    created_by      UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Unique: code per org, only on non-deleted rows
CREATE UNIQUE INDEX IF NOT EXISTS uq_project_categories_org_code
    ON project_categories (organization_id, code)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_project_categories_org_active
    ON project_categories (organization_id, is_active)
    WHERE deleted_at IS NULL;
