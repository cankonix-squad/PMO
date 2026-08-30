"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  BarChart2,
  ChevronDown,
  ChevronUp,
  Edit3,
  Plus,
  Trash2,
  X,
} from "lucide-react";
import { periodicReportService } from "@/services/periodic-report.service";
import {
  computeVariance,
  formatCurrency,
  formatPeriod,
  MONTH_FULL,
  MONTH_LABELS,
} from "@/types/periodic-report";
import type {
  CreatePeriodicReportRequest,
  PeriodicReport,
  UpdatePeriodicReportRequest,
} from "@/types/periodic-report";
import { cn } from "@/lib/utils";

// ── helpers ──────────────────────────────────────────────────────────────────

const currentYear = new Date().getFullYear();
const YEARS = Array.from({ length: 10 }, (_, i) => currentYear - i);
const MONTHS = Array.from({ length: 12 }, (_, i) => i + 1);

function pctColor(pct: number): string {
  if (pct >= 80) return "text-green-600";
  if (pct >= 50) return "text-yellow-600";
  return "text-red-500";
}

function varianceColor(v: number): string {
  return v >= 0 ? "text-green-600" : "text-red-500";
}

// ── sub-components ────────────────────────────────────────────────────────────

interface FormState {
  period_year: number;
  period_month: number;
  physical_progress_pct: string;
  financial_planned: string;
  financial_actual: string;
  notes: string;
}

const emptyForm = (): FormState => ({
  period_year: new Date().getFullYear(),
  period_month: new Date().getMonth() + 1,
  physical_progress_pct: "",
  financial_planned: "",
  financial_actual: "",
  notes: "",
});

function formFromReport(r: PeriodicReport): FormState {
  return {
    period_year: r.period_year,
    period_month: r.period_month,
    physical_progress_pct: String(r.physical_progress_pct),
    financial_planned: String(r.financial_planned),
    financial_actual: String(r.financial_actual),
    notes: r.notes ?? "",
  };
}

function validateForm(f: FormState): string | null {
  const year = Number(f.period_year);
  const month = Number(f.period_month);
  const phys = Number(f.physical_progress_pct);
  const planned = Number(f.financial_planned);
  const actual = Number(f.financial_actual);

  if (isNaN(year) || year < 2000 || year > 2100) return "Tahun tidak valid";
  if (isNaN(month) || month < 1 || month > 12) return "Bulan tidak valid";
  if (isNaN(phys) || phys < 0 || phys > 100)
    return "Progres fisik harus antara 0–100";
  if (isNaN(planned) || planned < 0)
    return "Rencana keuangan tidak boleh negatif";
  if (isNaN(actual) || actual < 0)
    return "Realisasi keuangan tidak boleh negatif";
  return null;
}

// ── main component ────────────────────────────────────────────────────────────

interface PeriodicReportPanelProps {
  projectId: string;
}

export function PeriodicReportPanel({ projectId }: PeriodicReportPanelProps) {
  const qc = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const [formState, setFormState] = useState<FormState>(emptyForm());
  const [formError, setFormError] = useState<string | null>(null);
  const [sortDesc, setSortDesc] = useState(true);

  const { data, isLoading, isError } = useQuery({
    queryKey: ["periodic-reports", projectId],
    queryFn: () =>
      periodicReportService.list(projectId, { page: 1, page_size: 50 }),
  });

  const reports: PeriodicReport[] = data?.data ?? [];

  const createMutation = useMutation({
    mutationFn: (req: CreatePeriodicReportRequest) =>
      periodicReportService.create(projectId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["periodic-reports", projectId] });
      qc.invalidateQueries({ queryKey: ["project", projectId] });
      setShowForm(false);
      setFormState(emptyForm());
      setFormError(null);
    },
    onError: (err: { response?: { data?: { message?: string } } }) => {
      setFormError(
        err?.response?.data?.message ?? "Gagal menyimpan laporan periodik"
      );
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({
      id,
      req,
    }: {
      id: string;
      req: UpdatePeriodicReportRequest;
    }) => periodicReportService.update(projectId, id, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["periodic-reports", projectId] });
      qc.invalidateQueries({ queryKey: ["project", projectId] });
      setEditingId(null);
      setFormState(emptyForm());
      setFormError(null);
    },
    onError: (err: { response?: { data?: { message?: string } } }) => {
      setFormError(
        err?.response?.data?.message ?? "Gagal memperbarui laporan periodik"
      );
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      periodicReportService.delete(projectId, id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["periodic-reports", projectId] });
      qc.invalidateQueries({ queryKey: ["project", projectId] });
      setConfirmDeleteId(null);
    },
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const err = validateForm(formState);
    if (err) {
      setFormError(err);
      return;
    }
    setFormError(null);
    if (editingId) {
      updateMutation.mutate({
        id: editingId,
        req: {
          physical_progress_pct: Number(formState.physical_progress_pct),
          financial_planned: Number(formState.financial_planned),
          financial_actual: Number(formState.financial_actual),
          notes: formState.notes || undefined,
        },
      });
    } else {
      createMutation.mutate({
        period_year: Number(formState.period_year),
        period_month: Number(formState.period_month),
        physical_progress_pct: Number(formState.physical_progress_pct),
        financial_planned: Number(formState.financial_planned),
        financial_actual: Number(formState.financial_actual),
        notes: formState.notes || undefined,
      });
    }
  }

  function openEdit(r: PeriodicReport) {
    setEditingId(r.id);
    setFormState(formFromReport(r));
    setFormError(null);
    setShowForm(true);
  }

  function cancelForm() {
    setShowForm(false);
    setEditingId(null);
    setFormState(emptyForm());
    setFormError(null);
  }

  const sortedReports = [...reports].sort((a, b) => {
    const aVal = a.period_year * 100 + a.period_month;
    const bVal = b.period_year * 100 + b.period_month;
    return sortDesc ? bVal - aVal : aVal - bVal;
  });

  const isSubmitting = createMutation.isPending || updateMutation.isPending;

  return (
    <section className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <BarChart2 className="h-4 w-4 text-blue-600" />
          <h3 className="font-semibold text-slate-800">Laporan Periodik</h3>
          <span className="text-xs text-slate-400 ml-1">
            (laporan periodik operasional)
          </span>
        </div>
        <button
          onClick={() => {
            setEditingId(null);
            setFormState(emptyForm());
            setFormError(null);
            setShowForm((v) => !v);
          }}
          className="flex items-center gap-1 rounded-md bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700 transition-colors"
        >
          <Plus className="h-3.5 w-3.5" />
          Tambah Laporan
        </button>
      </div>

      {/* Inline Form */}
      {showForm && (
        <div className="rounded-lg border border-blue-200 bg-blue-50 p-4 space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-slate-700">
              {editingId ? "Edit Laporan Periodik" : "Tambah Laporan Periodik"}
            </span>
            <button
              onClick={cancelForm}
              className="text-slate-400 hover:text-slate-600"
            >
              <X className="h-4 w-4" />
            </button>
          </div>

          <form onSubmit={handleSubmit} className="space-y-3">
            {/* Period row */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs font-medium text-slate-600 mb-1 block">
                  Tahun *
                </label>
                <select
                  value={formState.period_year}
                  onChange={(e) =>
                    setFormState((s) => ({
                      ...s,
                      period_year: Number(e.target.value),
                    }))
                  }
                  disabled={!!editingId}
                  className="w-full rounded-md border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-60"
                >
                  {YEARS.map((y) => (
                    <option key={y} value={y}>
                      {y}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-xs font-medium text-slate-600 mb-1 block">
                  Bulan *
                </label>
                <select
                  value={formState.period_month}
                  onChange={(e) =>
                    setFormState((s) => ({
                      ...s,
                      period_month: Number(e.target.value),
                    }))
                  }
                  disabled={!!editingId}
                  className="w-full rounded-md border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-60"
                >
                  {MONTHS.map((m) => (
                    <option key={m} value={m}>
                      {MONTH_FULL[m]}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {/* Progress */}
            <div>
              <label className="text-xs font-medium text-slate-600 mb-1 block">
                Progres Fisik (%) *
              </label>
              <input
                type="number"
                min={0}
                max={100}
                step={0.01}
                value={formState.physical_progress_pct}
                onChange={(e) =>
                  setFormState((s) => ({
                    ...s,
                    physical_progress_pct: e.target.value,
                  }))
                }
                placeholder="0 – 100"
                className="w-full rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            {/* Financial */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs font-medium text-slate-600 mb-1 block">
                  Rencana Keuangan (IDR) *
                </label>
                <input
                  type="number"
                  min={0}
                  step={1}
                  value={formState.financial_planned}
                  onChange={(e) =>
                    setFormState((s) => ({
                      ...s,
                      financial_planned: e.target.value,
                    }))
                  }
                  placeholder="0"
                  className="w-full rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-slate-600 mb-1 block">
                  Realisasi Keuangan (IDR) *
                </label>
                <input
                  type="number"
                  min={0}
                  step={1}
                  value={formState.financial_actual}
                  onChange={(e) =>
                    setFormState((s) => ({
                      ...s,
                      financial_actual: e.target.value,
                    }))
                  }
                  placeholder="0"
                  className="w-full rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>

            {/* Computed preview */}
            {formState.financial_planned && formState.financial_actual && (
              <div className="rounded bg-slate-100 px-3 py-2 text-xs text-slate-600 flex gap-4 flex-wrap">
                <span>
                  Realisasi keuangan:{" "}
                  <strong>
                    {Number(formState.financial_planned) > 0
                      ? (
                          (Number(formState.financial_actual) /
                            Number(formState.financial_planned)) *
                          100
                        ).toFixed(2)
                      : "0.00"}
                    %
                  </strong>
                </span>
                <span>
                  Variance:{" "}
                  <strong
                    className={
                      Number(formState.financial_planned) -
                        Number(formState.financial_actual) >=
                      0
                        ? "text-green-600"
                        : "text-red-500"
                    }
                  >
                    {formatCurrency(
                      Number(formState.financial_planned) -
                        Number(formState.financial_actual)
                    )}
                  </strong>
                </span>
              </div>
            )}

            {/* Notes */}
            <div>
              <label className="text-xs font-medium text-slate-600 mb-1 block">
                Catatan
              </label>
              <textarea
                rows={2}
                value={formState.notes}
                onChange={(e) =>
                  setFormState((s) => ({ ...s, notes: e.target.value }))
                }
                placeholder="Catatan opsional..."
                className="w-full rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
              />
            </div>

            {formError && (
              <p className="text-xs text-red-600 bg-red-50 rounded px-2 py-1">
                {formError}
              </p>
            )}

            <div className="flex justify-end gap-2">
              <button
                type="button"
                onClick={cancelForm}
                className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 hover:bg-slate-50 transition-colors"
              >
                Batal
              </button>
              <button
                type="submit"
                disabled={isSubmitting}
                className="rounded-md bg-blue-600 px-4 py-1.5 text-xs font-medium text-white hover:bg-blue-700 disabled:opacity-60 transition-colors"
              >
                {isSubmitting ? "Menyimpan..." : editingId ? "Simpan" : "Tambah"}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Table */}
      {isLoading && (
        <div className="py-8 text-center text-sm text-slate-400">
          Memuat laporan periodik...
        </div>
      )}
      {isError && (
        <div className="py-4 text-center text-sm text-red-500">
          Gagal memuat laporan periodik.
        </div>
      )}

      {!isLoading && !isError && reports.length === 0 && (
        <div className="py-8 text-center text-sm text-slate-400">
          Belum ada laporan periodik. Tambahkan laporan bulanan untuk mengisi
          grafik tren dashboard.
        </div>
      )}

      {!isLoading && !isError && reports.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full text-xs border-collapse">
            <thead>
              <tr className="bg-slate-50 border-b border-slate-200">
                <th
                  className="px-3 py-2 text-left font-medium text-slate-500 cursor-pointer whitespace-nowrap"
                  onClick={() => setSortDesc((v) => !v)}
                >
                  <span className="flex items-center gap-1">
                    Periode
                    {sortDesc ? (
                      <ChevronDown className="h-3 w-3" />
                    ) : (
                      <ChevronUp className="h-3 w-3" />
                    )}
                  </span>
                </th>
                <th className="px-3 py-2 text-right font-medium text-slate-500">
                  Fisik %
                </th>
                <th className="px-3 py-2 text-right font-medium text-slate-500">
                  Keuangan %
                </th>
                <th className="px-3 py-2 text-right font-medium text-slate-500 hidden sm:table-cell">
                  Rencana
                </th>
                <th className="px-3 py-2 text-right font-medium text-slate-500 hidden sm:table-cell">
                  Realisasi
                </th>
                <th className="px-3 py-2 text-right font-medium text-slate-500 hidden md:table-cell">
                  Variance
                </th>
                <th className="px-3 py-2 text-left font-medium text-slate-500 hidden lg:table-cell">
                  Catatan
                </th>
                <th className="px-3 py-2" />
              </tr>
            </thead>
            <tbody>
              {sortedReports.map((r) => {
                const variance = computeVariance(r);
                const isDeleting =
                  deleteMutation.isPending && confirmDeleteId === r.id;
                return (
                  <tr
                    key={r.id}
                    className="border-b border-slate-100 hover:bg-slate-50 transition-colors"
                  >
                    <td className="px-3 py-2 font-medium text-slate-700 whitespace-nowrap">
                      {MONTH_LABELS[r.period_month]} {r.period_year}
                    </td>
                    <td
                      className={cn(
                        "px-3 py-2 text-right font-semibold",
                        pctColor(r.physical_progress_pct)
                      )}
                    >
                      {r.physical_progress_pct.toFixed(1)}%
                    </td>
                    <td
                      className={cn(
                        "px-3 py-2 text-right font-semibold",
                        pctColor(r.financial_pct)
                      )}
                    >
                      {r.financial_pct.toFixed(1)}%
                    </td>
                    <td className="px-3 py-2 text-right text-slate-500 hidden sm:table-cell">
                      {formatCurrency(r.financial_planned)}
                    </td>
                    <td className="px-3 py-2 text-right text-slate-600 hidden sm:table-cell">
                      {formatCurrency(r.financial_actual)}
                    </td>
                    <td
                      className={cn(
                        "px-3 py-2 text-right font-medium hidden md:table-cell",
                        varianceColor(variance)
                      )}
                    >
                      {formatCurrency(variance)}
                    </td>
                    <td className="px-3 py-2 text-slate-400 max-w-[200px] truncate hidden lg:table-cell">
                      {r.notes ?? "—"}
                    </td>
                    <td className="px-3 py-2">
                      {confirmDeleteId === r.id ? (
                        <div className="flex items-center gap-1">
                          <button
                            onClick={() => deleteMutation.mutate(r.id)}
                            disabled={isDeleting}
                            className="text-red-600 hover:text-red-800 text-xs font-medium"
                          >
                            {isDeleting ? "..." : "Hapus"}
                          </button>
                          <button
                            onClick={() => setConfirmDeleteId(null)}
                            className="text-slate-400 hover:text-slate-600 text-xs"
                          >
                            Batal
                          </button>
                        </div>
                      ) : (
                        <div className="flex items-center gap-1 justify-end">
                          <button
                            onClick={() => openEdit(r)}
                            className="p-1 text-slate-400 hover:text-blue-600 rounded transition-colors"
                            title="Edit"
                          >
                            <Edit3 className="h-3.5 w-3.5" />
                          </button>
                          <button
                            onClick={() => setConfirmDeleteId(r.id)}
                            className="p-1 text-slate-400 hover:text-red-600 rounded transition-colors"
                            title="Hapus"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      <p className="text-xs text-slate-400">
        Laporan periodik operasional — sumber data grafik tren dashboard.
        Belum melewati proses validasi Data Governance.
      </p>
    </section>
  );
}
