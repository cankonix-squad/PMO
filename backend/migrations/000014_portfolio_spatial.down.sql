-- 000014_portfolio_spatial.down.sql

ALTER TABLE projects
    DROP COLUMN IF EXISTS river_basin_id,
    DROP COLUMN IF EXISTS region_id,
    DROP COLUMN IF EXISTS sector_id,
    DROP COLUMN IF EXISTS program_id;

DROP TABLE IF EXISTS river_basins;
DROP TABLE IF EXISTS regions;
DROP TABLE IF EXISTS sectors;
DROP TABLE IF EXISTS programs;
