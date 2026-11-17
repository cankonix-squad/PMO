import api from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  KpbuAvailabilityPayment,
  KpbuAvailabilityPaymentRequest,
  KpbuChecklist,
  KpbuFinanceSummary,
  KpbuFinancing,
  KpbuFinancingRequest,
  KpbuGovernmentSupport,
  KpbuGovernmentSupportRequest,
  KpbuDocument,
  KpbuDocumentRequirement,
  KpbuDocumentUpload,
  KpbuGuarantee,
  KpbuGuaranteeRequest,
  KpbuRequirementRequest,
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

  // K1.5 — document checklist
  checklist(id: string): Promise<ApiResponse<KpbuChecklist>> {
    return api.get(`${base}/${id}/checklist`).then((r) => r.data);
  },
  uploadDocument(id: string, input: KpbuDocumentUpload): Promise<ApiResponse<KpbuDocument>> {
    const fd = new FormData();
    fd.append("file", input.file);
    if (input.requirement_id) fd.append("requirement_id", input.requirement_id);
    if (input.name) fd.append("name", input.name);
    if (input.document_number) fd.append("document_number", input.document_number);
    if (input.document_date) fd.append("document_date", input.document_date);
    if (input.notes) fd.append("notes", input.notes);
    return api
      .post(`${base}/${id}/documents`, fd, { headers: { "Content-Type": "multipart/form-data" } })
      .then((r) => r.data);
  },
  reviewDocument(
    id: string,
    docId: string,
    decision: "APPROVE" | "REJECT",
    notes: string,
  ): Promise<ApiResponse<KpbuDocument>> {
    return api.post(`${base}/${id}/documents/${docId}/review`, { decision, notes }).then((r) => r.data);
  },
  deleteDocument(id: string, docId: string): Promise<void> {
    return api.delete(`${base}/${id}/documents/${docId}`).then(() => undefined);
  },
  async downloadDocument(id: string, doc: KpbuDocument): Promise<void> {
    const res = await api.get(`${base}/${id}/documents/${doc.id}/download`, { responseType: "blob" });
    const url = URL.createObjectURL(res.data as Blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = doc.name;
    a.click();
    URL.revokeObjectURL(url);
  },

  // K1.5 — guarantees
  listGuarantees(id: string): Promise<ApiResponse<KpbuGuarantee[]>> {
    return api.get(`${base}/${id}/guarantees`).then((r) => r.data);
  },
  createGuarantee(id: string, payload: KpbuGuaranteeRequest): Promise<ApiResponse<KpbuGuarantee>> {
    return api.post(`${base}/${id}/guarantees`, payload).then((r) => r.data);
  },
  updateGuarantee(id: string, gid: string, payload: KpbuGuaranteeRequest): Promise<ApiResponse<KpbuGuarantee>> {
    return api.put(`${base}/${id}/guarantees/${gid}`, payload).then((r) => r.data);
  },
  deleteGuarantee(id: string, gid: string): Promise<void> {
    return api.delete(`${base}/${id}/guarantees/${gid}`).then(() => undefined);
  },

  // K2 — finance
  financeSummary(id: string): Promise<ApiResponse<KpbuFinanceSummary>> {
    return api.get(`${base}/${id}/finance-summary`).then((r) => r.data);
  },
  listFinancings(id: string): Promise<ApiResponse<KpbuFinancing[]>> {
    return api.get(`${base}/${id}/financings`).then((r) => r.data);
  },
  saveFinancing(id: string, rid: string | null, payload: KpbuFinancingRequest): Promise<ApiResponse<KpbuFinancing>> {
    return (rid ? api.put(`${base}/${id}/financings/${rid}`, payload) : api.post(`${base}/${id}/financings`, payload)).then((r) => r.data);
  },
  deleteFinancing(id: string, rid: string): Promise<void> {
    return api.delete(`${base}/${id}/financings/${rid}`).then(() => undefined);
  },
  listSupports(id: string): Promise<ApiResponse<KpbuGovernmentSupport[]>> {
    return api.get(`${base}/${id}/government-supports`).then((r) => r.data);
  },
  saveSupport(id: string, rid: string | null, payload: KpbuGovernmentSupportRequest): Promise<ApiResponse<KpbuGovernmentSupport>> {
    return (rid ? api.put(`${base}/${id}/government-supports/${rid}`, payload) : api.post(`${base}/${id}/government-supports`, payload)).then((r) => r.data);
  },
  deleteSupport(id: string, rid: string): Promise<void> {
    return api.delete(`${base}/${id}/government-supports/${rid}`).then(() => undefined);
  },
  listPayments(id: string): Promise<ApiResponse<KpbuAvailabilityPayment[]>> {
    return api.get(`${base}/${id}/availability-payments`).then((r) => r.data);
  },
  savePayment(id: string, rid: string | null, payload: KpbuAvailabilityPaymentRequest): Promise<ApiResponse<KpbuAvailabilityPayment>> {
    return (rid ? api.put(`${base}/${id}/availability-payments/${rid}`, payload) : api.post(`${base}/${id}/availability-payments`, payload)).then((r) => r.data);
  },
  deletePayment(id: string, rid: string): Promise<void> {
    return api.delete(`${base}/${id}/availability-payments/${rid}`).then(() => undefined);
  },

  // K1.5 — checklist master
  listRequirements(includeInactive = false): Promise<ApiResponse<KpbuDocumentRequirement[]>> {
    return api
      .get("/kpbu/document-requirements", { params: includeInactive ? { include_inactive: "true" } : {} })
      .then((r) => r.data);
  },
  createRequirement(payload: KpbuRequirementRequest): Promise<ApiResponse<KpbuDocumentRequirement>> {
    return api.post("/kpbu/document-requirements", payload).then((r) => r.data);
  },
  updateRequirement(reqId: string, payload: KpbuRequirementRequest): Promise<ApiResponse<KpbuDocumentRequirement>> {
    return api.put(`/kpbu/document-requirements/${reqId}`, payload).then((r) => r.data);
  },
  deleteRequirement(reqId: string): Promise<void> {
    return api.delete(`/kpbu/document-requirements/${reqId}`).then(() => undefined);
  },
};

/** Extracts the backend error message from an axios error. */
export function kpbuErrorMessage(err: unknown, fallback: string): string {
  return (
    (err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? fallback
  );
}
