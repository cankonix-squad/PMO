-- Migration: 000042_kpbu_documents (KPBU-K1.5)
-- Business-entity document checklist, uploaded documents with verification,
-- guarantees, procurement method and agreement signing date.

ALTER TABLE kpbu_projects
    ADD COLUMN IF NOT EXISTS procurement_method    VARCHAR(30),
    ADD COLUMN IF NOT EXISTS procurement_reason    TEXT,
    ADD COLUMN IF NOT EXISTS agreement_signed_date DATE;

-- Master checklist, editable per organization. Default rows are seeded from
-- the "Matriks Dokumen Badan Usaha" (see internal/kpbu/requirements_default.go).
CREATE TABLE IF NOT EXISTS kpbu_document_requirements (
    id                     UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id        UUID          NOT NULL REFERENCES organizations(id),
    code                   VARCHAR(30)   NOT NULL,
    category               VARCHAR(30)   NOT NULL,
    name                   VARCHAR(300)  NOT NULL,
    description            TEXT,
    applies_to             VARCHAR(20)   NOT NULL DEFAULT 'ALL',
    required_before_stage  VARCHAR(30),
    is_mandatory           BOOLEAN       NOT NULL DEFAULT TRUE,
    legal_basis            VARCHAR(500),
    source_status          VARCHAR(20)   NOT NULL DEFAULT 'UNVERIFIED',
    sort_order             INT           NOT NULL DEFAULT 0,
    is_active              BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at             TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at             TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_kpbu_doc_req_org_code
    ON kpbu_document_requirements (organization_id, code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS kpbu_documents (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID          NOT NULL REFERENCES organizations(id),
    kpbu_project_id  UUID          NOT NULL REFERENCES kpbu_projects(id),
    requirement_id   UUID          REFERENCES kpbu_document_requirements(id),
    name             VARCHAR(500)  NOT NULL,
    file_url         VARCHAR(1000) NOT NULL,
    mime_type        VARCHAR(150),
    file_size        BIGINT        NOT NULL DEFAULT 0,
    document_number  VARCHAR(200),
    document_date    DATE,
    notes            TEXT,
    status           VARCHAR(20)   NOT NULL DEFAULT 'SUBMITTED',
    uploaded_by      UUID          NOT NULL,
    uploaded_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    reviewed_by      UUID,
    reviewed_at      TIMESTAMPTZ,
    review_notes     TEXT,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_kpbu_documents_project
    ON kpbu_documents (kpbu_project_id, requirement_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS kpbu_guarantees (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID          NOT NULL REFERENCES organizations(id),
    kpbu_project_id  UUID          NOT NULL REFERENCES kpbu_projects(id),
    type             VARCHAR(30)   NOT NULL,
    issuer           VARCHAR(300),
    guarantee_number VARCHAR(200),
    amount           NUMERIC(20,2) NOT NULL DEFAULT 0,
    currency         VARCHAR(10)   NOT NULL DEFAULT 'IDR',
    issued_date      DATE,
    expiry_date      DATE,
    status           VARCHAR(20)   NOT NULL DEFAULT 'ACTIVE',
    notes            TEXT,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_kpbu_guarantees_project
    ON kpbu_guarantees (kpbu_project_id) WHERE deleted_at IS NULL;
