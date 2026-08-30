import api from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type { DashboardStats, DashboardTrend } from "@/types/dashboard";

export const dashboardService = {
  getStats: () =>
    api.get<ApiResponse<DashboardStats>>("/dashboard"),

  getTrend: () =>
    api.get<ApiResponse<DashboardTrend>>("/dashboard/trend"),
};
