CREATE TABLE IF NOT EXISTS command_escalations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    project_id UUID REFERENCES projects(id),
    source_type VARCHAR(50) NOT NULL,
    source_id UUID NOT NULL,
    level VARCHAR(30) NOT NULL CHECK (level IN ('PROJECT_MANAGER','PROGRAM_MANAGER','EXECUTIVE')),
    reason TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN','ACKNOWLEDGED','CLOSED')),
    created_by UUID REFERENCES users(id),
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_command_escalations_scope ON command_escalations(organization_id, status, created_at);

CREATE TABLE IF NOT EXISTS executive_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    project_id UUID REFERENCES projects(id),
    escalation_id UUID REFERENCES command_escalations(id),
    subject VARCHAR(500) NOT NULL,
    decision_text TEXT NOT NULL,
    owner_id UUID REFERENCES users(id),
    due_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN','IN_PROGRESS','COMPLETED','CANCELLED')),
    decided_by UUID REFERENCES users(id),
    decided_at TIMESTAMPTZ,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_executive_decisions_scope ON executive_decisions(organization_id, status, due_date);