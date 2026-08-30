// Priority Scoring & Decision Support — Type definitions
// Mirrors backend/internal/modules/priority/model.go

export type FormulaStatus = "DRAFT" | "ACTIVE" | "ARCHIVED";
export type ScoreCategory = "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";
export type MissingDataRule = "SKIP" | "PENALIZE" | "ZERO";

export const COMPONENT_KEY_LABELS: Record<string, string> = {
  health_score: "Health Score",
  risk_score: "Risk Score",
  issue_severity: "Issue Severity",
  budget_usage: "Budget Usage",
  schedule_variance: "Schedule Variance",
  corrective_action_overdue: "Corrective Action Overdue",
  benefit_indicator: "Benefit Indicator",
};

export const SCORE_CATEGORY_COLOR: Record<ScoreCategory, string> = {
  LOW: "bg-emerald-100 text-emerald-800",
  MEDIUM: "bg-yellow-100 text-yellow-800",
  HIGH: "bg-orange-100 text-orange-800",
  CRITICAL: "bg-rose-100 text-rose-800",
};

// ── Formula ──────────────────────────────────────────────────────────────────

export interface FormulaComponent {
  id: string;
  formula_id: string;
  component_key: string;
  weight: number;
  created_at: string;
  updated_at: string;
}

export interface Formula {
  id: string;
  organization_id: string;
  name: string;
  version: number;
  status: FormulaStatus;
  missing_data_rule: MissingDataRule;
  components?: FormulaComponent[];
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface ComponentWeightInput {
  component_key: string;
  weight: number;
}

export interface CategoryThreshold {
  low_max: number;
  medium_max: number;
  high_max: number;
}

export interface CreateFormulaRequest {
  name: string;
  missing_data_rule?: MissingDataRule;
  components: ComponentWeightInput[];
  thresholds?: CategoryThreshold;
}

export interface UpdateFormulaRequest {
  name?: string;
  missing_data_rule?: MissingDataRule;
  components?: ComponentWeightInput[];
  thresholds?: CategoryThreshold;
}

// ── Score ─────────────────────────────────────────────────────────────────────

export interface ScoreComponent {
  id: string;
  score_id: string;
  organization_id?: string;
  project_id?: string;
  component_key: string;
  label?: string;
  raw_value: number | null;
  raw_unit?: string;
  normalized_score: number | null;
  weight: number;
  weighted_score: number;
  available: boolean;
  note: string;
  created_at: string;
}

export interface Score {
  id: string;
  organization_id: string;
  project_id: string;
  formula_id: string;
  total_score: number;
  score_category: ScoreCategory;
  rank_in_org: number;
  calculated_at: string;
  components?: ScoreComponent[];
}

export interface ProjectScoreSummary {
  score_id: string;
  project_id: string;
  project_name: string;
  project_code?: string;
  project_status: string;
  formula_id: string;
  formula_name: string;
  formula_version: number;
  total_score: number;
  score_category: ScoreCategory;
  rank_in_org: number | null;
  missing_components: number;
  calculated_at: string;
}

export interface RankingResponse {
  counts: Record<ScoreCategory, number>;
  projects: ProjectScoreSummary[];
}

export interface CalculateRequest {
  project_id: string;
  formula_id?: string;
}

export interface BatchCalculateRequest {
  project_ids?: string[];
  formula_id?: string;
}

export interface BatchCalculateResponse {
  formula_id: string;
  formula_version: number;
  calculated: number;
  skipped: number;
}
