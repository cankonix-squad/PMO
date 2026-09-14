"use client";
import type { Dispatch, SetStateAction } from "react";
import { projectService } from "@/services/project.service";
import type { CreateMilestoneRequest, Milestone, UpdateMilestoneRequest } from "@/types/project";
import { useMutation } from "@tanstack/react-query";

interface Options {
  projectId: string;
  setMilestoneFormOpen: Dispatch<SetStateAction<boolean>>;
  setEditingMilestone: Dispatch<SetStateAction<Milestone | null>>;
  setFormError: Dispatch<SetStateAction<string | null>>;
  refreshProjectData: () => void;
}

export function useMilestoneMutations({ projectId, setMilestoneFormOpen, setEditingMilestone, setFormError, refreshProjectData }: Options) {
  const createMilestoneMutation = useMutation({
    mutationFn: (payload: CreateMilestoneRequest) =>
      projectService.createMilestone(projectId, payload),
    onSuccess: () => {
      setMilestoneFormOpen(false);
      setEditingMilestone(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateMilestoneMutation = useMutation({
    mutationFn: ({
      milestoneId,
      payload,
    }: {
      milestoneId: string;
      payload: UpdateMilestoneRequest;
    }) => projectService.updateMilestone(projectId, milestoneId, payload),
    onSuccess: () => {
      setMilestoneFormOpen(false);
      setEditingMilestone(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteMilestoneMutation = useMutation({
    mutationFn: (milestoneId: string) =>
      projectService.deleteMilestone(projectId, milestoneId),
    onSuccess: refreshProjectData,
  });
  return { createMilestoneMutation, updateMilestoneMutation, deleteMilestoneMutation };
}
