"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  AlertTriangle,
  BarChart3,
  CheckCircle2,
  Clock,
  Crown,
  FileText,
  Loader2,
  ShieldAlert,
  TrendingUp,
  XCircle,
} from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { executiveService } from "@/services/executive.service";
import type {
  CriticalProject,
  DecisionItem,
  EscalationItem,
  ProgramKPISummary,
} from "@/types/executive";

// ── Helpers ───────────────────────────────────────────────────────────────────

function fmt(n: number): string {
  if (n >= 1_000_000_000_000) return `${(n / 1_000_000_000_000).toFixed(1)}T`;
  if (n >= 1_000_000_000) return `${(n / 1_000_000_000).toFixed(1)}M`;
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}Jt`;
  return n.toLocaleString("id-ID");
}

function healthBadge(hc: string) {
  const map: Record<string, string> = {
    GREEN: "bg-green-100 text-green-700",
    YELLOW: "bg-yellow-100 text-yellow-700",
    RED: "bg-red-100 text-red-700",
    CRITICAL: "bg-red-200 text-red-800 font-bold",
    UNSCORED: "bg-gray-100 text-gray-500",
  };
  const label: Record<string, string> = {
    GREEN: "Hijau",
    YELLOW: "Kuning",
    RED: "Merah",
    CRITICAL: "Kritis",
    UNSCORED: "Belum dinilai",
  };
  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${map[hc] ?? map.UNSCORED}`}
    >
      {label[hc] ?? hc}
    </span>
  );
}

// ── KPI Card ──────────────────────────────────────────────────────────────────

interface KpiCardProps {
  label: string;
  value: string | number;
  sub?: string;
  color?: string;
  icon: React.ReactNode;
}

function KpiCard({ label, value, sub, color = "bg-white", icon }: KpiCardProps) {
  return (
    <div className={`${color} rounded-xl border border-gray-200 p-4 shadow-sm`}>
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">
            {label}
          </p>
          <p className="mt-1 text-2xl font-bold text-gray-900">{value}</p>
          {sub && <p className="mt-0.5 text-xs text-gray-500">{sub}</p>}
        </div>
        <div className="text-gray-400">{icon}</div>
      </div>
    </div>
  );
}

// ── Health Bar ────────────────────────────────────────────────────────────────

function HealthBar({
  green, yellow, red, critical, unscored,
}: {
  green: number; yellow: number; red: number; critical: number; unscored: number;
}) {
  const total = green + yellow + red + critical + unscored || 1;
  const pct = (n: number) => `${((n / total) * 100).toFixed(1)}%`;
  return (
    <div className="flex flex-col gap-1">
      <div className="flex h-4 w-full overflow-hidden rounded-full bg-gray-100">
        {critical > 0 && (
          <div className="bg-red-700" style={{ width: pct(critical) }} title={`CRITICAL: ${critical}`} />
        )}
        {red > 0 && (
          <div className="bg-red-400" style={{ width: pct(red) }} title={`RED: ${red}`} />
        )}
        {yellow > 0 && (
          <div className="bg-yellow-400" style={{ width: pct(yellow) }} title={`YELLOW: ${yellow}`} />
        )}
        {green > 0 && (
          <div className="bg-green-400" style={{ width: pct(green) }} title={`GREEN: ${green}`} />
        )}
        {unscored > 0 && (
          <div className="bg-gray-300" style={{ width: pct(unscored) }} title={`Unscored: ${unscored}`} />
        )}
      </div>
      <div className="flex flex-wrap gap-2 text-xs text-gray-500">
        {critical > 0 && <span className="flex items-center gap-0.5"><span className="inline-block h-2 w-2 rounded-full bg-red-700" />{critical} Kritis</span>}
        {red > 0 && <span className="flex items-center gap-0.5"><span className="inline-block h-2 w-2 rounded-full bg-red-400" />{red} Merah</span>}
        {yellow > 0 && <span className="flex items-center gap-0.5"><span className="inline-block h-2 w-2 rounded-full bg-yellow-400" />{yellow} Kuning</span>}
        {green > 0 && <span className="flex items-center gap-0.5"><span className="inline-block h-2 w-2 rounded-full bg-green-400" />{green} Hijau</span>}
        {unscored > 0 && <span className="flex items-center gap-0.5"><span className="inline-block h-2 w-2 rounded-full bg-gray-300" />{unscored} Belum dinilai</span>}
      </div>
    </div>
  );
}

// ── Critical Projects Table ───────────────────────────────────────────────────

function CriticalProjectsTable({ projects }: { projects: CriticalProject[] }) {
  if (projects.length === 0)
    return <p className="py-6 text-center text-sm text-gray-500">Tidak ada proyek kritis.</p>;
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full text-sm">
        <thead>
          <tr className="border-b border-gray-100 bg-gray-50 text-left text-xs uppercase text-gray-500">
            <th className="px-3 py-2">Kode</th>
            <th className="px-3 py-2">Nama Proyek</th>
            <th className="px-3 py-2">Health</th>
            <th className="px-3 py-2 text-right">Fisik (%)</th>
            <th className="px-3 py-2 text-right">Anggaran</th>
            <th className="px-3 py-2 text-right">Risiko</th>
            <th className="px-3 py-2 text-right">Isu</th>
            <th className="px-3 py-2">Program</th>
          </tr>
        </thead>
        <tbody>
          {projects.map((p) => (
            <tr key={p.project_id} className="border-b border-gray-50 hover:bg-gray-50">
              <td className="px-3 py-2 font-mono text-xs text-gray-600">{p.project_code}</td>
              <td className="px-3 py-2 font-medium text-gray-900 max-w-[200px] truncate">{p.project_name}</td>
              <td className="px-3 py-2">{healthBadge(p.health_class)}</td>
              <td className="px-3 py-2 text-right">{p.physical_actual.toFixed(1)}</td>
              <td className="px-3 py-2 text-right text-gray-600">Rp {fmt(p.budget_total)}</td>
              <td className="px-3 py-2 text-right">
                <span className={p.open_risks > 0 ? "text-red-600 font-medium" : "text-gray-400"}>{p.open_risks}</span>
              </td>
              <td className="px-3 py-2 text-right">
                <span className={p.open_issues > 0 ? "text-orange-600 font-medium" : "text-gray-400"}>{p.open_issues}</span>
              </td>
              <td className="px-3 py-2 text-xs text-gray-500 truncate max-w-[120px]">{p.program_name || "—"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ── Escalation List ───────────────────────────────────────────────────────────

function EscalationList({ items }: { items: EscalationItem[] }) {
  if (items.length === 0)
    return <p className="py-6 text-center text-sm text-gray-500">Tidak ada eskalasi aktif.</p>;
  return (
    <ul className="divide-y divide-gray-100">
      {items.map((e) => (
        <li key={e.id} className="flex items-start gap-3 py-3">
          <ShieldAlert className="mt-0.5 h-4 w-4 shrink-0 text-red-500" />
          <div className="min-w-0 flex-1">
            <p className="text-sm font-medium text-gray-900 line-clamp-2">{e.reason}</p>
            <div className="mt-0.5 flex flex-wrap gap-2 text-xs text-gray-500">
              {e.project_name && <span>{e.project_name}</span>}
              <span className="rounded-full bg-gray-100 px-1.5 py-0.5">{e.level}</span>
              <span className="rounded-full bg-gray-100 px-1.5 py-0.5">{e.source_type}</span>
              <span className={`rounded-full px-1.5 py-0.5 ${e.status === "OPEN" ? "bg-red-100 text-red-700" : "bg-yellow-100 text-yellow-700"}`}>{e.status}</span>
            </div>
          </div>
          <span className="shrink-0 text-xs text-gray-400">
            {new Date(e.created_at).toLocaleDateString("id-ID")}
          </span>
        </li>
      ))}
    </ul>
  );
}

// ── Decision Queue ────────────────────────────────────────────────────────────

function DecisionQueue({ items }: { items: DecisionItem[] }) {
  if (items.length === 0)
    return <p className="py-6 text-center text-sm text-gray-500">Tidak ada keputusan pending.</p>;
  return (
    <ul className="divide-y divide-gray-100">
      {items.map((d) => (
        <li key={d.id} className="flex items-start gap-3 py-3">
          {d.is_overdue ? (
            <XCircle className="mt-0.5 h-4 w-4 shrink-0 text-red-500" />
          ) : (
            <Clock className="mt-0.5 h-4 w-4 shrink-0 text-yellow-500" />
          )}
          <div className="min-w-0 flex-1">
            <p className="text-sm font-medium text-gray-900 line-clamp-1">{d.subject}</p>
            <p className="mt-0.5 text-xs text-gray-500 line-clamp-2">{d.decision_text}</p>
            <div className="mt-1 flex flex-wrap gap-2 text-xs">
              {d.project_name && <span className="text-gray-500">{d.project_name}</span>}
              {d.due_date && (
                <span className={d.is_overdue ? "text-red-600 font-medium" : "text-gray-500"}>
                  Jatuh tempo: {d.due_date}
                </span>
              )}
            </div>
          </div>
        </li>
      ))}
    </ul>
  );
}

// ── Program Table ─────────────────────────────────────────────────────────────

function ProgramTable({ programs }: { programs: ProgramKPISummary[] }) {
  if (programs.length === 0)
    return <p className="py-6 text-center text-sm text-gray-500">Tidak ada data program.</p>;
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full text-sm">
        <thead>
          <tr className="border-b border-gray-100 bg-gray-50 text-left text-xs uppercase text-gray-500">
            <th className="px-3 py-2">Program</th>
            <th className="px-3 py-2 text-right">Proyek</th>
            <th className="px-3 py-2 text-right">Anggaran</th>
            <th className="px-3 py-2">Health</th>
            <th className="px-3 py-2 text-right">Risiko</th>
            <th className="px-3 py-2 text-right">Isu</th>
          </tr>
        </thead>
        <tbody>
          {programs.map((pg) => {
            const total = (pg.health_green + pg.health_yellow + pg.health_red + pg.health_critical) || 1;
            return (
              <tr key={pg.program_id} className="border-b border-gray-50 hover:bg-gray-50">
                <td className="px-3 py-2">
                  <div className="font-medium text-gray-900">{pg.program_name}</div>
                  <div className="text-xs text-gray-400">{pg.program_code}</div>
                </td>
                <td className="px-3 py-2 text-right">
                  <span className="font-medium">{pg.active_projects}</span>
                  <span className="text-gray-400">/{pg.total_projects}</span>
                </td>
                <td className="px-3 py-2 text-right text-gray-600">Rp {fmt(pg.total_budget)}</td>
                <td className="px-3 py-2">
                  <div className="flex h-2 w-24 overflow-hidden rounded-full bg-gray-100">
                    {pg.health_critical > 0 && <div className="bg-red-700" style={{ width: `${(pg.health_critical / total) * 100}%` }} />}
                    {pg.health_red > 0 && <div className="bg-red-400" style={{ width: `${(pg.health_red / total) * 100}%` }} />}
                    {pg.health_yellow > 0 && <div className="bg-yellow-400" style={{ width: `${(pg.health_yellow / total) * 100}%` }} />}
                    {pg.health_green > 0 && <div className="bg-green-400" style={{ width: `${(pg.health_green / total) * 100}%` }} />}
                  </div>
                </td>
                <td className="px-3 py-2 text-right">
                  <span className={pg.open_risks > 0 ? "text-red-600 font-medium" : "text-gray-400"}>{pg.open_risks}</span>
                </td>
                <td className="px-3 py-2 text-right">
                  <span className={pg.open_issues > 0 ? "text-orange-600 font-medium" : "text-gray-400"}>{pg.open_issues}</span>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function ExecutiveDashboardPage() {
  const [year, setYear] = useState<number | undefined>(undefined);
  const [month, setMonth] = useState<number | undefined>(undefined);

  const { data, isLoading, isError } = useQuery({
    queryKey: ["executive-dashboard", year, month],
    queryFn: () => executiveService.getDashboard({ year, month }),
  });

  const currentYear = new Date().getFullYear();
  const years = [currentYear - 1, currentYear, currentYear + 1];

  return (
    <DashboardLayout title="Executive Dashboard">
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-2 text-xs text-gray-500">
          <Crown className="h-4 w-4 text-yellow-500" />
          <span>
            Level 1 — Ringkasan Nasional
            {data?.as_of ? ` · per ${new Date(data.as_of).toLocaleString("id-ID")}` : ""}
          </span>
        </div>

        <div className="flex items-center gap-2">
          <select
            className="rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
            value={year ?? ""}
            onChange={(e) => setYear(e.target.value ? Number(e.target.value) : undefined)}
          >
            <option value="">Semua Tahun</option>
            {years.map((y) => (
              <option key={y} value={y}>{y}</option>
            ))}
          </select>
          <select
            className="rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
            value={month ?? ""}
            onChange={(e) => setMonth(e.target.value ? Number(e.target.value) : undefined)}
          >
            <option value="">Semua Bulan</option>
            {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
              <option key={m} value={m}>
                {new Date(2000, m - 1).toLocaleString("id-ID", { month: "long" })}
              </option>
            ))}
          </select>
        </div>
      </div>

      {isLoading && (
        <div className="flex h-64 items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
        </div>
      )}

      {isError && (
        <div className="rounded-xl border border-red-200 bg-red-50 p-6 text-center text-sm text-red-600">
          Gagal memuat data dashboard. Silakan coba lagi.
        </div>
      )}

      {data && (
        <div className="space-y-6">
          {/* ── KPI Hero Cards ──────────────────────────────────────────── */}
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
            <KpiCard
              label="Total Proyek"
              value={data.summary.total_projects}
              sub={`${data.summary.active_projects} aktif`}
              icon={<BarChart3 className="h-5 w-5" />}
            />
            <KpiCard
              label="Total Anggaran"
              value={`Rp ${fmt(data.summary.total_budget)}`}
              icon={<TrendingUp className="h-5 w-5" />}
            />
            <KpiCard
              label="Eskalasi Aktif"
              value={data.summary.open_escalations}
              sub={`${data.summary.pending_decisions} keputusan tertunda`}
              color={data.summary.open_escalations > 0 ? "bg-red-50" : "bg-white"}
              icon={<ShieldAlert className="h-5 w-5" />}
            />
            <KpiCard
              label="Risiko Terbuka"
              value={data.summary.open_risks}
              sub={`${data.summary.high_risks} tinggi/kritis`}
              color={data.summary.high_risks > 0 ? "bg-orange-50" : "bg-white"}
              icon={<AlertTriangle className="h-5 w-5" />}
            />
            <KpiCard
              label="Indikator Manfaat"
              value={data.summary.benefit_indicators}
              icon={<CheckCircle2 className="h-5 w-5" />}
            />
          </div>

          {/* ── Health Distribution ─────────────────────────────────────── */}
          <div className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm">
            <h2 className="mb-3 text-sm font-semibold text-gray-700">Distribusi Health Proyek</h2>
            <HealthBar
              green={data.summary.health_green}
              yellow={data.summary.health_yellow}
              red={data.summary.health_red}
              critical={data.summary.health_critical}
              unscored={data.summary.health_unscored}
            />
          </div>

          {/* ── Critical Projects + Escalations row ────────────────────── */}
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            {/* Critical Projects */}
            <div className="rounded-xl border border-gray-200 bg-white shadow-sm">
              <div className="flex items-center gap-2 border-b border-gray-100 px-4 py-3">
                <AlertTriangle className="h-4 w-4 text-red-500" />
                <h2 className="text-sm font-semibold text-gray-700">
                  Proyek Kritis ({data.critical_projects.length})
                </h2>
              </div>
              <div className="p-4">
                <CriticalProjectsTable projects={data.critical_projects} />
              </div>
            </div>

            {/* Escalations */}
            <div className="rounded-xl border border-gray-200 bg-white shadow-sm">
              <div className="flex items-center gap-2 border-b border-gray-100 px-4 py-3">
                <ShieldAlert className="h-4 w-4 text-orange-500" />
                <h2 className="text-sm font-semibold text-gray-700">
                  Eskalasi Aktif ({data.escalations.length})
                </h2>
              </div>
              <div className="max-h-72 overflow-y-auto px-4">
                <EscalationList items={data.escalations} />
              </div>
            </div>
          </div>

          {/* ── Decisions + Benefits row ────────────────────────────────── */}
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            {/* Pending Decisions */}
            <div className="rounded-xl border border-gray-200 bg-white shadow-sm">
              <div className="flex items-center gap-2 border-b border-gray-100 px-4 py-3">
                <FileText className="h-4 w-4 text-blue-500" />
                <h2 className="text-sm font-semibold text-gray-700">
                  Antrean Keputusan ({data.pending_decisions.length})
                </h2>
              </div>
              <div className="max-h-72 overflow-y-auto px-4">
                <DecisionQueue items={data.pending_decisions} />
              </div>
            </div>

            {/* Benefits */}
            <div className="rounded-xl border border-gray-200 bg-white shadow-sm">
              <div className="flex items-center gap-2 border-b border-gray-100 px-4 py-3">
                <CheckCircle2 className="h-4 w-4 text-green-500" />
                <h2 className="text-sm font-semibold text-gray-700">
                  Indikator Manfaat ({data.benefits.total_indicators})
                </h2>
              </div>
              <div className="p-4">
                {data.benefits.total_indicators === 0 ? (
                  <p className="py-4 text-center text-sm text-gray-500">Belum ada indikator manfaat.</p>
                ) : (
                  <>
                    <div className="mb-3 flex gap-4 text-sm">
                      <span className="flex items-center gap-1 text-green-600">
                        <CheckCircle2 className="h-3.5 w-3.5" />
                        On track: {data.benefits.on_track_count}
                      </span>
                      <span className="flex items-center gap-1 text-red-600">
                        <XCircle className="h-3.5 w-3.5" />
                        Di bawah target: {data.benefits.behind_count}
                      </span>
                    </div>
                    <ul className="divide-y divide-gray-100">
                      {data.benefits.indicators.map((b) => (
                        <li key={b.id} className="py-2">
                          <div className="flex items-center justify-between">
                            <span className="text-sm font-medium text-gray-800">{b.name}</span>
                            <span className={`text-xs font-semibold ${b.achievement_pct >= 80 ? "text-green-600" : "text-red-600"}`}>
                              {b.achievement_pct.toFixed(0)}%
                            </span>
                          </div>
                          <div className="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-gray-100">
                            <div
                              className={`h-full rounded-full ${b.achievement_pct >= 80 ? "bg-green-400" : "bg-red-400"}`}
                              style={{ width: `${Math.min(b.achievement_pct, 100)}%` }}
                            />
                          </div>
                          <p className="mt-0.5 text-xs text-gray-400">
                            {b.actual_value.toLocaleString("id-ID")} / {b.target_value.toLocaleString("id-ID")} {b.unit}
                          </p>
                        </li>
                      ))}
                    </ul>
                  </>
                )}
              </div>
            </div>
          </div>

          {/* ── Program Comparison ──────────────────────────────────────── */}
          <div className="rounded-xl border border-gray-200 bg-white shadow-sm">
            <div className="flex items-center gap-2 border-b border-gray-100 px-4 py-3">
              <BarChart3 className="h-4 w-4 text-indigo-500" />
              <h2 className="text-sm font-semibold text-gray-700">
                Perbandingan Program ({data.programs.length})
              </h2>
            </div>
            <div className="p-4">
              <ProgramTable programs={data.programs} />
            </div>
          </div>
        </div>
      )}
    </DashboardLayout>
  );
}
