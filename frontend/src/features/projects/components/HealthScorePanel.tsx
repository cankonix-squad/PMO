"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Activity, RefreshCw } from "lucide-react";
import { healthService } from "@/services/health.service";
import type { HealthClass } from "@/types/health";
import { cn, formatDate } from "@/lib/utils";

const classStyle: Record<HealthClass, string> = {
  GREEN: "bg-emerald-100 text-emerald-800",
  YELLOW: "bg-amber-100 text-amber-800",
  RED: "bg-rose-100 text-rose-800",
  CRITICAL: "bg-red-700 text-white",
};

const classLabel: Record<HealthClass, string> = {
  GREEN: "Hijau",
  YELLOW: "Kuning",
  RED: "Merah",
  CRITICAL: "Kritis",
};

export function HealthScorePanel({ projectId }: { projectId: string }) {
  const qc = useQueryClient();
  const now = new Date();
  const [year, setYear] = useState(String(now.getFullYear()));
  const [month, setMonth] = useState(String(now.getMonth() + 1));
  const [error, setError] = useState<string | null>(null);

  const query = useQuery({
    queryKey: ["projects", projectId, "health"],
    queryFn: () => healthService.listSnapshots(projectId),
  });

  const calculate = useMutation({
    mutationFn: () =>
      healthService.calculate(projectId, {
        period_year: Number(year),
        period_month: Number(month),
      }),
    onSuccess: () => {
      setError(null);
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "health"] });
    },
    onError: (e: Error) => setError(e.message),
  });

  const snapshots = query.data ?? [];

  return (
    <section className="mt-6 rounded-lg border bg-card p-5 shadow-sm">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <h3 className="flex items-center gap-2 text-base font-semibold text-foreground">
            <Activity className="h-4 w-4" />
            Skor Kesehatan Proyek
          </h3>
          <p className="mt-1 text-xs leading-5 text-muted-foreground">
            Formula bersifat versioned; snapshot historis tidak berubah setelah
            perhitungan baru.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <input
            aria-label="Tahun kesehatan"
            type="number"
            min="2000"
            max="2100"
            value={year}
            onChange={(e) => setYear(e.target.value)}
            className="h-9 w-24 rounded-md border bg-background px-2 text-sm"
          />
          <input
            aria-label="Bulan kesehatan"
            type="number"
            min="1"
            max="12"
            value={month}
            onChange={(e) => setMonth(e.target.value)}
            className="h-9 w-16 rounded-md border bg-background px-2 text-sm"
          />
          <button
            type="button"
            disabled={calculate.isPending}
            onClick={() => calculate.mutate()}
            className="inline-flex h-9 items-center gap-2 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground disabled:cursor-not-allowed disabled:opacity-50"
          >
            <RefreshCw className={cn("h-4 w-4", calculate.isPending && "animate-spin")} />
            {calculate.isPending ? "Menghitung..." : "Hitung"}
          </button>
        </div>
      </div>

      {error && (
        <p className="mt-3 rounded-md border border-rose-200 bg-rose-50 p-3 text-sm text-rose-700">
          {error}
        </p>
      )}

      {query.isLoading && (
        <p className="mt-4 rounded-md bg-muted/50 p-4 text-sm text-muted-foreground">
          Memuat snapshot kesehatan...
        </p>
      )}

      {query.isError && (
        <p className="mt-4 rounded-md border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">
          Snapshot kesehatan belum dapat dimuat.
        </p>
      )}

      {!query.isLoading && !query.isError && snapshots.length === 0 && (
        <p className="mt-4 rounded-md border border-dashed p-6 text-center text-sm text-muted-foreground">
          Belum ada snapshot kesehatan. Aktifkan formula sebelum menghitung.
        </p>
      )}

      <div className="mt-4 space-y-3">
        {snapshots.map((snapshot) => (
          <article key={snapshot.id} className="rounded-md border p-3">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <div className="min-w-0">
                <p className="font-semibold text-foreground">
                  {snapshot.period_year}-{String(snapshot.period_month).padStart(2, "0")}
                </p>
                <p className="mt-1 text-xs leading-5 text-muted-foreground">
                  {snapshot.explanation} · {formatDate(snapshot.calculated_at)}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xl font-bold text-foreground">{snapshot.score}</span>
                <span
                  className={cn(
                    "rounded-full px-2 py-1 text-xs font-semibold",
                    classStyle[snapshot.health_class]
                  )}
                >
                  {classLabel[snapshot.health_class]}
                </span>
              </div>
            </div>

            <div className="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-4">
              {Object.entries(snapshot.components).map(([name, component]) => (
                <div key={name} className="rounded border bg-background p-2">
                  <p className="text-[11px] capitalize text-muted-foreground">
                    {name.replace(/_/g, " ")}
                  </p>
                  <p className="font-semibold text-foreground">
                    {component.available ? component.score : "N/A"}
                  </p>
                  <p className="truncate text-[10px] text-muted-foreground" title={component.reason}>
                    {component.reason}
                  </p>
                </div>
              ))}
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
