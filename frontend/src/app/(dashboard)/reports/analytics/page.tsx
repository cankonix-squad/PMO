"use client";

/**
 * Reporting Center — P2-009
 *
 * Analytics data mart: 6 dataset tabs, Power BI config status,
 * export request queue. Membaca dari /api/v1/analytics/reports/...
 */

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  BarChart2,
  BarChart3,
  BookOpen,
  CheckCircle2,
  Clock,
  Download,
  FileText,
  LayoutGrid,
  Loader2,
  ShieldAlert,
  TrendingUp,
  Wallet,
  XCircle,
  Zap,
} from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { reportingService } from "@/services/reporting.service";
import { cn } from "@/lib/utils";
import type {
  DatasetFilter,
  DatasetKey,
  ExportFormat,
  ReportExportRequest,
} from "@/types/reporting";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function formatIDR(n: number): string {
  if (n >= 1_000_000_000_000) return `Rp ${(n / 1_000_000_000_000).toFixed(1)}T`;
  if (n >= 1_000_000_000) return `Rp ${(n / 1_000_000_000).toFixed(1)}M`;
  if (n >= 1_000_000) return `Rp ${(n / 1_000_000).toFixed(1)}jt`;
  return `Rp ${Math.round(n).toLocaleString()}`;
}

function pct(n: number): string {
  return `${n.toFixed(1)}%`;
}

function healthBadge(h: string | null) {
  if (!h) return <span className="text-slate-400 text-xs">—</span>;
  const map: Record<string, string> = {
    GREEN: "bg-green-100 text-green-700",
    YELLOW: "bg-yellow-100 text-yellow-700",
    RED: "bg-red-100 text-red-700",
    CRITICAL: "bg-red-200 text-red-900 font-bold",
  };
  return (
    <span className={cn("rounded-full px-2 py-0.5 text-xs font-medium", map[h] ?? "bg-slate-100 text-slate-600")}>
      {h}
    </span>
  );
}

function statusBadge(s: string) {
  const map: Record<string, string> = {
    ACTIVE: "bg-blue-100 text-blue-700",
    COMPLETED: "bg-green-100 text-green-700",
    ON_HOLD: "bg-amber-100 text-amber-700",
    DRAFT: "bg-slate-100 text-slate-600",
    PLANNING: "bg-purple-100 text-purple-700",
    CANCELLED: "bg-red-100 text-red-600",
  };
  return (
    <span className={cn("rounded-full px-2 py-0.5 text-xs font-medium", map[s] ?? "bg-slate-100 text-slate-600")}>
      {s}
    </span>
  );
}

function exportStatusBadge(s: string) {
  const map: Record<string, string> = {
    PENDING: "bg-slate-100 text-slate-600",
    PROCESSING: "bg-blue-100 text-blue-700",
    COMPLETED: "bg-green-100 text-green-700",
    FAILED: "bg-red-100 text-red-700",
  };
  return (
    <span className={cn("rounded-full px-2 py-0.5 text-xs font-medium", map[s] ?? "bg-slate-100 text-slate-600")}>
      {s}
    </span>
  );
}

function priorityCategoryBadge(c: string | null) {
  if (!c) return <span className="text-slate-400 text-xs">—</span>;
  const map: Record<string, string> = {
    LOW: "bg-slate-100 text-slate-600",
    MEDIUM: "bg-blue-100 text-blue-700",
    HIGH: "bg-amber-100 text-amber-700",
    CRITICAL: "bg-red-100 text-red-700",
  };
  return (
    <span className={cn("rounded-full px-2 py-0.5 text-xs font-medium", map[c] ?? "bg-slate-100 text-slate-600")}>
      {c}
    </span>
  );
}

// ---------------------------------------------------------------------------
// Tab definitions
// ---------------------------------------------------------------------------

type TabId = "executive" | "performance" | "risk" | "budget" | "benefits" | "priority" | "powerbi" | "exports";

const TABS: { id: TabId; label: string; icon: React.ReactNode }[] = [
  { id: "executive",   label: "Ringkasan Eksekutif", icon: <LayoutGrid className="w-4 h-4" /> },
  { id: "performance", label: "Kinerja Proyek", icon: <TrendingUp className="w-4 h-4" /> },
  { id: "risk",        label: "Risiko & Isu", icon: <ShieldAlert className="w-4 h-4" /> },
  { id: "budget",      label: "Anggaran", icon: <Wallet className="w-4 h-4" /> },
  { id: "benefits",    label: "Manfaat", icon: <BarChart3 className="w-4 h-4" /> },
  { id: "priority",    label: "Prioritas", icon: <Zap className="w-4 h-4" /> },
  { id: "powerbi",     label: "Konfigurasi Power BI", icon: <BarChart2 className="w-4 h-4" /> },
  { id: "exports",     label: "Antrean Export", icon: <Download className="w-4 h-4" /> },
];

const DATASET_KEY_MAP: Partial<Record<TabId, DatasetKey>> = {
  executive:   "executive-summary",
  performance: "project-performance",
  risk:        "risk-issue",
  budget:      "budget",
  benefits:    "benefits",
  priority:    "priority",
};

// ---------------------------------------------------------------------------
// Filter bar
// ---------------------------------------------------------------------------

function FilterBar({
  filter,
  onChange,
}: {
  filter: DatasetFilter;
  onChange: (f: DatasetFilter) => void;
}) {
  return (
    <div className="flex flex-wrap gap-3 items-center mb-4">
      <div className="flex flex-col gap-0.5">
        <label className="text-xs text-slate-500">Status Proyek</label>
        <select
          className="border rounded px-2 py-1 text-sm bg-white dark:bg-slate-800 dark:border-slate-600"
          value={filter.status ?? ""}
          onChange={(e) => onChange({ ...filter, status: e.target.value || undefined })}
        >
          <option value="">Semua</option>
          <option value="ACTIVE">Aktif</option>
          <option value="COMPLETED">Selesai</option>
          <option value="ON_HOLD">Ditunda</option>
          <option value="PLANNING">Perencanaan</option>
          <option value="DRAFT">Draft</option>
          <option value="CANCELLED">Dibatalkan</option>
        </select>
      </div>
      <div className="flex flex-col gap-0.5">
        <label className="text-xs text-slate-500">Provinsi</label>
        <input
          type="text"
          placeholder="Filter provinsi..."
          className="border rounded px-2 py-1 text-sm bg-white dark:bg-slate-800 dark:border-slate-600 w-40"
          value={filter.province ?? ""}
          onChange={(e) => onChange({ ...filter, province: e.target.value || undefined })}
        />
      </div>
      <div className="flex flex-col gap-0.5">
        <label className="text-xs text-slate-500">Mulai ≥</label>
        <input
          type="date"
          className="border rounded px-2 py-1 text-sm bg-white dark:bg-slate-800 dark:border-slate-600"
          value={filter.period_start ?? ""}
          onChange={(e) => onChange({ ...filter, period_start: e.target.value || undefined })}
        />
      </div>
      <div className="flex flex-col gap-0.5">
        <label className="text-xs text-slate-500">Selesai ≤</label>
        <input
          type="date"
          className="border rounded px-2 py-1 text-sm bg-white dark:bg-slate-800 dark:border-slate-600"
          value={filter.period_end ?? ""}
          onChange={(e) => onChange({ ...filter, period_end: e.target.value || undefined })}
        />
      </div>
      <button
        className="mt-4 px-3 py-1 text-sm rounded bg-slate-100 hover:bg-slate-200 dark:bg-slate-700 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-200"
        onClick={() => onChange({})}
      >
        Reset
      </button>
    </div>
  );
}

// ---------------------------------------------------------------------------
// KPI Card
// ---------------------------------------------------------------------------

function KpiCard({ label, value, sub, color }: { label: string; value: string | number; sub?: string; color?: string }) {
  return (
    <div className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl p-4 flex flex-col gap-1 shadow-sm">
      <span className="text-xs text-slate-500 dark:text-slate-400">{label}</span>
      <span className={cn("text-2xl font-bold", color ?? "text-slate-900 dark:text-slate-100")}>{value}</span>
      {sub && <span className="text-xs text-slate-400">{sub}</span>}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Empty / Loading / Error states
// ---------------------------------------------------------------------------

function LoadingState() {
  return (
    <div className="flex items-center justify-center h-40 text-slate-400 gap-2">
      <Loader2 className="w-5 h-5 animate-spin" />
      <span>Memuat data...</span>
    </div>
  );
}

function EmptyState({ msg }: { msg?: string }) {
  return (
    <div className="flex flex-col items-center justify-center h-40 text-slate-400 gap-2">
      <FileText className="w-8 h-8 opacity-30" />
      <span className="text-sm">{msg ?? "Tidak ada data"}</span>
    </div>
  );
}

function ErrorState({ msg }: { msg?: string }) {
  return (
    <div className="flex items-center justify-center h-40 text-red-400 gap-2">
      <XCircle className="w-5 h-5" />
      <span className="text-sm">{msg ?? "Gagal memuat data"}</span>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Export request form
// ---------------------------------------------------------------------------

function ExportRequestForm({
  activeTab,
  onSuccess,
}: {
  activeTab: TabId;
  onSuccess: () => void;
}) {
  const [format, setFormat] = useState<ExportFormat>("XLSX");
  const qc = useQueryClient();

  const mutation = useMutation({
    mutationFn: () =>
      reportingService.createExportRequest({
        dataset_key: DATASET_KEY_MAP[activeTab] ?? activeTab,
        format,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["reporting-exports"] });
      onSuccess();
    },
  });

  const datasetKey = DATASET_KEY_MAP[activeTab];
  if (!datasetKey) return null;

  return (
    <div className="flex items-center gap-2 mt-2">
      <select
        className="border rounded px-2 py-1 text-sm bg-white dark:bg-slate-800 dark:border-slate-600"
        value={format}
        onChange={(e) => setFormat(e.target.value as ExportFormat)}
      >
        <option value="XLSX">XLSX</option>
        <option value="CSV">CSV</option>
        <option value="PDF">PDF</option>
      </select>
      <button
        className="flex items-center gap-1 px-3 py-1.5 text-sm rounded bg-blue-600 hover:bg-blue-700 text-white disabled:opacity-50"
        onClick={() => mutation.mutate()}
        disabled={mutation.isPending}
      >
        {mutation.isPending ? (
          <Loader2 className="w-4 h-4 animate-spin" />
        ) : (
          <Download className="w-4 h-4" />
        )}
        Request Export
      </button>
      {mutation.isSuccess && (
        <span className="text-xs text-green-600 flex items-center gap-1">
          <CheckCircle2 className="w-3 h-3" /> Permintaan dikirim
        </span>
      )}
      {mutation.isError && (
        <span className="text-xs text-red-500">Gagal membuat export request</span>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Tab panels
// ---------------------------------------------------------------------------

function ExecutiveTab({ filter }: { filter: DatasetFilter }) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["reporting-executive", filter],
    queryFn: () => reportingService.getExecutiveSummary(filter),
  });

  if (isLoading) return <LoadingState />;
  if (isError || !data) return <ErrorState />;

  return (
    <div className="space-y-6">
      {/* Projects KPIs */}
      <div>
        <h3 className="text-sm font-semibold text-slate-600 dark:text-slate-400 mb-3">Proyek</h3>
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
          <KpiCard label="Total Proyek" value={data.total_projects} />
          <KpiCard label="Aktif" value={data.active_projects} color="text-blue-600" />
          <KpiCard label="Selesai" value={data.completed_projects} color="text-green-600" />
          <KpiCard label="On Hold" value={data.on_hold_projects} color="text-amber-600" />
          <KpiCard label="Rata-rata Progress" value={pct(data.avg_progress_pct)} color="text-purple-600" />
        </div>
      </div>

      {/* Budget KPIs */}
      <div>
        <h3 className="text-sm font-semibold text-slate-600 dark:text-slate-400 mb-3">Anggaran</h3>
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
          <KpiCard label="Total Rencana" value={formatIDR(data.total_budget_plan)} />
          <KpiCard label="Total Realisasi" value={formatIDR(data.total_budget_actual)} />
          <KpiCard
            label="Serapan Anggaran"
            value={pct(data.budget_usage_pct)}
            color={data.budget_usage_pct > 100 ? "text-red-600" : "text-green-600"}
          />
        </div>
      </div>

      {/* Risk & Issue KPIs */}
      <div>
        <h3 className="text-sm font-semibold text-slate-600 dark:text-slate-400 mb-3">Risiko & Isu</h3>
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
          <KpiCard label="Total Risiko" value={data.total_risks} />
          <KpiCard label="Risiko Terbuka" value={data.open_risks} color="text-amber-600" />
          <KpiCard label="Risiko Tinggi" value={data.high_risks} color="text-red-600" />
          <KpiCard label="Total Isu" value={data.total_issues} />
          <KpiCard label="Isu Terbuka" value={data.open_issues} color="text-amber-600" />
        </div>
      </div>

      {/* Health Distribution */}
      <div>
        <h3 className="text-sm font-semibold text-slate-600 dark:text-slate-400 mb-3">Distribusi Kesehatan Proyek</h3>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <KpiCard label="🟢 Green" value={data.green_health} color="text-green-600" />
          <KpiCard label="🟡 Yellow" value={data.yellow_health} color="text-yellow-600" />
          <KpiCard label="🔴 Red" value={data.red_health} color="text-red-600" />
          <KpiCard label="⚫ Critical" value={data.critical_health} color="text-red-900" />
        </div>
      </div>
    </div>
  );
}

function PerformanceTab({ filter }: { filter: DatasetFilter }) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["reporting-performance", filter],
    queryFn: () => reportingService.getProjectPerformance(filter),
  });

  if (isLoading) return <LoadingState />;
  if (isError) return <ErrorState />;
  if (!data || data.length === 0) return <EmptyState msg="Tidak ada data proyek" />;

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm border-collapse">
        <thead>
          <tr className="bg-slate-50 dark:bg-slate-800 text-left text-xs text-slate-500 uppercase tracking-wide">
            <th className="px-3 py-2 border-b dark:border-slate-700">Kode</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Nama</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Status</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Progress</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Rencana</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Realisasi</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Serapan</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Kesehatan</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Prioritas</th>
          </tr>
        </thead>
        <tbody>
          {data.map((row) => (
            <tr key={row.project_id} className="border-b dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800/50">
              <td className="px-3 py-2 font-mono text-xs text-slate-500">{row.project_code}</td>
              <td className="px-3 py-2 font-medium max-w-[200px] truncate">{row.project_name}</td>
              <td className="px-3 py-2">{statusBadge(row.status)}</td>
              <td className="px-3 py-2">
                <div className="flex items-center gap-2">
                  <div className="w-20 h-1.5 bg-slate-200 dark:bg-slate-600 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-blue-500 rounded-full"
                      style={{ width: `${Math.min(row.progress_pct, 100)}%` }}
                    />
                  </div>
                  <span className="text-xs text-slate-500">{pct(row.progress_pct)}</span>
                </div>
              </td>
              <td className="px-3 py-2 text-right tabular-nums text-xs">{formatIDR(row.budget_plan)}</td>
              <td className="px-3 py-2 text-right tabular-nums text-xs">{formatIDR(row.budget_actual)}</td>
              <td className="px-3 py-2 text-right tabular-nums text-xs">
                <span className={row.budget_usage_pct > 100 ? "text-red-600 font-medium" : "text-slate-700 dark:text-slate-300"}>
                  {pct(row.budget_usage_pct)}
                </span>
              </td>
              <td className="px-3 py-2">{healthBadge(row.health_class)}</td>
              <td className="px-3 py-2">{priorityCategoryBadge(row.priority_category)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function RiskTab({ filter }: { filter: DatasetFilter }) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["reporting-risk", filter],
    queryFn: () => reportingService.getRiskIssue(filter),
  });

  if (isLoading) return <LoadingState />;
  if (isError) return <ErrorState />;
  if (!data || data.length === 0) return <EmptyState msg="Tidak ada data risiko/isu" />;

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm border-collapse">
        <thead>
          <tr className="bg-slate-50 dark:bg-slate-800 text-left text-xs text-slate-500 uppercase tracking-wide">
            <th className="px-3 py-2 border-b dark:border-slate-700">Proyek</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-center">Total Risiko</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-center">Terbuka</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-center">Tinggi</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-center">Kritis</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-center">Total Isu</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-center">Terbuka</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-center">Tinggi</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-center">Kritis</th>
          </tr>
        </thead>
        <tbody>
          {data.map((row) => (
            <tr key={row.project_id} className="border-b dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800/50">
              <td className="px-3 py-2">
                <div className="font-medium text-sm">{row.project_name}</div>
                <div className="text-xs text-slate-400 font-mono">{row.project_code}</div>
              </td>
              <td className="px-3 py-2 text-center">{row.total_risks}</td>
              <td className="px-3 py-2 text-center">
                <span className={row.open_risks > 0 ? "text-amber-600 font-medium" : ""}>{row.open_risks}</span>
              </td>
              <td className="px-3 py-2 text-center">
                <span className={row.high_risks > 0 ? "text-red-500 font-medium" : ""}>{row.high_risks}</span>
              </td>
              <td className="px-3 py-2 text-center">
                <span className={row.critical_risks > 0 ? "text-red-700 font-bold" : ""}>{row.critical_risks}</span>
              </td>
              <td className="px-3 py-2 text-center">{row.total_issues}</td>
              <td className="px-3 py-2 text-center">
                <span className={row.open_issues > 0 ? "text-amber-600 font-medium" : ""}>{row.open_issues}</span>
              </td>
              <td className="px-3 py-2 text-center">
                <span className={row.high_issues > 0 ? "text-red-500 font-medium" : ""}>{row.high_issues}</span>
              </td>
              <td className="px-3 py-2 text-center">
                <span className={row.critical_issues > 0 ? "text-red-700 font-bold" : ""}>{row.critical_issues}</span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function BudgetTab({ filter }: { filter: DatasetFilter }) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["reporting-budget", filter],
    queryFn: () => reportingService.getBudget(filter),
  });

  if (isLoading) return <LoadingState />;
  if (isError) return <ErrorState />;
  if (!data || data.length === 0) return <EmptyState msg="Tidak ada data anggaran" />;

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm border-collapse">
        <thead>
          <tr className="bg-slate-50 dark:bg-slate-800 text-left text-xs text-slate-500 uppercase tracking-wide">
            <th className="px-3 py-2 border-b dark:border-slate-700">Proyek</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Status</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-right">Rencana</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-right">Realisasi</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-right">Varians</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-right">Serapan</th>
          </tr>
        </thead>
        <tbody>
          {data.map((row) => (
            <tr key={row.project_id} className="border-b dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800/50">
              <td className="px-3 py-2">
                <div className="font-medium text-sm">{row.project_name}</div>
                <div className="text-xs text-slate-400 font-mono">{row.project_code}</div>
              </td>
              <td className="px-3 py-2">{statusBadge(row.status)}</td>
              <td className="px-3 py-2 text-right tabular-nums text-xs">{formatIDR(row.budget_plan)}</td>
              <td className="px-3 py-2 text-right tabular-nums text-xs">{formatIDR(row.budget_actual)}</td>
              <td className="px-3 py-2 text-right tabular-nums text-xs">
                <span className={row.variance > 0 ? "text-red-600" : row.variance < 0 ? "text-green-600" : ""}>
                  {row.variance > 0 ? "+" : ""}{formatIDR(row.variance)}
                </span>
              </td>
              <td className="px-3 py-2 text-right tabular-nums text-xs">
                <span className={row.usage_pct > 100 ? "text-red-600 font-medium" : "text-slate-700 dark:text-slate-300"}>
                  {pct(row.usage_pct)}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function BenefitsTab({ filter }: { filter: DatasetFilter }) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["reporting-benefits", filter],
    queryFn: () => reportingService.getBenefits(filter),
  });

  if (isLoading) return <LoadingState />;
  if (isError) return <ErrorState />;
  if (!data || data.length === 0) return <EmptyState msg="Tidak ada data manfaat" />;

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm border-collapse">
        <thead>
          <tr className="bg-slate-50 dark:bg-slate-800 text-left text-xs text-slate-500 uppercase tracking-wide">
            <th className="px-3 py-2 border-b dark:border-slate-700">Proyek</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Indikator</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Satuan</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-right">Target</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-right">Aktual</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-right">Capaian</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Metode</th>
          </tr>
        </thead>
        <tbody>
          {data.map((row) => (
            <tr key={`${row.project_id}-${row.indicator_id}`} className="border-b dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800/50">
              <td className="px-3 py-2">
                <div className="font-medium text-sm">{row.project_name}</div>
                <div className="text-xs text-slate-400 font-mono">{row.project_code}</div>
              </td>
              <td className="px-3 py-2 text-sm">{row.indicator_name}</td>
              <td className="px-3 py-2 text-xs text-slate-500">{row.unit}</td>
              <td className="px-3 py-2 text-right tabular-nums text-sm">{row.target.toLocaleString()}</td>
              <td className="px-3 py-2 text-right tabular-nums text-sm">{row.actual.toLocaleString()}</td>
              <td className="px-3 py-2 text-right tabular-nums text-sm">
                <span className={row.achievement_pct >= 100 ? "text-green-600 font-medium" : row.achievement_pct < 50 ? "text-red-500" : "text-amber-600"}>
                  {pct(row.achievement_pct)}
                </span>
              </td>
              <td className="px-3 py-2 text-xs text-slate-500">{row.aggregation_method}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function PriorityTab() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["reporting-priority"],
    queryFn: () => reportingService.getPriority(),
  });

  if (isLoading) return <LoadingState />;
  if (isError) return <ErrorState />;
  if (!data || data.length === 0) return <EmptyState msg="Belum ada scoring prioritas" />;

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm border-collapse">
        <thead>
          <tr className="bg-slate-50 dark:bg-slate-800 text-left text-xs text-slate-500 uppercase tracking-wide">
            <th className="px-3 py-2 border-b dark:border-slate-700">Rank</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Proyek</th>
            <th className="px-3 py-2 border-b dark:border-slate-700 text-right">Skor</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Kategori</th>
            <th className="px-3 py-2 border-b dark:border-slate-700">Dihitung</th>
          </tr>
        </thead>
        <tbody>
          {data.map((row, i) => (
            <tr key={row.project_id} className="border-b dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800/50">
              <td className="px-3 py-2 text-slate-400 font-mono text-xs">{i + 1}</td>
              <td className="px-3 py-2">
                <div className="font-medium text-sm">{row.project_name}</div>
                <div className="text-xs text-slate-400 font-mono">{row.project_code}</div>
              </td>
              <td className="px-3 py-2 text-right font-bold tabular-nums">{row.total_score.toFixed(1)}</td>
              <td className="px-3 py-2">{priorityCategoryBadge(row.score_category)}</td>
              <td className="px-3 py-2 text-xs text-slate-400">
                {new Date(row.calculated_at).toLocaleDateString("id-ID")}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function PowerBITab() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["reporting-powerbi"],
    queryFn: () => reportingService.getPowerBIConfig(),
  });

  if (isLoading) return <LoadingState />;
  if (isError) return <ErrorState msg="Gagal memuat konfigurasi Power BI" />;
  if (!data) return <EmptyState />;

  return (
    <div className="max-w-lg space-y-4">
      <div className={cn(
        "flex items-center gap-3 rounded-xl p-4 border",
        data.configured
          ? "border-green-200 bg-green-50 dark:bg-green-950/30 dark:border-green-800"
          : "border-amber-200 bg-amber-50 dark:bg-amber-950/30 dark:border-amber-800"
      )}>
        {data.configured ? (
          <CheckCircle2 className="w-5 h-5 text-green-600 flex-shrink-0" />
        ) : (
          <Clock className="w-5 h-5 text-amber-600 flex-shrink-0" />
        )}
        <div>
          <p className={cn("font-semibold text-sm", data.configured ? "text-green-800 dark:text-green-300" : "text-amber-800 dark:text-amber-300")}>
            {data.configured ? "Power BI Terkonfigurasi" : "Power BI Belum Dikonfigurasi"}
          </p>
          <p className="text-xs text-slate-500 mt-0.5">
            {data.configured
              ? "Embed integration siap digunakan"
              : "Set environment variables POWERBI_WORKSPACE_ID, POWERBI_REPORT_ID, POWERBI_TENANT_ID di backend"}
          </p>
        </div>
      </div>

      {data.configured && (
        <div className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl p-4 space-y-3">
          <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300">Detail Konfigurasi</h3>
          {[
            { label: "Workspace ID", value: data.workspace_id },
            { label: "Report ID", value: data.report_id },
            { label: "Tenant ID", value: data.tenant_id },
            { label: "Auth Method", value: data.auth_method },
            { label: "Embed URL", value: data.embed_url },
          ].filter(x => x.value).map(({ label, value }) => (
            <div key={label} className="flex gap-3 text-sm">
              <span className="text-slate-500 w-32 flex-shrink-0">{label}</span>
              <span className="font-mono text-xs text-slate-700 dark:text-slate-300 break-all">{value}</span>
            </div>
          ))}
        </div>
      )}

      {!data.configured && (
        <div className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl p-4 space-y-2">
          <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300">Cara Konfigurasi</h3>
          <p className="text-xs text-slate-500">Tambahkan variabel berikut ke file <code className="bg-slate-100 dark:bg-slate-700 px-1 rounded">.env</code> backend:</p>
          <pre className="text-xs bg-slate-50 dark:bg-slate-900 rounded p-3 text-slate-700 dark:text-slate-300 overflow-x-auto">
{`POWERBI_WORKSPACE_ID=<your-workspace-id>
POWERBI_REPORT_ID=<your-report-id>
POWERBI_TENANT_ID=<your-tenant-id>
POWERBI_EMBED_URL=<embed-url>        # opsional
POWERBI_AUTH_METHOD=service_principal`}
          </pre>
        </div>
      )}
    </div>
  );
}

function ExportsTab() {
  const qc = useQueryClient();
  const [downloading, setDownloading] = useState<string | null>(null);

  const { data, isLoading, isError } = useQuery({
    queryKey: ["reporting-exports"],
    queryFn: () => reportingService.listExportRequests(),
    refetchInterval: 8_000, // poll setiap 8 detik untuk status update
  });

  if (isLoading) return <LoadingState />;
  if (isError) return <ErrorState />;
  if (!data || data.length === 0) return (
    <EmptyState msg="Belum ada export request. Gunakan tombol 'Request Export' di tab dataset." />
  );

  async function handleDownload(req: ReportExportRequest) {
    if (!req.file_name) return;
    setDownloading(req.id);
    try {
      await reportingService.downloadExportFile(req.id, req.file_name);
    } finally {
      setDownloading(null);
    }
  }

  function formatBytes(bytes: number | null): string {
    if (!bytes) return "—";
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <p className="text-xs text-slate-400 italic">
          Data bersumber dari reporting/read-model · Bukan write-back ke data operasional
        </p>
        <button
          className="text-xs text-slate-500 hover:text-blue-600 flex items-center gap-1"
          onClick={() => qc.invalidateQueries({ queryKey: ["reporting-exports"] })}
        >
          <Loader2 className="w-3 h-3" /> Refresh
        </button>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm border-collapse">
          <thead>
            <tr className="bg-slate-50 dark:bg-slate-800 text-left text-xs text-slate-500 uppercase tracking-wide">
              <th className="px-3 py-2 border-b dark:border-slate-700">Dataset</th>
              <th className="px-3 py-2 border-b dark:border-slate-700">Format</th>
              <th className="px-3 py-2 border-b dark:border-slate-700">Status</th>
              <th className="px-3 py-2 border-b dark:border-slate-700">Nama File</th>
              <th className="px-3 py-2 border-b dark:border-slate-700 text-right">Ukuran</th>
              <th className="px-3 py-2 border-b dark:border-slate-700">Dibuat</th>
              <th className="px-3 py-2 border-b dark:border-slate-700">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {data.map((req) => (
              <tr key={req.id} className="border-b dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800/50">
                <td className="px-3 py-2 font-mono text-xs">{req.dataset_key}</td>
                <td className="px-3 py-2 text-xs font-medium">{req.format}</td>
                <td className="px-3 py-2">{exportStatusBadge(req.status)}</td>
                <td className="px-3 py-2 text-xs text-slate-500 font-mono truncate max-w-[180px]">
                  {req.file_name ?? "—"}
                </td>
                <td className="px-3 py-2 text-xs text-slate-400 text-right tabular-nums">
                  {formatBytes(req.file_size)}
                </td>
                <td className="px-3 py-2 text-xs text-slate-400">
                  {req.generated_at
                    ? new Date(req.generated_at).toLocaleString("id-ID")
                    : new Date(req.created_at).toLocaleString("id-ID")}
                </td>
                <td className="px-3 py-2">
                  {req.status === "COMPLETED" && req.file_name ? (
                    <button
                      onClick={() => handleDownload(req)}
                      disabled={downloading === req.id}
                      className="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-800 disabled:opacity-50"
                    >
                      {downloading === req.id ? (
                        <Loader2 className="w-3 h-3 animate-spin" />
                      ) : (
                        <Download className="w-3 h-3" />
                      )}
                      Unduh
                    </button>
                  ) : req.status === "FAILED" && req.error_message ? (
                    <span className="text-xs text-red-500 flex items-center gap-1" title={req.error_message}>
                      <XCircle className="w-3 h-3" /> Error
                    </span>
                  ) : req.status === "PENDING" || req.status === "PROCESSING" ? (
                    <span className="text-xs text-slate-400 flex items-center gap-1">
                      <Clock className="w-3 h-3" /> Diproses
                    </span>
                  ) : (
                    <span className="text-xs text-slate-400">—</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main page
// ---------------------------------------------------------------------------

export default function ReportingAnalyticsPage() {
  const [activeTab, setActiveTab] = useState<TabId>("executive");
  const [filter, setFilter] = useState<DatasetFilter>({});
  const [showExportForm, setShowExportForm] = useState(false);

  const showFilter = ["executive", "performance", "risk", "budget", "benefits"].includes(activeTab);
  const showExportBtn = Object.keys(DATASET_KEY_MAP).includes(activeTab);

  return (
    <DashboardLayout title="Reporting Center">
      <div className="max-w-7xl mx-auto px-4 py-6 space-y-6">
        {/* Header */}
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <BookOpen className="w-5 h-5 text-blue-600" />
              <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Reporting Center</h1>
            </div>
            <p className="text-sm text-slate-500">
              Data mart analitik — 6 dataset siap ekspor ke Power BI atau format lainnya
            </p>
          </div>
        </div>

        {/* Tab navigation — horizontally scrollable on mobile */}
        <div className="overflow-x-auto -mx-4 px-4">
          <div className="flex gap-1 border-b border-slate-200 dark:border-slate-700 min-w-max">
            {TABS.map((tab) => (
              <button
                key={tab.id}
                onClick={() => {
                  setActiveTab(tab.id);
                  setShowExportForm(false);
                }}
                className={cn(
                  "flex items-center gap-1.5 px-3 py-2 text-sm rounded-t-lg border-b-2 transition-colors whitespace-nowrap",
                  activeTab === tab.id
                    ? "border-blue-600 text-blue-600 font-medium bg-blue-50/50 dark:bg-blue-950/20"
                    : "border-transparent text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 hover:border-slate-300"
                )}
              >
                {tab.icon}
                <span>{tab.label}</span>
              </button>
            ))}
          </div>
        </div>

        {/* Filter bar */}
        {showFilter && (
          <FilterBar filter={filter} onChange={setFilter} />
        )}

        {/* Export request form toggle */}
        {showExportBtn && (
          <div className="flex items-center gap-2">
            <button
              className="flex items-center gap-1.5 text-sm text-slate-600 dark:text-slate-400 hover:text-blue-600 underline underline-offset-2"
              onClick={() => setShowExportForm((v) => !v)}
            >
              <Download className="w-4 h-4" />
              {showExportForm ? "Sembunyikan form ekspor" : "Request ekspor dataset ini"}
            </button>
          </div>
        )}
        {showExportBtn && showExportForm && (
          <ExportRequestForm
            activeTab={activeTab}
            onSuccess={() => setShowExportForm(false)}
          />
        )}

        {/* Tab content */}
        <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-xl p-4 shadow-sm min-h-[300px]">
          {activeTab === "executive"   && <ExecutiveTab filter={filter} />}
          {activeTab === "performance" && <PerformanceTab filter={filter} />}
          {activeTab === "risk"        && <RiskTab filter={filter} />}
          {activeTab === "budget"      && <BudgetTab filter={filter} />}
          {activeTab === "benefits"    && <BenefitsTab filter={filter} />}
          {activeTab === "priority"    && <PriorityTab />}
          {activeTab === "powerbi"     && <PowerBITab />}
          {activeTab === "exports"     && <ExportsTab />}
        </div>

        <p className="text-xs text-slate-400 text-center">
          Data bersumber dari database PMO secara real-time · Export request diproses secara asinkron
        </p>
      </div>
    </DashboardLayout>
  );
}
