"use client";

import { Save } from "lucide-react";
import { primaryButton } from "../schemas";

export function FormActions({ isSaving, onCancel }: { isSaving: boolean; onCancel: () => void }) {
  return (
    <div className="mt-4 flex justify-end gap-2">
      <button
        type="button"
        onClick={onCancel}
        className="inline-flex h-9 items-center rounded-md border border-input px-3 text-sm hover:bg-accent"
      >
        Batal
      </button>
      <button type="submit" disabled={isSaving} className={primaryButton}>
        <Save className="h-4 w-4" aria-hidden="true" />
        {isSaving ? "Menyimpan..." : "Simpan"}
      </button>
    </div>
  );
}
