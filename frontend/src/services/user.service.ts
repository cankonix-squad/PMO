import api from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  CreateUserRequest,
  UpdateUserRequest,
  UserListFilter,
  UserProfile,
} from "@/types/user";

export const userService = {
  list: (params?: UserListFilter) =>
    api.get<ApiResponse<UserProfile[]>>("/users", { params }),

  get: (id: string) =>
    api.get<ApiResponse<UserProfile>>(`/users/${id}`),

  create: (data: CreateUserRequest) =>
    api.post<ApiResponse<UserProfile>>("/users", data),

  update: (id: string, data: UpdateUserRequest) =>
    api.put<ApiResponse<UserProfile>>(`/users/${id}`, data),

  deactivate: (id: string) =>
    api.post<ApiResponse<null>>(`/users/${id}/deactivate`),
};
