export type DashboardWarningType =
  | "OVERDUE_TASK"
  | "OVERDUE_MILESTONE"
  | "LOW_PROGRESS_NEAR_END"
  | "BUDGET_THRESHOLD"
  | "RISK_REGISTER";

export type DashboardWarningSeverity = "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";

export interface DashboardWarning {
  id: string;
  type: DashboardWarningType;
  severity: DashboardWarningSeverity;
  title: string;
  message: string;
  project_id?: string;
  project_name?: string;
  due_date?: string;
  value?: number;
  threshold?: number;
}

export interface TrendPoint {
  month: string;         // "YYYY-MM"
  physical_pct: number;  // avg physical progress %
  financial_pct: number; // actual/planned * 100
  planned: number;       // total planned budget for the month
  actual: number;        // total actual budget for the month
  data_type: string;     // "OPERATIONAL" | "SNAPSHOT"
}

export interface DashboardTrend {
  points: TrendPoint[];
  data_type: string; // "OPERATIONAL" — no official snapshot yet
}

export interface DashboardStats {
  total_projects: number;
  active_projects: number;
  on_hold_projects: number;
  closed_projects: number;
  total_tasks: number;
  todo_tasks: number;
  in_progress_tasks: number;
  done_tasks: number;
  overdue_tasks: number;
  total_milestones: number;
  pending_milestones: number;
  done_milestones: number;
  overdue_milestones: number;
  total_users: number;
  active_users: number;
  early_warnings: DashboardWarning[];
}
