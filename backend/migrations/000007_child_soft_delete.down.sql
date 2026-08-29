-- 000007_child_soft_delete.down.sql
-- Reverts child soft-delete columns added by 000007.

DROP INDEX IF EXISTS idx_project_teams_deleted_at;
ALTER TABLE project_teams DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_task_assignments_deleted_at;
ALTER TABLE task_assignments DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_project_budgets_deleted_at;
ALTER TABLE project_budgets DROP COLUMN IF EXISTS deleted_at;
