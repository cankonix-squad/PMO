import { api } from '@/lib/axios';
import type { ProjectControl } from '@/types/project-control';
interface ApiResponse<T> { message: string; data: T }
export const projectControlService = { get: async (projectId: string, year?: number, month?: number): Promise<ProjectControl> => { const response = await api.get<ApiResponse<ProjectControl>>(`/projects/${projectId}/control`, { params: year && month ? { year, month } : undefined }); return response.data.data; } };