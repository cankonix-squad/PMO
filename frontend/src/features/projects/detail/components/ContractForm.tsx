"use client";

import { useQuery } from "@tanstack/react-query";
import { cn } from "@/lib/utils";
import { projectService } from "@/services/project.service";
import { AGREEMENT_FORM_LABEL, type Contract, type Vendor, type VendorType } from "@/types/project";
import type { FormEvent } from "react";
import { CONTRACT_STATUSES, inputClassName } from "../schemas";
import { toDateInput } from "../utils";
import { Field } from "./Field";
import { FormActions } from "./FormActions";
import { FormHeader } from "./FormHeader";

export function ContractForm({
  contract,
  vendorOptions,
  consultantOptions,
  vendorsLoading,
  isSaving,
  onCreateVendor,
  onSubmit,
  onCancel,
}: {
  contract: Contract | null;
  vendorOptions: Vendor[];
  consultantOptions: Vendor[];
  vendorsLoading: boolean;
  isSaving: boolean;
  onCreateVendor: (type: VendorType) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onCancel: () => void;
}) {
  const { data: ppkOptions = [], isLoading: ppkLoading } = useQuery({
    queryKey: ["contract-ppk-candidates"],
    queryFn: () => projectService.listPPKCandidates().then((r) => r.data.data ?? []),
  });
  return (
    <form onSubmit={onSubmit} className="mb-6 rounded-lg border border-border bg-card p-5 shadow-sm">
      <FormHeader title={contract ? "Edit Kontrak" : "Tambah Kontrak"} onCancel={onCancel} />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <Field label="Contract Number">
          <input
            name="contract_number"
            defaultValue={contract?.contract_number ?? ""}
            className={inputClassName}
          />
        </Field>
        <Field label="Title">
          <input
            name="title"
            defaultValue={contract?.title ?? ""}
            className={inputClassName}
          />
        </Field>
        <Field label="Status">
          <select name="status" defaultValue={contract?.status ?? "DRAFT"} className={inputClassName}>
            {CONTRACT_STATUSES.map((status) => (
              <option key={status} value={status}>
                {status.replace("_", " ")}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Vendor">
          {vendorsLoading ? (
            <div className="flex h-9 items-center text-sm text-muted-foreground">
              Memuat vendor...
            </div>
          ) : vendorOptions.length === 0 ? (
            <div className="flex h-9 items-center gap-2 text-sm text-muted-foreground">
              <span>Belum ada vendor.</span>
              <button
                type="button"
                onClick={() => onCreateVendor("VENDOR")}
                className="text-primary hover:underline"
              >
                Tambah vendor
              </button>
            </div>
          ) : (
            <select
              name="vendor_id"
              defaultValue={contract?.vendor_id ?? ""}
              className={inputClassName}
            >
              <option value="">Select vendor...</option>
              {vendorOptions.map((v) => (
                <option key={v.id} value={v.id}>
                  {v.name}
                </option>
              ))}
            </select>
          )}
        </Field>
        <Field label="Consultant (optional)">
          {vendorsLoading ? (
            <div className="flex h-9 items-center text-sm text-muted-foreground">
              Memuat konsultan...
            </div>
          ) : consultantOptions.length === 0 ? (
            <div className="flex h-9 items-center gap-2 text-sm text-muted-foreground">
              <span>Belum ada konsultan.</span>
              <button
                type="button"
                onClick={() => onCreateVendor("CONSULTANT")}
                className="text-primary hover:underline"
              >
                Tambah konsultan
              </button>
            </div>
          ) : (
            <select
              name="consultant_id"
              defaultValue={contract?.consultant_id ?? ""}
              className={inputClassName}
            >
              <option value="">None</option>
              {consultantOptions.map((v) => (
                <option key={v.id} value={v.id}>
                  {v.name}
                </option>
              ))}
            </select>
          )}
        </Field>
        <Field label="Contract Value">
          <input
            name="contract_value"
            type="number"
            min="0"
            step="0.01"
            defaultValue={contract?.contract_value ?? 0}
            className={inputClassName}
          />
        </Field>
        <Field label="Currency">
          <input
            name="currency"
            defaultValue={contract?.currency ?? "IDR"}
            className={inputClassName}
          />
        </Field>
        <Field label="Signed Date">
          <input
            name="signed_date"
            type="date"
            defaultValue={toDateInput(contract?.signed_date)}
            className={inputClassName}
          />
        </Field>
        <Field label="Start Date">
          <input
            name="start_date"
            type="date"
            defaultValue={toDateInput(contract?.start_date)}
            className={inputClassName}
          />
        </Field>
        <Field label="End Date">
          <input
            name="end_date"
            type="date"
            defaultValue={toDateInput(contract?.end_date)}
            className={inputClassName}
          />
        </Field>
        <Field label="PPK (Pejabat Pembuat Komitmen)">
          <select name="ppk_user_id" defaultValue={contract?.ppk_user_id ?? ""} className={inputClassName} disabled={ppkLoading}>
            <option value="">{ppkLoading ? "Memuat pengguna..." : "— Belum ditetapkan —"}</option>
            {ppkOptions.map((u) => (
              <option key={u.id} value={u.id}>
                {`${u.first_name} ${u.last_name}`.trim() || u.email} · {u.email}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Bentuk Perjanjian">
          <select name="agreement_form" defaultValue={contract?.agreement_form ?? ""} className={inputClassName}>
            <option value="">— Belum ditetapkan —</option>
            {Object.entries(AGREEMENT_FORM_LABEL).map(([k, l]) => (
              <option key={k} value={k}>{l}</option>
            ))}
          </select>
        </Field>
        <Field label="Kode RUP (SIRUP)">
          <input name="sirup_code" defaultValue={contract?.sirup_code ?? ""} maxLength={50} className={inputClassName} />
        </Field>
        <Field label="Kode Tender (SPSE)">
          <input name="spse_tender_code" defaultValue={contract?.spse_tender_code ?? ""} maxLength={50} className={inputClassName} />
        </Field>
      </div>
      <p className="mt-2 text-xs text-muted-foreground">
        Bila PPK ditetapkan, hanya PPK tersebut (atau PMO/Admin) yang dapat memverifikasi laporan dan dokumen penyedia untuk kontrak ini (Perpres 16/2018 Ps.11).
      </p>
      <div className="mt-4">
        <Field label="Scope of Work">
          <textarea
            name="scope_of_work"
            defaultValue={contract?.scope_of_work ?? ""}
            className={cn(inputClassName, "min-h-20 py-2")}
          />
        </Field>
      </div>
      <FormActions isSaving={isSaving} onCancel={onCancel} />
    </form>
  );
}
