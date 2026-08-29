export interface ProjectControlItem { id: string; title: string; status: string; severity?: string; due_date?: string }
export interface ProjectControl {
  as_of: string;
  project: { id: string; code: string; name: string; status: string; progress_pct: number; budget_total: number; currency: string };
  contract: { count: number; total_value: number; active_count: number; currency: string };
  snapshot?: { period_year: number; period_month: number; physical_actual: number; physical_target: number; physical_variance: number; financial_actual: number; financial_target: number; financial_variance: number; currency: string; schedule_deviation_days?: number; status: string; source?: string };
  health?: { score: number; class: string; formula_id: string; explanation: string };
  evidence: { inspections: number; verified_inspections: number; evidence_files: number };
  issues: ProjectControlItem[];
  risks: ProjectControlItem[];
  actions: ProjectControlItem[];
}