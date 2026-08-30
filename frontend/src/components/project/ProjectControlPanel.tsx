"use client";

import { useQuery } from "@tanstack/react-query";
import { AlertCircle, FileCheck2, RefreshCw } from "lucide-react";
import { projectControlService } from "@/services/project-control.service";
import { cn } from "@/lib/utils";

type ControlItem = {
  id: string;
  title: string;
  status: string;
  severity?: string;
};

export function ProjectControlPanel({ projectId }: { projectId: string }) {
  const query = useQuery({
    queryKey: ["projects", projectId, "control"],
    queryFn: () => projectControlService.get(projectId),
  });
  const data = query.data;

  return (
    <section className="mt-6 rounded-lg border bg-card p-5 shadow-sm">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <h3 className="text-base font-semibold text-foreground">Kontrol Proyek Level 3</h3>
          <p className="mt-1 text-xs leading-5 text-muted-foreground">
            Rekonsiliasi kontrak, snapshot tervalidasi, kesehatan proyek, evidence,
            isu, risiko, dan tindak lanjut.
          </p>
        </div>
        <button
          type="button"
          onClick={() => void query.refetch()}
          className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md border text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
          aria-label="Muat ulang kontrol proyek"
          title="Muat ulang"
        >
          <RefreshCw className={cn("h-4 w-4", query.isFetching && "animate-spin")} />
        </button>
      </div>

      {query.isLoading && (
        <p className="mt-4 rounded-md bg-muted/50 p-4 text-sm text-muted-foreground">
          Memuat kontrol proyek...
        </p>
      )}

      {query.isError && (
        <p className="mt-4 rounded-md border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">
          Data kontrol proyek belum dapat dimuat.
        </p>
      )}

      {data && (
        <>
          <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <MetricCard
              label="Fisik tervalidasi"
              value={data.snapshot ? `${data.snapshot.physical_actual}%` : "N/A"}
              detail={`Deviasi ${data.snapshot?.physical_variance ?? "-"} · ${
                data.snapshot?.status ?? "Belum ada snapshot valid"
              }`}
            />
            <MetricCard
              label="Realisasi keuangan"
              value={
                data.snapshot
                  ? `${data.snapshot.financial_actual.toLocaleString("id-ID")} ${
                      data.snapshot.currency ?? data.project.currency
                    }`
                  : "N/A"
              }
              detail={`Deviasi ${data.snapshot?.financial_variance ?? "-"}`}
            />
            <MetricCard
              label="Kesehatan"
              value={data.health?.score ?? "N/A"}
              detail={data.health?.class ?? "Belum ada snapshot"}
            />
            <MetricCard
              label="Kontrak / evidence"
              value={`${data.contract.count} / ${data.evidence.evidence_files}`}
              detail={`${data.evidence.verified_inspections} inspeksi terverifikasi`}
            />
          </div>

          <div className="mt-4 grid gap-3 md:grid-cols-3">
            <ControlList title="Isu" items={data.issues} />
            <ControlList title="Risiko" items={data.risks} />
            <ControlList title="Tindak lanjut" items={data.actions} />
          </div>

          <p className="mt-4 text-[11px] leading-5 text-muted-foreground">
            Per {new Date(data.as_of).toLocaleString("id-ID")}. Hanya snapshot
            berstatus VALID yang ditampilkan sebagai data resmi.
          </p>
        </>
      )}
    </section>
  );
}

function MetricCard({
  label,
  value,
  detail,
}: {
  label: string;
  value: string | number;
  detail: string;
}) {
  return (
    <div className="rounded-md border bg-background p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 truncate text-xl font-bold text-foreground">{value}</p>
      <p className="mt-1 text-xs leading-4 text-muted-foreground">{detail}</p>
    </div>
  );
}

function ControlList({
  title,
  items = [],
}: {
  title: string;
  items?: ControlItem[] | null;
}) {
  const safeItems = items ?? [];

  return (
    <div className="rounded-md border bg-background p-3">
      <h4 className="flex items-center gap-2 text-sm font-semibold text-foreground">
        <AlertCircle className="h-4 w-4" />
        {title}
      </h4>
      {safeItems.length === 0 ? (
        <p className="mt-3 text-xs text-muted-foreground">Tidak ada item aktif.</p>
      ) : (
        <ul className="mt-3 space-y-2">
          {safeItems.slice(0, 5).map((item) => (
            <li key={item.id} className="flex items-start gap-2 text-xs">
              <FileCheck2 className="mt-0.5 h-3.5 w-3.5 shrink-0 text-muted-foreground" />
              <span className="min-w-0">
                <span className="block truncate font-medium text-foreground">{item.title}</span>
                <span className="text-muted-foreground">
                  {formatStatus(item.status)}
                  {item.severity ? ` · ${formatStatus(item.severity)}` : ""}
                </span>
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function formatStatus(value: string) {
  return value
    .toLowerCase()
    .replace(/_/g, " ")
    .replace(/\b\w/g, (letter) => letter.toUpperCase());
}
