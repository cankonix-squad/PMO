import { api } from '@/lib/axios';
import {
  Sector, CreateSectorRequest, UpdateSectorRequest,
  Region, CreateRegionRequest, UpdateRegionRequest,
  RiverBasin, CreateRiverBasinRequest, UpdateRiverBasinRequest,
} from '@/types/spatial';

export interface ApiResponse<T> {
  message: string;
  data: T;
}

// ---------------------------------------------------------------------------
// Sectors
// ---------------------------------------------------------------------------
export const sectorService = {
  list: async (includeInactive = false): Promise<Sector[]> => {
    const params = includeInactive ? '?include_inactive=true' : '';
    const res = await api.get<ApiResponse<Sector[]>>(`/sectors${params}`);
    return res.data.data;
  },

  get: async (id: string): Promise<Sector> => {
    const res = await api.get<ApiResponse<Sector>>(`/sectors/${id}`);
    return res.data.data;
  },

  create: async (payload: CreateSectorRequest): Promise<Sector> => {
    const res = await api.post<ApiResponse<Sector>>('/sectors', payload);
    return res.data.data;
  },

  update: async (id: string, payload: UpdateSectorRequest): Promise<Sector> => {
    const res = await api.put<ApiResponse<Sector>>(`/sectors/${id}`, payload);
    return res.data.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/sectors/${id}`);
  },
};

// ---------------------------------------------------------------------------
// Regions
// ---------------------------------------------------------------------------
export const regionService = {
  list: async (includeInactive = false): Promise<Region[]> => {
    const params = includeInactive ? '?include_inactive=true' : '';
    const res = await api.get<ApiResponse<Region[]>>(`/regions${params}`);
    return res.data.data;
  },

  get: async (id: string): Promise<Region> => {
    const res = await api.get<ApiResponse<Region>>(`/regions/${id}`);
    return res.data.data;
  },

  create: async (payload: CreateRegionRequest): Promise<Region> => {
    const res = await api.post<ApiResponse<Region>>('/regions', payload);
    return res.data.data;
  },

  update: async (id: string, payload: UpdateRegionRequest): Promise<Region> => {
    const res = await api.put<ApiResponse<Region>>(`/regions/${id}`, payload);
    return res.data.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/regions/${id}`);
  },
};

// ---------------------------------------------------------------------------
// River Basins
// ---------------------------------------------------------------------------
export const riverBasinService = {
  list: async (includeInactive = false): Promise<RiverBasin[]> => {
    const params = includeInactive ? '?include_inactive=true' : '';
    const res = await api.get<ApiResponse<RiverBasin[]>>(`/river-basins${params}`);
    return res.data.data;
  },

  get: async (id: string): Promise<RiverBasin> => {
    const res = await api.get<ApiResponse<RiverBasin>>(`/river-basins/${id}`);
    return res.data.data;
  },

  create: async (payload: CreateRiverBasinRequest): Promise<RiverBasin> => {
    const res = await api.post<ApiResponse<RiverBasin>>('/river-basins', payload);
    return res.data.data;
  },

  update: async (id: string, payload: UpdateRiverBasinRequest): Promise<RiverBasin> => {
    const res = await api.put<ApiResponse<RiverBasin>>(`/river-basins/${id}`, payload);
    return res.data.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/river-basins/${id}`);
  },
};
