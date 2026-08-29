-- Migration 000026 rollback: drop government connector tables
DROP TABLE IF EXISTS government_external_mappings;
DROP TABLE IF EXISTS government_sync_records;
DROP TABLE IF EXISTS government_sync_runs;
