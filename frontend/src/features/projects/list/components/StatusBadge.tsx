"use client";

import { cn } from "@/lib/utils";
import type { ProjectStatus } from "@/types/project";
import { STATUS_COLORS, STATUS_LABELS } from "../schemas";

export function StatusBadge({ status }: { status: ProjectStatus }) {
  return (
    <span
      className={cn(
        "inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium",
        STATUS_COLORS[status]
      )}
    >
      {STATUS_LABELS[status]}
    </span>
  );
}
