-- 000009_risks_tenant_score.down.sql
-- Reverts risk tenant guard and numeric score columns.

DROP INDEX IF EXISTS idx_risks_risk_score;
DROP INDEX IF EXISTS idx_risks_organization_id;

ALTER TABLE risks
    DROP CONSTRAINT IF EXISTS fk_risks_organization;

ALTER TABLE risks
    DROP COLUMN IF EXISTS severity;

ALTER TABLE risks
    DROP COLUMN IF EXISTS risk_score;

-- Revert probability/impact to the original VARCHAR semantics (MEDIUM default).
ALTER TABLE risks
    ALTER COLUMN probability TYPE VARCHAR(50)
    USING CASE
        WHEN probability <= 2 THEN 'LOW'
        WHEN probability = 3 THEN 'MEDIUM'
        WHEN probability = 4 THEN 'HIGH'
        ELSE 'VERY_HIGH'
    END;

ALTER TABLE risks
    ALTER COLUMN impact TYPE VARCHAR(50)
    USING CASE
        WHEN impact <= 2 THEN 'LOW'
        WHEN impact = 3 THEN 'MEDIUM'
        WHEN impact = 4 THEN 'HIGH'
        ELSE 'VERY_HIGH'
    END;

ALTER TABLE risks
    ALTER COLUMN probability SET DEFAULT 'MEDIUM';

ALTER TABLE risks
    ALTER COLUMN impact SET DEFAULT 'MEDIUM';

ALTER TABLE risks
    DROP COLUMN IF EXISTS organization_id;
