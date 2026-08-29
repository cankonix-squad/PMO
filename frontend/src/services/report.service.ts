// Report service — P1-007

import api from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  ReportSnapshot,
  ListReportFilter,
  GenerateReportRequest,
  CreateReportRequest,
  UpdateReportRequest,
  TransitionReportRequest,
} from "@/types/report";

const BASE = "/reports";

function filterToParams(filter: ListReportFilter): Record<string, string> {
  const p: Record<string, string> = {};
  if (filter.period_type) p.period_type = filter.period_type;
  if (filter.status) p.status = filter.status;
  if (filter.project_id) p.project_id = filter.project_id;
  if (filter.page) p.page = String(filter.page);
  if (filter.page_size) p.page_size = String(filter.page_size);
  return p;
}

export const reportService = {
  /** List report snapshots with optional filters. */
  listReports: async (
    filter: ListReportFilter = {}
  ): Promise<ApiResponse<ReportSnapshot[]>> => {
    const { data } = await api.get<ApiResponse<ReportSnapshot[]>>(BASE, {
      params: filterToParams(filter),
    });
    return data;
  },

  /** Get a single report snapshot by ID. */
  getReport: async (id: string): Promise<ReportSnapshot> => {
    const { data } = await api.get<{ data: ReportSnapshot }>(`${BASE}/${id}`);
    return data.data;
  },

  /** Generate a new report snapshot with live metrics computed. */
  generateReport: async (
    req: GenerateReportRequest
  ): Promise<ReportSnapshot> => {
    const { data } = await api.post<{ data: ReportSnapshot }>(
      `${BASE}/generate`,
      req
    );
    return data.data;
  },

  /** Create a manual report snapshot (no metric computation). */
  createReport: async (req: CreateReportRequest): Promise<ReportSnapshot> => {
    const { data } = await api.post<{ data: ReportSnapshot }>(BASE, req);
    return data.data;
  },

  /** Update executive summary of a report. */
  updateReport: async (
    id: string,
    req: UpdateReportRequest
  ): Promise<ReportSnapshot> => {
    const { data } = await api.put<{ data: ReportSnapshot }>(
      `${BASE}/${id}`,
      req
    );
    return data.data;
  },

  /** Delete a report snapshot (soft delete). */
  deleteReport: async (id: string): Promise<void> => {
    await api.delete(`${BASE}/${id}`);
  },

  /** Transition report status: DRAFT→PUBLISHED, PUBLISHED→ARCHIVED, etc. */
  transitionReport: async (
    id: string,
    toStatus: TransitionReportRequest["to_status"]
  ): Promise<ReportSnapshot> => {
    const { data } = await api.post<{ data: ReportSnapshot }>(
      `${BASE}/${id}/transition`,
      { to_status: toStatus }
    );
    return data.data;
  },
};
