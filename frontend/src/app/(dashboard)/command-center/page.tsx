"use client";

import Image from "next/image";
import Link from "next/link";
import { useMemo } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AlertTriangle,
  ArrowRight,
  BriefcaseBusiness,
  CalendarDays,
  CircleAlert,
  Clock3,
  FileCheck2,
  Filter,
  RefreshCw,
  ShieldAlert,
  type LucideIcon,
} from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { dashboardService } from "@/services/dashboard.service";
import { projectService } from "@/services/project.service";
import { commandCenterService } from "@/services/commandcenter.service";
import { useAuthStore } from "@/store/auth.store";
import type {
  DashboardStats,
  DashboardWarning,
  DashboardWarningSeverity,
  DashboardWarningType,
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

const severityOrder: Record<DashboardWarningSeverity, number> = {
  CRITICAL: 4,
  HIGH: 3,
  MEDIUM: 2,
  LOW: 1,
};

export default function CommandCenterPage() {
  const queryClient = useQueryClient();
  const accessToken = useAuthStore((state) => state.accessToken);
  const statsQuery = useQuery({
    queryKey: ["command-center", "stats"],
    queryFn: async () => (await dashboardService.getStats()).data.data,
    enabled: Boolean(accessToken),
    staleTime: 30_000,
  });
  const projectsQuery = useQuery({
    queryKey: ["command-center", "projects"],
    queryFn: async () =>
      (await projectService.list({ page: 1, page_size: 100, sort_by: "updated_at", sort_dir: "desc" })).data.data,
    enabled: Boolean(accessToken),
    staleTime: 30_000,
  });
  const commandQuery = useQuery({
    queryKey: ["command-center", "summary"],
    queryFn: () => commandCenterService.getSummary(),
    enabled: Boolean(accessToken),
    staleTime: 30_000,
  });

  const stats = statsQuery.data ?? emptyStats;
  const projects = projectsQuery.data ?? emptyProjects;
  const loading = !accessToken || statsQuery.isLoading || projectsQuery.isLoading || commandQuery.isLoading;
  const fetching = statsQuery.isFetching || projectsQuery.isFetching || commandQuery.isFetching;
  const command = commandQuery.data;
  const updateEscalation = useMutation({ mutationFn: ({ id, status }: { id: string; status: "ACKNOWLEDGED" | "CLOSED" }) => commandCenterService.updateEscalationStatus(id, status), onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["command-center", "summary"] }) });
  const updateDecision = useMutation({ mutationFn: ({ id, status }: { id: string; status: "IN_PROGRESS" | "COMPLETED" | "CANCELLED" }) => commandCenterService.updateDecisionStatus(id, status), onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["command-center", "summary"] }) });
  const warnings = useMemo(
    () => [...stats.early_warnings].sort((a, b) => severityOrder[b.severity] - severityOrder[a.severity]),
    [stats.early_warnings]
  );
  const uniquePriorityWarnings = useMemo(() => uniqueByProject(warnings), [warnings]);
  const riskRegisterWarnings = useMemo(
    () => warnings.filter((warning) => warning.type === "RISK_REGISTER").slice(0, 5),
    [warnings]
  );
  const priorityRiskWarnings = useMemo(() => {
    if (riskRegisterWarnings.length > 0) {
      return riskRegisterWarnings;
    }
    return uniquePriorityWarnings.slice(0, 5);
  }, [riskRegisterWarnings, uniquePriorityWarnings]);
  const watchlist = useMemo(() => rankWatchlist(projects, warnings).slice(0, 8), [projects, warnings]);
  const issueSummary = useMemo(() => summarizeWarningTypes(warnings), [warnings]);
  const highRiskProjects = new Set(
    warnings
      .filter((warning) => warning.severity === "HIGH" || warning.severity === "CRITICAL")
      .map((warning) => warning.project_id)
      .filter((id): id is string => Boolean(id))
  ).size;

  const refresh = () => void Promise.all([statsQuery.refetch(), projectsQuery.refetch()]);
  const kpis: CommandKpi[] = [
    {
      label: "Total Proyek Dipantau",
      value: formatNumber(stats.total_projects),
      detail: `${stats.active_projects} proyek aktif`,
      icon: BriefcaseBusiness,
      tone: "blue",
    },
    {
      label: "Alert Aktif",
      value: formatNumber(warnings.length),
      detail: "Dihasilkan rule dashboard",
      icon: AlertTriangle,
      tone: "red",
    },
    {
      label: "Validasi Tertunda",
      value: formatNumber(command?.validations.filter((item) => item.status === "SUBMITTED").length ?? 0),
      detail: "Submission menunggu keputusan",
      icon: FileCheck2,
      tone: "amber",
    },
    {
      label: "Tindak Lanjut Overdue",
      value: formatNumber(command?.actions.filter((item) => item.aging_days && item.aging_days > 0).length ?? 0),
      detail: "Corrective action melewati target",
      icon: Clock3,
      tone: "red",
    },
    {
      label: "Risiko Tinggi",
      value: formatNumber(highRiskProjects),
      detail: "Proyek dengan alert prioritas",
      icon: ShieldAlert,
      tone: "red",
    },
    {
      label: "Isu Kritis",
      value: formatNumber(warnings.filter((warning) => warning.severity === "CRITICAL").length),
      detail: "Memerlukan eskalasi",
      icon: CircleAlert,
      tone: "red",
    },
  ];

  return (
    <DashboardLayout title="PMO Command Center">
      <div className="mb-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="text-[10px] font-semibold uppercase text-[#0b5aa2]">Operasional dan eskalasi</p>
          <p className="mt-1 text-xs text-slate-500">Ringkasan alert, kualitas data, tindak lanjut, dan proyek prioritas.</p>
        </div>
        <div className="flex flex-col gap-2 sm:flex-row">
          <label className="flex h-9 items-center gap-2 rounded-md border border-slate-200 bg-white px-3 text-xs text-slate-600">
            <span className="font-semibold text-[#082e63]">Periode:</span>
            <input type="month" defaultValue={currentMonth()} className="bg-transparent outline-none" aria-label="Periode data" />
          </label>
          <button
            type="button"
            onClick={refresh}
            disabled={!accessToken || fetching}
            className="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-slate-200 bg-white px-3 text-xs font-semibold text-[#0b5aa2] hover:bg-blue-50 disabled:opacity-50"
          >
            <Filter className="h-4 w-4" aria-hidden="true" />
            Filter
            <RefreshCw className={cn("h-3.5 w-3.5", fetching && "animate-spin")} aria-hidden="true" />
          </button>
        </div>
      </div>

      <section className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3 min-[2200px]:grid-cols-6" aria-label="KPI Command Center">
        {kpis.map((kpi) => <KpiCard key={kpi.label} data={kpi} loading={loading} />)}
      </section>

      <section className="mt-3 grid grid-cols-1 gap-3 2xl:grid-cols-12">
        <Panel title="Alert Center" action={{ label: "Lihat proyek", href: "/projects" }} className="2xl:col-span-4">
          <AlertTable warnings={warnings.slice(0, 5)} loading={loading} />
        </Panel>
        <Panel title="Pending Validation" className="2xl:col-span-4">
          <ValidationTable items={command?.validations ?? []} loading={loading} />
        </Panel>
        <Panel title="Corrective Action Tracker" action={{ label: "Lihat proyek", href: "/projects" }} className="2xl:col-span-4">
          <CommandActionTable items={command?.actions ?? []} loading={loading} />
        </Panel>
      </section>

      <section className="mt-3 grid grid-cols-1 gap-3 xl:grid-cols-2 2xl:grid-cols-12">
        <Panel title="Risiko Prioritas" action={{ label: "Lihat proyek", href: "/projects" }} className="2xl:col-span-3">
          <p className="mb-2 text-[10px] text-slate-400">
            {riskRegisterWarnings.length > 0
              ? "Dari risk register proyek (score probability × impact)."
              : "Derived dari early warning (risk register belum memiliki data terbuka)."}
          </p>
          <RankedWarnings warnings={priorityRiskWarnings} />
        </Panel>
        <Panel title="Isu Utama" action={{ label: "Lihat proyek", href: "/projects" }} className="2xl:col-span-3">
          <IssueSummary rows={issueSummary} total={warnings.length} />
        </Panel>
        <Panel title="Data Quality & Freshness" className="2xl:col-span-3">
          <QualityTable projects={projects.slice(0, 5)} />
        </Panel>
        <Panel title="Heatmap / Peta Hotspot Proyek" action={{ label: "Buka GIS", href: "/gis" }} className="2xl:col-span-3">
          <div className="relative min-h-64 overflow-hidden rounded-md border border-slate-100 bg-[#dceffc]">
            <Image src="/images/indonesia-project-map.png" alt="Peta hotspot proyek Indonesia" fill sizes="(min-width: 1536px) 25vw, 100vw" className="object-cover" />
            <div className="absolute inset-x-3 bottom-3 rounded-md border border-white/80 bg-white/90 px-3 py-2 text-center text-[9px] text-slate-600">
              Titik bersifat ilustratif sampai koordinat proyek tersedia pada P2-008.
            </div>
          </div>
        </Panel>
      </section>

      <section className="mt-3 grid grid-cols-1 gap-3 xl:grid-cols-2">
        <Panel title="Jadwal Pelaporan">
          <ReportingSchedule />
        </Panel>
        <Panel title="Kalender Pelaporan">
          <ReportingCalendar />
        </Panel>
      </section>

      <section className="mt-3">
        <Panel title="Watchlist Proyek Prioritas" action={{ label: "Lihat semua proyek", href: "/projects" }}>
          <WatchlistTable projects={watchlist} warnings={warnings} loading={loading} />
        </Panel>
      </section>

      <section className="mt-3 grid grid-cols-1 gap-3 xl:grid-cols-2">
        <Panel title="Escalation Aktif"><CommandActionTable items={command?.escalations ?? []} loading={loading} onUpdate={(item) => updateEscalation.mutate({ id: item.id, status: item.status === "OPEN" ? "ACKNOWLEDGED" : "CLOSED" })} /></Panel>
        <Panel title="Executive Decision Follow-up"><CommandActionTable items={command?.decisions ?? []} loading={loading} onUpdate={(item) => updateDecision.mutate({ id: item.id, status: item.status === "OPEN" ? "IN_PROGRESS" : "COMPLETED" })} /></Panel>
      </section>
    </DashboardLayout>
  );
}

interface CommandKpi {
  label: string;
  value: string;
  detail: string;
  icon: LucideIcon;
  tone: "blue" | "red" | "amber";
}

function KpiCard({ data, loading }: { data: CommandKpi; loading: boolean }) {
  const Icon = data.icon;
  const tone = {
    blue: "bg-blue-50 text-blue-700",
    red: "bg-red-50 text-red-600",
    amber: "bg-amber-50 text-amber-600",
  }[data.tone];

  return (
    <article className="grid min-h-32 grid-cols-[auto_minmax(0,1fr)] items-center gap-4 rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <div className={cn("grid h-12 w-12 place-items-center rounded-md", tone)}><Icon className="h-6 w-6" aria-hidden="true" /></div>
      <div className="min-w-0">
        <p className="min-h-8 text-[10px] font-semibold uppercase leading-4 text-slate-500">{data.label}</p>
        {loading ? <div className="my-2 h-7 w-20 rounded bg-slate-100" /> : <p className="text-2xl font-bold leading-none text-[#082e63] tabular-nums">{data.value}</p>}
        <p className="mt-2 min-h-8 text-[10px] leading-4 text-slate-500">{data.detail}</p>
      </div>
    </article>
  );
}

function Panel({ title, children, action, className }: { title: string; children: React.ReactNode; action?: { label: string; href: string }; className?: string }) {
  return (
    <article className={cn("min-w-0 rounded-lg border border-slate-200 bg-white p-4 shadow-sm", className)}>
      <div className="mb-3 flex min-h-9 items-start justify-between gap-4 border-b border-slate-200 pb-2">
        <h2 className="min-w-0 text-xs font-bold uppercase leading-5 text-[#082e63]">{title}</h2>
        {action && <Link href={action.href} className="inline-flex shrink-0 items-center gap-1 whitespace-nowrap text-[10px] font-semibold leading-5 text-[#0b5aa2] hover:underline">{action.label}<ArrowRight className="h-3 w-3" aria-hidden="true" /></Link>}
      </div>
      {children}
    </article>
  );
}

function AlertTable({ warnings, loading }: { warnings: DashboardWarning[]; loading: boolean }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[620px] text-left text-[10px]">
        <thead><tr className="border-b border-slate-200 uppercase text-slate-500"><th className="pb-2">Proyek</th><th className="pb-2">Jenis Alert</th><th className="pb-2">Severity</th><th className="pb-2">Tenggat</th><th className="pb-2">Status</th></tr></thead>
        <tbody>
          {loading ? <LoadingRows columns={5} /> : warnings.length > 0 ? warnings.map((warning) => (
            <tr key={warning.id} className="border-b border-slate-100 last:border-0">
              <td className="max-w-40 py-2.5 pr-3"><ProjectLink warning={warning} /></td>
              <td className="py-2.5 pr-3 text-slate-600">{warningTypeLabel(warning.type)}</td>
              <td className="py-2.5 pr-3"><SeverityBadge severity={warning.severity} /></td>
              <td className="py-2.5 pr-3 text-slate-500">{warning.due_date ? formatDate(warning.due_date) : "-"}</td>
              <td className="py-2.5"><StatusBadge label="Aktif" tone="blue" /></td>
            </tr>
          )) : <EmptyTable columns={5} text="Tidak ada alert aktif." />}
        </tbody>
      </table>
    </div>
  );
}

function ActionTable({ warnings, projects, loading }: { warnings: DashboardWarning[]; projects: Project[]; loading: boolean }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[620px] text-left text-[10px]">
        <thead><tr className="border-b border-slate-200 uppercase text-slate-500"><th className="pb-2">Tindak Lanjut</th><th className="pb-2">Proyek</th><th className="pb-2">Progres</th><th className="pb-2">Status</th></tr></thead>
        <tbody>
          {loading ? <LoadingRows columns={4} /> : warnings.length > 0 ? warnings.map((warning) => {
            const project = projects.find((item) => item.id === warning.project_id);
            const progress = project?.progress_pct ?? 0;
            return (
              <tr key={warning.id} className="border-b border-slate-100 last:border-0">
                <td className="max-w-52 py-2.5 pr-3 text-slate-600">{warning.message}</td>
                <td className="max-w-40 py-2.5 pr-3"><ProjectLink warning={warning} /></td>
                <td className="py-2.5 pr-3"><ProgressBar value={progress} /></td>
                <td className="py-2.5"><StatusBadge label="Perlu aksi" tone="amber" /></td>
              </tr>
            );
          }) : <EmptyTable columns={4} text="Tidak ada tindak lanjut prioritas." />}
        </tbody>
      </table>
    </div>
  );
}

function ValidationTable({ items, loading }: { items: import("@/types/commandcenter").CommandItem[]; loading: boolean }) {
  return <div className="overflow-x-auto"><table className="w-full min-w-[520px] text-left text-[10px]"><thead><tr className="border-b border-slate-200 uppercase text-slate-500"><th className="pb-2">Project</th><th className="pb-2">Status</th><th className="pb-2">SLA</th><th className="pb-2">Aging</th></tr></thead><tbody>{loading ? <LoadingRows columns={4} /> : items.length > 0 ? items.map((item) => <tr key={item.id} className="border-b border-slate-100 last:border-0"><td className="py-2.5 pr-3"><Link href={item.project_id ? `/projects/${item.project_id}` : "/projects"} className="font-semibold text-[#0b4c91] hover:underline">{item.project_name ?? "Project"}</Link></td><td className="py-2.5 pr-3"><StatusBadge label={item.status} tone={item.status === "SUBMITTED" ? "amber" : "blue"} /></td><td className="py-2.5 pr-3 text-slate-500">{item.due_at ? formatDate(item.due_at) : "-"}</td><td className="py-2.5 text-slate-600">{item.aging_days ?? 0} hari</td></tr>) : <EmptyTable columns={4} text="Tidak ada validasi tertunda." />}</tbody></table></div>;
}

function CommandActionTable({ items, loading, onUpdate }: { items: import("@/types/commandcenter").CommandItem[]; loading: boolean; onUpdate?: (item: import("@/types/commandcenter").CommandItem) => void }) {
  return <div className="overflow-x-auto"><table className="w-full min-w-[620px] text-left text-[10px]"><thead><tr className="border-b border-slate-200 uppercase text-slate-500"><th className="pb-2">Tindak lanjut</th><th className="pb-2">Project</th><th className="pb-2">PIC</th><th className="pb-2">Target</th><th className="pb-2">Action</th></tr></thead><tbody>{loading ? <LoadingRows columns={5} /> : items.length > 0 ? items.map((item) => <tr key={item.id} className="border-b border-slate-100 last:border-0"><td className="max-w-52 py-2.5 pr-3 text-slate-600">{item.title}</td><td className="py-2.5 pr-3"><Link href={item.project_id ? `/projects/${item.project_id}` : "/projects"} className="font-semibold text-[#0b4c91] hover:underline">{item.project_name ?? "Project"}</Link></td><td className="py-2.5 pr-3 font-mono text-slate-500">{item.pic_user_id ? item.pic_user_id.slice(0, 8) : "-"}</td><td className="py-2.5 text-slate-500">{item.due_at ? formatDate(item.due_at) : "-"}</td><td className="py-2.5">{onUpdate && <button type="button" onClick={() => onUpdate(item)} className="rounded border px-2 py-1 text-[9px] font-semibold hover:bg-blue-50">{item.status === "OPEN" ? "Start" : "Close"}</button>}</td></tr>) : <EmptyTable columns={5} text="Tidak ada item terbuka." />}</tbody></table></div>;
}

function ModuleReadiness({ icon: Icon, title, description, ticket }: { icon: LucideIcon; title: string; description: string; ticket: string }) {
  return (
    <div className="grid min-h-64 place-items-center rounded-md border border-dashed border-slate-200 bg-slate-50 px-6 text-center">
      <div>
        <div className="mx-auto grid h-12 w-12 place-items-center rounded-md bg-amber-50 text-amber-600"><Icon className="h-6 w-6" aria-hidden="true" /></div>
        <p className="mt-4 text-sm font-semibold text-[#17345d]">{title}</p>
        <p className="mx-auto mt-2 max-w-md text-[10px] leading-5 text-slate-500">{description}</p>
        <span className="mt-4 inline-flex rounded bg-white px-2 py-1 font-mono text-[9px] text-slate-500 shadow-sm">{ticket}</span>
      </div>
    </div>
  );
}

function RankedWarnings({ warnings }: { warnings: DashboardWarning[] }) {
  return <div>{warnings.length > 0 ? warnings.map((warning, index) => <Link key={warning.id} href={warning.project_id ? `/projects/${warning.project_id}` : "/projects"} className="grid grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-2 border-b border-slate-100 py-3 text-[10px] last:border-0 hover:bg-blue-50/40"><span className={cn("grid h-5 w-5 place-items-center rounded-full font-bold text-white", severityColor(warning.severity))}>{index + 1}</span><span className="min-w-0"><span className="line-clamp-1 font-semibold text-[#17345d]">{warningTypeLabel(warning.type)}</span><span className="mt-0.5 block truncate text-[9px] text-slate-500">{warning.project_name ?? "Proyek"}</span></span><SeverityBadge severity={warning.severity} /></Link>) : <EmptyState text="Tidak ada risiko prioritas." />}</div>;
}

function IssueSummary({ rows, total }: { rows: Array<{ label: string; value: number; tone: string }>; total: number }) {
  return <div>{rows.map((row, index) => <div key={row.label} className="grid grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-2 border-b border-slate-100 py-3 text-[10px]"><span className={cn("grid h-5 w-5 place-items-center rounded-full font-bold text-white", row.tone)}>{index + 1}</span><span className="min-w-0 truncate text-slate-600">{row.label}</span><span className="font-bold text-[#082e63]">{row.value}</span></div>)}<div className="mt-3 flex justify-between border-t border-slate-200 pt-3 text-[10px]"><span className="text-slate-500">Total isu aktif</span><strong className="text-[#082e63]">{total}</strong></div></div>;
}

function QualityTable({ projects }: { projects: Project[] }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[420px] text-left text-[10px]">
        <thead><tr className="border-b border-slate-200 uppercase text-slate-500"><th className="pb-2">Proyek</th><th className="pb-2">Kelengkapan</th><th className="pb-2">Freshness</th><th className="pb-2">Status</th></tr></thead>
        <tbody>{projects.map((project) => {
          const completeness = projectCompleteness(project);
          const age = daysSince(project.updated_at);
          return <tr key={project.id} className="border-b border-slate-100 last:border-0"><td className="max-w-36 py-2.5 pr-3"><Link href={`/projects/${project.id}`} className="line-clamp-1 font-semibold text-[#0b4c91] hover:underline">{project.name}</Link></td><td className="py-2.5 pr-3 font-semibold text-slate-600">{completeness}%</td><td className="py-2.5 pr-3 text-slate-500">{age} hari</td><td className="py-2.5"><StatusBadge label={completeness >= 80 ? "Baik" : "Perlu cek"} tone={completeness >= 80 ? "green" : "amber"} /></td></tr>;
        })}</tbody>
      </table>
      <p className="mt-3 text-[9px] leading-4 text-slate-400">Kelengkapan dihitung dari field proyek yang tersedia; validasi resmi menunggu P1-012.</p>
    </div>
  );
}

function ReportingSchedule() {
  const rows = [
    ["Laporan Mingguan", "Senin berikutnya", "Proyek prioritas dan kritis", "Terjadwal"],
    ["Laporan Bulanan", "Tanggal 5", "Seluruh proyek dipantau", "Terjadwal"],
    ["Laporan Triwulanan", "Akhir triwulan", "Evaluasi kinerja dan keuangan", "Mendatang"],
    ["Laporan Semesteran", "Akhir semester", "Capaian dan rekap semester", "Mendatang"],
  ];
  return <div className="overflow-x-auto"><table className="w-full min-w-[620px] text-left text-[10px]"><thead><tr className="border-b border-slate-200 uppercase text-slate-500"><th className="pb-2">Frekuensi</th><th className="pb-2">Periode Berikutnya</th><th className="pb-2">Cakupan</th><th className="pb-2">Status</th></tr></thead><tbody>{rows.map((row) => <tr key={row[0]} className="border-b border-slate-100 last:border-0"><td className="py-3 font-semibold text-[#17345d]">{row[0]}</td><td className="py-3 text-slate-600">{row[1]}</td><td className="py-3 text-slate-600">{row[2]}</td><td className="py-3"><StatusBadge label={row[3]} tone={row[3] === "Terjadwal" ? "green" : "blue"} /></td></tr>)}</tbody></table></div>;
}

function ReportingCalendar() {
  const days = Array.from({ length: 31 }, (_, index) => index + 1);
  return <div className="grid min-h-48 grid-cols-[minmax(0,1fr)_160px] gap-5 max-sm:grid-cols-1"><div><div className="mb-3 flex items-center justify-between"><p className="text-xs font-semibold text-[#17345d]">{new Intl.DateTimeFormat("id-ID", { month: "long", year: "numeric" }).format(new Date())}</p><CalendarDays className="h-4 w-4 text-[#0b5aa2]" aria-hidden="true" /></div><div className="grid grid-cols-7 gap-1 text-center text-[9px]">{["Sen", "Sel", "Rab", "Kam", "Jum", "Sab", "Min"].map((day) => <span key={day} className="pb-1 font-semibold text-slate-400">{day}</span>)}{days.map((day) => <span key={day} className={cn("grid h-6 place-items-center rounded text-slate-600", [5, 15, 26].includes(day) && "bg-[#0b5aa2] font-bold text-white")}>{day}</span>)}</div></div><div className="space-y-3 border-l border-slate-200 pl-5 max-sm:border-l-0 max-sm:border-t max-sm:pl-0 max-sm:pt-4"><Legend color="bg-emerald-500" label="Mingguan" /><Legend color="bg-blue-600" label="Bulanan" /><Legend color="bg-violet-500" label="Triwulanan" /><Legend color="bg-orange-500" label="Semesteran" /></div></div>;
}

function WatchlistTable({ projects, warnings, loading }: { projects: Project[]; warnings: DashboardWarning[]; loading: boolean }) {
  return <div className="overflow-x-auto"><table className="w-full min-w-[1180px] text-left text-[10px]"><thead><tr className="border-b border-slate-200 uppercase text-slate-500"><th className="pb-2">No.</th><th className="pb-2">Proyek</th><th className="pb-2">Kode</th><th className="pb-2">Progres Fisik</th><th className="pb-2">Target</th><th className="pb-2">Risiko Utama</th><th className="pb-2">Tindak Lanjut</th><th className="pb-2">Status</th></tr></thead><tbody>{loading ? <LoadingRows columns={8} /> : projects.map((project, index) => { const warning = warnings.find((item) => item.project_id === project.id); return <tr key={project.id} className="border-b border-slate-100 last:border-0 hover:bg-blue-50/40"><td className="py-3 pr-3 text-slate-400">{index + 1}</td><td className="max-w-56 py-3 pr-4"><Link href={`/projects/${project.id}`} className="font-semibold text-[#0b4c91] hover:underline">{project.name}</Link></td><td className="py-3 pr-4 font-mono text-slate-500">{project.code}</td><td className="py-3 pr-4"><ProgressBar value={project.progress_pct} /></td><td className="py-3 pr-4 text-slate-600">{project.end_date ? formatDate(project.end_date) : "-"}</td><td className="max-w-44 py-3 pr-4 text-slate-600">{warning ? warningTypeLabel(warning.type) : "-"}</td><td className="max-w-64 py-3 pr-4 text-slate-600">{warning?.message ?? "Belum ada tindak lanjut"}</td><td className="py-3"><ProjectStatusBadge status={project.status} /></td></tr>; })}</tbody></table></div>;
}

function ProjectLink({ warning }: { warning: DashboardWarning }) {
  return <Link href={warning.project_id ? `/projects/${warning.project_id}` : "/projects"} className="line-clamp-2 font-semibold leading-4 text-[#0b4c91] hover:underline">{warning.project_name ?? warning.title}</Link>;
}

function SeverityBadge({ severity }: { severity: DashboardWarningSeverity }) {
  const tone = severity === "CRITICAL" ? "bg-red-100 text-red-700" : severity === "HIGH" ? "bg-orange-100 text-orange-700" : severity === "MEDIUM" ? "bg-amber-100 text-amber-700" : "bg-blue-100 text-blue-700";
  const label = severity === "CRITICAL" ? "Kritis" : severity === "HIGH" ? "Tinggi" : severity === "MEDIUM" ? "Sedang" : "Rendah";
  return <span className={cn("inline-flex shrink-0 rounded px-2 py-1 text-[9px] font-bold", tone)}>{label}</span>;
}

function StatusBadge({ label, tone }: { label: string; tone: "blue" | "green" | "amber" }) {
  const classes = { blue: "bg-blue-100 text-blue-700", green: "bg-emerald-100 text-emerald-700", amber: "bg-amber-100 text-amber-700" }[tone];
  return <span className={cn("inline-flex whitespace-nowrap rounded px-2 py-1 text-[9px] font-bold", classes)}>{label}</span>;
}

function ProjectStatusBadge({ status }: { status: ProjectStatus }) {
  const label: Record<ProjectStatus, string> = { DRAFT: "Draft", PLANNING: "Perencanaan", ACTIVE: "Aktif", ON_HOLD: "Ditunda", COMPLETED: "Selesai", CANCELLED: "Dibatalkan" };
  return <StatusBadge label={label[status]} tone={status === "ACTIVE" || status === "COMPLETED" ? "green" : status === "ON_HOLD" || status === "CANCELLED" ? "amber" : "blue"} />;
}

function ProgressBar({ value }: { value: number }) {
  const safe = Math.max(0, Math.min(value, 100));
  return <div className="flex min-w-28 items-center gap-2"><div className="h-2 flex-1 overflow-hidden rounded-full bg-slate-100"><div className="h-full rounded-full bg-[#1262b8]" style={{ width: `${safe}%` }} /></div><span className="w-9 text-right font-semibold text-slate-600">{formatDecimal(safe)}%</span></div>;
}

function Legend({ color, label }: { color: string; label: string }) { return <div className="flex items-center gap-2 text-[10px] text-slate-600"><span className={cn("h-2.5 w-2.5 rounded-sm", color)} />{label}</div>; }
function EmptyState({ text }: { text: string }) { return <div className="grid min-h-48 place-items-center text-center text-xs text-slate-500">{text}</div>; }
function EmptyTable({ columns, text }: { columns: number; text: string }) { return <tr><td colSpan={columns} className="py-12 text-center text-xs text-slate-500">{text}</td></tr>; }
function LoadingRows({ columns }: { columns: number }) { return <>{[1, 2, 3, 4, 5].map((row) => <tr key={row} className="border-b border-slate-100">{Array.from({ length: columns }, (_, column) => <td key={column} className="py-3 pr-3"><div className="h-4 rounded bg-slate-100" /></td>)}</tr>)}</>; }

function uniqueByProject(warnings: DashboardWarning[]) {
  const seen = new Set<string>();
  return warnings.filter((warning) => { const key = warning.project_id ?? warning.id; if (seen.has(key)) return false; seen.add(key); return true; });
}

function rankWatchlist(projects: Project[], warnings: DashboardWarning[]) {
  const warningScore = new Map<string, number>();
  warnings.forEach((warning) => { if (warning.project_id) warningScore.set(warning.project_id, Math.max(warningScore.get(warning.project_id) ?? 0, severityOrder[warning.severity])); });
  return [...projects].sort((a, b) => ((warningScore.get(b.id) ?? 0) * 100 + (100 - b.progress_pct)) - ((warningScore.get(a.id) ?? 0) * 100 + (100 - a.progress_pct)));
}

function summarizeWarningTypes(warnings: DashboardWarning[]) {
  const counts = new Map<DashboardWarningType, number>();
  warnings.forEach((warning) => counts.set(warning.type, (counts.get(warning.type) ?? 0) + 1));
  const tones = ["bg-red-500", "bg-orange-500", "bg-amber-400", "bg-blue-500"];
  return [...counts.entries()].sort((a, b) => b[1] - a[1]).map(([type, value], index) => ({ label: warningTypeLabel(type), value, tone: tones[index % tones.length] }));
}

function projectCompleteness(project: Project) {
  const fields = [project.code, project.name, project.description, project.objectives, project.category, project.start_date, project.end_date, project.budget_total > 0 ? project.budget_total : null, project.manager_id];
  return Math.round((fields.filter(Boolean).length / fields.length) * 100);
}

function warningTypeLabel(type: DashboardWarningType) {
  const labels: Record<DashboardWarningType, string> = { OVERDUE_TASK: "Task terlambat", OVERDUE_MILESTONE: "Milestone terlambat", LOW_PROGRESS_NEAR_END: "Progres rendah", BUDGET_THRESHOLD: "Ambang anggaran", RISK_REGISTER: "Risiko terbuka" };
  return labels[type];
}

function severityColor(severity: DashboardWarningSeverity) { return severity === "CRITICAL" ? "bg-red-600" : severity === "HIGH" ? "bg-orange-500" : severity === "MEDIUM" ? "bg-amber-500" : "bg-blue-500"; }
function formatNumber(value: number) { return new Intl.NumberFormat("id-ID").format(value); }
function formatDecimal(value: number) { return new Intl.NumberFormat("id-ID", { maximumFractionDigits: 1 }).format(value); }
function formatDate(value: string) { const date = new Date(value); return Number.isNaN(date.getTime()) ? "-" : new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short", year: "numeric" }).format(date); }
function daysSince(value: string) { const date = new Date(value); if (Number.isNaN(date.getTime())) return 0; return Math.max(0, Math.floor((Date.now() - date.getTime()) / 86_400_000)); }
function currentMonth() { return new Date().toISOString().slice(0, 7); }
