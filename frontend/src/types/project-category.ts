export interface ProjectCategory {
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

export interface CreateProjectCategoryRequest {
  code: string;
  name: string;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdateProjectCategoryRequest {
  code?: string;
  name?: string;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}
