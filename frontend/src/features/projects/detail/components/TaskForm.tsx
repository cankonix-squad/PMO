"use client";

import { cn } from "@/lib/utils";
import type { Task } from "@/types/project";
import type { FormEvent } from "react";
import { inputClassName } from "../schemas";
import { toDateInput } from "../utils";
import { Field } from "./Field";
import { FormActions } from "./FormActions";
import { FormHeader } from "./FormHeader";

export function TaskForm({
  task,
  isSaving,
  onSubmit,
  onCancel,
}: {
  task: Task | null;
  isSaving: boolean;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onCancel: () => void;
}) {
  return (
    <form onSubmit={onSubmit} className="mb-6 rounded-lg border border-border bg-card p-5 shadow-sm">
      <FormHeader title={task ? "Edit Task" : "Tambah Task"} onCancel={onCancel} />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-4">
        <Field label="Title">
          <input name="title" defaultValue={task?.title ?? ""} className={inputClassName} />
        </Field>
        <Field label="Priority">
          <select name="priority" defaultValue={task?.priority ?? "MEDIUM"} className={inputClassName}>
            {["LOW", "MEDIUM", "HIGH", "CRITICAL"].map((priority) => (
              <option key={priority} value={priority}>
                {priority}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Type">
          <select name="type" defaultValue={task?.type ?? "TASK"} className={inputClassName}>
            {["TASK", "BUG", "FEATURE", "RESEARCH"].map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Progress">
          <input
            name="progress_pct"
            type="number"
            min="0"
            max="100"
            defaultValue={task?.progress_pct ?? 0}
            className={inputClassName}
          />
        </Field>
        <Field label="Start Date">
          <input
            name="start_date"
            type="date"
            defaultValue={toDateInput(task?.start_date)}
            className={inputClassName}
          />
        </Field>
        <Field label="Due Date">
          <input
            name="due_date"
            type="date"
            defaultValue={toDateInput(task?.due_date)}
            className={inputClassName}
          />
        </Field>
        <Field label="Estimated Hours">
          <input
            name="est_hours"
            type="number"
            min="0"
            defaultValue={task?.est_hours ?? 0}
            className={inputClassName}
          />
        </Field>
      </div>
      <div className="mt-4">
        <Field label="Description">
          <textarea
            name="description"
            defaultValue={task?.description ?? ""}
            className={cn(inputClassName, "min-h-20 py-2")}
          />
        </Field>
      </div>
      <FormActions isSaving={isSaving} onCancel={onCancel} />
    </form>
  );
}
