-- Rollback migration 000023
DROP TABLE IF EXISTS report_export_requests;
DROP TABLE IF EXISTS report_definitions;
DROP TYPE  IF EXISTS export_status;
DROP TYPE  IF EXISTS export_format_type;
