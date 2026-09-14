"use client";

import { formatDate } from "@/lib/utils";
import type { Milestone, MilestoneStatus } from "@/types/project";
import { MILESTONE_NEXT_STATUS } from "../schemas";
import { EmptyState } from "./EmptyState";
import { ErrorState } from "./ErrorState";
import { LoadingState } from "./LoadingState";
import { ProgressBar } from "./ProgressBar";
import { RowActions } from "./RowActions";
import { SectionHeader } from "./SectionHeader";

export function MilestonePanel({
  milestones,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onMove,
  onDelete,
}: {
  milestones: Milestone[];
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (milestone: Milestone) => void;
  onMove: (milestone: Milestone, status: MilestoneStatus) => void;
  onDelete: (milestone: Milestone) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Milestone" actionLabel="Milestone Baru" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat milestone..." />
      ) : isError ? (
        <ErrorState label="Milestone belum dapat dimuat." />
      ) : milestones.length === 0 ? (
        <EmptyState label="Belum ada milestone." />
      ) : (
        <div className="divide-y divide-border">
          {milestones.map((milestone) => (
            <div key={milestone.id} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-sm font-medium text-foreground">{milestone.title}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {milestone.status.replace("_", " ")} · due {formatDate(milestone.due_date)}
                  </p>
                </div>
                <RowActions
                  onEdit={() => onEdit(milestone)}
                  onDelete={() => onDelete(milestone)}
                />
              </div>
              <div className="mt-3">
                <ProgressBar value={milestone.progress_pct} />
              </div>
              <div className="mt-3 flex flex-wrap gap-2">
                {MILESTONE_NEXT_STATUS[milestone.status].map((status) => (
                  <button
                    key={status}
                    type="button"
                    onClick={() => onMove(milestone, status)}
                    className="inline-flex h-7 items-center rounded-md border border-input px-2 text-xs hover:bg-accent"
                  >
                    {status.replace("_", " ")}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </aside>
  );
}
