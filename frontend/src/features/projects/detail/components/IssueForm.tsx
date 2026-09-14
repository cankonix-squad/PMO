"use client";

import { cn } from "@/lib/utils";
import type { Issue } from "@/types/project";
import type { FormEvent } from "react";
import { inputClassName, ISSUE_ESCALATIONS, ISSUE_SEVERITIES } from "../schemas";
import { toDateInput } from "../utils";
import { Field } from "./Field";
import { FormActions } from "./FormActions";
import { FormHeader } from "./FormHeader";

export function IssueForm({
  issue,
  isSaving,
  onSubmit,
  onCancel,
}: {
  issue: Issue | null;
  isSaving: boolean;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onCancel: () => void;
}) {
  return (
    <form onSubmit={onSubmit} className="mb-6 rounded-lg border border-border bg-card p-5 shadow-sm">
      <FormHeader title={issue ? "Edit Isu" : "Tambah Isu"} onCancel={onCancel} />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-4">
        <Field label="Title">
          <input name="title" defaultValue={issue?.title ?? ""} className={inputClassName} />
        </Field>
        <Field label="Severity">
          <select name="severity" defaultValue={issue?.severity ?? "MEDIUM"} className={inputClassName}>
            {ISSUE_SEVERITIES.map((severity) => (
              <option key={severity} value={severity}>
                {severity}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Escalation">
          <select name="escalation" defaultValue={issue?.escalation ?? "NONE"} className={inputClassName}>
            {ISSUE_ESCALATIONS.map((escalation) => (
              <option key={escalation} value={escalation}>
                {escalation.replace("_", " ")}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Due Date">
          <input
            name="due_date"
            type="date"
            defaultValue={toDateInput(issue?.due_date)}
            className={inputClassName}
          />
        </Field>
      </div>
      <div className="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Field label="Description">
          <textarea
            name="description"
            defaultValue={issue?.description ?? ""}
            className={cn(inputClassName, "min-h-20 py-2")}
          />
        </Field>
        <Field label="Resolution">
          <textarea
            name="resolution"
            defaultValue={issue?.resolution ?? ""}
            className={cn(inputClassName, "min-h-20 py-2")}
          />
        </Field>
      </div>
      <FormActions isSaving={isSaving} onCancel={onCancel} />
    </form>
  );
}
