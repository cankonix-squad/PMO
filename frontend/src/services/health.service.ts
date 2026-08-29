import { api } from '@/lib/axios';
import type { CalculateHealthRequest, HealthSnapshot } from '@/types/health';

interface ApiResponse<T> { message: string; data: T }

export const healthService = {
  listSnapshots: async (projectId: string): Promise<HealthSnapshot[]> => {
    const response = await api.get<ApiResponse<HealthSnapshot[]>>(`/projects/${projectId}/health`);
    return response.data.data;
  },
  calculate: async (projectId: string, data: CalculateHealthRequest): Promise<HealthSnapshot> => {
    const response = await api.post<ApiResponse<HealthSnapshot>>(`/projects/${projectId}/health/calculate`, data);
    return response.data.data;
  },
};