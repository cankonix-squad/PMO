"use client";

import { cn, formatDate } from "@/lib/utils";
import type { Baseline, Snapshot, SnapshotStatus } from "@/types/monitoring";
import { MONTH_NAMES, SNAPSHOT_STATUS_COLOR, SNAPSHOT_STATUS_LABEL } from "@/types/monitoring";
import { CheckCircle2, Edit3, Plus, Trash2 } from "lucide-react";
import { primaryButton } from "../schemas";
import { SnapshotActions } from "./SnapshotActions";

export function MonitoringPanel({
  baselines,
  snapshots,
  isLoadingBaselines,
  isLoadingSnapshots,
  isErrorBaselines,
  isErrorSnapshots,
  statusFilter,
  onStatusFilterChange,
  onCreateBaseline,
  onEditBaseline,
  onDeleteBaseline,
  onCreateSnapshot,
  onEditSnapshot,
  onTransitionSnapshot,
  onRejectSnapshot,
  onDeleteSnapshot,
}: {
  baselines: Baseline[];
  snapshots: Snapshot[];
  isLoadingBaselines: boolean;
  isLoadingSnapshots: boolean;
  isErrorBaselines: boolean;
  isErrorSnapshots: boolean;
  statusFilter: SnapshotStatus | "";
  onStatusFilterChange: (v: SnapshotStatus | "") => void;
  onCreateBaseline: () => void;
  onEditBaseline: (b: Baseline) => void;
  onDeleteBaseline: (b: Baseline) => void;
  onCreateSnapshot: () => void;
  onEditSnapshot: (s: Snapshot) => void;
  onTransitionSnapshot: (s: Snapshot, status: SnapshotStatus) => void;
  onRejectSnapshot: (s: Snapshot, reason: string) => void;
  onDeleteSnapshot: (s: Snapshot) => void;
}) {
  const filteredSnapshots =
    statusFilter
      ? snapshots.filter((s) => s.status === statusFilter)
      : snapshots;

  const activeBaseline = baselines.find((b) => b.is_active);

  return (
    <section className="space-y-6">
      {/* ── Header ── */}
      <div className="flex items-center justify-between">
        <h3 className="text-base font-semibold text-foreground">Monitoring</h3>
      </div>

      {/* ── Baselines ── */}
      <div className="rounded-lg border border-border bg-card p-5 space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-semibold text-foreground">Baseline</p>
            <p className="text-xs text-muted-foreground mt-0.5">
              Rencana resmi proyek yang digunakan sebagai acuan snapshot
            </p>
          </div>
          <button onClick={onCreateBaseline} className={primaryButton}>
            <Plus className="h-3.5 w-3.5" />
            Tambah Baseline
          </button>
        </div>

        {isErrorBaselines && (
          <p className="text-sm text-destructive">Gagal memuat baseline.</p>
        )}

        {isLoadingBaselines ? (
          <p className="text-sm text-muted-foreground">Memuat…</p>
        ) : baselines.length === 0 ? (
          <p className="text-sm text-muted-foreground italic">
            Belum ada baseline. Buat baseline pertama untuk mulai monitoring.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left text-xs text-muted-foreground">
                  <th className="pb-2 pr-4 font-medium">Ver.</th>
                  <th className="pb-2 pr-4 font-medium">Label</th>
                  <th className="pb-2 pr-4 font-medium">Mulai Rencana</th>
                  <th className="pb-2 pr-4 font-medium">Selesai Rencana</th>
                  <th className="pb-2 pr-4 font-medium">Target Fisik (%)</th>
                  <th className="pb-2 pr-4 font-medium">Anggaran</th>
                  <th className="pb-2 pr-4 font-medium">Status</th>
                  <th className="pb-2 font-medium" />
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {baselines.map((b) => (
                  <tr key={b.id} className="hover:bg-muted/30">
                    <td className="py-2.5 pr-4 font-mono text-xs">v{b.version}</td>
                    <td className="py-2.5 pr-4">{b.label ?? "—"}</td>
                    <td className="py-2.5 pr-4">{formatDate(b.planned_start)}</td>
                    <td className="py-2.5 pr-4">{formatDate(b.planned_end)}</td>
                    <td className="py-2.5 pr-4">{b.physical_target}%</td>
                    <td className="py-2.5 pr-4">
                      {b.budget_total.toLocaleString("id-ID")} {b.currency}
                    </td>
                    <td className="py-2.5 pr-4">
                      {b.is_active ? (
                        <span className="inline-flex items-center gap-1 rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
                          <CheckCircle2 className="h-3 w-3" /> Aktif
                        </span>
                      ) : (
                        <span className="inline-flex items-center rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
                          Tidak aktif
                        </span>
                      )}
                    </td>
                    <td className="py-2.5">
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => onEditBaseline(b)}
                          className="text-muted-foreground hover:text-foreground"
                          title="Edit"
                        >
                          <Edit3 className="h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => onDeleteBaseline(b)}
                          className="text-muted-foreground hover:text-destructive"
                          title="Hapus"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* ── Snapshots ── */}
      <div className="rounded-lg border border-border bg-card p-5 space-y-4">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          <div>
            <p className="text-sm font-semibold text-foreground">Snapshot Periodik</p>
            <p className="text-xs text-muted-foreground mt-0.5">
              Realisasi fisik & keuangan per periode terhadap baseline
              {activeBaseline ? ` (Baseline: v${activeBaseline.version}${activeBaseline.label ? ` — ${activeBaseline.label}` : ""})` : ""}
            </p>
          </div>
          <div className="flex items-center gap-3">
            <select
              value={statusFilter}
              onChange={(e) => onStatusFilterChange(e.target.value as SnapshotStatus | "")}
              className="rounded-md border border-input bg-background px-3 py-1.5 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="">Semua Status</option>
              {(["DRAFT", "SUBMITTED", "VALID", "REJECTED", "STALE"] as SnapshotStatus[]).map((s) => (
                <option key={s} value={s}>{SNAPSHOT_STATUS_LABEL[s]}</option>
              ))}
            </select>
            <button onClick={onCreateSnapshot} className={primaryButton}>
              <Plus className="h-3.5 w-3.5" />
              Tambah Snapshot
            </button>
          </div>
        </div>

        {isErrorSnapshots && (
          <p className="text-sm text-destructive">Gagal memuat snapshot.</p>
        )}

        {isLoadingSnapshots ? (
          <p className="text-sm text-muted-foreground">Memuat…</p>
        ) : filteredSnapshots.length === 0 ? (
          <p className="text-sm text-muted-foreground italic">
            {statusFilter ? `Tidak ada snapshot dengan status ${SNAPSHOT_STATUS_LABEL[statusFilter]}.` : "Belum ada snapshot periodik."}
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left text-xs text-muted-foreground">
                  <th className="pb-2 pr-3 font-medium">Periode</th>
                  <th className="pb-2 pr-3 font-medium">Fisik Rencana</th>
                  <th className="pb-2 pr-3 font-medium">Fisik Aktual</th>
                  <th className="pb-2 pr-3 font-medium">Deviasi Fisik</th>
                  <th className="pb-2 pr-3 font-medium">Keuangan Rencana</th>
                  <th className="pb-2 pr-3 font-medium">Keuangan Aktual</th>
                  <th className="pb-2 pr-3 font-medium">Deviasi Hari</th>
                  <th className="pb-2 pr-3 font-medium">Status</th>
                  <th className="pb-2 font-medium" />
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filteredSnapshots.map((s) => (
                  <tr key={s.id} className="hover:bg-muted/30">
                    <td className="py-2.5 pr-3 font-medium whitespace-nowrap">
                      {MONTH_NAMES[s.period_month]} {s.period_year}
                    </td>
                    <td className="py-2.5 pr-3">{s.physical_target}%</td>
                    <td className="py-2.5 pr-3">{s.physical_actual}%</td>
                    <td className={cn("py-2.5 pr-3 font-medium", s.physical_variance >= 0 ? "text-green-600" : "text-destructive")}>
                      {s.physical_variance >= 0 ? "+" : ""}{s.physical_variance.toFixed(2)}%
                    </td>
                    <td className="py-2.5 pr-3">
                      {s.financial_target.toLocaleString("id-ID")}
                    </td>
                    <td className="py-2.5 pr-3">
                      {s.financial_actual.toLocaleString("id-ID")}
                    </td>
                    <td className={cn("py-2.5 pr-3", (s.schedule_deviation_days ?? 0) > 0 ? "text-destructive" : "text-green-600")}>
                      {s.schedule_deviation_days != null
                        ? `${s.schedule_deviation_days > 0 ? "+" : ""}${s.schedule_deviation_days}h`
                        : "—"}
                    </td>
                    <td className="py-2.5 pr-3">
                      <span className={cn("inline-flex rounded-full px-2 py-0.5 text-xs font-medium", SNAPSHOT_STATUS_COLOR[s.status])}>
                        {SNAPSHOT_STATUS_LABEL[s.status]}
                      </span>
                    </td>
                    <td className="py-2.5">
                      <SnapshotActions
                        snapshot={s}
                        onEdit={onEditSnapshot}
                        onTransition={onTransitionSnapshot}
                        onReject={onRejectSnapshot}
                        onDelete={onDeleteSnapshot}
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  );
}
