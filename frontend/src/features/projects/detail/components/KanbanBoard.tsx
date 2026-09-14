"use client";

import type { Task, TaskStatus } from "@/types/project";
import { ErrorState } from "./ErrorState";
import { LoadingState } from "./LoadingState";
import { SectionHeader } from "./SectionHeader";
import { TaskCard } from "./TaskCard";

export function KanbanBoard({
  groups,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onMove,
  onDelete,
}: {
  groups: Array<{ status: TaskStatus; tasks: Task[] }>;
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (task: Task) => void;
  onMove: (task: Task, status: TaskStatus) => void;
  onDelete: (task: Task) => void;
}) {
  return (
    <section className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Papan Kanban" actionLabel="Task Baru" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat papan kerja..." />
      ) : isError ? (
        <ErrorState label="Papan kerja belum dapat dimuat." />
      ) : (
        <div className="grid grid-cols-1 gap-3 p-4 md:grid-cols-2 2xl:grid-cols-5">
          {groups.map((group) => (
            <div key={group.status} className="rounded-md border border-border bg-muted/30">
              <div className="border-b border-border px-3 py-2">
                <p className="text-xs font-semibold text-foreground">
                  {group.status.replace("_", " ")}
                </p>
                <p className="text-xs text-muted-foreground">
                  {group.tasks.length} task{group.tasks.length !== 1 ? "s" : ""}
                </p>
              </div>
              <div className="space-y-3 p-3">
                {group.tasks.length === 0 ? (
                  <p className="rounded-md bg-background px-3 py-4 text-center text-xs text-muted-foreground">
                    Empty
                  </p>
                ) : (
                  group.tasks.map((task) => (
                    <TaskCard
                      key={task.id}
                      task={task}
                      onEdit={onEdit}
                      onMove={onMove}
                      onDelete={onDelete}
                    />
                  ))
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
