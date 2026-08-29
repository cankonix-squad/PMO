export interface NationalSummary {
  total_projects: number;
  active_projects: number;
  draft_projects: number;
  total_budget: number;
  budget_realized: number;
  budget_usage_pct: number;
  avg_physical_progress: number;
  health_green: number;
  health_yellow: number;
  health_red: number;
  health_critical: number;
  health_unscored: number;
  open_risks: number;
  high_risks: number;
  open_issues: number;
  critical_issues: number;
  overdue_actions: number;
  open_escalations: number;
  pending_decisions: number;
  overdue_decisions: number;
  benefit_indicators: number;
  as_of: string;
}

export interface CriticalProject {
  project_id: string;
  project_code: string;
  project_name: string;
  status: string;
  health_class: string;
  physical_actual: number;
  budget_total: number;
  open_risks: number;
  open_issues: number;
  priority_score: number;
  priority_class: string;
  program_name: string;
  sector_name: string;
}

export interface EscalationItem {
  id: string;
  project_id?: string;
  project_name?: string;
  level: string;
  source_type: string;
  reason: string;
  status: string;
  created_at: string;
}

export interface DecisionItem {
  id: string;
  project_id?: string;
  project_name?: string;
  subject: string;
  decision_text: string;
  status: string;
  due_date?: string;
  is_overdue: boolean;
  created_at: string;
}

export interface BenefitItem {
  id: string;
  name: string;
  unit: string;
  target_value: number;
  actual_value: number;
  achievement_pct: number;
  aggregation_method: string;
}

export interface BenefitSummary {
  total_indicators: number;
  on_track_count: number;
  behind_count: number;
  indicators: BenefitItem[];
}

export interface ProgramKPISummary {
  program_id: string;
  program_code: string;
  program_name: string;
  total_projects: number;
  active_projects: number;
  total_budget: number;
  avg_physical_progress: number;
  health_green: number;
  health_yellow: number;
  health_red: number;
  health_critical: number;
  open_risks: number;
  open_issues: number;
}

export interface ExecutiveDashboard {
  summary: NationalSummary;
  critical_projects: CriticalProject[];
  escalations: EscalationItem[];
  pending_decisions: DecisionItem[];
  programs: ProgramKPISummary[];
  benefits: BenefitSummary;
  as_of: string;
}
