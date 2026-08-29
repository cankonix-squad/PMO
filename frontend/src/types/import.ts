export type ImportDatasetType =
  | "project_progress"
  | "project_budgets"
  | "risks"
  | "issues"
  | "benefit_measurements";

export type ImportStatus =
  | "UPLOADED"
  | "VALIDATED"
  | "COMMITTED"
  | "FAILED"
  | "CANCELLED";

export type ImportAction = "CREATE" | "UPDATE" | "SKIP";

export interface ImportColumnDef {
  name: string;
  required: boolean;
  description: string;
}

export interface ImportTemplate {
  dataset_type: ImportDatasetType;
  label: string;
  description: string;
  columns: ImportColumnDef[];
}

export interface ImportJob {
  id: string;
  organization_id: string;
  dataset_type: ImportDatasetType;
  file_name: string;
  file_size: number;
  mime_type: string;
  status: ImportStatus;
  total_rows: number;
  valid_rows: number;
  invalid_rows: number;
  error_summary: string; // raw JSON array string from backend, e.g. '["row 1: ..."]'
  uploaded_by: string;
  validated_at?: string;
  committed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ImportRow {
  id: string;
  job_id: string;
  row_number: number;
  raw_payload: Record<string, string>;
  normalized_payload: Record<string, unknown>;
  valid: boolean;
  errors: string[];
  action: ImportAction;
  target_entity_id?: string;
  created_at: string;
}

export interface ListJobsParams {
  dataset_type?: ImportDatasetType;
  status?: ImportStatus;
  page?: number;
  page_size?: number;
}

export interface ListRowsParams {
  valid?: boolean;
  page?: number;
  page_size?: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
}
