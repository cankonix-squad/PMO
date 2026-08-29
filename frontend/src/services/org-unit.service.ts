// Org unit service — P1-008

import api from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  OrgUnit,
  CreateOrgUnitRequest,
  UpdateOrgUnitRequest,
} from "@/types/org-unit";

const BASE = "/org-units";

export const orgUnitService = {
  /** List all org units for the current organization. */
  listOrgUnits: async (includeInactive = false): Promise<OrgUnit[]> => {
    const { data } = await api.get<ApiResponse<OrgUnit[]>>(BASE, {
      params: includeInactive ? { include_inactive: "true" } : {},
    });
    return data.data;
  },

  /** Get a single org unit by ID. */
  getOrgUnit: async (id: string): Promise<OrgUnit> => {
    const { data } = await api.get<{ data: OrgUnit }>(`${BASE}/${id}`);
    return data.data;
  },

  /** Create a new org unit. */
  createOrgUnit: async (req: CreateOrgUnitRequest): Promise<OrgUnit> => {
    const { data } = await api.post<{ data: OrgUnit }>(BASE, req);
    return data.data;
  },

  /** Update an existing org unit. */
  updateOrgUnit: async (
    id: string,
    req: UpdateOrgUnitRequest
  ): Promise<OrgUnit> => {
    const { data } = await api.put<{ data: OrgUnit }>(`${BASE}/${id}`, req);
    return data.data;
  },

  /** Soft-delete an org unit. */
  deleteOrgUnit: async (id: string): Promise<void> => {
    await api.delete(`${BASE}/${id}`);
  },
};
