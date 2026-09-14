"use client";
import type { QueryClient } from "@tanstack/react-query";
import type { Dispatch, SetStateAction } from "react";
import { baselineService } from "@/services/monitoring.service";
import type { Baseline, CreateBaselineRequest, UpdateBaselineRequest } from "@/types/monitoring";
import { useMutation } from "@tanstack/react-query";

interface Options {
  projectId: string;
  setBaselineFormOpen: Dispatch<SetStateAction<boolean>>;
  setEditingBaseline: Dispatch<SetStateAction<Baseline | null>>;
  setFormError: Dispatch<SetStateAction<string | null>>;
  qc: QueryClient;
}

export function useBaselineMutations({ projectId, setBaselineFormOpen, setEditingBaseline, setFormError, qc }: Options) {
  const createBaselineMutation = useMutation({
    mutationFn: (payload: CreateBaselineRequest) => baselineService.create(projectId, payload),
    onSuccess: () => {
      setBaselineFormOpen(false);
      setEditingBaseline(null);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "baselines"] });
    },
    onError: () => setFormError("Gagal menyimpan baseline."),
  });

  const updateBaselineMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateBaselineRequest }) =>
      baselineService.update(projectId, id, payload),
    onSuccess: () => {
      setBaselineFormOpen(false);
      setEditingBaseline(null);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "baselines"] });
    },
    onError: () => setFormError("Gagal memperbarui baseline."),
  });

  const deleteBaselineMutation = useMutation({
    mutationFn: (id: string) => baselineService.delete(projectId, id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["projects", projectId, "baselines"] }),
  });
  return { createBaselineMutation, updateBaselineMutation, deleteBaselineMutation };
}
