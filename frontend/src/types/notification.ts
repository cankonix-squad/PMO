// Notification Delivery Foundation — Type definitions
// Mirrors backend/internal/core/notification/model.go

export type NotificationStatus = "PENDING" | "SENT" | "FAILED" | "READ";
export type NotificationChannel = "IN_APP" | "EMAIL";
export type NotificationPriority = "LOW" | "NORMAL" | "HIGH" | "URGENT";

export const NOTIFICATION_STATUS_LABEL: Record<NotificationStatus, string> = {
  PENDING: "Pending",
  SENT: "Terkirim",
  FAILED: "Gagal",
  READ: "Dibaca",
};

export const NOTIFICATION_STATUS_COLOR: Record<NotificationStatus, string> = {
  PENDING: "bg-yellow-100 text-yellow-800",
  SENT: "bg-blue-100 text-blue-800",
  FAILED: "bg-rose-100 text-rose-800",
  READ: "bg-slate-100 text-slate-500",
};

export const NOTIFICATION_PRIORITY_COLOR: Record<NotificationPriority, string> = {
  LOW: "bg-slate-100 text-slate-500",
  NORMAL: "bg-blue-100 text-blue-700",
  HIGH: "bg-orange-100 text-orange-700",
  URGENT: "bg-rose-100 text-rose-800",
};

export const NOTIFICATION_CHANNEL_LABEL: Record<NotificationChannel, string> = {
  IN_APP: "In-App",
  EMAIL: "Email",
};

export interface Notification {
  id: string;
  organization_id: string;
  recipient_user_id?: string | null;
  channel: NotificationChannel;
  status: NotificationStatus;
  priority: NotificationPriority;
  subject: string;
  body: string;
  source_type?: string;
  source_id?: string;
  error_message?: string;
  sent_at?: string | null;
  read_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface NotificationSummary {
  total: number;
  unread: number;
  pending: number;
  failed: number;
}

export interface NotificationFilter {
  status?: NotificationStatus | "";
  channel?: NotificationChannel | "";
  priority?: NotificationPriority | "";
  source_type?: string;
  unread_only?: boolean;
  page?: number;
  page_size?: number;
}

export interface CreateTestNotificationRequest {
  subject: string;
  body: string;
  channel?: NotificationChannel;
  priority?: NotificationPriority;
  source_type?: string;
}
