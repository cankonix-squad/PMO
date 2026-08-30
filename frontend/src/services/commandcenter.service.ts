import { api } from '@/lib/axios';
import type { CommandSummary } from '@/types/commandcenter';

interface ApiResponse<T> { message: string; data: T }

export const commandCenterService = {
  getSummary: async (): Promise<CommandSummary> => {
    const response = await api.get<ApiResponse<CommandSummary>>('/command-center');
    return response.data.data;
  },
  updateEscalationStatus: async (id: string, status: 'ACKNOWLEDGED' | 'CLOSED'): Promise<void> => {
    await api.patch(`/command-center/escalations/${id}/status`, { status });
  },
  updateDecisionStatus: async (id: string, status: 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED'): Promise<void> => {
    await api.patch(`/command-center/decisions/${id}/status`, { status });
  },
};