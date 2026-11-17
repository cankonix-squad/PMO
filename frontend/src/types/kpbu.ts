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
  org_unit_id?: string | null;
  province?: string;
  latitude?: number | null;
  longitude?: number | null;
}

export interface KpbuTransitionRequest {
  to_stage: KpbuStage;
  date?: string;
  notes?: string;
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
