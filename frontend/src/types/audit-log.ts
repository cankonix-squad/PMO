// Audit Log Viewer — Type definitions
// Mirrors backend/internal/core/audit/model.go

export interface AuditLog {
  id: string;
  organization_id: string;
  actor_id?: string | null;
  actor_email?: string;
  action: string;
  entity_type: string;
  entity_id?: string;
  entity_label?: string;
  old_values?: string | null; // raw JSON string from JSONB
  new_values?: string | null; // raw JSON string from JSONB
  ip_address?: string;
  user_agent?: string;
  request_id?: string;
  created_at: string;
}

export interface AuditLogSummary {
  total_events: number;
  unique_actors: number;
  top_actions: ActionCount[];
  top_entities: EntityCount[];
}

export interface ActionCount {
  action: string;
  count: number;
}

export interface EntityCount {
  entity_type: string;
  count: number;
}

export interface AuditLogFilter {
  action?: string;
  entity_type?: string;
  entity_id?: string;
  actor_id?: string;
  search?: string;
  date_from?: string; // ISO date string or RFC3339
  date_to?: string;
  page?: number;
  page_size?: number;
}
