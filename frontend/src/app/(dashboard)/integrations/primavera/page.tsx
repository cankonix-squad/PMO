"use client";

import { useState, useRef } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Upload,
  RefreshCw,
  XCircle,
  ChevronDown,
  ChevronRight,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Loader2,
  FileCode2,
} from "lucide-react";
import { primaveraService } from "@/services/primavera.service";
import type {
  SyncRun,
  SyncStatus,
  SyncFormat,
  ConflictEntry,
  SyncErrorEntry,
  ActivityMapping,
} from "@/types/primavera";
import { DashboardLayout } from "@/components/layout/DashboardLayout";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function fmtBytes(b: number): string {
  if (b < 1024) return `${b} B`;
  if (b < 1024 * 1024) return `${(b / 1024).toFixed(1)} KB`;
  return `${(b / (1024 * 1024)).toFixed(1)} MB`;
}

function fmtDate(s?: string): string {
  if (!s) return "—";
  return new Date(s).toLocaleString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function StatusBadge({ status }: { status: SyncStatus }) {
  const map: Record<SyncStatus, { label: string; cls: string }> = {
    PENDING:   { label: "Pending",   cls: "bg-slate-100 text-slate-600" },
    RUNNING:   { label: "Running",   cls: "bg-blue-100 text-blue-700" },
    DONE:      { label: "Selesai",   cls: "bg-green-100 text-green-700" },
    FAILED:    { label: "Gagal",     cls: "bg-red-100 text-red-700" },
    CANCELLED: { label: "Dibatalkan", cls: "bg-yellow-100 text-yellow-700" },
  };
  const { label, cls } = map[status] ?? { label: status, cls: "bg-slate-100 text-slate-600" };
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${cls}`}>
      {label}
    </span>
  );
}

function StatusIcon({ status }: { status: SyncStatus }) {
  if (status === "DONE") return <CheckCircle2 className="h-4 w-4 text-green-500" />;
  if (status === "FAILED") return <AlertTriangle className="h-4 w-4 text-red-500" />;
  if (status === "RUNNING") return <Loader2 className="h-4 w-4 animate-spin text-blue-500" />;
  if (status === "CANCELLED") return <XCircle className="h-4 w-4 text-yellow-500" />;
  return <Clock className="h-4 w-4 text-slate-400" />;
}

// ---------------------------------------------------------------------------
// RunCard
// ---------------------------------------------------------------------------

function RunCard({ run }: { run: SyncRun }) {
  const [expanded, setExpanded] = useState(false);
  const [mappingsPage, setMappingsPage] = useState(1);

  const { data: mappingsData, isLoading: mappingsLoading } = useQuery({
    queryKey: ["p6-mappings", run.id, mappingsPage],
    queryFn: () => primaveraService.listMappings(run.id, { page: mappingsPage, page_size: 20 }),
    enabled: expanded,
  });

  const errors: SyncErrorEntry[] = (() => {
    try { return JSON.parse(run.error_summary) as SyncErrorEntry[]; } catch { return []; }
  })();
  const conflicts: ConflictEntry[] = (() => {
    try { return JSON.parse(run.conflict_report) as ConflictEntry[]; } catch { return []; }
  })();

  return (
    <div className="rounded-lg border border-slate-200 bg-white shadow-sm">
      {/* Header row */}
      <button
        type="button"
        className="flex w-full items-center gap-3 px-4 py-3 text-left"
        onClick={() => setExpanded((v) => !v)}
      >
        <StatusIcon status={run.status} />
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium text-slate-800">{run.source_file_name}</p>
          <p className="text-xs text-slate-500">
            {run.format} · {fmtBytes(run.source_file_size)} · {fmtDate(run.created_at)}
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-3">
          <StatusBadge status={run.status} />
          {run.status === "DONE" && (
            <span className="text-xs text-slate-500">
              {run.imported_activities} impor / {run.skipped_activities} lewat / {run.failed_activities} gagal
            </span>
          )}
          {expanded ? (
            <ChevronDown className="h-4 w-4 text-slate-400" />
          ) : (
            <ChevronRight className="h-4 w-4 text-slate-400" />
          )}
        </div>
      </button>

      {/* Expanded detail */}
      {expanded && (
        <div className="border-t border-slate-100 px-4 pb-4 pt-3 space-y-4">
          {/* Stats row */}
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            {[
              { label: "Total Aktivitas", value: run.total_activities },
              { label: "Diimpor", value: run.imported_activities },
              { label: "Dilewati", value: run.skipped_activities },
              { label: "Konflik", value: run.conflict_count },
            ].map(({ label, value }) => (
              <div key={label} className="rounded-md bg-slate-50 px-3 py-2">
                <p className="text-xs text-slate-500">{label}</p>
                <p className="text-lg font-semibold text-slate-800">{value}</p>
              </div>
            ))}
          </div>

          {/* Errors */}
          {errors.length > 0 && (
            <div>
              <p className="mb-1 text-xs font-semibold text-red-600">Error ({errors.length})</p>
              <ul className="space-y-1">
                {errors.slice(0, 10).map((e, i) => (
                  <li key={i} className="rounded bg-red-50 px-3 py-1.5 text-xs text-red-700">
                    [{e.code}] {e.message}
                    {e.activity_id && <span className="ml-1 font-mono">({e.activity_id})</span>}
                  </li>
                ))}
                {errors.length > 10 && (
                  <li className="text-xs text-slate-500">… dan {errors.length - 10} error lainnya</li>
                )}
              </ul>
            </div>
          )}

          {/* Conflicts */}
          {conflicts.length > 0 && (
            <div>
              <p className="mb-1 text-xs font-semibold text-yellow-700">Konflik ({conflicts.length})</p>
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="bg-yellow-50 text-left text-yellow-700">
                      <th className="px-2 py-1">Activity ID</th>
                      <th className="px-2 py-1">Field</th>
                      <th className="px-2 py-1">Existing</th>
                      <th className="px-2 py-1">Incoming</th>
                    </tr>
                  </thead>
                  <tbody>
                    {conflicts.slice(0, 20).map((c, i) => (
                      <tr key={i} className="border-t border-yellow-100">
                        <td className="px-2 py-1 font-mono">{c.activity_id}</td>
                        <td className="px-2 py-1">{c.field}</td>
                        <td className="px-2 py-1 text-slate-500">{c.existing}</td>
                        <td className="px-2 py-1 font-medium text-slate-800">{c.incoming}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Mappings */}
          {run.status === "DONE" && (
            <div>
              <p className="mb-2 text-xs font-semibold text-slate-600">Activity Mappings</p>
              {mappingsLoading ? (
                <div className="flex items-center gap-2 text-xs text-slate-500">
                  <Loader2 className="h-3 w-3 animate-spin" /> Memuat mappings…
                </div>
              ) : mappingsData && mappingsData.data.length > 0 ? (
                <>
                  <div className="overflow-x-auto rounded border border-slate-100">
                    <table className="w-full text-xs">
                      <thead className="bg-slate-50 text-slate-500">
                        <tr>
                          <th className="px-2 py-1 text-left">P6 Activity ID</th>
                          <th className="px-2 py-1 text-left">Nama</th>
                          <th className="px-2 py-1 text-left">WBS</th>
                          <th className="px-2 py-1 text-right">Baseline %</th>
                          <th className="px-2 py-1 text-right">Actual %</th>
                          <th className="px-2 py-1 text-left">Action</th>
                        </tr>
                      </thead>
                      <tbody>
                        {mappingsData.data.map((m: ActivityMapping) => (
                          <tr key={m.id} className="border-t border-slate-100 hover:bg-slate-50">
                            <td className="px-2 py-1 font-mono">{m.p6_activity_id}</td>
                            <td className="px-2 py-1 max-w-[200px] truncate">{m.p6_activity_name}</td>
                            <td className="px-2 py-1 text-slate-500">{m.p6_wbs_code || "—"}</td>
                            <td className="px-2 py-1 text-right">{m.baseline_physical.toFixed(1)}%</td>
                            <td className="px-2 py-1 text-right">{m.actual_physical.toFixed(1)}%</td>
                            <td className="px-2 py-1">
                              <span className={`rounded px-1 py-0.5 text-[10px] font-medium ${
                                m.action === "CREATE" ? "bg-green-100 text-green-700"
                                : m.action === "UPDATE" ? "bg-blue-100 text-blue-700"
                                : m.action === "CONFLICT" ? "bg-yellow-100 text-yellow-700"
                                : "bg-slate-100 text-slate-500"
                              }`}>
                                {m.action}
                              </span>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  {/* Pagination */}
                  {mappingsData.meta.total > 20 && (
                    <div className="mt-2 flex items-center justify-between text-xs text-slate-500">
                      <span>{mappingsData.meta.total} total</span>
                      <div className="flex gap-1">
                        <button
                          disabled={mappingsPage === 1}
                          onClick={() => setMappingsPage((p) => p - 1)}
                          className="rounded border px-2 py-1 disabled:opacity-40"
                        >← Prev</button>
                        <span className="px-2 py-1">Hal {mappingsPage}</span>
                        <button
                          disabled={mappingsPage * 20 >= mappingsData.meta.total}
                          onClick={() => setMappingsPage((p) => p + 1)}
                          className="rounded border px-2 py-1 disabled:opacity-40"
                        >Next →</button>
                      </div>
                    </div>
                  )}
                </>
              ) : (
                <p className="text-xs text-slate-400">Tidak ada mapping tercatat.</p>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main Page
// ---------------------------------------------------------------------------

export default function PrimaveraPage() {
  const qc = useQueryClient();
  const fileRef = useRef<HTMLInputElement>(null);

  const [projectID, setProjectID] = useState("");
  const [format, setFormat] = useState<SyncFormat | "">("");
  const [p6Version, setP6Version] = useState("");
  const [filterStatus, setFilterStatus] = useState<SyncStatus | "">("");
  const [filterFormat, setFilterFormat] = useState<SyncFormat | "">("");
  const [page, setPage] = useState(1);

  const { data: runsData, isLoading, refetch } = useQuery({
    queryKey: ["p6-runs", filterStatus, filterFormat, page],
    queryFn: () =>
      primaveraService.listRuns({
        status: filterStatus || undefined,
        format: filterFormat || undefined,
        page,
        page_size: 20,
      }),
  });

  const uploadMutation = useMutation({
    mutationFn: async (file: File) => {
      if (!projectID.trim()) throw new Error("Project ID wajib diisi");
      return primaveraService.createRun(projectID.trim(), file, {
        format: format || undefined,
        p6Version: p6Version || undefined,
      });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["p6-runs"] });
      setProjectID("");
      setFormat("");
      setP6Version("");
      if (fileRef.current) fileRef.current.value = "";
    },
  });

  const cancelMutation = useMutation({
    mutationFn: (runID: string) => primaveraService.cancelRun(runID),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["p6-runs"] }),
  });

  const handleUpload = () => {
    const file = fileRef.current?.files?.[0];
    if (!file) return;
    uploadMutation.mutate(file);
  };

  const runs = runsData?.data ?? [];
  const total = runsData?.meta.total ?? 0;

  return (
    <DashboardLayout title="Primavera P6 Import">
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-800">Primavera P6 Import</h1>
          <p className="text-sm text-slate-500 mt-0.5">
            Upload file XER atau PMXML dari Oracle Primavera P6 untuk sinkronisasi aktivitas ke proyek.
          </p>
        </div>
        <button
          onClick={() => refetch()}
          className="flex items-center gap-1.5 rounded-lg border border-slate-200 px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-50"
        >
          <RefreshCw className="h-4 w-4" />
          Muat Ulang
        </button>
      </div>

      {/* Upload card */}
      <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
        <div className="flex items-center gap-2 mb-4">
          <FileCode2 className="h-5 w-5 text-blue-600" />
          <h2 className="font-semibold text-slate-700">Upload File P6</h2>
        </div>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {/* Project ID */}
          <div className="lg:col-span-2">
            <label className="mb-1 block text-xs font-medium text-slate-600">
              Project ID <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              placeholder="UUID project tujuan"
              value={projectID}
              onChange={(e) => setProjectID(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>

          {/* Format */}
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-600">Format</label>
            <select
              value={format}
              onChange={(e) => setFormat(e.target.value as SyncFormat | "")}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            >
              <option value="">Auto-detect</option>
              <option value="XER">XER</option>
              <option value="PMXML">PMXML</option>
            </select>
          </div>

          {/* P6 Version */}
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-600">Versi P6</label>
            <input
              type="text"
              placeholder="Contoh: 22.12"
              value={p6Version}
              onChange={(e) => setP6Version(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>

          {/* File picker */}
          <div className="sm:col-span-2 lg:col-span-3">
            <label className="mb-1 block text-xs font-medium text-slate-600">
              File <span className="text-red-500">*</span>
            </label>
            <input
              ref={fileRef}
              type="file"
              accept=".xer,.xml,.pmxml"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm file:mr-3 file:rounded file:border-0 file:bg-blue-50 file:px-3 file:py-1 file:text-xs file:font-medium file:text-blue-700 hover:file:bg-blue-100"
            />
          </div>

          {/* Upload button */}
          <div className="flex items-end">
            <button
              onClick={handleUpload}
              disabled={uploadMutation.isPending}
              className="flex w-full items-center justify-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60"
            >
              {uploadMutation.isPending ? (
                <><Loader2 className="h-4 w-4 animate-spin" /> Memproses…</>
              ) : (
                <><Upload className="h-4 w-4" /> Upload & Proses</>
              )}
            </button>
          </div>
        </div>

        {/* Error/success feedback */}
        {uploadMutation.isError && (
          <div className="mt-3 rounded-lg bg-red-50 px-4 py-2 text-sm text-red-700">
            {(uploadMutation.error as Error).message}
          </div>
        )}
        {uploadMutation.isSuccess && (
          <div className="mt-3 rounded-lg bg-green-50 px-4 py-2 text-sm text-green-700">
            Sync run selesai — status:{" "}
            <span className="font-semibold">{uploadMutation.data?.status}</span>
            {" · "}
            {uploadMutation.data?.imported_activities} aktivitas diimpor.
          </div>
        )}
      </div>

      {/* Filters + list */}
      <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 className="font-semibold text-slate-700">Riwayat Sync Run</h2>
          <div className="flex gap-2">
            <select
              value={filterStatus}
              onChange={(e) => { setFilterStatus(e.target.value as SyncStatus | ""); setPage(1); }}
              className="rounded-lg border border-slate-300 px-2 py-1.5 text-xs focus:outline-none"
            >
              <option value="">Semua Status</option>
              <option value="PENDING">Pending</option>
              <option value="RUNNING">Running</option>
              <option value="DONE">Selesai</option>
              <option value="FAILED">Gagal</option>
              <option value="CANCELLED">Dibatalkan</option>
            </select>
            <select
              value={filterFormat}
              onChange={(e) => { setFilterFormat(e.target.value as SyncFormat | ""); setPage(1); }}
              className="rounded-lg border border-slate-300 px-2 py-1.5 text-xs focus:outline-none"
            >
              <option value="">Semua Format</option>
              <option value="XER">XER</option>
              <option value="PMXML">PMXML</option>
            </select>
          </div>
        </div>

        {isLoading ? (
          <div className="flex items-center gap-2 py-8 justify-center text-sm text-slate-500">
            <Loader2 className="h-5 w-5 animate-spin" /> Memuat…
          </div>
        ) : runs.length === 0 ? (
          <div className="py-12 text-center text-sm text-slate-400">
            Belum ada sync run. Upload file P6 di atas untuk memulai.
          </div>
        ) : (
          <div className="space-y-3">
            {runs.map((run) => (
              <div key={run.id} className="relative">
                <RunCard run={run} />
                {run.status === "PENDING" && (
                  <button
                    onClick={() => cancelMutation.mutate(run.id)}
                    disabled={cancelMutation.isPending}
                    className="absolute right-12 top-3 flex items-center gap-1 rounded px-2 py-1 text-xs text-red-600 hover:bg-red-50"
                  >
                    <XCircle className="h-3 w-3" /> Batalkan
                  </button>
                )}
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        {total > 20 && (
          <div className="mt-4 flex items-center justify-between text-xs text-slate-500">
            <span>{total} total run</span>
            <div className="flex gap-1">
              <button
                disabled={page === 1}
                onClick={() => setPage((p) => p - 1)}
                className="rounded border px-2 py-1 disabled:opacity-40"
              >← Prev</button>
              <span className="px-2 py-1">Hal {page}</span>
              <button
                disabled={page * 20 >= total}
                onClick={() => setPage((p) => p + 1)}
                className="rounded border px-2 py-1 disabled:opacity-40"
              >Next →</button>
            </div>
          </div>
        )}
      </div>
    </div>
    </DashboardLayout>
  );
}
