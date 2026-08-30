// CANKORA-DASH-002: Periodic Progress & Financial Report types
// Classification: OPERATIONAL (laporan periodik operasional — belum official-governed)

export interface PeriodicReport {
  id: string;
  organization_id: string;
  project_id: string;
  period_year: number;
  period_month: number;
  physical_progress_pct: number;
  financial_planned: number;
  financial_actual: number;
  /** Backend-computed: financial_actual / financial_planned * 100; 0 if planned = 0 */
  financial_pct: number;
  notes: string | null;
  reported_by: string | null;
  reported_at: string;
  created_at: string;
  updated_at: string;
}

export interface CreatePeriodicReportRequest {
  period_year: number;
  period_month: number;
  physical_progress_pct: number;
  financial_planned: number;
  financial_actual: number;
  notes?: string;
  reported_at?: string;
}

export interface UpdatePeriodicReportRequest {
  physical_progress_pct?: number;
  financial_planned?: number;
  financial_actual?: number;
  notes?: string;
  reported_at?: string;
}

export interface PeriodicReportListFilter {
  year?: number;
  month?: number;
  page?: number;
  page_size?: number;
}

/** Computed display values for a periodic report row */
export function computeVariance(r: PeriodicReport): number {
  return r.financial_planned - r.financial_actual;
}

export const MONTH_LABELS: Record<number, string> = {
  1: "Jan", 2: "Feb", 3: "Mar", 4: "Apr", 5: "Mei", 6: "Jun",
  7: "Jul", 8: "Agu", 9: "Sep", 10: "Okt", 11: "Nov", 12: "Des",
};

export const MONTH_FULL: Record<number, string> = {
  1: "Januari", 2: "Februari", 3: "Maret", 4: "April",
  5: "Mei", 6: "Juni", 7: "Juli", 8: "Agustus",
  9: "September", 10: "Oktober", 11: "November", 12: "Desember",
};

export function formatPeriod(year: number, month: number): string {
  return `${MONTH_FULL[month] ?? month} ${year}`;
}

export function formatCurrency(value: number, currency = "IDR"): string {
  if (value >= 1_000_000_000_000) {
    return `${currency} ${(value / 1_000_000_000_000).toFixed(2)} T`;
  }
  if (value >= 1_000_000_000) {
    return `${currency} ${(value / 1_000_000_000).toFixed(2)} M`;
  }
  if (value >= 1_000_000) {
    return `${currency} ${(value / 1_000_000).toFixed(2)} Jt`;
  }
  return `${currency} ${value.toLocaleString("id-ID")}`;
}
