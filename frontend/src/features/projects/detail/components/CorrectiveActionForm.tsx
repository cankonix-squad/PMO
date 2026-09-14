"use client";

import { cn } from "@/lib/utils";
import type { CorrectiveAction } from "@/types/project";
import { inputClassName, primaryButton, secondaryButton } from "../schemas";
import { toDateInput } from "../utils";

export function CorrectiveActionForm({
  ca,
  isSaving,
  formError,
  onSubmit,
  onClose,
}: {
  ca: CorrectiveAction | null;
  isSaving: boolean;
  formError: string | null;
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void;
  onClose: () => void;
}) {
  const isEdit = ca !== null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-lg rounded-xl border border-border bg-background shadow-xl">
        <div className="flex items-center justify-between border-b border-border px-6 py-4">
          <h3 className="text-base font-semibold">
            {isEdit ? "Edit Tindak Lanjut" : "Tindak Lanjut Baru"}
          </h3>
          <button
            type="button"
            onClick={onClose}
            className="text-muted-foreground hover:text-foreground"
            aria-label="Close"
          >
            ✕
          </button>
        </div>

        <form onSubmit={onSubmit} className="space-y-4 px-6 py-5">
          {formError && (
            <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {formError}
            </p>
          )}

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="ca-title">
              Title <span className="text-destructive">*</span>
            </label>
            <input
              id="ca-title"
              name="title"
              className={inputClassName}
              defaultValue={ca?.title ?? ""}
              placeholder="Judul singkat temuan"
              required
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="ca-deviation">
              Deviation <span className="text-destructive">*</span>
            </label>
            <textarea
              id="ca-deviation"
              name="deviation"
              rows={3}
              className={cn(inputClassName, "h-auto resize-none")}
              defaultValue={ca?.deviation ?? ""}
              placeholder="Jelaskan deviasi atau ketidaksesuaian"
              required
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="ca-root-cause">
              Root Cause
            </label>
            <textarea
              id="ca-root-cause"
              name="root_cause"
              rows={2}
              className={cn(inputClassName, "h-auto resize-none")}
              defaultValue={ca?.root_cause ?? ""}
              placeholder="Analysis of the root cause"
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="ca-recommendation">
              Recommendation
            </label>
            <textarea
              id="ca-recommendation"
              name="recommendation"
              rows={2}
              className={cn(inputClassName, "h-auto resize-none")}
              defaultValue={ca?.recommendation ?? ""}
              placeholder="Rekomendasi tindak lanjut"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="ca-source-type">
                Source Type
              </label>
              <select
                id="ca-source-type"
                name="source_type"
                className={inputClassName}
                defaultValue={ca?.source_type ?? ""}
              >
                <option value="">— none —</option>
                <option value="issue">Issue</option>
                <option value="risk">Risk</option>
                <option value="task">Task</option>
              </select>
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="ca-target-date">
                Target Date
              </label>
              <input
                id="ca-target-date"
                name="target_date"
                type="date"
                className={inputClassName}
                defaultValue={toDateInput(ca?.target_date)}
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="ca-evidence">
              Evidence Note
            </label>
            <textarea
              id="ca-evidence"
              name="evidence_note"
              rows={2}
              className={cn(inputClassName, "h-auto resize-none")}
              defaultValue={ca?.evidence_note ?? ""}
              placeholder="Evidence atau referensi dokumen pendukung"
            />
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className={secondaryButton}>
              Batal
            </button>
            <button type="submit" disabled={isSaving} className={primaryButton}>
              {isSaving ? "Menyimpan..." : isEdit ? "Perbarui" : "Tambah"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
