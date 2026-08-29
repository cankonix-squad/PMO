import { api } from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  GovernanceSubmission,
  GovernanceSubmissionDetail,
  GovernanceSubmissionItem,
  GovernanceLockPeriod,
  GovernanceSubmissionListResponse,
  GovernanceLockPeriodListResponse,
  CreateGovernanceSubmissionRequest,
  ReviewGovernanceRequest,
  RejectGovernanceRequest,
  LockGovernanceRequest,
  CancelGovernanceRequest,
  CreateLockPeriodRequest,
  GovernanceSubmissionFilter,
  GovernanceLockPeriodFilter,
} from "@/types/governance";

const BASE = "/governance";

// The backend envelopes list responses twice: { data: { data: [...], meta } }.
// The helpers unwrap the inner envelope.
interface ListEnvelope<T> {
  data: T[];
  meta: { total: number; page: number; page_size: number };
}

export const governanceService = {
  // ---------------------------------------------------------------------------
  // Submissions
  // ---------------------------------------------------------------------------

  listSubmissions: async (
    filter: GovernanceSubmissionFilter = {}
  ): Promise<GovernanceSubmissionListResponse> => {
    const params = new URLSearchParams();
    if (filter.status) params.set("status", filter.status);
    if (filter.dataset_type) params.set("dataset_type", filter.dataset_type);
    if (filter.source_type) params.set("source_type", filter.source_type);
    if (filter.period_year) params.set("period_year", String(filter.period_year));
    if (filter.page) params.set("page", String(filter.page));
    if (filter.page_size) params.set("page_size", String(filter.page_size));
    const qs = params.toString();
    const response = await api.get<ApiResponse<ListEnvelope<GovernanceSubmission>>>(
      `${BASE}/submissions${qs ? `?${qs}` : ""}`
    );
    return response.data.data;
  },

  getSubmission: async (id: string): Promise<GovernanceSubmissionDetail> => {
    const response = await api.get<ApiResponse<GovernanceSubmissionDetail>>(
      `${BASE}/submissions/${id}`
    );
    return response.data.data;
  },

  createSubmission: async (
    payload: CreateGovernanceSubmissionRequest
  ): Promise<GovernanceSubmission> => {
    const response = await api.post<ApiResponse<GovernanceSubmission>>(
      `${BASE}/submissions`,
      payload
    );
    return response.data.data;
  },

  submit: async (id: string): Promise<GovernanceSubmission> => {
    const response = await api.post<ApiResponse<GovernanceSubmission>>(
      `${BASE}/submissions/${id}/submit`,
      {}
    );
    return response.data.data;
  },

  startReview: async (
    id: string,
    payload: ReviewGovernanceRequest = {}
  ): Promise<GovernanceSubmissionDetail> => {
    const response = await api.post<ApiResponse<GovernanceSubmissionDetail>>(
      `${BASE}/submissions/${id}/review`,
      payload
    );
    return response.data.data;
  },

  approve: async (id: string): Promise<GovernanceSubmission> => {
    const response = await api.post<ApiResponse<GovernanceSubmission>>(
      `${BASE}/submissions/${id}/approve`,
      {}
    );
    return response.data.data;
  },

  reject: async (id: string, payload: RejectGovernanceRequest): Promise<GovernanceSubmission> => {
    const response = await api.post<ApiResponse<GovernanceSubmission>>(
      `${BASE}/submissions/${id}/reject`,
      payload
    );
    return response.data.data;
  },

  lock: async (id: string, payload: LockGovernanceRequest = {}): Promise<GovernanceSubmission> => {
    const response = await api.post<ApiResponse<GovernanceSubmission>>(
      `${BASE}/submissions/${id}/lock`,
      payload
    );
    return response.data.data;
  },

  cancel: async (
    id: string,
    payload: CancelGovernanceRequest = {}
  ): Promise<GovernanceSubmission> => {
    const response = await api.post<ApiResponse<GovernanceSubmission>>(
      `${BASE}/submissions/${id}/cancel`,
      payload
    );
    return response.data.data;
  },

  // ---------------------------------------------------------------------------
  // Lock Periods
  // ---------------------------------------------------------------------------

  listLockPeriods: async (
    filter: GovernanceLockPeriodFilter = {}
  ): Promise<GovernanceLockPeriodListResponse> => {
    const params = new URLSearchParams();
    if (filter.dataset_type) params.set("dataset_type", filter.dataset_type);
    if (filter.status) params.set("status", filter.status);
    if (filter.period_year) params.set("period_year", String(filter.period_year));
    if (filter.page) params.set("page", String(filter.page));
    if (filter.page_size) params.set("page_size", String(filter.page_size));
    const qs = params.toString();
    const response = await api.get<ApiResponse<ListEnvelope<GovernanceLockPeriod>>>(
      `${BASE}/lock-periods${qs ? `?${qs}` : ""}`
    );
    return response.data.data;
  },

  createLockPeriod: async (payload: CreateLockPeriodRequest): Promise<GovernanceLockPeriod> => {
    const response = await api.post<ApiResponse<GovernanceLockPeriod>>(
      `${BASE}/lock-periods`,
      payload
    );
    return response.data.data;
  },

  lockPeriod: async (id: string, payload: LockGovernanceRequest = {}): Promise<GovernanceLockPeriod> => {
    const response = await api.post<ApiResponse<GovernanceLockPeriod>>(
      `${BASE}/lock-periods/${id}/lock`,
      payload
    );
    return response.data.data;
  },

  // Convenience: unwrap a single item from a double-enveloped detail response.
  unwrapItem: (items: GovernanceSubmissionItem[]): GovernanceSubmissionItem[] => items,
};
