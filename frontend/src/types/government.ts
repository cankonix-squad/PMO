export type ConnectorState = "NOT_CONFIGURED" | "SANDBOX_SAMPLE" | "ACTIVE";
export type SyncMode = "SAMPLE" | "DRY_RUN" | "COMMIT";
export type RunStatus = "PENDING" | "RUNNING" | "SUCCEEDED" | "FAILED" | "CANCELLED";
export type RecordStatus = "ACCEPTED" | "REJECTED" | "SKIPPED";
export type RecordAction = "CREATE" | "UPDATE" | "SKIP" | "CONFLICT";
export type MatchStatus = "PENDING_MATCH" | "MATCHED" | "REJECTED";
export type MatchConfidence = "EXACT_CODE" | "EXACT_NAME" | "EXACT_NPWP" | "PARTIAL_NAME" | "LOW_CONFIDENCE";

export interface ConnectorDefinition {
  key: string;
  name: string;
  description: string;
  dataset_types: string[];
  state: ConnectorState;
}

export interface ConnectorConfig {
  enabled: boolean;
  base_url: string;
  sandbox: boolean;
  state: ConnectorState;
}

export interface SyncRun {
  id: string;
  organization_id: string;
  connector_key: string;
  dataset_type: string;
  mode: SyncMode;
  status: RunStatus;
  started_by: string;
  total_records: number;
  accepted_records: number;
  rejected_records: number;
  error_summary: unknown; // JSONB
  source_hash: string;
  idempotency_key: string;
  started_at?: string;
  finished_at?: string;
  created_at: string;
  updated_at: string;
}

export interface SyncRecord {
  id: string;
  sync_run_id: string;
  organization_id: string;
  external_id: string;
  dataset_type: string;
  status: RecordStatus;
  action: RecordAction;
  validation_errors: unknown; // JSONB
  raw_payload: unknown; // JSONB
  created_at: string;
}

export interface ExternalMapping {
  id: string;
  organization_id: string;
  connector_key: string;
  dataset_type: string;
  external_id: string;
  internal_entity_type: string;
  internal_entity_id: string | null;
  match_status: MatchStatus;
  match_confidence?: number;
  match_reason?: string;
  matched_by?: string;
  matched_at?: string;
  rejected_by?: string;
  rejected_at?: string;
  reject_reason?: string;
  source_payload_hash: string;
  last_seen_at: string;
  sync_run_id: string;
  created_at: string;
  updated_at: string;
}

// ---------------------------------------------------------------------------
// Resolution types (P3-002)
// ---------------------------------------------------------------------------

export interface ResolutionCandidate {
  entity_id: string;
  entity_type: string;
  name: string;
  code?: string;
  confidence: number;
  reason: string;
}

export interface MatchMappingRequest {
  internal_entity_id: string;
  internal_entity_type: string;
  match_reason?: string;
  match_confidence?: number;
}

export interface RejectMappingRequest {
  reject_reason?: string;
}

export interface ListPendingMappingsParams {
  connector_key?: string;
  dataset_type?: string;
  page?: number;
  page_size?: number;
}

export interface CandidatesResponse {
  data: ResolutionCandidate[];
  meta: { total: number };
}

export interface CreateRunRequest {
  connector_key: string;
  dataset_type: string;
  mode: SyncMode;
  idempotency_key?: string;
}

export interface ListRunsParams {
  connector_key?: string;
  dataset_type?: string;
  status?: RunStatus | "";
  mode?: SyncMode | "";
  page?: number;
  page_size?: number;
}

export interface ListRecordsParams {
  status?: RecordStatus | "";
  action?: RecordAction | "";
  page?: number;
  page_size?: number;
}

export interface ListMappingsParams {
  connector_key?: string;
  dataset_type?: string;
  match_status?: MatchStatus | "";
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

export interface PaginatedRecordsResponse {
  data: SyncRecord[];
  meta: {
    total: number;
    page: number;
    page_size: number;
  };
}

export interface PaginatedMappingsResponse {
  data: ExternalMapping[];
  meta: {
    total: number;
    page: number;
    page_size: number;
  };
}
