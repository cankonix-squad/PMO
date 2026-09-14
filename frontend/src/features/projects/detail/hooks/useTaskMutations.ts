"use client";
import type { Dispatch, SetStateAction } from "react";
import { projectService } from "@/services/project.service";
import type { CreateTaskRequest, Task, UpdateTaskRequest } from "@/types/project";
import { useMutation } from "@tanstack/react-query";

interface Options {
  projectId: string;
  setTaskFormOpen: Dispatch<SetStateAction<boolean>>;
  setEditingTask: Dispatch<SetStateAction<Task | null>>;
  setFormError: Dispatch<SetStateAction<string | null>>;
  refreshProjectData: () => void;
}

export function useTaskMutations({ projectId, setTaskFormOpen, setEditingTask, setFormError, refreshProjectData }: Options) {
  const createTaskMutation = useMutation({
    mutationFn: (payload: CreateTaskRequest) => projectService.createTask(projectId, payload),
    onSuccess: () => {
      setTaskFormOpen(false);
      setEditingTask(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateTaskMutation = useMutation({
    mutationFn: ({ taskId, payload }: { taskId: string; payload: UpdateTaskRequest }) =>
      projectService.updateTask(projectId, taskId, payload),
    onSuccess: () => {
      setTaskFormOpen(false);
      setEditingTask(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteTaskMutation = useMutation({
    mutationFn: (taskId: string) => projectService.deleteTask(projectId, taskId),
    onSuccess: refreshProjectData,
  });
  return { createTaskMutation, updateTaskMutation, deleteTaskMutation };
}
