export interface Sector {
  id: string;
  organization_id: string;
  code: string;
  name: string;
  description?: string;
  is_active: boolean;
  sort_order: number;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateSectorRequest {
  code: string;
  name: string;
  description?: string;
  sort_order?: number;
}

export interface UpdateSectorRequest {
  code?: string;
  name?: string;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface Region {
  id: string;
  organization_id: string;
  parent_id?: string;
  code: string;
  name: string;
  level: number;
  description?: string;
  is_active: boolean;
  sort_order: number;
  created_by?: string;
  created_at: string;
  updated_at: string;
  children?: Region[];
}

export interface CreateRegionRequest {
  parent_id?: string;
  code: string;
  name: string;
  level: number;
  description?: string;
  sort_order?: number;
}

export interface UpdateRegionRequest {
  parent_id?: string;
  code?: string;
  name?: string;
  level?: number;
  is_active?: boolean;
  sort_order?: number;
}

export interface RiverBasin {
  id: string;
  organization_id: string;
  region_id?: string;
  code: string;
  name: string;
  description?: string;
  area_km2?: number;
  is_active: boolean;
  sort_order: number;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateRiverBasinRequest {
  region_id?: string;
  code: string;
  name: string;
  description?: string;
  area_km2?: number;
  sort_order?: number;
}

export interface UpdateRiverBasinRequest {
  region_id?: string;
  code?: string;
  name?: string;
  description?: string;
  area_km2?: number;
  is_active?: boolean;
  sort_order?: number;
}

export const RegionLevelLabel: Record<number, string> = {
  1: 'Nasional',
  2: 'Provinsi',
  3: 'Kabupaten/Kota',
  4: 'Kecamatan',
};
