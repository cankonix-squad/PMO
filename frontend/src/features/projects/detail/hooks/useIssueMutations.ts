"use client";
import type { Dispatch, SetStateAction } from "react";
import { projectService } from "@/services/project.service";
import type { CreateIssueRequest, Issue, UpdateIssueRequest } from "@/types/project";
import { useMutation } from "@tanstack/react-query";

interface Options {
  projectId: string;
  setIssueFormOpen: Dispatch<SetStateAction<boolean>>;
  setEditingIssue: Dispatch<SetStateAction<Issue | null>>;
  setFormError: Dispatch<SetStateAction<string | null>>;
  refreshProjectData: () => void;
}

export function useIssueMutations({ projectId, setIssueFormOpen, setEditingIssue, setFormError, refreshProjectData }: Options) {
  const createIssueMutation = useMutation({
    mutationFn: (payload: CreateIssueRequest) =>
      projectService.createIssue(projectId, payload),
    onSuccess: () => {
      setIssueFormOpen(false);
      setEditingIssue(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateIssueMutation = useMutation({
    mutationFn: ({
      issueId,
      payload,
    }: {
      issueId: string;
      payload: UpdateIssueRequest;
    }) => projectService.updateIssue(projectId, issueId, payload),
    onSuccess: () => {
      setIssueFormOpen(false);
      setEditingIssue(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteIssueMutation = useMutation({
    mutationFn: (issueId: string) =>
      projectService.deleteIssue(projectId, issueId),
    onSuccess: refreshProjectData,
  });
  return { createIssueMutation, updateIssueMutation, deleteIssueMutation };
}
