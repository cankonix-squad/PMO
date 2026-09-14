"use client";

import type { ProjectDocument } from "@/types/project";
import type { FormEvent } from "react";
import { DOCUMENT_CATEGORIES, inputClassName } from "../schemas";
import { Field } from "./Field";
import { FormActions } from "./FormActions";
import { FormHeader } from "./FormHeader";

export function DocumentForm({
  document,
  isSaving,
  onSubmit,
  onCancel,
}: {
  document: ProjectDocument | null;
  isSaving: boolean;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onCancel: () => void;
}) {
  return (
    <form onSubmit={onSubmit} className="mb-6 rounded-lg border border-border bg-card p-5 shadow-sm">
      <FormHeader
        title={document ? "Edit Metadata Dokumen" : "Upload Dokumen"}
        onCancel={onCancel}
      />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <Field label="Document Name">
          <input
            name="name"
            defaultValue={document?.name ?? ""}
            placeholder="Kosongkan untuk memakai nama file"
            className={inputClassName}
          />
        </Field>
        <Field label="Category">
          <select
            name="category"
            defaultValue={document?.category ?? "OTHER"}
            className={inputClassName}
          >
            {DOCUMENT_CATEGORIES.map((category) => (
              <option key={category} value={category}>
                {category.replace("_", " ")}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Version">
          <input
            name="version"
            defaultValue={document?.version ?? ""}
            placeholder="e.g. v1.0"
            className={inputClassName}
          />
        </Field>
        {!document && (
          <Field label="File">
            <input name="file" type="file" className={inputClassName} />
          </Field>
        )}
      </div>
      {!document && (
        <p className="mt-3 text-xs text-muted-foreground">
          Allowed: PDF, DOC(X), XLS(X), PPT(X), PNG, JPG, TXT, CSV (max 20 MB).
        </p>
      )}
      <FormActions isSaving={isSaving} onCancel={onCancel} />
    </form>
  );
}
