import { api } from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  ActivityMapping,
  ListMappingsParams,
  ListRunsParams,
  PaginatedMappingsResponse,
  PaginatedRunsResponse,
  SyncRun,
} from "@/types/primavera";

const BASE = "/integrations/primavera";

export const primaveraService = {
  listRuns: async (params?: ListRunsParams): Promise<PaginatedRunsResponse> => {
    const response = await api.get<ApiResponse<PaginatedRunsResponse>>(`${BASE}/runs`, { params });
    return response.data.data;
  },

  getRun: async (runID: string): Promise<SyncRun> => {
    const response = await api.get<ApiResponse<SyncRun>>(`${BASE}/runs/${runID}`);
    return response.data.data;
  },

  createRun: async (
    projectID: string,
    file: File,
    opts?: {
      format?: string;
      sourceProjectID?: string;
      exportedAt?: string;
      p6Version?: string;
      operator?: string;
    }
  ): Promise<SyncRun> => {
    const form = new FormData();
    form.append("project_id", projectID);
    form.append("file", file);
    if (opts?.format) form.append("format", opts.format);
    if (opts?.sourceProjectID) form.append("source_project_id", opts.sourceProjectID);
    if (opts?.exportedAt) form.append("exported_at", opts.exportedAt);
    if (opts?.p6Version) form.append("p6_version", opts.p6Version);
    if (opts?.operator) form.append("operator", opts.operator);

    const response = await api.post<ApiResponse<SyncRun>>(`${BASE}/runs`, form, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return response.data.data;
  },

  cancelRun: async (runID: string): Promise<SyncRun> => {
    const response = await api.post<ApiResponse<SyncRun>>(`${BASE}/runs/${runID}/cancel`);
    return response.data.data;
  },

  listMappings: async (
    runID: string,
    params?: ListMappingsParams
  ): Promise<PaginatedMappingsResponse> => {
    const response = await api.get<ApiResponse<PaginatedMappingsResponse>>(
      `${BASE}/runs/${runID}/mappings`,
      { params }
    );
    return response.data.data;
  },
};
