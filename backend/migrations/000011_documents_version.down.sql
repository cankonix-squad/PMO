-- 000011_documents_version.down.sql
-- Reverts document version + updated_at columns (P1-005).

ALTER TABLE project_documents
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE project_documents
    DROP COLUMN IF EXISTS version;
