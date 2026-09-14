"use client";

import type { Task, TaskStatus } from "@/types/project";
import { TASK_NEXT_STATUS } from "../schemas";
import { ProgressBar } from "./ProgressBar";
import { RowActions } from "./RowActions";

export function TaskCard({
  task,
  onEdit,
  onMove,
  onDelete,
}: {
  task: Task;
  onEdit: (task: Task) => void;
  onMove: (task: Task, status: TaskStatus) => void;
  onDelete: (task: Task) => void;
}) {
  return (
    <article className="rounded-md border border-border bg-background p-3 shadow-sm">
      <div className="flex items-start justify-between gap-2">
        <div>
          <h4 className="text-sm font-medium text-foreground">{task.title}</h4>
          <p className="mt-1 text-xs text-muted-foreground">{task.priority}</p>
        </div>
        <RowActions onEdit={() => onEdit(task)} onDelete={() => onDelete(task)} />
      </div>
      <div className="mt-3">
        <ProgressBar value={task.progress_pct} />
      </div>
      <div className="mt-3 flex flex-wrap gap-2">
        {TASK_NEXT_STATUS[task.status].map((status) => (
          <button
            key={status}
            type="button"
            onClick={() => onMove(task, status)}
            className="inline-flex h-7 items-center rounded-md border border-input px-2 text-xs hover:bg-accent"
          >
            {status.replace("_", " ")}
          </button>
        ))}
      </div>
    </article>
  );
}
