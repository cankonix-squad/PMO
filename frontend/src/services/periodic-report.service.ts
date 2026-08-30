// PMO-DASH-002: Periodic report API service
import { api } from "@/lib/axios";
import type {
  PeriodicReport,
  CreatePeriodicReportRequest,
  UpdatePeriodicReportRequest,
  PeriodicReportListFilter,
} from "@/types/periodic-report";
import type { ApiResponse } from "@/types/api";

export const periodicReportService = {
  async list(
    projectId: string,
    filter: PeriodicReportListFilter = {}
  ): Promise<ApiResponse<PeriodicReport[]>> {
    const params: Record<string, string | number> = {};
    if (filter.year) params.year = filter.year;
    if (filter.month) params.month = filter.month;
    if (filter.page) params.page = filter.page;
    if (filter.page_size) params.page_size = filter.page_size;

    const { data } = await api.get<ApiResponse<PeriodicReport[]>>(
      `/projects/${projectId}/periodic-reports`,
      { params }
    );
    return data;
  },

  async get(projectId: string, reportId: string): Promise<PeriodicReport> {
    const { data } = await api.get<ApiResponse<PeriodicReport>>(
      `/projects/${projectId}/periodic-reports/${reportId}`
    );
    return data.data;
  },

  async create(
    projectId: string,
    req: CreatePeriodicReportRequest
  ): Promise<PeriodicReport> {
    const { data } = await api.post<ApiResponse<PeriodicReport>>(
      `/projects/${projectId}/periodic-reports`,
      req
    );
    return data.data;
  },

  async update(
    projectId: string,
    reportId: string,
    req: UpdatePeriodicReportRequest
  ): Promise<PeriodicReport> {
    const { data } = await api.put<ApiResponse<PeriodicReport>>(
      `/projects/${projectId}/periodic-reports/${reportId}`,
      req
    );
    return data.data;
  },

  async delete(projectId: string, reportId: string): Promise<void> {
    await api.delete(
      `/projects/${projectId}/periodic-reports/${reportId}`
    );
  },
};
