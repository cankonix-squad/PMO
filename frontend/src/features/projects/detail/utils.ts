"use client";

import type {
  BudgetStatus,
  ContractStatus,
  CorrectiveActionStatus,
  CreateBudgetRequest,
  CreateContractRequest,
  CreateIssueRequest,
  CreateMilestoneRequest,
  CreateRiskRequest,
  CreateTaskRequest,
  CreateVendorRequest,
  RiskSeverity,
  UpdateBudgetRequest,
  UpdateContractRequest,
  UpdateDocumentRequest,
  UpdateIssueRequest,
  UpdateMilestoneRequest,
  UpdateRiskRequest,
  UpdateTaskRequest,
  VendorType,
} from "@/types/project";
import type {
  BudgetFormValues,
  ContractFormValues,
  DocumentFormValues,
  IssueFormValues,
  MilestoneFormValues,
  RiskFormValues,
  TaskFormValues,
  VendorFormValues,
} from "./schemas";

export function readTaskForm(form: HTMLFormElement): TaskFormValues {
  const data = new FormData(form);
  return {
    title: getString(data, "title"),
    description: getString(data, "description"),
    priority: getString(data, "priority") as TaskFormValues["priority"],
    type: getString(data, "type") as TaskFormValues["type"],
    start_date: getString(data, "start_date"),
    due_date: getString(data, "due_date"),
    est_hours: Number(getString(data, "est_hours") || 0),
    progress_pct: Number(getString(data, "progress_pct") || 0),
  };
}

export function readMilestoneForm(form: HTMLFormElement): MilestoneFormValues {
  const data = new FormData(form);
  return {
    title: getString(data, "title"),
    description: getString(data, "description"),
    due_date: getString(data, "due_date"),
    progress_pct: Number(getString(data, "progress_pct") || 0),
  };
}

export function toTaskCreatePayload(values: TaskFormValues): CreateTaskRequest {
  return {
    title: values.title,
    description: optionalString(values.description),
    priority: values.priority,
    type: values.type,
    start_date: optionalString(values.start_date),
    due_date: optionalString(values.due_date),
    est_hours: values.est_hours,
  };
}

export function toTaskUpdatePayload(values: TaskFormValues): UpdateTaskRequest {
  return {
    ...toTaskCreatePayload(values),
    progress_pct: values.progress_pct,
  };
}

export function toMilestoneCreatePayload(values: MilestoneFormValues): CreateMilestoneRequest {
  return {
    title: values.title,
    description: optionalString(values.description),
    due_date: optionalString(values.due_date),
  };
}

export function toMilestoneUpdatePayload(values: MilestoneFormValues): UpdateMilestoneRequest {
  return {
    ...toMilestoneCreatePayload(values),
    progress_pct: values.progress_pct,
  };
}

export function readIssueForm(form: HTMLFormElement): IssueFormValues {
  const data = new FormData(form);
  return {
    title: getString(data, "title"),
    description: getString(data, "description"),
    severity: getString(data, "severity") as IssueFormValues["severity"],
    escalation: getString(data, "escalation") as IssueFormValues["escalation"],
    assigned_to: getString(data, "assigned_to"),
    due_date: getString(data, "due_date"),
    resolution: getString(data, "resolution"),
  };
}

export function toIssueCreatePayload(values: IssueFormValues): CreateIssueRequest {
  return {
    title: values.title,
    description: optionalString(values.description),
    severity: values.severity,
    escalation: values.escalation,
    assigned_to: optionalString(values.assigned_to),
    due_date: optionalString(values.due_date),
    resolution: optionalString(values.resolution),
  };
}

export function toIssueUpdatePayload(values: IssueFormValues): UpdateIssueRequest {
  return {
    ...toIssueCreatePayload(values),
  };
}

export function readRiskForm(form: HTMLFormElement): RiskFormValues {
  const data = new FormData(form);
  return {
    title: getString(data, "title"),
    description: getString(data, "description"),
    probability: Number(getString(data, "probability") || 0),
    impact: Number(getString(data, "impact") || 0),
    mitigation: getString(data, "mitigation"),
    owned_by: getString(data, "owned_by"),
    due_date: getString(data, "due_date"),
  };
}

export function toRiskCreatePayload(values: RiskFormValues): CreateRiskRequest {
  return {
    title: values.title,
    description: optionalString(values.description),
    probability: values.probability,
    impact: values.impact,
    mitigation: optionalString(values.mitigation),
    owned_by: optionalString(values.owned_by),
    due_date: optionalString(values.due_date),
  };
}

export function toRiskUpdatePayload(values: RiskFormValues): UpdateRiskRequest {
  return {
    ...toRiskCreatePayload(values),
  };
}

export function riskSeverityColor(severity: RiskSeverity) {
  switch (severity) {
    case "CRITICAL":
      return "font-semibold text-red-600";
    case "HIGH":
      return "font-semibold text-orange-600";
    case "MEDIUM":
      return "font-semibold text-yellow-600";
    default:
      return "font-semibold text-green-600";
  }
}

export function readBudgetForm(form: HTMLFormElement): BudgetFormValues {
  const data = new FormData(form);
  return {
    category: getString(data, "category"),
    description: getString(data, "description"),
    planned: Number(getString(data, "planned") || 0),
    actual: Number(getString(data, "actual") || 0),
    currency: getString(data, "currency"),
  };
}

export function toBudgetCreatePayload(values: BudgetFormValues): CreateBudgetRequest {
  return {
    category: values.category,
    description: optionalString(values.description),
    planned: values.planned,
    actual: values.actual,
    currency: optionalString(values.currency) ?? "IDR",
  };
}

export function toBudgetUpdatePayload(values: BudgetFormValues): UpdateBudgetRequest {
  return {
    ...toBudgetCreatePayload(values),
  };
}

export function readContractForm(form: HTMLFormElement): ContractFormValues {
  const data = new FormData(form);
  return {
    contract_number: getString(data, "contract_number"),
    title: getString(data, "title"),
    vendor_id: getString(data, "vendor_id"),
    consultant_id: getString(data, "consultant_id"),
    contract_value: Number(getString(data, "contract_value") || 0),
    currency: getString(data, "currency"),
    signed_date: getString(data, "signed_date"),
    start_date: getString(data, "start_date"),
    end_date: getString(data, "end_date"),
    status: getString(data, "status") as ContractFormValues["status"],
    scope_of_work: getString(data, "scope_of_work"),
    ppk_user_id: getString(data, "ppk_user_id"),
    agreement_form: getString(data, "agreement_form"),
    sirup_code: getString(data, "sirup_code"),
    spse_tender_code: getString(data, "spse_tender_code"),
  };
}

export function toContractCreatePayload(values: ContractFormValues): CreateContractRequest {
  return {
    contract_number: values.contract_number,
    title: values.title,
    vendor_id: values.vendor_id,
    consultant_id: optionalString(values.consultant_id),
    contract_value: values.contract_value,
    currency: optionalString(values.currency) ?? "IDR",
    signed_date: optionalString(values.signed_date),
    start_date: optionalString(values.start_date),
    end_date: optionalString(values.end_date),
    status: values.status,
    scope_of_work: optionalString(values.scope_of_work),
    ppk_user_id: optionalString(values.ppk_user_id) ?? null,
    agreement_form: values.agreement_form ?? "",
    sirup_code: values.sirup_code ?? "",
    spse_tender_code: values.spse_tender_code ?? "",
  };
}

export function toContractUpdatePayload(values: ContractFormValues): UpdateContractRequest {
  return {
    contract_number: values.contract_number,
    title: values.title,
    vendor_id: values.vendor_id,
    consultant_id: values.consultant_id ? values.consultant_id : null,
    contract_value: values.contract_value,
    currency: optionalString(values.currency) ?? "IDR",
    signed_date: optionalString(values.signed_date),
    start_date: optionalString(values.start_date),
    end_date: optionalString(values.end_date),
    status: values.status,
    scope_of_work: optionalString(values.scope_of_work),
    ppk_user_id: values.ppk_user_id ? values.ppk_user_id : "00000000-0000-0000-0000-000000000000",
    agreement_form: values.agreement_form ?? "",
    sirup_code: values.sirup_code ?? "",
    spse_tender_code: values.spse_tender_code ?? "",
  };
}

export function readVendorForm(form: HTMLFormElement): VendorFormValues {
  const data = new FormData(form);
  return {
    name: getString(data, "name"),
    legal_name: getString(data, "legal_name"),
    tax_id: getString(data, "tax_id"),
    contact_person: getString(data, "contact_person"),
    email: getString(data, "email"),
    phone: getString(data, "phone"),
    address: getString(data, "address"),
  };
}

export function toVendorCreatePayload(
  values: VendorFormValues,
  type: VendorType
): CreateVendorRequest {
  return {
    name: values.name,
    type,
    legal_name: optionalString(values.legal_name),
    tax_id: optionalString(values.tax_id),
    contact_person: optionalString(values.contact_person),
    email: optionalString(values.email),
    phone: optionalString(values.phone),
    address: optionalString(values.address),
  };
}

export function readDocumentForm(form: HTMLFormElement): DocumentFormValues {
  const data = new FormData(form);
  const file = data.get("file");
  return {
    name: getString(data, "name"),
    category: getString(data, "category"),
    version: getString(data, "version"),
    file: file instanceof File && file.size > 0 ? file : undefined,
  };
}

export function toDocumentUpdatePayload(
  values: DocumentFormValues
): UpdateDocumentRequest {
  return {
    name: optionalString(values.name),
    category: optionalString(values.category),
    version: optionalString(values.version),
  };
}

export function documentCategoryTone(category: string | null | undefined) {
  switch (category) {
    case "CONTRACT":
      return "bg-blue-100 text-blue-700";
    case "BAST":
      return "bg-purple-100 text-purple-700";
    case "RELATION_EVIDENCE":
    case "EVIDENCE":
      return "bg-green-100 text-green-700";
    case "TOR_KAK":
      return "bg-amber-100 text-amber-700";
    default:
      return "bg-gray-100 text-gray-700";
  }
}

export function formatFileSize(bytes: number) {
  if (!bytes) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  const i = Math.min(
    Math.floor(Math.log(bytes) / Math.log(1024)),
    units.length - 1
  );
  const value = bytes / Math.pow(1024, i);
  return `${value.toFixed(value >= 100 || i === 0 ? 0 : 1)} ${units[i]}`;
}

export function contractStatusTone(status: ContractStatus) {
  switch (status) {
    case "ACTIVE":
      return "bg-green-100 text-green-700";
    case "AMENDED":
      return "bg-yellow-100 text-yellow-700";
    case "COMPLETED":
      return "bg-blue-100 text-blue-700";
    case "TERMINATED":
      return "bg-red-100 text-red-700";
    default:
      return "bg-gray-100 text-gray-700";
  }
}

export function budgetStatusTone(status: BudgetStatus) {
  switch (status) {
    case "OVERRUN":
      return "bg-red-100 text-red-700";
    case "RISK":
      return "bg-orange-100 text-orange-700";
    case "WATCH":
      return "bg-yellow-100 text-yellow-700";
    default:
      return "bg-green-100 text-green-700";
  }
}

export function formatCurrency(value: number) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(value);
}

export function getString(data: FormData, key: string) {
  const value = data.get(key);
  return typeof value === "string" ? value : "";
}

export function optionalString(value: string | undefined) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

export function toDateInput(value: string | null | undefined) {
  if (!value) return "";
  return value.slice(0, 10);
}

export function caStatusTone(status: CorrectiveActionStatus): string {
  switch (status) {
    case "DRAFT": return "bg-gray-100 text-gray-700";
    case "SUBMITTED": return "bg-blue-100 text-blue-700";
    case "IN_PROGRESS": return "bg-yellow-100 text-yellow-800";
    case "COMPLETED": return "bg-green-100 text-green-700";
    case "REJECTED": return "bg-red-100 text-red-700";
    default: return "bg-gray-100 text-gray-700";
  }
}
