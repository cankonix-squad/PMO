-- 000001_create_organizations.up.sql
-- Creates the organizations and org_units tables (tenant foundation)

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS organizations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(50)  NOT NULL UNIQUE,
    name        VARCHAR(300) NOT NULL,
    short_name  VARCHAR(100),
    logo_url    TEXT,
    address     TEXT,
    website     VARCHAR(300),
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at ON organizations(deleted_at);

CREATE TABLE IF NOT EXISTS org_units (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES organizations(id),
    parent_id       UUID         REFERENCES org_units(id),
    code            VARCHAR(100) NOT NULL,
    name            VARCHAR(300) NOT NULL,
    level           SMALLINT     NOT NULL CHECK (level BETWEEN 1 AND 5),
    head_user_id    UUID,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    sort_order      INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_org_units_organization_id ON org_units(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_units_parent_id       ON org_units(parent_id);
CREATE INDEX IF NOT EXISTS idx_org_units_deleted_at      ON org_units(deleted_at);
