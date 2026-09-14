"use client";

import { cn } from "@/lib/utils";
import type { VendorType } from "@/types/project";
import type { FormEvent } from "react";
import { inputClassName } from "../schemas";
import { Field } from "./Field";
import { FormActions } from "./FormActions";
import { FormHeader } from "./FormHeader";

export function VendorForm({
  type,
  isSaving,
  onSubmit,
  onCancel,
}: {
  type: VendorType;
  isSaving: boolean;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onCancel: () => void;
}) {
  return (
    <form onSubmit={onSubmit} className="mb-6 rounded-lg border border-border bg-card p-5 shadow-sm">
      <FormHeader
        title={type === "CONSULTANT" ? "Tambah Konsultan" : "Tambah Vendor"}
        onCancel={onCancel}
      />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Field label="Name">
          <input name="name" className={inputClassName} />
        </Field>
        <Field label="Legal Name">
          <input name="legal_name" className={inputClassName} />
        </Field>
        <Field label="Tax ID (NPWP)">
          <input name="tax_id" className={inputClassName} />
        </Field>
        <Field label="Contact Person">
          <input name="contact_person" className={inputClassName} />
        </Field>
        <Field label="Email">
          <input name="email" type="email" className={inputClassName} />
        </Field>
        <Field label="Phone">
          <input name="phone" className={inputClassName} />
        </Field>
      </div>
      <div className="mt-4">
        <Field label="Address">
          <textarea name="address" className={cn(inputClassName, "min-h-16 py-2")} />
        </Field>
      </div>
      <FormActions isSaving={isSaving} onCancel={onCancel} />
    </form>
  );
}
