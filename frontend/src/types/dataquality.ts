export type ValidationStatus = 'DRAFT' | 'SUBMITTED' | 'VALID' | 'REJECTED' | 'STALE';

export interface ValidationSubmission {
  id: string;
  organization_id: string;
  project_id: string;
  snapshot_id: string;
  source?: string;
  source_reference?: string;
  period_year: number;
  period_month: number;
  status: ValidationStatus;
  completeness_pct: number;
  freshness_at?: string;
  freshness_days?: number;
  sla_due_at?: string;
  submitted_by?: string;
  submitted_at?: string;
  validator_id?: string;
  validated_at?: string;
  rejection_reason?: string;
  lineage: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface CreateValidationSubmissionRequest {
  snapshot_id: string;
  completeness_pct?: number;
  freshness_at?: string;
  sla_hours?: number;
  source_reference?: string;
  lineage?: Record<string, unknown>;
}

export interface TransitionValidationRequest {
  status: 'VALID' | 'REJECTED' | 'STALE';
  rejection_reason?: string;
}