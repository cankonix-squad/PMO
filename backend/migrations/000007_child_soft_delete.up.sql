-- 000007_child_soft_delete.up.sql
-- Adds deleted_at to child tables that lack it so a project soft delete can
-- cascade soft deletes to every business child row.
-- Rationale: "Jangan hard delete business data" — relational join rows
-- (project_teams, task_assignments) and financial line items (project_budgets)
-- must be soft-deletable when their parent project is deleted.

ALTER TABLE project_teams ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_project_teams_deleted_at ON project_teams(deleted_at);

ALTER TABLE task_assignments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_task_assignments_deleted_at ON task_assignments(deleted_at);

ALTER TABLE project_budgets ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_project_budgets_deleted_at ON project_budgets(deleted_at);
