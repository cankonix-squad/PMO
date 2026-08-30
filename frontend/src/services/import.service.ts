import { api } from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  ImportJob,
  ImportRow,
  ImportTemplate,
  ListJobsParams,
  ListRowsParams,
  PaginatedResponse,
} from "@/types/import";

export const importService = {
  listTemplates: async (): Promise<ImportTemplate[]> => {
    const response = await api.get<ApiResponse<ImportTemplate[]>>("/imports/templates");
    return response.data.data;
  },

  listJobs: async (params?: ListJobsParams): Promise<PaginatedResponse<ImportJob>> => {
    const response = await api.get<ApiResponse<PaginatedResponse<ImportJob>>>("/imports/jobs", {
      params,
    });
    return response.data.data;
  },

  getJob: async (jobID: string): Promise<ImportJob> => {
    const response = await api.get<ApiResponse<ImportJob>>(`/imports/jobs/${jobID}`);
    return response.data.data;
  },

  createJob: async (datasetType: string, file: File): Promise<ImportJob> => {
    const form = new FormData();
    form.append("dataset_type", datasetType);
    form.append("file", file);
    const response = await api.post<ApiResponse<ImportJob>>("/imports/jobs", form, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return response.data.data;
  },

  validateJob: async (jobID: string): Promise<ImportJob> => {
    const response = await api.post<ApiResponse<ImportJob>>(`/imports/jobs/${jobID}/validate`);
    return response.data.data;
  },

  commitJob: async (jobID: string): Promise<ImportJob> => {
    const response = await api.post<ApiResponse<ImportJob>>(`/imports/jobs/${jobID}/commit`);
    return response.data.data;
  },

  cancelJob: async (jobID: string): Promise<ImportJob> => {
    const response = await api.post<ApiResponse<ImportJob>>(`/imports/jobs/${jobID}/cancel`);
    return response.data.data;
  },

  listRows: async (jobID: string, params?: ListRowsParams): Promise<PaginatedResponse<ImportRow>> => {
    const response = await api.get<ApiResponse<PaginatedResponse<ImportRow>>>(
      `/imports/jobs/${jobID}/rows`,
      { params }
    );
    return response.data.data;
  },
};
