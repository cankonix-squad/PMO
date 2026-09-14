"use client";
import type { QueryClient } from "@tanstack/react-query";
import type { Dispatch, SetStateAction } from "react";
import { projectService } from "@/services/project.service";
import type { CorrectiveAction, CreateCorrectiveActionRequest, UpdateCorrectiveActionRequest } from "@/types/project";
import { useMutation } from "@tanstack/react-query";

interface Options {
  projectId: string;
  qc: QueryClient;
  setCAFormOpen: Dispatch<SetStateAction<boolean>>;
  setEditingCA: Dispatch<SetStateAction<CorrectiveAction | null>>;
}

export function useCAMutations({ projectId, qc, setCAFormOpen, setEditingCA }: Options) {
  const createCAMutation = useMutation({
    mutationFn: (data: CreateCorrectiveActionRequest) =>
      projectService.createCorrectiveAction(projectId, data),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "corrective-actions"] });
      setCAFormOpen(false);
      setEditingCA(null);
    },
  });

  const updateCAMutation = useMutation({
    mutationFn: ({ caId, payload }: { caId: string; payload: UpdateCorrectiveActionRequest }) =>
      projectService.updateCorrectiveAction(projectId, caId, payload),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "corrective-actions"] });
      setCAFormOpen(false);
      setEditingCA(null);
    },
  });

  const deleteCAMutation = useMutation({
    mutationFn: (caId: string) => projectService.deleteCorrectiveAction(projectId, caId),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "corrective-actions"] }),
  });

  const transitionCAMutation = useMutation({
    mutationFn: ({ caId, toStatus }: { caId: string; toStatus: string }) =>
      projectService.transitionCorrectiveAction(projectId, caId, toStatus),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "corrective-actions"] }),
  });
  return { createCAMutation, updateCAMutation, deleteCAMutation, transitionCAMutation };
}
