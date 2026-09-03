import api from "@/lib/axios";
import { ApiResponse } from "@/types/api";
import { CreateRoleRequest, Role, UpdateRoleRequest } from "@/types/rbac";

const rbacService = {
  listRoles: async (): Promise<Role[]> => {
    const res = await api.get<ApiResponse<Role[]>>("/roles");
    return res.data.data ?? [];
  },

  createRole: async (req: CreateRoleRequest): Promise<Role> => {
    const res = await api.post<ApiResponse<Role>>("/roles", req);
    return res.data.data!;
  },

  updateRole: async (id: string, req: UpdateRoleRequest): Promise<Role> => {
    const res = await api.put<ApiResponse<Role>>(`/roles/${id}`, req);
    return res.data.data!;
  },

  deleteRole: async (id: string): Promise<void> => {
    await api.delete(`/roles/${id}`);
  },
};

export default rbacService;
