-- 000009_risks_tenant_score.up.sql
-- Adds tenant guard (organization_id), numeric probability/impact, risk_score,
-- and severity to the risks table.

ALTER TABLE risks
    ADD COLUMN IF NOT EXISTS organization_id UUID;

UPDATE risks r
SET organization_id = p.organization_id
FROM projects p
WHERE r.project_id = p.id
  AND r.organization_id IS NULL;

ALTER TABLE risks
    ALTER COLUMN organization_id SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_risks_organization'
    ) THEN
        ALTER TABLE risks
            ADD CONSTRAINT fk_risks_organization
            FOREIGN KEY (organization_id) REFERENCES organizations(id);
    END IF;
END $$;

-- probability/impact change from VARCHAR to INT (1-5).
-- Drop the string default first so the type can be altered, then restore it.
ALTER TABLE risks
    ALTER COLUMN probability DROP DEFAULT;

ALTER TABLE risks
    ALTER COLUMN impact DROP DEFAULT;

ALTER TABLE risks
    ALTER COLUMN probability TYPE INTEGER
    USING CASE probability
        WHEN 'LOW' THEN 2
        WHEN 'MEDIUM' THEN 3
        WHEN 'HIGH' THEN 4
        WHEN 'VERY_HIGH' THEN 5
        ELSE 3
    END;

ALTER TABLE risks
    ALTER COLUMN impact TYPE INTEGER
    USING CASE impact
        WHEN 'LOW' THEN 2
        WHEN 'MEDIUM' THEN 3
        WHEN 'HIGH' THEN 4
        WHEN 'VERY_HIGH' THEN 5
        ELSE 3
    END;

ALTER TABLE risks
    ALTER COLUMN probability SET DEFAULT 3;

ALTER TABLE risks
    ALTER COLUMN impact SET DEFAULT 3;

ALTER TABLE risks
    ADD COLUMN IF NOT EXISTS risk_score INTEGER NOT NULL DEFAULT 9;

ALTER TABLE risks
    ADD COLUMN IF NOT EXISTS severity VARCHAR(20) NOT NULL DEFAULT 'MEDIUM';

-- Backfill score/severity for existing rows.
UPDATE risks
SET risk_score = GREATEST(1, LEAST(probability, 5)) * GREATEST(1, LEAST(impact, 5)),
    severity   = CASE
        WHEN GREATEST(1, LEAST(probability, 5)) * GREATEST(1, LEAST(impact, 5)) >= 16 THEN 'CRITICAL'
        WHEN GREATEST(1, LEAST(probability, 5)) * GREATEST(1, LEAST(impact, 5)) >= 10 THEN 'HIGH'
        WHEN GREATEST(1, LEAST(probability, 5)) * GREATEST(1, LEAST(impact, 5)) >= 5  THEN 'MEDIUM'
        ELSE 'LOW'
    END
WHERE risk_score = 9 AND severity = 'MEDIUM';

CREATE INDEX IF NOT EXISTS idx_risks_organization_id ON risks(organization_id);
CREATE INDEX IF NOT EXISTS idx_risks_risk_score ON risks(risk_score) WHERE deleted_at IS NULL;
