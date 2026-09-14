"use client";

import { cn, formatDate } from "@/lib/utils";
import type { CorrectiveAction, CorrectiveActionStatus } from "@/types/project";
import { CA_NEXT_STATUS } from "../schemas";
import { caStatusTone } from "../utils";
import { EmptyState } from "./EmptyState";
import { ErrorState } from "./ErrorState";
import { LoadingState } from "./LoadingState";
import { RowActions } from "./RowActions";
import { SectionHeader } from "./SectionHeader";

export function CorrectiveActionPanel({
  correctiveActions,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onMove,
  onDelete,
}: {
  correctiveActions: CorrectiveAction[];
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (ca: CorrectiveAction) => void;
  onMove: (ca: CorrectiveAction, toStatus: CorrectiveActionStatus) => void;
  onDelete: (ca: CorrectiveAction) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader
        title="Corrective Actions"
        actionLabel="Tindak Lanjut Baru"
        onAction={onCreate}
      />
      {isLoading ? (
        <LoadingState label="Memuat tindak lanjut..." />
      ) : isError ? (
        <ErrorState label="Tindak lanjut belum dapat dimuat." />
      ) : correctiveActions.length === 0 ? (
        <EmptyState label="Belum ada tindak lanjut." />
      ) : (
        <div className="divide-y divide-border">
          {correctiveActions.map((ca) => (
            <div key={ca.id} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="text-sm font-medium text-foreground">{ca.title}</p>
                    <span
                      className={cn(
                        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
                        caStatusTone(ca.status as CorrectiveActionStatus)
                      )}
                    >
                      {ca.status.replace("_", " ")}
                    </span>
                    {ca.source_type && (
                      <span className="inline-flex items-center rounded-full bg-purple-100 px-2 py-0.5 text-xs font-medium text-purple-700">
                        {ca.source_type}
                      </span>
                    )}
                  </div>
                  <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">
                    {ca.deviation}
                  </p>
                  {ca.root_cause && (
                    <p className="mt-1 text-xs text-muted-foreground">
                      <span className="font-medium text-foreground">Root cause:</span>{" "}
                      {ca.root_cause}
                    </p>
                  )}
                  {ca.target_date && (
                    <p className="mt-1 text-xs text-muted-foreground">
                      Target: {formatDate(ca.target_date)}
                    </p>
                  )}
                </div>
                <RowActions
                  onEdit={() => onEdit(ca)}
                  onDelete={() => onDelete(ca)}
                />
              </div>
              <div className="mt-3 flex flex-wrap gap-2">
                {(CA_NEXT_STATUS[ca.status as CorrectiveActionStatus] ?? []).map((toStatus) => (
                  <button
                    key={toStatus}
                    type="button"
                    onClick={() => onMove(ca, toStatus)}
                    className="inline-flex h-7 items-center rounded-md border border-input px-2 text-xs hover:bg-accent"
                  >
                    {toStatus.replace("_", " ")}
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
