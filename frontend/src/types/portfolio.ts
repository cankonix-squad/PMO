export interface Program {
  id: string;
  organization_id: string;
  code: string;
  name: string;
  description?: string;
  fiscal_year?: number;
  is_active: boolean;
  sort_order: number;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateProgramRequest {
  code: string;
  name: string;
  description?: string;
  fiscal_year?: number;
  sort_order?: number;
}

export interface UpdateProgramRequest {
  code?: string;
  name?: string;
  description?: string;
  fiscal_year?: number;
  is_active?: boolean;
  sort_order?: number;
}
