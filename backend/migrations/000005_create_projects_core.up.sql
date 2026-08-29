-- 000005_create_projects_core.up.sql
-- Creates all project-related tables

CREATE TABLE IF NOT EXISTS projects (
    id              UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID             NOT NULL REFERENCES organizations(id),
    org_unit_id     UUID             REFERENCES org_units(id),
    code            VARCHAR(100)     NOT NULL,
    name            VARCHAR(500)     NOT NULL,
    description     TEXT,
    objectives      TEXT,
    status          VARCHAR(50)      NOT NULL DEFAULT 'DRAFT',
    priority        VARCHAR(50)      NOT NULL DEFAULT 'MEDIUM',
    category        VARCHAR(100),
    start_date      DATE,
    end_date        DATE,
    actual_end_date DATE,
    budget_total    NUMERIC(20,2)    NOT NULL DEFAULT 0,
    currency        VARCHAR(10)      NOT NULL DEFAULT 'IDR',
    progress_pct    NUMERIC(5,2)     NOT NULL DEFAULT 0,
    manager_id      UUID             REFERENCES users(id),
    created_by      UUID             NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE (organization_id, code)
);

CREATE INDEX IF NOT EXISTS idx_projects_organization_id ON projects(organization_id);
CREATE INDEX IF NOT EXISTS idx_projects_status          ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_deleted_at      ON projects(deleted_at);

CREATE TABLE IF NOT EXISTS project_teams (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    UUID         NOT NULL REFERENCES users(id),
    role       VARCHAR(100) NOT NULL DEFAULT 'MEMBER',
    joined_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_project_teams_project_id ON project_teams(project_id);

CREATE TABLE IF NOT EXISTS milestones (
    id           UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID          NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name         VARCHAR(500)  NOT NULL,
    description  TEXT,
    due_date     DATE,
    status       VARCHAR(50)   NOT NULL DEFAULT 'PENDING',
    progress_pct NUMERIC(5,2)  NOT NULL DEFAULT 0,
    created_by   UUID          NOT NULL REFERENCES users(id),
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_milestones_project_id  ON milestones(project_id);
CREATE INDEX IF NOT EXISTS idx_milestones_deleted_at  ON milestones(deleted_at);

CREATE TABLE IF NOT EXISTS tasks (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID          NOT NULL REFERENCES organizations(id),
    project_id      UUID          NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    milestone_id    UUID          REFERENCES milestones(id),
    parent_id       UUID          REFERENCES tasks(id),
    wbs_code        VARCHAR(50),
    title           VARCHAR(500)  NOT NULL,
    description     TEXT,
    status          VARCHAR(50)   NOT NULL DEFAULT 'TODO',
    priority        VARCHAR(50)   NOT NULL DEFAULT 'MEDIUM',
    type            VARCHAR(50)   NOT NULL DEFAULT 'TASK',
    start_date      DATE,
    due_date        DATE,
    est_hours       NUMERIC(8,2)  NOT NULL DEFAULT 0,
    actual_hours    NUMERIC(8,2)  NOT NULL DEFAULT 0,
    progress_pct    NUMERIC(5,2)  NOT NULL DEFAULT 0,
    created_by      UUID          NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_tasks_project_id      ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_milestone_id    ON tasks(milestone_id);
CREATE INDEX IF NOT EXISTS idx_tasks_parent_id       ON tasks(parent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status          ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at      ON tasks(deleted_at);

CREATE TABLE IF NOT EXISTS task_assignments (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id     UUID        NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users(id),
    is_lead     BOOLEAN     NOT NULL DEFAULT FALSE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (task_id, user_id)
);

CREATE TABLE IF NOT EXISTS issues (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_id      UUID         REFERENCES tasks(id),
    title        VARCHAR(500) NOT NULL,
    description  TEXT,
    status       VARCHAR(50)  NOT NULL DEFAULT 'OPEN',
    severity     VARCHAR(50)  NOT NULL DEFAULT 'MEDIUM',
    reported_by  UUID         NOT NULL REFERENCES users(id),
    assigned_to  UUID         REFERENCES users(id),
    due_date     DATE,
    resolution   TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_issues_project_id ON issues(project_id);
CREATE INDEX IF NOT EXISTS idx_issues_deleted_at ON issues(deleted_at);

CREATE TABLE IF NOT EXISTS risks (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title        VARCHAR(500) NOT NULL,
    description  TEXT,
    status       VARCHAR(50)  NOT NULL DEFAULT 'IDENTIFIED',
    probability  VARCHAR(50)  NOT NULL DEFAULT 'MEDIUM',
    impact       VARCHAR(50)  NOT NULL DEFAULT 'MEDIUM',
    mitigation   TEXT,
    owned_by     UUID         REFERENCES users(id),
    due_date     DATE,
    created_by   UUID         NOT NULL REFERENCES users(id),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_risks_project_id ON risks(project_id);
CREATE INDEX IF NOT EXISTS idx_risks_deleted_at ON risks(deleted_at);

CREATE TABLE IF NOT EXISTS project_budgets (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID          NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    category    VARCHAR(200)  NOT NULL,
    description TEXT,
    planned     NUMERIC(20,2) NOT NULL DEFAULT 0,
    actual      NUMERIC(20,2) NOT NULL DEFAULT 0,
    currency    VARCHAR(10)   NOT NULL DEFAULT 'IDR',
    created_by  UUID          NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_project_budgets_project_id ON project_budgets(project_id);

CREATE TABLE IF NOT EXISTS project_documents (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        VARCHAR(500) NOT NULL,
    file_url    TEXT         NOT NULL,
    file_size   BIGINT,
    mime_type   VARCHAR(200),
    category    VARCHAR(100),
    uploaded_by UUID         NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_project_documents_project_id ON project_documents(project_id);

CREATE TABLE IF NOT EXISTS project_progress_history (
    id           UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID          NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    progress_pct NUMERIC(5,2)  NOT NULL,
    notes        TEXT,
    recorded_by  UUID          NOT NULL REFERENCES users(id),
    recorded_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_project_progress_history_project_id ON project_progress_history(project_id);

CREATE TABLE IF NOT EXISTS approval_requests (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type  VARCHAR(100) NOT NULL,
    entity_id    UUID         NOT NULL,
    from_status  VARCHAR(100) NOT NULL,
    to_status    VARCHAR(100) NOT NULL,
    requested_by UUID         NOT NULL REFERENCES users(id),
    reviewed_by  UUID         REFERENCES users(id),
    status       VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    comment      TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_approval_requests_entity ON approval_requests(entity_type, entity_id);
