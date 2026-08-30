-- Migration 000024 down: drop import_jobs + import_rows
DROP TABLE IF EXISTS import_rows;
DROP TABLE IF EXISTS import_jobs;
