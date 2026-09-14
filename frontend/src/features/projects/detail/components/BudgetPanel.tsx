"use client";

import { cn } from "@/lib/utils";
import type { BudgetLine } from "@/types/project";
import { budgetStatusTone, formatCurrency } from "../utils";
import { EmptyState } from "./EmptyState";
import { ErrorState } from "./ErrorState";
import { LoadingState } from "./LoadingState";
import { RowActions } from "./RowActions";
import { SectionHeader } from "./SectionHeader";

export function BudgetPanel({
  budgets,
  totals,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onDelete,
}: {
  budgets: BudgetLine[];
  totals: { planned: number; actual: number; variance: number; usage: number };
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (budget: BudgetLine) => void;
  onDelete: (budget: BudgetLine) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Anggaran" actionLabel="Baris Anggaran Baru" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat anggaran..." />
      ) : isError ? (
        <ErrorState label="Anggaran belum dapat dimuat." />
      ) : budgets.length === 0 ? (
        <EmptyState label="Belum ada baris anggaran." />
      ) : (
        <div>
          <div className="grid grid-cols-2 gap-2 border-b border-border p-4 text-xs">
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Total Planned</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {formatCurrency(totals.planned)}
              </p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Total Actual</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {formatCurrency(totals.actual)}
              </p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Variance</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {formatCurrency(totals.variance)}
              </p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Usage</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {totals.usage}%
              </p>
            </div>
          </div>
          <div className="divide-y divide-border">
            {budgets.map((budget) => (
              <div key={budget.id} className="p-4">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-sm font-medium text-foreground">{budget.category}</p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Planned {formatCurrency(budget.planned)} · Actual{" "}
                      {formatCurrency(budget.actual)} · Variance{" "}
                      {formatCurrency(budget.variance)} · Usage {budget.usage_pct}%
                    </p>
                  </div>
                  <div className="flex flex-col items-end gap-2">
                    <span
                      className={cn(
                        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold",
                        budgetStatusTone(budget.status)
                      )}
                    >
                      {budget.status.replace("_", " ")}
                    </span>
                    <RowActions
                      onEdit={() => onEdit(budget)}
                      onDelete={() => onDelete(budget)}
                    />
                  </div>
                </div>
                {budget.description && (
                  <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">
                    {budget.description}
                  </p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
      <p className="border-t border-border px-4 py-2 text-[10px] text-muted-foreground">
        Operational input budget/realisasi — belum validated/published (menunggu P1-011/P1-012).
      </p>
    </aside>
  );
}
