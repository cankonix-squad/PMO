import { api } from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  ConnectorDefinition,
  ConnectorConfig,
  SyncRun,
  ExternalMapping,
  CreateRunRequest,
  ListRunsParams,
  ListRecordsParams,
  ListMappingsParams,
  ListPendingMappingsParams,
  PaginatedRunsResponse,
  PaginatedRecordsResponse,
  PaginatedMappingsResponse,
  MatchMappingRequest,
  RejectMappingRequest,
  CandidatesResponse,
} from "@/types/government";

const BASE = "/integrations/government";

export const governmentService = {
  // --- Connectors ---

  listConnectors: async (): Promise<ConnectorDefinition[]> => {
    const response = await api.get<ApiResponse<ConnectorDefinition[]>>(`${BASE}/connectors`);
    return response.data.data;
  },

  getConnector: async (key: string): Promise<ConnectorDefinition> => {
    const response = await api.get<ApiResponse<ConnectorDefinition>>(`${BASE}/connectors/${key}`);
    return response.data.data;
  },

  getConfig: async (): Promise<Record<string, ConnectorConfig>> => {
    const response = await api.get<ApiResponse<Record<string, ConnectorConfig>>>(`${BASE}/config`);
    return response.data.data;
  },

  // --- Sync Runs ---

  listRuns: async (params?: ListRunsParams): Promise<PaginatedRunsResponse> => {
    const response = await api.get<ApiResponse<PaginatedRunsResponse>>(`${BASE}/runs`, { params });
    return response.data.data;
  },

  createRun: async (payload: CreateRunRequest): Promise<SyncRun> => {
    const response = await api.post<ApiResponse<SyncRun>>(`${BASE}/runs`, payload);
    return response.data.data;
  },

  getRun: async (runID: string): Promise<SyncRun> => {
    const response = await api.get<ApiResponse<SyncRun>>(`${BASE}/runs/${runID}`);
    return response.data.data;
  },

  cancelRun: async (runID: string): Promise<SyncRun> => {
    const response = await api.post<ApiResponse<SyncRun>>(`${BASE}/runs/${runID}/cancel`);
    return response.data.data;
  },

  // --- Records & Mappings ---

  listRecords: async (runID: string, params?: ListRecordsParams): Promise<PaginatedRecordsResponse> => {
    const response = await api.get<ApiResponse<PaginatedRecordsResponse>>(
      `${BASE}/runs/${runID}/records`,
      { params }
    );
    return response.data.data;
  },

  listMappings: async (params?: ListMappingsParams): Promise<PaginatedMappingsResponse> => {
    const response = await api.get<ApiResponse<PaginatedMappingsResponse>>(`${BASE}/mappings`, {
      params,
    });
    return response.data.data;
  },

  // --- Entity Resolution (P3-002) ---

  listPendingMappings: async (params?: ListPendingMappingsParams): Promise<PaginatedMappingsResponse> => {
    const response = await api.get<ApiResponse<PaginatedMappingsResponse>>(
      `${BASE}/mappings/pending`,
      { params }
    );
    return response.data.data;
  },

  getMapping: async (id: string): Promise<ExternalMapping> => {
    const response = await api.get<ApiResponse<ExternalMapping>>(`${BASE}/mappings/${id}`);
    return response.data.data;
  },

  getMappingCandidates: async (id: string): Promise<CandidatesResponse> => {
    const response = await api.get<ApiResponse<CandidatesResponse>>(
      `${BASE}/mappings/${id}/candidates`
    );
    return response.data.data;
  },

  matchMapping: async (id: string, req: MatchMappingRequest): Promise<ExternalMapping> => {
    const response = await api.post<ApiResponse<ExternalMapping>>(
      `${BASE}/mappings/${id}/match`,
      req
    );
    return response.data.data;
  },

  unmatchMapping: async (id: string): Promise<ExternalMapping> => {
    const response = await api.post<ApiResponse<ExternalMapping>>(
      `${BASE}/mappings/${id}/unmatch`,
      {}
    );
    return response.data.data;
  },

  rejectMapping: async (id: string, req: RejectMappingRequest): Promise<ExternalMapping> => {
    const response = await api.post<ApiResponse<ExternalMapping>>(
      `${BASE}/mappings/${id}/reject`,
      req
    );
    return response.data.data;
  },
};
