-- 000012_corrective_actions.up.sql
-- Adds corrective_actions table for P1-006.
-- A corrective action records a deviation finding, its root cause, recommended
-- fix, PIC, target date, and follow-up status.  It links to a source entity
-- (issue, risk, or task) via nullable FK columns so the source type can vary.
-- FSM: DRAFT → SUBMITTED → IN_PROGRESS → COMPLETED | REJECTED
--      REJECTED → DRAFT  (revise & resubmit)

CREATE TABLE IF NOT EXISTS corrective_actions (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID            NOT NULL REFERENCES organizations(id),
    project_id          UUID            NOT NULL REFERENCES projects(id),

    -- Deviation description
    title               VARCHAR(500)    NOT NULL,
    deviation           TEXT            NOT NULL,
    root_cause          TEXT,
    recommendation      TEXT,

    -- PIC & schedule
    pic_user_id         UUID            REFERENCES users(id),
    target_date         DATE,

    -- Source linkage (at most one non-null expected per row)
    source_type         VARCHAR(50),    -- 'issue' | 'risk' | 'task' | null
    source_issue_id     UUID            REFERENCES issues(id)  ON DELETE SET NULL,
    source_risk_id      UUID            REFERENCES risks(id)   ON DELETE SET NULL,
    source_task_id      UUID            REFERENCES tasks(id)   ON DELETE SET NULL,

    -- Workflow
    status              VARCHAR(50)     NOT NULL DEFAULT 'DRAFT',
    -- DRAFT | SUBMITTED | IN_PROGRESS | COMPLETED | REJECTED

    -- Evidence note (free-text reference; full file linkage via P1-005 documents)
    evidence_note       TEXT,

    -- Audit
    created_by          UUID            NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_corrective_actions_org    ON corrective_actions(organization_id);
CREATE INDEX IF NOT EXISTS idx_corrective_actions_proj   ON corrective_actions(project_id);
CREATE INDEX IF NOT EXISTS idx_corrective_actions_status ON corrective_actions(status);
CREATE INDEX IF NOT EXISTS idx_corrective_actions_del    ON corrective_actions(deleted_at);
