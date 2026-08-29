-- 000014_portfolio_spatial.up.sql
-- Master data: programs, sectors, regions, river_basins (DAS)
-- All tenant-scoped with soft delete and unique code per organization.

-- Programs (master program / umbrella program)
CREATE TABLE IF NOT EXISTS programs (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES organizations(id),
    code            VARCHAR(100) NOT NULL,
    name            VARCHAR(300) NOT NULL,
    description     TEXT,
    fiscal_year     SMALLINT,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    sort_order      INT          NOT NULL DEFAULT 0,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE (organization_id, code)
);

CREATE INDEX IF NOT EXISTS idx_programs_organization_id ON programs(organization_id);
CREATE INDEX IF NOT EXISTS idx_programs_deleted_at      ON programs(deleted_at);

-- Sectors (sektor SDA — e.g. Irigasi, Bendungan, Sungai, Air Baku, dll.)
CREATE TABLE IF NOT EXISTS sectors (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES organizations(id),
    code            VARCHAR(100) NOT NULL,
    name            VARCHAR(300) NOT NULL,
    description     TEXT,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    sort_order      INT          NOT NULL DEFAULT 0,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE (organization_id, code)
);

CREATE INDEX IF NOT EXISTS idx_sectors_organization_id ON sectors(organization_id);
CREATE INDEX IF NOT EXISTS idx_sectors_deleted_at      ON sectors(deleted_at);

-- Regions (wilayah administratif — e.g. Provinsi, Kabupaten)
CREATE TABLE IF NOT EXISTS regions (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES organizations(id),
    parent_id       UUID         REFERENCES regions(id),
    code            VARCHAR(100) NOT NULL,
    name            VARCHAR(300) NOT NULL,
    level           SMALLINT     NOT NULL DEFAULT 1 CHECK (level BETWEEN 1 AND 4),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    sort_order      INT          NOT NULL DEFAULT 0,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE (organization_id, code)
);

CREATE INDEX IF NOT EXISTS idx_regions_organization_id ON regions(organization_id);
CREATE INDEX IF NOT EXISTS idx_regions_parent_id       ON regions(parent_id);
CREATE INDEX IF NOT EXISTS idx_regions_deleted_at      ON regions(deleted_at);

-- River basins (DAS — Daerah Aliran Sungai)
CREATE TABLE IF NOT EXISTS river_basins (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES organizations(id),
    region_id       UUID         REFERENCES regions(id),
    code            VARCHAR(100) NOT NULL,
    name            VARCHAR(300) NOT NULL,
    description     TEXT,
    area_km2        DECIMAL(12,2),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    sort_order      INT          NOT NULL DEFAULT 0,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE (organization_id, code)
);

CREATE INDEX IF NOT EXISTS idx_river_basins_organization_id ON river_basins(organization_id);
CREATE INDEX IF NOT EXISTS idx_river_basins_region_id       ON river_basins(region_id);
CREATE INDEX IF NOT EXISTS idx_river_basins_deleted_at      ON river_basins(deleted_at);

-- Add classification FK columns to projects
ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS program_id     UUID REFERENCES programs(id),
    ADD COLUMN IF NOT EXISTS sector_id      UUID REFERENCES sectors(id),
    ADD COLUMN IF NOT EXISTS region_id      UUID REFERENCES regions(id),
    ADD COLUMN IF NOT EXISTS river_basin_id UUID REFERENCES river_basins(id);

CREATE INDEX IF NOT EXISTS idx_projects_program_id     ON projects(program_id);
CREATE INDEX IF NOT EXISTS idx_projects_sector_id      ON projects(sector_id);
CREATE INDEX IF NOT EXISTS idx_projects_region_id      ON projects(region_id);
CREATE INDEX IF NOT EXISTS idx_projects_river_basin_id ON projects(river_basin_id);
