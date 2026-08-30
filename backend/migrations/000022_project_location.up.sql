-- 000022_project_location.up.sql
-- Tambah kolom lokasi pada tabel projects untuk fitur GIS Map (P2-008)

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS latitude      DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS longitude     DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS province      VARCHAR(100),
    ADD COLUMN IF NOT EXISTS city          VARCHAR(150),
    ADD COLUMN IF NOT EXISTS location_name VARCHAR(200);

CREATE INDEX IF NOT EXISTS idx_projects_location
    ON projects (latitude, longitude)
    WHERE deleted_at IS NULL;
