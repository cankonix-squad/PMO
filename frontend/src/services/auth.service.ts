import api from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  LoginRequest,
  LoginResponse,
  TokenPair,
  UserInfo,
  ChangePasswordRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
} from "@/types/auth";

export const authService = {
  login: (data: LoginRequest) =>
    api.post<ApiResponse<LoginResponse>>(
      "/auth/login",
      data
    ),

  me: () =>
    api.get<ApiResponse<UserInfo>>("/auth/me"),

  refresh: (refreshToken: string) =>
    api.post<ApiResponse<TokenPair>>("/auth/refresh", {
      refresh_token: refreshToken,
    }),

  logout: () =>
    api.post<ApiResponse<null>>("/auth/logout"),

  logoutAll: () =>
    api.post<ApiResponse<null>>("/auth/logout-all"),

  changePassword: (data: ChangePasswordRequest) =>
    api.post<ApiResponse<null>>("/auth/change-password", data),

  forgotPassword: (data: ForgotPasswordRequest) =>
    api.post<ApiResponse<null>>("/auth/forgot-password", data),

  resetPassword: (data: ResetPasswordRequest) =>
    api.post<ApiResponse<null>>("/auth/reset-password", data),
};
