import api from "@/lib/axios";
import { ApiResponse } from "@/types/api";
import {
  Organization,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
} from "@/types/organization";

const organizationService = {
  list: async (): Promise<Organization[]> => {
    const res = await api.get<ApiResponse<Organization[]>>("/organizations");
    return res.data.data ?? [];
  },

  get: async (id: string): Promise<Organization> => {
    const res = await api.get<ApiResponse<Organization>>(`/organizations/${id}`);
    return res.data.data!;
  },

  create: async (req: CreateOrganizationRequest): Promise<Organization> => {
    const res = await api.post<ApiResponse<Organization>>("/organizations", req);
    return res.data.data!;
  },

  update: async (id: string, req: UpdateOrganizationRequest): Promise<Organization> => {
    const res = await api.put<ApiResponse<Organization>>(`/organizations/${id}`, req);
    return res.data.data!;
  },
};

export default organizationService;
