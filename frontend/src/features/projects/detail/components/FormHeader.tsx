"use client";

import { X } from "lucide-react";

export function FormHeader({ title, onCancel }: { title: string; onCancel: () => void }) {
  return (
    <div className="mb-4 flex items-center justify-between gap-3">
      <h3 className="text-sm font-semibold text-foreground">{title}</h3>
      <button
        type="button"
        onClick={onCancel}
        className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
        aria-label="Close form"
      >
        <X className="h-4 w-4" aria-hidden="true" />
      </button>
    </div>
  );
}
