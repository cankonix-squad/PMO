-- Rollback migration 000021: Priority Scoring & Decision Support

DROP INDEX IF EXISTS idx_pps_components_org_project;
DROP INDEX IF EXISTS idx_pps_components_score;
DROP TABLE IF EXISTS project_priority_score_components;

DROP INDEX IF EXISTS idx_project_priority_scores_latest;
DROP INDEX IF EXISTS idx_project_priority_scores_org_score;
DROP INDEX IF EXISTS idx_project_priority_scores_formula;
DROP INDEX IF EXISTS idx_project_priority_scores_org_project;
DROP TABLE IF EXISTS project_priority_scores;

DROP INDEX IF EXISTS idx_priority_formula_components_formula;
DROP INDEX IF EXISTS uidx_priority_formula_components_key;
DROP TABLE IF EXISTS priority_formula_components;

DROP INDEX IF EXISTS idx_priority_formulas_org;
DROP INDEX IF EXISTS uidx_priority_formulas_active_org;
DROP TABLE IF EXISTS priority_formulas;
