"use client";
import type { Dispatch, SetStateAction } from "react";
import { projectService } from "@/services/project.service";
import type { CreateRiskRequest, Risk, UpdateRiskRequest } from "@/types/project";
import { useMutation } from "@tanstack/react-query";

interface Options {
  projectId: string;
  setRiskFormOpen: Dispatch<SetStateAction<boolean>>;
  setEditingRisk: Dispatch<SetStateAction<Risk | null>>;
  setFormError: Dispatch<SetStateAction<string | null>>;
  refreshProjectData: () => void;
}

export function useRiskMutations({ projectId, setRiskFormOpen, setEditingRisk, setFormError, refreshProjectData }: Options) {
  const createRiskMutation = useMutation({
    mutationFn: (payload: CreateRiskRequest) =>
      projectService.createRisk(projectId, payload),
    onSuccess: () => {
      setRiskFormOpen(false);
      setEditingRisk(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateRiskMutation = useMutation({
    mutationFn: ({
      riskId,
      payload,
    }: {
      riskId: string;
      payload: UpdateRiskRequest;
    }) => projectService.updateRisk(projectId, riskId, payload),
    onSuccess: () => {
      setRiskFormOpen(false);
      setEditingRisk(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteRiskMutation = useMutation({
    mutationFn: (riskId: string) =>
      projectService.deleteRisk(projectId, riskId),
    onSuccess: refreshProjectData,
  });
  return { createRiskMutation, updateRiskMutation, deleteRiskMutation };
}
