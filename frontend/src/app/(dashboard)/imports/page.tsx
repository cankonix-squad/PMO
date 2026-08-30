"use client";

import { useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowUpFromLine,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  FileDown,
  Loader2,
  RefreshCw,
  XCircle,
} from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { importService } from "@/services/import.service";
import { cn } from "@/lib/utils";
import type { ImportDatasetType, ImportJob, ImportStatus } from "@/types/import";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const DATASET_LABELS: Record<ImportDatasetType, string> = {
  project_progress: "Progress Proyek",
  project_budgets: "Anggaran Proyek",
  risks: "Risiko",
  issues: "Isu",
  benefit_measurements: "Pengukuran Benefit",
};

const STATUS_BADGE: Record<ImportStatus, { label: string; cls: string }> = {
  UPLOADED:  { label: "Diunggah",    cls: "bg-blue-100 text-blue-700" },
  VALIDATED: { label: "Tervalidasi", cls: "bg-yellow-100 text-yellow-700" },
  COMMITTED: { label: "Dikomit",     cls: "bg-green-100 text-green-700" },
  FAILED:    { label: "Gagal",       cls: "bg-red-100 text-red-700" },
  CANCELLED: { label: "Dibatalkan",  cls: "bg-gray-100 text-gray-500" },
};

function StatusBadge({ status }: { status: ImportStatus }) {
  const { label, cls } = STATUS_BADGE[status] ?? { label: status, cls: "bg-gray-100 text-gray-600" };
  return (
    <span className={cn("inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium", cls)}>
      {label}
    </span>
  );
}

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

// ---------------------------------------------------------------------------
// Row preview panel
// ---------------------------------------------------------------------------

function RowPreview({ jobID }: { jobID: string }) {
  const [showInvalid, setShowInvalid] = useState(false);
  const rowsQuery = useQuery({
    queryKey: ["import-rows", jobID, showInvalid],
    queryFn: () =>
      importService.listRows(jobID, {
        valid: showInvalid ? false : undefined,
        page_size: 50,
      }),
  });

  const rows = rowsQuery.data?.data ?? [];

  return (
    <div className="mt-4 rounded-lg border border-gray-200 bg-gray-50 p-4">
      <div className="mb-3 flex items-center justify-between">
        <p className="text-sm font-semibold text-gray-700">Preview Baris</p>
        <label className="flex cursor-pointer items-center gap-2 text-xs text-gray-600">
          <input
            type="checkbox"
            className="rounded border-gray-300"
            checked={showInvalid}
            onChange={(e) => setShowInvalid(e.target.checked)}
          />
          Tampilkan hanya baris invalid
        </label>
      </div>

      {rowsQuery.isLoading && (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-gray-400" />
        </div>
      )}

      {rows.length === 0 && !rowsQuery.isLoading && (
        <p className="py-4 text-center text-sm text-gray-500">Tidak ada data baris.</p>
      )}

      {rows.length > 0 && (
        <div className="overflow-x-auto">
          <table className="min-w-full text-xs">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="py-2 pr-3 text-left font-medium text-gray-500">#</th>
                <th className="py-2 pr-3 text-left font-medium text-gray-500">Aksi</th>
                <th className="py-2 pr-3 text-left font-medium text-gray-500">Valid</th>
                <th className="py-2 pr-3 text-left font-medium text-gray-500">Error</th>
                <th className="py-2 text-left font-medium text-gray-500">Data Mentah</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {rows.map((row) => (
                <tr key={row.id} className={row.valid ? "" : "bg-red-50"}>
                  <td className="py-1.5 pr-3 text-gray-600">{row.row_number}</td>
                  <td className="py-1.5 pr-3">
                    <span
                      className={cn(
                        "rounded px-1.5 py-0.5 text-[10px] font-medium",
                        row.action === "CREATE"
                          ? "bg-green-100 text-green-700"
                          : row.action === "UPDATE"
                          ? "bg-blue-100 text-blue-700"
                          : "bg-gray-100 text-gray-500"
                      )}
                    >
                      {row.action}
                    </span>
                  </td>
                  <td className="py-1.5 pr-3">
                    {row.valid ? (
                      <CheckCircle2 className="h-4 w-4 text-green-500" />
                    ) : (
                      <XCircle className="h-4 w-4 text-red-500" />
                    )}
                  </td>
                  <td className="py-1.5 pr-3 text-red-600">
                    {(row.errors ?? []).join("; ")}
                  </td>
                  <td className="max-w-xs truncate py-1.5 text-gray-600">
                    {Object.entries(row.raw_payload ?? {})
                      .slice(0, 4)
                      .map(([k, v]) => `${k}=${v}`)
                      .join(", ")}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {(rowsQuery.data?.total ?? 0) > 50 && (
            <p className="mt-2 text-xs text-gray-500">
              Menampilkan 50 dari {rowsQuery.data?.total} baris.
            </p>
          )}
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Job card
// ---------------------------------------------------------------------------

function JobCard({ job }: { job: ImportJob }) {
  const [expanded, setExpanded] = useState(false);
  const queryClient = useQueryClient();

  const validateMut = useMutation({
    mutationFn: () => importService.validateJob(job.id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["import-jobs"] }),
  });
  const commitMut = useMutation({
    mutationFn: () => importService.commitJob(job.id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["import-jobs"] }),
  });
  const cancelMut = useMutation({
    mutationFn: () => importService.cancelJob(job.id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["import-jobs"] }),
  });

  const busy = validateMut.isPending || commitMut.isPending || cancelMut.isPending;
  const errors: string[] = (() => {
    try {
      return JSON.parse(job.error_summary ?? "[]") as string[];
    } catch {
      return [];
    }
  })();

  return (
    <div className="rounded-lg border border-gray-200 bg-white shadow-sm">
      {/* Header row */}
      <div className="flex flex-wrap items-center gap-3 px-4 py-3">
        <button
          type="button"
          onClick={() => setExpanded((v) => !v)}
          className="text-gray-500 hover:text-gray-700"
          aria-label={expanded ? "Tutup detail" : "Buka detail"}
        >
          {expanded ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </button>

        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium text-gray-900">{job.file_name}</p>
          <p className="text-xs text-gray-500">
            {DATASET_LABELS[job.dataset_type] ?? job.dataset_type} &middot;{" "}
            {formatBytes(job.file_size)} &middot;{" "}
            {new Date(job.created_at).toLocaleString("id-ID")}
          </p>
        </div>

        <StatusBadge status={job.status} />

        <div className="flex items-center gap-1.5 text-xs text-gray-500">
          <span className="text-green-600 font-medium">{job.valid_rows} valid</span>
          <span>/</span>
          <span className="text-red-500 font-medium">{job.invalid_rows} invalid</span>
          <span>/</span>
          <span>{job.total_rows} total</span>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2">
          {job.status === "UPLOADED" && (
            <button
              type="button"
              disabled={busy}
              onClick={() => validateMut.mutate()}
              className="flex items-center gap-1.5 rounded-md bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {validateMut.isPending && <Loader2 className="h-3 w-3 animate-spin" />}
              Validasi
            </button>
          )}
          {job.status === "VALIDATED" && (
            <button
              type="button"
              disabled={busy || job.valid_rows === 0}
              onClick={() => commitMut.mutate()}
              className="flex items-center gap-1.5 rounded-md bg-green-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-green-700 disabled:opacity-50"
            >
              {commitMut.isPending && <Loader2 className="h-3 w-3 animate-spin" />}
              Commit
            </button>
          )}
          {(job.status === "UPLOADED" || job.status === "VALIDATED") && (
            <button
              type="button"
              disabled={busy}
              onClick={() => cancelMut.mutate()}
              className="flex items-center gap-1.5 rounded-md border border-gray-300 px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
            >
              {cancelMut.isPending && <Loader2 className="h-3 w-3 animate-spin" />}
              Batalkan
            </button>
          )}
        </div>
      </div>

      {/* Error summary */}
      {errors.length > 0 && (
        <div className="border-t border-red-100 bg-red-50 px-4 py-2">
          <p className="mb-1 text-xs font-medium text-red-700">Error Summary:</p>
          <ul className="list-inside list-disc space-y-0.5">
            {errors.slice(0, 5).map((e, i) => (
              <li key={i} className="text-xs text-red-600">
                {e}
              </li>
            ))}
            {errors.length > 5 && (
              <li className="text-xs text-red-500">+{errors.length - 5} error lainnya...</li>
            )}
          </ul>
        </div>
      )}

      {/* Expanded: row preview */}
      {expanded && (
        <div className="border-t border-gray-100 px-4 pb-4">
          <RowPreview jobID={job.id} />
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main page
// ---------------------------------------------------------------------------

export default function ImportsPage() {
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [selectedDataset, setSelectedDataset] = useState<ImportDatasetType>("project_progress");
  const [filterStatus, setFilterStatus] = useState<ImportStatus | "">("");
  const [filterDataset, setFilterDataset] = useState<ImportDatasetType | "">("");

  const templatesQuery = useQuery({
    queryKey: ["import-templates"],
    queryFn: importService.listTemplates,
  });

  const jobsQuery = useQuery({
    queryKey: ["import-jobs", filterStatus, filterDataset],
    queryFn: () =>
      importService.listJobs({
        status: filterStatus || undefined,
        dataset_type: filterDataset || undefined,
        page_size: 20,
      }),
  });

  const uploadMut = useMutation({
    mutationFn: ({ dataset, file }: { dataset: ImportDatasetType; file: File }) =>
      importService.createJob(dataset, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["import-jobs"] });
      if (fileInputRef.current) fileInputRef.current.value = "";
    },
  });

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    uploadMut.mutate({ dataset: selectedDataset, file });
  }

  const jobs = jobsQuery.data?.data ?? [];
  const templates = templatesQuery.data ?? [];
  const selectedTemplate = templates.find((t) => t.dataset_type === selectedDataset);

  return (
    <DashboardLayout title="Import Data">
      <div className="space-y-6">
        {/* ── Upload card ── */}
        <div className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
          <h2 className="mb-4 text-base font-semibold text-gray-900">Unggah File CSV</h2>

          <div className="flex flex-wrap gap-4">
            {/* Dataset selector */}
            <div className="min-w-[200px] flex-1">
              <label className="mb-1.5 block text-xs font-medium text-gray-700">
                Tipe Dataset
              </label>
              <select
                value={selectedDataset}
                onChange={(e) => setSelectedDataset(e.target.value as ImportDatasetType)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                {Object.entries(DATASET_LABELS).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </div>

            {/* File upload */}
            <div className="min-w-[240px] flex-1">
              <label className="mb-1.5 block text-xs font-medium text-gray-700">
                File CSV / Excel
              </label>
              <div className="flex items-center gap-3">
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".csv,.xlsx,.xls"
                  onChange={handleFileChange}
                  disabled={uploadMut.isPending}
                  className="block w-full text-sm text-gray-600 file:mr-3 file:cursor-pointer file:rounded-lg file:border-0 file:bg-blue-50 file:px-3 file:py-1.5 file:text-xs file:font-medium file:text-blue-700 hover:file:bg-blue-100 disabled:opacity-50"
                />
                {uploadMut.isPending && (
                  <Loader2 className="h-4 w-4 animate-spin text-blue-500" />
                )}
              </div>
              {uploadMut.isError && (
                <p className="mt-1 text-xs text-red-600">
                  Upload gagal. Periksa format dan ukuran file (maks. 10 MB).
                </p>
              )}
              {uploadMut.isSuccess && (
                <p className="mt-1 text-xs text-green-600">
                  File berhasil diunggah. Validasi otomatis berjalan…
                </p>
              )}
            </div>
          </div>

          {/* Template hint */}
          {selectedTemplate && (
            <div className="mt-4 rounded-lg bg-blue-50 p-4">
              <div className="mb-2 flex items-center justify-between">
                <p className="text-xs font-semibold text-blue-800">
                  Kolom wajib untuk <em>{selectedTemplate.label}</em>
                </p>
                <span className="text-[10px] text-blue-500">
                  {selectedTemplate.description}
                </span>
              </div>
              <div className="flex flex-wrap gap-2">
                {selectedTemplate.columns.map((col) => (
                  <span
                    key={col.name}
                    title={col.description}
                    className={cn(
                      "rounded-full px-2.5 py-0.5 text-[11px] font-medium",
                      col.required
                        ? "bg-blue-200 text-blue-800"
                        : "bg-blue-100 text-blue-600"
                    )}
                  >
                    {col.name}
                    {col.required && " *"}
                  </span>
                ))}
              </div>
              <p className="mt-2 text-[10px] text-blue-500">
                * Kolom wajib &nbsp;|&nbsp; Row pertama harus menjadi header.
              </p>
            </div>
          )}
        </div>

        {/* ── Job history ── */}
        <div className="rounded-xl border border-gray-200 bg-white shadow-sm">
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-gray-100 px-6 py-4">
            <h2 className="text-base font-semibold text-gray-900">Riwayat Import</h2>

            <div className="flex flex-wrap items-center gap-3">
              {/* Filter dataset */}
              <select
                value={filterDataset}
                onChange={(e) => setFilterDataset(e.target.value as ImportDatasetType | "")}
                className="rounded-lg border border-gray-300 px-3 py-1.5 text-xs focus:border-blue-500 focus:outline-none"
              >
                <option value="">Semua Dataset</option>
                {Object.entries(DATASET_LABELS).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>

              {/* Filter status */}
              <select
                value={filterStatus}
                onChange={(e) => setFilterStatus(e.target.value as ImportStatus | "")}
                className="rounded-lg border border-gray-300 px-3 py-1.5 text-xs focus:border-blue-500 focus:outline-none"
              >
                <option value="">Semua Status</option>
                {Object.entries(STATUS_BADGE).map(([value, { label }]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>

              <button
                type="button"
                onClick={() => queryClient.invalidateQueries({ queryKey: ["import-jobs"] })}
                className="flex items-center gap-1.5 rounded-lg border border-gray-300 px-3 py-1.5 text-xs font-medium text-gray-600 hover:bg-gray-50"
                aria-label="Muat ulang"
              >
                <RefreshCw className="h-3.5 w-3.5" />
                Muat Ulang
              </button>
            </div>
          </div>

          <div className="p-6">
            {jobsQuery.isLoading && (
              <div className="flex justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
              </div>
            )}

            {!jobsQuery.isLoading && jobs.length === 0 && (
              <div className="flex flex-col items-center justify-center py-12 text-gray-400">
                <FileDown className="mb-3 h-10 w-10" />
                <p className="text-sm">Belum ada riwayat import.</p>
                <p className="mt-1 text-xs">Unggah file CSV di atas untuk memulai.</p>
              </div>
            )}

            {jobs.length > 0 && (
              <div className="space-y-3">
                {jobs.map((job) => (
                  <JobCard key={job.id} job={job} />
                ))}
                {(jobsQuery.data?.total ?? 0) > jobs.length && (
                  <p className="text-center text-xs text-gray-500">
                    Menampilkan {jobs.length} dari {jobsQuery.data?.total} job.
                  </p>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Visually hidden upload trigger for screen readers */}
      <div className="sr-only" aria-live="polite">
        {uploadMut.isPending && "Sedang mengunggah file…"}
        {uploadMut.isSuccess && "File berhasil diunggah dan divalidasi."}
      </div>
    </DashboardLayout>
  );
}
