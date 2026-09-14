"use client";

import { cn } from "@/lib/utils";
import type {
  ContractStatus,
  CorrectiveActionStatus,
  DocumentCategory,
  IssueEscalation,
  IssueSeverity,
  IssueStatus,
  MilestoneStatus,
  RiskSeverity,
  RiskStatus,
  TaskStatus,
} from "@/types/project";
import { z } from "zod";

export const TASK_STATUSES: TaskStatus[] = [
  "TODO",
  "IN_PROGRESS",
  "IN_REVIEW",
  "BLOCKED",
  "DONE",
];

export const TASK_NEXT_STATUS: Record<TaskStatus, TaskStatus[]> = {
  TODO: ["IN_PROGRESS"],
  IN_PROGRESS: ["IN_REVIEW", "BLOCKED"],
  IN_REVIEW: ["DONE", "IN_PROGRESS"],
  BLOCKED: ["IN_PROGRESS"],
  DONE: ["IN_PROGRESS"],
};

export const MILESTONE_NEXT_STATUS: Record<MilestoneStatus, MilestoneStatus[]> = {
  PENDING: ["IN_PROGRESS"],
  IN_PROGRESS: ["COMPLETED", "DELAYED"],
  DELAYED: ["IN_PROGRESS"],
  COMPLETED: [],
};

export const ISSUE_NEXT_STATUS: Record<IssueStatus, IssueStatus[]> = {
  OPEN: ["IN_PROGRESS"],
  IN_PROGRESS: ["RESOLVED"],
  RESOLVED: ["CLOSED", "OPEN"],
  CLOSED: ["OPEN"],
};

export const CA_NEXT_STATUS: Record<CorrectiveActionStatus, CorrectiveActionStatus[]> = {
  DRAFT: ["SUBMITTED"],
  SUBMITTED: ["IN_PROGRESS", "REJECTED"],
  IN_PROGRESS: ["COMPLETED", "REJECTED"],
  COMPLETED: [],
  REJECTED: ["DRAFT"],
};

export const ISSUE_SEVERITIES: IssueSeverity[] = ["LOW", "MEDIUM", "HIGH", "CRITICAL"];

export const ISSUE_ESCALATIONS: IssueEscalation[] = [
  "NONE",
  "PROJECT_MANAGER",
  "PROGRAM_MANAGER",
  "EXECUTIVE",
];

export const RISK_NEXT_STATUS: Record<RiskStatus, RiskStatus[]> = {
  IDENTIFIED: ["ASSESSED"],
  ASSESSED: ["MITIGATED", "ACCEPTED", "ESCALATED"],
  MITIGATED: ["CLOSED"],
  ACCEPTED: ["CLOSED"],
  ESCALATED: ["MITIGATED"],
  CLOSED: [],
};

export const RISK_SEVERITIES: RiskSeverity[] = ["LOW", "MEDIUM", "HIGH", "CRITICAL"];

export const CONTRACT_STATUSES: ContractStatus[] = [
  "DRAFT",
  "ACTIVE",
  "AMENDED",
  "COMPLETED",
  "TERMINATED",
];

export const DOCUMENT_CATEGORIES: DocumentCategory[] = [
  "CONTRACT",
  "REPORT",
  "EVIDENCE",
  "PHOTO",
  "BAST",
  "TOR_KAK",
  "OTHER",
];

export const taskSchema = z.object({
  title: z.string().trim().min(1, "Task title is required").max(500),
  description: z.string().trim().optional(),
  priority: z.enum(["LOW", "MEDIUM", "HIGH", "CRITICAL"]),
  type: z.enum(["TASK", "BUG", "FEATURE", "RESEARCH"]),
  start_date: z.string().trim().optional(),
  due_date: z.string().trim().optional(),
  est_hours: z.coerce.number().min(0),
  progress_pct: z.coerce.number().min(0).max(100),
});

export const milestoneSchema = z.object({
  title: z.string().trim().min(1, "Milestone title is required").max(500),
  description: z.string().trim().optional(),
  due_date: z.string().trim().optional(),
  progress_pct: z.coerce.number().min(0).max(100),
});

export const issueSchema = z.object({
  title: z.string().trim().min(1, "Issue title is required").max(500),
  description: z.string().trim().optional(),
  severity: z.enum(["LOW", "MEDIUM", "HIGH", "CRITICAL"]),
  escalation: z.enum(["NONE", "PROJECT_MANAGER", "PROGRAM_MANAGER", "EXECUTIVE"]),
  assigned_to: z.string().trim().optional(),
  due_date: z.string().trim().optional(),
  resolution: z.string().trim().optional(),
});

export const riskSchema = z.object({
  title: z.string().trim().min(1, "Risk title is required").max(500),
  description: z.string().trim().optional(),
  probability: z.coerce.number().min(1).max(5),
  impact: z.coerce.number().min(1).max(5),
  mitigation: z.string().trim().optional(),
  owned_by: z.string().trim().optional(),
  due_date: z.string().trim().optional(),
});

export const budgetSchema = z.object({
  category: z.string().trim().min(1, "Category is required").max(200),
  description: z.string().trim().optional(),
  planned: z.coerce.number().min(0, "Planned cannot be negative"),
  actual: z.coerce.number().min(0, "Actual cannot be negative"),
  currency: z.string().trim().max(10).optional(),
});

export const contractSchema = z.object({
  contract_number: z.string().trim().min(1, "Contract number is required").max(200),
  title: z.string().trim().min(1, "Title is required").max(500),
  vendor_id: z.string().trim().min(1, "Vendor is required"),
  consultant_id: z.string().trim().optional(),
  contract_value: z.coerce.number().min(0, "Value cannot be negative"),
  currency: z.string().trim().max(10).optional(),
  signed_date: z.string().trim().optional(),
  start_date: z.string().trim().optional(),
  end_date: z.string().trim().optional(),
  status: z.enum(["DRAFT", "ACTIVE", "AMENDED", "COMPLETED", "TERMINATED"]),
  scope_of_work: z.string().trim().optional(),
});

export const vendorSchema = z.object({
  name: z.string().trim().min(1, "Vendor name is required").max(500),
  legal_name: z.string().trim().optional(),
  tax_id: z.string().trim().optional(),
  contact_person: z.string().trim().optional(),
  email: z.string().trim().optional(),
  phone: z.string().trim().optional(),
  address: z.string().trim().optional(),
});

export const documentSchema = z.object({
  name: z.string().trim().max(500).optional(),
  category: z.string().trim().max(100).optional(),
  version: z.string().trim().max(100).optional(),
  file: z.instanceof(File).optional(),
});

export type TaskFormValues = z.infer<typeof taskSchema>;

export type MilestoneFormValues = z.infer<typeof milestoneSchema>;

export type IssueFormValues = z.infer<typeof issueSchema>;

export type RiskFormValues = z.infer<typeof riskSchema>;

export type BudgetFormValues = z.infer<typeof budgetSchema>;

export type ContractFormValues = z.infer<typeof contractSchema>;

export type VendorFormValues = z.infer<typeof vendorSchema>;

export type DocumentFormValues = z.infer<typeof documentSchema>;

export const inputClassName = cn(
  "h-9 w-full rounded-md border border-input bg-background px-3 text-sm",
  "placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
);

export const primaryButton = cn(
  "inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground",
  "hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
);

export const secondaryButton = cn(
  "inline-flex h-9 items-center justify-center gap-2 rounded-md border border-input px-3 text-sm font-medium",
  "hover:bg-accent hover:text-accent-foreground"
);
