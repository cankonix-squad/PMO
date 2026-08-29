-- 000008_issues_tenant_escalation.down.sql
-- Reverts issue tenant guard and escalation column.

DROP INDEX IF EXISTS idx_issues_organization_id;
DROP INDEX IF EXISTS idx_issues_deleted_at;

ALTER TABLE issues
    DROP CONSTRAINT IF EXISTS fk_issues_organization;

ALTER TABLE issues
    DROP COLUMN IF EXISTS escalation;

ALTER TABLE issues
    DROP COLUMN IF EXISTS organization_id;
