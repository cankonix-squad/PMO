import { api } from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  BIMModel,
  BIMModelVersion,
  BIMProjectMapping,
  BIMModelListResponse,
  BIMVersionListResponse,
  BIMMappingListResponse,
  CreateBIMModelRequest,
  UpdateBIMModelRequest,
  CreateVersionRequest,
  LinkProjectRequest,
} from "@/types/bim";

const BASE = "/integrations/bim";

export const bimService = {
  // ---------------------------------------------------------------------------
  // BIM Models
  // ---------------------------------------------------------------------------

  listModels: async (): Promise<BIMModelListResponse> => {
    const response = await api.get<ApiResponse<BIMModelListResponse>>(`${BASE}/models`);
    return response.data.data;
  },

  getModel: async (id: string): Promise<BIMModel> => {
    const response = await api.get<ApiResponse<{ data: BIMModel }>>(`${BASE}/models/${id}`);
    return response.data.data.data;
  },

  createModel: async (payload: CreateBIMModelRequest): Promise<BIMModel> => {
    const response = await api.post<ApiResponse<{ data: BIMModel }>>(`${BASE}/models`, payload);
    return response.data.data.data;
  },

  updateModel: async (id: string, payload: UpdateBIMModelRequest): Promise<BIMModel> => {
    const response = await api.patch<ApiResponse<{ data: BIMModel }>>(
      `${BASE}/models/${id}`,
      payload
    );
    return response.data.data.data;
  },

  deleteModel: async (id: string): Promise<void> => {
    await api.delete(`${BASE}/models/${id}`);
  },

  // ---------------------------------------------------------------------------
  // Versions
  // ---------------------------------------------------------------------------

  listVersions: async (modelId: string): Promise<BIMVersionListResponse> => {
    const response = await api.get<ApiResponse<BIMVersionListResponse>>(
      `${BASE}/models/${modelId}/versions`
    );
    return response.data.data;
  },

  addVersion: async (modelId: string, payload: CreateVersionRequest): Promise<BIMModelVersion> => {
    const response = await api.post<ApiResponse<{ data: BIMModelVersion }>>(
      `${BASE}/models/${modelId}/versions`,
      payload
    );
    return response.data.data.data;
  },

  // ---------------------------------------------------------------------------
  // Project Mappings
  // ---------------------------------------------------------------------------

  listMappings: async (modelId: string): Promise<BIMMappingListResponse> => {
    const response = await api.get<ApiResponse<BIMMappingListResponse>>(
      `${BASE}/models/${modelId}/mappings`
    );
    return response.data.data;
  },

  linkProject: async (
    modelId: string,
    payload: LinkProjectRequest
  ): Promise<BIMProjectMapping> => {
    const response = await api.post<ApiResponse<{ data: BIMProjectMapping }>>(
      `${BASE}/models/${modelId}/mappings`,
      payload
    );
    return response.data.data.data;
  },

  unlinkProject: async (modelId: string, projectId: string): Promise<void> => {
    await api.delete(`${BASE}/models/${modelId}/mappings/${projectId}`);
  },
};
