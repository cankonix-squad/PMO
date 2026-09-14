"use client";

import { orgUnitService } from "@/services/org-unit.service";
import { programService } from "@/services/portfolio.service";
import { projectCategoryService } from "@/services/project-category.service";
import { projectService } from "@/services/project.service";
import { regionService, riverBasinService, sectorService } from "@/services/spatial.service";
import type { CreateProjectRequest, Project, TransitionProjectRequest, UpdateProjectRequest } from "@/types/project";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import type { FormMode, ProjectFormValues } from "../schemas";
import { toCreatePayload, toUpdatePayload } from "../utils";

export function useProjects() {

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
  return {
    projects,
    search,
    setSearch,
    startCreate,
    showForm,
    formMode,
    selectedProject,
    formError,
    isSaving,
    orgUnits,
    programs,
    sectors,
    regions,
    riverBasins,
    projectCategories,
    descAttachments,
    objAttachments,
    setDescAttachments,
    setObjAttachments,
    handleSubmit,
    setShowForm,
    setFormError,
    projectListQuery,
    setSelectedId,
    selectedId,
    detailQuery,
    historyQuery,
    transitionMutation,
    startEdit,
  };
}
