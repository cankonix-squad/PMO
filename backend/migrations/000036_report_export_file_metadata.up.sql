-- Migration 000036: add file metadata columns to report_export_requests
-- UAT-002 Report Export Real File
-- Safe and idempotent via IF NOT EXISTS / DO $$ checks

ALTER TABLE report_export_requests
    ADD COLUMN IF NOT EXISTS file_name       VARCHAR(255),
    ADD COLUMN IF NOT EXISTS storage_key     TEXT,
    ADD COLUMN IF NOT EXISTS mime_type       VARCHAR(100),
    ADD COLUMN IF NOT EXISTS file_size       BIGINT,
    ADD COLUMN IF NOT EXISTS generated_at   TIMESTAMPTZ;

-- Index for faster status polling
CREATE INDEX IF NOT EXISTS idx_report_export_requests_status
    ON report_export_requests(organization_id, status);

COMMENT ON COLUMN report_export_requests.file_name   IS 'Human-readable filename, e.g. executive-summary_2026-08-29.csv';
COMMENT ON COLUMN report_export_requests.storage_key IS 'Relative path under REPORT_STORAGE_PATH, safe from path traversal';
COMMENT ON COLUMN report_export_requests.mime_type   IS 'MIME type of generated file, e.g. text/csv';
COMMENT ON COLUMN report_export_requests.file_size   IS 'File size in bytes after generation';
COMMENT ON COLUMN report_export_requests.generated_at IS 'Timestamp when file was successfully written to storage';
