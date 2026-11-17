DROP TABLE IF EXISTS kpbu_guarantees;
DROP TABLE IF EXISTS kpbu_documents;
DROP TABLE IF EXISTS kpbu_document_requirements;
ALTER TABLE kpbu_projects
    DROP COLUMN IF EXISTS agreement_signed_date,
    DROP COLUMN IF EXISTS procurement_reason,
    DROP COLUMN IF EXISTS procurement_method;
