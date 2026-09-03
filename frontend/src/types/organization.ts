export interface Organization {
  id: string;
  code: string;
  name: string;
  short_name?: string;
  logo_url?: string;
  address?: string;
  website?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateOrganizationRequest {
  code: string;
  name: string;
  short_name?: string;
  address?: string;
  website?: string;
}

export interface UpdateOrganizationRequest {
  name?: string;
  short_name?: string;
  logo_url?: string;
  address?: string;
  website?: string;
  is_active?: boolean;
}
