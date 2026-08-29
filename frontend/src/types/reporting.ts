// P2-009 Reporting Analytics types

export interface ReportDefinition {
  id: string
  organization_id: string
  name: string
  description: string
  category: 'EXECUTIVE' | 'PORTFOLIO' | 'PROJECT' | 'RISK' | 'BUDGET' | 'BENEFIT' | 'PRIORITY' | 'GENERAL'
  dataset_key: string
  visualization_type: 'TABLE' | 'BAR_CHART' | 'LINE_CHART' | 'PIE_CHART' | 'MAP' | 'KPI_CARD'
  available: boolean
  requires_powerbi: boolean
  embed_configured: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface DatasetFilter {
  period_start?: string
  period_end?: string
  program_id?: string
  status?: string
  province?: string
}

export interface ExecutiveSummaryData {
  total_projects: number
  active_projects: number
  completed_projects: number
  on_hold_projects: number
  avg_progress_pct: number
  total_budget_plan: number
  total_budget_actual: number
  budget_usage_pct: number
  total_risks: number
  open_risks: number
  high_risks: number
  total_issues: number
  open_issues: number
  green_health: number
  yellow_health: number
  red_health: number
  critical_health: number
}

export interface ProjectPerformanceRow {
  project_id: string
  project_code: string
  project_name: string
  status: string
  progress_pct: number
  start_date: string | null
  end_date: string | null
  budget_plan: number
  budget_actual: number
  budget_usage_pct: number
  health_class: 'GREEN' | 'YELLOW' | 'RED' | 'CRITICAL' | null
  province: string | null
  priority_score: number | null
  priority_category: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL' | null
}

export interface RiskIssueRow {
  project_id: string
  project_code: string
  project_name: string
  total_risks: number
  open_risks: number
  high_risks: number
  critical_risks: number
  total_issues: number
  open_issues: number
  high_issues: number
  critical_issues: number
}

export interface BudgetRow {
  project_id: string
  project_code: string
  project_name: string
  status: string
  budget_plan: number
  budget_actual: number
  variance: number
  usage_pct: number
}

export interface BenefitRow {
  project_id: string
  project_code: string
  project_name: string
  indicator_id: string
  indicator_name: string
  unit: string
  target: number
  actual: number
  achievement_pct: number
  aggregation_method: string
}

export interface PriorityRow {
  project_id: string
  project_code: string
  project_name: string
  total_score: number
  score_category: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL'
  calculated_at: string
}

export interface PowerBIConfig {
  configured: boolean
  workspace_id: string
  report_id: string
  tenant_id: string
  embed_url: string
  auth_method: string
}

export type ExportFormat = 'PDF' | 'XLSX' | 'CSV'
export type ExportStatus = 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED'

export interface ReportExportRequest {
  id: string
  organization_id: string
  report_id: string | null
  dataset_key: string
  format: ExportFormat
  status: ExportStatus
  parameters: Record<string, unknown>
  // Legacy field kept for backward compat
  file_url: string | null
  error_message: string | null
  // UAT-002 file metadata (populated when status === COMPLETED)
  file_name: string | null
  storage_key: string | null
  mime_type: string | null
  file_size: number | null
  generated_at: string | null
  requested_by: string
  started_at: string | null
  completed_at: string | null
  created_at: string
  updated_at: string
}

export interface CreateExportRequestInput {
  dataset_key: string
  report_id?: string
  format: ExportFormat
  parameters?: Record<string, unknown>
}

export type DatasetKey =
  | 'executive-summary'
  | 'project-performance'
  | 'risk-issue'
  | 'budget'
  | 'benefits'
  | 'priority'
