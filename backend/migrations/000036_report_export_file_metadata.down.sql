-- Rollback migration 000036
ALTER TABLE report_export_requests
    DROP COLUMN IF EXISTS file_name,
    DROP COLUMN IF EXISTS storage_key,
    DROP COLUMN IF EXISTS mime_type,
    DROP COLUMN IF EXISTS file_size,
    DROP COLUMN IF EXISTS generated_at;

DROP INDEX IF EXISTS idx_report_export_requests_status;
