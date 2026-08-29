-- Down migration: drop project_periodic_reports
DROP INDEX IF EXISTS idx_periodic_report_project;
DROP INDEX IF EXISTS idx_periodic_report_org_period;
DROP INDEX IF EXISTS uq_periodic_report_active;
DROP TABLE IF EXISTS project_periodic_reports;
