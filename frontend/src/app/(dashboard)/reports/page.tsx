"use client";

/**
 * Reports page — P1-007
 *
 * Lists periodic report snapshots (weekly/monthly/quarterly).
 * Supports generating new snapshots with live metrics, viewing detail,
 * publishing/archiving, and soft-delete.
 */

import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  BarChart2,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  FileText,
  Plus,
  RefreshCw,
  Trash2,
} from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { reportService } from "@/services/report.service";
import { cn, formatDate } from "@/lib/utils";
import type {
  ReportSnapshot,
  ReportPeriodType,
  ReportStatus,
  GenerateReportRequest,
} from "@/types/report";
import type { PaginationMeta } from "@/types/api";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function periodLabel(type: ReportPeriodType): string {
  switch (type) {
    case "WEEKLY":    return "Weekly";
    case "MONTHLY":   return "Monthly";
    case "QUARTERLY": return "Quarterly";
  }
}

function statusBadge(status: ReportStatus) {
  const map: Record<ReportStatus, string> = {
    DRAFT:     "bg-slate-100 text-slate-700 dark:bg-slate-700 dark:text-slate-200",
    PUBLISHED: "bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-200",
    ARCHIVED:  "bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-200",
  };
  return (
    <span className={cn("rounded-full px-2 py-0.5 text-xs font-medium", map[status])}>
      {status}
    </span>
  );
}

function formatIDR(n: number): string {
  if (n >= 1_000_000_000) return `Rp ${(n / 1_000_000_000).toFixed(1)}M`;
  if (n >= 1_000_000)     return `Rp ${(n / 1_000_000).toFixed(1)}jt`;
  if (n >= 1_000)         return `Rp ${(n / 1_000).toFixed(0)}rb`;
  return `Rp ${n}`;
}

// ---------------------------------------------------------------------------
// Generate period helpers
// ---------------------------------------------------------------------------

function getWeekRange(offset = 0): { label: string; start: string; end: string } {
  const now = new Date();
  const day = now.getDay();
  const monday = new Date(now);
  monday.setDate(now.getDate() - ((day + 6) % 7) + offset * 7);
  const sunday = new Date(monday);
  sunday.setDate(monday.getDate() + 6);
  const fmt = (d: Date) => d.toISOString().split("T")[0];
  const yr = monday.getFullYear();
  const wk = String(Math.ceil(monday.getDate() / 7)).padStart(2, "0");
  return { label: `${yr}-W${wk}`, start: fmt(monday), end: fmt(sunday) };
}

function getMonthRange(offset = 0): { label: string; start: string; end: string } {
  const d = new Date();
  d.setMonth(d.getMonth() + offset);
  const yr = d.getFullYear();
  const mo = d.getMonth();
  const start = new Date(yr, mo, 1);
  const end = new Date(yr, mo + 1, 0);
  const fmt = (x: Date) => x.toISOString().split("T")[0];
  const label = `${yr}-${String(mo + 1).padStart(2, "0")}`;
  return { label, start: fmt(start), end: fmt(end) };
}

function getQuarterRange(offset = 0): { label: string; start: string; end: string } {
  const d = new Date();
  const q = Math.floor(d.getMonth() / 3) + offset;
  const yr = d.getFullYear() + Math.floor(q / 4);
  const qNorm = ((q % 4) + 4) % 4;
  const startMonth = qNorm * 3;
  const start = new Date(yr, startMonth, 1);
  const end = new Date(yr, startMonth + 3, 0);
  const fmt = (x: Date) => x.toISOString().split("T")[0];
  return { label: `${yr}-Q${qNorm + 1}`, start: fmt(start), end: fmt(end) };
}

// ---------------------------------------------------------------------------
// Metric card
// ---------------------------------------------------------------------------

interface MetricCardProps {
  label: string;
  value: string | number;
  sub?: string;
  accent?: string;
}

function MetricCard({ label, value, sub, accent }: MetricCardProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-3">
      <p className="text-[11px] text-muted-foreground">{label}</p>
      <p className={cn("mt-0.5 text-xl font-bold", accent ?? "text-foreground")}>{value}</p>
      {sub && <p className="text-[10px] text-muted-foreground">{sub}</p>}
    </div>
  );
}

// ---------------------------------------------------------------------------
// SnapshotDetail — expandable row showing full metrics
// ---------------------------------------------------------------------------

function SnapshotDetail({ snap }: { snap: ReportSnapshot }) {
  const m = snap.metrics;
  return (
    <div className="border-t border-border bg-muted/20 px-4 py-4">
      {/* Executive summary */}
      {snap.executive_summary && (
        <div className="mb-4 rounded-md border border-border bg-card p-3">
          <p className="mb-1 text-xs font-semibold text-muted-foreground">Executive Summary</p>
          <p className="text-sm text-foreground whitespace-pre-wrap">{snap.executive_summary}</p>
        </div>
      )}

      {/* Metrics grid */}
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
        <MetricCard label="Total Projects"   value={m.total_projects}   sub={`${m.active_projects} active`} />
        <MetricCard label="Tasks Done"       value={`${m.done_tasks}/${m.total_tasks}`} sub={`${m.overdue_tasks} overdue`} accent={m.overdue_tasks > 0 ? "text-destructive" : undefined} />
        <MetricCard label="Avg Progress"     value={`${m.avg_progress_pct}%`} />
        <MetricCard label="Milestones Done"  value={`${m.done_milestones}/${m.total_milestones}`} sub={`${m.overdue_milestones} overdue`} accent={m.overdue_milestones > 0 ? "text-destructive" : undefined} />
        <MetricCard label="Open Risks"       value={m.open_risks} sub={`${m.high_risks} high/critical`} accent={m.high_risks > 0 ? "text-orange-500" : undefined} />
        <MetricCard label="Open Issues"      value={m.open_issues} sub={`of ${m.total_issues}`} />
        <MetricCard label="Planned Budget"   value={formatIDR(m.total_planned_budget)} />
        <MetricCard label="Actual Budget"    value={formatIDR(m.total_actual_budget)} sub={`${m.budget_usage_pct}% used`} accent={m.budget_usage_pct >= 100 ? "text-destructive" : m.budget_usage_pct >= 90 ? "text-orange-500" : undefined} />
        <MetricCard label="Open CA"          value={m.open_corrective_actions} sub={`of ${m.total_corrective_actions}`} accent={m.open_corrective_actions > 0 ? "text-amber-500" : undefined} />
      </div>

      <p className="mt-3 text-[10px] text-muted-foreground">
        Created by {snap.created_by_user?.name ?? snap.created_by} · {formatDate(snap.created_at)}
        {snap.published_at ? ` · Published ${formatDate(snap.published_at)}` : ""}
      </p>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main page
// ---------------------------------------------------------------------------

export default function ReportsPage() {
  const qc = useQueryClient();

  // Filters
  const [periodTypeFilter, setPeriodTypeFilter] = useState<ReportPeriodType | "">("");
  const [statusFilter, setStatusFilter]         = useState<ReportStatus | "">("");
  const [page, setPage]                         = useState(1);

  // Expanded row
  const [expandedId, setExpandedId] = useState<string | null>(null);

  // Generate form state
  const [genOpen, setGenOpen]                 = useState(false);
  const [genPeriodType, setGenPeriodType]     = useState<ReportPeriodType>("MONTHLY");
  const [genFormError, setGenFormError]       = useState<string | null>(null);

  // ---------------------------------------------------------------------------
  // Queries
  // ---------------------------------------------------------------------------

  const reportsQuery = useQuery({
    queryKey: ["reports", periodTypeFilter, statusFilter, page],
    queryFn: () =>
      reportService.listReports({
        period_type: periodTypeFilter || undefined,
        status: statusFilter || undefined,
        page,
        page_size: 20,
      }),
  });

  const reports: ReportSnapshot[] = useMemo(
    () => reportsQuery.data?.data ?? [],
    [reportsQuery.data]
  );
  const meta: PaginationMeta | undefined = reportsQuery.data?.meta;

  // ---------------------------------------------------------------------------
  // Mutations
  // ---------------------------------------------------------------------------

  const generateMutation = useMutation({
    mutationFn: (req: GenerateReportRequest) => reportService.generateReport(req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["reports"] });
      setGenOpen(false);
      setGenFormError(null);
    },
    onError: (e: Error) => setGenFormError(e.message),
  });

  const publishMutation = useMutation({
    mutationFn: (id: string) => reportService.transitionReport(id, "PUBLISHED"),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["reports"] }),
  });

  const archiveMutation = useMutation({
    mutationFn: (id: string) => reportService.transitionReport(id, "ARCHIVED"),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["reports"] }),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => reportService.deleteReport(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["reports"] }),
  });

  // ---------------------------------------------------------------------------
  // Generate handler
  // ---------------------------------------------------------------------------

  function handleGenerate() {
    setGenFormError(null);
    let range: { label: string; start: string; end: string };
    switch (genPeriodType) {
      case "WEEKLY":    range = getWeekRange(0);    break;
      case "MONTHLY":   range = getMonthRange(0);   break;
      case "QUARTERLY": range = getQuarterRange(0); break;
    }
    generateMutation.mutate({
      period_type:  genPeriodType,
      period_label: range.label,
      period_start: range.start,
      period_end:   range.end,
    });
  }

  // ---------------------------------------------------------------------------
  // Render
  // ---------------------------------------------------------------------------

  return (
    <DashboardLayout title="Laporan Periodik">
      <div className="mx-auto max-w-6xl space-y-6 p-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-foreground">Laporan Periodik</h1>
            <p className="text-sm text-muted-foreground">
              Snapshot mingguan, bulanan, dan kuartalan kondisi proyek
            </p>
          </div>
          <button
            onClick={() => setGenOpen(true)}
            className="flex items-center gap-1.5 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            <Plus className="h-4 w-4" />
            Generate Laporan
          </button>
        </div>

        {/* Generate modal */}
        {genOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
            <div className="w-full max-w-sm rounded-lg border border-border bg-card p-6 shadow-xl">
              <h2 className="mb-4 text-base font-semibold">Generate Laporan Baru</h2>

              <div className="space-y-3">
                <div>
                  <label className="mb-1 block text-xs font-medium text-muted-foreground">
                    Periode
                  </label>
                  <select
                    value={genPeriodType}
                    onChange={(e) => setGenPeriodType(e.target.value as ReportPeriodType)}
                    className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="WEEKLY">Weekly — minggu ini</option>
                    <option value="MONTHLY">Monthly — bulan ini</option>
                    <option value="QUARTERLY">Quarterly — kuartal ini</option>
                  </select>
                </div>

                {genFormError && (
                  <p className="rounded bg-destructive/10 px-2 py-1 text-xs text-destructive">
                    {genFormError}
                  </p>
                )}
              </div>

              <div className="mt-5 flex justify-end gap-2">
                <button
                  onClick={() => { setGenOpen(false); setGenFormError(null); }}
                  className="rounded-md border border-input px-3 py-1.5 text-sm hover:bg-muted"
                >
                  Batal
                </button>
                <button
                  onClick={handleGenerate}
                  disabled={generateMutation.isPending}
                  className="flex items-center gap-1 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-60"
                >
                  {generateMutation.isPending ? (
                    <RefreshCw className="h-3.5 w-3.5 animate-spin" />
                  ) : (
                    <BarChart2 className="h-3.5 w-3.5" />
                  )}
                  Generate
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Filters */}
        <div className="flex flex-wrap gap-3">
          <select
            value={periodTypeFilter}
            onChange={(e) => { setPeriodTypeFilter(e.target.value as ReportPeriodType | ""); setPage(1); }}
            className="rounded-md border border-input bg-background px-3 py-1.5 text-sm"
          >
            <option value="">Semua Periode</option>
            <option value="WEEKLY">Weekly</option>
            <option value="MONTHLY">Monthly</option>
            <option value="QUARTERLY">Quarterly</option>
          </select>

          <select
            value={statusFilter}
            onChange={(e) => { setStatusFilter(e.target.value as ReportStatus | ""); setPage(1); }}
            className="rounded-md border border-input bg-background px-3 py-1.5 text-sm"
          >
            <option value="">Semua Status</option>
            <option value="DRAFT">Draft</option>
            <option value="PUBLISHED">Published</option>
            <option value="ARCHIVED">Archived</option>
          </select>

          <button
            onClick={() => reportsQuery.refetch()}
            className="flex items-center gap-1 rounded-md border border-input bg-background px-3 py-1.5 text-sm hover:bg-muted"
          >
            <RefreshCw className={cn("h-3.5 w-3.5", reportsQuery.isFetching && "animate-spin")} />
            Muat Ulang
          </button>
        </div>

        {/* Table */}
        <div className="rounded-lg border border-border bg-card shadow-sm">
          {reportsQuery.isLoading ? (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
              Memuat laporan...
            </div>
          ) : reportsQuery.isError ? (
            <div className="flex items-center justify-center py-16 text-sm text-destructive">
              Gagal memuat laporan.
            </div>
          ) : reports.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-2 py-16">
              <FileText className="h-10 w-10 text-muted-foreground/40" />
              <p className="text-sm text-muted-foreground">
                Belum ada laporan. Klik &quot;Generate Laporan&quot; untuk membuat snapshot pertama.
              </p>
            </div>
          ) : (
            <div className="divide-y divide-border">
              {/* Header row */}
              <div className="grid grid-cols-[1.5rem_1fr_6rem_8rem_8rem_6rem_5rem] items-center gap-3 px-4 py-2 text-xs font-medium text-muted-foreground">
                <span />
                <span>Label</span>
                <span>Tipe</span>
                <span>Mulai</span>
                <span>Selesai</span>
                <span>Status</span>
                <span className="text-right">Aksi</span>
              </div>

              {reports.map((snap) => (
                <div key={snap.id}>
                  {/* Summary row */}
                  <div
                    className="grid cursor-pointer grid-cols-[1.5rem_1fr_6rem_8rem_8rem_6rem_5rem] items-center gap-3 px-4 py-3 hover:bg-muted/30"
                    onClick={() => setExpandedId(expandedId === snap.id ? null : snap.id)}
                  >
                    <span className="text-muted-foreground">
                      {expandedId === snap.id ? (
                        <ChevronDown className="h-3.5 w-3.5" />
                      ) : (
                        <ChevronRight className="h-3.5 w-3.5" />
                      )}
                    </span>
                    <span className="truncate text-sm font-medium text-foreground">
                      {snap.period_label}
                      {snap.project?.name && (
                        <span className="ml-1.5 text-xs text-muted-foreground">· {snap.project.name}</span>
                      )}
                    </span>
                    <span className="text-xs text-muted-foreground">{periodLabel(snap.period_type)}</span>
                    <span className="text-xs text-muted-foreground">{formatDate(snap.period_start)}</span>
                    <span className="text-xs text-muted-foreground">{formatDate(snap.period_end)}</span>
                    <span>{statusBadge(snap.status)}</span>
                    <div
                      className="flex justify-end gap-1"
                      onClick={(e) => e.stopPropagation()}
                    >
                      {snap.status === "DRAFT" && (
                        <button
                          onClick={() => publishMutation.mutate(snap.id)}
                          disabled={publishMutation.isPending}
                          title="Publish"
                          className="rounded p-1 hover:bg-green-100 dark:hover:bg-green-900"
                        >
                          <CheckCircle2 className="h-3.5 w-3.5 text-green-600" />
                        </button>
                      )}
                      {snap.status === "PUBLISHED" && (
                        <button
                          onClick={() => archiveMutation.mutate(snap.id)}
                          disabled={archiveMutation.isPending}
                          title="Archive"
                          className="rounded p-1 hover:bg-amber-100 dark:hover:bg-amber-900"
                        >
                          <FileText className="h-3.5 w-3.5 text-amber-600" />
                        </button>
                      )}
                      <button
                        onClick={() => {
                          if (window.confirm(`Hapus laporan "${snap.period_label}"?`)) {
                            deleteMutation.mutate(snap.id);
                          }
                        }}
                        title="Hapus"
                        className="rounded p-1 hover:bg-red-100 dark:hover:bg-red-900"
                      >
                        <Trash2 className="h-3.5 w-3.5 text-destructive" />
                      </button>
                    </div>
                  </div>

                  {/* Expanded detail */}
                  {expandedId === snap.id && <SnapshotDetail snap={snap} />}
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Pagination */}
        {meta && meta.total_pages > 1 && (
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <span>
              {((meta.page - 1) * meta.page_size) + 1}–{Math.min(meta.page * meta.page_size, meta.total)} dari {meta.total}
            </span>
            <div className="flex gap-2">
              <button
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
                className="rounded border border-input px-2 py-1 hover:bg-muted disabled:opacity-40"
              >
                ‹ Prev
              </button>
              <button
                disabled={page >= meta.total_pages}
                onClick={() => setPage((p) => p + 1)}
                className="rounded border border-input px-2 py-1 hover:bg-muted disabled:opacity-40"
              >
                Next ›
              </button>
            </div>
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}
