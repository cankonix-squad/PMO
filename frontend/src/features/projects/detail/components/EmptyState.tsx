"use client";

import { CheckCircle2 } from "lucide-react";

export function EmptyState({ label }: { label: string }) {
  return (
    <div className="flex items-start gap-3 p-6 text-sm text-muted-foreground">
      <CheckCircle2 className="mt-0.5 h-4 w-4 text-green-700" aria-hidden="true" />
      {label}
    </div>
  );
}
