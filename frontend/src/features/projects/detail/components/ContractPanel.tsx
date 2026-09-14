"use client";

import { cn } from "@/lib/utils";
import type { Contract } from "@/types/project";
import { contractStatusTone, formatCurrency, toDateInput } from "../utils";
import { EmptyState } from "./EmptyState";
import { ErrorState } from "./ErrorState";
import { LoadingState } from "./LoadingState";
import { RowActions } from "./RowActions";
import { SectionHeader } from "./SectionHeader";

export function ContractPanel({
  contracts,
  totals,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onDelete,
}: {
  contracts: Contract[];
  totals: {
    totalValue: number;
    activeCount: number;
    mainVendorId: string | null;
    mainValue: number;
    period: string;
  };
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (contract: Contract) => void;
  onDelete: (contract: Contract) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Kontrak" actionLabel="Kontrak Baru" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat kontrak..." />
      ) : isError ? (
        <ErrorState label="Kontrak belum dapat dimuat." />
      ) : contracts.length === 0 ? (
        <EmptyState label="Belum ada kontrak." />
      ) : (
        <div>
          <div className="grid grid-cols-2 gap-2 border-b border-border p-4 text-xs">
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Total Value</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {formatCurrency(totals.totalValue)}
              </p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Kontrak Aktif</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {totals.activeCount}
              </p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Contract Period</p>
              <p className="mt-0.5 font-semibold text-foreground">{totals.period}</p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Main Vendor Value</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {formatCurrency(totals.mainValue)}
              </p>
            </div>
          </div>
          <div className="divide-y divide-border">
            {contracts.map((contract) => (
              <div key={contract.id} className="p-4">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-sm font-medium text-foreground">{contract.title}</p>
                    <p className="mt-0.5 text-xs text-muted-foreground">
                      {contract.contract_number}
                    </p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {contract.vendor?.name ?? ""}
                      {contract.consultant ? ` · Konsultan: ${contract.consultant.name}` : ""}
                      {" · "}
                      {formatCurrency(contract.contract_value)}
                      {" · "}
                      {contract.start_date
                        ? `${toDateInput(contract.start_date)} — ${contract.end_date ? toDateInput(contract.end_date) : "?"}`
                        : "TBD"}
                    </p>
                  </div>
                  <div className="flex flex-col items-end gap-2">
                    <span
                      className={cn(
                        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold",
                        contractStatusTone(contract.status)
                      )}
                    >
                      {contract.status.replace("_", " ")}
                    </span>
                    <RowActions
                      onEdit={() => onEdit(contract)}
                      onDelete={() => onDelete(contract)}
                    />
                  </div>
                </div>
                {contract.scope_of_work && (
                  <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">
                    {contract.scope_of_work}
                  </p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
      <p className="border-t border-border px-4 py-2 text-[10px] text-muted-foreground">
        Operational contract input — belum validated/published (menunggu P1-011/P1-012).
      </p>
    </aside>
  );
}
