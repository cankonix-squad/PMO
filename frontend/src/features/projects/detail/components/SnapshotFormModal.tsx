"use client";

import { cn } from "@/lib/utils";
import type { Baseline, CreateSnapshotRequest, Snapshot, UpdateSnapshotRequest } from "@/types/monitoring";
import { MONTH_NAMES } from "@/types/monitoring";
import { X } from "lucide-react";
import { inputClassName, primaryButton, secondaryButton } from "../schemas";

export function SnapshotFormModal({
  snapshot,
  baselines,
  isSaving,
  formError,
  onSubmit,
  onClose,
}: {
  snapshot: Snapshot | null;
  baselines: Baseline[];
  isSaving: boolean;
  formError: string | null;
  onSubmit: (payload: CreateSnapshotRequest | UpdateSnapshotRequest) => void;
  onClose: () => void;
}) {
  const isEdit = snapshot !== null;
  const activeBaseline = baselines.find((b) => b.is_active);

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    if (isEdit) {
      const payload: UpdateSnapshotRequest = {
        physical_actual: parseFloat(fd.get("physical_actual") as string),
        physical_target: parseFloat(fd.get("physical_target") as string),
        financial_actual: parseFloat(fd.get("financial_actual") as string),
        financial_target: parseFloat(fd.get("financial_target") as string),
        schedule_deviation_days: fd.get("schedule_deviation_days")
          ? parseInt(fd.get("schedule_deviation_days") as string)
          : undefined,
        source: (fd.get("source") as string) || undefined,
        notes: (fd.get("notes") as string) || undefined,
      };
      onSubmit(payload);
    } else {
      const payload: CreateSnapshotRequest = {
        baseline_id: (fd.get("baseline_id") as string) || undefined,
        period_year: parseInt(fd.get("period_year") as string),
        period_month: parseInt(fd.get("period_month") as string),
        physical_actual: parseFloat(fd.get("physical_actual") as string),
        physical_target: parseFloat(fd.get("physical_target") as string),
        financial_actual: parseFloat(fd.get("financial_actual") as string),
        financial_target: parseFloat(fd.get("financial_target") as string),
        currency: (fd.get("currency") as string) || "IDR",
        schedule_deviation_days: fd.get("schedule_deviation_days")
          ? parseInt(fd.get("schedule_deviation_days") as string)
          : undefined,
        source: (fd.get("source") as string) || undefined,
        notes: (fd.get("notes") as string) || undefined,
      };
      onSubmit(payload);
    }
  }

  const currentYear = new Date().getFullYear();

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-lg rounded-xl border border-border bg-card shadow-xl overflow-y-auto max-h-[90vh]">
        <div className="flex items-center justify-between border-b border-border px-6 py-4">
          <h3 className="text-base font-semibold">
            {isEdit ? "Edit Snapshot" : "Tambah Snapshot Periodik"}
          </h3>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground" aria-label="Close">
            <X className="h-4 w-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 px-6 py-5">
          {formError && (
            <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{formError}</p>
          )}

          {!isEdit && (
            <>
              <div className="space-y-1">
                <label className="text-sm font-medium" htmlFor="sn-baseline">Baseline</label>
                <select
                  id="sn-baseline"
                  name="baseline_id"
                  className={inputClassName}
                  defaultValue={activeBaseline?.id ?? ""}
                >
                  <option value="">— Tanpa baseline —</option>
                  {baselines.map((b) => (
                    <option key={b.id} value={b.id}>
                      v{b.version}{b.label ? ` — ${b.label}` : ""}{b.is_active ? " (aktif)" : ""}
                    </option>
                  ))}
                </select>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <label className="text-sm font-medium" htmlFor="sn-year">
                    Tahun <span className="text-destructive">*</span>
                  </label>
                  <input
                    id="sn-year"
                    name="period_year"
                    type="number"
                    min={2000}
                    max={2100}
                    className={inputClassName}
                    defaultValue={currentYear}
                    required
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium" htmlFor="sn-month">
                    Bulan <span className="text-destructive">*</span>
                  </label>
                  <select
                    id="sn-month"
                    name="period_month"
                    className={inputClassName}
                    defaultValue={new Date().getMonth() + 1}
                    required
                  >
                    {MONTH_NAMES.slice(1).map((name, i) => (
                      <option key={i + 1} value={i + 1}>{name}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="space-y-1">
                <label className="text-sm font-medium" htmlFor="sn-currency">Mata Uang</label>
                <input
                  id="sn-currency"
                  name="currency"
                  className={inputClassName}
                  defaultValue="IDR"
                  placeholder="IDR"
                />
              </div>
            </>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="sn-physical-target">
                Target Fisik (%) <span className="text-destructive">*</span>
              </label>
              <input
                id="sn-physical-target"
                name="physical_target"
                type="number"
                min={0}
                max={100}
                step={0.01}
                className={inputClassName}
                defaultValue={snapshot?.physical_target ?? ""}
                required
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="sn-physical-actual">
                Aktual Fisik (%) <span className="text-destructive">*</span>
              </label>
              <input
                id="sn-physical-actual"
                name="physical_actual"
                type="number"
                min={0}
                max={100}
                step={0.01}
                className={inputClassName}
                defaultValue={snapshot?.physical_actual ?? ""}
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="sn-fin-target">
                Target Keuangan <span className="text-destructive">*</span>
              </label>
              <input
                id="sn-fin-target"
                name="financial_target"
                type="number"
                min={0}
                step={1}
                className={inputClassName}
                defaultValue={snapshot?.financial_target ?? ""}
                required
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="sn-fin-actual">
                Aktual Keuangan <span className="text-destructive">*</span>
              </label>
              <input
                id="sn-fin-actual"
                name="financial_actual"
                type="number"
                min={0}
                step={1}
                className={inputClassName}
                defaultValue={snapshot?.financial_actual ?? ""}
                required
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="sn-sched-dev">
              Deviasi Jadwal (hari)
            </label>
            <input
              id="sn-sched-dev"
              name="schedule_deviation_days"
              type="number"
              className={inputClassName}
              defaultValue={snapshot?.schedule_deviation_days ?? ""}
              placeholder="0 = sesuai jadwal, positif = terlambat"
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="sn-source">Sumber Data</label>
            <input
              id="sn-source"
              name="source"
              className={inputClassName}
              defaultValue={snapshot?.source ?? ""}
              placeholder="misal: Laporan Bulanan"
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="sn-notes">Catatan</label>
            <textarea
              id="sn-notes"
              name="notes"
              rows={2}
              className={cn(inputClassName, "h-auto resize-none")}
              defaultValue={snapshot?.notes ?? ""}
            />
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className={secondaryButton}>Batal</button>
            <button type="submit" disabled={isSaving} className={primaryButton}>
              {isSaving ? "Menyimpan…" : isEdit ? "Perbarui" : "Simpan"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
