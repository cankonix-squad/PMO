"use client";

import { AlertCircle } from "lucide-react";

export function ErrorState({ label }: { label: string }) {
  return (
    <div className="flex items-start gap-3 p-6 text-sm text-destructive">
      <AlertCircle className="mt-0.5 h-4 w-4" aria-hidden="true" />
      {label}
    </div>
  );
}
