import { api } from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type { ExecutiveDashboard } from "@/types/executive";

export interface ExecutiveFilter {
  year?: number;
  month?: number;
}

function buildParams(f?: ExecutiveFilter): Record<string, string> {
  const p: Record<string, string> = {};
  if (f?.year) p.year = String(f.year);
  if (f?.month) p.month = String(f.month);
  return p;
}

export const executiveService = {
  getDashboard: async (filter?: ExecutiveFilter): Promise<ExecutiveDashboard> => {
    const response = await api.get<ApiResponse<ExecutiveDashboard>>(
      "/analytics/executive",
      { params: buildParams(filter) }
    );
    return response.data.data;
  },
};
