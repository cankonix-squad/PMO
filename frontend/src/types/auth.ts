export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in?: number;
  expires_at?: string;
}

export interface LoginResponse extends TokenPair {
  user: UserInfo;
}

export interface UserInfo {
  id: string;
  organization_id: string;
  first_name: string;
  last_name: string;
  email: string;
  job_title: string | null;
  avatar_url: string | null;
  is_active: boolean;
  must_change_pwd: boolean;
  permissions: string[];
  roles: string[];
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  new_password: string;
}
