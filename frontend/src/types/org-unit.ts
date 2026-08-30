export type OrgUnitLevel = 1 | 2 | 3 | 4 | 5;

export const OrgUnitLevelLabel: Record<OrgUnitLevel, string> = {
  1: 'Kementerian',
  2: 'Ditjen',
  3: 'Direktorat',
  4: 'Subdit',
  5: 'Unit / Satker',
};

export interface OrgUnit {
  id: string;
  organization_id: string;
  parent_id?: string | null;
  code: string;
  name: string;
  level: OrgUnitLevel;
  head_user_id?: string | null;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
  children?: OrgUnit[];
}

export interface CreateOrgUnitRequest {
  parent_id?: string | null;
  code: string;
  name: string;
  level: OrgUnitLevel;
  sort_order?: number;
}

export interface UpdateOrgUnitRequest {
  parent_id?: string | null;
  code?: string;
  name?: string;
  level?: OrgUnitLevel;
  head_user_id?: string | null;
  is_active?: boolean;
  sort_order?: number;
}
