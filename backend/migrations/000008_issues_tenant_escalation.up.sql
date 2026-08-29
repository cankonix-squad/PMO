-- 000008_issues_tenant_escalation.up.sql
-- Adds tenant guard (organization_id) and escalation level to issues.

ALTER TABLE issues
    ADD COLUMN IF NOT EXISTS organization_id UUID;

UPDATE issues i
SET organization_id = p.organization_id
FROM projects p
WHERE i.project_id = p.id
  AND i.organization_id IS NULL;

ALTER TABLE issues
    ALTER COLUMN organization_id SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_issues_organization'
    ) THEN
        ALTER TABLE issues
            ADD CONSTRAINT fk_issues_organization
            FOREIGN KEY (organization_id) REFERENCES organizations(id);
    END IF;
END $$;

ALTER TABLE issues
    ADD COLUMN IF NOT EXISTS escalation VARCHAR(50) NOT NULL DEFAULT 'NONE';

CREATE INDEX IF NOT EXISTS idx_issues_organization_id ON issues(organization_id);
CREATE INDEX IF NOT EXISTS idx_issues_deleted_at ON issues(deleted_at);
