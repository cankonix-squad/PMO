"use client";

import { cn } from "@/lib/utils";
import type { Risk } from "@/types/project";
import type { FormEvent } from "react";
import { inputClassName } from "../schemas";
import { toDateInput } from "../utils";
import { Field } from "./Field";
import { FormActions } from "./FormActions";
import { FormHeader } from "./FormHeader";

export function RiskForm({
  risk,
  isSaving,
  onSubmit,
  onCancel,
}: {
  risk: Risk | null;
  isSaving: boolean;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onCancel: () => void;
}) {
  return (
    <form onSubmit={onSubmit} className="mb-6 rounded-lg border border-border bg-card p-5 shadow-sm">
      <FormHeader title={risk ? "Edit Risiko" : "Tambah Risiko"} onCancel={onCancel} />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-4">
        <Field label="Title">
          <input name="title" defaultValue={risk?.title ?? ""} className={inputClassName} />
        </Field>
        <Field label="Probability (1-5)">
          <input
            name="probability"
            type="number"
            min="1"
            max="5"
            defaultValue={risk?.probability ?? 3}
            className={inputClassName}
          />
        </Field>
        <Field label="Impact (1-5)">
          <input
            name="impact"
            type="number"
            min="1"
            max="5"
            defaultValue={risk?.impact ?? 3}
            className={inputClassName}
          />
        </Field>
        <Field label="Due Date">
          <input
            name="due_date"
            type="date"
            defaultValue={toDateInput(risk?.due_date)}
            className={inputClassName}
          />
        </Field>
      </div>
      <div className="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Field label="Description">
          <textarea
            name="description"
            defaultValue={risk?.description ?? ""}
            className={cn(inputClassName, "min-h-20 py-2")}
          />
        </Field>
        <Field label="Mitigation">
          <textarea
            name="mitigation"
            defaultValue={risk?.mitigation ?? ""}
            className={cn(inputClassName, "min-h-20 py-2")}
          />
        </Field>
      </div>
      <FormActions isSaving={isSaving} onCancel={onCancel} />
    </form>
  );
}
