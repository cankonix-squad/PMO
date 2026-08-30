-- Migration 000025 rollback: drop primavera_sync tables
DROP TABLE IF EXISTS primavera_activity_mappings;
DROP TABLE IF EXISTS primavera_sync_runs;
