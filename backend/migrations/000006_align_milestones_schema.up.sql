-- 000006_align_milestones_schema.up.sql
-- Align milestones table with the Go model and tenant rules.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'milestones' AND column_name = 'name'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'milestones' AND column_name = 'title'
    ) THEN
        ALTER TABLE milestones RENAME COLUMN name TO title;
    ELSIF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'milestones' AND column_name = 'title'
    ) THEN
        ALTER TABLE milestones ADD COLUMN title VARCHAR(500);
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'milestones' AND column_name = 'name'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'milestones' AND column_name = 'title'
    ) THEN
        UPDATE milestones SET title = COALESCE(title, name);
        ALTER TABLE milestones DROP COLUMN name;
    END IF;
END $$;

ALTER TABLE milestones
    ALTER COLUMN title SET NOT NULL;

ALTER TABLE milestones
    ADD COLUMN IF NOT EXISTS organization_id UUID;

UPDATE milestones
SET organization_id = projects.organization_id
FROM projects
WHERE milestones.project_id = projects.id
  AND milestones.organization_id IS NULL;

ALTER TABLE milestones
    ALTER COLUMN organization_id SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_milestones_organization'
    ) THEN
        ALTER TABLE milestones
            ADD CONSTRAINT fk_milestones_organization
            FOREIGN KEY (organization_id) REFERENCES organizations(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_milestones_organization_id ON milestones(organization_id);
