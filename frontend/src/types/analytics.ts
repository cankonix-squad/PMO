// Program & Sector Dashboard Analytics — Type definitions
// Mirrors backend/internal/modules/analytics/program/model.go

export type GroupType = "program" | "sector";

export const HEALTH_CLASS_COLOR: Record<string, string> = {
  GREEN: "bg-emerald-100 text-emerald-800",
  YELLOW: "bg-yellow-100 text-yellow-800",
  RED: "bg-rose-100 text-rose-800",
  CRITICAL: "bg-red-200 text-red-900",
  "": "bg-gray-100 text-gray-500",
};

export const HEALTH_CLASS_LABEL: Record<string, string> = {
  GREEN: "Baik",
  YELLOW: "Perlu Perhatian",
  RED: "Berisiko",
  CRITICAL: "Kritis",
  "": "Belum Dinilai",
};

// ── KPI ──────────────────────────────────────────────────────────────────────

export interface ProgramKPI {
  group_id: string;
  group_code: string;
  group_name: string;
  group_type: GroupType;
  total_projects: number;
  active_projects: number;
  // Budget
  total_budget: number;
  budget_realized: number;
  budget_usage_pct: number;
  currency: string;
  // Physical progress
  avg_physical_actual: number;
  avg_physical_target: number;
  physical_variance: number;
  // Health distribution
  health_green: number;
  health_yellow: number;
  health_red: number;
  health_critical: number;
  health_unscored: number;
  // Risk & issues
  open_risks: number;
  high_risks: number;
  open_issues: number;
  critical_issues: number;
  overdue_actions: number;
  // Priority
  avg_priority_score: number;
  critical_priority_count: number;
  // Benefit
  benefit_indicators: number;
  as_of: string;
}

// ── Project row ───────────────────────────────────────────────────────────────

export interface ProjectRow {
  project_id: string;
  project_code: string;
  project_name: string;
  status: string;
  physical_actual: number;
  physical_target: number;
  physical_variance: number;
  budget_total: number;
  budget_usage_pct: number;
  health_class: string;
  health_score: number;
  open_risks: number;
  open_issues: number;
  priority_score: number;
  priority_class: string;
}

// ── Deviation ────────────────────────────────────────────────────────────────

export interface TopDeviation {
  project_id: string;
  project_code: string;
  project_name: string;
  value: number;
  label: string;
  group_name: string;
}

// ── Dashboard ────────────────────────────────────────────────────────────────

export interface ProgramDashboard {
  kpi: ProgramKPI;
  projects: ProjectRow[];
  top_physical_deviation: TopDeviation[];
  top_budget_deviation: TopDeviation[];
  high_risk_projects: ProjectRow[];
  as_of: string;
}

export interface ListResponse {
  groups: ProgramKPI[];
  as_of: string;
}
