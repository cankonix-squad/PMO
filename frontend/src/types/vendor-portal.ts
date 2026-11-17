// Contractor portal (VENDOR-PORTAL-001).

export type SubmissionStatus = "SUBMITTED" | "VERIFIED" | "REJECTED";

export const SUBMISSION_STATUS_LABEL: Record<SubmissionStatus, string> = {
  SUBMITTED: "Menunggu Verifikasi",
  VERIFIED: "Terverifikasi",
  REJECTED: "Ditolak",
};

export type VendorDocCategory = "PROGRESS_REPORT" | "MC" | "BAST" | "PHOTO" | "SHOP_DRAWING" | "AS_BUILT" | "OTHER";

export const VENDOR_DOC_CATEGORY_LABEL: Record<VendorDocCategory, string> = {
  PROGRESS_REPORT: "Laporan Kemajuan",
  MC: "Berita Acara Kemajuan (MC)",
  BAST: "BAST",
  PHOTO: "Foto Lapangan",
  SHOP_DRAWING: "Shop Drawing",
  AS_BUILT: "As-Built Drawing",
  OTHER: "Lainnya",
};

export interface SubmissionContext {
  contract_number?: string;
  project_code?: string;
  project_name?: string;
  vendor_name?: string;
  ppk_user_id?: string;
  ppk_name?: string;
}

export interface VendorContract {
  id: string;
  contract_number: string;
  title: string;
  status: string;
  contract_value: number;
  currency: string;
  start_date: string | null;
  end_date: string | null;
  project_id: string;
  project_code: string;
  project_name: string;
  latest_verified_progress_pct: number | null;
  pending_reports: number;
  pending_documents: number;
  rejected_submissions: number;
}

export interface VendorProgressReport extends SubmissionContext {
  id: string;
  project_id: string;
  contract_id: string;
  vendor_id: string;
  period_year: number;
  period_month: number;
  physical_progress_pct: number;
  planned_progress_pct: number | null;
  notes: string;
  status: SubmissionStatus;
  submitted_at: string;
  reviewed_at: string | null;
  review_notes: string;
}

export interface VendorDocument extends SubmissionContext {
  id: string;
  project_id: string;
  contract_id: string;
  vendor_id: string;
  name: string;
  category: VendorDocCategory;
  period_year: number | null;
  period_month: number | null;
  mime_type: string;
  file_size: number;
  notes: string;
  status: SubmissionStatus;
  submitted_at: string;
  reviewed_at: string | null;
  review_notes: string;
}

export interface CreateProgressReportRequest {
  period_year: number;
  period_month: number;
  physical_progress_pct: number;
  planned_progress_pct?: number | null;
  notes?: string;
}

export interface UploadVendorDocumentInput {
  file: File;
  name?: string;
  category: VendorDocCategory;
  period_year?: number;
  period_month?: number;
  notes?: string;
}

export interface VendorSubmissions {
  progress_reports: VendorProgressReport[];
  documents: VendorDocument[];
}

export interface ReviewRequest {
  decision: "VERIFIED" | "REJECTED";
  notes?: string;
}

export interface VendorAccount {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  is_active: boolean;
  must_change_pwd: boolean;
  last_login_at: string | null;
  created_at: string;
}

export interface CreateVendorAccountRequest {
  first_name: string;
  last_name?: string;
  email: string;
  phone?: string;
}

export interface CreatedVendorAccount {
  account: VendorAccount;
  temporary_password: string;
}
