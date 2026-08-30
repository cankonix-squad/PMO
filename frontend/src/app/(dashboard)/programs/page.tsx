"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  AlertTriangle,
  BarChart3,
  Building2,
  ChevronRight,
  Layers3,
  RefreshCw,
  TrendingDown,
  TrendingUp,
  Wallet,
} from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { analyticsService } from "@/services/analytics.service";
import { cn } from "@/lib/utils";
import type { ProgramDashboard, ProgramKPI, ProjectRow, TopDeviation } from "@/types/analytics";
import { HEALTH_CLASS_COLOR, HEALTH_CLASS_LABEL } from "@/types/analytics";

// ── Helpers ───────────────────────────────────────────────────────────────────

function fmt(n: number) {
  return new Intl.NumberFormat("id-ID").format(Math.round(n));
}

function fmtBudget(n: number) {
  if (n >= 1_000_000_000_000) return `Rp ${(n / 1_000_000_000_000).toFixed(1)} T`;
  if (n >= 1_000_000_000) return `Rp ${(n / 1_000_000_000).toFixed(1)} M`;
  if (n >= 1_000_000) return `Rp ${(n / 1_000_000).toFixed(0)} Jt`;
  return `Rp ${fmt(n)}`;
}

function fmtPct(n: number) {
  return `${n >= 0 ? "+" : ""}${n.toFixed(1)}%`;
}

// ── Small components ──────────────────────────────────────────────────────────

function HealthDot({ cls }: { cls: string }) {
  const colors: Record<string, string> = {
    GREEN: "bg-emerald-500",
    YELLOW: "bg-yellow-400",
    RED: "bg-rose-500",
    CRITICAL: "bg-red-700",
    "": "bg-gray-300",
  };
  return (
    <span
      className={cn("inline-block h-2.5 w-2.5 rounded-full", colors[cls] ?? "bg-gray-300")}
    />
  );
}

function HealthBadge({ cls }: { cls: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium",
        HEALTH_CLASS_COLOR[cls] ?? HEALTH_CLASS_COLOR[""]
      )}
    >
      <HealthDot cls={cls} />
      {HEALTH_CLASS_LABEL[cls] ?? cls}
    </span>
  );
}

function ProgressBar({ actual, target }: { actual: number; target: number }) {
  const pct = Math.min(100, actual);
  const color =
    actual >= target ? "bg-emerald-500" : actual >= target - 5 ? "bg-yellow-400" : "bg-rose-500";
  return (
    <div className="space-y-0.5">
      <div className="h-2 w-full overflow-hidden rounded-full bg-slate-200">
        <div className={cn("h-full rounded-full", color)} style={{ width: `${pct}%` }} />
      </div>
      <div className="flex justify-between text-[11px] text-slate-500">
        <span>Aktual {actual.toFixed(1)}%</span>
        <span>Target {target.toFixed(1)}%</span>
      </div>
    </div>
  );
}

// ── KPI summary card ──────────────────────────────────────────────────────────

function KPICard({
  kpi,
  onClick,
  selected,
}: {
  kpi: ProgramKPI;
  onClick: () => void;
  selected: boolean;
}) {
  const healthTotal =
    kpi.health_green + kpi.health_yellow + kpi.health_red + kpi.health_critical;
  const variance = kpi.avg_physical_actual - kpi.avg_physical_target;

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "w-full rounded-xl border p-4 text-left transition-all hover:shadow-md",
        selected
          ? "border-blue-500 bg-blue-50 shadow-md"
          : "border-slate-200 bg-white hover:border-slate-300"
      )}
    >
      <div className="mb-3 flex items-start justify-between gap-2">
        <div>
          <p className="text-[11px] font-medium uppercase tracking-wider text-slate-400">
            {kpi.group_code}
          </p>
          <p className="mt-0.5 font-semibold text-slate-800 leading-snug">{kpi.group_name}</p>
        </div>
        <span className="shrink-0 rounded-full bg-blue-100 px-2 py-0.5 text-xs font-semibold text-blue-700">
          {kpi.total_projects} Proyek
        </span>
      </div>

      {/* Budget */}
      <div className="mb-3 space-y-1">
        <div className="flex items-center justify-between text-xs text-slate-500">
          <span className="flex items-center gap-1">
            <Wallet className="h-3 w-3" /> Anggaran
          </span>
          <span className="font-medium text-slate-700">{fmtBudget(kpi.total_budget)}</span>
        </div>
        <div className="h-1.5 w-full overflow-hidden rounded-full bg-slate-200">
          <div
            className={cn(
              "h-full rounded-full",
              kpi.budget_usage_pct > 90
                ? "bg-rose-500"
                : kpi.budget_usage_pct > 70
                  ? "bg-yellow-400"
                  : "bg-emerald-500"
            )}
            style={{ width: `${Math.min(100, kpi.budget_usage_pct)}%` }}
          />
        </div>
        <p className="text-right text-[11px] text-slate-500">
          {kpi.budget_usage_pct.toFixed(1)}% terpakai
        </p>
      </div>

      {/* Physical progress */}
      <div className="mb-3">
        <ProgressBar actual={kpi.avg_physical_actual} target={kpi.avg_physical_target} />
      </div>

      {/* Health distribution */}
      {healthTotal > 0 && (
        <div className="mb-3 flex gap-0.5 overflow-hidden rounded-full">
          {kpi.health_green > 0 && (
            <div
              className="h-2 bg-emerald-500"
              style={{ width: `${(kpi.health_green / healthTotal) * 100}%` }}
              title={`Baik: ${kpi.health_green}`}
            />
          )}
          {kpi.health_yellow > 0 && (
            <div
              className="h-2 bg-yellow-400"
              style={{ width: `${(kpi.health_yellow / healthTotal) * 100}%` }}
              title={`Perlu Perhatian: ${kpi.health_yellow}`}
            />
          )}
          {kpi.health_red > 0 && (
            <div
              className="h-2 bg-rose-500"
              style={{ width: `${(kpi.health_red / healthTotal) * 100}%` }}
              title={`Berisiko: ${kpi.health_red}`}
            />
          )}
          {kpi.health_critical > 0 && (
            <div
              className="h-2 bg-red-700"
              style={{ width: `${(kpi.health_critical / healthTotal) * 100}%` }}
              title={`Kritis: ${kpi.health_critical}`}
            />
          )}
        </div>
      )}

      {/* Risks & issues */}
      <div className="flex items-center justify-between text-xs">
        <span
          className={cn(
            "flex items-center gap-1",
            kpi.high_risks > 0 ? "text-rose-600 font-medium" : "text-slate-500"
          )}
        >
          <AlertTriangle className="h-3 w-3" />
          {kpi.open_risks} risiko
        </span>
        <span
          className={cn(
            "flex items-center gap-1",
            variance < 0 ? "text-rose-600 font-medium" : "text-emerald-600 font-medium"
          )}
        >
          {variance < 0 ? (
            <TrendingDown className="h-3 w-3" />
          ) : (
            <TrendingUp className="h-3 w-3" />
          )}
          {fmtPct(variance)}
        </span>
      </div>
    </button>
  );
}

// ── Project table ─────────────────────────────────────────────────────────────

function ProjectTable({ projects }: { projects: ProjectRow[] }) {
  if (projects.length === 0) {
    return (
      <p className="py-8 text-center text-sm text-slate-400">Belum ada proyek dalam program ini.</p>
    );
  }
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-slate-200 text-left text-xs font-medium uppercase tracking-wider text-slate-400">
            <th className="pb-2 pr-3">Proyek</th>
            <th className="pb-2 pr-3">Status</th>
            <th className="pb-2 pr-3 text-right">Fisik Aktual</th>
            <th className="pb-2 pr-3 text-right">Deviasi</th>
            <th className="pb-2 pr-3 text-right">Anggaran</th>
            <th className="pb-2 pr-3">Health</th>
            <th className="pb-2 pr-3 text-right">Risiko</th>
            <th className="pb-2 text-right">Isu</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100">
          {projects.map((p) => (
            <tr key={p.project_id} className="hover:bg-slate-50">
              <td className="py-2 pr-3">
                <p className="font-medium text-slate-800">{p.project_name}</p>
                <p className="text-[11px] text-slate-400">{p.project_code}</p>
              </td>
              <td className="py-2 pr-3">
                <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-600">
                  {p.status}
                </span>
              </td>
              <td className="py-2 pr-3 text-right tabular-nums">
                {p.physical_actual.toFixed(1)}%
              </td>
              <td
                className={cn(
                  "py-2 pr-3 text-right tabular-nums font-medium",
                  p.physical_variance < 0 ? "text-rose-600" : "text-emerald-600"
                )}
              >
                {fmtPct(p.physical_variance)}
              </td>
              <td className="py-2 pr-3 text-right tabular-nums text-slate-600">
                {fmtBudget(p.budget_total)}
              </td>
              <td className="py-2 pr-3">
                <HealthBadge cls={p.health_class} />
              </td>
              <td
                className={cn(
                  "py-2 pr-3 text-right tabular-nums",
                  p.open_risks > 0 ? "font-medium text-rose-600" : "text-slate-500"
                )}
              >
                {p.open_risks}
              </td>
              <td
                className={cn(
                  "py-2 text-right tabular-nums",
                  p.open_issues > 0 ? "font-medium text-orange-600" : "text-slate-500"
                )}
              >
                {p.open_issues}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ── Deviation list ────────────────────────────────────────────────────────────

function DeviationList({ items, title }: { items: TopDeviation[]; title: string }) {
  if (items.length === 0) return null;
  return (
    <div>
      <h4 className="mb-2 text-xs font-semibold uppercase tracking-wider text-slate-400">
        {title}
      </h4>
      <ul className="space-y-2">
        {items.map((d) => (
          <li key={d.project_id} className="flex items-center justify-between gap-2 text-sm">
            <span className="truncate text-slate-700">{d.project_name}</span>
            <span className="shrink-0 font-semibold text-rose-600">{d.label}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}

// ── Detail panel ──────────────────────────────────────────────────────────────

function DetailPanel({ dashboard }: { dashboard: ProgramDashboard }) {
  const { kpi } = dashboard;
  return (
    <div className="space-y-6">
      {/* KPI metrics row */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <div className="rounded-lg border border-slate-200 bg-white p-3">
          <p className="text-xs text-slate-400">Total Proyek</p>
          <p className="text-2xl font-bold text-slate-800">{kpi.total_projects}</p>
          <p className="text-xs text-slate-500">{kpi.active_projects} aktif</p>
        </div>
        <div className="rounded-lg border border-slate-200 bg-white p-3">
          <p className="text-xs text-slate-400">Total Anggaran</p>
          <p className="text-xl font-bold text-slate-800">{fmtBudget(kpi.total_budget)}</p>
          <p className="text-xs text-slate-500">{kpi.budget_usage_pct.toFixed(1)}% terpakai</p>
        </div>
        <div className="rounded-lg border border-slate-200 bg-white p-3">
          <p className="text-xs text-slate-400">Progress Fisik</p>
          <p className="text-2xl font-bold text-slate-800">
            {kpi.avg_physical_actual.toFixed(1)}%
          </p>
          <p
            className={cn(
              "text-xs font-medium",
              kpi.physical_variance < 0 ? "text-rose-600" : "text-emerald-600"
            )}
          >
            {fmtPct(kpi.physical_variance)} vs target
          </p>
        </div>
        <div className="rounded-lg border border-slate-200 bg-white p-3">
          <p className="text-xs text-slate-400">Open Risiko</p>
          <p className="text-2xl font-bold text-slate-800">{kpi.open_risks}</p>
          <p className={cn("text-xs", kpi.high_risks > 0 ? "text-rose-600 font-medium" : "text-slate-500")}>
            {kpi.high_risks} risiko tinggi
          </p>
        </div>
      </div>

      {/* Health distribution */}
      <div className="rounded-lg border border-slate-200 bg-white p-4">
        <h3 className="mb-3 text-sm font-semibold text-slate-700">Distribusi Health</h3>
        <div className="flex flex-wrap gap-3">
          {[
            { label: "Baik", count: kpi.health_green, color: "bg-emerald-500" },
            { label: "Perlu Perhatian", count: kpi.health_yellow, color: "bg-yellow-400" },
            { label: "Berisiko", count: kpi.health_red, color: "bg-rose-500" },
            { label: "Kritis", count: kpi.health_critical, color: "bg-red-700" },
            { label: "Belum Dinilai", count: kpi.health_unscored, color: "bg-gray-300" },
          ].map(({ label, count, color }) => (
            <div key={label} className="flex items-center gap-2">
              <span className={cn("h-3 w-3 rounded-full", color)} />
              <span className="text-sm text-slate-600">
                {label}: <strong>{count}</strong>
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Top deviations */}
      {(dashboard.top_physical_deviation.length > 0 ||
        dashboard.top_budget_deviation.length > 0) && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          {dashboard.top_physical_deviation.length > 0 && (
            <div className="rounded-lg border border-slate-200 bg-white p-4">
              <DeviationList
                items={dashboard.top_physical_deviation}
                title="Top Deviasi Fisik"
              />
            </div>
          )}
          {dashboard.top_budget_deviation.length > 0 && (
            <div className="rounded-lg border border-slate-200 bg-white p-4">
              <DeviationList
                items={dashboard.top_budget_deviation}
                title="Top Utilisasi Anggaran"
              />
            </div>
          )}
        </div>
      )}

      {/* High risk projects */}
      {dashboard.high_risk_projects.length > 0 && (
        <div className="rounded-lg border border-rose-100 bg-white p-4">
          <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold text-rose-700">
            <AlertTriangle className="h-4 w-4" />
            Proyek Berisiko Tinggi
          </h3>
          <ProjectTable projects={dashboard.high_risk_projects} />
        </div>
      )}

      {/* All projects */}
      <div className="rounded-lg border border-slate-200 bg-white p-4">
        <h3 className="mb-3 text-sm font-semibold text-slate-700">
          Semua Proyek ({dashboard.projects.length})
        </h3>
        <ProjectTable projects={dashboard.projects} />
      </div>
    </div>
  );
}

// ── Tab toggle ────────────────────────────────────────────────────────────────

type Tab = "program" | "sector";

// ── Main page ─────────────────────────────────────────────────────────────────

export default function ProgramsPage() {
  const [tab, setTab] = useState<Tab>("program");
  const [selectedID, setSelectedID] = useState<string | null>(null);

  const listQuery = useQuery({
    queryKey: ["analytics", tab],
    queryFn: () =>
      tab === "program" ? analyticsService.listPrograms() : analyticsService.listSectors(),
  });

  const detailQuery = useQuery({
    queryKey: ["analytics", tab, selectedID],
    queryFn: () =>
      selectedID
        ? tab === "program"
          ? analyticsService.getProgram(selectedID)
          : analyticsService.getSector(selectedID)
        : Promise.resolve(null),
    enabled: !!selectedID,
  });

  const groups = listQuery.data?.groups ?? [];

  // Auto-select first on list load
  const handleTabChange = (t: Tab) => {
    setTab(t);
    setSelectedID(null);
  };

  const handleSelect = (id: string) => {
    setSelectedID((prev) => (prev === id ? null : id));
  };

  return (
    <DashboardLayout title="Program Dashboard">
      <div className="space-y-4 px-4 py-6 sm:px-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-slate-800">Program Dashboard</h1>
            <p className="mt-0.5 text-sm text-slate-500">
              Agregasi KPI per program dan sektor — Level 2 PMO Control Tower
            </p>
          </div>
          <button
            type="button"
            onClick={() => listQuery.refetch()}
            disabled={listQuery.isFetching}
            className="flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-sm text-slate-600 transition hover:bg-slate-50 disabled:opacity-50"
          >
            <RefreshCw className={cn("h-4 w-4", listQuery.isFetching && "animate-spin")} />
            Muat Ulang
          </button>
        </div>

        {/* Tab */}
        <div className="flex gap-1 rounded-lg border border-slate-200 bg-slate-100 p-1 w-fit">
          {(["program", "sector"] as Tab[]).map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => handleTabChange(t)}
              className={cn(
                "rounded-md px-4 py-1.5 text-sm font-medium transition",
                tab === t
                  ? "bg-white text-blue-700 shadow-sm"
                  : "text-slate-500 hover:text-slate-700"
              )}
            >
              {t === "program" ? (
                <span className="flex items-center gap-1.5">
                  <Layers3 className="h-4 w-4" /> Program
                </span>
              ) : (
                <span className="flex items-center gap-1.5">
                  <BarChart3 className="h-4 w-4" /> Sektor
                </span>
              )}
            </button>
          ))}
        </div>

        {/* Error */}
        {listQuery.isError && (
          <div className="rounded-lg border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">
            Gagal memuat data. Coba refresh halaman.
          </div>
        )}

        {/* Loading skeleton */}
        {listQuery.isLoading && (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
            {Array.from({ length: 5 }).map((_, i) => (
              <div
                key={i}
                className="h-52 animate-pulse rounded-xl border border-slate-200 bg-slate-100"
              />
            ))}
          </div>
        )}

        {/* Main content */}
        {!listQuery.isLoading && (
          <div className="flex flex-col gap-4 lg:flex-row">
            {/* Left: KPI cards grid */}
            <div
              className={cn(
                "grid gap-3 content-start",
                selectedID
                  ? "grid-cols-1 sm:grid-cols-2 lg:w-[420px] lg:shrink-0"
                  : "grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 w-full"
              )}
            >
              {groups.length === 0 && !listQuery.isLoading && (
                <div className="col-span-full rounded-xl border border-slate-200 bg-white p-8 text-center text-slate-400">
                  <Building2 className="mx-auto mb-2 h-8 w-8 opacity-30" />
                  <p className="text-sm">
                    Belum ada {tab === "program" ? "program" : "sektor"} yang terdaftar.
                  </p>
                </div>
              )}
              {groups.map((kpi) => (
                <KPICard
                  key={kpi.group_id}
                  kpi={kpi}
                  selected={selectedID === kpi.group_id}
                  onClick={() => handleSelect(kpi.group_id)}
                />
              ))}
            </div>

            {/* Right: Detail panel */}
            {selectedID && (
              <div className="min-w-0 flex-1">
                {detailQuery.isLoading && (
                  <div className="space-y-3">
                    {Array.from({ length: 3 }).map((_, i) => (
                      <div
                        key={i}
                        className="h-32 animate-pulse rounded-xl border border-slate-200 bg-slate-100"
                      />
                    ))}
                  </div>
                )}
                {detailQuery.isError && (
                  <div className="rounded-lg border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">
                    Gagal memuat detail program.
                  </div>
                )}
                {detailQuery.data && (
                  <div>
                    <div className="mb-4 flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => setSelectedID(null)}
                        className="text-xs text-slate-400 hover:text-slate-600"
                      >
                        ← Kembali
                      </button>
                      <ChevronRight className="h-3 w-3 text-slate-300" />
                      <h2 className="font-semibold text-slate-800">
                        {detailQuery.data.kpi.group_name}
                      </h2>
                    </div>
                    <DetailPanel dashboard={detailQuery.data} />
                  </div>
                )}
              </div>
            )}
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}
