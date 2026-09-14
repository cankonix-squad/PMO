"use client";
import type { QueryClient } from "@tanstack/react-query";
import type { Dispatch, SetStateAction } from "react";
import { snapshotService } from "@/services/monitoring.service";
import type { CreateSnapshotRequest, Snapshot, TransitionSnapshotRequest, UpdateSnapshotRequest } from "@/types/monitoring";
import { useMutation } from "@tanstack/react-query";

interface Options {
  projectId: string;
  setSnapshotFormOpen: Dispatch<SetStateAction<boolean>>;
  setEditingSnapshot: Dispatch<SetStateAction<Snapshot | null>>;
  setFormError: Dispatch<SetStateAction<string | null>>;
  qc: QueryClient;
}

export function useSnapshotMutations({ projectId, setSnapshotFormOpen, setEditingSnapshot, setFormError, qc }: Options) {
  const createSnapshotMutation = useMutation({
    mutationFn: (payload: CreateSnapshotRequest) => snapshotService.create(projectId, payload),
    onSuccess: () => {
      setSnapshotFormOpen(false);
      setEditingSnapshot(null);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "snapshots"] });
    },
    onError: () => setFormError("Gagal menyimpan snapshot."),
  });

  const updateSnapshotMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateSnapshotRequest }) =>
      snapshotService.update(projectId, id, payload),
    onSuccess: () => {
      setSnapshotFormOpen(false);
      setEditingSnapshot(null);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "snapshots"] });
    },
    onError: () => setFormError("Gagal memperbarui snapshot."),
  });

  const transitionSnapshotMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: TransitionSnapshotRequest }) =>
      snapshotService.transition(projectId, id, payload),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["projects", projectId, "snapshots"] }),
  });

  const deleteSnapshotMutation = useMutation({
    mutationFn: (id: string) => snapshotService.delete(projectId, id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["projects", projectId, "snapshots"] }),
  });
  return { createSnapshotMutation, updateSnapshotMutation, transitionSnapshotMutation, deleteSnapshotMutation };
}
