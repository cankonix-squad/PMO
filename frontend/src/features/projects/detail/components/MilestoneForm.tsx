"use client";

import { cn } from "@/lib/utils";
import type { Milestone } from "@/types/project";
import type { FormEvent } from "react";
import { inputClassName } from "../schemas";
import { toDateInput } from "../utils";
import { Field } from "./Field";
import { FormActions } from "./FormActions";
import { FormHeader } from "./FormHeader";

export function MilestoneForm({
  milestone,
  isSaving,
  onSubmit,
  onCancel,
}: {
  milestone: Milestone | null;
  isSaving: boolean;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onCancel: () => void;
}) {
  return (
    <form onSubmit={onSubmit} className="mb-6 rounded-lg border border-border bg-card p-5 shadow-sm">
      <FormHeader title={milestone ? "Edit Milestone" : "Tambah Milestone"} onCancel={onCancel} />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <Field label="Title">
          <input name="title" defaultValue={milestone?.title ?? ""} className={inputClassName} />
        </Field>
        <Field label="Due Date">
          <input
            name="due_date"
            type="date"
            defaultValue={toDateInput(milestone?.due_date)}
            className={inputClassName}
          />
        </Field>
        <Field label="Progress">
          <input
            name="progress_pct"
            type="number"
            min="0"
            max="100"
            defaultValue={milestone?.progress_pct ?? 0}
            className={inputClassName}
          />
        </Field>
      </div>
      <div className="mt-4">
        <Field label="Description">
          <textarea
            name="description"
            defaultValue={milestone?.description ?? ""}
            className={cn(inputClassName, "min-h-20 py-2")}
          />
        </Field>
      </div>
      <FormActions isSaving={isSaving} onCancel={onCancel} />
    </form>
  );
}
