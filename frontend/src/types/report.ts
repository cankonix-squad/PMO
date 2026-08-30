// Report types — P1-007

export type ReportPeriodType = "WEEKLY" | "MONTHLY" | "QUARTERLY";
export type ReportStatus = "DRAFT" | "PUBLISHED" | "ARCHIVED";

export interface SnapshotMetrics {
  // Projects
  total_projects: number;
  active_projects: number;
  done_projects: number;
  // Tasks
  total_tasks: number;
  done_tasks: number;
  overdue_tasks: number;
  avg_progress_pct: number;
  // Milestones
  total_milestones: number;
  done_milestones: number;
  overdue_milestones: number;
  // Risks
  total_risks: number;
  open_risks: number;
  high_risks: number;
  // Issues
  total_issues: number;
  open_issues: number;
  // Budget
  total_planned_budget: number;
  total_actual_budget: number;
  budget_usage_pct: number;
  // Corrective Actions
  total_corrective_actions: number;
  open_corrective_actions: number;
}

export interface ReportSnapshot {
  id: string;
  organization_id: string;
  period_type: ReportPeriodType;
  period_label: string;
  period_start: string; // date string
  period_end: string;   // date string
  project_id?: string | null;
  metrics: SnapshotMetrics;
  executive_summary?: string | null;
  status: ReportStatus;
  export_format?: string | null;
  export_url?: string | null;
  created_by: string;
  published_at?: string | null;
  published_by?: string | null;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
  // preloaded relations
  created_by_user?: { id: string; name: string } | null;
  project?: { id: string; name: string } | null;
}

export interface ListReportFilter {
  period_type?: ReportPeriodType | string;
  status?: ReportStatus | string;
  project_id?: string;
  page?: number;
  page_size?: number;
}

export interface GenerateReportRequest {
  period_type: ReportPeriodType;
  period_label: string;
  period_start: string; // YYYY-MM-DD
  period_end: string;   // YYYY-MM-DD
  project_id?: string;
}

export interface CreateReportRequest {
  period_type: ReportPeriodType;
  period_label: string;
  period_start: string;
  period_end: string;
  project_id?: string;
  executive_summary?: string;
}

export interface UpdateReportRequest {
  executive_summary?: string;
}

export interface TransitionReportRequest {
  to_status: ReportStatus;
}
