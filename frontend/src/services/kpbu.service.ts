import api from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  KpbuListFilter,
  KpbuParty,
  KpbuPartyRequest,
  KpbuProject,
  KpbuProjectDetail,
  KpbuProjectRequest,
  KpbuStageMilestone,
  KpbuStageMilestoneRequest,
  KpbuTransitionRequest,
} from "@/types/kpbu";

const base = "/kpbu/projects";

export const kpbuService = {
  list(filter: KpbuListFilter = {}): Promise<ApiResponse<KpbuProject[]>> {
    return api.get(base, { params: filter }).then((r) => r.data);
  },
  get(id: string): Promise<ApiResponse<KpbuProjectDetail>> {
    return api.get(`${base}/${id}`).then((r) => r.data);
  },
  create(payload: KpbuProjectRequest): Promise<ApiResponse<KpbuProject>> {
    return api.post(base, payload).then((r) => r.data);
  },
  update(id: string, payload: KpbuProjectRequest): Promise<ApiResponse<KpbuProject>> {
    return api.put(`${base}/${id}`, payload).then((r) => r.data);
  },
  delete(id: string): Promise<void> {
    return api.delete(`${base}/${id}`).then(() => undefined);
  },
  transition(id: string, payload: KpbuTransitionRequest): Promise<ApiResponse<KpbuProject>> {
    return api.post(`${base}/${id}/transition`, payload).then((r) => r.data);
  },
  upsertStageMilestone(
    id: string,
    payload: KpbuStageMilestoneRequest,
  ): Promise<ApiResponse<KpbuStageMilestone>> {
    return api.put(`${base}/${id}/stage-milestones`, payload).then((r) => r.data);
  },
  createParty(id: string, payload: KpbuPartyRequest): Promise<ApiResponse<KpbuParty>> {
    return api.post(`${base}/${id}/parties`, payload).then((r) => r.data);
  },
  updateParty(id: string, partyId: string, payload: KpbuPartyRequest): Promise<ApiResponse<KpbuParty>> {
    return api.put(`${base}/${id}/parties/${partyId}`, payload).then((r) => r.data);
  },
  deleteParty(id: string, partyId: string): Promise<void> {
    return api.delete(`${base}/${id}/parties/${partyId}`).then(() => undefined);
  },
};

/** Extracts the backend error message from an axios error. */
export function kpbuErrorMessage(err: unknown, fallback: string): string {
  return (
    (err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? fallback
  );
}
