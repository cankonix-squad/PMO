"use client";

import { formatDate } from "@/lib/utils";
import type { Project } from "@/types/project";

export function ProjectHeader({ project }: { project: Project | undefined }) {
  if (!project) return null;

  return (
    <div>
      <p className="text-xs font-medium uppercase text-muted-foreground">{project.code}</p>
      <h2 className="text-xl font-semibold text-foreground">{project.name}</h2>
      <p className="mt-1 text-sm text-muted-foreground">
        {project.status.replace("_", " ")} · {project.progress_pct}% progress ·{" "}
        {formatDate(project.end_date)}
      </p>
    </div>
  );
}
