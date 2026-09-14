"use client";

import { cn } from "@/lib/utils";
import type { BudgetLine } from "@/types/project";
import type { FormEvent } from "react";
import { inputClassName } from "../schemas";
import { Field } from "./Field";
import { FormActions } from "./FormActions";
import { FormHeader } from "./FormHeader";

export function BudgetForm({
  budget,
  isSaving,
  onSubmit,
  onCancel,
}: {
  budget: BudgetLine | null;
  isSaving: boolean;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onCancel: () => void;
}) {
  return (
    <form onSubmit={onSubmit} className="mb-6 rounded-lg border border-border bg-card p-5 shadow-sm">
      <FormHeader title={budget ? "Edit Baris Anggaran" : "Tambah Baris Anggaran"} onCancel={onCancel} />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-4">
        <Field label="Category">
          <input name="category" defaultValue={budget?.category ?? ""} className={inputClassName} />
        </Field>
        <Field label="Planned">
          <input
            name="planned"
            type="number"
            min="0"
            step="0.01"
            defaultValue={budget?.planned ?? 0}
            className={inputClassName}
          />
        </Field>
        <Field label="Actual">
          <input
            name="actual"
            type="number"
            min="0"
            step="0.01"
            defaultValue={budget?.actual ?? 0}
            className={inputClassName}
          />
        </Field>
        <Field label="Currency">
          <input
            name="currency"
            defaultValue={budget?.currency ?? "IDR"}
            className={inputClassName}
          />
        </Field>
      </div>
      <div className="mt-4">
        <Field label="Description">
          <textarea
            name="description"
            defaultValue={budget?.description ?? ""}
            className={cn(inputClassName, "min-h-20 py-2")}
          />
        </Field>
      </div>
      <FormActions isSaving={isSaving} onCancel={onCancel} />
    </form>
  );
}
