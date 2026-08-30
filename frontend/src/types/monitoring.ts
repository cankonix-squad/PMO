export interface Baseline {
  id: string;
  organization_id: string;
  project_id: string;
  version: number;
  label?: string;
  approved_at?: string;
  approved_by?: string;
  physical_target: number;
  budget_total: number;
  currency: string;
  planned_start: string;
  planned_end: string;
  source?: string;
  notes?: string;
  is_active: boolean;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateBaselineRequest {
  version?: number;
  label?: string;
  physical_target: number;
  budget_total: number;
  currency?: string;
  planned_start: string;
  planned_end: string;
  source?: string;
  notes?: string;
}

export interface UpdateBaselineRequest {
  label?: string;
  physical_target?: number;
  budget_total?: number;
  planned_start?: string;
  planned_end?: string;
  source?: string;
  notes?: string;
  is_active?: boolean;
}

export type SnapshotStatus = 'DRAFT' | 'SUBMITTED' | 'VALID' | 'REJECTED' | 'STALE';

export interface Snapshot {
  id: string;
  organization_id: string;
  project_id: string;
  baseline_id?: string;
  period_year: number;
  period_month: number;
  period_label?: string;
  physical_actual: number;
  physical_target: number;
  physical_variance: number;
  financial_actual: number;
  financial_target: number;
  financial_variance: number;
  currency: string;
  schedule_actual_start?: string;
  schedule_actual_end?: string;
  schedule_deviation_days?: number;
  status: SnapshotStatus;
  submitted_at?: string;
  submitted_by?: string;
  validated_at?: string;
  validated_by?: string;
  rejection_reason?: string;
  source?: string;
  notes?: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateSnapshotRequest {
  baseline_id?: string;
  period_year: number;
  period_month: number;
  physical_actual: number;
  physical_target: number;
  financial_actual: number;
  financial_target: number;
  currency?: string;
  schedule_deviation_days?: number;
  source?: string;
  notes?: string;
}

export interface UpdateSnapshotRequest {
  physical_actual?: number;
  physical_target?: number;
  financial_actual?: number;
  financial_target?: number;
  schedule_deviation_days?: number;
  source?: string;
  notes?: string;
}

export interface TransitionSnapshotRequest {
  status: SnapshotStatus;
  rejection_reason?: string;
}

export const SNAPSHOT_STATUS_LABEL: Record<SnapshotStatus, string> = {
  DRAFT: 'Draft',
  SUBMITTED: 'Diajukan',
  VALID: 'Valid',
  REJECTED: 'Ditolak',
  STALE: 'Kedaluwarsa',
};

export const SNAPSHOT_STATUS_COLOR: Record<SnapshotStatus, string> = {
  DRAFT: 'bg-muted text-muted-foreground',
  SUBMITTED: 'bg-blue-100 text-blue-700',
  VALID: 'bg-green-100 text-green-700',
  REJECTED: 'bg-red-100 text-red-700',
  STALE: 'bg-yellow-100 text-yellow-700',
};

export const MONTH_NAMES = ['', 'Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des'];
