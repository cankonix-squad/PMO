-- Migration: 000013_report_snapshots (down)
DROP TABLE IF EXISTS report_snapshots;
DROP TYPE IF EXISTS report_status;
DROP TYPE IF EXISTS report_period;
