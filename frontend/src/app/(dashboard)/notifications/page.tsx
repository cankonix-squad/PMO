"use client";

import { useState, useCallback } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Bell,
  BellOff,
  CheckCheck,
  ChevronDown,
  ChevronRight,
  Filter,
  Loader2,
  Plus,
  RefreshCw,
  RotateCcw,
  X,
} from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { notificationService } from "@/services/notification.service";
import type {
  Notification,
  NotificationChannel,
  NotificationFilter,
  NotificationPriority,
  NotificationStatus,
} from "@/types/notification";
import {
  NOTIFICATION_CHANNEL_LABEL,
  NOTIFICATION_PRIORITY_COLOR,
  NOTIFICATION_STATUS_COLOR,
  NOTIFICATION_STATUS_LABEL,
} from "@/types/notification";

// ── helpers ───────────────────────────────────────────────────────────────────

function formatDate(iso: string): string {
  try {
    return new Intl.DateTimeFormat("id-ID", {
      dateStyle: "short",
      timeStyle: "short",
    }).format(new Date(iso));
  } catch {
    return iso;
  }
}

// ── Badges ────────────────────────────────────────────────────────────────────

function StatusBadge({ status }: { status: NotificationStatus }) {
  return (
    <span
      className={`inline-block rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${NOTIFICATION_STATUS_COLOR[status]}`}
    >
      {NOTIFICATION_STATUS_LABEL[status]}
    </span>
  );
}

function PriorityBadge({ priority }: { priority: NotificationPriority }) {
  return (
    <span
      className={`inline-block rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${NOTIFICATION_PRIORITY_COLOR[priority]}`}
    >
      {priority}
    </span>
  );
}

// ── Summary cards ─────────────────────────────────────────────────────────────

function SummaryCards({ onFilterUnread }: { onFilterUnread: () => void }) {
  const { data, isLoading } = useQuery({
    queryKey: ["notifications", "summary"],
    queryFn: () => notificationService.summary(),
    staleTime: 30_000,
  });

  if (isLoading)
    return (
      <div className="flex gap-3">
        {[1, 2, 3, 4].map((i) => (
          <div
            key={i}
            className="h-16 flex-1 animate-pulse rounded-lg bg-slate-200"
          />
        ))}
      </div>
    );

  if (!data) return null;

  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
      <div className="rounded-lg border border-slate-200 bg-white px-4 py-3 shadow-sm">
        <p className="text-[10px] font-semibold uppercase tracking-wide text-slate-500">Total</p>
        <p className="mt-0.5 text-lg font-bold text-slate-800">{data.total.toLocaleString()}</p>
      </div>
      <button
        type="button"
        onClick={onFilterUnread}
        className="rounded-lg border border-slate-200 bg-white px-4 py-3 shadow-sm text-left hover:border-blue-400 transition-colors"
      >
        <p className="text-[10px] font-semibold uppercase tracking-wide text-slate-500">Belum Dibaca</p>
        <p className={`mt-0.5 text-lg font-bold ${data.unread > 0 ? "text-blue-700" : "text-slate-800"}`}>
          {data.unread.toLocaleString()}
        </p>
      </button>
      <div className="rounded-lg border border-slate-200 bg-white px-4 py-3 shadow-sm">
        <p className="text-[10px] font-semibold uppercase tracking-wide text-slate-500">Pending</p>
        <p className={`mt-0.5 text-lg font-bold ${data.pending > 0 ? "text-yellow-700" : "text-slate-800"}`}>
          {data.pending.toLocaleString()}
        </p>
      </div>
      <div className="rounded-lg border border-slate-200 bg-white px-4 py-3 shadow-sm">
        <p className="text-[10px] font-semibold uppercase tracking-wide text-slate-500">Gagal</p>
        <p className={`mt-0.5 text-lg font-bold ${data.failed > 0 ? "text-rose-700" : "text-slate-800"}`}>
          {data.failed.toLocaleString()}
        </p>
      </div>
    </div>
  );
}

// ── Notification row ──────────────────────────────────────────────────────────

function NotificationRow({
  notif,
  onMarkRead,
  onRetry,
}: {
  notif: Notification;
  onMarkRead: (id: string) => void;
  onRetry: (id: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const isUnread = notif.status !== "READ";

  return (
    <>
      <tr
        className={`border-b border-slate-100 text-sm transition-colors hover:bg-slate-50 ${
          isUnread ? "bg-blue-50/30" : ""
        } ${open ? "bg-slate-50" : ""}`}
      >
        <td className="px-3 py-2 text-xs text-slate-500 whitespace-nowrap">
          {formatDate(notif.created_at)}
        </td>
        <td className="px-3 py-2 max-w-[240px]">
          <p className={`truncate text-sm ${isUnread ? "font-semibold text-slate-800" : "text-slate-700"}`}>
            {notif.subject}
          </p>
          {notif.source_type && (
            <p className="text-[10px] text-slate-400 mt-0.5">{notif.source_type}</p>
          )}
        </td>
        <td className="px-3 py-2">
          <StatusBadge status={notif.status} />
        </td>
        <td className="px-3 py-2">
          <PriorityBadge priority={notif.priority} />
        </td>
        <td className="px-3 py-2 text-xs text-slate-500">
          {NOTIFICATION_CHANNEL_LABEL[notif.channel]}
        </td>
        <td className="px-3 py-2">
          <div className="flex items-center gap-1.5">
            {isUnread && (
              <button
                type="button"
                onClick={() => onMarkRead(notif.id)}
                className="text-xs text-blue-600 hover:text-blue-800 underline"
                title="Tandai dibaca"
              >
                Baca
              </button>
            )}
            {notif.status === "FAILED" && (
              <button
                type="button"
                onClick={() => onRetry(notif.id)}
                className="text-xs text-amber-600 hover:text-amber-800 underline flex items-center gap-0.5"
                title="Coba ulang"
              >
                <RotateCcw className="h-3 w-3" />
                Retry
              </button>
            )}
            <button
              type="button"
              onClick={() => setOpen((v) => !v)}
              className="flex items-center gap-0.5 text-xs text-slate-500 hover:text-slate-700"
              aria-expanded={open}
            >
              {open ? (
                <ChevronDown className="h-3.5 w-3.5" />
              ) : (
                <ChevronRight className="h-3.5 w-3.5" />
              )}
              Detail
            </button>
          </div>
        </td>
      </tr>

      {open && (
        <tr className="bg-slate-50">
          <td colSpan={6} className="px-4 pb-3 pt-1">
            <div className="space-y-2 text-xs text-slate-600">
              <p className="whitespace-pre-wrap break-words">{notif.body}</p>
              <div className="flex flex-wrap gap-4 text-slate-400">
                {notif.source_id && (
                  <span>
                    <span className="font-medium text-slate-600">Source ID: </span>
                    {notif.source_id}
                  </span>
                )}
                {notif.sent_at && (
                  <span>
                    <span className="font-medium text-slate-600">Terkirim: </span>
                    {formatDate(notif.sent_at)}
                  </span>
                )}
                {notif.read_at && (
                  <span>
                    <span className="font-medium text-slate-600">Dibaca: </span>
                    {formatDate(notif.read_at)}
                  </span>
                )}
                {notif.error_message && (
                  <span className="text-rose-500">
                    <span className="font-medium">Error: </span>
                    {notif.error_message}
                  </span>
                )}
              </div>
            </div>
          </td>
        </tr>
      )}
    </>
  );
}

// ── Test notification modal ───────────────────────────────────────────────────

function TestModal({ onClose }: { onClose: () => void }) {
  const queryClient = useQueryClient();
  const [subject, setSubject] = useState("Test Notification");
  const [body, setBody] = useState("Ini adalah test notification dari PMO UAT-004.");
  const [channel, setChannel] = useState<NotificationChannel>("IN_APP");
  const [priority, setPriority] = useState<NotificationPriority>("NORMAL");

  const mutation = useMutation({
    mutationFn: () =>
      notificationService.createTest({ subject, body, channel, priority }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      onClose();
    },
  });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/50 p-4">
      <div className="w-full max-w-md rounded-lg bg-white p-5 shadow-xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-slate-800">Test Notification</h2>
          <button type="button" onClick={onClose}>
            <X className="h-4 w-4 text-slate-400 hover:text-slate-600" />
          </button>
        </div>
        <div className="space-y-3">
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-600">Subject</label>
            <input
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              className="w-full rounded border border-slate-300 px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-400"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-600">Body</label>
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              rows={3}
              className="w-full rounded border border-slate-300 px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-400"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-600">Channel</label>
              <select
                value={channel}
                onChange={(e) => setChannel(e.target.value as NotificationChannel)}
                className="w-full rounded border border-slate-300 px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-400"
              >
                <option value="IN_APP">In-App</option>
                <option value="EMAIL">Email</option>
              </select>
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-600">Priority</label>
              <select
                value={priority}
                onChange={(e) => setPriority(e.target.value as NotificationPriority)}
                className="w-full rounded border border-slate-300 px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-400"
              >
                <option value="LOW">Low</option>
                <option value="NORMAL">Normal</option>
                <option value="HIGH">High</option>
                <option value="URGENT">Urgent</option>
              </select>
            </div>
          </div>
        </div>
        <div className="mt-4 flex justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded border border-slate-300 px-3 py-1.5 text-xs hover:bg-slate-50"
          >
            Batal
          </button>
          <button
            type="button"
            onClick={() => mutation.mutate()}
            disabled={mutation.isPending || !subject.trim() || !body.trim()}
            className="flex items-center gap-1.5 rounded bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700 disabled:opacity-40"
          >
            {mutation.isPending && <Loader2 className="h-3 w-3 animate-spin" />}
            Kirim Test
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Pagination ────────────────────────────────────────────────────────────────

function Pagination({
  page,
  totalPages,
  onPage,
}: {
  page: number;
  totalPages: number;
  onPage: (p: number) => void;
}) {
  if (totalPages <= 1) return null;
  return (
    <div className="flex items-center gap-2 text-sm">
      <button
        type="button"
        disabled={page <= 1}
        onClick={() => onPage(page - 1)}
        className="rounded border border-slate-300 px-2 py-1 text-xs disabled:opacity-40 hover:bg-slate-50"
      >
        ‹ Prev
      </button>
      <span className="text-slate-500">
        {page} / {totalPages}
      </span>
      <button
        type="button"
        disabled={page >= totalPages}
        onClick={() => onPage(page + 1)}
        className="rounded border border-slate-300 px-2 py-1 text-xs disabled:opacity-40 hover:bg-slate-50"
      >
        Next ›
      </button>
    </div>
  );
}

// ── Main page ─────────────────────────────────────────────────────────────────

export default function NotificationsPage() {
  const queryClient = useQueryClient();
  const [filter, setFilter] = useState<NotificationFilter>({
    page: 1,
    page_size: 25,
  });
  const [statusFilter, setStatusFilter] = useState<NotificationStatus | "">("");
  const [channelFilter, setChannelFilter] = useState<NotificationChannel | "">("");
  const [showTestModal, setShowTestModal] = useState(false);

  const { data, isLoading, isError, refetch, isFetching } = useQuery({
    queryKey: ["notifications", filter],
    queryFn: () => notificationService.list(filter),
    placeholderData: (prev) => prev,
    staleTime: 15_000,
  });

  const applyFilters = useCallback(() => {
    setFilter((f) => ({
      ...f,
      page: 1,
      status: statusFilter || undefined,
      channel: channelFilter || undefined,
    }));
  }, [statusFilter, channelFilter]);

  const clearFilters = useCallback(() => {
    setStatusFilter("");
    setChannelFilter("");
    setFilter({ page: 1, page_size: 25 });
  }, []);

  const markReadMutation = useMutation({
    mutationFn: (id: string) => notificationService.markRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  const markAllReadMutation = useMutation({
    mutationFn: () => notificationService.markAllRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  const retryMutation = useMutation({
    mutationFn: (id: string) => notificationService.retry(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  const filterUnread = useCallback(() => {
    setStatusFilter("");
    setFilter((f) => ({ ...f, page: 1, unread_only: true, status: undefined }));
  }, []);

  const notifications: Notification[] = data?.data ?? [];
  const meta = data?.meta;
  const hasFilters = filter.status || filter.channel || filter.unread_only;

  return (
    <DashboardLayout title="Notifikasi">
      <div className="mx-auto max-w-screen-xl space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Bell className="h-5 w-5 text-slate-600" />
          <h1 className="text-lg font-bold text-slate-800">Notifikasi</h1>
          {isFetching && (
            <Loader2 className="h-4 w-4 animate-spin text-slate-400" />
          )}
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => refetch()}
            className="flex items-center gap-1.5 rounded border border-slate-300 px-3 py-1.5 text-xs hover:bg-slate-50"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            Muat Ulang
          </button>
          <button
            type="button"
            onClick={() => markAllReadMutation.mutate()}
            disabled={markAllReadMutation.isPending}
            className="flex items-center gap-1.5 rounded border border-slate-300 px-3 py-1.5 text-xs hover:bg-slate-50 disabled:opacity-40"
          >
            <CheckCheck className="h-3.5 w-3.5" />
            Tandai Semua Dibaca
          </button>
          <button
            type="button"
            onClick={() => setShowTestModal(true)}
            className="flex items-center gap-1.5 rounded bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700"
          >
            <Plus className="h-3.5 w-3.5" />
            Test Notifikasi
          </button>
        </div>
      </div>

      {/* Summary */}
      <SummaryCards onFilterUnread={filterUnread} />

      {/* Filter bar */}
      <div className="rounded-lg border border-slate-200 bg-white p-3 shadow-sm">
        <div className="mb-2 flex items-center gap-1.5 text-xs font-semibold text-slate-500">
          <Filter className="h-3.5 w-3.5" />
          Filter
          {hasFilters && (
            <button
              type="button"
              onClick={clearFilters}
              className="ml-1 flex items-center gap-0.5 text-rose-500 hover:text-rose-700"
            >
              <X className="h-3 w-3" />
              Clear
            </button>
          )}
        </div>
        <div className="grid gap-2 sm:grid-cols-3">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as NotificationStatus | "")}
            className="rounded border border-slate-300 px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-blue-400"
          >
            <option value="">Semua Status</option>
            <option value="PENDING">Pending</option>
            <option value="SENT">Terkirim</option>
            <option value="FAILED">Gagal</option>
            <option value="READ">Dibaca</option>
          </select>
          <select
            value={channelFilter}
            onChange={(e) => setChannelFilter(e.target.value as NotificationChannel | "")}
            className="rounded border border-slate-300 px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-blue-400"
          >
            <option value="">Semua Channel</option>
            <option value="IN_APP">In-App</option>
            <option value="EMAIL">Email</option>
          </select>
          <div className="flex items-center gap-2">
            <label className="flex items-center gap-1.5 text-xs text-slate-600 cursor-pointer">
              <input
                type="checkbox"
                checked={filter.unread_only === true}
                onChange={(e) =>
                  setFilter((f) => ({
                    ...f,
                    page: 1,
                    unread_only: e.target.checked ? true : undefined,
                  }))
                }
                className="rounded"
              />
              Belum dibaca saja
            </label>
          </div>
        </div>
        <div className="mt-2 flex justify-end">
          <button
            type="button"
            onClick={applyFilters}
            className="rounded bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700"
          >
            Terapkan Filter
          </button>
        </div>
      </div>

      {/* Table */}
      <div className="rounded-lg border border-slate-200 bg-white shadow-sm overflow-hidden">
        {isLoading ? (
          <div className="flex items-center justify-center gap-2 p-12 text-slate-400">
            <Loader2 className="h-5 w-5 animate-spin" />
            <span className="text-sm">Memuat notifikasi…</span>
          </div>
        ) : isError ? (
          <div className="flex flex-col items-center justify-center gap-2 p-12 text-rose-500">
            <BellOff className="h-8 w-8 opacity-50" />
            <p className="text-sm font-medium">Gagal memuat notifikasi</p>
            <button
              type="button"
              onClick={() => refetch()}
              className="text-xs underline hover:no-underline"
            >
              Coba lagi
            </button>
          </div>
        ) : notifications.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 p-12 text-slate-400">
            <Bell className="h-8 w-8 opacity-30" />
            <p className="text-sm">Tidak ada notifikasi</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead className="border-b border-slate-200 bg-slate-50 text-[10px] font-semibold uppercase tracking-wide text-slate-500">
                <tr>
                  <th className="px-3 py-2">Waktu</th>
                  <th className="px-3 py-2">Subject</th>
                  <th className="px-3 py-2">Status</th>
                  <th className="px-3 py-2">Prioritas</th>
                  <th className="px-3 py-2">Channel</th>
                  <th className="px-3 py-2">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {notifications.map((n) => (
                  <NotificationRow
                    key={n.id}
                    notif={n}
                    onMarkRead={(id) => markReadMutation.mutate(id)}
                    onRetry={(id) => retryMutation.mutate(id)}
                  />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Pagination */}
      {meta && (
        <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-slate-500">
          <span>
            {meta.total.toLocaleString()} total · halaman {meta.page}/
            {meta.total_pages}
          </span>
          <Pagination
            page={meta.page}
            totalPages={meta.total_pages}
            onPage={(p) => setFilter((f) => ({ ...f, page: p }))}
          />
        </div>
      )}

      {showTestModal && <TestModal onClose={() => setShowTestModal(false)} />}
      </div>
    </DashboardLayout>
  );
}
