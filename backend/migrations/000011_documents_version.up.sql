-- 000011_documents_version.up.sql
-- Adds document versioning + updated_at to project_documents (P1-005).
-- The document table already exists (000005) with metadata columns; this
-- migration extends it with version + updated_at so clients can track
-- revisions of operational evidence before validation (P1-012).

ALTER TABLE project_documents
    ADD COLUMN IF NOT EXISTS version VARCHAR(100);

ALTER TABLE project_documents
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
