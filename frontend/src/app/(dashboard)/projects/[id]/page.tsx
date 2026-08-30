"use client";

import Link from "next/link";
import { type FormEvent, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AlertCircle,
  ArrowLeft,
  Banknote,
  CalendarDays,
  CheckCircle2,
  Download,
  Edit3,
  FileText,
  Flag,
  Milestone as MilestoneIcon,
  Plus,
  Save,
  ShieldAlert,
  Trash2,
  X,
} from "lucide-react";
import { z } from "zod";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { GanttPanel } from "@/components/project/GanttPanel";
import { FieldInspectionPanel } from "@/components/project/FieldInspectionPanel";
import { HealthScorePanel } from "@/components/project/HealthScorePanel";
import { ProjectControlPanel } from "@/components/project/ProjectControlPanel";
import { PeriodicReportPanel } from "@/components/project/PeriodicReportPanel";
import { projectService } from "@/services/project.service";
import { baselineService, snapshotService } from "@/services/monitoring.service";
import { cn, formatDate } from "@/lib/utils";
import type {
  Baseline,
  CreateBaselineRequest,
  UpdateBaselineRequest,
  Snapshot,
  CreateSnapshotRequest,
  UpdateSnapshotRequest,
  TransitionSnapshotRequest,
  SnapshotStatus,
} from "@/types/monitoring";
import {
  SNAPSHOT_STATUS_LABEL,
  SNAPSHOT_STATUS_COLOR,
  MONTH_NAMES,
} from "@/types/monitoring";
import type {
  BudgetLine,
  BudgetStatus,
  Contract,
  ContractStatus,
  CreateBudgetRequest,
  CreateContractRequest,
  CreateIssueRequest,
  CreateMilestoneRequest,
  CreateRiskRequest,
  CreateTaskRequest,
  CreateVendorRequest,
  DocumentCategory,
  Issue,
  IssueEscalation,
  IssueSeverity,
  IssueStatus,
  Milestone,
  MilestoneStatus,
  Project,
  ProjectDocument,
  Risk,
  RiskSeverity,
  RiskStatus,
  Task,
  TaskStatus,
  UpdateBudgetRequest,
  UpdateContractRequest,
  UpdateDocumentRequest,
  UpdateIssueRequest,
  UpdateMilestoneRequest,
  UpdateRiskRequest,
  UpdateTaskRequest,
  UpdateVendorRequest,
  Vendor,
  VendorType,
  CorrectiveAction,
  CorrectiveActionStatus,
  CreateCorrectiveActionRequest,
  UpdateCorrectiveActionRequest,
} from "@/types/project";

const TASK_STATUSES: TaskStatus[] = [
  "TODO",
  "IN_PROGRESS",
  "IN_REVIEW",
  "BLOCKED",
  "DONE",
];

const TASK_NEXT_STATUS: Record<TaskStatus, TaskStatus[]> = {
  TODO: ["IN_PROGRESS"],
  IN_PROGRESS: ["IN_REVIEW", "BLOCKED"],
  IN_REVIEW: ["DONE", "IN_PROGRESS"],
  BLOCKED: ["IN_PROGRESS"],
  DONE: ["IN_PROGRESS"],
};

const MILESTONE_NEXT_STATUS: Record<MilestoneStatus, MilestoneStatus[]> = {
  PENDING: ["IN_PROGRESS"],
  IN_PROGRESS: ["COMPLETED", "DELAYED"],
  DELAYED: ["IN_PROGRESS"],
  COMPLETED: [],
};

const ISSUE_NEXT_STATUS: Record<IssueStatus, IssueStatus[]> = {
  OPEN: ["IN_PROGRESS"],
  IN_PROGRESS: ["RESOLVED"],
  RESOLVED: ["CLOSED", "OPEN"],
  CLOSED: ["OPEN"],
};

const CA_NEXT_STATUS: Record<CorrectiveActionStatus, CorrectiveActionStatus[]> = {
  DRAFT: ["SUBMITTED"],
  SUBMITTED: ["IN_PROGRESS", "REJECTED"],
  IN_PROGRESS: ["COMPLETED", "REJECTED"],
  COMPLETED: [],
  REJECTED: ["DRAFT"],
};

const ISSUE_SEVERITIES: IssueSeverity[] = ["LOW", "MEDIUM", "HIGH", "CRITICAL"];
const ISSUE_ESCALATIONS: IssueEscalation[] = [
  "NONE",
  "PROJECT_MANAGER",
  "PROGRAM_MANAGER",
  "EXECUTIVE",
];

const RISK_NEXT_STATUS: Record<RiskStatus, RiskStatus[]> = {
  IDENTIFIED: ["ASSESSED"],
  ASSESSED: ["MITIGATED", "ACCEPTED", "ESCALATED"],
  MITIGATED: ["CLOSED"],
  ACCEPTED: ["CLOSED"],
  ESCALATED: ["MITIGATED"],
  CLOSED: [],
};

const RISK_SEVERITIES: RiskSeverity[] = ["LOW", "MEDIUM", "HIGH", "CRITICAL"];
const CONTRACT_STATUSES: ContractStatus[] = [
  "DRAFT",
  "ACTIVE",
  "AMENDED",
  "COMPLETED",
  "TERMINATED",
];

const DOCUMENT_CATEGORIES: DocumentCategory[] = [
  "CONTRACT",
  "REPORT",
  "EVIDENCE",
  "PHOTO",
  "BAST",
  "TOR_KAK",
  "OTHER",
];

const taskSchema = z.object({
  title: z.string().trim().min(1, "Task title is required").max(500),
  description: z.string().trim().optional(),
  priority: z.enum(["LOW", "MEDIUM", "HIGH", "CRITICAL"]),
  type: z.enum(["TASK", "BUG", "FEATURE", "RESEARCH"]),
  start_date: z.string().trim().optional(),
  due_date: z.string().trim().optional(),
  est_hours: z.coerce.number().min(0),
  progress_pct: z.coerce.number().min(0).max(100),
});

const milestoneSchema = z.object({
  title: z.string().trim().min(1, "Milestone title is required").max(500),
  description: z.string().trim().optional(),
  due_date: z.string().trim().optional(),
  progress_pct: z.coerce.number().min(0).max(100),
});

const issueSchema = z.object({
  title: z.string().trim().min(1, "Issue title is required").max(500),
  description: z.string().trim().optional(),
  severity: z.enum(["LOW", "MEDIUM", "HIGH", "CRITICAL"]),
  escalation: z.enum(["NONE", "PROJECT_MANAGER", "PROGRAM_MANAGER", "EXECUTIVE"]),
  assigned_to: z.string().trim().optional(),
  due_date: z.string().trim().optional(),
  resolution: z.string().trim().optional(),
});

const riskSchema = z.object({
  title: z.string().trim().min(1, "Risk title is required").max(500),
  description: z.string().trim().optional(),
  probability: z.coerce.number().min(1).max(5),
  impact: z.coerce.number().min(1).max(5),
  mitigation: z.string().trim().optional(),
  owned_by: z.string().trim().optional(),
  due_date: z.string().trim().optional(),
});

const budgetSchema = z.object({
  category: z.string().trim().min(1, "Category is required").max(200),
  description: z.string().trim().optional(),
  planned: z.coerce.number().min(0, "Planned cannot be negative"),
  actual: z.coerce.number().min(0, "Actual cannot be negative"),
  currency: z.string().trim().max(10).optional(),
});

const contractSchema = z.object({
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

const vendorSchema = z.object({
  name: z.string().trim().min(1, "Vendor name is required").max(500),
  legal_name: z.string().trim().optional(),
  tax_id: z.string().trim().optional(),
  contact_person: z.string().trim().optional(),
  email: z.string().trim().optional(),
  phone: z.string().trim().optional(),
  address: z.string().trim().optional(),
});

const documentSchema = z.object({
  name: z.string().trim().max(500).optional(),
  category: z.string().trim().max(100).optional(),
  version: z.string().trim().max(100).optional(),
  file: z.instanceof(File).optional(),
});

type TaskFormValues = z.infer<typeof taskSchema>;
type MilestoneFormValues = z.infer<typeof milestoneSchema>;
type IssueFormValues = z.infer<typeof issueSchema>;
type RiskFormValues = z.infer<typeof riskSchema>;
type BudgetFormValues = z.infer<typeof budgetSchema>;
type ContractFormValues = z.infer<typeof contractSchema>;
type VendorFormValues = z.infer<typeof vendorSchema>;
type DocumentFormValues = z.infer<typeof documentSchema>;

export default function ProjectDetailPage() {
  const params = useParams<{ id: string }>();
  const projectId = params.id;
  const qc = useQueryClient();
  const [taskFormOpen, setTaskFormOpen] = useState(false);
  const [milestoneFormOpen, setMilestoneFormOpen] = useState(false);
  const [issueFormOpen, setIssueFormOpen] = useState(false);
  const [riskFormOpen, setRiskFormOpen] = useState(false);
  const [budgetFormOpen, setBudgetFormOpen] = useState(false);
  const [contractFormOpen, setContractFormOpen] = useState(false);
  const [vendorFormOpen, setVendorFormOpen] = useState(false);
  const [vendorFormType, setVendorFormType] = useState<VendorType>("VENDOR");
  const [documentFormOpen, setDocumentFormOpen] = useState(false);
  const [editingDocument, setEditingDocument] = useState<ProjectDocument | null>(null);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [editingMilestone, setEditingMilestone] = useState<Milestone | null>(null);
  const [editingIssue, setEditingIssue] = useState<Issue | null>(null);
  const [editingRisk, setEditingRisk] = useState<Risk | null>(null);
  const [editingBudget, setEditingBudget] = useState<BudgetLine | null>(null);
  const [editingContract, setEditingContract] = useState<Contract | null>(null);
  const [editingVendor, setEditingVendor] = useState<Vendor | null>(null);
  const [caFormOpen, setCAFormOpen] = useState(false);
  const [editingCA, setEditingCA] = useState<CorrectiveAction | null>(null);
  const [formError, setFormError] = useState<string | null>(null);

  // ── Monitoring state ──────────────────────────────────────────────────────
  const [baselineFormOpen, setBaselineFormOpen] = useState(false);
  const [editingBaseline, setEditingBaseline] = useState<Baseline | null>(null);
  const [snapshotFormOpen, setSnapshotFormOpen] = useState(false);
  const [editingSnapshot, setEditingSnapshot] = useState<Snapshot | null>(null);
  const [snapshotStatusFilter, setSnapshotStatusFilter] = useState<SnapshotStatus | "">("");

  const projectQuery = useQuery({
    queryKey: ["projects", projectId],
    queryFn: () => projectService.get(projectId),
    select: (res) => res.data.data,
  });

  const tasksQuery = useQuery({
    queryKey: ["projects", projectId, "tasks"],
    queryFn: () => projectService.listTasks(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const milestonesQuery = useQuery({
    queryKey: ["projects", projectId, "milestones"],
    queryFn: () => projectService.listMilestones(projectId),
    select: (res) => res.data.data ?? [],
  });

  const issuesQuery = useQuery({
    queryKey: ["projects", projectId, "issues"],
    queryFn: () => projectService.listIssues(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const risksQuery = useQuery({
    queryKey: ["projects", projectId, "risks"],
    queryFn: () => projectService.listRisks(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const budgetsQuery = useQuery({
    queryKey: ["projects", projectId, "budgets"],
    queryFn: () => projectService.listBudgets(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const contractsQuery = useQuery({
    queryKey: ["projects", projectId, "contracts"],
    queryFn: () => projectService.listContracts(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const documentsQuery = useQuery({
    queryKey: ["projects", projectId, "documents"],
    queryFn: () => projectService.listDocuments(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const vendorsQuery = useQuery({
    queryKey: ["vendors", "all"],
    queryFn: () => projectService.listVendors({ page: 1, page_size: 100, is_active: true }),
    select: (res) => res.data.data ?? [],
  });

  const correctiveActionsQuery = useQuery({
    queryKey: ["projects", projectId, "corrective-actions"],
    queryFn: () => projectService.listCorrectiveActions(projectId, { page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const tasks = useMemo(() => tasksQuery.data ?? [], [tasksQuery.data]);
  const milestones = useMemo(
    () => milestonesQuery.data ?? [],
    [milestonesQuery.data]
  );
  const issues = useMemo(() => issuesQuery.data ?? [], [issuesQuery.data]);
  const risks = useMemo(() => risksQuery.data ?? [], [risksQuery.data]);
  const budgets = useMemo(() => budgetsQuery.data ?? [], [budgetsQuery.data]);
  const contracts = useMemo(() => contractsQuery.data ?? [], [contractsQuery.data]);
  const documents = useMemo(() => documentsQuery.data ?? [], [documentsQuery.data]);
  const vendors = useMemo(() => vendorsQuery.data ?? [], [vendorsQuery.data]);
  const correctiveActions = useMemo(() => correctiveActionsQuery.data ?? [], [correctiveActionsQuery.data]);
  const vendorOptions = useMemo(
    () => vendors.filter((v) => v.type === "VENDOR" && v.is_active),
    [vendors]
  );
  const consultantOptions = useMemo(
    () => vendors.filter((v) => v.type === "CONSULTANT" && v.is_active),
    [vendors]
  );

  const contractTotals = useMemo(() => {
    let totalValue = 0;
    let activeCount = 0;
    let mainVendorId: string | null = null;
    const vendorValue: Record<string, number> = {};
    const sorted = [...contracts].sort((a, b) => b.contract_value - a.contract_value);
    for (const contract of contracts) {
      totalValue += contract.contract_value;
      if (contract.status === "ACTIVE") activeCount += 1;
      vendorValue[contract.vendor_id] =
        (vendorValue[contract.vendor_id] ?? 0) + contract.contract_value;
    }
    const main = sorted[0] ?? null;
    if (main) mainVendorId = main.vendor_id;
    let period = "–";
    if (contracts.length > 0) {
      const start = contracts.reduce<string | null>(
        (acc, c) => (c.start_date && (!acc || c.start_date < acc) ? c.start_date : acc),
        null
      );
      const end = contracts.reduce<string | null>(
        (acc, c) => (c.end_date && (!acc || c.end_date > acc) ? c.end_date : acc),
        null
      );
      if (start || end) period = `${start ? toDateInput(start) : "?"} — ${end ? toDateInput(end) : "?"}`;
    }
    return { totalValue, activeCount, mainVendorId, mainValue: main?.contract_value ?? 0, period };
  }, [contracts]);

  const budgetTotals = useMemo(() => {
    let planned = 0;
    let actual = 0;
    for (const line of budgets) {
      planned += line.planned;
      actual += line.actual;
    }
    const usage = planned > 0 ? Math.round((actual / planned) * 10000) / 100 : 0;
    return { planned, actual, variance: planned - actual, usage };
  }, [budgets]);

  const documentTotals = useMemo(() => {
    let totalSize = 0;
    const byCategory: Record<string, number> = {};
    for (const doc of documents) {
      totalSize += doc.file_size ?? 0;
      const cat = doc.category || "UNKNOWN";
      byCategory[cat] = (byCategory[cat] ?? 0) + 1;
    }
    return { count: documents.length, totalSize, byCategory };
  }, [documents]);

  const taskGroups = useMemo(
    () =>
      TASK_STATUSES.map((status) => ({
        status,
        tasks: tasks.filter((task) => task.status === status),
      })),
    [tasks]
  );

  const refreshProjectData = () => {
    void qc.invalidateQueries({ queryKey: ["projects", projectId] });
    void qc.invalidateQueries({ queryKey: ["projects", projectId, "tasks"] });
    void qc.invalidateQueries({ queryKey: ["projects", projectId, "milestones"] });
    void qc.invalidateQueries({ queryKey: ["projects", projectId, "issues"] });
    void qc.invalidateQueries({ queryKey: ["projects", projectId, "risks"] });
    void qc.invalidateQueries({ queryKey: ["projects", projectId, "budgets"] });
    void qc.invalidateQueries({ queryKey: ["projects", projectId, "contracts"] });
    void qc.invalidateQueries({ queryKey: ["projects", projectId, "documents"] });
    void qc.invalidateQueries({ queryKey: ["vendors", "all"] });
    void qc.invalidateQueries({ queryKey: ["dashboard", "stats"] });
    void qc.invalidateQueries({ queryKey: ["projects", projectId, "baselines"] });
    void qc.invalidateQueries({ queryKey: ["projects", projectId, "snapshots"] });
  };

  // ── Monitoring queries ───────────────────────────────────────────────────
  const baselinesQuery = useQuery({
    queryKey: ["projects", projectId, "baselines"],
    queryFn: () => baselineService.list(projectId),
    select: (data) => data,
  });
  const baselines = baselinesQuery.data ?? [];

  const snapshotsQuery = useQuery({
    queryKey: ["projects", projectId, "snapshots"],
    queryFn: () => snapshotService.list(projectId),
    select: (data) => data,
  });
  const snapshots = snapshotsQuery.data ?? [];

  // ── Monitoring mutations ──────────────────────────────────────────────────
  const createBaselineMutation = useMutation({
    mutationFn: (payload: CreateBaselineRequest) => baselineService.create(projectId, payload),
    onSuccess: () => {
      setBaselineFormOpen(false);
      setEditingBaseline(null);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "baselines"] });
    },
    onError: () => setFormError("Gagal menyimpan baseline."),
  });

  const updateBaselineMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateBaselineRequest }) =>
      baselineService.update(projectId, id, payload),
    onSuccess: () => {
      setBaselineFormOpen(false);
      setEditingBaseline(null);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "baselines"] });
    },
    onError: () => setFormError("Gagal memperbarui baseline."),
  });

  const deleteBaselineMutation = useMutation({
    mutationFn: (id: string) => baselineService.delete(projectId, id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["projects", projectId, "baselines"] }),
  });

  const createSnapshotMutation = useMutation({
    mutationFn: (payload: CreateSnapshotRequest) => snapshotService.create(projectId, payload),
    onSuccess: () => {
      setSnapshotFormOpen(false);
      setEditingSnapshot(null);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "snapshots"] });
    },
    onError: () => setFormError("Gagal menyimpan snapshot."),
  });

  const updateSnapshotMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateSnapshotRequest }) =>
      snapshotService.update(projectId, id, payload),
    onSuccess: () => {
      setSnapshotFormOpen(false);
      setEditingSnapshot(null);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "snapshots"] });
    },
    onError: () => setFormError("Gagal memperbarui snapshot."),
  });

  const transitionSnapshotMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: TransitionSnapshotRequest }) =>
      snapshotService.transition(projectId, id, payload),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["projects", projectId, "snapshots"] }),
  });

  const deleteSnapshotMutation = useMutation({
    mutationFn: (id: string) => snapshotService.delete(projectId, id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["projects", projectId, "snapshots"] }),
  });

  const createTaskMutation = useMutation({
    mutationFn: (payload: CreateTaskRequest) => projectService.createTask(projectId, payload),
    onSuccess: () => {
      setTaskFormOpen(false);
      setEditingTask(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateTaskMutation = useMutation({
    mutationFn: ({ taskId, payload }: { taskId: string; payload: UpdateTaskRequest }) =>
      projectService.updateTask(projectId, taskId, payload),
    onSuccess: () => {
      setTaskFormOpen(false);
      setEditingTask(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteTaskMutation = useMutation({
    mutationFn: (taskId: string) => projectService.deleteTask(projectId, taskId),
    onSuccess: refreshProjectData,
  });

  const createMilestoneMutation = useMutation({
    mutationFn: (payload: CreateMilestoneRequest) =>
      projectService.createMilestone(projectId, payload),
    onSuccess: () => {
      setMilestoneFormOpen(false);
      setEditingMilestone(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateMilestoneMutation = useMutation({
    mutationFn: ({
      milestoneId,
      payload,
    }: {
      milestoneId: string;
      payload: UpdateMilestoneRequest;
    }) => projectService.updateMilestone(projectId, milestoneId, payload),
    onSuccess: () => {
      setMilestoneFormOpen(false);
      setEditingMilestone(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteMilestoneMutation = useMutation({
    mutationFn: (milestoneId: string) =>
      projectService.deleteMilestone(projectId, milestoneId),
    onSuccess: refreshProjectData,
  });

  const createIssueMutation = useMutation({
    mutationFn: (payload: CreateIssueRequest) =>
      projectService.createIssue(projectId, payload),
    onSuccess: () => {
      setIssueFormOpen(false);
      setEditingIssue(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateIssueMutation = useMutation({
    mutationFn: ({
      issueId,
      payload,
    }: {
      issueId: string;
      payload: UpdateIssueRequest;
    }) => projectService.updateIssue(projectId, issueId, payload),
    onSuccess: () => {
      setIssueFormOpen(false);
      setEditingIssue(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteIssueMutation = useMutation({
    mutationFn: (issueId: string) =>
      projectService.deleteIssue(projectId, issueId),
    onSuccess: refreshProjectData,
  });

  const createRiskMutation = useMutation({
    mutationFn: (payload: CreateRiskRequest) =>
      projectService.createRisk(projectId, payload),
    onSuccess: () => {
      setRiskFormOpen(false);
      setEditingRisk(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateRiskMutation = useMutation({
    mutationFn: ({
      riskId,
      payload,
    }: {
      riskId: string;
      payload: UpdateRiskRequest;
    }) => projectService.updateRisk(projectId, riskId, payload),
    onSuccess: () => {
      setRiskFormOpen(false);
      setEditingRisk(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteRiskMutation = useMutation({
    mutationFn: (riskId: string) =>
      projectService.deleteRisk(projectId, riskId),
    onSuccess: refreshProjectData,
  });

  const createBudgetMutation = useMutation({
    mutationFn: (payload: CreateBudgetRequest) =>
      projectService.createBudget(projectId, payload),
    onSuccess: () => {
      setBudgetFormOpen(false);
      setEditingBudget(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateBudgetMutation = useMutation({
    mutationFn: ({
      budgetId,
      payload,
    }: {
      budgetId: string;
      payload: UpdateBudgetRequest;
    }) => projectService.updateBudget(projectId, budgetId, payload),
    onSuccess: () => {
      setBudgetFormOpen(false);
      setEditingBudget(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteBudgetMutation = useMutation({
    mutationFn: (budgetId: string) =>
      projectService.deleteBudget(projectId, budgetId),
    onSuccess: refreshProjectData,
  });

  const createVendorMutation = useMutation({
    mutationFn: (payload: CreateVendorRequest) =>
      projectService.createVendor(payload),
    onSuccess: () => {
      setVendorFormOpen(false);
      setEditingVendor(null);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["vendors", "all"] });
    },
  });

  const createContractMutation = useMutation({
    mutationFn: (payload: CreateContractRequest) =>
      projectService.createContract(projectId, payload),
    onSuccess: () => {
      setContractFormOpen(false);
      setEditingContract(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateContractMutation = useMutation({
    mutationFn: ({
      contractId,
      payload,
    }: {
      contractId: string;
      payload: UpdateContractRequest;
    }) => projectService.updateContract(projectId, contractId, payload),
    onSuccess: () => {
      setContractFormOpen(false);
      setEditingContract(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const deleteContractMutation = useMutation({
    mutationFn: (contractId: string) =>
      projectService.deleteContract(projectId, contractId),
    onSuccess: refreshProjectData,
  });

  const uploadDocumentMutation = useMutation({
    mutationFn: (payload: { file: File; name?: string; category?: string; version?: string }) =>
      projectService.uploadDocument(projectId, payload),
    onSuccess: () => {
      setDocumentFormOpen(false);
      setEditingDocument(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const updateDocumentMutation = useMutation({
    mutationFn: ({
      documentId,
      payload,
    }: {
      documentId: string;
      payload: UpdateDocumentRequest;
    }) => projectService.updateDocument(projectId, documentId, payload),
    onSuccess: () => {
      setDocumentFormOpen(false);
      setEditingDocument(null);
      setFormError(null);
      refreshProjectData();
    },
  });

  const createCAMutation = useMutation({
    mutationFn: (data: CreateCorrectiveActionRequest) =>
      projectService.createCorrectiveAction(projectId, data),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "corrective-actions"] });
      setCAFormOpen(false);
      setEditingCA(null);
    },
  });

  const updateCAMutation = useMutation({
    mutationFn: ({ caId, payload }: { caId: string; payload: UpdateCorrectiveActionRequest }) =>
      projectService.updateCorrectiveAction(projectId, caId, payload),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "corrective-actions"] });
      setCAFormOpen(false);
      setEditingCA(null);
    },
  });

  const deleteCAMutation = useMutation({
    mutationFn: (caId: string) => projectService.deleteCorrectiveAction(projectId, caId),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "corrective-actions"] }),
  });

  const transitionCAMutation = useMutation({
    mutationFn: ({ caId, toStatus }: { caId: string; toStatus: string }) =>
      projectService.transitionCorrectiveAction(projectId, caId, toStatus),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ["projects", projectId, "corrective-actions"] }),
  });

  const deleteDocumentMutation = useMutation({
    mutationFn: (documentId: string) =>
      projectService.deleteDocument(projectId, documentId),
    onSuccess: refreshProjectData,
  });

  function submitTask(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);
    const parsed = taskSchema.safeParse(readTaskForm(event.currentTarget));
    if (!parsed.success) {
      setFormError(parsed.error.issues[0]?.message ?? "Invalid task data.");
      return;
    }

    if (editingTask) {
      updateTaskMutation.mutate({
        taskId: editingTask.id,
        payload: toTaskUpdatePayload(parsed.data),
      });
      return;
    }

    createTaskMutation.mutate(toTaskCreatePayload(parsed.data));
  }

  function submitMilestone(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);
    const parsed = milestoneSchema.safeParse(readMilestoneForm(event.currentTarget));
    if (!parsed.success) {
      setFormError(parsed.error.issues[0]?.message ?? "Invalid milestone data.");
      return;
    }

    if (editingMilestone) {
      updateMilestoneMutation.mutate({
        milestoneId: editingMilestone.id,
        payload: toMilestoneUpdatePayload(parsed.data),
      });
      return;
    }

    createMilestoneMutation.mutate(toMilestoneCreatePayload(parsed.data));
  }

  function submitIssue(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);
    const parsed = issueSchema.safeParse(readIssueForm(event.currentTarget));
    if (!parsed.success) {
      setFormError(parsed.error.issues[0]?.message ?? "Invalid issue data.");
      return;
    }

    if (editingIssue) {
      updateIssueMutation.mutate({
        issueId: editingIssue.id,
        payload: toIssueUpdatePayload(parsed.data),
      });
      return;
    }

    createIssueMutation.mutate(toIssueCreatePayload(parsed.data));
  }

  function submitRisk(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);
    const parsed = riskSchema.safeParse(readRiskForm(event.currentTarget));
    if (!parsed.success) {
      setFormError(parsed.error.issues[0]?.message ?? "Invalid risk data.");
      return;
    }

    if (editingRisk) {
      updateRiskMutation.mutate({
        riskId: editingRisk.id,
        payload: toRiskUpdatePayload(parsed.data),
      });
      return;
    }

    createRiskMutation.mutate(toRiskCreatePayload(parsed.data));
  }

  function submitBudget(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);
    const parsed = budgetSchema.safeParse(readBudgetForm(event.currentTarget));
    if (!parsed.success) {
      setFormError(parsed.error.issues[0]?.message ?? "Invalid budget data.");
      return;
    }

    if (editingBudget) {
      updateBudgetMutation.mutate({
        budgetId: editingBudget.id,
        payload: toBudgetUpdatePayload(parsed.data),
      });
      return;
    }

    createBudgetMutation.mutate(toBudgetCreatePayload(parsed.data));
  }

  function submitContract(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);
    const parsed = contractSchema.safeParse(readContractForm(event.currentTarget));
    if (!parsed.success) {
      setFormError(parsed.error.issues[0]?.message ?? "Invalid contract data.");
      return;
    }

    if (editingContract) {
      updateContractMutation.mutate({
        contractId: editingContract.id,
        payload: toContractUpdatePayload(parsed.data),
      });
      return;
    }

    createContractMutation.mutate(toContractCreatePayload(parsed.data));
  }

  function submitVendor(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);
    const parsed = vendorSchema.safeParse(readVendorForm(event.currentTarget));
    if (!parsed.success) {
      setFormError(parsed.error.issues[0]?.message ?? "Invalid vendor data.");
      return;
    }

    createVendorMutation.mutate(toVendorCreatePayload(parsed.data, vendorFormType));
  }

  function submitDocument(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);
    const parsed = documentSchema.safeParse(readDocumentForm(event.currentTarget));
    if (!parsed.success) {
      setFormError(parsed.error.issues[0]?.message ?? "Invalid document data.");
      return;
    }

    if (editingDocument) {
      updateDocumentMutation.mutate({
        documentId: editingDocument.id,
        payload: toDocumentUpdatePayload(parsed.data),
      });
      return;
    }

    if (!parsed.data.file) {
      setFormError("A file is required.");
      return;
    }

    uploadDocumentMutation.mutate({
      file: parsed.data.file,
      name: optionalString(parsed.data.name),
      category: optionalString(parsed.data.category),
      version: optionalString(parsed.data.version),
    });
  }

  function openTaskForm(task: Task | null = null) {
    setEditingTask(task);
    setTaskFormOpen(true);
    setMilestoneFormOpen(false);
    setFormError(null);
  }

  function openMilestoneForm(milestone: Milestone | null = null) {
    setEditingMilestone(milestone);
    setMilestoneFormOpen(true);
    setTaskFormOpen(false);
    setIssueFormOpen(false);
    setFormError(null);
  }

  function openIssueForm(issue: Issue | null = null) {
    setEditingIssue(issue);
    setIssueFormOpen(true);
    setTaskFormOpen(false);
    setMilestoneFormOpen(false);
    setFormError(null);
  }

  function openRiskForm(risk: Risk | null = null) {
    setEditingRisk(risk);
    setRiskFormOpen(true);
    setTaskFormOpen(false);
    setMilestoneFormOpen(false);
    setIssueFormOpen(false);
    setFormError(null);
  }

  function openBudgetForm(budget: BudgetLine | null = null) {
    setEditingBudget(budget);
    setBudgetFormOpen(true);
    setTaskFormOpen(false);
    setMilestoneFormOpen(false);
    setIssueFormOpen(false);
    setRiskFormOpen(false);
    setFormError(null);
  }

  function openContractForm(contract: Contract | null = null) {
    setEditingContract(contract);
    setContractFormOpen(true);
    setTaskFormOpen(false);
    setMilestoneFormOpen(false);
    setIssueFormOpen(false);
    setRiskFormOpen(false);
    setBudgetFormOpen(false);
    setFormError(null);
  }

  function openVendorForm(type: VendorType) {
    setEditingVendor(null);
    setVendorFormOpen(true);
    setFormError(null);
    setVendorFormType(type);
  }

  function openDocumentForm(document: ProjectDocument | null = null) {
    setEditingDocument(document);
    setDocumentFormOpen(true);
    setTaskFormOpen(false);
    setMilestoneFormOpen(false);
    setIssueFormOpen(false);
    setRiskFormOpen(false);
    setBudgetFormOpen(false);
    setFormError(null);
  }

  function openCAForm(ca: CorrectiveAction | null = null) {
    setEditingCA(ca);
    setCAFormOpen(true);
    setFormError(null);
  }

  function handleCASubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const title = getString(data, "title").trim();
    const deviation = getString(data, "deviation").trim();
    if (!title || !deviation) {
      setFormError("Title dan Deviation wajib diisi.");
      return;
    }
    setFormError(null);
    const payload = {
      title,
      deviation,
      root_cause: optionalString(getString(data, "root_cause")),
      recommendation: optionalString(getString(data, "recommendation")),
      pic_user_id: optionalString(getString(data, "pic_user_id")) ?? null,
      target_date: optionalString(getString(data, "target_date")) ?? null,
      source_type: optionalString(getString(data, "source_type")),
      evidence_note: optionalString(getString(data, "evidence_note")),
    };
    if (editingCA) {
      updateCAMutation.mutate({ caId: editingCA.id, payload });
    } else {
      createCAMutation.mutate(payload as CreateCorrectiveActionRequest);
    }
  }

  return (
    <DashboardLayout title="Project Detail">
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <Link
            href="/projects"
            className="mb-3 inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft className="h-4 w-4" aria-hidden="true" />
            Projects
          </Link>
          {projectQuery.isLoading ? (
            <div className="h-7 w-64 rounded-md bg-muted" />
          ) : projectQuery.isError ? (
            <h2 className="text-lg font-semibold text-destructive">Project not available</h2>
          ) : (
            <ProjectHeader project={projectQuery.data} />
          )}
        </div>
        <div className="flex gap-2">
          <button type="button" onClick={() => openIssueForm()} className={secondaryButton}>
            <Flag className="h-4 w-4" aria-hidden="true" />
            Issue
          </button>
          <button type="button" onClick={() => openRiskForm()} className={secondaryButton}>
            <ShieldAlert className="h-4 w-4" aria-hidden="true" />
            Risk
          </button>
          <button type="button" onClick={() => openBudgetForm()} className={secondaryButton}>
            <Banknote className="h-4 w-4" aria-hidden="true" />
            Budget
          </button>
          <button type="button" onClick={() => openContractForm()} className={secondaryButton}>
            <FileText className="h-4 w-4" aria-hidden="true" />
            Kontrak
          </button>
          <button type="button" onClick={() => openDocumentForm()} className={secondaryButton}>
            <Download className="h-4 w-4" aria-hidden="true" />
            Dokumen
          </button>
          <button type="button" onClick={() => openMilestoneForm()} className={secondaryButton}>
            <MilestoneIcon className="h-4 w-4" aria-hidden="true" />
            Milestone
          </button>
          <button type="button" onClick={() => openTaskForm()} className={primaryButton}>
            <Plus className="h-4 w-4" aria-hidden="true" />
            Task
          </button>
        </div>
      </div>

      {formError && (
        <div className="mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {formError}
        </div>
      )}

      <FieldInspectionPanel projectId={projectId} />
      <HealthScorePanel projectId={projectId} />
      <ProjectControlPanel projectId={projectId} />
      <PeriodicReportPanel projectId={projectId} />

      {taskFormOpen && (
        <TaskForm
          task={editingTask}
          isSaving={createTaskMutation.isPending || updateTaskMutation.isPending}
          onSubmit={submitTask}
          onCancel={() => {
            setTaskFormOpen(false);
            setEditingTask(null);
            setFormError(null);
          }}
        />
      )}

      {milestoneFormOpen && (
        <MilestoneForm
          milestone={editingMilestone}
          isSaving={createMilestoneMutation.isPending || updateMilestoneMutation.isPending}
          onSubmit={submitMilestone}
          onCancel={() => {
            setMilestoneFormOpen(false);
            setEditingMilestone(null);
            setFormError(null);
          }}
        />
      )}

      {issueFormOpen && (
        <IssueForm
          issue={editingIssue}
          isSaving={createIssueMutation.isPending || updateIssueMutation.isPending}
          onSubmit={submitIssue}
          onCancel={() => {
            setIssueFormOpen(false);
            setEditingIssue(null);
            setFormError(null);
          }}
        />
      )}

      {riskFormOpen && (
        <RiskForm
          risk={editingRisk}
          isSaving={createRiskMutation.isPending || updateRiskMutation.isPending}
          onSubmit={submitRisk}
          onCancel={() => {
            setRiskFormOpen(false);
            setEditingRisk(null);
            setFormError(null);
          }}
        />
      )}

      {budgetFormOpen && (
        <BudgetForm
          budget={editingBudget}
          isSaving={createBudgetMutation.isPending || updateBudgetMutation.isPending}
          onSubmit={submitBudget}
          onCancel={() => {
            setBudgetFormOpen(false);
            setEditingBudget(null);
            setFormError(null);
          }}
        />
      )}

      {vendorFormOpen && (
        <VendorForm
          type={vendorFormType}
          isSaving={createVendorMutation.isPending}
          onSubmit={submitVendor}
          onCancel={() => {
            setVendorFormOpen(false);
            setEditingVendor(null);
            setFormError(null);
          }}
        />
      )}

      {contractFormOpen && (
        <ContractForm
          contract={editingContract}
          vendorOptions={vendorOptions}
          consultantOptions={consultantOptions}
          vendorsLoading={vendorsQuery.isLoading}
          isSaving={createContractMutation.isPending || updateContractMutation.isPending}
          onCreateVendor={openVendorForm}
          onSubmit={submitContract}
          onCancel={() => {
            setContractFormOpen(false);
            setEditingContract(null);
            setFormError(null);
          }}
        />
      )}

      {documentFormOpen && (
        <DocumentForm
          document={editingDocument}
          isSaving={uploadDocumentMutation.isPending || updateDocumentMutation.isPending}
          onSubmit={submitDocument}
          onCancel={() => {
            setDocumentFormOpen(false);
            setEditingDocument(null);
            setFormError(null);
          }}
        />
      )}

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.35fr)_minmax(360px,0.65fr)]">
        <section className="space-y-6">
          <KanbanBoard
            groups={taskGroups}
            isLoading={tasksQuery.isLoading}
            isError={tasksQuery.isError}
            onCreate={() => openTaskForm()}
            onEdit={openTaskForm}
            onMove={(task, status) =>
              updateTaskMutation.mutate({
                taskId: task.id,
                payload: { status },
              })
            }
            onDelete={(task) => {
              if (window.confirm(`Hapus task "${task.title}"?`)) {
                deleteTaskMutation.mutate(task.id);
              }
            }}
          />

          <TaskList
            tasks={tasks}
            isLoading={tasksQuery.isLoading}
            isError={tasksQuery.isError}
            onEdit={openTaskForm}
            onMove={(task, status) =>
              updateTaskMutation.mutate({ taskId: task.id, payload: { status } })
            }
            onDelete={(task) => {
              if (window.confirm(`Hapus task "${task.title}"?`)) {
                deleteTaskMutation.mutate(task.id);
              }
            }}
          />
        </section>

        <MilestonePanel
          milestones={milestones}
          isLoading={milestonesQuery.isLoading}
          isError={milestonesQuery.isError}
          onCreate={() => openMilestoneForm()}
          onEdit={openMilestoneForm}
          onMove={(milestone, status) =>
            updateMilestoneMutation.mutate({
              milestoneId: milestone.id,
              payload: { status },
            })
          }
          onDelete={(milestone) => {
            if (window.confirm(`Hapus milestone "${milestone.title}"?`)) {
              deleteMilestoneMutation.mutate(milestone.id);
            }
          }}
        />

        <GanttPanel
          tasks={tasks}
          milestones={milestones}
          isLoading={tasksQuery.isLoading || milestonesQuery.isLoading}
          isError={tasksQuery.isError || milestonesQuery.isError}
        />

        <IssuePanel
          issues={issues}
          isLoading={issuesQuery.isLoading}
          isError={issuesQuery.isError}
          onCreate={() => openIssueForm()}
          onEdit={openIssueForm}
          onMove={(issue, status) =>
            updateIssueMutation.mutate({ issueId: issue.id, payload: { status } })
          }
          onDelete={(issue) => {
            if (window.confirm(`Hapus isu "${issue.title}"?`)) {
              deleteIssueMutation.mutate(issue.id);
            }
          }}
        />

        <RiskPanel
          risks={risks}
          isLoading={risksQuery.isLoading}
          isError={risksQuery.isError}
          onCreate={() => openRiskForm()}
          onEdit={openRiskForm}
          onMove={(risk, status) =>
            updateRiskMutation.mutate({ riskId: risk.id, payload: { status } })
          }
          onDelete={(risk) => {
            if (window.confirm(`Hapus risiko "${risk.title}"?`)) {
              deleteRiskMutation.mutate(risk.id);
            }
          }}
        />

        <BudgetPanel
          budgets={budgets}
          totals={budgetTotals}
          isLoading={budgetsQuery.isLoading}
          isError={budgetsQuery.isError}
          onCreate={() => openBudgetForm()}
          onEdit={openBudgetForm}
          onDelete={(budget) => {
            if (window.confirm(`Hapus baris anggaran "${budget.category}"?`)) {
              deleteBudgetMutation.mutate(budget.id);
            }
          }}
        />

        <ContractPanel
          contracts={contracts}
          totals={contractTotals}
          isLoading={contractsQuery.isLoading}
          isError={contractsQuery.isError}
          onCreate={() => openContractForm()}
          onEdit={openContractForm}
          onDelete={(contract) => {
            if (window.confirm(`Hapus kontrak "${contract.title}"?`)) {
              deleteContractMutation.mutate(contract.id);
            }
          }}
        />

        <DocumentPanel
          documents={documents}
          totals={documentTotals}
          isLoading={documentsQuery.isLoading}
          isError={documentsQuery.isError}
          onCreate={() => openDocumentForm()}
          onEdit={openDocumentForm}
          onDownload={async (document) => {
            try {
              const { blob, filename } = await projectService.downloadDocument(
                projectId,
                document.id
              );
              const url = URL.createObjectURL(blob);
              const anchor = window.document.createElement("a");
              anchor.href = url;
              anchor.download = filename;
              window.document.body.appendChild(anchor);
              anchor.click();
              anchor.remove();
              URL.revokeObjectURL(url);
            } catch {
              window.alert("Dokumen belum dapat diunduh.");
            }
          }}
          onDelete={(document) => {
            if (window.confirm(`Hapus dokumen "${document.name}"?`)) {
              deleteDocumentMutation.mutate(document.id);
            }
          }}
        />

        <CorrectiveActionPanel
          correctiveActions={correctiveActions}
          isLoading={correctiveActionsQuery.isLoading}
          isError={correctiveActionsQuery.isError}
          onCreate={() => openCAForm()}
          onEdit={openCAForm}
          onMove={(ca, toStatus) =>
            transitionCAMutation.mutate({ caId: ca.id, toStatus })
          }
          onDelete={(ca) => {
            if (window.confirm(`Hapus tindak lanjut "${ca.title}"?`)) {
              deleteCAMutation.mutate(ca.id);
            }
          }}
        />

        {caFormOpen && (
          <CorrectiveActionForm
            ca={editingCA}
            isSaving={createCAMutation.isPending || updateCAMutation.isPending}
            formError={formError}
            onSubmit={handleCASubmit}
            onClose={() => {
              setCAFormOpen(false);
              setEditingCA(null);
              setFormError(null);
            }}
          />
        )}

        {/* ── Monitoring ──────────────────────────────────────────────── */}
        <MonitoringPanel
          baselines={baselines}
          snapshots={snapshots}
          isLoadingBaselines={baselinesQuery.isLoading}
          isLoadingSnapshots={snapshotsQuery.isLoading}
          isErrorBaselines={baselinesQuery.isError}
          isErrorSnapshots={snapshotsQuery.isError}
          statusFilter={snapshotStatusFilter}
          onStatusFilterChange={setSnapshotStatusFilter}
          onCreateBaseline={() => {
            setEditingBaseline(null);
            setFormError(null);
            setBaselineFormOpen(true);
          }}
          onEditBaseline={(b) => {
            setEditingBaseline(b);
            setFormError(null);
            setBaselineFormOpen(true);
          }}
          onDeleteBaseline={(b) => {
            if (window.confirm(`Hapus baseline "${b.label ?? `v${b.version}`}"?`)) {
              deleteBaselineMutation.mutate(b.id);
            }
          }}
          onCreateSnapshot={() => {
            setEditingSnapshot(null);
            setFormError(null);
            setSnapshotFormOpen(true);
          }}
          onEditSnapshot={(s) => {
            setEditingSnapshot(s);
            setFormError(null);
            setSnapshotFormOpen(true);
          }}
          onTransitionSnapshot={(s, status) =>
            transitionSnapshotMutation.mutate({ id: s.id, payload: { status } })
          }
          onRejectSnapshot={(s, reason) =>
            transitionSnapshotMutation.mutate({
              id: s.id,
              payload: { status: "REJECTED", rejection_reason: reason },
            })
          }
          onDeleteSnapshot={(s) => {
            if (window.confirm(`Hapus snapshot ${MONTH_NAMES[s.period_month]} ${s.period_year}?`)) {
              deleteSnapshotMutation.mutate(s.id);
            }
          }}
        />

        {baselineFormOpen && (
          <BaselineFormModal
            baseline={editingBaseline}
            isSaving={createBaselineMutation.isPending || updateBaselineMutation.isPending}
            formError={formError}
            onSubmit={(payload) => {
              if (editingBaseline) {
                updateBaselineMutation.mutate({ id: editingBaseline.id, payload });
              } else {
                createBaselineMutation.mutate(payload as CreateBaselineRequest);
              }
            }}
            onClose={() => {
              setBaselineFormOpen(false);
              setEditingBaseline(null);
              setFormError(null);
            }}
          />
        )}

        {snapshotFormOpen && (
          <SnapshotFormModal
            snapshot={editingSnapshot}
            baselines={baselines}
            isSaving={createSnapshotMutation.isPending || updateSnapshotMutation.isPending}
            formError={formError}
            onSubmit={(payload) => {
              if (editingSnapshot) {
                updateSnapshotMutation.mutate({ id: editingSnapshot.id, payload: payload as UpdateSnapshotRequest });
              } else {
                createSnapshotMutation.mutate(payload as CreateSnapshotRequest);
              }
            }}
            onClose={() => {
              setSnapshotFormOpen(false);
              setEditingSnapshot(null);
              setFormError(null);
            }}
          />
        )}
      </div>
    </DashboardLayout>
  );
}

function ProjectHeader({ project }: { project: Project | undefined }) {
  if (!project) return null;

  return (
    <div>
      <p className="text-xs font-medium uppercase text-muted-foreground">{project.code}</p>
      <h2 className="text-xl font-semibold text-foreground">{project.name}</h2>
      <p className="mt-1 text-sm text-muted-foreground">
        {project.status.replace("_", " ")} · {project.progress_pct}% progress ·{" "}
        {formatDate(project.end_date)}
      </p>
    </div>
  );
}

function KanbanBoard({
  groups,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onMove,
  onDelete,
}: {
  groups: Array<{ status: TaskStatus; tasks: Task[] }>;
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (task: Task) => void;
  onMove: (task: Task, status: TaskStatus) => void;
  onDelete: (task: Task) => void;
}) {
  return (
    <section className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Papan Kanban" actionLabel="Task Baru" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat papan kerja..." />
      ) : isError ? (
        <ErrorState label="Papan kerja belum dapat dimuat." />
      ) : (
        <div className="grid grid-cols-1 gap-3 p-4 md:grid-cols-2 2xl:grid-cols-5">
          {groups.map((group) => (
            <div key={group.status} className="rounded-md border border-border bg-muted/30">
              <div className="border-b border-border px-3 py-2">
                <p className="text-xs font-semibold text-foreground">
                  {group.status.replace("_", " ")}
                </p>
                <p className="text-xs text-muted-foreground">
                  {group.tasks.length} task{group.tasks.length !== 1 ? "s" : ""}
                </p>
              </div>
              <div className="space-y-3 p-3">
                {group.tasks.length === 0 ? (
                  <p className="rounded-md bg-background px-3 py-4 text-center text-xs text-muted-foreground">
                    Empty
                  </p>
                ) : (
                  group.tasks.map((task) => (
                    <TaskCard
                      key={task.id}
                      task={task}
                      onEdit={onEdit}
                      onMove={onMove}
                      onDelete={onDelete}
                    />
                  ))
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

function TaskCard({
  task,
  onEdit,
  onMove,
  onDelete,
}: {
  task: Task;
  onEdit: (task: Task) => void;
  onMove: (task: Task, status: TaskStatus) => void;
  onDelete: (task: Task) => void;
}) {
  return (
    <article className="rounded-md border border-border bg-background p-3 shadow-sm">
      <div className="flex items-start justify-between gap-2">
        <div>
          <h4 className="text-sm font-medium text-foreground">{task.title}</h4>
          <p className="mt-1 text-xs text-muted-foreground">{task.priority}</p>
        </div>
        <RowActions onEdit={() => onEdit(task)} onDelete={() => onDelete(task)} />
      </div>
      <div className="mt-3">
        <ProgressBar value={task.progress_pct} />
      </div>
      <div className="mt-3 flex flex-wrap gap-2">
        {TASK_NEXT_STATUS[task.status].map((status) => (
          <button
            key={status}
            type="button"
            onClick={() => onMove(task, status)}
            className="inline-flex h-7 items-center rounded-md border border-input px-2 text-xs hover:bg-accent"
          >
            {status.replace("_", " ")}
          </button>
        ))}
      </div>
    </article>
  );
}

function TaskList({
  tasks,
  isLoading,
  isError,
  onEdit,
  onMove,
  onDelete,
}: {
  tasks: Task[];
  isLoading: boolean;
  isError: boolean;
  onEdit: (task: Task) => void;
  onMove: (task: Task, status: TaskStatus) => void;
  onDelete: (task: Task) => void;
}) {
  return (
    <section className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Task List" />
      {isLoading ? (
        <LoadingState label="Memuat task..." />
      ) : isError ? (
        <ErrorState label="Task belum dapat dimuat." />
      ) : tasks.length === 0 ? (
        <EmptyState label="Belum ada task." />
      ) : (
        <div className="divide-y divide-border">
          {tasks.map((task) => (
            <div key={task.id} className="p-4">
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <p className="text-sm font-medium text-foreground">{task.title}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {task.status.replace("_", " ")} · due {formatDate(task.due_date)}
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  {TASK_NEXT_STATUS[task.status].map((status) => (
                    <button
                      key={status}
                      type="button"
                      onClick={() => onMove(task, status)}
                      className="inline-flex h-8 items-center rounded-md border border-input px-2 text-xs hover:bg-accent"
                    >
                      {status.replace("_", " ")}
                    </button>
                  ))}
                  <RowActions onEdit={() => onEdit(task)} onDelete={() => onDelete(task)} />
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

function MilestonePanel({
  milestones,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onMove,
  onDelete,
}: {
  milestones: Milestone[];
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (milestone: Milestone) => void;
  onMove: (milestone: Milestone, status: MilestoneStatus) => void;
  onDelete: (milestone: Milestone) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Milestone" actionLabel="Milestone Baru" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat milestone..." />
      ) : isError ? (
        <ErrorState label="Milestone belum dapat dimuat." />
      ) : milestones.length === 0 ? (
        <EmptyState label="Belum ada milestone." />
      ) : (
        <div className="divide-y divide-border">
          {milestones.map((milestone) => (
            <div key={milestone.id} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-sm font-medium text-foreground">{milestone.title}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {milestone.status.replace("_", " ")} · due {formatDate(milestone.due_date)}
                  </p>
                </div>
                <RowActions
                  onEdit={() => onEdit(milestone)}
                  onDelete={() => onDelete(milestone)}
                />
              </div>
              <div className="mt-3">
                <ProgressBar value={milestone.progress_pct} />
              </div>
              <div className="mt-3 flex flex-wrap gap-2">
                {MILESTONE_NEXT_STATUS[milestone.status].map((status) => (
                  <button
                    key={status}
                    type="button"
                    onClick={() => onMove(milestone, status)}
                    className="inline-flex h-7 items-center rounded-md border border-input px-2 text-xs hover:bg-accent"
                  >
                    {status.replace("_", " ")}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </aside>
  );
}

function IssuePanel({
  issues,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onMove,
  onDelete,
}: {
  issues: Issue[];
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (issue: Issue) => void;
  onMove: (issue: Issue, status: IssueStatus) => void;
  onDelete: (issue: Issue) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Isu" actionLabel="Isu Baru" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat isu..." />
      ) : isError ? (
        <ErrorState label="Isu belum dapat dimuat." />
      ) : issues.length === 0 ? (
        <EmptyState label="Belum ada isu." />
      ) : (
        <div className="divide-y divide-border">
          {issues.map((issue) => (
            <div key={issue.id} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-sm font-medium text-foreground">{issue.title}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {issue.status.replace("_", " ")} · {issue.severity} ·{" "}
                    {issue.escalation.replace("_", " ")}
                    {issue.due_date ? ` · due ${formatDate(issue.due_date)}` : ""}
                  </p>
                </div>
                <RowActions
                  onEdit={() => onEdit(issue)}
                  onDelete={() => onDelete(issue)}
                />
              </div>
              {issue.description && (
                <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">
                  {issue.description}
                </p>
              )}
              {issue.resolution && (
                <p className="mt-2 text-xs text-muted-foreground">
                  <span className="font-medium text-foreground">Resolution:</span>{" "}
                  {issue.resolution}
                </p>
              )}
              <div className="mt-3 flex flex-wrap gap-2">
                {ISSUE_NEXT_STATUS[issue.status].map((status) => (
                  <button
                    key={status}
                    type="button"
                    onClick={() => onMove(issue, status)}
                    className="inline-flex h-7 items-center rounded-md border border-input px-2 text-xs hover:bg-accent"
                  >
                    {status.replace("_", " ")}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </aside>
  );
}

function RiskPanel({
  risks,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onMove,
  onDelete,
}: {
  risks: Risk[];
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (risk: Risk) => void;
  onMove: (risk: Risk, status: RiskStatus) => void;
  onDelete: (risk: Risk) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Risiko" actionLabel="Risiko Baru" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat risiko..." />
      ) : isError ? (
        <ErrorState label="Risiko belum dapat dimuat." />
      ) : risks.length === 0 ? (
        <EmptyState label="Belum ada risiko." />
      ) : (
        <div className="divide-y divide-border">
          {risks.map((risk) => (
            <div key={risk.id} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-sm font-medium text-foreground">{risk.title}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {risk.status.replace("_", " ")} ·{" "}
                    <span className={riskSeverityColor(risk.severity)}>
                      {risk.severity}
                    </span>{" "}
                    · score {risk.risk_score} ({risk.probability}×{risk.impact})
                    {risk.due_date ? ` · due ${formatDate(risk.due_date)}` : ""}
                  </p>
                </div>
                <RowActions
                  onEdit={() => onEdit(risk)}
                  onDelete={() => onDelete(risk)}
                />
              </div>
              {risk.description && (
                <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">
                  {risk.description}
                </p>
              )}
              {risk.mitigation && (
                <p className="mt-2 text-xs text-muted-foreground">
                  <span className="font-medium text-foreground">Mitigation:</span>{" "}
                  {risk.mitigation}
                </p>
              )}
              <div className="mt-3 flex flex-wrap gap-2">
                {RISK_NEXT_STATUS[risk.status].map((status) => (
                  <button
                    key={status}
                    type="button"
                    onClick={() => onMove(risk, status)}
                    className="inline-flex h-7 items-center rounded-md border border-input px-2 text-xs hover:bg-accent"
                  >
                    {status.replace("_", " ")}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </aside>
  );
}

function TaskForm({
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

function MilestoneForm({
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

function IssueForm({
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

function RiskForm({
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

function BudgetPanel({
  budgets,
  totals,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onDelete,
}: {
  budgets: BudgetLine[];
  totals: { planned: number; actual: number; variance: number; usage: number };
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (budget: BudgetLine) => void;
  onDelete: (budget: BudgetLine) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Anggaran" actionLabel="Baris Anggaran Baru" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat anggaran..." />
      ) : isError ? (
        <ErrorState label="Anggaran belum dapat dimuat." />
      ) : budgets.length === 0 ? (
        <EmptyState label="Belum ada baris anggaran." />
      ) : (
        <div>
          <div className="grid grid-cols-2 gap-2 border-b border-border p-4 text-xs">
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Total Planned</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {formatCurrency(totals.planned)}
              </p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Total Actual</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {formatCurrency(totals.actual)}
              </p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Variance</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {formatCurrency(totals.variance)}
              </p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Usage</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {totals.usage}%
              </p>
            </div>
          </div>
          <div className="divide-y divide-border">
            {budgets.map((budget) => (
              <div key={budget.id} className="p-4">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-sm font-medium text-foreground">{budget.category}</p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Planned {formatCurrency(budget.planned)} · Actual{" "}
                      {formatCurrency(budget.actual)} · Variance{" "}
                      {formatCurrency(budget.variance)} · Usage {budget.usage_pct}%
                    </p>
                  </div>
                  <div className="flex flex-col items-end gap-2">
                    <span
                      className={cn(
                        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold",
                        budgetStatusTone(budget.status)
                      )}
                    >
                      {budget.status.replace("_", " ")}
                    </span>
                    <RowActions
                      onEdit={() => onEdit(budget)}
                      onDelete={() => onDelete(budget)}
                    />
                  </div>
                </div>
                {budget.description && (
                  <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">
                    {budget.description}
                  </p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
      <p className="border-t border-border px-4 py-2 text-[10px] text-muted-foreground">
        Operational input budget/realisasi — belum validated/published (menunggu P1-011/P1-012).
      </p>
    </aside>
  );
}

function BudgetForm({
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

function VendorForm({
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

function ContractForm({
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
      </div>
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

function ContractPanel({
  contracts,
  totals,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onDelete,
}: {
  contracts: Contract[];
  totals: {
    totalValue: number;
    activeCount: number;
    mainVendorId: string | null;
    mainValue: number;
    period: string;
  };
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (contract: Contract) => void;
  onDelete: (contract: Contract) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Kontrak" actionLabel="Kontrak Baru" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat kontrak..." />
      ) : isError ? (
        <ErrorState label="Kontrak belum dapat dimuat." />
      ) : contracts.length === 0 ? (
        <EmptyState label="Belum ada kontrak." />
      ) : (
        <div>
          <div className="grid grid-cols-2 gap-2 border-b border-border p-4 text-xs">
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Total Value</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {formatCurrency(totals.totalValue)}
              </p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Kontrak Aktif</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {totals.activeCount}
              </p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Contract Period</p>
              <p className="mt-0.5 font-semibold text-foreground">{totals.period}</p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Main Vendor Value</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {formatCurrency(totals.mainValue)}
              </p>
            </div>
          </div>
          <div className="divide-y divide-border">
            {contracts.map((contract) => (
              <div key={contract.id} className="p-4">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-sm font-medium text-foreground">{contract.title}</p>
                    <p className="mt-0.5 text-xs text-muted-foreground">
                      {contract.contract_number}
                    </p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {contract.vendor?.name ?? ""}
                      {contract.consultant ? ` · Konsultan: ${contract.consultant.name}` : ""}
                      {" · "}
                      {formatCurrency(contract.contract_value)}
                      {" · "}
                      {contract.start_date
                        ? `${toDateInput(contract.start_date)} — ${contract.end_date ? toDateInput(contract.end_date) : "?"}`
                        : "TBD"}
                    </p>
                  </div>
                  <div className="flex flex-col items-end gap-2">
                    <span
                      className={cn(
                        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold",
                        contractStatusTone(contract.status)
                      )}
                    >
                      {contract.status.replace("_", " ")}
                    </span>
                    <RowActions
                      onEdit={() => onEdit(contract)}
                      onDelete={() => onDelete(contract)}
                    />
                  </div>
                </div>
                {contract.scope_of_work && (
                  <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">
                    {contract.scope_of_work}
                  </p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
      <p className="border-t border-border px-4 py-2 text-[10px] text-muted-foreground">
        Operational contract input — belum validated/published (menunggu P1-011/P1-012).
      </p>
    </aside>
  );
}

function DocumentPanel({
  documents,
  totals,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onDownload,
  onDelete,
}: {
  documents: ProjectDocument[];
  totals: { count: number; totalSize: number; byCategory: Record<string, number> };
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (document: ProjectDocument) => void;
  onDownload: (document: ProjectDocument) => void;
  onDelete: (document: ProjectDocument) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Dokumen" actionLabel="Upload" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat dokumen..." />
      ) : isError ? (
        <ErrorState label="Dokumen belum dapat dimuat." />
      ) : documents.length === 0 ? (
        <EmptyState label="Belum ada dokumen." />
      ) : (
        <div>
          <div className="grid grid-cols-2 gap-2 border-b border-border p-4 text-xs">
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Total Files</p>
              <p className="mt-0.5 font-semibold text-foreground">{totals.count}</p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Total Size</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {formatFileSize(totals.totalSize)}
              </p>
            </div>
          </div>
          <div className="divide-y divide-border">
            {documents.map((doc) => (
              <div key={doc.id} className="p-4">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium text-foreground">{doc.name}</p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {formatFileSize(doc.file_size ?? 0)}
                      {doc.category ? ` · ${doc.category.replace("_", " ")}` : ""}
                      {doc.version ? ` · ${doc.version}` : ""}
                      {" · "}
                      {formatDate(doc.created_at)}
                    </p>
                  </div>
                  <div className="flex flex-col items-end gap-2">
                    <span
                      className={cn(
                        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold",
                        documentCategoryTone(doc.category)
                      )}
                    >
                      {(doc.category ?? "OTHER").replace("_", " ")}
                    </span>
                    <div className="flex gap-1">
                      <button
                        type="button"
                        onClick={() => onDownload(doc)}
                        className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        aria-label="Download"
                      >
                        <Download className="h-4 w-4" aria-hidden="true" />
                      </button>
                      <button
                        type="button"
                        onClick={() => onEdit(doc)}
                        className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        aria-label="Edit"
                      >
                        <Edit3 className="h-4 w-4" aria-hidden="true" />
                      </button>
                      <button
                        type="button"
                        onClick={() => onDelete(doc)}
                        className="rounded-md p-2 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                        aria-label="Delete"
                      >
                        <Trash2 className="h-4 w-4" aria-hidden="true" />
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
      <p className="border-t border-border px-4 py-2 text-[10px] text-muted-foreground">
        File tersimpan lokal (storage/documents). Operational evidence input — belum
        validated/published (menunggu P1-012).
      </p>
    </aside>
  );
}

function DocumentForm({
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

function SectionHeader({
  title,
  actionLabel,
  onAction,
}: {
  title: string;
  actionLabel?: string;
  onAction?: () => void;
}) {
  return (
    <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-3">
      <h3 className="text-sm font-semibold text-foreground">{title}</h3>
      {actionLabel && onAction && (
        <button type="button" onClick={onAction} className={secondaryButton}>
          <Plus className="h-4 w-4" aria-hidden="true" />
          {actionLabel}
        </button>
      )}
    </div>
  );
}

function FormHeader({ title, onCancel }: { title: string; onCancel: () => void }) {
  return (
    <div className="mb-4 flex items-center justify-between gap-3">
      <h3 className="text-sm font-semibold text-foreground">{title}</h3>
      <button
        type="button"
        onClick={onCancel}
        className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
        aria-label="Close form"
      >
        <X className="h-4 w-4" aria-hidden="true" />
      </button>
    </div>
  );
}

function FormActions({ isSaving, onCancel }: { isSaving: boolean; onCancel: () => void }) {
  return (
    <div className="mt-4 flex justify-end gap-2">
      <button
        type="button"
        onClick={onCancel}
        className="inline-flex h-9 items-center rounded-md border border-input px-3 text-sm hover:bg-accent"
      >
        Batal
      </button>
      <button type="submit" disabled={isSaving} className={primaryButton}>
        <Save className="h-4 w-4" aria-hidden="true" />
        {isSaving ? "Menyimpan..." : "Simpan"}
      </button>
    </div>
  );
}

function RowActions({ onEdit, onDelete }: { onEdit: () => void; onDelete: () => void }) {
  return (
    <div className="flex gap-1">
      <button
        type="button"
        onClick={onEdit}
        className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
        aria-label="Edit"
      >
        <Edit3 className="h-4 w-4" aria-hidden="true" />
      </button>
      <button
        type="button"
        onClick={onDelete}
        className="rounded-md p-2 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
        aria-label="Delete"
      >
        <Trash2 className="h-4 w-4" aria-hidden="true" />
      </button>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="mb-1 block text-sm font-medium text-foreground">{label}</span>
      {children}
    </label>
  );
}

function ProgressBar({ value }: { value: number }) {
  return (
    <div>
      <div className="mb-1 flex items-center justify-between text-xs text-muted-foreground">
        <span>Progress</span>
        <span>{value}%</span>
      </div>
      <div className="h-1.5 overflow-hidden rounded-full bg-muted">
        <div className="h-full rounded-full bg-primary" style={{ width: `${value}%` }} />
      </div>
    </div>
  );
}

function EmptyState({ label }: { label: string }) {
  return (
    <div className="flex items-start gap-3 p-6 text-sm text-muted-foreground">
      <CheckCircle2 className="mt-0.5 h-4 w-4 text-green-700" aria-hidden="true" />
      {label}
    </div>
  );
}

function ErrorState({ label }: { label: string }) {
  return (
    <div className="flex items-start gap-3 p-6 text-sm text-destructive">
      <AlertCircle className="mt-0.5 h-4 w-4" aria-hidden="true" />
      {label}
    </div>
  );
}

function LoadingState({ label }: { label: string }) {
  return (
    <div className="flex items-center gap-3 p-6 text-sm text-muted-foreground">
      <CalendarDays className="h-4 w-4" aria-hidden="true" />
      {label}
    </div>
  );
}

function readTaskForm(form: HTMLFormElement): TaskFormValues {
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

function readMilestoneForm(form: HTMLFormElement): MilestoneFormValues {
  const data = new FormData(form);
  return {
    title: getString(data, "title"),
    description: getString(data, "description"),
    due_date: getString(data, "due_date"),
    progress_pct: Number(getString(data, "progress_pct") || 0),
  };
}

function toTaskCreatePayload(values: TaskFormValues): CreateTaskRequest {
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

function toTaskUpdatePayload(values: TaskFormValues): UpdateTaskRequest {
  return {
    ...toTaskCreatePayload(values),
    progress_pct: values.progress_pct,
  };
}

function toMilestoneCreatePayload(values: MilestoneFormValues): CreateMilestoneRequest {
  return {
    title: values.title,
    description: optionalString(values.description),
    due_date: optionalString(values.due_date),
  };
}

function toMilestoneUpdatePayload(values: MilestoneFormValues): UpdateMilestoneRequest {
  return {
    ...toMilestoneCreatePayload(values),
    progress_pct: values.progress_pct,
  };
}

function readIssueForm(form: HTMLFormElement): IssueFormValues {
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

function toIssueCreatePayload(values: IssueFormValues): CreateIssueRequest {
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

function toIssueUpdatePayload(values: IssueFormValues): UpdateIssueRequest {
  return {
    ...toIssueCreatePayload(values),
  };
}

function readRiskForm(form: HTMLFormElement): RiskFormValues {
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

function toRiskCreatePayload(values: RiskFormValues): CreateRiskRequest {
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

function toRiskUpdatePayload(values: RiskFormValues): UpdateRiskRequest {
  return {
    ...toRiskCreatePayload(values),
  };
}

function riskSeverityColor(severity: RiskSeverity) {
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

function readBudgetForm(form: HTMLFormElement): BudgetFormValues {
  const data = new FormData(form);
  return {
    category: getString(data, "category"),
    description: getString(data, "description"),
    planned: Number(getString(data, "planned") || 0),
    actual: Number(getString(data, "actual") || 0),
    currency: getString(data, "currency"),
  };
}

function toBudgetCreatePayload(values: BudgetFormValues): CreateBudgetRequest {
  return {
    category: values.category,
    description: optionalString(values.description),
    planned: values.planned,
    actual: values.actual,
    currency: optionalString(values.currency) ?? "IDR",
  };
}

function toBudgetUpdatePayload(values: BudgetFormValues): UpdateBudgetRequest {
  return {
    ...toBudgetCreatePayload(values),
  };
}

function readContractForm(form: HTMLFormElement): ContractFormValues {
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
  };
}

function toContractCreatePayload(values: ContractFormValues): CreateContractRequest {
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
  };
}

function toContractUpdatePayload(values: ContractFormValues): UpdateContractRequest {
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
  };
}

function readVendorForm(form: HTMLFormElement): VendorFormValues {
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

function toVendorCreatePayload(
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

function readDocumentForm(form: HTMLFormElement): DocumentFormValues {
  const data = new FormData(form);
  const file = data.get("file");
  return {
    name: getString(data, "name"),
    category: getString(data, "category"),
    version: getString(data, "version"),
    file: file instanceof File && file.size > 0 ? file : undefined,
  };
}

function toDocumentUpdatePayload(
  values: DocumentFormValues
): UpdateDocumentRequest {
  return {
    name: optionalString(values.name),
    category: optionalString(values.category),
    version: optionalString(values.version),
  };
}

function documentCategoryTone(category: string | null | undefined) {
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

function formatFileSize(bytes: number) {
  if (!bytes) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  const i = Math.min(
    Math.floor(Math.log(bytes) / Math.log(1024)),
    units.length - 1
  );
  const value = bytes / Math.pow(1024, i);
  return `${value.toFixed(value >= 100 || i === 0 ? 0 : 1)} ${units[i]}`;
}

function contractStatusTone(status: ContractStatus) {
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

function budgetStatusTone(status: BudgetStatus) {
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

function formatCurrency(value: number) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(value);
}

function getString(data: FormData, key: string) {
  const value = data.get(key);
  return typeof value === "string" ? value : "";
}

function optionalString(value: string | undefined) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

function toDateInput(value: string | null | undefined) {
  if (!value) return "";
  return value.slice(0, 10);
}

const inputClassName = cn(
  "h-9 w-full rounded-md border border-input bg-background px-3 text-sm",
  "placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
);

const primaryButton = cn(
  "inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground",
  "hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
);

const secondaryButton = cn(
  "inline-flex h-9 items-center justify-center gap-2 rounded-md border border-input px-3 text-sm font-medium",
  "hover:bg-accent hover:text-accent-foreground"
);

// ---------------------------------------------------------------------------
// Corrective Action Panel (P1-006)
// ---------------------------------------------------------------------------

function caStatusTone(status: CorrectiveActionStatus): string {
  switch (status) {
    case "DRAFT":       return "bg-gray-100 text-gray-700";
    case "SUBMITTED":   return "bg-blue-100 text-blue-700";
    case "IN_PROGRESS": return "bg-yellow-100 text-yellow-800";
    case "COMPLETED":   return "bg-green-100 text-green-700";
    case "REJECTED":    return "bg-red-100 text-red-700";
    default:            return "bg-gray-100 text-gray-700";
  }
}

function CorrectiveActionPanel({
  correctiveActions,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onMove,
  onDelete,
}: {
  correctiveActions: CorrectiveAction[];
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (ca: CorrectiveAction) => void;
  onMove: (ca: CorrectiveAction, toStatus: CorrectiveActionStatus) => void;
  onDelete: (ca: CorrectiveAction) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader
        title="Corrective Actions"
        actionLabel="Tindak Lanjut Baru"
        onAction={onCreate}
      />
      {isLoading ? (
        <LoadingState label="Memuat tindak lanjut..." />
      ) : isError ? (
        <ErrorState label="Tindak lanjut belum dapat dimuat." />
      ) : correctiveActions.length === 0 ? (
        <EmptyState label="Belum ada tindak lanjut." />
      ) : (
        <div className="divide-y divide-border">
          {correctiveActions.map((ca) => (
            <div key={ca.id} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="text-sm font-medium text-foreground">{ca.title}</p>
                    <span
                      className={cn(
                        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
                        caStatusTone(ca.status as CorrectiveActionStatus)
                      )}
                    >
                      {ca.status.replace("_", " ")}
                    </span>
                    {ca.source_type && (
                      <span className="inline-flex items-center rounded-full bg-purple-100 px-2 py-0.5 text-xs font-medium text-purple-700">
                        {ca.source_type}
                      </span>
                    )}
                  </div>
                  <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">
                    {ca.deviation}
                  </p>
                  {ca.root_cause && (
                    <p className="mt-1 text-xs text-muted-foreground">
                      <span className="font-medium text-foreground">Root cause:</span>{" "}
                      {ca.root_cause}
                    </p>
                  )}
                  {ca.target_date && (
                    <p className="mt-1 text-xs text-muted-foreground">
                      Target: {formatDate(ca.target_date)}
                    </p>
                  )}
                </div>
                <RowActions
                  onEdit={() => onEdit(ca)}
                  onDelete={() => onDelete(ca)}
                />
              </div>
              <div className="mt-3 flex flex-wrap gap-2">
                {(CA_NEXT_STATUS[ca.status as CorrectiveActionStatus] ?? []).map((toStatus) => (
                  <button
                    key={toStatus}
                    type="button"
                    onClick={() => onMove(ca, toStatus)}
                    className="inline-flex h-7 items-center rounded-md border border-input px-2 text-xs hover:bg-accent"
                  >
                    {toStatus.replace("_", " ")}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </aside>
  );
}

// ---------------------------------------------------------------------------
// Corrective Action Form (modal)
// ---------------------------------------------------------------------------

function CorrectiveActionForm({
  ca,
  isSaving,
  formError,
  onSubmit,
  onClose,
}: {
  ca: CorrectiveAction | null;
  isSaving: boolean;
  formError: string | null;
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void;
  onClose: () => void;
}) {
  const isEdit = ca !== null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-lg rounded-xl border border-border bg-background shadow-xl">
        <div className="flex items-center justify-between border-b border-border px-6 py-4">
          <h3 className="text-base font-semibold">
            {isEdit ? "Edit Tindak Lanjut" : "Tindak Lanjut Baru"}
          </h3>
          <button
            type="button"
            onClick={onClose}
            className="text-muted-foreground hover:text-foreground"
            aria-label="Close"
          >
            ✕
          </button>
        </div>

        <form onSubmit={onSubmit} className="space-y-4 px-6 py-5">
          {formError && (
            <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {formError}
            </p>
          )}

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="ca-title">
              Title <span className="text-destructive">*</span>
            </label>
            <input
              id="ca-title"
              name="title"
              className={inputClassName}
              defaultValue={ca?.title ?? ""}
              placeholder="Judul singkat temuan"
              required
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="ca-deviation">
              Deviation <span className="text-destructive">*</span>
            </label>
            <textarea
              id="ca-deviation"
              name="deviation"
              rows={3}
              className={cn(inputClassName, "h-auto resize-none")}
              defaultValue={ca?.deviation ?? ""}
              placeholder="Jelaskan deviasi atau ketidaksesuaian"
              required
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="ca-root-cause">
              Root Cause
            </label>
            <textarea
              id="ca-root-cause"
              name="root_cause"
              rows={2}
              className={cn(inputClassName, "h-auto resize-none")}
              defaultValue={ca?.root_cause ?? ""}
              placeholder="Analysis of the root cause"
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="ca-recommendation">
              Recommendation
            </label>
            <textarea
              id="ca-recommendation"
              name="recommendation"
              rows={2}
              className={cn(inputClassName, "h-auto resize-none")}
              defaultValue={ca?.recommendation ?? ""}
              placeholder="Rekomendasi tindak lanjut"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="ca-source-type">
                Source Type
              </label>
              <select
                id="ca-source-type"
                name="source_type"
                className={inputClassName}
                defaultValue={ca?.source_type ?? ""}
              >
                <option value="">— none —</option>
                <option value="issue">Issue</option>
                <option value="risk">Risk</option>
                <option value="task">Task</option>
              </select>
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="ca-target-date">
                Target Date
              </label>
              <input
                id="ca-target-date"
                name="target_date"
                type="date"
                className={inputClassName}
                defaultValue={toDateInput(ca?.target_date)}
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="ca-evidence">
              Evidence Note
            </label>
            <textarea
              id="ca-evidence"
              name="evidence_note"
              rows={2}
              className={cn(inputClassName, "h-auto resize-none")}
              defaultValue={ca?.evidence_note ?? ""}
              placeholder="Evidence atau referensi dokumen pendukung"
            />
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className={secondaryButton}>
              Batal
            </button>
            <button type="submit" disabled={isSaving} className={primaryButton}>
              {isSaving ? "Menyimpan..." : isEdit ? "Perbarui" : "Tambah"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// MONITORING — MonitoringPanel
// ─────────────────────────────────────────────────────────────────────────────

function MonitoringPanel({
  baselines,
  snapshots,
  isLoadingBaselines,
  isLoadingSnapshots,
  isErrorBaselines,
  isErrorSnapshots,
  statusFilter,
  onStatusFilterChange,
  onCreateBaseline,
  onEditBaseline,
  onDeleteBaseline,
  onCreateSnapshot,
  onEditSnapshot,
  onTransitionSnapshot,
  onRejectSnapshot,
  onDeleteSnapshot,
}: {
  baselines: Baseline[];
  snapshots: Snapshot[];
  isLoadingBaselines: boolean;
  isLoadingSnapshots: boolean;
  isErrorBaselines: boolean;
  isErrorSnapshots: boolean;
  statusFilter: SnapshotStatus | "";
  onStatusFilterChange: (v: SnapshotStatus | "") => void;
  onCreateBaseline: () => void;
  onEditBaseline: (b: Baseline) => void;
  onDeleteBaseline: (b: Baseline) => void;
  onCreateSnapshot: () => void;
  onEditSnapshot: (s: Snapshot) => void;
  onTransitionSnapshot: (s: Snapshot, status: SnapshotStatus) => void;
  onRejectSnapshot: (s: Snapshot, reason: string) => void;
  onDeleteSnapshot: (s: Snapshot) => void;
}) {
  const filteredSnapshots =
    statusFilter
      ? snapshots.filter((s) => s.status === statusFilter)
      : snapshots;

  const activeBaseline = baselines.find((b) => b.is_active);

  return (
    <section className="space-y-6">
      {/* ── Header ── */}
      <div className="flex items-center justify-between">
        <h3 className="text-base font-semibold text-foreground">Monitoring</h3>
      </div>

      {/* ── Baselines ── */}
      <div className="rounded-lg border border-border bg-card p-5 space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-semibold text-foreground">Baseline</p>
            <p className="text-xs text-muted-foreground mt-0.5">
              Rencana resmi proyek yang digunakan sebagai acuan snapshot
            </p>
          </div>
          <button onClick={onCreateBaseline} className={primaryButton}>
            <Plus className="h-3.5 w-3.5" />
            Tambah Baseline
          </button>
        </div>

        {isErrorBaselines && (
          <p className="text-sm text-destructive">Gagal memuat baseline.</p>
        )}

        {isLoadingBaselines ? (
          <p className="text-sm text-muted-foreground">Memuat…</p>
        ) : baselines.length === 0 ? (
          <p className="text-sm text-muted-foreground italic">
            Belum ada baseline. Buat baseline pertama untuk mulai monitoring.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left text-xs text-muted-foreground">
                  <th className="pb-2 pr-4 font-medium">Ver.</th>
                  <th className="pb-2 pr-4 font-medium">Label</th>
                  <th className="pb-2 pr-4 font-medium">Mulai Rencana</th>
                  <th className="pb-2 pr-4 font-medium">Selesai Rencana</th>
                  <th className="pb-2 pr-4 font-medium">Target Fisik (%)</th>
                  <th className="pb-2 pr-4 font-medium">Anggaran</th>
                  <th className="pb-2 pr-4 font-medium">Status</th>
                  <th className="pb-2 font-medium" />
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {baselines.map((b) => (
                  <tr key={b.id} className="hover:bg-muted/30">
                    <td className="py-2.5 pr-4 font-mono text-xs">v{b.version}</td>
                    <td className="py-2.5 pr-4">{b.label ?? "—"}</td>
                    <td className="py-2.5 pr-4">{formatDate(b.planned_start)}</td>
                    <td className="py-2.5 pr-4">{formatDate(b.planned_end)}</td>
                    <td className="py-2.5 pr-4">{b.physical_target}%</td>
                    <td className="py-2.5 pr-4">
                      {b.budget_total.toLocaleString("id-ID")} {b.currency}
                    </td>
                    <td className="py-2.5 pr-4">
                      {b.is_active ? (
                        <span className="inline-flex items-center gap-1 rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
                          <CheckCircle2 className="h-3 w-3" /> Aktif
                        </span>
                      ) : (
                        <span className="inline-flex items-center rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
                          Tidak aktif
                        </span>
                      )}
                    </td>
                    <td className="py-2.5">
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => onEditBaseline(b)}
                          className="text-muted-foreground hover:text-foreground"
                          title="Edit"
                        >
                          <Edit3 className="h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => onDeleteBaseline(b)}
                          className="text-muted-foreground hover:text-destructive"
                          title="Hapus"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* ── Snapshots ── */}
      <div className="rounded-lg border border-border bg-card p-5 space-y-4">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          <div>
            <p className="text-sm font-semibold text-foreground">Snapshot Periodik</p>
            <p className="text-xs text-muted-foreground mt-0.5">
              Realisasi fisik & keuangan per periode terhadap baseline
              {activeBaseline ? ` (Baseline: v${activeBaseline.version}${activeBaseline.label ? ` — ${activeBaseline.label}` : ""})` : ""}
            </p>
          </div>
          <div className="flex items-center gap-3">
            <select
              value={statusFilter}
              onChange={(e) => onStatusFilterChange(e.target.value as SnapshotStatus | "")}
              className="rounded-md border border-input bg-background px-3 py-1.5 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="">Semua Status</option>
              {(["DRAFT", "SUBMITTED", "VALID", "REJECTED", "STALE"] as SnapshotStatus[]).map((s) => (
                <option key={s} value={s}>{SNAPSHOT_STATUS_LABEL[s]}</option>
              ))}
            </select>
            <button onClick={onCreateSnapshot} className={primaryButton}>
              <Plus className="h-3.5 w-3.5" />
              Tambah Snapshot
            </button>
          </div>
        </div>

        {isErrorSnapshots && (
          <p className="text-sm text-destructive">Gagal memuat snapshot.</p>
        )}

        {isLoadingSnapshots ? (
          <p className="text-sm text-muted-foreground">Memuat…</p>
        ) : filteredSnapshots.length === 0 ? (
          <p className="text-sm text-muted-foreground italic">
            {statusFilter ? `Tidak ada snapshot dengan status ${SNAPSHOT_STATUS_LABEL[statusFilter]}.` : "Belum ada snapshot periodik."}
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left text-xs text-muted-foreground">
                  <th className="pb-2 pr-3 font-medium">Periode</th>
                  <th className="pb-2 pr-3 font-medium">Fisik Rencana</th>
                  <th className="pb-2 pr-3 font-medium">Fisik Aktual</th>
                  <th className="pb-2 pr-3 font-medium">Deviasi Fisik</th>
                  <th className="pb-2 pr-3 font-medium">Keuangan Rencana</th>
                  <th className="pb-2 pr-3 font-medium">Keuangan Aktual</th>
                  <th className="pb-2 pr-3 font-medium">Deviasi Hari</th>
                  <th className="pb-2 pr-3 font-medium">Status</th>
                  <th className="pb-2 font-medium" />
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filteredSnapshots.map((s) => (
                  <tr key={s.id} className="hover:bg-muted/30">
                    <td className="py-2.5 pr-3 font-medium whitespace-nowrap">
                      {MONTH_NAMES[s.period_month]} {s.period_year}
                    </td>
                    <td className="py-2.5 pr-3">{s.physical_target}%</td>
                    <td className="py-2.5 pr-3">{s.physical_actual}%</td>
                    <td className={cn("py-2.5 pr-3 font-medium", s.physical_variance >= 0 ? "text-green-600" : "text-destructive")}>
                      {s.physical_variance >= 0 ? "+" : ""}{s.physical_variance.toFixed(2)}%
                    </td>
                    <td className="py-2.5 pr-3">
                      {s.financial_target.toLocaleString("id-ID")}
                    </td>
                    <td className="py-2.5 pr-3">
                      {s.financial_actual.toLocaleString("id-ID")}
                    </td>
                    <td className={cn("py-2.5 pr-3", (s.schedule_deviation_days ?? 0) > 0 ? "text-destructive" : "text-green-600")}>
                      {s.schedule_deviation_days != null
                        ? `${s.schedule_deviation_days > 0 ? "+" : ""}${s.schedule_deviation_days}h`
                        : "—"}
                    </td>
                    <td className="py-2.5 pr-3">
                      <span className={cn("inline-flex rounded-full px-2 py-0.5 text-xs font-medium", SNAPSHOT_STATUS_COLOR[s.status])}>
                        {SNAPSHOT_STATUS_LABEL[s.status]}
                      </span>
                    </td>
                    <td className="py-2.5">
                      <SnapshotActions
                        snapshot={s}
                        onEdit={onEditSnapshot}
                        onTransition={onTransitionSnapshot}
                        onReject={onRejectSnapshot}
                        onDelete={onDeleteSnapshot}
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  );
}

function SnapshotActions({
  snapshot,
  onEdit,
  onTransition,
  onReject,
  onDelete,
}: {
  snapshot: Snapshot;
  onEdit: (s: Snapshot) => void;
  onTransition: (s: Snapshot, status: SnapshotStatus) => void;
  onReject: (s: Snapshot, reason: string) => void;
  onDelete: (s: Snapshot) => void;
}) {
  const s = snapshot;
  return (
    <div className="flex items-center gap-1.5 flex-wrap">
      {s.status === "DRAFT" && (
        <>
          <button
            onClick={() => onEdit(s)}
            className="text-xs text-muted-foreground hover:text-foreground"
            title="Edit"
          >
            <Edit3 className="h-3.5 w-3.5" />
          </button>
          <button
            onClick={() => onTransition(s, "SUBMITTED")}
            className="rounded px-2 py-0.5 text-xs bg-blue-100 text-blue-700 hover:bg-blue-200"
          >
            Ajukan
          </button>
          <button
            onClick={() => onDelete(s)}
            className="text-muted-foreground hover:text-destructive"
            title="Hapus"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        </>
      )}
      {s.status === "SUBMITTED" && (
        <>
          <button
            onClick={() => onTransition(s, "VALID")}
            className="rounded px-2 py-0.5 text-xs bg-green-100 text-green-700 hover:bg-green-200"
          >
            Validasi
          </button>
          <button
            onClick={() => {
              const reason = window.prompt("Alasan penolakan:");
              if (reason !== null) onReject(s, reason);
            }}
            className="rounded px-2 py-0.5 text-xs bg-red-100 text-red-700 hover:bg-red-200"
          >
            Tolak
          </button>
        </>
      )}
      {(s.status === "VALID" || s.status === "REJECTED") && (
        <button
          onClick={() => onTransition(s, "STALE")}
          className="rounded px-2 py-0.5 text-xs bg-yellow-100 text-yellow-700 hover:bg-yellow-200"
        >
          Kedaluwarsa
        </button>
      )}
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// MONITORING — BaselineFormModal
// ─────────────────────────────────────────────────────────────────────────────

function BaselineFormModal({
  baseline,
  isSaving,
  formError,
  onSubmit,
  onClose,
}: {
  baseline: Baseline | null;
  isSaving: boolean;
  formError: string | null;
  onSubmit: (payload: CreateBaselineRequest | UpdateBaselineRequest) => void;
  onClose: () => void;
}) {
  const isEdit = baseline !== null;

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const payload = {
      label: (fd.get("label") as string) || undefined,
      physical_target: parseFloat(fd.get("physical_target") as string),
      budget_total: parseFloat(fd.get("budget_total") as string),
      currency: (fd.get("currency") as string) || "IDR",
      planned_start: fd.get("planned_start") as string,
      planned_end: fd.get("planned_end") as string,
      source: (fd.get("source") as string) || undefined,
      notes: (fd.get("notes") as string) || undefined,
      ...(isEdit ? { is_active: (fd.get("is_active") as string) === "true" } : {}),
    };
    onSubmit(payload);
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-lg rounded-xl border border-border bg-card shadow-xl">
        <div className="flex items-center justify-between border-b border-border px-6 py-4">
          <h3 className="text-base font-semibold">
            {isEdit ? "Edit Baseline" : "Tambah Baseline"}
          </h3>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground" aria-label="Close">
            <X className="h-4 w-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 px-6 py-5">
          {formError && (
            <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{formError}</p>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-label">Label</label>
              <input
                id="bl-label"
                name="label"
                className={inputClassName}
                defaultValue={baseline?.label ?? ""}
                placeholder="misal: Baseline Kontrak"
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-currency">Mata Uang</label>
              <input
                id="bl-currency"
                name="currency"
                className={inputClassName}
                defaultValue={baseline?.currency ?? "IDR"}
                placeholder="IDR"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-physical-target">
                Target Fisik (%) <span className="text-destructive">*</span>
              </label>
              <input
                id="bl-physical-target"
                name="physical_target"
                type="number"
                min={0}
                max={100}
                step={0.01}
                className={inputClassName}
                defaultValue={baseline?.physical_target ?? 100}
                required
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-budget-total">
                Anggaran Total <span className="text-destructive">*</span>
              </label>
              <input
                id="bl-budget-total"
                name="budget_total"
                type="number"
                min={0}
                step={1}
                className={inputClassName}
                defaultValue={baseline?.budget_total ?? ""}
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-planned-start">
                Mulai Rencana <span className="text-destructive">*</span>
              </label>
              <input
                id="bl-planned-start"
                name="planned_start"
                type="date"
                className={inputClassName}
                defaultValue={baseline?.planned_start?.slice(0, 10) ?? ""}
                required
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-planned-end">
                Selesai Rencana <span className="text-destructive">*</span>
              </label>
              <input
                id="bl-planned-end"
                name="planned_end"
                type="date"
                className={inputClassName}
                defaultValue={baseline?.planned_end?.slice(0, 10) ?? ""}
                required
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="bl-source">Sumber</label>
            <input
              id="bl-source"
              name="source"
              className={inputClassName}
              defaultValue={baseline?.source ?? ""}
              placeholder="misal: Dokumen Kontrak No. 001"
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="bl-notes">Catatan</label>
            <textarea
              id="bl-notes"
              name="notes"
              rows={2}
              className={cn(inputClassName, "h-auto resize-none")}
              defaultValue={baseline?.notes ?? ""}
            />
          </div>

          {isEdit && (
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="bl-is-active">Status Aktif</label>
              <select
                id="bl-is-active"
                name="is_active"
                className={inputClassName}
                defaultValue={String(baseline?.is_active ?? false)}
              >
                <option value="true">Aktif</option>
                <option value="false">Tidak aktif</option>
              </select>
            </div>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className={secondaryButton}>Batal</button>
            <button type="submit" disabled={isSaving} className={primaryButton}>
              {isSaving ? "Menyimpan…" : isEdit ? "Perbarui" : "Simpan"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// MONITORING — SnapshotFormModal
// ─────────────────────────────────────────────────────────────────────────────

function SnapshotFormModal({
  snapshot,
  baselines,
  isSaving,
  formError,
  onSubmit,
  onClose,
}: {
  snapshot: Snapshot | null;
  baselines: Baseline[];
  isSaving: boolean;
  formError: string | null;
  onSubmit: (payload: CreateSnapshotRequest | UpdateSnapshotRequest) => void;
  onClose: () => void;
}) {
  const isEdit = snapshot !== null;
  const activeBaseline = baselines.find((b) => b.is_active);

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    if (isEdit) {
      const payload: UpdateSnapshotRequest = {
        physical_actual: parseFloat(fd.get("physical_actual") as string),
        physical_target: parseFloat(fd.get("physical_target") as string),
        financial_actual: parseFloat(fd.get("financial_actual") as string),
        financial_target: parseFloat(fd.get("financial_target") as string),
        schedule_deviation_days: fd.get("schedule_deviation_days")
          ? parseInt(fd.get("schedule_deviation_days") as string)
          : undefined,
        source: (fd.get("source") as string) || undefined,
        notes: (fd.get("notes") as string) || undefined,
      };
      onSubmit(payload);
    } else {
      const payload: CreateSnapshotRequest = {
        baseline_id: (fd.get("baseline_id") as string) || undefined,
        period_year: parseInt(fd.get("period_year") as string),
        period_month: parseInt(fd.get("period_month") as string),
        physical_actual: parseFloat(fd.get("physical_actual") as string),
        physical_target: parseFloat(fd.get("physical_target") as string),
        financial_actual: parseFloat(fd.get("financial_actual") as string),
        financial_target: parseFloat(fd.get("financial_target") as string),
        currency: (fd.get("currency") as string) || "IDR",
        schedule_deviation_days: fd.get("schedule_deviation_days")
          ? parseInt(fd.get("schedule_deviation_days") as string)
          : undefined,
        source: (fd.get("source") as string) || undefined,
        notes: (fd.get("notes") as string) || undefined,
      };
      onSubmit(payload);
    }
  }

  const currentYear = new Date().getFullYear();

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-lg rounded-xl border border-border bg-card shadow-xl overflow-y-auto max-h-[90vh]">
        <div className="flex items-center justify-between border-b border-border px-6 py-4">
          <h3 className="text-base font-semibold">
            {isEdit ? "Edit Snapshot" : "Tambah Snapshot Periodik"}
          </h3>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground" aria-label="Close">
            <X className="h-4 w-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 px-6 py-5">
          {formError && (
            <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{formError}</p>
          )}

          {!isEdit && (
            <>
              <div className="space-y-1">
                <label className="text-sm font-medium" htmlFor="sn-baseline">Baseline</label>
                <select
                  id="sn-baseline"
                  name="baseline_id"
                  className={inputClassName}
                  defaultValue={activeBaseline?.id ?? ""}
                >
                  <option value="">— Tanpa baseline —</option>
                  {baselines.map((b) => (
                    <option key={b.id} value={b.id}>
                      v{b.version}{b.label ? ` — ${b.label}` : ""}{b.is_active ? " (aktif)" : ""}
                    </option>
                  ))}
                </select>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <label className="text-sm font-medium" htmlFor="sn-year">
                    Tahun <span className="text-destructive">*</span>
                  </label>
                  <input
                    id="sn-year"
                    name="period_year"
                    type="number"
                    min={2000}
                    max={2100}
                    className={inputClassName}
                    defaultValue={currentYear}
                    required
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium" htmlFor="sn-month">
                    Bulan <span className="text-destructive">*</span>
                  </label>
                  <select
                    id="sn-month"
                    name="period_month"
                    className={inputClassName}
                    defaultValue={new Date().getMonth() + 1}
                    required
                  >
                    {MONTH_NAMES.slice(1).map((name, i) => (
                      <option key={i + 1} value={i + 1}>{name}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="space-y-1">
                <label className="text-sm font-medium" htmlFor="sn-currency">Mata Uang</label>
                <input
                  id="sn-currency"
                  name="currency"
                  className={inputClassName}
                  defaultValue="IDR"
                  placeholder="IDR"
                />
              </div>
            </>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="sn-physical-target">
                Target Fisik (%) <span className="text-destructive">*</span>
              </label>
              <input
                id="sn-physical-target"
                name="physical_target"
                type="number"
                min={0}
                max={100}
                step={0.01}
                className={inputClassName}
                defaultValue={snapshot?.physical_target ?? ""}
                required
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="sn-physical-actual">
                Aktual Fisik (%) <span className="text-destructive">*</span>
              </label>
              <input
                id="sn-physical-actual"
                name="physical_actual"
                type="number"
                min={0}
                max={100}
                step={0.01}
                className={inputClassName}
                defaultValue={snapshot?.physical_actual ?? ""}
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="sn-fin-target">
                Target Keuangan <span className="text-destructive">*</span>
              </label>
              <input
                id="sn-fin-target"
                name="financial_target"
                type="number"
                min={0}
                step={1}
                className={inputClassName}
                defaultValue={snapshot?.financial_target ?? ""}
                required
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium" htmlFor="sn-fin-actual">
                Aktual Keuangan <span className="text-destructive">*</span>
              </label>
              <input
                id="sn-fin-actual"
                name="financial_actual"
                type="number"
                min={0}
                step={1}
                className={inputClassName}
                defaultValue={snapshot?.financial_actual ?? ""}
                required
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="sn-sched-dev">
              Deviasi Jadwal (hari)
            </label>
            <input
              id="sn-sched-dev"
              name="schedule_deviation_days"
              type="number"
              className={inputClassName}
              defaultValue={snapshot?.schedule_deviation_days ?? ""}
              placeholder="0 = sesuai jadwal, positif = terlambat"
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="sn-source">Sumber Data</label>
            <input
              id="sn-source"
              name="source"
              className={inputClassName}
              defaultValue={snapshot?.source ?? ""}
              placeholder="misal: Laporan Bulanan"
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-medium" htmlFor="sn-notes">Catatan</label>
            <textarea
              id="sn-notes"
              name="notes"
              rows={2}
              className={cn(inputClassName, "h-auto resize-none")}
              defaultValue={snapshot?.notes ?? ""}
            />
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className={secondaryButton}>Batal</button>
            <button type="submit" disabled={isSaving} className={primaryButton}>
              {isSaving ? "Menyimpan…" : isEdit ? "Perbarui" : "Simpan"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
