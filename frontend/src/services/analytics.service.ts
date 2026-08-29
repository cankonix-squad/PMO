import { api } from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type { ListResponse, ProgramDashboard } from "@/types/analytics";

export interface AnalyticsFilter {
  year?: number;
  month?: number;
}

function buildParams(f?: AnalyticsFilter): Record<string, string> {
  const p: Record<string, string> = {};
  if (f?.year) p.year = String(f.year);
  if (f?.month) p.month = String(f.month);
  return p;
}

export const analyticsService = {
  // ── Programs ──────────────────────────────────────────────────────────────

  listPrograms: async (filter?: AnalyticsFilter): Promise<ListResponse> => {
    const response = await api.get<ApiResponse<ListResponse>>("/analytics/programs", {
      params: buildParams(filter),
    });
    return response.data.data;
  },

  getProgram: async (id: string, filter?: AnalyticsFilter): Promise<ProgramDashboard> => {
    const response = await api.get<ApiResponse<ProgramDashboard>>(`/analytics/programs/${id}`, {
      params: buildParams(filter),
    });
    return response.data.data;
  },

  // ── Sectors ───────────────────────────────────────────────────────────────

  listSectors: async (filter?: AnalyticsFilter): Promise<ListResponse> => {
    const response = await api.get<ApiResponse<ListResponse>>("/analytics/sectors", {
      params: buildParams(filter),
    });
    return response.data.data;
  },

  getSector: async (id: string, filter?: AnalyticsFilter): Promise<ProgramDashboard> => {
    const response = await api.get<ApiResponse<ProgramDashboard>>(`/analytics/sectors/${id}`, {
      params: buildParams(filter),
    });
    return response.data.data;
  },
};
