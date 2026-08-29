export type VerificationStatus = 'PENDING' | 'VERIFIED' | 'REJECTED';

export interface FieldEvidence {
  id: string;
  inspection_id: string;
  file_name: string;
  mime_type: string;
  file_size: number;
  checksum_sha256: string;
  captured_at?: string;
  latitude?: number;
  longitude?: number;
  verification_status: VerificationStatus;
}

export interface FieldInspection {
  id: string;
  project_id: string;
  inspected_at: string;
  latitude?: number;
  longitude?: number;
  inspector_id: string;
  notes?: string;
  verification_status: VerificationStatus;
  verified_at?: string;
  evidence?: FieldEvidence[];
}

export interface CreateFieldInspectionRequest {
  inspected_at: string;
  latitude?: number;
  longitude?: number;
  notes?: string;
  file?: File;
}