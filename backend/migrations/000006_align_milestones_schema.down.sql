-- 000006_align_milestones_schema.down.sql
-- Reverts milestone schema alignment.

DROP INDEX IF EXISTS idx_milestones_organization_id;

ALTER TABLE milestones
    DROP CONSTRAINT IF EXISTS fk_milestones_organization;

ALTER TABLE milestones
    DROP COLUMN IF EXISTS organization_id;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'milestones' AND column_name = 'title'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'milestones' AND column_name = 'name'
    ) THEN
        ALTER TABLE milestones RENAME COLUMN title TO name;
    ELSIF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'milestones' AND column_name = 'title'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'milestones' AND column_name = 'name'
    ) THEN
        UPDATE milestones SET name = COALESCE(name, title);
        ALTER TABLE milestones DROP COLUMN title;
    END IF;
END $$;
