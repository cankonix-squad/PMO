"use client";

import { formatDate } from "@/lib/utils";
import type { Issue, IssueStatus } from "@/types/project";
import { ISSUE_NEXT_STATUS } from "../schemas";
import { EmptyState } from "./EmptyState";
import { ErrorState } from "./ErrorState";
import { LoadingState } from "./LoadingState";
import { RowActions } from "./RowActions";
import { SectionHeader } from "./SectionHeader";

export function IssuePanel({
  issues,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onMove,
  onDelete,
}: {
  issues: Issue[];
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (issue: Issue) => void;
  onMove: (issue: Issue, status: IssueStatus) => void;
  onDelete: (issue: Issue) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Isu" actionLabel="Isu Baru" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat isu..." />
      ) : isError ? (
        <ErrorState label="Isu belum dapat dimuat." />
      ) : issues.length === 0 ? (
        <EmptyState label="Belum ada isu." />
      ) : (
        <div className="divide-y divide-border">
          {issues.map((issue) => (
            <div key={issue.id} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-sm font-medium text-foreground">{issue.title}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {issue.status.replace("_", " ")} · {issue.severity} ·{" "}
                    {issue.escalation.replace("_", " ")}
                    {issue.due_date ? ` · due ${formatDate(issue.due_date)}` : ""}
                  </p>
                </div>
                <RowActions
                  onEdit={() => onEdit(issue)}
                  onDelete={() => onDelete(issue)}
                />
              </div>
              {issue.description && (
                <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">
                  {issue.description}
                </p>
              )}
              {issue.resolution && (
                <p className="mt-2 text-xs text-muted-foreground">
                  <span className="font-medium text-foreground">Resolution:</span>{" "}
                  {issue.resolution}
                </p>
              )}
              <div className="mt-3 flex flex-wrap gap-2">
                {ISSUE_NEXT_STATUS[issue.status].map((status) => (
                  <button
                    key={status}
                    type="button"
                    onClick={() => onMove(issue, status)}
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
