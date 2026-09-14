"use client";

import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { FieldInspectionPanel } from "@/features/projects/components/FieldInspectionPanel";
import { GanttPanel } from "@/features/projects/components/GanttPanel";
import { HealthScorePanel } from "@/features/projects/components/HealthScorePanel";
import { PeriodicReportPanel } from "@/features/projects/components/PeriodicReportPanel";
import { ProjectControlPanel } from "@/features/projects/components/ProjectControlPanel";
import { projectService } from "@/services/project.service";
import type { CreateBaselineRequest, CreateSnapshotRequest, UpdateSnapshotRequest } from "@/types/monitoring";
import { MONTH_NAMES } from "@/types/monitoring";
import { ArrowLeft, Banknote, Download, FileText, Flag, Milestone as MilestoneIcon, Plus, ShieldAlert } from "lucide-react";
import Link from "next/link";
import { BaselineFormModal } from "./components/BaselineFormModal";
import { BudgetForm } from "./components/BudgetForm";
import { BudgetPanel } from "./components/BudgetPanel";
import { ContractForm } from "./components/ContractForm";
import { ContractPanel } from "./components/ContractPanel";
import { CorrectiveActionForm } from "./components/CorrectiveActionForm";
import { CorrectiveActionPanel } from "./components/CorrectiveActionPanel";
import { DocumentForm } from "./components/DocumentForm";
import { DocumentPanel } from "./components/DocumentPanel";
import { IssueForm } from "./components/IssueForm";
import { IssuePanel } from "./components/IssuePanel";
import { KanbanBoard } from "./components/KanbanBoard";
import { MilestoneForm } from "./components/MilestoneForm";
import { MilestonePanel } from "./components/MilestonePanel";
import { MonitoringPanel } from "./components/MonitoringPanel";
import { ProjectHeader } from "./components/ProjectHeader";
import { RiskForm } from "./components/RiskForm";
import { RiskPanel } from "./components/RiskPanel";
import { SnapshotFormModal } from "./components/SnapshotFormModal";
import { TaskForm } from "./components/TaskForm";
import { TaskList } from "./components/TaskList";
import { VendorForm } from "./components/VendorForm";
import { useProjectDetail } from "./hooks/useProjectDetail";
import { primaryButton, secondaryButton } from "./schemas";

export default function ProjectDetailPage() {
  const {
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
  } = useProjectDetail();

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
