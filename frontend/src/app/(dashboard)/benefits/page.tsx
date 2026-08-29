"use client";

import { type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { BarChart3, Plus, RefreshCw, Trash2 } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { benefitService } from "@/services/benefit.service";
import { projectService } from "@/services/project.service";
import { cn } from "@/lib/utils";
import type {
  BenefitAggregationMethod,
  BenefitIndicator,
  BenefitValidationStatus,
  CreateBenefitIndicatorRequest,
  CreateBenefitMeasurementRequest,
} from "@/types/benefit";

const aggregationMethods: BenefitAggregationMethod[] = ["SUM", "AVERAGE", "LATEST"];
const validationStatuses: BenefitValidationStatus[] = ["DRAFT", "SUBMITTED", "VALID", "REJECTED", "STALE"];

export default function BenefitsPage() {
  const queryClient = useQueryClient();
  const benefitsQuery = useQuery({ queryKey: ["benefits"], queryFn: benefitService.list });
  const summaryQuery = useQuery({ queryKey: ["benefit-summary"], queryFn: benefitService.summary });
  const projectsQuery = useQuery({
    queryKey: ["projects", "benefit-select"],
    queryFn: async () => (await projectService.list({ page_size: 100 })).data.data,
  });

  const createIndicator = useMutation({
    mutationFn: (payload: CreateBenefitIndicatorRequest) => benefitService.create(payload),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["benefits"] }),
        queryClient.invalidateQueries({ queryKey: ["benefit-summary"] }),
      ]);
    },
  });

  function handleCreateIndicator(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const payload: CreateBenefitIndicatorRequest = {
      name: String(form.get("name") ?? ""),
      unit: String(form.get("unit") ?? ""),
      aggregation_method: String(form.get("aggregation_method") ?? "SUM") as BenefitAggregationMethod,
      source: String(form.get("source") ?? ""),
      description: String(form.get("description") ?? ""),
    };
    const projectID = String(form.get("project_id") ?? "");
    if (projectID) payload.project_id = projectID;
    createIndicator.mutate(payload, { onSuccess: () => event.currentTarget.reset() });
  }

  const indicators = benefitsQuery.data ?? [];
  const summary = summaryQuery.data ?? [];

  return (
    <DashboardLayout title="Indikator Manfaat">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">
          Indikator manfaat berbasis measurement tervalidasi, dengan agregasi dipisahkan per unit.
        </p>
        <button
          type="button"
          onClick={() => {
            void benefitsQuery.refetch();
            void summaryQuery.refetch();
          }}
          className="inline-flex h-9 w-9 items-center justify-center rounded-md border"
          aria-label="Muat ulang indikator manfaat"
          title="Muat ulang"
        >
          <RefreshCw className={cn("h-4 w-4", (benefitsQuery.isFetching || summaryQuery.isFetching) && "animate-spin")} />
        </button>
      </div>

      <section className="mt-5 grid gap-3 md:grid-cols-3">
        {summary.length === 0 ? (
          <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground md:col-span-3">
            Belum ada benefit tervalidasi untuk diagregasi.
          </div>
        ) : (
          summary.map((item) => (
            <div key={`${item.unit}-${item.aggregation_method}`} className="rounded-md border bg-card p-4">
              <p className="text-xs uppercase text-muted-foreground">{item.unit} / {item.aggregation_method}</p>
              <p className="mt-2 text-2xl font-semibold">{item.value.toLocaleString("id-ID")}</p>
              <p className="mt-1 text-xs text-muted-foreground">{item.count} indikator kompatibel</p>
            </div>
          ))
        )}
      </section>

      <form onSubmit={handleCreateIndicator} className="mt-5 grid gap-3 rounded-md border bg-card p-4 lg:grid-cols-6">
        <input name="name" required placeholder="Nama indikator" className="rounded-md border px-3 py-2 text-sm lg:col-span-2" />
        <input name="unit" required placeholder="Unit, mis. Ha" className="rounded-md border px-3 py-2 text-sm" />
        <select name="aggregation_method" defaultValue="SUM" className="rounded-md border px-3 py-2 text-sm">
          {aggregationMethods.map((method) => <option key={method}>{method}</option>)}
        </select>
        <select name="project_id" defaultValue="" className="rounded-md border px-3 py-2 text-sm">
          <option value="">Level portofolio</option>
          {(projectsQuery.data ?? []).map((project) => (
            <option key={project.id} value={project.id}>{project.code} - {project.name}</option>
          ))}
        </select>
        <input name="source" placeholder="Sumber data" className="rounded-md border px-3 py-2 text-sm" />
        <textarea name="description" placeholder="Deskripsi" className="rounded-md border px-3 py-2 text-sm lg:col-span-5" />
        <button type="submit" disabled={createIndicator.isPending} className="inline-flex items-center justify-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-60">
          <Plus className="h-4 w-4" /> Simpan
        </button>
      </form>

      {benefitsQuery.isLoading && <p className="mt-6 text-sm text-muted-foreground">Memuat indikator manfaat...</p>}
      {benefitsQuery.isError && <p className="mt-6 text-sm text-rose-700">Indikator manfaat belum dapat dimuat.</p>}
      {!benefitsQuery.isLoading && !benefitsQuery.isError && indicators.length === 0 && (
        <p className="mt-6 rounded-md border border-dashed p-8 text-center text-sm text-muted-foreground">
          Belum ada indikator manfaat. Tambahkan indikator untuk mulai mencatat outcome.
        </p>
      )}
      <div className="mt-6 grid gap-4 xl:grid-cols-2">
        {indicators.map((indicator) => <IndicatorCard key={indicator.id} indicator={indicator} />)}
      </div>
    </DashboardLayout>
  );
}

function IndicatorCard({ indicator }: { indicator: BenefitIndicator }) {
  const queryClient = useQueryClient();
  const measurementsQuery = useQuery({
    queryKey: ["benefit-measurements", indicator.id],
    queryFn: () => benefitService.listMeasurements(indicator.id),
  });
  const aggregateQuery = useQuery({
    queryKey: ["benefit-aggregate", indicator.id],
    queryFn: () => benefitService.aggregate(indicator.id),
  });
  const createMeasurement = useMutation({
    mutationFn: (payload: CreateBenefitMeasurementRequest) => benefitService.createMeasurement(indicator.id, payload),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["benefit-measurements", indicator.id] }),
        queryClient.invalidateQueries({ queryKey: ["benefit-aggregate", indicator.id] }),
        queryClient.invalidateQueries({ queryKey: ["benefit-summary"] }),
      ]);
    },
  });
  const deleteIndicator = useMutation({
    mutationFn: () => benefitService.delete(indicator.id),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["benefits"] }),
        queryClient.invalidateQueries({ queryKey: ["benefit-summary"] }),
      ]);
    },
  });
  const deleteMeasurement = useMutation({
    mutationFn: (measurementID: string) => benefitService.deleteMeasurement(indicator.id, measurementID),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["benefit-measurements", indicator.id] }),
        queryClient.invalidateQueries({ queryKey: ["benefit-aggregate", indicator.id] }),
        queryClient.invalidateQueries({ queryKey: ["benefit-summary"] }),
      ]);
    },
  });

  function handleMeasurement(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const payload: CreateBenefitMeasurementRequest = {
      period_year: Number(form.get("period_year")),
      period_month: Number(form.get("period_month")),
      baseline: Number(form.get("baseline")),
      target: Number(form.get("target")),
      actual: Number(form.get("actual")),
      validation_status: String(form.get("validation_status") ?? "DRAFT") as BenefitValidationStatus,
      source: String(form.get("source") ?? ""),
    };
    createMeasurement.mutate(payload, { onSuccess: () => event.currentTarget.reset() });
  }

  const measurements = measurementsQuery.data ?? [];
  const aggregate = aggregateQuery.data;

  return (
    <article className="rounded-md border bg-card p-4 shadow-sm">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="font-semibold">{indicator.name}</h2>
          <p className="mt-1 text-xs text-muted-foreground">{indicator.unit} · {indicator.aggregation_method}</p>
        </div>
        <div className="flex items-center gap-2">
          <BarChart3 className="h-5 w-5 text-primary" />
          <button
            type="button"
            onClick={() => {
              if (window.confirm(`Hapus indikator ${indicator.name}?`)) deleteIndicator.mutate();
            }}
            className="rounded-md border p-2"
            aria-label="Hapus indikator manfaat"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      </div>
      <p className="mt-4 text-2xl font-bold">{aggregate?.value === null || aggregate?.value === undefined ? "N/A" : aggregate.value.toLocaleString("id-ID")}</p>
      <p className="mt-1 text-xs text-muted-foreground">{aggregate?.count ?? 0} measurement tervalidasi</p>

      <form onSubmit={handleMeasurement} className="mt-4 grid gap-2 md:grid-cols-4">
        <input name="period_year" required type="number" min={2000} max={2100} placeholder="Tahun" className="rounded-md border px-3 py-2 text-sm" />
        <input name="period_month" required type="number" min={1} max={12} placeholder="Bulan" className="rounded-md border px-3 py-2 text-sm" />
        <input name="baseline" required type="number" step="0.01" placeholder="Baseline" className="rounded-md border px-3 py-2 text-sm" />
        <input name="target" required type="number" step="0.01" placeholder="Target" className="rounded-md border px-3 py-2 text-sm" />
        <input name="actual" required type="number" step="0.01" placeholder="Realisasi" className="rounded-md border px-3 py-2 text-sm" />
        <select name="validation_status" defaultValue="DRAFT" className="rounded-md border px-3 py-2 text-sm">
          {validationStatuses.map((status) => <option key={status}>{status}</option>)}
        </select>
        <input name="source" placeholder="Sumber data" className="rounded-md border px-3 py-2 text-sm" />
        <button type="submit" disabled={createMeasurement.isPending} className="inline-flex items-center justify-center gap-2 rounded-md border px-3 py-2 text-sm disabled:opacity-60">
          <Plus className="h-4 w-4" /> Tambah
        </button>
      </form>

      <div className="mt-4 overflow-x-auto">
        <table className="w-full min-w-[560px] text-left text-sm">
          <thead className="text-xs uppercase text-muted-foreground">
            <tr>
              <th className="py-2">Periode</th>
              <th>Baseline</th>
              <th>Target</th>
              <th>Realisasi</th>
              <th>Status</th>
              <th className="w-10"><span className="sr-only">Aksi</span></th>
            </tr>
          </thead>
          <tbody>
            {measurements.map((measurement) => (
              <tr key={measurement.id} className="border-t">
                <td className="py-2">{measurement.period_year}-{String(measurement.period_month).padStart(2, "0")}</td>
                <td>{measurement.baseline.toLocaleString("id-ID")}</td>
                <td>{measurement.target.toLocaleString("id-ID")}</td>
                <td>{measurement.actual.toLocaleString("id-ID")}</td>
                <td>{measurement.validation_status}</td>
                <td>
                  <button
                    type="button"
                    onClick={() => {
                      if (window.confirm("Hapus measurement ini?")) deleteMeasurement.mutate(measurement.id);
                    }}
                    className="rounded-md border p-1"
                    aria-label="Hapus measurement manfaat"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </article>
  );
}
