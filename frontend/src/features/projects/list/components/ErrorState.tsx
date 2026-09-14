"use client";

import { AlertCircle } from "lucide-react";

export function ErrorState({ message }: { message: string }) {
  return (
    <div className="flex items-center gap-3 p-6 text-sm text-destructive">
      <AlertCircle className="h-4 w-4" aria-hidden="true" />
      {message}
    </div>
  );
}
