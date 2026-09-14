"use client";

import { formatDate } from "@/lib/utils";
import type { Task, TaskStatus } from "@/types/project";
import { TASK_NEXT_STATUS } from "../schemas";
import { EmptyState } from "./EmptyState";
import { ErrorState } from "./ErrorState";
import { LoadingState } from "./LoadingState";
import { RowActions } from "./RowActions";
import { SectionHeader } from "./SectionHeader";

export function TaskList({
  tasks,
  isLoading,
  isError,
  onEdit,
  onMove,
  onDelete,
}: {
  tasks: Task[];
  isLoading: boolean;
  isError: boolean;
  onEdit: (task: Task) => void;
  onMove: (task: Task, status: TaskStatus) => void;
  onDelete: (task: Task) => void;
}) {
  return (
    <section className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Task List" />
      {isLoading ? (
        <LoadingState label="Memuat task..." />
      ) : isError ? (
        <ErrorState label="Task belum dapat dimuat." />
      ) : tasks.length === 0 ? (
        <EmptyState label="Belum ada task." />
      ) : (
        <div className="divide-y divide-border">
          {tasks.map((task) => (
            <div key={task.id} className="p-4">
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <p className="text-sm font-medium text-foreground">{task.title}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {task.status.replace("_", " ")} · due {formatDate(task.due_date)}
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  {TASK_NEXT_STATUS[task.status].map((status) => (
                    <button
                      key={status}
                      type="button"
                      onClick={() => onMove(task, status)}
                      className="inline-flex h-8 items-center rounded-md border border-input px-2 text-xs hover:bg-accent"
                    >
                      {status.replace("_", " ")}
                    </button>
                  ))}
                  <RowActions onEdit={() => onEdit(task)} onDelete={() => onDelete(task)} />
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
