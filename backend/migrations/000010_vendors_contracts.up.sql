-- 000010_vendors_contracts.up.sql
-- Adds vendor/party master table and project contract table (P1-004).
--
-- vendors: tenant-scoped master of penyedia (VENDOR) and konsultan
--          supervisi (CONSULTANT).
-- contracts: project contracts linked to a project + organization. The
--            contract_number is unique per organization (partial unique
--            index over non-deleted rows so a number can be reused after
--            a soft delete).

-- ---------------------------------------------------------------------------
-- vendors
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vendors (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    name           VARCHAR(500) NOT NULL,
    type           VARCHAR(50)  NOT NULL DEFAULT 'VENDOR', -- VENDOR | CONSULTANT
    legal_name     VARCHAR(500),
    tax_id         VARCHAR(100),
    contact_person VARCHAR(200),
    email          VARCHAR(200),
    phone          VARCHAR(50),
    address        TEXT,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_by     UUID NOT NULL REFERENCES users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_vendors_organization_id ON vendors(organization_id);
CREATE INDEX IF NOT EXISTS idx_vendors_type ON vendors(type);
CREATE INDEX IF NOT EXISTS idx_vendors_deleted_at ON vendors(deleted_at);

-- ---------------------------------------------------------------------------
-- contracts
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS contracts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    project_id      UUID NOT NULL REFERENCES projects(id),
    contract_number VARCHAR(200) NOT NULL,
    title           VARCHAR(500) NOT NULL,
    vendor_id       UUID NOT NULL REFERENCES vendors(id),
    consultant_id   UUID REFERENCES vendors(id),
    contract_value  NUMERIC(20,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(10) NOT NULL DEFAULT 'IDR',
    signed_date     TIMESTAMPTZ,
    start_date      TIMESTAMPTZ,
    end_date        TIMESTAMPTZ,
    status          VARCHAR(50) NOT NULL DEFAULT 'DRAFT', -- DRAFT | ACTIVE | AMENDED | COMPLETED | TERMINATED
    scope_of_work   TEXT,
    created_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_contracts_organization_id ON contracts(organization_id);
CREATE INDEX IF NOT EXISTS idx_contracts_project_id ON contracts(project_id);
CREATE INDEX IF NOT EXISTS idx_contracts_vendor_id ON contracts(vendor_id);
CREATE INDEX IF NOT EXISTS idx_contracts_status ON contracts(status);
CREATE INDEX IF NOT EXISTS idx_contracts_deleted_at ON contracts(deleted_at);

-- contract_number must be unique per organization among non-deleted rows.
CREATE UNIQUE INDEX IF NOT EXISTS uq_contracts_org_number
    ON contracts(organization_id, contract_number)
    WHERE deleted_at IS NULL;
