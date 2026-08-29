import { api } from '@/lib/axios';
import type { CreateFieldInspectionRequest, FieldInspection, VerificationStatus } from '@/types/field';

interface ApiResponse<T> { message: string; data: T }

export const fieldService = {
  list: async (projectId: string): Promise<FieldInspection[]> => {
    const response = await api.get<ApiResponse<FieldInspection[]>>(`/projects/${projectId}/inspections`);
    return response.data.data;
  },
  create: async (projectId: string, data: CreateFieldInspectionRequest): Promise<FieldInspection> => {
    const form = new FormData();
    form.append('inspected_at', data.inspected_at);
    if (data.latitude !== undefined) form.append('latitude', String(data.latitude));
    if (data.longitude !== undefined) form.append('longitude', String(data.longitude));
    if (data.notes) form.append('notes', data.notes);
    if (data.file) form.append('file', data.file);
    const response = await api.post<ApiResponse<FieldInspection>>(`/projects/${projectId}/inspections`, form);
    return response.data.data;
  },
  verify: async (projectId: string, inspectionId: string, status: VerificationStatus): Promise<FieldInspection> => {
    const response = await api.patch<ApiResponse<FieldInspection>>(`/projects/${projectId}/inspections/${inspectionId}/verification`, { status });
    return response.data.data;
  },
  remove: async (projectId: string, inspectionId: string): Promise<void> => {
    await api.delete(`/projects/${projectId}/inspections/${inspectionId}`);
  },
  addEvidence: async (
    projectId: string,
    inspectionId: string,
    file: File,
    latitude?: number,
    longitude?: number,
  ): Promise<FieldInspection> => {
    const form = new FormData();
    form.append('file', file);
    if (latitude !== undefined) form.append('latitude', String(latitude));
    if (longitude !== undefined) form.append('longitude', String(longitude));
    const response = await api.post<ApiResponse<FieldInspection>>(
      `/projects/${projectId}/inspections/${inspectionId}/evidence`,
      form,
    );
    return response.data.data;
  },
  downloadUrl: (projectId: string, inspectionId: string, evidenceId: string): string =>
    `${api.defaults.baseURL}/projects/${projectId}/inspections/${inspectionId}/evidence/${evidenceId}/download`,
  download: async (projectId: string, inspectionId: string, evidenceId: string): Promise<Blob> => {
    const response = await api.get(`/projects/${projectId}/inspections/${inspectionId}/evidence/${evidenceId}/download`, { responseType: 'blob' });
    return response.data as Blob;
  },
};