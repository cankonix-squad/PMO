"use client";

import Link from "next/link";
import { type FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AlertCircle,
  CheckCircle2,
  Edit3,
  Filter,
  History,
  Paperclip,
  Plus,
  Save,
  Search,
  Trash2,
  X,
} from "lucide-react";
import { z } from "zod";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { RichTextEditor } from "@/components/editor/RichTextEditor";
import { projectService } from "@/services/project.service";
import { orgUnitService } from "@/services/org-unit.service";
import type { OrgUnit } from "@/types/org-unit";
import { OrgUnitLevelLabel } from "@/types/org-unit";
import { programService } from "@/services/portfolio.service";
import { sectorService, regionService, riverBasinService } from "@/services/spatial.service";
import { projectCategoryService } from "@/services/project-category.service";
import type { ProjectCategory } from "@/types/project-category";
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
  description: z.string().trim().min(1, "Deskripsi wajib diisi"),
  objectives: z.string().trim().min(1, "Tujuan wajib diisi"),
  priority: z.enum(["LOW", "MEDIUM", "HIGH", "CRITICAL"]),
  category: z.string().trim().min(1, "Kategori wajib dipilih"),
  start_date: z.string().trim().min(1, "Tanggal mulai wajib diisi"),
  end_date: z.string().trim().min(1, "Tanggal selesai wajib diisi"),
  budget_total: z.coerce.number().min(1, "Anggaran wajib lebih dari 0"),
  currency: z.string().trim().min(1).max(10),
  progress_pct: z.coerce.number().min(0).max(100),
  org_unit_id: z.string().trim().min(1, "Balai / Unit Pemilik wajib dipilih"),
  program_id: z.string().trim().min(1, "Program wajib dipilih"),
  sector_id: z.string().trim().min(1, "Sektor SDA wajib dipilih"),
  region_id: z.string().trim().min(1, "Wilayah wajib dipilih"),
  river_basin_id: z.string().trim().min(1, "DAS wajib dipilih"),
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

  const { data: projectCategories = [] } = useQuery({
    queryKey: ["project-categories", false],
    queryFn: () => projectCategoryService.list(false).then((r) => r.data ?? []),
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
    onSuccess: async (res) => {
      const project = res.data.data;
      if (project?.id) {
        setSelectedId(project.id);
        await uploadAttachments(project.id);
      }
      setShowForm(false);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["projects"] });
      void qc.invalidateQueries({ queryKey: ["dashboard", "stats"] });
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ??
        "Gagal menyimpan proyek. Periksa kembali data yang diisi.";
      setFormError(msg);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateProjectRequest }) =>
      projectService.update(id, payload),
    onSuccess: async (res) => {
      const project = res.data.data;
      if (project?.id) {
        setSelectedId(project.id);
        await uploadAttachments(project.id);
      }
      setShowForm(false);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["projects"] });
      void qc.invalidateQueries({ queryKey: ["projects", project.id] });
      void qc.invalidateQueries({ queryKey: ["dashboard", "stats"] });
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ??
        "Gagal memperbarui proyek. Periksa kembali data yang diisi.";
      setFormError(msg);
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

  // Attachment queues: files waiting to be uploaded after project is saved
  const [descAttachments, setDescAttachments] = useState<File[]>([]);
  const [objAttachments, setObjAttachments] = useState<File[]>([]);

  // Upload all queued attachments to the document store
  async function uploadAttachments(projectId: string) {
    const pairs: Array<{ file: File; category: string }> = [
      ...descAttachments.map((f) => ({ file: f, category: "EVIDENCE" as const })),
      ...objAttachments.map((f) => ({ file: f, category: "OTHER" as const })),
    ];
    await Promise.allSettled(
      pairs.map(({ file, category }) =>
        projectService.uploadDocument(projectId, {
          file,
          name: file.name,
          category,
          version: "1",
        })
      )
    );
    setDescAttachments([]);
    setObjAttachments([]);
  }

  function handleSubmit(values: ProjectFormValues) {
    setFormError(null);
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
          projectCategories={projectCategories}
          descAttachments={descAttachments}
          objAttachments={objAttachments}
          onDescAttach={(file) => setDescAttachments((prev) => [...prev, file])}
          onObjAttach={(file) => setObjAttachments((prev) => [...prev, file])}
          onDescRemove={(idx) => setDescAttachments((prev) => prev.filter((_, i) => i !== idx))}
          onObjRemove={(idx) => setObjAttachments((prev) => prev.filter((_, i) => i !== idx))}
          onSubmit={handleSubmit}
          onCancel={() => {
            setShowForm(false);
            setFormError(null);
            setDescAttachments([]);
            setObjAttachments([]);
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
  projectCategories,
  descAttachments,
  objAttachments,
  onDescAttach,
  onObjAttach,
  onDescRemove,
  onObjRemove,
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
  projectCategories: ProjectCategory[];
  descAttachments: File[];
  objAttachments: File[];
  onDescAttach: (file: File) => void;
  onObjAttach: (file: File) => void;
  onDescRemove: (index: number) => void;
  onObjRemove: (index: number) => void;
  onSubmit: (values: ProjectFormValues) => void;
  onCancel: () => void;
}) {
  // ---- Local controlled state for rich fields + selects ----
  const [description, setDescription] = useState(project?.description ?? "");
  const [objectives, setObjectives] = useState(project?.objectives ?? "");
  const [localError, setLocalError] = useState<string | null>(null);

  // Hidden file input refs for attachment pickers
  const descFileRef = useRef<HTMLInputElement>(null);
  const objFileRef = useRef<HTMLInputElement>(null);

  // Empty master data warnings
  const emptyMasters: string[] = [];
  if (orgUnits.length === 0) emptyMasters.push("Balai / Unit Pemilik");
  if (programs.length === 0) emptyMasters.push("Program");
  if (sectors.length === 0) emptyMasters.push("Sektor SDA");
  if (regions.length === 0) emptyMasters.push("Wilayah");
  if (riverBasins.length === 0) emptyMasters.push("DAS");
  if (projectCategories.length === 0) emptyMasters.push("Kategori");

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLocalError(null);

    const data = new FormData(event.currentTarget);
    const raw = {
      code: (data.get("code") as string) ?? "",
      name: (data.get("name") as string) ?? "",
      description,
      objectives,
      priority: (data.get("priority") as string) ?? "",
      category: (data.get("category") as string) ?? "",
      start_date: (data.get("start_date") as string) ?? "",
      end_date: (data.get("end_date") as string) ?? "",
      budget_total: parseRupiahInput((data.get("budget_total") as string) ?? "0"),
      currency: (data.get("currency") as string) || "IDR",
      progress_pct: Number((data.get("progress_pct") as string) || 0),
      org_unit_id: (data.get("org_unit_id") as string) ?? "",
      program_id: (data.get("program_id") as string) ?? "",
      sector_id: (data.get("sector_id") as string) ?? "",
      region_id: (data.get("region_id") as string) ?? "",
      river_basin_id: (data.get("river_basin_id") as string) ?? "",
    };

    const parsed = projectFormSchema.safeParse(raw);
    if (!parsed.success) {
      setLocalError(parsed.error.issues[0]?.message ?? "Data proyek belum valid.");
      return;
    }
    onSubmit(parsed.data);
  }

  const displayedError = localError ?? error;

  return (
    <form
      onSubmit={handleSubmit}
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

      {emptyMasters.length > 0 && (
        <div className="mb-4 rounded-md border border-amber-300 bg-amber-50 px-3 py-2 text-sm text-amber-800">
          <span className="font-medium">Data master belum tersedia:</span>{" "}
          {emptyMasters.join(", ")}.
          Lengkapi data master di menu Pengaturan sebelum membuat proyek.
        </div>
      )}

      {displayedError && (
        <div className="mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {displayedError}
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-4">
        <Field label="Kode" required>
          <input
            name="code"
            defaultValue={project?.code ?? ""}
            readOnly={mode === "edit"}
            className={inputClassName + (mode === "edit" ? " cursor-not-allowed opacity-60" : "")}
            placeholder="PMO-001"
          />
        </Field>
        <Field label="Nama" required>
          <input
            name="name"
            defaultValue={project?.name ?? ""}
            className={inputClassName}
            placeholder="Nama proyek"
          />
        </Field>
        <Field label="Prioritas" required>
          <select name="priority" defaultValue={project?.priority ?? "MEDIUM"} className={inputClassName}>
            {PRIORITIES.map((priority) => (
              <option key={priority} value={priority}>
                {PRIORITY_LABELS[priority]}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Anggaran" required>
          <input
            name="budget_total"
            inputMode="numeric"
            defaultValue={formatRupiahInput(project?.budget_total ?? 0)}
            className={inputClassName}
            placeholder="Rp. 100,000,000,-"
            onInput={formatRupiahInputEvent}
          />
        </Field>
        <Field label="Kategori" required>
          <select
            name="category"
            defaultValue={project?.category ?? ""}
            className={inputClassName}
          >
            <option value="">— Pilih kategori —</option>
            {projectCategories.map((cat) => (
              <option key={cat.id} value={cat.code}>
                {cat.code} – {cat.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Tanggal Mulai" required>
          <input
            name="start_date"
            type="date"
            defaultValue={toDateInput(project?.start_date)}
            className={inputClassName}
          />
        </Field>
        <Field label="Tanggal Selesai" required>
          <input
            name="end_date"
            type="date"
            defaultValue={toDateInput(project?.end_date)}
            className={inputClassName}
          />
        </Field>
        <Field label="Progres">
          <div className="relative">
            <input
              name="progress_pct"
              type="number"
              min="0"
              max="100"
              step="0.1"
              defaultValue={project?.progress_pct ?? 0}
              className={cn(inputClassName, "pr-9")}
            />
            <span className="pointer-events-none absolute inset-y-0 right-3 flex items-center text-sm font-medium text-muted-foreground">
              %
            </span>
          </div>
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
        <div>
          <Field label="Deskripsi" required>
            <RichTextEditor
              value={description}
              onChange={setDescription}
              placeholder="Deskripsi operasional singkat"
              disabled={isSaving}
              onAttachFile={() => descFileRef.current?.click()}
              onAttachImage={(file) => onDescAttach(file)}
            />
          </Field>
          {/* Hidden file input for description attachments */}
          <input
            ref={descFileRef}
            type="file"
            className="sr-only"
            multiple
            onChange={(e) => {
              Array.from(e.target.files ?? []).forEach(onDescAttach);
              e.target.value = "";
            }}
          />
          {/* Attachment queue for description */}
          {descAttachments.length > 0 && (
            <ul className="mt-2 space-y-1">
              {descAttachments.map((file, idx) => (
                <li
                  key={idx}
                  className="flex items-center justify-between rounded-md border border-border bg-muted/40 px-3 py-1.5 text-xs"
                >
                  <span className="flex items-center gap-1.5 truncate text-foreground">
                    <Paperclip className="h-3 w-3 shrink-0 text-muted-foreground" />
                    <span className="truncate">{file.name}</span>
                    <span className="shrink-0 text-muted-foreground">({formatBytes(file.size)})</span>
                  </span>
                  <button
                    type="button"
                    onClick={() => onDescRemove(idx)}
                    className="ml-2 shrink-0 rounded p-0.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                    aria-label="Hapus lampiran"
                  >
                    <Trash2 className="h-3 w-3" />
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div>
          <Field label="Tujuan" required>
            <RichTextEditor
              value={objectives}
              onChange={setObjectives}
              placeholder="Hasil yang diharapkan"
              disabled={isSaving}
              onAttachFile={() => objFileRef.current?.click()}
              onAttachImage={(file) => onObjAttach(file)}
            />
          </Field>
          {/* Hidden file input for objectives attachments */}
          <input
            ref={objFileRef}
            type="file"
            className="sr-only"
            multiple
            onChange={(e) => {
              Array.from(e.target.files ?? []).forEach(onObjAttach);
              e.target.value = "";
            }}
          />
          {/* Attachment queue for objectives */}
          {objAttachments.length > 0 && (
            <ul className="mt-2 space-y-1">
              {objAttachments.map((file, idx) => (
                <li
                  key={idx}
                  className="flex items-center justify-between rounded-md border border-border bg-muted/40 px-3 py-1.5 text-xs"
                >
                  <span className="flex items-center gap-1.5 truncate text-foreground">
                    <Paperclip className="h-3 w-3 shrink-0 text-muted-foreground" />
                    <span className="truncate">{file.name}</span>
                    <span className="shrink-0 text-muted-foreground">({formatBytes(file.size)})</span>
                  </span>
                  <button
                    type="button"
                    onClick={() => onObjRemove(idx)}
                    className="ml-2 shrink-0 rounded p-0.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                    aria-label="Hapus lampiran"
                  >
                    <Trash2 className="h-3 w-3" />
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
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
        <div
          className="prose prose-sm max-w-none text-muted-foreground [&_ul]:list-disc [&_ul]:pl-4 [&_ol]:list-decimal [&_ol]:pl-4 [&_hr]:my-2"
          dangerouslySetInnerHTML={{ __html: project.description }}
        />
      )}

      {project.objectives && (
        <div>
          <h4 className="mb-1 text-sm font-semibold text-foreground">Tujuan</h4>
          <div
            className="prose prose-sm max-w-none text-muted-foreground [&_ul]:list-disc [&_ul]:pl-4 [&_ol]:list-decimal [&_ol]:pl-4 [&_hr]:my-2"
            dangerouslySetInnerHTML={{ __html: project.objectives }}
          />
        </div>
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

function Field({ label, required, children }: { label: string; required?: boolean; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="mb-1 block text-sm font-medium text-foreground">
        {label}
        {required && <span className="ml-0.5 text-destructive" aria-hidden="true">*</span>}
      </span>
      {children}
    </label>
  );
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
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
    budget_total: parseRupiahInput(getString(data, "budget_total")),
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
    progress_pct: values.progress_pct,
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

function parseRupiahInput(value: string) {
  return Number(value.replace(/[^\d]/g, "") || 0);
}

function formatRupiahInput(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    return "";
  }
  return `Rp. ${Math.round(value).toLocaleString("en-US")},-`;
}

function formatRupiahInputEvent(event: FormEvent<HTMLInputElement>) {
  const input = event.currentTarget;
  input.value = formatRupiahInput(parseRupiahInput(input.value));
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
