"use client";
import type { Dispatch, SetStateAction } from "react";
import { projectService } from "@/services/project.service";
import type { Contract, CreateContractRequest, UpdateContractRequest } from "@/types/project";
import { useMutation } from "@tanstack/react-query";

interface Options {
  projectId: string;
  setContractFormOpen: Dispatch<SetStateAction<boolean>>;
  setEditingContract: Dispatch<SetStateAction<Contract | null>>;
  setFormError: Dispatch<SetStateAction<string | null>>;
  refreshProjectData: () => void;
}

export function useContractMutations({ projectId, setContractFormOpen, setEditingContract, setFormError, refreshProjectData }: Options) {
  const createContractMutation = useMutation({
    mutationFn: (payload: CreateContractRequest) =>
      projectService.createContract(projectId, payload),
    onSuccess: () => {
      setContractFormOpen(false);
      setEditingContract(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateContractMutation = useMutation({
    mutationFn: ({
      contractId,
      payload,
    }: {
      contractId: string;
      payload: UpdateContractRequest;
    }) => projectService.updateContract(projectId, contractId, payload),
    onSuccess: () => {
      setContractFormOpen(false);
      setEditingContract(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteContractMutation = useMutation({
    mutationFn: (contractId: string) =>
      projectService.deleteContract(projectId, contractId),
    onSuccess: refreshProjectData,
  });
  return { createContractMutation, updateContractMutation, deleteContractMutation };
}
