// ---------------------------------------------------------------------------
// Data Governance — Official Data Validation & Approval Workflow (CANKORA-P3-003)
// ---------------------------------------------------------------------------

export type GovernanceSourceType =
  | "MANUAL"
  | "CSV_IMPORT"
  | "PRIMAVERA"
  | "GOVERNMENT"
  | "BIM"
  | "API";

export type GovernanceDatasetType =
  | "PROJECT_PROGRESS"
  | "BUDGET"
  | "RISK"
  | "ISSUE"
  | "BENEFIT"
  | "LOCATION"
  | "CONTRACT"
  | "DOCUMENT"
  | "OTHER";

export type GovernanceSubmissionStatus =
  | "DRAFT"
  | "SUBMITTED"
  | "IN_REVIEW"
  | "APPROVED"
  | "REJECTED"
  | "LOCKED"
  | "CANCELLED";

export type GovernanceItemAction =
  | "CREATE"
  | "UPDATE"
  | "DELETE"
  | "UPSERT"
  | "VALIDATE_ONLY";

export type GovernanceItemValidationStatus = "PENDING" | "VALID" | "INVALID";

export type GovernanceLockStatus = "OPEN" | "LOCKED";

// ---------------------------------------------------------------------------
// Submission Item
// ---------------------------------------------------------------------------

export interface GovernanceSubmissionItem {
  id: string;
  submission_id: string;
  entity_type: string;
  entity_id?: string;
  action: GovernanceItemAction;
  payload_before?: Record<string, unknown>;
  payload_after: Record<string, unknown>;
  validation_status: GovernanceItemValidationStatus;
  validation_errors: string[];
  created_at: string;
}

// ---------------------------------------------------------------------------
// Submission
// ---------------------------------------------------------------------------

export interface GovernanceSubmission {
  id: string;
  organization_id: string;
  project_id?: string;
  snapshot_id?: string;
  source?: string;
  source_reference?: string;
  source_type: GovernanceSourceType;
  dataset_type: GovernanceDatasetType;
  source_entity_type?: string;
  source_entity_id?: string;
  period_year: number;
  period_month?: number;
  status: GovernanceSubmissionStatus;
  completeness_pct: number;
  freshness_at?: string;
  freshness_days?: number;
  sla_due_at?: string;
  submitted_by?: string;
  submitted_at?: string;
  validator_id?: string;
  validated_at?: string;
  rejection_reason?: string;
  review_notes?: string;
  reviewed_by?: string;
  reviewed_at?: string;
  approved_by?: string;
  approved_at?: string;
  locked_by?: string;
  locked_at?: string;
  lineage: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface GovernanceSubmissionDetail extends GovernanceSubmission {
  items: GovernanceSubmissionItem[];
}

// ---------------------------------------------------------------------------
// Lock Period
// ---------------------------------------------------------------------------

export interface GovernanceLockPeriod {
  id: string;
  organization_id: string;
  dataset_type: GovernanceDatasetType;
  period_year: number;
  period_month?: number;
  status: GovernanceLockStatus;
  locked_by?: string;
  locked_at?: string;
  lock_reason?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

export interface CreateGovernanceItemRequest {
  entity_type: string;
  entity_id?: string;
  action: GovernanceItemAction;
  payload_after?: Record<string, unknown>;
  payload_before?: Record<string, unknown>;
}

export interface CreateGovernanceSubmissionRequest {
  dataset_type: GovernanceDatasetType;
  source_type: GovernanceSourceType;
  source_entity_type?: string;
  source_entity_id?: string;
  period_year: number;
  period_month?: number;
  source_reference?: string;
  items: CreateGovernanceItemRequest[];
}

export interface ReviewGovernanceRequest {
  review_notes?: string;
}

export interface RejectGovernanceRequest {
  rejection_reason: string;
}

export interface LockGovernanceRequest {
  lock_reason?: string;
}

export interface CancelGovernanceRequest {
  cancel_reason?: string;
}

export interface CreateLockPeriodRequest {
  dataset_type: GovernanceDatasetType;
  period_year: number;
  period_month?: number;
  lock_reason?: string;
  lock_now: boolean;
}

// ---------------------------------------------------------------------------
// List filters
// ---------------------------------------------------------------------------

export interface GovernanceSubmissionFilter {
  status?: GovernanceSubmissionStatus;
  dataset_type?: GovernanceDatasetType;
  source_type?: GovernanceSourceType;
  period_year?: number;
  page?: number;
  page_size?: number;
}

export interface GovernanceLockPeriodFilter {
  dataset_type?: GovernanceDatasetType;
  status?: GovernanceLockStatus;
  period_year?: number;
  page?: number;
  page_size?: number;
}

// ---------------------------------------------------------------------------
// Response wrappers
// ---------------------------------------------------------------------------

export interface GovernanceSubmissionListResponse {
  data: GovernanceSubmission[];
  meta: { total: number; page: number; page_size: number };
}

export interface GovernanceLockPeriodListResponse {
  data: GovernanceLockPeriod[];
  meta: { total: number; page: number; page_size: number };
}

// ---------------------------------------------------------------------------
// Labels for UI
// ---------------------------------------------------------------------------

export const GOVERNANCE_DATASET_LABELS: Record<GovernanceDatasetType, string> = {
  PROJECT_PROGRESS: "Progress Proyek",
  BUDGET: "Anggaran",
  RISK: "Risiko",
  ISSUE: "Isu",
  BENEFIT: "Benefit",
  LOCATION: "Lokasi",
  CONTRACT: "Kontrak",
  DOCUMENT: "Dokumen",
  OTHER: "Lainnya",
};

export const GOVERNANCE_SOURCE_LABELS: Record<GovernanceSourceType, string> = {
  MANUAL: "Manual",
  CSV_IMPORT: "Import CSV",
  PRIMAVERA: "Primavera P6",
  GOVERNMENT: "Pemerintah",
  BIM: "BIM / Digital Twin",
  API: "API",
};

export const GOVERNANCE_STATUS_LABELS: Record<GovernanceSubmissionStatus, string> = {
  DRAFT: "Draft",
  SUBMITTED: "Diajukan",
  IN_REVIEW: "Dalam Review",
  APPROVED: "Disetujui",
  REJECTED: "Ditolak",
  LOCKED: "Terkunci",
  CANCELLED: "Dibatalkan",
};
