"use client";

import { formatDate } from "@/lib/utils";
import type { Risk, RiskStatus } from "@/types/project";
import { RISK_NEXT_STATUS } from "../schemas";
import { riskSeverityColor } from "../utils";
import { EmptyState } from "./EmptyState";
import { ErrorState } from "./ErrorState";
import { LoadingState } from "./LoadingState";
import { RowActions } from "./RowActions";
import { SectionHeader } from "./SectionHeader";

export function RiskPanel({
  risks,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onMove,
  onDelete,
}: {
  risks: Risk[];
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (risk: Risk) => void;
  onMove: (risk: Risk, status: RiskStatus) => void;
  onDelete: (risk: Risk) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Risiko" actionLabel="Risiko Baru" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat risiko..." />
      ) : isError ? (
        <ErrorState label="Risiko belum dapat dimuat." />
      ) : risks.length === 0 ? (
        <EmptyState label="Belum ada risiko." />
      ) : (
        <div className="divide-y divide-border">
          {risks.map((risk) => (
            <div key={risk.id} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-sm font-medium text-foreground">{risk.title}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {risk.status.replace("_", " ")} ·{" "}
                    <span className={riskSeverityColor(risk.severity)}>
                      {risk.severity}
                    </span>{" "}
                    · score {risk.risk_score} ({risk.probability}×{risk.impact})
                    {risk.due_date ? ` · due ${formatDate(risk.due_date)}` : ""}
                  </p>
                </div>
                <RowActions
                  onEdit={() => onEdit(risk)}
                  onDelete={() => onDelete(risk)}
                />
              </div>
              {risk.description && (
                <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">
                  {risk.description}
                </p>
              )}
              {risk.mitigation && (
                <p className="mt-2 text-xs text-muted-foreground">
                  <span className="font-medium text-foreground">Mitigation:</span>{" "}
                  {risk.mitigation}
                </p>
              )}
              <div className="mt-3 flex flex-wrap gap-2">
                {RISK_NEXT_STATUS[risk.status].map((status) => (
                  <button
                    key={status}
                    type="button"
                    onClick={() => onMove(risk, status)}
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
