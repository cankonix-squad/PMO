-- Migration 000021: Priority Scoring & Decision Support (P2-004)
-- Creates tables for priority formulas, formula components, score snapshots, and score components.

-- 1. Priority formulas (versioned, one ACTIVE per org)
CREATE TABLE IF NOT EXISTS priority_formulas (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID        NOT NULL REFERENCES organizations(id),
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    version         INT         NOT NULL DEFAULT 1,
    status          VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT','ACTIVE','ARCHIVED')),
    missing_data_rule VARCHAR(20) NOT NULL DEFAULT 'PENALIZE' CHECK (missing_data_rule IN ('PENALIZE','EXCLUDE','NEUTRAL')),
    -- category thresholds stored as JSONB: {"LOW":{"min":0,"max":39},...}
    category_thresholds JSONB   NOT NULL DEFAULT '{}',
    activated_by    UUID        REFERENCES users(id),
    activated_at    TIMESTAMPTZ,
    created_by      UUID        REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_priority_formulas_active_org
    ON priority_formulas (organization_id)
    WHERE status = 'ACTIVE' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_priority_formulas_org
    ON priority_formulas (organization_id)
    WHERE deleted_at IS NULL;

-- 2. Priority formula components (weights per scoring dimension)
CREATE TABLE IF NOT EXISTS priority_formula_components (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    formula_id      UUID        NOT NULL REFERENCES priority_formulas(id) ON DELETE CASCADE,
    organization_id UUID        NOT NULL REFERENCES organizations(id),
    component_key   VARCHAR(100) NOT NULL,   -- e.g. health_score, risk_score, issue_severity, budget_usage, schedule_variance, corrective_action_overdue, benefit_indicator
    label           VARCHAR(200) NOT NULL,
    weight          NUMERIC(5,4) NOT NULL DEFAULT 0 CHECK (weight >= 0 AND weight <= 1),
    sort_order      INT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_priority_formula_components_key
    ON priority_formula_components (formula_id, component_key);

CREATE INDEX IF NOT EXISTS idx_priority_formula_components_formula
    ON priority_formula_components (formula_id);

-- 3. Project priority scores (snapshot per calculation run)
CREATE TABLE IF NOT EXISTS project_priority_scores (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID        NOT NULL REFERENCES organizations(id),
    project_id      UUID        NOT NULL REFERENCES projects(id),
    formula_id      UUID        NOT NULL REFERENCES priority_formulas(id),
    formula_version INT         NOT NULL,
    total_score     NUMERIC(6,3) NOT NULL DEFAULT 0,  -- 0-100
    score_category  VARCHAR(20) NOT NULL DEFAULT 'LOW' CHECK (score_category IN ('LOW','MEDIUM','HIGH','CRITICAL')),
    rank_in_org     INT,        -- populated during batch calculation
    missing_components INT      NOT NULL DEFAULT 0,
    calculated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    calculated_by   UUID        REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_project_priority_scores_org_project
    ON project_priority_scores (organization_id, project_id);

CREATE INDEX IF NOT EXISTS idx_project_priority_scores_formula
    ON project_priority_scores (formula_id);

CREATE INDEX IF NOT EXISTS idx_project_priority_scores_org_score
    ON project_priority_scores (organization_id, total_score DESC);

-- Latest score per project (for fast lookups)
CREATE INDEX IF NOT EXISTS idx_project_priority_scores_latest
    ON project_priority_scores (organization_id, project_id, calculated_at DESC);

-- 4. Project priority score components (explainability per dimension)
CREATE TABLE IF NOT EXISTS project_priority_score_components (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    score_id        UUID        NOT NULL REFERENCES project_priority_scores(id) ON DELETE CASCADE,
    organization_id UUID        NOT NULL REFERENCES organizations(id),
    project_id      UUID        NOT NULL REFERENCES projects(id),
    component_key   VARCHAR(100) NOT NULL,
    label           VARCHAR(200) NOT NULL,
    raw_value       NUMERIC(10,4),      -- actual value from source (null = missing)
    raw_unit        VARCHAR(50),        -- e.g. "%", "score", "count"
    normalized_score NUMERIC(6,3),      -- 0-100 normalized score
    weight          NUMERIC(5,4) NOT NULL DEFAULT 0,
    weighted_score  NUMERIC(6,3),       -- normalized_score * weight
    available       BOOLEAN      NOT NULL DEFAULT TRUE,
    note            TEXT,               -- source/explanation
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pps_components_score
    ON project_priority_score_components (score_id);

CREATE INDEX IF NOT EXISTS idx_pps_components_org_project
    ON project_priority_score_components (organization_id, project_id);
