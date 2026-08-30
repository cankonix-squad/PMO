import api from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  Project,
  CreateProjectRequest,
  UpdateProjectRequest,
  TransitionProjectRequest,
  ProjectListFilter,
  ProgressHistory,
  Task,
  TaskListFilter,
  CreateTaskRequest,
  UpdateTaskRequest,
  Milestone,
  CreateMilestoneRequest,
  UpdateMilestoneRequest,
  Issue,
  IssueListFilter,
  CreateIssueRequest,
  UpdateIssueRequest,
  Risk,
  RiskListFilter,
  CreateRiskRequest,
  UpdateRiskRequest,
  BudgetLine,
  BudgetListFilter,
  CreateBudgetRequest,
  UpdateBudgetRequest,
  Vendor,
  VendorListFilter,
  CreateVendorRequest,
  UpdateVendorRequest,
  Contract,
  ContractListFilter,
  CreateContractRequest,
  UpdateContractRequest,
  ProjectDocument,
  DocumentListFilter,
  UploadDocumentRequest,
  UpdateDocumentRequest,
  CorrectiveAction,
  CorrectiveActionListFilter,
  CreateCorrectiveActionRequest,
  UpdateCorrectiveActionRequest,
} from "@/types/project";

export const projectService = {
  list: (params?: ProjectListFilter) =>
    api.get<ApiResponse<Project[]>>("/projects", { params }),

  get: (id: string) =>
    api.get<ApiResponse<Project>>(`/projects/${id}`),

  create: (data: CreateProjectRequest) =>
    api.post<ApiResponse<Project>>("/projects", data),

  update: (id: string, data: UpdateProjectRequest) =>
    api.put<ApiResponse<Project>>(`/projects/${id}`, data),

  transition: (id: string, data: TransitionProjectRequest) =>
    api.post<ApiResponse<Project>>(`/projects/${id}/transition`, data),

  delete: (id: string) =>
    api.delete<ApiResponse<null>>(`/projects/${id}`),

  getProgressHistory: (id: string) =>
    api.get<ApiResponse<ProgressHistory[]>>(`/projects/${id}/progress-history`),

  listTasks: (projectId: string, params?: TaskListFilter) =>
    api.get<ApiResponse<Task[]>>(`/projects/${projectId}/tasks`, { params }),

  getTask: (projectId: string, taskId: string) =>
    api.get<ApiResponse<Task>>(`/projects/${projectId}/tasks/${taskId}`),

  createTask: (projectId: string, data: CreateTaskRequest) =>
    api.post<ApiResponse<Task>>(`/projects/${projectId}/tasks`, data),

  updateTask: (projectId: string, taskId: string, data: UpdateTaskRequest) =>
    api.put<ApiResponse<Task>>(`/projects/${projectId}/tasks/${taskId}`, data),

  deleteTask: (projectId: string, taskId: string) =>
    api.delete<ApiResponse<null>>(`/projects/${projectId}/tasks/${taskId}`),

  listMilestones: (projectId: string) =>
    api.get<ApiResponse<Milestone[]>>(`/projects/${projectId}/milestones`),

  getMilestone: (projectId: string, milestoneId: string) =>
    api.get<ApiResponse<Milestone>>(
      `/projects/${projectId}/milestones/${milestoneId}`
    ),

  createMilestone: (projectId: string, data: CreateMilestoneRequest) =>
    api.post<ApiResponse<Milestone>>(`/projects/${projectId}/milestones`, data),

  updateMilestone: (
    projectId: string,
    milestoneId: string,
    data: UpdateMilestoneRequest
  ) =>
    api.put<ApiResponse<Milestone>>(
      `/projects/${projectId}/milestones/${milestoneId}`,
      data
    ),

  deleteMilestone: (projectId: string, milestoneId: string) =>
    api.delete<ApiResponse<null>>(
      `/projects/${projectId}/milestones/${milestoneId}`
    ),

  listIssues: (projectId: string, params?: IssueListFilter) =>
    api.get<ApiResponse<Issue[]>>(`/projects/${projectId}/issues`, { params }),

  getIssue: (projectId: string, issueId: string) =>
    api.get<ApiResponse<Issue>>(`/projects/${projectId}/issues/${issueId}`),

  createIssue: (projectId: string, data: CreateIssueRequest) =>
    api.post<ApiResponse<Issue>>(`/projects/${projectId}/issues`, data),

  updateIssue: (projectId: string, issueId: string, data: UpdateIssueRequest) =>
    api.put<ApiResponse<Issue>>(
      `/projects/${projectId}/issues/${issueId}`,
      data
    ),

  deleteIssue: (projectId: string, issueId: string) =>
    api.delete<ApiResponse<null>>(
      `/projects/${projectId}/issues/${issueId}`
    ),

  listRisks: (projectId: string, params?: RiskListFilter) =>
    api.get<ApiResponse<Risk[]>>(`/projects/${projectId}/risks`, { params }),

  getRisk: (projectId: string, riskId: string) =>
    api.get<ApiResponse<Risk>>(`/projects/${projectId}/risks/${riskId}`),

  createRisk: (projectId: string, data: CreateRiskRequest) =>
    api.post<ApiResponse<Risk>>(`/projects/${projectId}/risks`, data),

  updateRisk: (projectId: string, riskId: string, data: UpdateRiskRequest) =>
    api.put<ApiResponse<Risk>>(
      `/projects/${projectId}/risks/${riskId}`,
      data
    ),

  deleteRisk: (projectId: string, riskId: string) =>
    api.delete<ApiResponse<null>>(
      `/projects/${projectId}/risks/${riskId}`
    ),

  listBudgets: (projectId: string, params?: BudgetListFilter) =>
    api.get<ApiResponse<BudgetLine[]>>(
      `/projects/${projectId}/budgets`,
      { params }
    ),

  getBudget: (projectId: string, budgetId: string) =>
    api.get<ApiResponse<BudgetLine>>(
      `/projects/${projectId}/budgets/${budgetId}`
    ),

  createBudget: (projectId: string, data: CreateBudgetRequest) =>
    api.post<ApiResponse<BudgetLine>>(
      `/projects/${projectId}/budgets`,
      data
    ),

  updateBudget: (
    projectId: string,
    budgetId: string,
    data: UpdateBudgetRequest
  ) =>
    api.put<ApiResponse<BudgetLine>>(
      `/projects/${projectId}/budgets/${budgetId}`,
      data
    ),

  deleteBudget: (projectId: string, budgetId: string) =>
    api.delete<ApiResponse<null>>(
      `/projects/${projectId}/budgets/${budgetId}`
    ),

  listVendors: (params?: VendorListFilter) =>
    api.get<ApiResponse<Vendor[]>>("/vendors", { params }),

  getVendor: (vendorId: string) =>
    api.get<ApiResponse<Vendor>>(`/vendors/${vendorId}`),

  createVendor: (data: CreateVendorRequest) =>
    api.post<ApiResponse<Vendor>>("/vendors", data),

  updateVendor: (vendorId: string, data: UpdateVendorRequest) =>
    api.put<ApiResponse<Vendor>>(`/vendors/${vendorId}`, data),

  deleteVendor: (vendorId: string) =>
    api.delete<ApiResponse<null>>(`/vendors/${vendorId}`),

  listContracts: (projectId: string, params?: ContractListFilter) =>
    api.get<ApiResponse<Contract[]>>(
      `/projects/${projectId}/contracts`,
      { params }
    ),

  getContract: (projectId: string, contractId: string) =>
    api.get<ApiResponse<Contract>>(
      `/projects/${projectId}/contracts/${contractId}`
    ),

  createContract: (projectId: string, data: CreateContractRequest) =>
    api.post<ApiResponse<Contract>>(
      `/projects/${projectId}/contracts`,
      data
    ),

  updateContract: (
    projectId: string,
    contractId: string,
    data: UpdateContractRequest
  ) =>
    api.put<ApiResponse<Contract>>(
      `/projects/${projectId}/contracts/${contractId}`,
      data
    ),

  deleteContract: (projectId: string, contractId: string) =>
    api.delete<ApiResponse<null>>(
      `/projects/${projectId}/contracts/${contractId}`
    ),

  listDocuments: (projectId: string, params?: DocumentListFilter) =>
    api.get<ApiResponse<ProjectDocument[]>>(
      `/projects/${projectId}/documents`,
      { params }
    ),

  getDocument: (projectId: string, documentId: string) =>
    api.get<ApiResponse<ProjectDocument>>(
      `/projects/${projectId}/documents/${documentId}`
    ),

  uploadDocument: (projectId: string, data: UploadDocumentRequest) => {
    const form = new FormData();
    form.append("file", data.file);
    if (data.name) form.append("name", data.name);
    if (data.category) form.append("category", data.category);
    if (data.version) form.append("version", data.version);
    return api.post<ApiResponse<ProjectDocument>>(
      `/projects/${projectId}/documents`,
      form,
      { headers: { "Content-Type": "multipart/form-data" } }
    );
  },

  updateDocument: (
    projectId: string,
    documentId: string,
    data: UpdateDocumentRequest
  ) =>
    api.put<ApiResponse<ProjectDocument>>(
      `/projects/${projectId}/documents/${documentId}`,
      data
    ),

  deleteDocument: (projectId: string, documentId: string) =>
    api.delete<ApiResponse<null>>(
      `/projects/${projectId}/documents/${documentId}`
    ),

  downloadDocument: async (projectId: string, documentId: string) => {
    const response = await api.get<Blob>(
      `/projects/${projectId}/documents/${documentId}/download`,
      { responseType: "blob" }
    );
    const disposition = response.headers["content-disposition"] ?? "";
    const match = disposition.match(/filename\*=UTF-8''([^;]+)/i);
    const fallback = disposition.match(/filename="?([^";]+)"?/i);
    let filename = `document-${documentId}`;
    if (match?.[1]) {
      filename = decodeURIComponent(match[1]);
    } else if (fallback?.[1]) {
      filename = fallback[1];
    }
    return { blob: response.data, filename };
  },

  // -------------------------------------------------------------------------
  // Corrective Actions (P1-006)
  // -------------------------------------------------------------------------

  listCorrectiveActions: (projectId: string, params?: CorrectiveActionListFilter) =>
    api.get<ApiResponse<CorrectiveAction[]>>(
      `/projects/${projectId}/corrective-actions`,
      { params }
    ),

  getCorrectiveAction: (projectId: string, caId: string) =>
    api.get<ApiResponse<CorrectiveAction>>(
      `/projects/${projectId}/corrective-actions/${caId}`
    ),

  createCorrectiveAction: (projectId: string, data: CreateCorrectiveActionRequest) =>
    api.post<ApiResponse<CorrectiveAction>>(
      `/projects/${projectId}/corrective-actions`,
      data
    ),

  updateCorrectiveAction: (
    projectId: string,
    caId: string,
    data: UpdateCorrectiveActionRequest
  ) =>
    api.put<ApiResponse<CorrectiveAction>>(
      `/projects/${projectId}/corrective-actions/${caId}`,
      data
    ),

  deleteCorrectiveAction: (projectId: string, caId: string) =>
    api.delete<ApiResponse<null>>(
      `/projects/${projectId}/corrective-actions/${caId}`
    ),

  transitionCorrectiveAction: (
    projectId: string,
    caId: string,
    toStatus: string
  ) =>
    api.post<ApiResponse<CorrectiveAction>>(
      `/projects/${projectId}/corrective-actions/${caId}/transition`,
      { to_status: toStatus }
    ),
};
