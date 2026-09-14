"use client";
import type { Dispatch, SetStateAction } from "react";
import { projectService } from "@/services/project.service";
import type { ProjectDocument, UpdateDocumentRequest } from "@/types/project";
import { useMutation } from "@tanstack/react-query";

interface Options {
  projectId: string;
  setDocumentFormOpen: Dispatch<SetStateAction<boolean>>;
  setEditingDocument: Dispatch<SetStateAction<ProjectDocument | null>>;
  setFormError: Dispatch<SetStateAction<string | null>>;
  refreshProjectData: () => void;
}

export function useDocumentMutations({ projectId, setDocumentFormOpen, setEditingDocument, setFormError, refreshProjectData }: Options) {
  const uploadDocumentMutation = useMutation({
    mutationFn: (payload: { file: File; name?: string; category?: string; version?: string }) =>
      projectService.uploadDocument(projectId, payload),
    onSuccess: () => {
      setDocumentFormOpen(false);
      setEditingDocument(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateDocumentMutation = useMutation({
    mutationFn: ({
      documentId,
      payload,
    }: {
      documentId: string;
      payload: UpdateDocumentRequest;
    }) => projectService.updateDocument(projectId, documentId, payload),
    onSuccess: () => {
      setDocumentFormOpen(false);
      setEditingDocument(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteDocumentMutation = useMutation({
    mutationFn: (documentId: string) =>
      projectService.deleteDocument(projectId, documentId),
    onSuccess: refreshProjectData,
  });
  return { uploadDocumentMutation, updateDocumentMutation, deleteDocumentMutation };
}
