DROP TABLE IF EXISTS vendor_documents;
DROP TABLE IF EXISTS vendor_progress_reports;
DROP INDEX IF EXISTS idx_users_vendor_id;
ALTER TABLE users DROP COLUMN IF EXISTS vendor_id;
