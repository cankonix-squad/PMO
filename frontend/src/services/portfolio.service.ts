import { api } from '@/lib/axios';
import { Program, CreateProgramRequest, UpdateProgramRequest } from '@/types/portfolio';

export interface ApiResponse<T> {
  message: string;
  data: T;
}

export const programService = {
  list: async (includeInactive = false): Promise<Program[]> => {
    const params = includeInactive ? '?include_inactive=true' : '';
    const res = await api.get<ApiResponse<Program[]>>(`/programs${params}`);
    return res.data.data;
  },

  get: async (id: string): Promise<Program> => {
    const res = await api.get<ApiResponse<Program>>(`/programs/${id}`);
    return res.data.data;
  },

  create: async (payload: CreateProgramRequest): Promise<Program> => {
    const res = await api.post<ApiResponse<Program>>('/programs', payload);
    return res.data.data;
  },

  update: async (id: string, payload: UpdateProgramRequest): Promise<Program> => {
    const res = await api.put<ApiResponse<Program>>(`/programs/${id}`, payload);
    return res.data.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/programs/${id}`);
  },
};
