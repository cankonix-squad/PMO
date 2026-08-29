export type SyncDirection = "IMPORT" | "EXPORT";
export type SyncFormat = "XER" | "PMXML";
export type SyncStatus = "PENDING" | "RUNNING" | "DONE" | "FAILED" | "CANCELLED";
export type ActivityAction = "CREATE" | "UPDATE" | "SKIP" | "CONFLICT";

export interface SyncRun {
  id: string;
  organization_id: string;
  project_id?: string;
  direction: SyncDirection;
  source_file_name: string;
  source_file_size: number;
  source_mime_type: string;
  format: SyncFormat;
  status: SyncStatus;
  total_activities: number;
  imported_activities: number;
  skipped_activities: number;
  failed_activities: number;
  conflict_count: number;
  error_summary: string; // raw JSON array string
  conflict_report: string; // raw JSON array string
  lineage: string; // raw JSON object string
  started_at?: string;
  finished_at?: string;
  triggered_by: string;
  created_at: string;
  updated_at: string;
}

export interface ActivityMapping {
  id: string;
  organization_id: string;
  project_id: string;
  sync_run_id: string;
  p6_activity_id: string;
  p6_wbs_code: string;
  p6_activity_name: string;
  entity_type: string;
  entity_id: string;
  action: ActivityAction;
  baseline_physical: number;
  actual_physical: number;
  planned_start?: string;
  planned_end?: string;
  actual_start?: string;
  actual_end?: string;
  raw_payload: string;
  created_at: string;
  updated_at: string;
}

export interface ListRunsParams {
  project_id?: string;
  status?: SyncStatus | "";
  format?: SyncFormat | "";
  page?: number;
  page_size?: number;
}

export interface ListMappingsParams {
  page?: number;
  page_size?: number;
}

export interface PaginatedRunsResponse {
  data: SyncRun[];
  meta: {
    total: number;
    page: number;
    page_size: number;
  };
}

export interface PaginatedMappingsResponse {
  data: ActivityMapping[];
  meta: {
    total: number;
    page: number;
    page_size: number;
  };
}

export interface ConflictEntry {
  activity_id: string;
  field: string;
  existing: string;
  incoming: string;
}

export interface SyncErrorEntry {
  code: string;
  message: string;
  activity_id?: string;
  row?: number;
}

export interface LineageMeta {
  source_project_id?: string;
  exported_at?: string;
  p6_version?: string;
  operator?: string;
}
