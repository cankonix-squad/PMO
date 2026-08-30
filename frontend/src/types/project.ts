import type { PaginationParams } from "./api";

export type ProjectStatus =
  | "DRAFT"
  | "PLANNING"
  | "ACTIVE"
  | "ON_HOLD"
  | "COMPLETED"
  | "CANCELLED";

export type Priority = "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";

export type TaskStatus = "TODO" | "IN_PROGRESS" | "IN_REVIEW" | "DONE" | "BLOCKED";

export type TaskType = "TASK" | "BUG" | "FEATURE" | "RESEARCH";

export type MilestoneStatus = "PENDING" | "IN_PROGRESS" | "COMPLETED" | "DELAYED";

export interface Project {
  id: string;
  organization_id: string;
  org_unit_id: string | null;
  org_unit_name: string | null;
  program_id: string | null;
  sector_id: string | null;
  region_id: string | null;
  river_basin_id: string | null;
  code: string;
  name: string;
  description: string | null;
  objectives: string | null;
  status: ProjectStatus;
  priority: Priority;
  category: string | null;
  start_date: string | null;
  end_date: string | null;
  actual_end_date: string | null;
  budget_total: number;
  currency: string;
  progress_pct: number;
  manager_id: string | null;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateProjectRequest {
  code: string;
  name: string;
  description?: string;
  objectives?: string;
  priority?: Priority;
  category?: string;
  start_date?: string;
  end_date?: string;
  budget_total?: number;
  currency?: string;
  progress_pct?: number;
  org_unit_id?: string;
  program_id?: string;
  sector_id?: string;
  region_id?: string;
  river_basin_id?: string;
}

export interface UpdateProjectRequest extends Partial<CreateProjectRequest> {
  manager_id?: string;
  org_unit_id?: string;
  program_id?: string;
  sector_id?: string;
  region_id?: string;
  river_basin_id?: string;
  progress_pct?: number;
}

export interface TransitionProjectRequest {
  to_status: ProjectStatus;
  comment?: string;
}

export interface ProjectListFilter extends PaginationParams {
  status?: ProjectStatus;
  priority?: Priority;
  category?: string;
}

export interface ProgressHistory {
  id: string;
  project_id: string;
  progress_pct: number;
  notes: string | null;
  recorded_by: string;
  recorded_at: string;
}

export interface TaskAssignment {
  id: string;
  task_id: string;
  user_id: string;
  is_lead: boolean;
  assigned_at: string;
}

export interface Task {
  id: string;
  organization_id: string;
  project_id: string;
  milestone_id: string | null;
  parent_id: string | null;
  wbs_code: string | null;
  title: string;
  description: string | null;
  status: TaskStatus;
  priority: Priority;
  type: TaskType;
  start_date: string | null;
  due_date: string | null;
  est_hours: number;
  actual_hours: number;
  progress_pct: number;
  created_by: string;
  created_at: string;
  updated_at: string;
  subtasks?: Task[];
  assignments?: TaskAssignment[];
}

export interface TaskListFilter extends PaginationParams {
  milestone_id?: string;
  status?: TaskStatus;
  assigned_to?: string;
}

export interface CreateTaskRequest {
  milestone_id?: string;
  parent_id?: string;
  wbs_code?: string;
  title: string;
  description?: string;
  priority?: Priority;
  type?: TaskType;
  start_date?: string;
  due_date?: string;
  est_hours?: number;
}

export interface UpdateTaskRequest extends Partial<CreateTaskRequest> {
  status?: TaskStatus;
  actual_hours?: number;
  progress_pct?: number;
}

export interface Milestone {
  id: string;
  organization_id: string;
  project_id: string;
  title: string;
  description: string | null;
  due_date: string | null;
  status: MilestoneStatus;
  progress_pct: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateMilestoneRequest {
  title: string;
  description?: string;
  due_date?: string;
}

export interface UpdateMilestoneRequest extends Partial<CreateMilestoneRequest> {
  status?: MilestoneStatus;
  progress_pct?: number;
}

export type IssueStatus = "OPEN" | "IN_PROGRESS" | "RESOLVED" | "CLOSED";
export type IssueSeverity = "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";
export type IssueEscalation =
  | "NONE"
  | "PROJECT_MANAGER"
  | "PROGRAM_MANAGER"
  | "EXECUTIVE";

export interface Issue {
  id: string;
  organization_id: string;
  project_id: string;
  task_id: string | null;
  title: string;
  description: string | null;
  status: IssueStatus;
  severity: IssueSeverity;
  escalation: IssueEscalation;
  reported_by: string;
  assigned_to: string | null;
  due_date: string | null;
  resolution: string | null;
  created_at: string;
  updated_at: string;
}

export interface IssueListFilter extends PaginationParams {
  status?: IssueStatus;
  severity?: IssueSeverity;
  assigned_to?: string;
}

export interface CreateIssueRequest {
  task_id?: string;
  title: string;
  description?: string;
  severity?: IssueSeverity;
  escalation?: IssueEscalation;
  assigned_to?: string;
  due_date?: string;
  resolution?: string;
}

export interface UpdateIssueRequest extends Partial<CreateIssueRequest> {
  status?: IssueStatus;
}

export type RiskStatus =
  | "IDENTIFIED"
  | "ASSESSED"
  | "MITIGATED"
  | "ACCEPTED"
  | "ESCALATED"
  | "CLOSED";

export type RiskSeverity = "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";

export interface Risk {
  id: string;
  organization_id: string;
  project_id: string;
  title: string;
  description: string | null;
  status: RiskStatus;
  probability: number;
  impact: number;
  risk_score: number;
  severity: RiskSeverity;
  mitigation: string | null;
  owned_by: string | null;
  due_date: string | null;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface RiskListFilter extends PaginationParams {
  status?: RiskStatus;
  severity?: RiskSeverity;
  owned_by?: string;
}

export interface CreateRiskRequest {
  title: string;
  description?: string;
  probability?: number;
  impact?: number;
  mitigation?: string;
  owned_by?: string;
  due_date?: string;
}

export interface UpdateRiskRequest extends Partial<CreateRiskRequest> {
  status?: RiskStatus;
}

export type BudgetStatus = "NORMAL" | "WATCH" | "RISK" | "OVERRUN";

export interface BudgetLine {
  id: string;
  project_id: string;
  category: string;
  description: string | null;
  planned: number;
  actual: number;
  currency: string;
  variance: number;
  usage_pct: number;
  status: BudgetStatus;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface BudgetListFilter extends PaginationParams {
  category?: string;
}

export interface CreateBudgetRequest {
  category: string;
  description?: string;
  planned?: number;
  actual?: number;
  currency?: string;
}

export interface UpdateBudgetRequest extends Partial<CreateBudgetRequest> {}

export type VendorType = "VENDOR" | "CONSULTANT";

export interface Vendor {
  id: string;
  organization_id: string;
  name: string;
  type: VendorType;
  legal_name: string | null;
  tax_id: string | null;
  contact_person: string | null;
  email: string | null;
  phone: string | null;
  address: string | null;
  is_active: boolean;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface VendorListFilter extends PaginationParams {
  type?: VendorType;
  search?: string;
  is_active?: boolean;
}

export interface CreateVendorRequest {
  name: string;
  type: VendorType;
  legal_name?: string;
  tax_id?: string;
  contact_person?: string;
  email?: string;
  phone?: string;
  address?: string;
  is_active?: boolean;
}

export interface UpdateVendorRequest extends Partial<CreateVendorRequest> {}

export type ContractStatus =
  | "DRAFT"
  | "ACTIVE"
  | "AMENDED"
  | "COMPLETED"
  | "TERMINATED";

export interface Contract {
  id: string;
  organization_id: string;
  project_id: string;
  contract_number: string;
  title: string;
  vendor_id: string;
  consultant_id: string | null;
  contract_value: number;
  currency: string;
  signed_date: string | null;
  start_date: string | null;
  end_date: string | null;
  status: ContractStatus;
  scope_of_work: string | null;
  created_by: string;
  created_at: string;
  updated_at: string;
  vendor?: Vendor;
  consultant?: Vendor;
}

export interface ContractListFilter extends PaginationParams {
  status?: ContractStatus;
  vendor_id?: string;
  search?: string;
}

export interface CreateContractRequest {
  contract_number: string;
  title: string;
  vendor_id: string;
  consultant_id?: string | null;
  contract_value?: number;
  currency?: string;
  signed_date?: string;
  start_date?: string;
  end_date?: string;
  status?: ContractStatus;
  scope_of_work?: string;
}

export interface UpdateContractRequest extends Partial<CreateContractRequest> {
  consultant_id?: string | null;
  contract_value?: number;
}

export type DocumentCategory =
  | "CONTRACT"
  | "REPORT"
  | "EVIDENCE"
  | "PHOTO"
  | "BAST"
  | "TOR_KAK"
  | "OTHER";

export interface ProjectDocument {
  id: string;
  project_id: string;
  name: string;
  category: DocumentCategory | string | null;
  version: string | null;
  file_url: string;
  file_size: number | null;
  mime_type: string | null;
  uploaded_by: string;
  created_at: string;
  updated_at: string;
}

export interface DocumentListFilter extends PaginationParams {
  category?: string;
  search?: string;
}

export interface UploadDocumentRequest {
  file: File;
  name?: string;
  category?: string;
  version?: string;
}

export interface UpdateDocumentRequest {
  name?: string;
  category?: string;
  version?: string;
}

// ---------------------------------------------------------------------------
// Corrective Action (P1-006)
// ---------------------------------------------------------------------------

export type CorrectiveActionStatus =
  | "DRAFT"
  | "SUBMITTED"
  | "IN_PROGRESS"
  | "COMPLETED"
  | "REJECTED";

export type CorrectiveActionSourceType = "issue" | "risk" | "task";

export interface CorrectiveAction {
  id: string;
  organization_id: string;
  project_id: string;
  title: string;
  deviation: string;
  root_cause?: string;
  recommendation?: string;
  pic_user_id?: string | null;
  target_date?: string | null;
  source_type?: CorrectiveActionSourceType | string;
  source_issue_id?: string | null;
  source_risk_id?: string | null;
  source_task_id?: string | null;
  status: CorrectiveActionStatus;
  evidence_note?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CorrectiveActionListFilter extends PaginationParams {
  status?: CorrectiveActionStatus | string;
  source_type?: CorrectiveActionSourceType | string;
  search?: string;
}

export interface CreateCorrectiveActionRequest {
  title: string;
  deviation: string;
  root_cause?: string;
  recommendation?: string;
  pic_user_id?: string | null;
  target_date?: string | null;
  source_type?: CorrectiveActionSourceType | string;
  source_issue_id?: string | null;
  source_risk_id?: string | null;
  source_task_id?: string | null;
  evidence_note?: string;
}

export interface UpdateCorrectiveActionRequest {
  title?: string;
  deviation?: string;
  root_cause?: string;
  recommendation?: string;
  pic_user_id?: string | null;
  target_date?: string | null;
  source_type?: CorrectiveActionSourceType | string;
  source_issue_id?: string | null;
  source_risk_id?: string | null;
  source_task_id?: string | null;
  status?: CorrectiveActionStatus;
  evidence_note?: string;
}
