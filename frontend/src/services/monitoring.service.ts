import { api } from '@/lib/axios';
import type {
  Baseline,
  CreateBaselineRequest,
  UpdateBaselineRequest,
  Snapshot,
  CreateSnapshotRequest,
  UpdateSnapshotRequest,
  TransitionSnapshotRequest,
} from '@/types/monitoring';

interface ApiResponse<T> {
  message: string;
  data: T;
}

// ── Baselines ──────────────────────────────────────────────────────────────

export const baselineService = {
  list: async (projectId: string): Promise<Baseline[]> => {
    const res = await api.get<ApiResponse<Baseline[]>>(`/projects/${projectId}/baselines`);
    return res.data.data;
  },

  get: async (projectId: string, id: string): Promise<Baseline> => {
    const res = await api.get<ApiResponse<Baseline>>(`/projects/${projectId}/baselines/${id}`);
    return res.data.data;
  },

  create: async (projectId: string, data: CreateBaselineRequest): Promise<Baseline> => {
    const res = await api.post<ApiResponse<Baseline>>(`/projects/${projectId}/baselines`, data);
    return res.data.data;
  },

  update: async (projectId: string, id: string, data: UpdateBaselineRequest): Promise<Baseline> => {
    const res = await api.put<ApiResponse<Baseline>>(`/projects/${projectId}/baselines/${id}`, data);
    return res.data.data;
  },

  delete: async (projectId: string, id: string): Promise<void> => {
    await api.delete(`/projects/${projectId}/baselines/${id}`);
  },
};

// ── Snapshots ──────────────────────────────────────────────────────────────

export const snapshotService = {
  list: async (projectId: string, status?: string): Promise<Snapshot[]> => {
    const params = status ? { status } : {};
    const res = await api.get<ApiResponse<Snapshot[]>>(`/projects/${projectId}/snapshots`, { params });
    return res.data.data;
  },

  get: async (projectId: string, id: string): Promise<Snapshot> => {
    const res = await api.get<ApiResponse<Snapshot>>(`/projects/${projectId}/snapshots/${id}`);
    return res.data.data;
  },

  create: async (projectId: string, data: CreateSnapshotRequest): Promise<Snapshot> => {
    const res = await api.post<ApiResponse<Snapshot>>(`/projects/${projectId}/snapshots`, data);
    return res.data.data;
  },

  update: async (projectId: string, id: string, data: UpdateSnapshotRequest): Promise<Snapshot> => {
    const res = await api.put<ApiResponse<Snapshot>>(`/projects/${projectId}/snapshots/${id}`, data);
    return res.data.data;
  },

  transition: async (projectId: string, id: string, data: TransitionSnapshotRequest): Promise<Snapshot> => {
    const res = await api.patch<ApiResponse<Snapshot>>(`/projects/${projectId}/snapshots/${id}/status`, data);
    return res.data.data;
  },

  delete: async (projectId: string, id: string): Promise<void> => {
    await api.delete(`/projects/${projectId}/snapshots/${id}`);
  },
};
