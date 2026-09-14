"use client";

import { CalendarDays } from "lucide-react";

export function LoadingState({ label }: { label: string }) {
  return (
    <div className="flex items-center gap-3 p-6 text-sm text-muted-foreground">
      <CalendarDays className="h-4 w-4" aria-hidden="true" />
      {label}
    </div>
  );
}
