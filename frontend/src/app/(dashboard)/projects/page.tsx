"use client";

import Link from "next/link";
import { type FormEvent, useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AlertCircle,
  CheckCircle2,
  Edit3,
  Filter,
  History,
  Plus,
  Save,
  Search,
  X,
} from "lucide-react";
import { z } from "zod";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { projectService } from "@/services/project.service";
import { orgUnitService } from "@/services/org-unit.service";
import type { OrgUnit } from "@/types/org-unit";
import { OrgUnitLevelLabel } from "@/types/org-unit";
import { programService } from "@/services/portfolio.service";
import { sectorService, regionService, riverBasinService } from "@/services/spatial.service";
import type { Program } from "@/types/portfolio";
import type { Sector, Region, RiverBasin } from "@/types/spatial";
import { cn, formatDate, formatIDR } from "@/lib/utils";
import type {
  CreateProjectRequest,
  Priority,
  Project,
  ProjectStatus,
  TransitionProjectRequest,
  UpdateProjectRequest,
} from "@/types/project";

const STATUS_COLORS: Record<ProjectStatus, string> = {
  DRAFT: "bg-gray-100 text-gray-700",
  PLANNING: "bg-blue-100 text-blue-700",
  ACTIVE: "bg-green-100 text-green-700",
  ON_HOLD: "bg-yellow-100 text-yellow-700",
  COMPLETED: "bg-emerald-100 text-emerald-700",
  CANCELLED: "bg-red-100 text-red-700",
};

const PRIORITIES: Priority[] = ["LOW", "MEDIUM", "HIGH", "CRITICAL"];

const STATUS_LABELS: Record<ProjectStatus, string> = {
  DRAFT: "Draft",
  PLANNING: "Perencanaan",
  ACTIVE: "Aktif",
  ON_HOLD: "Ditunda",
  COMPLETED: "Selesai",
  CANCELLED: "Dibatalkan",
};

const PRIORITY_LABELS: Record<Priority, string> = {
  LOW: "Rendah",
  MEDIUM: "Sedang",
  HIGH: "Tinggi",
  CRITICAL: "Kritis",
};

const NEXT_STATUS: Record<ProjectStatus, ProjectStatus[]> = {
  DRAFT: ["PLANNING", "CANCELLED"],
  PLANNING: ["ACTIVE", "CANCELLED"],
  ACTIVE: ["ON_HOLD", "COMPLETED", "CANCELLED"],
  ON_HOLD: ["ACTIVE", "CANCELLED"],
  COMPLETED: ["ACTIVE"],
  CANCELLED: [],
};

const projectFormSchema = z.object({
  code: z.string().trim().min(1, "Kode wajib diisi").max(100),
  name: z.string().trim().min(1, "Nama proyek wajib diisi").max(500),
  description: z.string().trim().optional(),
  objectives: z.string().trim().optional(),
  priority: z.enum(["LOW", "MEDIUM", "HIGH", "CRITICAL"]),
  category: z.string().trim().optional(),
  start_date: z.string().trim().optional(),
  end_date: z.string().trim().optional(),
  budget_total: z.coerce.number().min(0, "Anggaran tidak boleh negatif"),
  currency: z.string().trim().min(1).max(10),
  progress_pct: z.coerce.number().min(0).max(100),
  org_unit_id: z.string().trim().optional(),
  program_id: z.string().trim().optional(),
  sector_id: z.string().trim().optional(),
  region_id: z.string().trim().optional(),
  river_basin_id: z.string().trim().optional(),
});

type ProjectFormValues = z.infer<typeof projectFormSchema>;
type FormMode = "create" | "edit";

export default function ProjectsPage() {
  const qc = useQueryClient();
  const [search, setSearch] = useState("");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [formMode, setFormMode] = useState<FormMode>("create");

  const { data: orgUnits = [] } = useQuery({
    queryKey: ["org-units"],
    queryFn: () => orgUnitService.listOrgUnits(false),
  });

  const { data: programs = [] } = useQuery({
    queryKey: ["programs", false],
    queryFn: () => programService.list(false),
  });

  const { data: sectors = [] } = useQuery({
    queryKey: ["sectors", false],
    queryFn: () => sectorService.list(false),
  });

  const { data: regions = [] } = useQuery({
    queryKey: ["regions", false],
    queryFn: () => regionService.list(false),
  });

  const { data: riverBasins = [] } = useQuery({
    queryKey: ["river-basins", false],
    queryFn: () => riverBasinService.list(false),
  });
  const [showForm, setShowForm] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const projectListQuery = useQuery({
    queryKey: ["projects", { search }],
    queryFn: () =>
      projectService.list({ search: search || undefined, page: 1, page_size: 50 }),
    select: (res) => res.data.data ?? [],
  });

  const projects = useMemo(() => projectListQuery.data ?? [], [projectListQuery.data]);

  useEffect(() => {
    if (!selectedId && projects.length > 0) {
      setSelectedId(projects[0].id);
    }
    if (selectedId && projects.length > 0 && !projects.some((p) => p.id === selectedId)) {
      setSelectedId(projects[0].id);
    }
  }, [projects, selectedId]);

  const detailQuery = useQuery({
    queryKey: ["projects", selectedId],
    queryFn: () => projectService.get(selectedId ?? ""),
    select: (res) => res.data.data,
    enabled: Boolean(selectedId),
  });

  const historyQuery = useQuery({
    queryKey: ["projects", selectedId, "progress-history"],
    queryFn: () => projectService.getProgressHistory(selectedId ?? ""),
    select: (res) => res.data.data ?? [],
    enabled: Boolean(selectedId),
  });

  const selectedProject = detailQuery.data ?? projects.find((p) => p.id === selectedId) ?? null;

  const createMutation = useMutation({
    mutationFn: (payload: CreateProjectRequest) => projectService.create(payload),
    onSuccess: (res) => {
      const project = res.data.data;
      setSelectedId(project.id);
      setShowForm(false);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["projects"] });
      void qc.invalidateQueries({ queryKey: ["dashboard", "stats"] });
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateProjectRequest }) =>
      projectService.update(id, payload),
    onSuccess: (res) => {
      const project = res.data.data;
      setSelectedId(project.id);
      setShowForm(false);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["projects"] });
      void qc.invalidateQueries({ queryKey: ["projects", project.id] });
      void qc.invalidateQueries({ queryKey: ["dashboard", "stats"] });
    },
  });

  const transitionMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: TransitionProjectRequest }) =>
      projectService.transition(id, payload),
    onSuccess: (res) => {
      const project = res.data.data;
      void qc.invalidateQueries({ queryKey: ["projects"] });
      void qc.invalidateQueries({ queryKey: ["projects", project.id] });
      void qc.invalidateQueries({ queryKey: ["dashboard", "stats"] });
    },
  });

  const isSaving = createMutation.isPending || updateMutation.isPending;

  function startCreate() {
    setFormMode("create");
    setFormError(null);
    setShowForm(true);
  }

  function startEdit(project: Project) {
    setSelectedId(project.id);
    setFormMode("edit");
    setFormError(null);
    setShowForm(true);
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);

    const parsed = projectFormSchema.safeParse(formValues(event.currentTarget));
    if (!parsed.success) {
      setFormError(parsed.error.issues[0]?.message ?? "Data proyek belum valid.");
      return;
    }

    const values = parsed.data;
    if (formMode === "create") {
      createMutation.mutate(toCreatePayload(values));
      return;
    }

    if (!selectedProject) {
      setFormError("Pilih proyek sebelum menyimpan perubahan.");
      return;
    }
    updateMutation.mutate({
      id: selectedProject.id,
      payload: toUpdatePayload(values),
    });
  }

  return (
    <DashboardLayout title="Proyek">
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-lg font-semibold text-foreground">Ruang Kerja Proyek</h2>
          <p className="text-sm text-muted-foreground">
            {projects.length} proyek dalam cakupan
          </p>
        </div>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              type="search"
              placeholder="Cari proyek"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              className={cn(
                "h-9 w-full rounded-md border border-input bg-background pl-9 pr-3 text-sm sm:w-56",
                "placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
              )}
            />
          </div>
          <button
            type="button"
            onClick={startCreate}
            className={cn(
              "inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground",
              "transition-colors hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-ring"
            )}
          >
            <Plus className="h-4 w-4" aria-hidden="true" />
            Proyek Baru
          </button>
        </div>
      </div>

      {showForm && (
        <ProjectForm
          key={formMode === "edit" ? (selectedProject?.id ?? "edit") : "create"}
          mode={formMode}
          project={formMode === "edit" ? selectedProject : null}
          error={formError}
          isSaving={isSaving}
          orgUnits={orgUnits}
          programs={programs}
          sectors={sectors}
          regions={regions}
          riverBasins={riverBasins}
          onSubmit={handleSubmit}
          onCancel={() => {
            setShowForm(false);
            setFormError(null);
          }}
        />
      )}

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.25fr)_minmax(360px,0.75fr)]">
        <section className="rounded-lg border border-border bg-card shadow-sm">
          <div className="border-b border-border px-4 py-3">
            <h3 className="text-sm font-semibold text-foreground">Daftar Proyek</h3>
          </div>
          {projectListQuery.isLoading ? (
            <div className="p-6 text-sm text-muted-foreground">Memuat daftar proyek...</div>
          ) : projectListQuery.isError ? (
            <ErrorState message="Daftar proyek belum dapat dimuat." />
          ) : projects.length === 0 ? (
            <EmptyState
              search={search}
              onCreate={startCreate}
            />
          ) : (
            <div className="overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border bg-muted/40">
                    <th className="px-4 py-3 text-left font-medium text-muted-foreground">
                      Proyek
                    </th>
                    <th className="hidden px-4 py-3 text-left font-medium text-muted-foreground sm:table-cell">
                      Status
                    </th>
                    <th className="hidden px-4 py-3 text-left font-medium text-muted-foreground md:table-cell">
                      Progres
                    </th>
                    <th className="hidden px-4 py-3 text-left font-medium text-muted-foreground lg:table-cell">
                      Anggaran
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {projects.map((project) => (
                    <tr
                      key={project.id}
                      onClick={() => setSelectedId(project.id)}
                      className={cn(
                        "cursor-pointer transition-colors hover:bg-muted/30",
                        selectedId === project.id && "bg-muted/40"
                      )}
                    >
                      <td className="px-4 py-3">
                        <div>
                          <Link
                            href={`/projects/${project.id}`}
                            className="font-medium text-foreground hover:text-primary hover:underline"
                            onClick={(event) => event.stopPropagation()}
                          >
                            {project.name}
                          </Link>
                          <p className="text-xs text-muted-foreground">{project.code}</p>
                        </div>
                      </td>
                      <td className="hidden px-4 py-3 sm:table-cell">
                        <StatusBadge status={project.status} />
                      </td>
                      <td className="hidden px-4 py-3 md:table-cell">
                        <ProgressCell value={project.progress_pct} />
                      </td>
                      <td className="hidden px-4 py-3 text-muted-foreground lg:table-cell">
                        {formatIDR(project.budget_total)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>

        <ProjectDetail
          project={selectedProject}
          isLoading={detailQuery.isLoading}
          history={historyQuery.data ?? []}
          historyLoading={historyQuery.isLoading}
          transitionPending={transitionMutation.isPending}
          onEdit={startEdit}
          onTransition={(project, status) =>
            transitionMutation.mutate({
              id: project.id,
              payload: { to_status: status, comment: `Status diubah ke ${status}` },
            })
          }
        />
      </div>
    </DashboardLayout>
  );
}

function ProjectForm({
  mode,
  project,
  error,
  isSaving,
  orgUnits,
  programs,
  sectors,
  regions,
  riverBasins,
  onSubmit,
  onCancel,
}: {
  mode: FormMode;
  project: Project | null;
  error: string | null;
  isSaving: boolean;
  orgUnits: OrgUnit[];
  programs: Program[];
  sectors: Sector[];
  regions: Region[];
  riverBasins: RiverBasin[];
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onCancel: () => void;
}) {
  return (
    <form
      onSubmit={onSubmit}
      className="mb-6 rounded-lg border border-border bg-card p-5 shadow-sm"
    >
      <div className="mb-4 flex items-center justify-between">
        <div>
          <h3 className="text-sm font-semibold text-foreground">
            {mode === "create" ? "Tambah Proyek" : "Edit Proyek"}
          </h3>
          <p className="text-xs text-muted-foreground">
            Lengkapi data utama proyek, lalu simpan perubahan.
          </p>
        </div>
        <button
          type="button"
          onClick={onCancel}
          className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          aria-label="Tutup form proyek"
        >
          <X className="h-4 w-4" aria-hidden="true" />
        </button>
      </div>

      {error && (
        <div className="mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-4">
        <Field label="Kode">
          <input
            name="code"
            defaultValue={project?.code ?? ""}
            readOnly={mode === "edit"}
            className={inputClassName + (mode === "edit" ? " cursor-not-allowed opacity-60" : "")}
            placeholder="PMO-001"
          />
        </Field>
        <Field label="Nama">
          <input
            name="name"
            defaultValue={project?.name ?? ""}
            className={inputClassName}
            placeholder="Nama proyek"
          />
        </Field>
        <Field label="Prioritas">
          <select name="priority" defaultValue={project?.priority ?? "MEDIUM"} className={inputClassName}>
            {PRIORITIES.map((priority) => (
              <option key={priority} value={priority}>
                {PRIORITY_LABELS[priority]}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Anggaran">
          <input
            name="budget_total"
            type="number"
            min="0"
            defaultValue={project?.budget_total ?? 0}
            className={inputClassName}
          />
        </Field>
        <Field label="Kategori">
          <input
            name="category"
            defaultValue={project?.category ?? ""}
            className={inputClassName}
            placeholder="Infrastruktur"
          />
        </Field>
        <Field label="Tanggal Mulai">
          <input
            name="start_date"
            type="date"
            defaultValue={toDateInput(project?.start_date)}
            className={inputClassName}
          />
        </Field>
        <Field label="Tanggal Selesai">
          <input
            name="end_date"
            type="date"
            defaultValue={toDateInput(project?.end_date)}
            className={inputClassName}
          />
        </Field>
        <Field label="Progres">
          <input
            name="progress_pct"
            type="number"
            min="0"
            max="100"
            defaultValue={project?.progress_pct ?? 0}
            className={inputClassName}
          />
        </Field>
        <Field label="Balai / Unit Pemilik">
          <select
            name="org_unit_id"
            defaultValue={project?.org_unit_id ?? ""}
            className={inputClassName}
          >
            <option value="">— Pilih org unit —</option>
            {orgUnits.map((u) => (
              <option key={u.id} value={u.id}>
                {u.name} ({OrgUnitLevelLabel[u.level]})
              </option>
            ))}
          </select>
        </Field>
        <Field label="Program">
          <select
            name="program_id"
            defaultValue={project?.program_id ?? ""}
            className={inputClassName}
          >
            <option value="">— Pilih program —</option>
            {programs.map((p) => (
              <option key={p.id} value={p.id}>
                {p.code} – {p.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Sektor SDA">
          <select
            name="sector_id"
            defaultValue={project?.sector_id ?? ""}
            className={inputClassName}
          >
            <option value="">— Pilih sektor —</option>
            {sectors.map((s) => (
              <option key={s.id} value={s.id}>
                {s.code} – {s.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Wilayah">
          <select
            name="region_id"
            defaultValue={project?.region_id ?? ""}
            className={inputClassName}
          >
            <option value="">— Pilih wilayah —</option>
            {regions.map((r) => (
              <option key={r.id} value={r.id}>
                {r.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="DAS">
          <select
            name="river_basin_id"
            defaultValue={project?.river_basin_id ?? ""}
            className={inputClassName}
          >
            <option value="">— Pilih DAS —</option>
            {riverBasins.map((rb) => (
              <option key={rb.id} value={rb.id}>
                {rb.code} – {rb.name}
              </option>
            ))}
          </select>
        </Field>
      </div>

      <div className="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Field label="Deskripsi">
          <textarea
            name="description"
            defaultValue={project?.description ?? ""}
            className={cn(inputClassName, "min-h-20 py-2")}
            placeholder="Deskripsi operasional singkat"
          />
        </Field>
        <Field label="Tujuan">
          <textarea
            name="objectives"
            defaultValue={project?.objectives ?? ""}
            className={cn(inputClassName, "min-h-20 py-2")}
            placeholder="Hasil yang diharapkan"
          />
        </Field>
      </div>

      <input name="currency" type="hidden" value={project?.currency ?? "IDR"} />

      <div className="mt-4 flex justify-end gap-2">
        <button
          type="button"
          onClick={onCancel}
          className="inline-flex h-9 items-center justify-center rounded-md border border-input px-3 text-sm font-medium hover:bg-accent"
        >
          Batal
        </button>
        <button
          type="submit"
          disabled={isSaving}
          className="inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
        >
          <Save className="h-4 w-4" aria-hidden="true" />
          {isSaving ? "Menyimpan..." : "Simpan"}
        </button>
      </div>
    </form>
  );
}

function ProjectDetail({
  project,
  isLoading,
  history,
  historyLoading,
  transitionPending,
  onEdit,
  onTransition,
}: {
  project: Project | null;
  isLoading: boolean;
  history: Array<{
    id: string;
    progress_pct: number;
    notes: string | null;
    recorded_at: string;
  }>;
  historyLoading: boolean;
  transitionPending: boolean;
  onEdit: (project: Project) => void;
  onTransition: (project: Project, status: ProjectStatus) => void;
}) {
  if (isLoading) {
    return (
      <aside className="rounded-lg border border-border bg-card p-5 shadow-sm">
        <div className="h-5 w-32 rounded-md bg-muted" />
        <div className="mt-4 h-24 rounded-md bg-muted" />
      </aside>
    );
  }

  if (!project) {
    return (
      <aside className="rounded-lg border border-border bg-card p-5 shadow-sm">
        <p className="text-sm text-muted-foreground">Pilih proyek untuk melihat detail.</p>
      </aside>
    );
  }

  return (
    <aside className="space-y-6 rounded-lg border border-border bg-card p-5 shadow-sm">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-medium uppercase text-muted-foreground">{project.code}</p>
          <h3 className="mt-1 text-lg font-semibold text-foreground">{project.name}</h3>
          <div className="mt-2 flex flex-wrap gap-2">
            <StatusBadge status={project.status} />
            <span className="inline-flex rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-muted-foreground">
              {PRIORITY_LABELS[project.priority]}
            </span>
          </div>
        </div>
        <button
          type="button"
          onClick={() => onEdit(project)}
          className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          aria-label="Edit proyek"
        >
          <Edit3 className="h-4 w-4" aria-hidden="true" />
        </button>
      </div>

      <div>
        <div className="mb-2 flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Progres</span>
          <span className="font-medium text-foreground">{project.progress_pct}%</span>
        </div>
        <div className="h-2 overflow-hidden rounded-full bg-muted">
          <div
            className="h-full rounded-full bg-primary"
            style={{ width: `${project.progress_pct}%` }}
          />
        </div>
      </div>

      <dl className="grid grid-cols-2 gap-4 text-sm">
        <div>
          <dt className="text-muted-foreground">Anggaran</dt>
          <dd className="font-medium text-foreground">{formatIDR(project.budget_total)}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Tanggal Selesai</dt>
          <dd className="font-medium text-foreground">{formatDate(project.end_date)}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Kategori</dt>
          <dd className="font-medium text-foreground">{project.category || "-"}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Diperbarui</dt>
          <dd className="font-medium text-foreground">{formatDate(project.updated_at)}</dd>
        </div>
      </dl>

      {project.description && (
        <p className="text-sm leading-relaxed text-muted-foreground">{project.description}</p>
      )}

      <div>
        <h4 className="mb-2 text-sm font-semibold text-foreground">Transisi Status</h4>
        <div className="flex flex-wrap gap-2">
          {NEXT_STATUS[project.status].length > 0 ? (
            NEXT_STATUS[project.status].map((status) => (
              <button
                key={status}
                type="button"
                disabled={transitionPending}
                onClick={() => onTransition(project, status)}
                className="inline-flex h-8 items-center justify-center rounded-md border border-input px-3 text-xs font-medium hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
              >
                {STATUS_LABELS[status]}
              </button>
            ))
          ) : (
            <p className="text-sm text-muted-foreground">Tidak ada transisi lanjutan.</p>
          )}
        </div>
      </div>

      <div>
        <div className="mb-3 flex items-center gap-2">
          <History className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
          <h4 className="text-sm font-semibold text-foreground">Riwayat Progres</h4>
        </div>
        {historyLoading ? (
          <p className="text-sm text-muted-foreground">Memuat riwayat...</p>
        ) : history.length === 0 ? (
          <div className="flex items-start gap-3 rounded-md bg-muted/40 p-3">
            <CheckCircle2 className="mt-0.5 h-4 w-4 text-green-700" aria-hidden="true" />
            <p className="text-sm text-muted-foreground">Belum ada riwayat progres.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {history.map((item) => (
              <div key={item.id} className="rounded-md border border-border p-3">
                <div className="flex items-center justify-between text-sm">
                  <span className="font-medium text-foreground">{item.progress_pct}%</span>
                  <span className="text-xs text-muted-foreground">
                    {formatDate(item.recorded_at)}
                  </span>
                </div>
                {item.notes && (
                  <p className="mt-1 text-sm text-muted-foreground">{item.notes}</p>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </aside>
  );
}

function EmptyState({ search, onCreate }: { search: string; onCreate: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <div className="mb-4 rounded-full bg-muted p-4">
        <Filter className="h-6 w-6 text-muted-foreground" aria-hidden="true" />
      </div>
      <p className="text-sm font-medium text-foreground">Proyek tidak ditemukan</p>
      <p className="mt-1 text-sm text-muted-foreground">
        {search ? "Coba kata kunci lain." : "Tambahkan proyek pertama untuk memulai."}
      </p>
      {!search && (
        <button
          type="button"
          onClick={onCreate}
          className="mt-4 inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" aria-hidden="true" />
          Proyek Baru
        </button>
      )}
    </div>
  );
}

function ErrorState({ message }: { message: string }) {
  return (
    <div className="flex items-center gap-3 p-6 text-sm text-destructive">
      <AlertCircle className="h-4 w-4" aria-hidden="true" />
      {message}
    </div>
  );
}

function StatusBadge({ status }: { status: ProjectStatus }) {
  return (
    <span
      className={cn(
        "inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium",
        STATUS_COLORS[status]
      )}
    >
      {STATUS_LABELS[status]}
    </span>
  );
}

function ProgressCell({ value }: { value: number }) {
  return (
    <div className="flex items-center gap-2">
      <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
        <div className="h-full rounded-full bg-primary" style={{ width: `${value}%` }} />
      </div>
      <span className="w-10 text-right text-xs text-muted-foreground">{value}%</span>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="mb-1 block text-sm font-medium text-foreground">{label}</span>
      {children}
    </label>
  );
}

function formValues(form: HTMLFormElement): ProjectFormValues {
  const data = new FormData(form);
  return {
    code: getString(data, "code"),
    name: getString(data, "name"),
    description: getString(data, "description"),
    objectives: getString(data, "objectives"),
    priority: getString(data, "priority") as Priority,
    category: getString(data, "category"),
    start_date: getString(data, "start_date"),
    end_date: getString(data, "end_date"),
    budget_total: Number(getString(data, "budget_total") || 0),
    currency: getString(data, "currency") || "IDR",
    progress_pct: Number(getString(data, "progress_pct") || 0),
    org_unit_id: getString(data, "org_unit_id"),
    program_id: getString(data, "program_id"),
    sector_id: getString(data, "sector_id"),
    region_id: getString(data, "region_id"),
    river_basin_id: getString(data, "river_basin_id"),
  };
}

function getString(data: FormData, key: string) {
  const value = data.get(key);
  return typeof value === "string" ? value : "";
}

function toCreatePayload(values: ProjectFormValues): CreateProjectRequest {
  return {
    code: values.code,
    name: values.name,
    description: optionalString(values.description),
    objectives: optionalString(values.objectives),
    priority: values.priority,
    category: optionalString(values.category),
    start_date: optionalString(values.start_date),
    end_date: optionalString(values.end_date),
    budget_total: values.budget_total,
    currency: values.currency,
    org_unit_id: optionalString(values.org_unit_id),
    program_id: optionalString(values.program_id),
    sector_id: optionalString(values.sector_id),
    region_id: optionalString(values.region_id),
    river_basin_id: optionalString(values.river_basin_id),
  };
}

function toUpdatePayload(values: ProjectFormValues): UpdateProjectRequest {
  return {
    name: values.name,
    description: optionalString(values.description),
    objectives: optionalString(values.objectives),
    priority: values.priority,
    category: optionalString(values.category),
    start_date: optionalString(values.start_date),
    end_date: optionalString(values.end_date),
    budget_total: values.budget_total,
    currency: values.currency,
    progress_pct: values.progress_pct,
    org_unit_id: optionalString(values.org_unit_id),
    program_id: optionalString(values.program_id),
    sector_id: optionalString(values.sector_id),
    region_id: optionalString(values.region_id),
    river_basin_id: optionalString(values.river_basin_id),
  };
}

function optionalString(value: string | undefined) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

function toDateInput(value: string | null | undefined) {
  if (!value) return "";
  return value.slice(0, 10);
}

const inputClassName = cn(
  "h-9 w-full rounded-md border border-input bg-background px-3 text-sm",
  "placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring",
  "disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground"
);
