"use client";

import { cn } from "@/lib/utils";
import type { Baseline, CreateBaselineRequest, UpdateBaselineRequest } from "@/types/monitoring";
import { X } from "lucide-react";
import { inputClassName, primaryButton, secondaryButton } from "../schemas";

export function BaselineFormModal({
  baseline,
  isSaving,
  formError,
  onSubmit,
  onClose,
}: {
  baseline: Baseline | null;
  isSaving: boolean;
  formError: string | null;
  onSubmit: (payload: CreateBaselineRequest | UpdateBaselineRequest) => void;
  onClose: () => void;
}) {
  const isEdit = baseline !== null;

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const payload = {
      label: (fd.get("label") as string) || undefined,
      physical_target: parseFloat(fd.get("physical_target") as string),
      budget_total: parseFloat(fd.get("budget_total") as string),
      currency: (fd.get("currency") as string) || "IDR",
      planned_start: fd.get("planned_start") as string,
      planned_end: fd.get("planned_end") as string,
      source: (fd.get("source") as string) || undefined,
      notes: (fd.get("notes") as string) || undefined,
      ...(isEdit ? { is_active: (fd.get("is_active") as string) === "true" } : {}),
    };
    onSubmit(payload);
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-lg rounded-xl border border-border bg-card shadow-xl">
        <div className="flex items-center justify-between border-b border-border px-6 py-4">
          <h3 className="text-base font-semibold">
            {isEdit ? "Edit Baseline" : "Tambah Baseline"}
          </h3>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground" aria-label="Close">
            <X className="h-4 w-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 px-6 py-5">
          {formError && (
            <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{formError}</p>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-label">Label</label>
              <input
                id="bl-label"
                name="label"
                className={inputClassName}
                defaultValue={baseline?.label ?? ""}
                placeholder="misal: Baseline Kontrak"
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-currency">Mata Uang</label>
              <input
                id="bl-currency"
                name="currency"
                className={inputClassName}
                defaultValue={baseline?.currency ?? "IDR"}
                placeholder="IDR"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-physical-target">
                Target Fisik (%) <span className="text-destructive">*</span>
              </label>
              <input
                id="bl-physical-target"
                name="physical_target"
                type="number"
                min={0}
                max={100}
                step={0.01}
                className={inputClassName}
                defaultValue={baseline?.physical_target ?? 100}
                required
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-budget-total">
                Anggaran Total <span className="text-destructive">*</span>
              </label>
              <input
                id="bl-budget-total"
                name="budget_total"
                type="number"
                min={0}
                step={1}
                className={inputClassName}
                defaultValue={baseline?.budget_total ?? ""}
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-planned-start">
                Mulai Rencana <span className="text-destructive">*</span>
              </label>
              <input
                id="bl-planned-start"
                name="planned_start"
                type="date"
                className={inputClassName}
                defaultValue={baseline?.planned_start?.slice(0, 10) ?? ""}
                required
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-planned-end">
                Selesai Rencana <span className="text-destructive">*</span>
              </label>
              <input
                id="bl-planned-end"
                name="planned_end"
                type="date"
                className={inputClassName}
                defaultValue={baseline?.planned_end?.slice(0, 10) ?? ""}
                required
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="bl-source">Sumber</label>
            <input
              id="bl-source"
              name="source"
              className={inputClassName}
              defaultValue={baseline?.source ?? ""}
              placeholder="misal: Dokumen Kontrak No. 001"
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="bl-notes">Catatan</label>
            <textarea
              id="bl-notes"
              name="notes"
              rows={2}
              className={cn(inputClassName, "h-auto resize-none")}
              defaultValue={baseline?.notes ?? ""}
            />
          </div>

          {isEdit && (
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-is-active">Status Aktif</label>
              <select
                id="bl-is-active"
                name="is_active"
                className={inputClassName}
                defaultValue={String(baseline?.is_active ?? false)}
              >
                <option value="true">Aktif</option>
                <option value="false">Tidak aktif</option>
              </select>
            </div>
          )}

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
