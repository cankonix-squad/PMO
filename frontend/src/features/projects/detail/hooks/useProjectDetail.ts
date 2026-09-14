"use client";
import { useBaselineMutations } from "./useBaselineMutations";
import { useBudgetMutations } from "./useBudgetMutations";
import { useCAMutations } from "./useCAMutations";
import { useContractMutations } from "./useContractMutations";
import { useDocumentMutations } from "./useDocumentMutations";
import { useIssueMutations } from "./useIssueMutations";
import { useMilestoneMutations } from "./useMilestoneMutations";
import { useProjectQueries } from "./useProjectQueries";
import { useRiskMutations } from "./useRiskMutations";
import { useSnapshotMutations } from "./useSnapshotMutations";
import { useTaskMutations } from "./useTaskMutations";
import { useVendorMutations } from "./useVendorMutations";

import type { Baseline, Snapshot, SnapshotStatus } from "@/types/monitoring";
import type {
  BudgetLine,
  Contract,
  CorrectiveAction,
  CreateCorrectiveActionRequest,
  Issue,
  Milestone,
  ProjectDocument,
  Risk,
  Task,
  Vendor,
  VendorType,
} from "@/types/project";
import { useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import type { FormEvent } from "react";
import { useState } from "react";
import {
  budgetSchema,
  contractSchema,
  documentSchema,
  issueSchema,
  milestoneSchema,
  riskSchema,
  taskSchema,
  vendorSchema,
} from "../schemas";
import {
  getString,
  optionalString,
  readBudgetForm,
  readContractForm,
  readDocumentForm,
  readIssueForm,
  readMilestoneForm,
  readRiskForm,
  readTaskForm,
  readVendorForm,
  toBudgetCreatePayload,
  toBudgetUpdatePayload,
  toContractCreatePayload,
  toContractUpdatePayload,
  toDocumentUpdatePayload,
  toIssueCreatePayload,
  toIssueUpdatePayload,
  toMilestoneCreatePayload,
  toMilestoneUpdatePayload,
  toRiskCreatePayload,
  toRiskUpdatePayload,
  toTaskCreatePayload,
  toTaskUpdatePayload,
  toVendorCreatePayload,
} from "../utils";

export function useProjectDetail() {

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

  const {
    projectQuery,
    tasksQuery,
    milestonesQuery,
    issuesQuery,
    risksQuery,
    budgetsQuery,
    contractsQuery,
    documentsQuery,
    vendorsQuery,
    correctiveActionsQuery,
    tasks,
    milestones,
    issues,
    risks,
    budgets,
    contracts,
    documents,
    vendors,
    correctiveActions,
    vendorOptions,
    consultantOptions,
    contractTotals,
    budgetTotals,
    documentTotals,
    taskGroups,
    baselinesQuery,
    baselines,
    snapshotsQuery,
    snapshots,
  } = useProjectQueries({ projectId });

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

  // ── Monitoring mutations ──────────────────────────────────────────────────
  const { createBaselineMutation, updateBaselineMutation, deleteBaselineMutation } = useBaselineMutations({ projectId, setBaselineFormOpen, setEditingBaseline, setFormError, qc });

  const { createSnapshotMutation, updateSnapshotMutation, transitionSnapshotMutation, deleteSnapshotMutation } = useSnapshotMutations({ projectId, setSnapshotFormOpen, setEditingSnapshot, setFormError, qc });

  const { createTaskMutation, updateTaskMutation, deleteTaskMutation } = useTaskMutations({ projectId, setTaskFormOpen, setEditingTask, setFormError, refreshProjectData });

  const { createMilestoneMutation, updateMilestoneMutation, deleteMilestoneMutation } = useMilestoneMutations({ projectId, setMilestoneFormOpen, setEditingMilestone, setFormError, refreshProjectData });

  const { createIssueMutation, updateIssueMutation, deleteIssueMutation } = useIssueMutations({ projectId, setIssueFormOpen, setEditingIssue, setFormError, refreshProjectData });

  const { createRiskMutation, updateRiskMutation, deleteRiskMutation } = useRiskMutations({ projectId, setRiskFormOpen, setEditingRisk, setFormError, refreshProjectData });

  const { createBudgetMutation, updateBudgetMutation, deleteBudgetMutation } = useBudgetMutations({ projectId, setBudgetFormOpen, setEditingBudget, setFormError, refreshProjectData });

  const { createVendorMutation } = useVendorMutations({ setVendorFormOpen, setEditingVendor, setFormError, qc });

  const { createContractMutation, updateContractMutation, deleteContractMutation } = useContractMutations({ projectId, setContractFormOpen, setEditingContract, setFormError, refreshProjectData });

  const { uploadDocumentMutation, updateDocumentMutation, deleteDocumentMutation } = useDocumentMutations({ projectId, setDocumentFormOpen, setEditingDocument, setFormError, refreshProjectData });

  const { createCAMutation, updateCAMutation, deleteCAMutation, transitionCAMutation } = useCAMutations({ projectId, qc, setCAFormOpen, setEditingCA });

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
  return {
    projectQuery,
    openIssueForm,
    openRiskForm,
    openBudgetForm,
    openContractForm,
    openDocumentForm,
    openMilestoneForm,
    openTaskForm,
    formError,
    projectId,
    taskFormOpen,
    editingTask,
    createTaskMutation,
    updateTaskMutation,
    submitTask,
    setTaskFormOpen,
    setEditingTask,
    setFormError,
    milestoneFormOpen,
    editingMilestone,
    createMilestoneMutation,
    updateMilestoneMutation,
    submitMilestone,
    setMilestoneFormOpen,
    setEditingMilestone,
    issueFormOpen,
    editingIssue,
    createIssueMutation,
    updateIssueMutation,
    submitIssue,
    setIssueFormOpen,
    setEditingIssue,
    riskFormOpen,
    editingRisk,
    createRiskMutation,
    updateRiskMutation,
    submitRisk,
    setRiskFormOpen,
    setEditingRisk,
    budgetFormOpen,
    editingBudget,
    createBudgetMutation,
    updateBudgetMutation,
    submitBudget,
    setBudgetFormOpen,
    setEditingBudget,
    vendorFormOpen,
    vendorFormType,
    createVendorMutation,
    submitVendor,
    setVendorFormOpen,
    setEditingVendor,
    contractFormOpen,
    editingContract,
    vendorOptions,
    consultantOptions,
    vendorsQuery,
    createContractMutation,
    updateContractMutation,
    openVendorForm,
    submitContract,
    setContractFormOpen,
    setEditingContract,
    documentFormOpen,
    editingDocument,
    uploadDocumentMutation,
    updateDocumentMutation,
    submitDocument,
    setDocumentFormOpen,
    setEditingDocument,
    taskGroups,
    tasksQuery,
    deleteTaskMutation,
    tasks,
    milestones,
    milestonesQuery,
    deleteMilestoneMutation,
    issues,
    issuesQuery,
    deleteIssueMutation,
    risks,
    risksQuery,
    deleteRiskMutation,
    budgets,
    budgetTotals,
    budgetsQuery,
    deleteBudgetMutation,
    contracts,
    contractTotals,
    contractsQuery,
    deleteContractMutation,
    documents,
    documentTotals,
    documentsQuery,
    deleteDocumentMutation,
    correctiveActions,
    correctiveActionsQuery,
    openCAForm,
    transitionCAMutation,
    deleteCAMutation,
    caFormOpen,
    editingCA,
    createCAMutation,
    updateCAMutation,
    handleCASubmit,
    setCAFormOpen,
    setEditingCA,
    baselines,
    snapshots,
    baselinesQuery,
    snapshotsQuery,
    snapshotStatusFilter,
    setSnapshotStatusFilter,
    setEditingBaseline,
    setBaselineFormOpen,
    deleteBaselineMutation,
    setEditingSnapshot,
    setSnapshotFormOpen,
    transitionSnapshotMutation,
    deleteSnapshotMutation,
    baselineFormOpen,
    editingBaseline,
    createBaselineMutation,
    updateBaselineMutation,
    snapshotFormOpen,
    editingSnapshot,
    createSnapshotMutation,
    updateSnapshotMutation,
  };
}
