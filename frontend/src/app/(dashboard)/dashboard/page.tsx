"use client";

import Link from "next/link";
import Image from "next/image";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  AlertCircle,
  AlertTriangle,
  ArrowRight,
  BriefcaseBusiness,
  CircleDollarSign,
  CircleGauge,
  Filter,
  Gavel,
  Maximize2,
  RefreshCw,
  WalletCards,
  X,
  type LucideIcon,
} from "lucide-react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  Cell,
  Legend,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { dashboardService } from "@/services/dashboard.service";
import { projectService } from "@/services/project.service";
import { useAuthStore } from "@/store/auth.store";
import type {
  DashboardStats,
  DashboardTrend,
  DashboardWarning,
  DashboardWarningSeverity,
  TrendPoint,
} from "@/types/dashboard";
import type { Project, ProjectStatus } from "@/types/project";
import { cn } from "@/lib/utils";

const emptyStats: DashboardStats = {
  total_projects: 0,
  active_projects: 0,
  on_hold_projects: 0,
  closed_projects: 0,
  total_tasks: 0,
  todo_tasks: 0,
  in_progress_tasks: 0,
  done_tasks: 0,
  overdue_tasks: 0,
  total_milestones: 0,
  pending_milestones: 0,
  done_milestones: 0,
  overdue_milestones: 0,
  total_users: 0,
  active_users: 0,
  early_warnings: [],
};

const emptyProjects: Project[] = [];

const statusColors: Record<string, string> = {
  Aktif: "#17864b",
  "Dalam Perencanaan": "#1671b8",
  "Ditunda": "#f2a900",
  Selesai: "#21a8b8",
  Lainnya: "#94a3b8",
};

const dashboardSections = [
  { label: "Ringkasan", href: "#overview" },
  { label: "Portofolio", href: "#portfolio" },
  { label: "Monitoring", href: "#monitoring" },
  { label: "Risiko", href: "#risiko" },
  { label: "Isu", href: "#isu" },
  { label: "Tindak Lanjut", href: "#tindak-lanjut" },
  { label: "Keputusan", href: "#keputusan" },
];

export default function DashboardPage() {
  const accessToken = useAuthStore((state) => state.accessToken);

  const statsQuery = useQuery({
    queryKey: ["dashboard", "stats"],
    queryFn: async () => {
      const { data: response } = await dashboardService.getStats();
      return response.data;
    },
    enabled: Boolean(accessToken),
    staleTime: 30_000,
  });

  const projectsQuery = useQuery({
    queryKey: ["dashboard", "projects"],
    // staleTime 0: data projects selalu refetch saat window focus agar
    // perubahan org_unit_name (dari /settings/org-units) langsung terefleksi.
    queryFn: async () => {
      const { data: response } = await projectService.list({
        page: 1,
        page_size: 100,
        sort_by: "updated_at",
        sort_dir: "desc",
      });
      return response.data;
    },
    enabled: Boolean(accessToken),
    staleTime: 0,
  });

  const stats = statsQuery.data ?? emptyStats;
  const projects = projectsQuery.data ?? emptyProjects;
  const isLoading = !accessToken || statsQuery.isLoading || projectsQuery.isLoading;
  const isFetching = statsQuery.isFetching || projectsQuery.isFetching;
  const isError = statsQuery.isError || projectsQuery.isError;

  const portfolio = useMemo(() => summarizePortfolio(projects, stats), [projects, stats]);
  const statusData = useMemo(() => buildStatusData(projects), [projects]);
  const fallbackTrend = useMemo(() => buildOperationalTrendFromProjects(projects), [projects]);
  const attentionWarnings = useMemo(
    () =>
      deduplicateWarnings(stats.early_warnings)
        .sort((a, b) => severityRank(b.severity) - severityRank(a.severity))
        .slice(0, 5),
    [stats.early_warnings]
  );
  const milestoneWarnings = stats.early_warnings
    .filter((warning) => warning.type === "OVERDUE_MILESTONE")
    .slice(0, 5);
  const priorityProjects = useMemo(() => rankProjects(projects).slice(0, 8), [projects]);
  const decisionWarnings = attentionWarnings.filter(
    (warning) => warning.severity === "CRITICAL" || warning.severity === "HIGH"
  ).slice(0, 4);
  const issueSummary = useMemo(() => summarizeWarnings(stats.early_warnings), [stats.early_warnings]);
  // Prefer real risk-register entries (RISK_REGISTER) for the "Risiko Utama"
  // panel. Fall back to generic CRITICAL/HIGH early warnings when the backend
  // has not yet produced any risk-register data (e.g. no open risks).
  const riskRegisterWarnings = useMemo(
    () => deduplicateWarnings(stats.early_warnings.filter((warning) => warning.type === "RISK_REGISTER")).slice(0, 5),
    [stats.early_warnings]
  );
  const riskWarnings = useMemo(() => {
    if (riskRegisterWarnings.length > 0) {
      return riskRegisterWarnings;
    }
    return deduplicateWarnings(
      stats.early_warnings.filter(
        (warning) => warning.severity === "CRITICAL" || warning.severity === "HIGH"
      )
    ).slice(0, 5);
  }, [riskRegisterWarnings, stats.early_warnings]);
  const uniqueRiskWarnings = riskWarnings;

  const refresh = () => {
    void Promise.all([statsQuery.refetch(), projectsQuery.refetch()]);
  };

  const kpis: KpiData[] = [
    {
      label: "Total Proyek Aktif",
      value: formatNumber(stats.total_projects),
      detail: `${stats.active_projects} proyek aktif`,
      icon: BriefcaseBusiness,
      tone: "blue",
    },
    {
      label: "Nilai Portofolio",
      value: formatCompactCurrency(portfolio.totalBudget),
      detail: "Total nilai proyek tercatat",
      icon: CircleDollarSign,
      tone: "navy",
    },
    {
      label: "Progres Fisik Nasional",
      value: `${formatDecimal(portfolio.averageProgress)}%`,
      detail: "Rata-rata progres proyek",
      icon: CircleGauge,
      tone: "teal",
    },
    {
      label: "Penyerapan Keuangan",
      value: "N/A",
      detail: "Menunggu data realisasi",
      icon: WalletCards,
      tone: "green",
    },
    {
      label: "Proyek Kritis",
      value: formatNumber(portfolio.criticalProjects),
      detail: `${portfolio.highWarnings} peringatan prioritas tinggi`,
      icon: AlertTriangle,
      tone: "red",
    },
    {
      label: "Anggaran Berisiko",
      value: formatCompactCurrency(portfolio.atRiskBudget),
      detail: "Berdasarkan warning anggaran",
      icon: CircleDollarSign,
      tone: "amber",
    },
  ];

  return (
    <DashboardLayout title="Dashboard">
      <section id="overview" className="scroll-mt-24">
        <div className="mb-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <nav
            className="-mx-1 overflow-x-auto px-1"
            aria-label="Navigasi bagian dashboard"
          >
            <div className="flex min-w-max items-center gap-1 rounded-md border border-slate-200 bg-white p-1 shadow-sm">
              {dashboardSections.map((section) => (
                <Link
                  key={section.href}
                  href={section.href}
                  className="inline-flex h-8 items-center rounded px-3 text-xs font-semibold text-slate-600 transition-colors hover:bg-blue-50 hover:text-[#0b5aa2]"
                >
                  {section.label}
                </Link>
              ))}
            </div>
          </nav>

          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <label className="flex h-9 items-center gap-2 rounded-md border border-slate-200 bg-white px-3 text-xs text-slate-600">
              <span className="font-semibold text-[#082e63]">Periode Data:</span>
              <input
                type="month"
                defaultValue={new Date().toISOString().slice(0, 7)}
                className="bg-transparent text-xs outline-none"
                aria-label="Periode data"
              />
            </label>
            <button
              type="button"
              onClick={refresh}
              disabled={!accessToken || isFetching}
              className="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-slate-200 bg-white px-3 text-xs font-semibold text-[#0b5aa2] hover:bg-blue-50 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <Filter className="h-4 w-4" aria-hidden="true" />
              Muat Ulang
              <RefreshCw className={cn("h-3.5 w-3.5", isFetching && "animate-spin")} aria-hidden="true" />
            </button>
          </div>
        </div>
      </section>

      {isError && (
        <div role="alert" className="mb-5 flex items-start gap-3 rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
          <div>
            <p className="font-semibold">Data dashboard belum dapat dimuat.</p>
            <p className="mt-1 text-xs">Periksa koneksi backend atau perbarui sesi login.</p>
          </div>
        </div>
      )}

      <section
        id="portfolio"
        className="grid scroll-mt-24 grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3 min-[2200px]:grid-cols-6"
        aria-label="Indikator utama"
      >
        {kpis.map((kpi) => (
          <KpiCard key={kpi.label} data={kpi} loading={isLoading} />
        ))}
      </section>

      <section className="mt-3 grid grid-cols-1 gap-3 xl:grid-cols-12">
        <Panel title="Kesehatan Proyek" className="xl:col-span-3">
          <div className="flex min-h-64 flex-col">
            <div className="relative mx-auto h-44 w-full max-w-64">
              {isLoading ? (
                <ChartSkeleton />
              ) : (
                <>
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={statusData.chart}
                        dataKey="value"
                        nameKey="name"
                        innerRadius={48}
                        outerRadius={70}
                        paddingAngle={2}
                        strokeWidth={0}
                      >
                        {statusData.chart.map((entry) => (
                          <Cell key={entry.name} fill={entry.color} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                  <div className="pointer-events-none absolute inset-0 grid place-items-center text-center">
                    <div>
                      <p className="text-[9px] uppercase text-slate-500">Total Proyek</p>
                      <p className="text-xl font-bold text-[#082e63]">{stats.total_projects}</p>
                    </div>
                  </div>
                </>
              )}
            </div>
            <div className="mt-2 grid grid-cols-1 gap-x-4 sm:grid-cols-2 xl:grid-cols-1">
              {statusData.legend.map((entry) => (
                <div key={entry.name} className="grid min-w-0 grid-cols-[auto_minmax(0,1fr)_auto_auto] items-center gap-2 border-b border-slate-100 py-2 text-[10px]">
                  <span className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: entry.color }} aria-hidden="true" />
                  <span className="min-w-0 flex-1 text-slate-600">{entry.name}</span>
                  <span className="font-bold text-[#082e63]">{entry.value}</span>
                  <span className="w-8 text-right text-slate-400">{entry.percent}%</span>
                </div>
              ))}
            </div>
          </div>
          <p className="mt-1 text-[9px] text-slate-400">
            Distribusi memakai status operasional; formula Health Score menunggu modul P1-014.
          </p>
        </Panel>

        <Panel
          title="Perhatian Pimpinan"
          action={{ label: "Lihat semua", href: "/projects" }}
          className="xl:col-span-5"
        >
          <div className="overflow-x-auto">
            <table className="w-full min-w-[560px] text-left text-[10px]">
              <thead>
                <tr className="border-b border-slate-200 text-[10px] uppercase text-slate-500">
                  <th className="pb-2 font-semibold">Proyek</th>
                  <th className="pb-2 font-semibold">Deviasi / Risiko</th>
                  <th className="pb-2 font-semibold">Rekomendasi</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <TableLoading columns={3} />
                ) : attentionWarnings.length > 0 ? (
                  attentionWarnings.map((warning) => (
                    <tr key={warning.id} className="border-b border-slate-100 last:border-0">
                      <td className="max-w-48 py-2.5 pr-3">
                        <Link href={warning.project_id ? `/projects/${warning.project_id}` : "/projects"} className="font-semibold text-[#0b4c91] hover:underline">
                          {warning.project_name ?? warning.title}
                        </Link>
                        <p className="mt-1 line-clamp-1 text-[9px] text-slate-500">{warning.due_date ? formatDate(warning.due_date) : "Tanpa tenggat"}</p>
                      </td>
                      <td className="py-2.5 pr-3">
                        <SeverityBadge severity={warning.severity} />
                        <p className="mt-1 text-[9px] text-slate-500">{formatWarningType(warning.type)}</p>
                      </td>
                      <td className="max-w-56 py-2.5 text-slate-600">{warning.message}</td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={3} className="py-12 text-center text-xs text-slate-500">
                      Tidak ada perhatian aktif berdasarkan rule saat ini.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </Panel>

        <Panel
          title="Keputusan Diperlukan"
          action={{ label: "Lihat semua", href: "/projects" }}
          className="xl:col-span-4"
          id="keputusan"
        >
          <div className="space-y-1">
            {isLoading ? (
              <ListLoading />
            ) : decisionWarnings.length > 0 ? (
              decisionWarnings.map((warning) => (
                <DecisionRow key={warning.id} warning={warning} />
              ))
            ) : (
              <EmptyState text="Belum ada keputusan prioritas yang perlu ditindaklanjuti." />
            )}
          </div>
        </Panel>
      </section>

      <section
        id="monitoring"
        className="mt-3 grid scroll-mt-24 grid-cols-1 items-stretch gap-3 xl:grid-cols-2 2xl:grid-cols-12"
      >
        <Panel title="Peta Sebaran Proyek" className="xl:col-span-1 2xl:col-span-6" id="peta">
          <ProjectMapPanel />
        </Panel>

        <Panel title="Tren Progres Fisik vs Keuangan" className="xl:col-span-1 2xl:col-span-6">
          <TrendPanel fallbackTrend={fallbackTrend} />
        </Panel>

        <Panel
          title="Isu Utama"
          action={{ label: "Lihat semua", href: "/projects" }}
          className="xl:col-span-1 2xl:col-span-4"
          id="isu"
        >
          <RankedSummary rows={issueSummary} total={stats.early_warnings.length} label="Total isu aktif" />
        </Panel>

        <Panel
          title="Risiko Utama"
          action={{ label: "Lihat semua", href: "/projects" }}
          className="xl:col-span-1 2xl:col-span-4"
          id="risiko"
        >
          <p className="mb-2 text-[10px] text-slate-400">
            {riskRegisterWarnings.length > 0
              ? "Dari risk register proyek (score probability × impact)."
              : "Derived dari early warning (risk register belum memiliki data terbuka)."}
          </p>
          <div className="space-y-1">
            {uniqueRiskWarnings.length > 0 ? (
              uniqueRiskWarnings.map((warning) => <WarningRow key={warning.id} warning={warning} compact />)
            ) : (
              <EmptyState text="Belum ada risiko tinggi." />
            )}
          </div>
        </Panel>

        <Panel
          title="Milestone Kritis"
          action={{ label: "Lihat semua", href: "/projects" }}
          className="xl:col-span-2 2xl:col-span-4"
          id="tindak-lanjut"
        >
          <div className="space-y-1">
            {isLoading ? (
              <ListLoading />
            ) : milestoneWarnings.length > 0 ? (
              milestoneWarnings.map((warning) => <WarningRow key={warning.id} warning={warning} compact />)
            ) : (
              <EmptyState text="Tidak ada milestone yang melewati target." />
            )}
          </div>
        </Panel>
      </section>

      <section className="mt-3" id="laporan">
        <Panel
          title="10 Proyek Prioritas"
          action={{ label: "Lihat semua proyek kritis", href: "/projects" }}
        >
          <div className="overflow-x-auto">
            <table className="w-full min-w-[1040px] text-left text-[10px]">
              <thead>
                <tr className="border-b border-slate-200 uppercase text-slate-500">
                  <th className="pb-2 font-semibold">No.</th>
                  <th className="pb-2 font-semibold">Proyek</th>
                  <th className="pb-2 font-semibold">Kode</th>
                  <th className="pb-2 font-semibold">Balai</th>
                  <th className="pb-2 font-semibold">Nilai Kontrak</th>
                  <th className="pb-2 font-semibold">Progres Fisik</th>
                  <th className="pb-2 font-semibold">Status</th>
                  <th className="pb-2 font-semibold">Risiko Utama</th>
                  <th className="pb-2 font-semibold">Rencana Tindak Lanjut</th>
                  <th className="pb-2 text-right font-semibold">Target Selesai</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <TableLoading columns={10} />
                ) : priorityProjects.length > 0 ? (
                  priorityProjects.map((project, index) => {
                    const warning = stats.early_warnings.find((item) => item.project_id === project.id);
                    return (
                      <tr key={project.id} className="border-b border-slate-100 last:border-0 hover:bg-blue-50/40">
                        <td className="py-2.5 pr-3 text-slate-400">{index + 1}</td>
                        <td className="max-w-64 py-2.5 pr-4">
                          <Link href={`/projects/${project.id}`} className="font-semibold text-[#0b4c91] hover:underline">
                            {project.name}
                          </Link>
                        </td>
                        <td className="py-2.5 pr-4 font-mono text-[9px] text-slate-500">{project.code}</td>
                        <td className="py-2.5 pr-4 max-w-[120px] truncate text-slate-600" title={project.org_unit_name ?? undefined}>{project.org_unit_name || <span className="text-slate-400">-</span>}</td>
                        <td className="py-2.5 pr-4 font-medium text-slate-700">{formatCompactCurrency(project.budget_total)}</td>
                        <td className="py-2.5 pr-4">
                          <div className="flex min-w-32 items-center gap-2">
                            <div className="h-2 flex-1 overflow-hidden rounded-full bg-slate-100">
                              <div className="h-full rounded-full bg-[#1262b8]" style={{ width: `${clampPercent(project.progress_pct)}%` }} />
                            </div>
                            <span className="w-9 text-right font-semibold text-slate-600">{formatDecimal(project.progress_pct)}%</span>
                          </div>
                        </td>
                        <td className="py-2.5 pr-4"><ProjectStatusBadge status={project.status} /></td>
                        <td className="max-w-40 py-2.5 pr-4 text-slate-600">{warning ? formatWarningType(warning.type) : "-"}</td>
                        <td className="max-w-56 py-2.5 pr-4 text-slate-600">{warning?.message ?? "Belum ada tindak lanjut tercatat"}</td>
                        <td className="py-2.5 text-right text-slate-600">{project.end_date ? formatDate(project.end_date) : "-"}</td>
                      </tr>
                    );
                  })
                ) : (
                  <tr>
                    <td colSpan={9} className="py-12 text-center text-xs text-slate-500">
                      Belum ada proyek. <Link href="/projects" className="font-semibold text-[#0b4c91] hover:underline">Buat proyek pertama</Link>
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </Panel>
      </section>
    </DashboardLayout>
  );
}

interface KpiData {
  label: string;
  value: string;
  detail: string;
  icon: LucideIcon;
  tone: "blue" | "navy" | "teal" | "green" | "red" | "amber";
}

const kpiTones: Record<KpiData["tone"], { icon: string; box: string; detail: string }> = {
  blue: { icon: "text-blue-700", box: "bg-blue-50", detail: "text-blue-700" },
  navy: { icon: "text-[#0b4c91]", box: "bg-slate-100", detail: "text-slate-500" },
  teal: { icon: "text-cyan-700", box: "bg-cyan-50", detail: "text-cyan-700" },
  green: { icon: "text-emerald-700", box: "bg-emerald-50", detail: "text-emerald-700" },
  red: { icon: "text-red-600", box: "bg-red-50", detail: "text-red-600" },
  amber: { icon: "text-amber-600", box: "bg-amber-50", detail: "text-amber-700" },
};

function KpiCard({ data, loading }: { data: KpiData; loading: boolean }) {
  const Icon = data.icon;
  const tone = kpiTones[data.tone];

  return (
    <article className="grid min-h-32 grid-cols-[auto_minmax(0,1fr)] items-center gap-4 rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <div className={cn("grid h-12 w-12 shrink-0 place-items-center rounded-md", tone.box)}>
        <Icon className={cn("h-6 w-6", tone.icon)} aria-hidden="true" />
      </div>
      <div className="min-w-0">
        <p className="min-h-8 text-[10px] font-semibold uppercase leading-4 text-slate-500">{data.label}</p>
        {loading ? (
          <div className="my-2 h-7 w-24 rounded-md bg-slate-100" />
        ) : (
          <p className="mt-1 break-words text-2xl font-bold leading-none text-[#082e63] tabular-nums">{data.value}</p>
        )}
        <p className={cn("mt-2 min-h-8 text-[10px] leading-4", tone.detail)}>{data.detail}</p>
      </div>
    </article>
  );
}

function Panel({
  title,
  children,
  action,
  className,
  id,
}: {
  title: string;
  children: React.ReactNode;
  action?: { label: string; href: string };
  className?: string;
  id?: string;
}) {
  return (
    <article id={id} className={cn("min-w-0 scroll-mt-24 rounded-lg border border-slate-200 bg-white p-4 shadow-sm", className)}>
      <div className="mb-3 flex min-h-9 items-start justify-between gap-4 border-b border-slate-200 pb-2">
        <h3 className="min-w-0 text-xs font-bold uppercase leading-5 text-[#082e63]">{title}</h3>
        {action && (
          <Link href={action.href} className="inline-flex shrink-0 items-center gap-1 whitespace-nowrap text-[10px] font-semibold leading-5 text-[#0b5aa2] hover:underline">
            {action.label}
            <ArrowRight className="h-3 w-3" aria-hidden="true" />
          </Link>
        )}
      </div>
      {children}
    </article>
  );
}

function WarningRow({ warning, compact = false }: { warning: DashboardWarning; compact?: boolean }) {
  return (
    <Link
      href={warning.project_id ? `/projects/${warning.project_id}` : "/projects"}
      className="flex items-start gap-3 border-b border-slate-100 py-3 last:border-0 hover:bg-blue-50/50"
    >
      <span className={cn("mt-1 h-2.5 w-2.5 shrink-0 rounded-full", severityDot(warning.severity))} />
      <div className="min-w-0 flex-1">
        <p className="line-clamp-2 text-[11px] font-semibold leading-4 text-[#17345d]">{warning.title}</p>
        {!compact && <p className="mt-1 line-clamp-1 text-[10px] text-slate-500">{warning.project_name ?? warning.message}</p>}
      </div>
      {warning.due_date && (
        <span className="shrink-0 text-[10px] text-slate-500">{formatShortDate(warning.due_date)}</span>
      )}
    </Link>
  );
}

function DecisionRow({ warning }: { warning: DashboardWarning }) {
  return (
    <Link
      href={warning.project_id ? `/projects/${warning.project_id}` : "/projects"}
      className="flex items-start gap-3 border-b border-slate-100 py-2.5 last:border-0 hover:bg-blue-50/50"
    >
      <div className="grid h-8 w-8 shrink-0 place-items-center rounded-md bg-blue-50 text-[#0b4c91]">
        <Gavel className="h-4 w-4" aria-hidden="true" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-3">
          <p className="line-clamp-2 text-[10px] font-semibold leading-4 text-[#17345d]">{warning.title}</p>
          <SeverityBadge severity={warning.severity} />
        </div>
        <p className="mt-1 line-clamp-1 text-[9px] text-slate-500">{warning.message}</p>
        {warning.due_date && (
          <p className="mt-1 text-[9px] text-[#0b5aa2]">Batas: {formatDate(warning.due_date)}</p>
        )}
      </div>
    </Link>
  );
}

// ─── Trend Chart ────────────────────────────────────────────────────────────

function ProjectMapPanel() {
  const [showModal, setShowModal] = useState(false);

  return (
    <>
      <div className="relative min-h-72 overflow-hidden rounded-md border border-slate-100 bg-[#dceffc] 2xl:min-h-80">
        <button
          type="button"
          onClick={() => setShowModal(true)}
          className="absolute right-2 top-2 z-10 rounded-md bg-white/90 p-1.5 text-slate-500 shadow-sm ring-1 ring-slate-200 hover:bg-white hover:text-slate-700"
          aria-label="Fullscreen peta"
          title="Fullscreen peta"
        >
          <Maximize2 className="h-3.5 w-3.5" aria-hidden="true" />
        </button>
        <Image
          src="/images/indonesia-project-map.png"
          alt="Peta Indonesia untuk sebaran proyek"
          fill
          sizes="(min-width: 1280px) 28vw, 100vw"
          className="object-cover"
        />
        <div className="absolute inset-x-3 bottom-3 rounded-md border border-white/80 bg-white/90 px-3 py-2 text-center text-[9px] text-slate-600">
          Peta ringkas sebaran proyek. Detail interaktif tersedia di GIS.
        </div>
      </div>

      {showModal && <ProjectMapModal onClose={() => setShowModal(false)} />}
    </>
  );
}

function ProjectMapModal({ onClose }: { onClose: () => void }) {
  return (
    <div
      className="fixed inset-0 z-50 bg-black/55"
      role="dialog"
      aria-modal="true"
      aria-label="Peta Sebaran Proyek"
    >
      <div className="flex h-screen w-screen flex-col overflow-hidden bg-white shadow-2xl">
        <div className="flex min-h-16 items-center justify-between border-b border-slate-200 px-5 py-3 sm:px-8">
          <div>
            <h2 className="text-base font-semibold text-slate-800 sm:text-lg">Peta Sebaran Proyek</h2>
            <p className="mt-0.5 text-[10px] text-slate-500">Ringkasan nasional berbasis data proyek aktif</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-md p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-700"
            aria-label="Tutup"
          >
            <X className="h-5 w-5" aria-hidden="true" />
          </button>
        </div>

        <div className="relative min-h-0 flex-1 bg-[#dceffc]">
          <Image
            src="/images/indonesia-project-map.png"
            alt="Peta Indonesia untuk sebaran proyek"
            fill
            sizes="100vw"
            priority
            className="object-contain"
          />
          <div className="absolute bottom-5 left-1/2 w-[calc(100%-2.5rem)] max-w-3xl -translate-x-1/2 rounded-md border border-white/80 bg-white/90 px-4 py-3 text-center text-xs font-medium text-slate-600 shadow-sm">
            Peta ringkas sebaran proyek. Buka modul GIS untuk filter, marker, dan detail proyek interaktif.
          </div>
        </div>
      </div>
    </div>
  );
}

function formatIDR(v: number): string {
  if (v >= 1_000_000_000) return `Rp ${(v / 1_000_000_000).toFixed(1)}M`;
  if (v >= 1_000_000) return `Rp ${(v / 1_000_000).toFixed(0)}jt`;
  return `Rp ${v.toLocaleString("id-ID")}`;
}

function buildOperationalTrendFromProjects(projects: Project[]): DashboardTrend {
  const activeProjects = projects.filter((project) => project.status !== "CANCELLED");
  const totalBudget = activeProjects.reduce((sum, project) => sum + Number(project.budget_total || 0), 0);
  const currentPhysical = activeProjects.length > 0
    ? activeProjects.reduce((sum, project) => sum + Number(project.progress_pct || 0), 0) / activeProjects.length
    : 0;
  const currentFinancial = totalBudget > 0
    ? activeProjects.reduce(
        (sum, project) => sum + Number(project.budget_total || 0) * (Number(project.progress_pct || 0) / 100),
        0
      ) / totalBudget * 100
    : 0;

  const now = new Date();
  const points: TrendPoint[] = Array.from({ length: 12 }, (_, index) => {
    const monthDate = new Date(now.getFullYear(), now.getMonth() - (11 - index), 1);
    const month = `${monthDate.getFullYear()}-${String(monthDate.getMonth() + 1).padStart(2, "0")}`;
    const ramp = (index + 1) / 12;
    const physical = currentPhysical * ramp;
    const financial = currentFinancial * ramp;
    const planned = totalBudget * ramp;
    const actual = totalBudget * (financial / 100);

    return {
      month,
      physical_pct: Number(physical.toFixed(2)),
      financial_pct: Number(financial.toFixed(2)),
      planned: Number(planned.toFixed(2)),
      actual: Number(actual.toFixed(2)),
      data_type: "OPERATIONAL",
    };
  });

  return {
    points,
    data_type: "OPERATIONAL",
  };
}

function TrendChartContent({ points, height = 200 }: { points: TrendPoint[]; height?: number }) {
  return (
    <ResponsiveContainer width="100%" height={height}>
      <AreaChart data={points} margin={{ top: 4, right: 8, left: -20, bottom: 0 }}>
        <defs>
          <linearGradient id="gradPhysical" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#1d4ed8" stopOpacity={0.25} />
            <stop offset="95%" stopColor="#1d4ed8" stopOpacity={0} />
          </linearGradient>
          <linearGradient id="gradFinancial" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#06b6d4" stopOpacity={0.25} />
            <stop offset="95%" stopColor="#06b6d4" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
        <XAxis
          dataKey="month"
          tick={{ fontSize: 9, fill: "#94a3b8" }}
          tickLine={false}
          axisLine={false}
        />
        <YAxis
          domain={[0, 100]}
          tick={{ fontSize: 9, fill: "#94a3b8" }}
          tickLine={false}
          axisLine={false}
          tickFormatter={(v: number) => `${v}%`}
        />
        <Tooltip
          contentStyle={{ fontSize: 11, borderRadius: 6, border: "1px solid #e2e8f0" }}
          formatter={(value: number, name: string) => [
            `${value.toFixed(1)}%`,
            name === "physical_pct" ? "Progres Fisik" : "Penyerapan Keuangan",
          ]}
        />
        <Legend
          iconSize={8}
          wrapperStyle={{ fontSize: 10 }}
          formatter={(v: string) =>
            v === "physical_pct" ? "Progres Fisik" : "Penyerapan Keuangan"
          }
        />
        <Area
          type="monotone"
          dataKey="physical_pct"
          stroke="#1d4ed8"
          strokeWidth={2}
          fill="url(#gradPhysical)"
          dot={false}
          activeDot={{ r: 4 }}
        />
        <Area
          type="monotone"
          dataKey="financial_pct"
          stroke="#06b6d4"
          strokeWidth={2}
          fill="url(#gradFinancial)"
          dot={false}
          activeDot={{ r: 4 }}
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}

function TrendModal({ trend, onClose }: { trend: DashboardTrend; onClose: () => void }) {
  return (
    <div
      className="fixed inset-0 z-50 bg-black/55"
      role="dialog"
      aria-modal="true"
      aria-label="Tren Progres Fisik vs Keuangan"
    >
      <div className="flex h-screen w-screen flex-col overflow-hidden bg-white shadow-2xl">
        <div className="flex min-h-16 items-center justify-between border-b border-slate-200 px-5 py-3 sm:px-8">
          <div>
            <h2 className="text-base font-semibold text-slate-800 sm:text-lg">Tren Progres Fisik vs Keuangan</h2>
            {trend.data_type === "PERIODIC_REPORT" ? (
              <p className="mt-0.5 text-[10px] text-blue-600">
                Laporan periodik operasional — data dari input periodik proyek
              </p>
            ) : trend.data_type === "OPERATIONAL" && (
              <p className="mt-0.5 text-[10px] text-amber-600">
                Data operasional sementara, belum snapshot resmi
              </p>
            )}
          </div>
          <button
            onClick={onClose}
            className="rounded-md p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-700"
            aria-label="Tutup"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="min-h-0 flex-1 px-5 pt-5 sm:px-8">
          <TrendChartContent points={trend.points} height={360} />
        </div>

        <div className="max-h-[42vh] min-h-0 overflow-auto border-t border-slate-200 px-5 pb-6 pt-3 sm:px-8">
          <table className="w-full min-w-[720px] text-[11px]">
            <thead>
              <tr className="border-b border-slate-200 text-left text-slate-500">
                <th className="py-2 pr-4 font-medium">Periode</th>
                <th className="py-2 pr-4 font-medium">Fisik %</th>
                <th className="py-2 pr-4 font-medium">Keuangan %</th>
                <th className="py-2 pr-4 font-medium">Planned</th>
                <th className="py-2 pr-4 font-medium">Actual</th>
                <th className="py-2 font-medium">Variance</th>
              </tr>
            </thead>
            <tbody>
              {trend.points.map((p) => {
                const variance = p.planned - p.actual;
                return (
                  <tr key={p.month} className="border-b border-slate-100 hover:bg-slate-50">
                    <td className="py-1.5 pr-4 font-medium text-slate-700">{p.month}</td>
                    <td className="py-1.5 pr-4">
                      <span
                        className={cn(
                          "font-semibold",
                          p.physical_pct >= 80 ? "text-green-600" :
                          p.physical_pct >= 50 ? "text-amber-600" : "text-slate-600"
                        )}
                      >
                        {p.physical_pct.toFixed(1)}%
                      </span>
                    </td>
                    <td className="py-1.5 pr-4">
                      <span
                        className={cn(
                          "font-semibold",
                          p.financial_pct >= 80 ? "text-green-600" :
                          p.financial_pct >= 50 ? "text-cyan-600" : "text-slate-600"
                        )}
                      >
                        {p.financial_pct.toFixed(1)}%
                      </span>
                    </td>
                    <td className="py-1.5 pr-4 text-slate-600">{p.planned > 0 ? formatIDR(p.planned) : "—"}</td>
                    <td className="py-1.5 pr-4 text-slate-600">{p.actual > 0 ? formatIDR(p.actual) : "—"}</td>
                    <td className={cn("py-1.5", variance >= 0 ? "text-green-600" : "text-red-500")}>
                      {p.planned > 0 ? (variance >= 0 ? "+" : "") + formatIDR(variance) : "—"}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function TrendPanel({ fallbackTrend }: { fallbackTrend: DashboardTrend }) {
  const [showModal, setShowModal] = useState(false);

  const { data: trendResp, isLoading, isError } = useQuery({
    queryKey: ["dashboard-trend"],
    queryFn: () => dashboardService.getTrend().then((r) => r.data.data),
    staleTime: 5 * 60 * 1000,
  });

  const trend = trendResp ?? fallbackTrend;
  const hasData = trend.points.some((p) => p.physical_pct > 0 || p.financial_pct > 0);
  const usingFallback = !trendResp && hasData;

  return (
    <div className="relative flex min-h-72 flex-col overflow-hidden rounded-md bg-white 2xl:min-h-80">
      {/* Expand button */}
      {hasData && (
        <button
          onClick={() => setShowModal(true)}
          className="absolute right-2 top-2 z-10 rounded-md p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
          aria-label="Fullscreen"
        >
          <Maximize2 className="h-3.5 w-3.5" />
        </button>
      )}

      {/* Operational/periodic data disclaimer */}
      {hasData && trend.data_type === "PERIODIC_REPORT" && (
        <div className="mx-3 mt-2 rounded border border-blue-200 bg-blue-50 px-2.5 py-1 text-[9px] text-blue-700">
          Laporan periodik operasional — data dari input periodik proyek
        </div>
      )}
      {hasData && trend.data_type === "OPERATIONAL" && (
        <div className="mx-3 mt-2 rounded border border-amber-200 bg-amber-50 px-2.5 py-1 text-[9px] text-amber-700">
          {usingFallback
            ? "Data operasional sementara dari progres dan nilai proyek"
            : "Data operasional sementara, belum snapshot resmi"}
        </div>
      )}

      {/* Chart or loading/error states */}
      <div className="flex-1 px-1 py-2">
        {isLoading && (
          <div className="flex h-48 items-center justify-center text-xs text-slate-400">
            Memuat data tren…
          </div>
        )}
        {isError && !hasData && (
          <div className="flex h-48 items-center justify-center text-xs text-red-400">
            Gagal memuat data tren.
          </div>
        )}
        {!isLoading && !isError && !hasData && (
          <div className="flex h-48 flex-col items-center justify-center text-center">
            <p className="text-xs font-medium text-slate-500">Belum ada data tren tersedia</p>
            <p className="mt-1 text-[9px] leading-4 text-slate-400">
              Data akan muncul setelah progres dan anggaran proyek dicatat.
            </p>
          </div>
        )}
        {!isLoading && !isError && hasData && trend && (
          <TrendChartContent points={trend.points} height={210} />
        )}
      </div>

      {/* Fullscreen modal */}
      {showModal && trend && (
        <TrendModal trend={trend} onClose={() => setShowModal(false)} />
      )}
    </div>
  );
}

function RankedSummary({
  rows,
  total,
  label,
}: {
  rows: Array<{ label: string; value: number; tone: string }>;
  total: number;
  label: string;
}) {
  return (
    <div>
      <div className="space-y-1">
        {rows.length > 0 ? (
          rows.slice(0, 5).map((row) => (
            <div key={row.label} className="flex items-center gap-2 border-b border-slate-100 py-2.5 text-[10px] last:border-0">
              <span className={cn("h-2 w-2 shrink-0 rounded-full", row.tone)} />
              <span className="min-w-0 flex-1 truncate text-slate-600">{row.label}</span>
              <span className="font-bold text-[#082e63]">{row.value}</span>
            </div>
          ))
        ) : (
          <EmptyState text="Belum ada isu aktif." />
        )}
      </div>
      <div className="mt-3 flex items-center justify-between border-t border-slate-200 pt-3 text-[10px]">
        <span className="text-slate-500">{label}</span>
        <span className="font-bold text-[#082e63]">{total}</span>
      </div>
    </div>
  );
}

function SeverityBadge({ severity }: { severity: DashboardWarningSeverity }) {
  const classes =
    severity === "CRITICAL"
      ? "bg-red-100 text-red-700"
      : severity === "HIGH"
        ? "bg-orange-100 text-orange-700"
        : severity === "MEDIUM"
          ? "bg-amber-100 text-amber-700"
          : "bg-blue-100 text-blue-700";

  return <span className={cn("inline-flex rounded px-2 py-1 text-[10px] font-bold", classes)}>{formatSeverity(severity)}</span>;
}

function ProjectStatusBadge({ status }: { status: ProjectStatus }) {
  const classes: Record<ProjectStatus, string> = {
    DRAFT: "bg-slate-100 text-slate-700",
    PLANNING: "bg-blue-100 text-blue-700",
    ACTIVE: "bg-emerald-100 text-emerald-700",
    ON_HOLD: "bg-amber-100 text-amber-700",
    COMPLETED: "bg-cyan-100 text-cyan-700",
    CANCELLED: "bg-red-100 text-red-700",
  };
  return <span className={cn("inline-flex rounded px-2 py-1 text-[10px] font-bold", classes[status])}>{formatStatus(status)}</span>;
}

function EmptyState({ text }: { text: string }) {
  return <div className="grid min-h-40 place-items-center px-4 text-center text-xs text-slate-500">{text}</div>;
}

function ChartSkeleton() {
  return <div className="mx-auto mt-5 h-36 w-36 rounded-full border-[20px] border-slate-100" />;
}

function ListLoading() {
  return (
    <div className="space-y-4 py-2">
      {[1, 2, 3, 4].map((item) => <div key={item} className="h-9 rounded-md bg-slate-100" />)}
    </div>
  );
}

function TableLoading({ columns }: { columns: number }) {
  return (
    <>
      {[1, 2, 3, 4].map((row) => (
        <tr key={row} className="border-b border-slate-100">
          {Array.from({ length: columns }, (_, column) => (
            <td key={column} className="py-3 pr-3"><div className="h-4 rounded-md bg-slate-100" /></td>
          ))}
        </tr>
      ))}
    </>
  );
}

function summarizePortfolio(projects: Project[], stats: DashboardStats) {
  const totalBudget = projects.reduce((sum, project) => sum + project.budget_total, 0);
  const averageProgress =
    projects.length > 0
      ? projects.reduce((sum, project) => sum + project.progress_pct, 0) / projects.length
      : 0;
  const criticalProjectIds = new Set(
    stats.early_warnings
      .filter((warning) => warning.severity === "HIGH" || warning.severity === "CRITICAL")
      .map((warning) => warning.project_id)
      .filter((projectId): projectId is string => Boolean(projectId))
  );
  const budgetRiskProjectIds = new Set(
    stats.early_warnings
      .filter((warning) => warning.type === "BUDGET_THRESHOLD")
      .map((warning) => warning.project_id)
      .filter((projectId): projectId is string => Boolean(projectId))
  );

  return {
    totalBudget,
    averageProgress,
    highWarnings: stats.early_warnings.filter((warning) => warning.severity === "HIGH" || warning.severity === "CRITICAL").length,
    criticalProjects: criticalProjectIds.size,
    atRiskBudget: projects
      .filter((project) => budgetRiskProjectIds.has(project.id))
      .reduce((sum, project) => sum + project.budget_total, 0),
  };
}

function summarizeWarnings(warnings: DashboardWarning[]) {
  const counts = new Map<string, number>();
  warnings.forEach((warning) => {
    const label = formatWarningType(warning.type);
    counts.set(label, (counts.get(label) ?? 0) + 1);
  });
  const tones = ["bg-red-500", "bg-orange-500", "bg-amber-400", "bg-blue-500"];
  return [...counts.entries()]
    .sort((a, b) => b[1] - a[1])
    .map(([label, value], index) => ({ label, value, tone: tones[index % tones.length] }));
}

function deduplicateWarnings(warnings: DashboardWarning[]) {
  const seenProjects = new Set<string>();

  return warnings.filter((warning) => {
    const key = warning.project_id ?? `${warning.type}:${warning.title}`;
    if (seenProjects.has(key)) return false;
    seenProjects.add(key);
    return true;
  });
}

function buildStatusData(projects: Project[]) {
  const counts = {
    Aktif: 0,
    "Dalam Perencanaan": 0,
    Ditunda: 0,
    Selesai: 0,
    Lainnya: 0,
  };
  projects.forEach((project) => {
    if (project.status === "ACTIVE") counts.Aktif += 1;
    else if (project.status === "DRAFT" || project.status === "PLANNING") counts["Dalam Perencanaan"] += 1;
    else if (project.status === "ON_HOLD") counts.Ditunda += 1;
    else if (project.status === "COMPLETED") counts.Selesai += 1;
    else counts.Lainnya += 1;
  });
  const total = projects.length;
  const legend = Object.entries(counts).map(([name, value]) => ({
    name,
    value,
    color: statusColors[name],
    percent: total > 0 ? Math.round((value / total) * 100) : 0,
  }));
  const nonZero = legend.filter((entry) => entry.value > 0);
  return {
    legend,
    chart: nonZero.length > 0 ? nonZero : [{ name: "Belum ada data", value: 1, color: "#cbd5e1", percent: 0 }],
  };
}

function rankProjects(projects: Project[]) {
  const priorityScore = { CRITICAL: 4, HIGH: 3, MEDIUM: 2, LOW: 1 };
  const statusScore: Record<ProjectStatus, number> = {
    ON_HOLD: 5,
    ACTIVE: 4,
    PLANNING: 3,
    DRAFT: 2,
    COMPLETED: 1,
    CANCELLED: 0,
  };
  return [...projects].sort((a, b) => {
    const scoreA = priorityScore[a.priority] * 100 + statusScore[a.status] * 10 + (100 - a.progress_pct);
    const scoreB = priorityScore[b.priority] * 100 + statusScore[b.status] * 10 + (100 - b.progress_pct);
    return scoreB - scoreA;
  });
}

function severityRank(severity: DashboardWarningSeverity) {
  return { LOW: 1, MEDIUM: 2, HIGH: 3, CRITICAL: 4 }[severity];
}

function severityDot(severity: DashboardWarningSeverity) {
  return {
    LOW: "bg-blue-500",
    MEDIUM: "bg-amber-400",
    HIGH: "bg-orange-500",
    CRITICAL: "bg-red-600",
  }[severity];
}

function formatWarningType(type: string) {
  const labels: Record<string, string> = {
    OVERDUE_TASK: "Task terlambat",
    OVERDUE_MILESTONE: "Milestone terlambat",
    LOW_PROGRESS_NEAR_END: "Progres rendah",
    BUDGET_THRESHOLD: "Ambang anggaran",
  };
  return labels[type] ?? type;
}

function formatSeverity(severity: DashboardWarningSeverity) {
  return { LOW: "Rendah", MEDIUM: "Sedang", HIGH: "Tinggi", CRITICAL: "Kritis" }[severity];
}

function formatStatus(status: ProjectStatus) {
  return {
    DRAFT: "Draft",
    PLANNING: "Perencanaan",
    ACTIVE: "Aktif",
    ON_HOLD: "Ditunda",
    COMPLETED: "Selesai",
    CANCELLED: "Dibatalkan",
  }[status];
}

function formatCompactCurrency(value: number) {
  if (value >= 1_000_000_000_000) return `Rp ${formatDecimal(value / 1_000_000_000_000)} T`;
  if (value >= 1_000_000_000) return `Rp ${formatDecimal(value / 1_000_000_000)} M`;
  if (value >= 1_000_000) return `Rp ${formatDecimal(value / 1_000_000)} Jt`;
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(value);
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("id-ID").format(value);
}

function formatDecimal(value: number) {
  return new Intl.NumberFormat("id-ID", { maximumFractionDigits: 1 }).format(value);
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short", year: "numeric" }).format(new Date(value));
}

function formatShortDate(value: string) {
  return new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short" }).format(new Date(value));
}

function clampPercent(value: number) {
  return Math.max(0, Math.min(100, value));
}
