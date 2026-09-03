export interface Role {
  id: string;
  organization_id: string;
  code: string;
  name: string;
  description?: string;
  is_system: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateRoleRequest {
  code: string;
  name: string;
  description?: string;
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
}
