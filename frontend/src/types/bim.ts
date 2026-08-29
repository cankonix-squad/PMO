export type BIMDiscipline =
  | "ARCHITECTURAL"
  | "STRUCTURAL"
  | "MEP"
  | "CIVIL"
  | "LANDSCAPE"
  | "OTHER";

export type BIMProvider =
  | "autodesk_bim360"
  | "trimble_connect"
  | "bentley_projectwise"
  | "local"
  | "other";

export type BIMModelStatus = "DRAFT" | "ACTIVE" | "ARCHIVED";

export type BIMModelRole = "PRIMARY" | "REFERENCE" | "ASBUILT" | "OTHER";

// ---------------------------------------------------------------------------
// BIM Model
// ---------------------------------------------------------------------------

export interface BIMModel {
  id: string;
  organization_id: string;
  name: string;
  description: string;
  discipline: BIMDiscipline;
  provider: BIMProvider;
  external_model_id: string;
  viewer_url: string;
  status: BIMModelStatus;
  metadata: Record<string, unknown>;
  created_by: string;
  deleted_at?: string;
  created_at: string;
  updated_at: string;
}

// ---------------------------------------------------------------------------
// BIM Model Version (immutable)
// ---------------------------------------------------------------------------

export interface BIMModelVersion {
  id: string;
  bim_model_id: string;
  organization_id: string;
  version_label: string;
  external_version_id: string;
  change_summary: string;
  file_size_bytes: number;
  checksum: string;
  published_at?: string;
  created_by: string;
  created_at: string;
}

// ---------------------------------------------------------------------------
// BIM Project Mapping
// ---------------------------------------------------------------------------

export interface BIMProjectMapping {
  id: string;
  organization_id: string;
  bim_model_id: string;
  project_id: string;
  model_role: BIMModelRole;
  notes: string;
  linked_by: string;
  linked_at: string;
}

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

export interface CreateBIMModelRequest {
  name: string;
  description?: string;
  discipline: BIMDiscipline;
  provider: BIMProvider;
  external_model_id: string;
  viewer_url?: string;
  metadata?: Record<string, unknown>;
}

export interface UpdateBIMModelRequest {
  name?: string;
  description?: string;
  discipline?: BIMDiscipline;
  viewer_url?: string;
  status?: BIMModelStatus;
  metadata?: Record<string, unknown>;
}

export interface CreateVersionRequest {
  version_label: string;
  external_version_id?: string;
  change_summary?: string;
  file_size_bytes?: number;
  checksum?: string;
}

export interface LinkProjectRequest {
  project_id: string;
  model_role: BIMModelRole;
  notes?: string;
}

// ---------------------------------------------------------------------------
// Response wrappers
// ---------------------------------------------------------------------------

export interface BIMModelListResponse {
  data: BIMModel[];
  meta: { total: number };
}

export interface BIMVersionListResponse {
  data: BIMModelVersion[];
  meta: { total: number };
}

export interface BIMMappingListResponse {
  data: BIMProjectMapping[];
  meta: { total: number };
}
