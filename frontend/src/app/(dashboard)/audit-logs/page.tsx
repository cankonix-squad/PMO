"use client";

import { useState, useCallback } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Activity,
  ChevronDown,
  ChevronRight,
  Download,
  Filter,
  Loader2,
  RefreshCw,
  Search,
  ShieldAlert,
  X,
} from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { auditLogService } from "@/services/audit-log.service";
import type { AuditLog, AuditLogFilter } from "@/types/audit-log";

// ── Helpers ───────────────────────────────────────────────────────────────────

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

/** Safely parse a JSON string for display. Returns null if invalid. */
function safeParseJson(raw: string | null | undefined): unknown {
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return raw;
  }
}

/** Pretty-print a parsed value, capped at reasonable depth. */
function JsonDisplay({ value }: { value: unknown }) {
  if (value === null || value === undefined) return null;
  const text =
    typeof value === "string" ? value : JSON.stringify(value, null, 2);
  return (
    <pre className="max-h-48 overflow-auto whitespace-pre-wrap break-all rounded bg-slate-800 p-2 text-xs text-slate-100">
      {text}
    </pre>
  );
}

// ── Action badge colours ───────────────────────────────────────────────────────

const ACTION_COLORS: Record<string, string> = {
  create: "bg-emerald-100 text-emerald-800",
  created: "bg-emerald-100 text-emerald-800",
  update: "bg-blue-100 text-blue-800",
  updated: "bg-blue-100 text-blue-800",
  delete: "bg-rose-100 text-rose-800",
  deleted: "bg-rose-100 text-rose-800",
  login: "bg-indigo-100 text-indigo-800",
  logout: "bg-slate-100 text-slate-700",
  approve: "bg-amber-100 text-amber-800",
  approved: "bg-amber-100 text-amber-800",
  export: "bg-purple-100 text-purple-800",
  view: "bg-slate-100 text-slate-600",
};

function actionBadgeClass(action: string): string {
  const key = action.split(".").pop()?.toLowerCase() ?? action.toLowerCase();
  return ACTION_COLORS[key] ?? "bg-slate-100 text-slate-700";
}

function ActionBadge({ action }: { action: string }) {
  return (
    <span
      className={`inline-block rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${actionBadgeClass(action)}`}
    >
      {action}
    </span>
  );
}

function EntityBadge({ entity }: { entity: string }) {
  return (
    <span className="inline-block rounded bg-sky-100 px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-sky-800">
      {entity}
    </span>
  );
}

// ── Row with expandable detail ────────────────────────────────────────────────

function AuditRow({ log }: { log: AuditLog }) {
  const [open, setOpen] = useState(false);

  const oldValues = safeParseJson(log.old_values);
  const newValues = safeParseJson(log.new_values);
  const hasDetail =
    oldValues !== null ||
    newValues !== null ||
    log.ip_address ||
    log.user_agent ||
    log.request_id;

  return (
    <>
      <tr
        className={`border-b border-slate-100 text-sm transition-colors hover:bg-slate-50 ${open ? "bg-slate-50" : ""}`}
      >
        <td className="px-3 py-2 text-xs text-slate-500 whitespace-nowrap">
          {formatDate(log.created_at)}
        </td>
        <td className="px-3 py-2 text-xs text-slate-700 max-w-[160px] truncate">
          {log.actor_email ?? <span className="text-slate-400">system</span>}
        </td>
        <td className="px-3 py-2">
          <ActionBadge action={log.action} />
        </td>
        <td className="px-3 py-2">
          <EntityBadge entity={log.entity_type} />
        </td>
        <td className="px-3 py-2 text-xs text-slate-600 max-w-[180px] truncate">
          {log.entity_label || log.entity_id || (
            <span className="text-slate-400">—</span>
          )}
        </td>
        <td className="px-3 py-2">
          {hasDetail ? (
            <button
              type="button"
              onClick={() => setOpen((v) => !v)}
              className="flex items-center gap-0.5 text-xs text-blue-600 hover:text-blue-800"
              aria-expanded={open}
            >
              {open ? (
                <ChevronDown className="h-3.5 w-3.5" />
              ) : (
                <ChevronRight className="h-3.5 w-3.5" />
              )}
              Detail
            </button>
          ) : (
            <span className="text-xs text-slate-400">—</span>
          )}
        </td>
      </tr>

      {open && hasDetail && (
        <tr className="bg-slate-50">
          <td colSpan={6} className="px-4 pb-3 pt-1">
            <div className="grid gap-3 text-xs sm:grid-cols-2">
              {oldValues !== null && (
                <div>
                  <p className="mb-1 font-semibold text-slate-500 uppercase tracking-wide">
                    Old Values
                  </p>
                  <JsonDisplay value={oldValues} />
                </div>
              )}
              {newValues !== null && (
                <div>
                  <p className="mb-1 font-semibold text-slate-500 uppercase tracking-wide">
                    New Values
                  </p>
                  <JsonDisplay value={newValues} />
                </div>
              )}
              {(log.ip_address || log.user_agent || log.request_id) && (
                <div className="sm:col-span-2 flex flex-wrap gap-4 text-slate-500">
                  {log.ip_address && (
                    <span>
                      <span className="font-medium text-slate-600">IP: </span>
                      {log.ip_address}
                    </span>
                  )}
                  {log.request_id && (
                    <span>
                      <span className="font-medium text-slate-600">
                        Request:{" "}
                      </span>
                      {log.request_id}
                    </span>
                  )}
                  {log.user_agent && (
                    <span className="break-all">
                      <span className="font-medium text-slate-600">UA: </span>
                      {log.user_agent}
                    </span>
                  )}
                </div>
              )}
            </div>
          </td>
        </tr>
      )}
    </>
  );
}

// ── Summary cards ─────────────────────────────────────────────────────────────

function SummaryCards() {
  const { data, isLoading } = useQuery({
    queryKey: ["audit-logs", "summary"],
    queryFn: () => auditLogService.summary(),
    staleTime: 60_000,
  });

  if (isLoading)
    return (
      <div className="flex gap-3">
        {[1, 2, 3].map((i) => (
          <div
            key={i}
            className="h-16 flex-1 animate-pulse rounded-lg bg-slate-200"
          />
        ))}
      </div>
    );

  if (!data) return null;

  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
      <StatCard label="Total Events" value={data.total_events.toLocaleString()} />
      <StatCard label="Unique Actors" value={data.unique_actors.toLocaleString()} />
      {data.top_actions[0] && (
        <StatCard
          label="Top Action"
          value={data.top_actions[0].action}
          sub={`${data.top_actions[0].count.toLocaleString()}×`}
        />
      )}
      {data.top_entities[0] && (
        <StatCard
          label="Top Entity"
          value={data.top_entities[0].entity_type}
          sub={`${data.top_entities[0].count.toLocaleString()}×`}
        />
      )}
    </div>
  );
}

function StatCard({
  label,
  value,
  sub,
}: {
  label: string;
  value: string;
  sub?: string;
}) {
  return (
    <div className="rounded-lg border border-slate-200 bg-white px-4 py-3 shadow-sm">
      <p className="text-[10px] font-semibold uppercase tracking-wide text-slate-500">
        {label}
      </p>
      <p className="mt-0.5 truncate text-lg font-bold text-slate-800">
        {value}
      </p>
      {sub && <p className="text-xs text-slate-400">{sub}</p>}
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

export default function AuditLogsPage() {
  const [filter, setFilter] = useState<AuditLogFilter>({
    page: 1,
    page_size: 25,
  });
  const [actionInput, setActionInput] = useState("");
  const [entityInput, setEntityInput] = useState("");
  const [searchInput, setSearchInput] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [exporting, setExporting] = useState(false);

  const { data, isLoading, isError, refetch, isFetching } = useQuery({
    queryKey: ["audit-logs", filter],
    queryFn: () => auditLogService.list(filter),
    placeholderData: (prev) => prev,
    staleTime: 15_000,
  });

  const applyFilters = useCallback(() => {
    setFilter((f) => ({
      ...f,
      page: 1,
      action: actionInput || undefined,
      entity_type: entityInput || undefined,
      search: searchInput || undefined,
      date_from: dateFrom || undefined,
      date_to: dateTo || undefined,
    }));
  }, [actionInput, entityInput, searchInput, dateFrom, dateTo]);

  const clearFilters = useCallback(() => {
    setActionInput("");
    setEntityInput("");
    setSearchInput("");
    setDateFrom("");
    setDateTo("");
    setFilter({ page: 1, page_size: 25 });
  }, []);

  const hasActiveFilters =
    filter.action ||
    filter.entity_type ||
    filter.search ||
    filter.date_from ||
    filter.date_to;

  const handleExport = async () => {
    setExporting(true);
    try {
      await auditLogService.exportCsv(filter);
    } finally {
      setExporting(false);
    }
  };

  const logs: AuditLog[] = data?.data ?? [];
  const meta = data?.meta;

  return (
    <DashboardLayout title="Audit Logs">
      <div className="mx-auto max-w-screen-xl space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <ShieldAlert className="h-5 w-5 text-slate-600" />
          <h1 className="text-lg font-bold text-slate-800">Audit Logs</h1>
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
            onClick={handleExport}
            disabled={exporting}
            className="flex items-center gap-1.5 rounded border border-slate-300 px-3 py-1.5 text-xs hover:bg-slate-50 disabled:opacity-40"
          >
            {exporting ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Download className="h-3.5 w-3.5" />
            )}
            Export CSV
          </button>
        </div>
      </div>

      {/* Summary */}
      <SummaryCards />

      {/* Filter bar */}
      <div className="rounded-lg border border-slate-200 bg-white p-3 shadow-sm">
        <div className="mb-2 flex items-center gap-1.5 text-xs font-semibold text-slate-500">
          <Filter className="h-3.5 w-3.5" />
          Filter
          {hasActiveFilters && (
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
        <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-5">
          <div className="relative">
            <Search className="pointer-events-none absolute left-2 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              placeholder="Cari keyword…"
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && applyFilters()}
              className="w-full rounded border border-slate-300 py-1.5 pl-7 pr-2 text-xs focus:outline-none focus:ring-1 focus:ring-blue-400"
            />
          </div>
          <input
            type="text"
            placeholder="Action (mis. project.created)"
            value={actionInput}
            onChange={(e) => setActionInput(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && applyFilters()}
            className="rounded border border-slate-300 px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-blue-400"
          />
          <input
            type="text"
            placeholder="Entity type (mis. project)"
            value={entityInput}
            onChange={(e) => setEntityInput(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && applyFilters()}
            className="rounded border border-slate-300 px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-blue-400"
          />
          <input
            type="date"
            value={dateFrom}
            onChange={(e) => setDateFrom(e.target.value)}
            className="rounded border border-slate-300 px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-blue-400"
            aria-label="Date from"
          />
          <input
            type="date"
            value={dateTo}
            onChange={(e) => setDateTo(e.target.value)}
            className="rounded border border-slate-300 px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-blue-400"
            aria-label="Date to"
          />
        </div>
        <div className="mt-2 flex justify-end">
          <button
            type="button"
            onClick={applyFilters}
            className="rounded bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700 active:bg-blue-800"
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
            <span className="text-sm">Memuat audit logs…</span>
          </div>
        ) : isError ? (
          <div className="flex flex-col items-center justify-center gap-2 p-12 text-rose-500">
            <Activity className="h-8 w-8 opacity-50" />
            <p className="text-sm font-medium">Gagal memuat audit logs</p>
            <button
              type="button"
              onClick={() => refetch()}
              className="text-xs underline hover:no-underline"
            >
              Coba lagi
            </button>
          </div>
        ) : logs.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 p-12 text-slate-400">
            <Activity className="h-8 w-8 opacity-30" />
            <p className="text-sm">Tidak ada audit log yang sesuai filter</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead className="border-b border-slate-200 bg-slate-50 text-[10px] font-semibold uppercase tracking-wide text-slate-500">
                <tr>
                  <th className="px-3 py-2">Waktu</th>
                  <th className="px-3 py-2">Aktor</th>
                  <th className="px-3 py-2">Action</th>
                  <th className="px-3 py-2">Entity</th>
                  <th className="px-3 py-2">Label / ID</th>
                  <th className="px-3 py-2">Detail</th>
                </tr>
              </thead>
              <tbody>
                {logs.map((log) => (
                  <AuditRow key={log.id} log={log} />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Pagination footer */}
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
      </div>
    </DashboardLayout>
  );
}
