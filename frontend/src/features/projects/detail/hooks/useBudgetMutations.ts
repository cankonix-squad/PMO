"use client";
import type { Dispatch, SetStateAction } from "react";
import { projectService } from "@/services/project.service";
import type { BudgetLine, CreateBudgetRequest, UpdateBudgetRequest } from "@/types/project";
import { useMutation } from "@tanstack/react-query";

interface Options {
  projectId: string;
  setBudgetFormOpen: Dispatch<SetStateAction<boolean>>;
  setEditingBudget: Dispatch<SetStateAction<BudgetLine | null>>;
  setFormError: Dispatch<SetStateAction<string | null>>;
  refreshProjectData: () => void;
}

export function useBudgetMutations({ projectId, setBudgetFormOpen, setEditingBudget, setFormError, refreshProjectData }: Options) {
  const createBudgetMutation = useMutation({
    mutationFn: (payload: CreateBudgetRequest) =>
      projectService.createBudget(projectId, payload),
    onSuccess: () => {
      setBudgetFormOpen(false);
      setEditingBudget(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateBudgetMutation = useMutation({
    mutationFn: ({
      budgetId,
      payload,
    }: {
      budgetId: string;
      payload: UpdateBudgetRequest;
    }) => projectService.updateBudget(projectId, budgetId, payload),
    onSuccess: () => {
      setBudgetFormOpen(false);
      setEditingBudget(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteBudgetMutation = useMutation({
    mutationFn: (budgetId: string) =>
      projectService.deleteBudget(projectId, budgetId),
    onSuccess: refreshProjectData,
  });
  return { createBudgetMutation, updateBudgetMutation, deleteBudgetMutation };
}
