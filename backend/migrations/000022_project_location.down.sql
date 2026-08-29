-- 000022_project_location.down.sql

DROP INDEX IF EXISTS idx_projects_location;

ALTER TABLE projects
    DROP COLUMN IF EXISTS latitude,
    DROP COLUMN IF EXISTS longitude,
    DROP COLUMN IF EXISTS province,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS location_name;
