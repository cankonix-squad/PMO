// KPBU (PPP) monitoring — phase K1. See docs/planning/KPBU-GRAND-DESIGN.md.

export const KPBU_MAIN_STAGES = [
  "IDENTIFIED",
  "PLANNING",
  "PREPARATION_OBC",
  "PREPARATION_FBC",
  "TRANSACTION",
  "CONTRACT_SIGNED",
  "FINANCIAL_CLOSE",
  "CONSTRUCTION",
  "OPERATION",
  "HANDOVER",
] as const;

export type KpbuMainStage = (typeof KPBU_MAIN_STAGES)[number];
export type KpbuStage = KpbuMainStage | "CLOSED" | "ON_HOLD" | "CANCELLED" | "TERMINATED";

export const KPBU_STAGE_LABEL: Record<KpbuStage, string> = {
  IDENTIFIED: "Identifikasi",
  PLANNING: "Perencanaan",
  PREPARATION_OBC: "Penyiapan — OBC",
  PREPARATION_FBC: "Penyiapan — FBC",
  TRANSACTION: "Transaksi",
  CONTRACT_SIGNED: "Perjanjian Ditandatangani",
  FINANCIAL_CLOSE: "Financial Close",
  CONSTRUCTION: "Konstruksi",
  OPERATION: "Operasi",
  HANDOVER: "Serah Terima",
  CLOSED: "Selesai",
  ON_HOLD: "Ditunda",
  CANCELLED: "Dibatalkan",
  TERMINATED: "Diakhiri",
};

export type KpbuSector = "SDA" | "JALAN" | "JALAN_TOL" | "AIR_MINUM" | "SANITASI" | "PERUMAHAN" | "LAINNYA";
export const KPBU_SECTOR_LABEL: Record<KpbuSector, string> = {
  SDA: "Sumber Daya Air",
  JALAN: "Jalan (non-tol)",
  JALAN_TOL: "Jalan Tol",
  AIR_MINUM: "Air Minum",
  SANITASI: "Sanitasi",
  PERUMAHAN: "Perumahan",
  LAINNYA: "Lainnya",
};

export type KpbuInitiative = "SOLICITED" | "UNSOLICITED";
export const KPBU_INITIATIVE_LABEL: Record<KpbuInitiative, string> = {
  SOLICITED: "Prakarsa Pemerintah (Solicited)",
  UNSOLICITED: "Prakarsa Badan Usaha (Unsolicited)",
};

export type KpbuReturnScheme = "USER_CHARGE" | "AVAILABILITY_PAYMENT" | "HYBRID";
export const KPBU_RETURN_SCHEME_LABEL: Record<KpbuReturnScheme, string> = {
  USER_CHARGE: "User Charge (tarif pengguna)",
  AVAILABILITY_PAYMENT: "Availability Payment",
  HYBRID: "Hybrid",
};

export type KpbuPartyRole =
  | "PJPK"
  | "BUP"
  | "SPONSOR"
  | "LENDER"
  | "GUARANTOR"
  | "CONSULTANT"
  | "INDEPENDENT_ENGINEER"
  | "OTHER";
export const KPBU_PARTY_ROLE_LABEL: Record<KpbuPartyRole, string> = {
  PJPK: "PJPK",
  BUP: "Badan Usaha Pelaksana (BUP)",
  SPONSOR: "Sponsor",
  LENDER: "Lender",
  GUARANTOR: "Penjamin",
  CONSULTANT: "Konsultan",
  INDEPENDENT_ENGINEER: "Independent Engineer",
  OTHER: "Lainnya",
};

export interface KpbuProject {
  id: string;
  organization_id: string;
  project_id: string | null;
  code: string;
  name: string;
  description: string;
  sector: KpbuSector;
  stage: KpbuStage;
  previous_stage: KpbuStage | null;
  initiative: KpbuInitiative;
  return_scheme: KpbuReturnScheme;
  pjpk_name: string;
  pjpk_unit: string;
  investment_value: number;
  currency: string;
  concession_years: number | null;
  concession_start: string | null;
  concession_end: string | null;
  financial_close_target: string | null;
  financial_close_actual: string | null;
  cod_target: string | null;
  cod_actual: string | null;
  procurement_method: KpbuProcurementMethod | null;
  procurement_reason: string;
  agreement_signed_date: string | null;
  financial_close_deadline: string | null;
  financial_close_days_left: number | null;
  financial_close_status: KpbuFinancialCloseStatus;
  org_unit_id: string | null;
  org_unit_name?: string;
  province: string;
  latitude: number | null;
  longitude: number | null;
  created_at: string;
  updated_at: string;
  stage_delay_days: number;
  stage_is_delayed: boolean;
  stage_planned_end: string | null;
}

export interface KpbuStageMilestone {
  id: string;
  kpbu_project_id: string;
  stage: KpbuMainStage;
  planned_start: string | null;
  planned_end: string | null;
  actual_start: string | null;
  actual_end: string | null;
  notes: string;
  is_delayed: boolean;
  delay_days: number;
}

export interface KpbuParty {
  id: string;
  kpbu_project_id: string;
  role: KpbuPartyRole;
  name: string;
  vendor_id: string | null;
  share_pct: number | null;
  contact_person: string;
  email: string;
  phone: string;
  notes: string;
}

export interface KpbuProjectDetail {
  project: KpbuProject;
  stage_milestones: KpbuStageMilestone[];
  parties: KpbuParty[];
  allowed_next_stages: KpbuStage[];
}

export interface KpbuProjectRequest {
  code: string;
  name: string;
  description?: string;
  sector: KpbuSector;
  initiative?: KpbuInitiative;
  return_scheme?: KpbuReturnScheme;
  pjpk_name?: string;
  pjpk_unit?: string;
  investment_value: number;
  currency?: string;
  concession_years?: number | null;
  concession_start?: string | null;
  concession_end?: string | null;
  financial_close_target?: string | null;
  cod_target?: string | null;
  procurement_method?: KpbuProcurementMethod | null;
  procurement_reason?: string;
  agreement_signed_date?: string | null;
  org_unit_id?: string | null;
  province?: string;
  latitude?: number | null;
  longitude?: number | null;
}

export interface KpbuTransitionRequest {
  to_stage: KpbuStage;
  date?: string;
  notes?: string;
  override_document_check?: boolean;
  override_reason?: string;
}

export interface KpbuStageMilestoneRequest {
  stage: KpbuMainStage;
  planned_start?: string | null;
  planned_end?: string | null;
  actual_start?: string | null;
  actual_end?: string | null;
  notes?: string;
}

export interface KpbuPartyRequest {
  role: KpbuPartyRole;
  name: string;
  share_pct?: number | null;
  contact_person?: string;
  email?: string;
  phone?: string;
  notes?: string;
}

export interface KpbuListFilter {
  stage?: string;
  sector?: string;
  search?: string;
}

// ── K1.5: document checklist, guarantees, procurement ───────────────────────

export type KpbuProcurementMethod = "TENDER" | "PENUNJUKAN_LANGSUNG";
export const KPBU_PROCUREMENT_LABEL: Record<KpbuProcurementMethod, string> = {
  TENDER: "Tender (lelang)",
  PENUNJUKAN_LANGSUNG: "Penunjukan Langsung",
};

export type KpbuFinancialCloseStatus = "NONE" | "ON_TRACK" | "WARNING" | "OVERDUE" | "ACHIEVED";

export type KpbuDocCategory =
  | "PRAKARSA"
  | "PRAKUALIFIKASI"
  | "PENAWARAN"
  | "PENANDATANGANAN"
  | "PEMBIAYAAN_PELAKSANAAN";
export const KPBU_DOC_CATEGORY_LABEL: Record<KpbuDocCategory, string> = {
  PRAKARSA: "A. Pengajuan Prakarsa",
  PRAKUALIFIKASI: "B. Prakualifikasi",
  PENAWARAN: "C. Penawaran",
  PENANDATANGANAN: "D. Penetapan s.d. Penandatanganan",
  PEMBIAYAAN_PELAKSANAAN: "E. Pembiayaan & Pelaksanaan",
};

export type KpbuAppliesTo = "ALL" | "SOLICITED" | "UNSOLICITED";
export const KPBU_APPLIES_LABEL: Record<KpbuAppliesTo, string> = {
  ALL: "Semua prakarsa",
  SOLICITED: "Prakarsa Pemerintah",
  UNSOLICITED: "Prakarsa Badan Usaha",
};

export type KpbuSourceStatus = "VERIFIED" | "SECONDARY" | "UNVERIFIED";
export const KPBU_SOURCE_LABEL: Record<KpbuSourceStatus, string> = {
  VERIFIED: "Terverifikasi (teks regulasi)",
  SECONDARY: "Sumber sekunder",
  UNVERIFIED: "Belum terverifikasi",
};

export type KpbuDocStatus = "MISSING" | "SUBMITTED" | "APPROVED" | "REJECTED";
export const KPBU_DOC_STATUS_LABEL: Record<KpbuDocStatus, string> = {
  MISSING: "Belum ada",
  SUBMITTED: "Menunggu verifikasi",
  APPROVED: "Disetujui",
  REJECTED: "Ditolak",
};

export interface KpbuDocumentRequirement {
  id: string;
  code: string;
  category: KpbuDocCategory;
  name: string;
  description: string;
  applies_to: KpbuAppliesTo;
  required_before_stage: KpbuStage | null;
  is_mandatory: boolean;
  legal_basis: string;
  source_status: KpbuSourceStatus;
  sort_order: number;
  is_active: boolean;
}

export interface KpbuRequirementRequest {
  code: string;
  category: KpbuDocCategory;
  name: string;
  description?: string;
  applies_to: KpbuAppliesTo;
  required_before_stage: KpbuStage | null;
  is_mandatory: boolean;
  legal_basis?: string;
  source_status: KpbuSourceStatus;
  sort_order: number;
  is_active: boolean;
}

export interface KpbuDocument {
  id: string;
  kpbu_project_id: string;
  requirement_id: string | null;
  name: string;
  mime_type: string;
  file_size: number;
  document_number: string;
  document_date: string | null;
  notes: string;
  status: Exclude<KpbuDocStatus, "MISSING">;
  uploaded_by: string;
  uploaded_at: string;
  reviewed_by: string | null;
  reviewed_at: string | null;
  review_notes: string;
}

export interface KpbuChecklistItem {
  requirement: KpbuDocumentRequirement;
  status: KpbuDocStatus;
  documents: KpbuDocument[];
}

export interface KpbuChecklist {
  items: KpbuChecklistItem[];
  other_documents: KpbuDocument[];
  initiative: KpbuInitiative;
  summary: {
    mandatory_total: number;
    mandatory_approved: number;
    pending_review: number;
    blocking_next_stage: string[];
    next_stage: KpbuStage | "";
  };
}

export interface KpbuDocumentUpload {
  file: File;
  requirement_id?: string;
  name?: string;
  document_number?: string;
  document_date?: string;
  notes?: string;
}

export type KpbuGuaranteeType = "BID" | "PERFORMANCE" | "GOVERNMENT" | "OTHER";
export const KPBU_GUARANTEE_TYPE_LABEL: Record<KpbuGuaranteeType, string> = {
  BID: "Jaminan Penawaran",
  PERFORMANCE: "Jaminan Pelaksanaan",
  GOVERNMENT: "Penjaminan Pemerintah (PII)",
  OTHER: "Lainnya",
};

export type KpbuGuaranteeStatus = "ACTIVE" | "RELEASED" | "CLAIMED" | "EXPIRED";
export const KPBU_GUARANTEE_STATUS_LABEL: Record<KpbuGuaranteeStatus, string> = {
  ACTIVE: "Aktif",
  RELEASED: "Dikembalikan",
  CLAIMED: "Dicairkan",
  EXPIRED: "Kedaluwarsa",
};

export interface KpbuGuarantee {
  id: string;
  kpbu_project_id: string;
  type: KpbuGuaranteeType;
  issuer: string;
  guarantee_number: string;
  amount: number;
  currency: string;
  issued_date: string | null;
  expiry_date: string | null;
  status: KpbuGuaranteeStatus;
  notes: string;
  pct_of_investment: number | null;
  exceeds_limit: boolean;
  days_to_expiry: number | null;
}

export interface KpbuGuaranteeRequest {
  type: KpbuGuaranteeType;
  issuer?: string;
  guarantee_number?: string;
  amount: number;
  currency?: string;
  issued_date?: string | null;
  expiry_date?: string | null;
  status?: KpbuGuaranteeStatus;
  notes?: string;
}
