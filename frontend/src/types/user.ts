import type { PaginationParams } from "./api";

export interface RoleRef {
  id: string;
  code: string;
  name: string;
}

export interface UserProfile {
  id: string;
  organization_id: string;
  org_unit_id?: string | null;
  org_unit_name?: string;
  employee_id?: string | null;
  first_name: string;
  last_name: string;
  full_name: string;
  email: string;
  phone?: string;
  job_title?: string;
  avatar_url?: string;
  is_active: boolean;
  must_change_pwd: boolean;
  last_login_at?: string | null;
  roles?: RoleRef[];
  created_at: string;
}

export interface UserListFilter extends PaginationParams {
  is_active?: boolean;
}

export interface CreateUserRequest {
  first_name: string;
  last_name?: string;
  email: string;
  password: string;
  phone?: string;
  job_title?: string;
  employee_id?: string;
  is_active?: boolean;
  role_ids?: string[];
}

export interface UpdateUserRequest {
  first_name?: string;
  last_name?: string;
  phone?: string;
  job_title?: string;
  employee_id?: string;
  avatar_url?: string;
  is_active?: boolean;
  role_ids?: string[];
}
