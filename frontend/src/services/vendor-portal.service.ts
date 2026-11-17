import api from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  CreateProgressReportRequest,
  CreateVendorAccountRequest,
  CreatedVendorAccount,
  ReviewRequest,
  UploadVendorDocumentInput,
  VendorAccount,
  VendorContract,
  VendorDocument,
  VendorProgressReport,
  VendorSubmissions,
} from "@/types/vendor-portal";

/** Triggers a browser download of a Blob. */
function saveBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

/** Contractor side — only the caller's own contracts. */
export const vendorPortalService = {
  myContracts(): Promise<ApiResponse<VendorContract[]>> {
    return api.get("/vendor-portal/contracts").then((r) => r.data);
  },
  myContract(id: string): Promise<ApiResponse<VendorContract>> {
    return api.get(`/vendor-portal/contracts/${id}`).then((r) => r.data);
  },
  listReports(contractId: string): Promise<ApiResponse<VendorProgressReport[]>> {
    return api.get(`/vendor-portal/contracts/${contractId}/progress-reports`).then((r) => r.data);
  },
  createReport(contractId: string, payload: CreateProgressReportRequest): Promise<ApiResponse<VendorProgressReport>> {
    return api.post(`/vendor-portal/contracts/${contractId}/progress-reports`, payload).then((r) => r.data);
  },
  listDocuments(contractId: string): Promise<ApiResponse<VendorDocument[]>> {
    return api.get(`/vendor-portal/contracts/${contractId}/documents`).then((r) => r.data);
  },
  uploadDocument(contractId: string, input: UploadVendorDocumentInput): Promise<ApiResponse<VendorDocument>> {
    const form = new FormData();
    form.append("file", input.file);
    form.append("category", input.category);
    if (input.name) form.append("name", input.name);
    if (input.notes) form.append("notes", input.notes);
    if (input.period_year && input.period_month) {
      form.append("period_year", String(input.period_year));
      form.append("period_month", String(input.period_month));
    }
    return api
      .post(`/vendor-portal/contracts/${contractId}/documents`, form, {
        headers: { "Content-Type": "multipart/form-data" },
      })
      .then((r) => r.data);
  },
  async downloadOwnDocument(doc: VendorDocument): Promise<void> {
    const r = await api.get(`/vendor-portal/documents/${doc.id}/download`, { responseType: "blob" });
    saveBlob(r.data as Blob, doc.name);
  },
};

/** Balai review queue. */
export const vendorSubmissionService = {
  list(params: { status?: string; project_id?: string } = {}): Promise<ApiResponse<VendorSubmissions>> {
    return api.get("/vendor-submissions", { params }).then((r) => r.data);
  },
  reviewReport(id: string, payload: ReviewRequest): Promise<ApiResponse<VendorProgressReport>> {
    return api.post(`/vendor-submissions/progress-reports/${id}/review`, payload).then((r) => r.data);
  },
  reviewDocument(id: string, payload: ReviewRequest): Promise<ApiResponse<VendorDocument>> {
    return api.post(`/vendor-submissions/documents/${id}/review`, payload).then((r) => r.data);
  },
  async downloadDocument(doc: VendorDocument): Promise<void> {
    const r = await api.get(`/vendor-submissions/documents/${doc.id}/download`, { responseType: "blob" });
    saveBlob(r.data as Blob, doc.name);
  },
};

/** Contractor accounts (admin). */
export const vendorAccountService = {
  list(vendorId: string): Promise<ApiResponse<VendorAccount[]>> {
    return api.get(`/vendors/${vendorId}/accounts`).then((r) => r.data);
  },
  create(vendorId: string, payload: CreateVendorAccountRequest): Promise<ApiResponse<CreatedVendorAccount>> {
    return api.post(`/vendors/${vendorId}/accounts`, payload).then((r) => r.data);
  },
};

export function apiErrorMessage(err: unknown, fallback: string): string {
  return (err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? fallback;
}
