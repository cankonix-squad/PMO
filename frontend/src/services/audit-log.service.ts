import { api } from "@/lib/axios";
import type { ApiResponse, PaginationMeta } from "@/types/api";
import type { AuditLog, AuditLogFilter, AuditLogSummary } from "@/types/audit-log";

export interface AuditLogListResponse {
  data: AuditLog[];
  meta: PaginationMeta;
}

export const auditLogService = {
  /**
   * List audit logs with optional filters. Returns paginated results.
   */
  list: async (filter?: AuditLogFilter): Promise<AuditLogListResponse> => {
    const params: Record<string, string | number> = {};
    if (filter?.action) params.action = filter.action;
    if (filter?.entity_type) params.entity_type = filter.entity_type;
    if (filter?.entity_id) params.entity_id = filter.entity_id;
    if (filter?.actor_id) params.actor_id = filter.actor_id;
    if (filter?.search) params.search = filter.search;
    if (filter?.date_from) params.date_from = filter.date_from;
    if (filter?.date_to) params.date_to = filter.date_to;
    if (filter?.page) params.page = filter.page;
    if (filter?.page_size) params.page_size = filter.page_size;

    const res = await api.get<ApiResponse<AuditLog[]>>("/audit-logs", { params });
    return {
      data: res.data.data,
      meta: res.data.meta ?? { page: 1, page_size: 20, total: 0, total_pages: 0 },
    };
  },

  /**
   * Get a single audit log entry by ID.
   */
  getById: async (id: string): Promise<AuditLog> => {
    const res = await api.get<ApiResponse<AuditLog>>(`/audit-logs/${id}`);
    return res.data.data;
  },

  /**
   * Get aggregate summary (total events, unique actors, top actions/entities).
   */
  summary: async (): Promise<AuditLogSummary> => {
    const res = await api.get<ApiResponse<AuditLogSummary>>("/audit-logs/summary");
    return res.data.data;
  },

  /**
   * Trigger a CSV download for the current filter (max 1000 rows).
   * Opens a blob download in the browser.
   */
  exportCsv: async (filter?: AuditLogFilter): Promise<void> => {
    const params: Record<string, string | number> = {};
    if (filter?.action) params.action = filter.action;
    if (filter?.entity_type) params.entity_type = filter.entity_type;
    if (filter?.entity_id) params.entity_id = filter.entity_id;
    if (filter?.actor_id) params.actor_id = filter.actor_id;
    if (filter?.search) params.search = filter.search;
    if (filter?.date_from) params.date_from = filter.date_from;
    if (filter?.date_to) params.date_to = filter.date_to;

    const res = await api.get("/audit-logs/export", {
      params,
      responseType: "blob",
    });

    const url = URL.createObjectURL(res.data as Blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `audit-logs-${new Date().toISOString().slice(0, 10)}.csv`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  },
};
