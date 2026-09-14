"use client";

import { RichTextEditor } from "@/components/editor/RichTextEditor";
import { cn } from "@/lib/utils";
import type { OrgUnit } from "@/types/org-unit";
import { OrgUnitLevelLabel } from "@/types/org-unit";
import type { Program } from "@/types/portfolio";
import type { Project } from "@/types/project";
import type { ProjectCategory } from "@/types/project-category";
import type { Region, RiverBasin, Sector } from "@/types/spatial";
import { Paperclip, Save, Trash2, X } from "lucide-react";
import type { FormEvent } from "react";
import { useRef, useState } from "react";
import type { FormMode, ProjectFormValues } from "../schemas";
import { inputClassName, PRIORITIES, PRIORITY_LABELS, projectFormSchema } from "../schemas";
import { formatBytes, formatRupiahInput, formatRupiahInputEvent, parseRupiahInput, toDateInput } from "../utils";
import { Field } from "./Field";

export function ProjectForm({
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
