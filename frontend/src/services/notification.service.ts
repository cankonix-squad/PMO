import { api } from "@/lib/axios";
import type { ApiResponse, PaginationMeta } from "@/types/api";
import type {
  CreateTestNotificationRequest,
  Notification,
  NotificationFilter,
  NotificationSummary,
} from "@/types/notification";

export interface NotificationListResponse {
  data: Notification[];
  meta: PaginationMeta;
}

export const notificationService = {
  list: async (filter?: NotificationFilter): Promise<NotificationListResponse> => {
    const params: Record<string, string | number | boolean> = {};
    if (filter?.status) params.status = filter.status;
    if (filter?.channel) params.channel = filter.channel;
    if (filter?.priority) params.priority = filter.priority;
    if (filter?.source_type) params.source_type = filter.source_type;
    if (filter?.unread_only) params.unread_only = true;
    if (filter?.page) params.page = filter.page;
    if (filter?.page_size) params.page_size = filter.page_size;

    const res = await api.get<ApiResponse<Notification[]>>("/notifications", { params });
    return {
      data: res.data.data,
      meta: res.data.meta ?? { page: 1, page_size: 20, total: 0, total_pages: 0 },
    };
  },

  getById: async (id: string): Promise<Notification> => {
    const res = await api.get<ApiResponse<Notification>>(`/notifications/${id}`);
    return res.data.data;
  },

  summary: async (): Promise<NotificationSummary> => {
    const res = await api.get<ApiResponse<NotificationSummary>>("/notifications/summary");
    return res.data.data;
  },

  markRead: async (id: string): Promise<void> => {
    await api.patch(`/notifications/${id}/read`);
  },

  markAllRead: async (): Promise<void> => {
    await api.patch("/notifications/read-all");
  },

  retry: async (id: string): Promise<void> => {
    await api.post(`/notifications/${id}/retry`);
  },

  createTest: async (payload: CreateTestNotificationRequest): Promise<void> => {
    await api.post("/notifications/test", payload);
  },
};
