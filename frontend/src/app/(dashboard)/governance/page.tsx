"use client";

import { useMemo, useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  ShieldCheck,
  Plus,
  Loader2,
  RefreshCw,
  Send,
  Search,
  Lock,
  XCircle,
  CheckCircle2,
  Eye,
  Ban,
  FileCheck2,
  CalendarRange,
  AlertTriangle,
} from "lucide-react";
import { governanceService } from "@/services/governance.service";
import type {
  GovernanceSubmission,
  GovernanceSubmissionDetail,
  GovernanceSubmissionStatus,
  GovernanceDatasetType,
  GovernanceSourceType,
  GovernanceItemAction,
  CreateGovernanceSubmissionRequest,
  GovernanceLockPeriod,
} from "@/types/governance";
import {
  GOVERNANCE_DATASET_LABELS,
  GOVERNANCE_SOURCE_LABELS,
  GOVERNANCE_STATUS_LABELS,
} from "@/types/governance";
import { DashboardLayout } from "@/components/layout/DashboardLayout";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function fmtDate(s?: string): string {
  if (!s) return "—";
  return new Date(s).toLocaleString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

const DATASETS = Object.keys(GOVERNANCE_DATASET_LABELS) as GovernanceDatasetType[];
const SOURCES = Object.keys(GOVERNANCE_SOURCE_LABELS) as GovernanceSourceType[];
const STATUSES = Object.keys(GOVERNANCE_STATUS_LABELS) as GovernanceSubmissionStatus[];
const ITEM_ACTIONS: GovernanceItemAction[] = [
  "CREATE",
  "UPDATE",
  "DELETE",
  "UPSERT",
  "VALIDATE_ONLY",
];

// ---------------------------------------------------------------------------
// Badges
// ---------------------------------------------------------------------------

function StatusBadge({ status }: { status: GovernanceSubmissionStatus }) {
  const map: Record<GovernanceSubmissionStatus, string> = {
    DRAFT: "bg-slate-100 text-slate-600",
    SUBMITTED: "bg-blue-100 text-blue-700",
    IN_REVIEW: "bg-yellow-100 text-yellow-700",
    APPROVED: "bg-green-100 text-green-700",
    REJECTED: "bg-red-100 text-red-700",
    LOCKED: "bg-slate-800 text-white",
    CANCELLED: "bg-gray-200 text-gray-600",
  };
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${map[status] ?? "bg-slate-100 text-slate-600"}`}>
      {GOVERNANCE_STATUS_LABELS[status] ?? status}
    </span>
  );
}

function ItemStatusBadge({ status }: { status: string }) {
  const map: Record<string, string> = {
    PENDING: "bg-slate-100 text-slate-600",
    VALID: "bg-green-100 text-green-700",
    INVALID: "bg-red-100 text-red-700",
  };
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${map[status] ?? "bg-slate-100 text-slate-600"}`}>
      {status}
    </span>
  );
}

// ---------------------------------------------------------------------------
// Create submission modal
// ---------------------------------------------------------------------------

interface CreateForm {
  dataset_type: GovernanceDatasetType;
  source_type: GovernanceSourceType;
  period_year: number;
  period_month: string;
  source_reference: string;
  items: { entity_type: string; entity_id: string; action: GovernanceItemAction; payload: string }[];
}

function CreateSubmissionModal({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient();
  const year = new Date().getFullYear();
  const [form, setForm] = useState<CreateForm>({
    dataset_type: "PROJECT_PROGRESS",
    source_type: "MANUAL",
    period_year: year,
    period_month: String(new Date().getMonth() + 1),
    source_reference: "",
    items: [{ entity_type: "project", entity_id: "", action: "VALIDATE_ONLY", payload: "{}" }],
  });
  const [error, setError] = useState("");

  const mutation = useMutation({
    mutationFn: () => {
      const payload: CreateGovernanceSubmissionRequest = {
        dataset_type: form.dataset_type,
        source_type: form.source_type,
        period_year: form.period_year,
        items: form.items.map((it) => {
          const base: CreateGovernanceSubmissionRequest["items"][number] = {
            entity_type: it.entity_type,
            action: it.action,
          };
          if (it.entity_id.trim()) base.entity_id = it.entity_id.trim();
          try {
            const parsed = it.payload.trim() ? JSON.parse(it.payload) : {};
            base.payload_after = parsed;
          } catch {
            throw new Error("Payload JSON item tidak valid");
          }
          return base;
        }),
      };
      if (form.period_month.trim()) payload.period_month = Number(form.period_month);
      if (form.source_reference.trim()) payload.source_reference = form.source_reference.trim();
      return governanceService.createSubmission(payload);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["governance-submissions"] });
      onClose();
    },
    onError: (err: Error) => {
      setError(err.message || "Gagal membuat submission");
    },
  });

  const updateItem = (idx: number, patch: Partial<CreateForm["items"][number]>) => {
    setForm((f) => ({
      ...f,
      items: f.items.map((it, i) => (i === idx ? { ...it, ...patch } : it)),
    }));
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-black/40 p-4">
      <div className="my-8 w-full max-w-2xl rounded-xl bg-white p-6 shadow-2xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-800">Buat Submission Baru</h2>
          <button type="button" onClick={onClose} className="rounded-md p-1 text-slate-400 hover:bg-slate-100" aria-label="Tutup">
            <XCircle className="h-5 w-5" />
          </button>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <label className="block text-sm">
            <span className="mb-1 block font-medium text-slate-600">Tipe Dataset</span>
            <select
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
              value={form.dataset_type}
              onChange={(e) => setForm({ ...form, dataset_type: e.target.value as GovernanceDatasetType })}
            >
              {DATASETS.map((d) => (
                <option key={d} value={d}>{GOVERNANCE_DATASET_LABELS[d]}</option>
              ))}
            </select>
          </label>
          <label className="block text-sm">
            <span className="mb-1 block font-medium text-slate-600">Sumber Data</span>
            <select
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
              value={form.source_type}
              onChange={(e) => setForm({ ...form, source_type: e.target.value as GovernanceSourceType })}
            >
              {SOURCES.map((s) => (
                <option key={s} value={s}>{GOVERNANCE_SOURCE_LABELS[s]}</option>
              ))}
            </select>
          </label>
          <label className="block text-sm">
            <span className="mb-1 block font-medium text-slate-600">Tahun Periode</span>
            <input
              type="number"
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
              value={form.period_year}
              onChange={(e) => setForm({ ...form, period_year: Number(e.target.value) })}
            />
          </label>
          <label className="block text-sm">
            <span className="mb-1 block font-medium text-slate-600">Bulan (kosongkan = full tahun)</span>
            <input
              type="number"
              min={1}
              max={12}
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
              value={form.period_month}
              onChange={(e) => setForm({ ...form, period_month: e.target.value })}
              placeholder="1-12"
            />
          </label>
          <label className="block text-sm sm:col-span-2">
            <span className="mb-1 block font-medium text-slate-600">Referensi Sumber</span>
            <input
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
              value={form.source_reference}
              onChange={(e) => setForm({ ...form, source_reference: e.target.value })}
              placeholder="contoh: import-job-id / run-id"
            />
          </label>
        </div>

        <div className="mt-5">
          <div className="mb-2 flex items-center justify-between">
            <h3 className="text-sm font-semibold text-slate-700">Item Submission</h3>
            <button
              type="button"
              className="inline-flex items-center gap-1 rounded-md bg-blue-50 px-2 py-1 text-xs font-medium text-blue-700 hover:bg-blue-100"
              onClick={() =>
                setForm((f) => ({
                  ...f,
                  items: [...f.items, { entity_type: "project", entity_id: "", action: "VALIDATE_ONLY", payload: "{}" }],
                }))
              }
            >
              <Plus className="h-3.5 w-3.5" /> Tambah Item
            </button>
          </div>
          <div className="space-y-3">
            {form.items.map((it, idx) => (
              <div key={idx} className="rounded-lg border border-slate-200 p-3">
                <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
                  <input
                    className="rounded-md border border-slate-300 px-2 py-1.5 text-sm"
                    placeholder="entity_type (project, budget...)"
                    value={it.entity_type}
                    onChange={(e) => updateItem(idx, { entity_type: e.target.value })}
                  />
                  <input
                    className="rounded-md border border-slate-300 px-2 py-1.5 text-sm"
                    placeholder="entity_id (UUID, opsional)"
                    value={it.entity_id}
                    onChange={(e) => updateItem(idx, { entity_id: e.target.value })}
                  />
                  <select
                    className="rounded-md border border-slate-300 px-2 py-1.5 text-sm"
                    value={it.action}
                    onChange={(e) => updateItem(idx, { action: e.target.value as GovernanceItemAction })}
                  >
                    {ITEM_ACTIONS.map((a) => (
                      <option key={a} value={a}>{a}</option>
                    ))}
                  </select>
                </div>
                <textarea
                  className="mt-2 w-full rounded-md border border-slate-300 px-2 py-1.5 font-mono text-xs"
                  rows={2}
                  placeholder='payload_after JSON, contoh: {"progress_pct": 45}'
                  value={it.payload}
                  onChange={(e) => updateItem(idx, { payload: e.target.value })}
                />
              </div>
            ))}
          </div>
        </div>

        {error && (
          <div className="mt-3 flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
            <AlertTriangle className="h-4 w-4" /> {error}
          </div>
        )}

        <div className="mt-5 flex justify-end gap-2">
          <button
            type="button"
            className="rounded-md border border-slate-300 px-4 py-2 text-sm text-slate-600 hover:bg-slate-50"
            onClick={onClose}
          >
            Batal
          </button>
          <button
            type="button"
            disabled={mutation.isPending}
            className="inline-flex items-center gap-2 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            onClick={() => mutation.mutate()}
          >
            {mutation.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
            Buat Draft
          </button>
        </div>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Submission detail modal (timeline + items + actions)
// ---------------------------------------------------------------------------

function ActionButton({
  label,
  onClick,
  kind,
  pending,
}: {
  label: string;
  onClick: () => void;
  kind: "primary" | "approve" | "reject" | "lock" | "ghost";
  pending?: boolean;
}) {
  const cls: Record<string, string> = {
    primary: "bg-blue-600 text-white hover:bg-blue-700",
    approve: "bg-green-600 text-white hover:bg-green-700",
    reject: "bg-red-600 text-white hover:bg-red-700",
    lock: "bg-slate-800 text-white hover:bg-slate-900",
    ghost: "border border-slate-300 text-slate-600 hover:bg-slate-50",
  };
  return (
    <button
      type="button"
      disabled={pending}
      onClick={onClick}
      className={`inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium disabled:opacity-50 ${cls[kind]}`}
    >
      {pending && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
      {label}
    </button>
  );
}

function SubmissionDetailModal({
  submission,
  onClose,
}: {
  submission: GovernanceSubmissionDetail;
  onClose: () => void;
}) {
  const qc = useQueryClient();
  const [reason, setReason] = useState("");
  const [rejecting, setRejecting] = useState(false);
  const [notes, setNotes] = useState("");

  const invalidate = () => qc.invalidateQueries({ queryKey: ["governance-submissions"] });

  const mutation = useMutation({
    mutationFn: async (action: string) => {
      switch (action) {
        case "submit":
          await governanceService.submit(submission.id);
          break;
        case "review":
          await governanceService.startReview(submission.id, { review_notes: notes || undefined });
          break;
        case "approve":
          await governanceService.approve(submission.id);
          break;
        case "reject":
          if (!reason.trim()) throw new Error("Alasan penolakan wajib diisi");
          await governanceService.reject(submission.id, { rejection_reason: reason.trim() });
          break;
        case "lock":
          await governanceService.lock(submission.id, { lock_reason: reason.trim() || undefined });
          break;
        case "cancel":
          await governanceService.cancel(submission.id, { cancel_reason: reason.trim() || undefined });
          break;
      }
    },
    onSuccess: () => {
      setReason("");
      setRejecting(false);
      invalidate();
      onClose();
    },
  });

  const canSubmit = submission.status === "DRAFT";
  const canReview = submission.status === "SUBMITTED";
  const canApprove = submission.status === "IN_REVIEW";
  const canReject = submission.status === "IN_REVIEW";
  const canLock = submission.status === "APPROVED";
  const canCancel = submission.status === "DRAFT" || submission.status === "SUBMITTED";

  const timeline = [
    { label: "Dibuat (DRAFT)", at: submission.created_at },
    { label: "Diajukan (SUBMITTED)", at: submission.submitted_at },
    { label: "Review (IN_REVIEW)", at: submission.reviewed_at },
    { label: "Disetujui (APPROVED)", at: submission.approved_at },
    { label: "Dikunci (LOCKED)", at: submission.locked_at },
  ].filter((t) => t.at);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-black/40 p-4">
      <div className="my-8 w-full max-w-3xl rounded-xl bg-white p-6 shadow-2xl">
        <div className="mb-4 flex items-start justify-between">
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-lg font-semibold text-slate-800">
                {GOVERNANCE_DATASET_LABELS[submission.dataset_type] ?? submission.dataset_type}
              </h2>
              <StatusBadge status={submission.status} />
            </div>
            <p className="mt-1 text-xs text-slate-500">
              {GOVERNANCE_SOURCE_LABELS[submission.source_type] ?? submission.source_type} · Periode{" "}
              {submission.period_month ? `${submission.period_month}/${submission.period_year}` : submission.period_year}
              {submission.source_reference ? ` · Ref: ${submission.source_reference}` : ""}
            </p>
          </div>
          <button type="button" onClick={onClose} className="rounded-md p-1 text-slate-400 hover:bg-slate-100" aria-label="Tutup">
            <XCircle className="h-5 w-5" />
          </button>
        </div>

        {/* Timeline */}
        <div className="mb-5 rounded-lg border border-slate-200 p-4">
          <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-700">
            <CalendarRange className="h-4 w-4" /> Alur Status
          </h3>
          <div className="flex flex-wrap items-center gap-1">
            {timeline.map((t, i) => (
              <div key={t.label} className="flex items-center gap-1">
                <div className="rounded-md bg-slate-50 px-2 py-1 text-xs">
                  <span className="font-medium text-slate-700">{t.label}</span>
                  <span className="ml-1 text-slate-400">{fmtDate(t.at)}</span>
                </div>
                {i < timeline.length - 1 && <span className="text-slate-300">→</span>}
              </div>
            ))}
          </div>
          {submission.rejection_reason && (
            <p className="mt-2 rounded-md bg-red-50 px-2 py-1 text-xs text-red-700">
              Alasan penolakan: {submission.rejection_reason}
            </p>
          )}
          {submission.review_notes && (
            <p className="mt-2 rounded-md bg-blue-50 px-2 py-1 text-xs text-blue-700">
              Catatan review: {submission.review_notes}
            </p>
          )}
        </div>

        {/* Items */}
        <div className="mb-5">
          <h3 className="mb-2 text-sm font-semibold text-slate-700">Item ({submission.items.length})</h3>
          <div className="overflow-x-auto rounded-lg border border-slate-200">
            <table className="w-full min-w-[640px] text-left text-sm">
              <thead className="bg-slate-50 text-xs text-slate-500">
                <tr>
                  <th className="px-3 py-2 font-medium">Entity</th>
                  <th className="px-3 py-2 font-medium">Action</th>
                  <th className="px-3 py-2 font-medium">Validasi</th>
                  <th className="px-3 py-2 font-medium">Payload</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {submission.items.map((it) => (
                  <tr key={it.id}>
                    <td className="px-3 py-2">
                      <span className="font-medium text-slate-700">{it.entity_type}</span>
                      {it.entity_id && <span className="block text-xs text-slate-400">{it.entity_id.slice(0, 8)}…</span>}
                    </td>
                    <td className="px-3 py-2">
                      <span className="rounded bg-slate-100 px-1.5 py-0.5 text-xs font-medium text-slate-600">{it.action}</span>
                    </td>
                    <td className="px-3 py-2">
                      <ItemStatusBadge status={it.validation_status} />
                      {it.validation_errors?.length > 0 && (
                        <ul className="mt-1 list-inside list-disc text-xs text-red-600">
                          {it.validation_errors.map((e, i) => (
                            <li key={i}>{e}</li>
                          ))}
                        </ul>
                      )}
                    </td>
                    <td className="max-w-[220px] truncate px-3 py-2 font-mono text-xs text-slate-500">
                      {JSON.stringify(it.payload_after ?? {})}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Actions */}
        {(canSubmit || canReview || canApprove || canReject || canLock || canCancel) && (
          <div className="rounded-lg border border-slate-200 p-4">
            <h3 className="mb-2 text-sm font-semibold text-slate-700">Tindakan</h3>
            {(canReview || canReject || canLock || canCancel) && (
              <textarea
                className="mb-2 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                rows={2}
                placeholder={canReject ? "Alasan penolakan (wajib)" : "Catatan / alasan (opsional)"}
                value={reason}
                onChange={(e) => setReason(e.target.value)}
              />
            )}
            {canReview && (
              <div className="mb-2">
                <textarea
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                  rows={2}
                  placeholder="Catatan review (opsional)"
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                />
              </div>
            )}
            <div className="flex flex-wrap gap-2">
              {canSubmit && (
                <ActionButton label="Submit" kind="primary" pending={mutation.isPending} onClick={() => mutation.mutate("submit")} />
              )}
              {canReview && (
                <ActionButton label="Mulai Review" kind="primary" pending={mutation.isPending} onClick={() => mutation.mutate("review")} />
              )}
              {canApprove && (
                <ActionButton label="Approve" kind="approve" pending={mutation.isPending} onClick={() => mutation.mutate("approve")} />
              )}
              {canReject && (
                <ActionButton label="Tolak" kind="reject" pending={mutation.isPending} onClick={() => { setRejecting(true); mutation.mutate("reject"); }} />
              )}
              {canLock && (
                <ActionButton label="Kunci Periode" kind="lock" pending={mutation.isPending} onClick={() => mutation.mutate("lock")} />
              )}
              {canCancel && (
                <ActionButton label="Batalkan" kind="ghost" pending={mutation.isPending} onClick={() => mutation.mutate("cancel")} />
              )}
            </div>
            {rejecting && !reason.trim() && (
              <p className="mt-2 text-xs text-red-600">Alasan penolakan wajib diisi.</p>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Lock period modal
// ---------------------------------------------------------------------------

function CreateLockPeriodModal({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient();
  const [form, setForm] = useState({
    dataset_type: "BUDGET" as GovernanceDatasetType,
    period_year: new Date().getFullYear(),
    period_month: "",
    lock_reason: "",
    lock_now: false,
  });
  const mutation = useMutation({
    mutationFn: () =>
      governanceService.createLockPeriod({
        dataset_type: form.dataset_type,
        period_year: form.period_year,
        period_month: form.period_month ? Number(form.period_month) : undefined,
        lock_reason: form.lock_reason || undefined,
        lock_now: form.lock_now,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["governance-lock-periods"] });
      onClose();
    },
  });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-2xl">
        <h2 className="mb-4 text-lg font-semibold text-slate-800">Buat Lock Period</h2>
        <div className="space-y-3">
          <label className="block text-sm">
            <span className="mb-1 block font-medium text-slate-600">Tipe Dataset</span>
            <select
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
              value={form.dataset_type}
              onChange={(e) => setForm({ ...form, dataset_type: e.target.value as GovernanceDatasetType })}
            >
              {DATASETS.map((d) => (
                <option key={d} value={d}>{GOVERNANCE_DATASET_LABELS[d]}</option>
              ))}
            </select>
          </label>
          <div className="grid grid-cols-2 gap-3">
            <label className="block text-sm">
              <span className="mb-1 block font-medium text-slate-600">Tahun</span>
              <input
                type="number"
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                value={form.period_year}
                onChange={(e) => setForm({ ...form, period_year: Number(e.target.value) })}
              />
            </label>
            <label className="block text-sm">
              <span className="mb-1 block font-medium text-slate-600">Bulan</span>
              <input
                type="number"
                min={1}
                max={12}
                placeholder="1-12 (kosong = tahun)"
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                value={form.period_month}
                onChange={(e) => setForm({ ...form, period_month: e.target.value })}
              />
            </label>
          </div>
          <label className="block text-sm">
            <span className="mb-1 block font-medium text-slate-600">Alasan Lock</span>
            <textarea
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
              rows={2}
              value={form.lock_reason}
              onChange={(e) => setForm({ ...form, lock_reason: e.target.value })}
            />
          </label>
          <label className="flex items-center gap-2 text-sm text-slate-700">
            <input
              type="checkbox"
              checked={form.lock_now}
              onChange={(e) => setForm({ ...form, lock_now: e.target.checked })}
              className="h-4 w-4 rounded border-slate-300"
            />
            Langsung kunci (LOCKED) sekarang
          </label>
        </div>
        <div className="mt-5 flex justify-end gap-2">
          <button
            type="button"
            className="rounded-md border border-slate-300 px-4 py-2 text-sm text-slate-600 hover:bg-slate-50"
            onClick={onClose}
          >
            Batal
          </button>
          <button
            type="button"
            disabled={mutation.isPending}
            className="inline-flex items-center gap-2 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            onClick={() => mutation.mutate()}
          >
            {mutation.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Lock className="h-4 w-4" />}
            Simpan
          </button>
        </div>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Lock period panel
// ---------------------------------------------------------------------------

function LockPeriodPanel() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["governance-lock-periods"],
    queryFn: () => governanceService.listLockPeriods({ page_size: 50 }),
  });
  const qc = useQueryClient();
  const [creating, setCreating] = useState(false);
  const [locking, setLocking] = useState<string | null>(null);

  const lockMutation = useMutation({
    mutationFn: (id: string) => governanceService.lockPeriod(id, { lock_reason: "Kunci via UI" }),
    onSuccess: () => {
      setLocking(null);
      qc.invalidateQueries({ queryKey: ["governance-lock-periods"] });
    },
  });

  const periods = data?.data ?? [];

  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="flex items-center gap-2 text-base font-semibold text-slate-800">
          <Lock className="h-5 w-5 text-slate-500" /> Lock Period
        </h2>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => refetch()}
            className="inline-flex items-center gap-1 rounded-md border border-slate-300 px-2.5 py-1.5 text-xs text-slate-600 hover:bg-slate-50"
          >
            <RefreshCw className="h-3.5 w-3.5" /> Muat Ulang
          </button>
          <button
            type="button"
            onClick={() => setCreating(true)}
            className="inline-flex items-center gap-1 rounded-md bg-blue-600 px-2.5 py-1.5 text-xs font-medium text-white hover:bg-blue-700"
          >
            <Plus className="h-3.5 w-3.5" /> Buat
          </button>
        </div>
      </div>

      {isLoading && (
        <div className="flex items-center justify-center py-8 text-slate-400">
          <Loader2 className="h-5 w-5 animate-spin" />
        </div>
      )}
      {error && (
        <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">Gagal memuat lock period.</div>
      )}
      {!isLoading && !error && periods.length === 0 && (
        <div className="rounded-md bg-slate-50 px-3 py-6 text-center text-sm text-slate-400">
          Belum ada lock period. Buat untuk menutup periode pelaporan.
        </div>
      )}
      {periods.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full min-w-[560px] text-left text-sm">
            <thead className="bg-slate-50 text-xs text-slate-500">
              <tr>
                <th className="px-3 py-2 font-medium">Dataset</th>
                <th className="px-3 py-2 font-medium">Periode</th>
                <th className="px-3 py-2 font-medium">Status</th>
                <th className="px-3 py-2 font-medium">Alasan</th>
                <th className="px-3 py-2 font-medium"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {periods.map((p: GovernanceLockPeriod) => (
                <tr key={p.id}>
                  <td className="px-3 py-2 font-medium text-slate-700">
                    {GOVERNANCE_DATASET_LABELS[p.dataset_type] ?? p.dataset_type}
                  </td>
                  <td className="px-3 py-2 text-slate-600">
                    {p.period_month ? `${p.period_month}/${p.period_year}` : p.period_year}
                  </td>
                  <td className="px-3 py-2">
                    {p.status === "LOCKED" ? (
                      <span className="inline-flex items-center rounded-full bg-slate-800 px-2 py-0.5 text-xs font-medium text-white">
                        LOCKED
                      </span>
                    ) : (
                      <span className="inline-flex items-center rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
                        OPEN
                      </span>
                    )}
                  </td>
                  <td className="max-w-[200px] truncate px-3 py-2 text-xs text-slate-500">{p.lock_reason || "—"}</td>
                  <td className="px-3 py-2 text-right">
                    {p.status === "OPEN" && (
                      <button
                        type="button"
                        disabled={locking === p.id}
                        className="inline-flex items-center gap-1 rounded-md bg-slate-800 px-2 py-1 text-xs font-medium text-white hover:bg-slate-900 disabled:opacity-50"
                        onClick={() => {
                          setLocking(p.id);
                          lockMutation.mutate(p.id);
                        }}
                      >
                        {locking === p.id ? <Loader2 className="h-3 w-3 animate-spin" /> : <Lock className="h-3 w-3" />}
                        Kunci
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {creating && <CreateLockPeriodModal onClose={() => setCreating(false)} />}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main page
// ---------------------------------------------------------------------------

export default function GovernancePage() {
  const [status, setStatus] = useState<string>("");
  const [dataset, setDataset] = useState<string>("");
  const [source, setSource] = useState<string>("");
  const [year, setYear] = useState<string>("");
  const [creating, setCreating] = useState(false);
  const [detail, setDetail] = useState<GovernanceSubmissionDetail | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  const filter = useMemo(
    () => ({
      status: (status || undefined) as GovernanceSubmissionStatus | undefined,
      dataset_type: (dataset || undefined) as GovernanceDatasetType | undefined,
      source_type: (source || undefined) as GovernanceSourceType | undefined,
      period_year: year ? Number(year) : undefined,
      page_size: 50,
    }),
    [status, dataset, source, year]
  );

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["governance-submissions", filter],
    queryFn: () => governanceService.listSubmissions(filter),
  });

  const submissions = data?.data ?? [];

  const openDetail = async (id: string) => {
    setLoadingDetail(true);
    try {
      const detailData = await governanceService.getSubmission(id);
      setDetail(detailData);
    } catch {
      // keep modal closed on error; list stays visible
    } finally {
      setLoadingDetail(false);
    }
  };

  return (
    <DashboardLayout title="Data Governance">
    <div className="mx-auto max-w-7xl space-y-6">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="flex items-center gap-2 text-xl font-bold text-slate-800">
            <ShieldCheck className="h-6 w-6 text-blue-600" /> Data Governance
          </h1>
          <p className="mt-1 text-sm text-slate-500">
            Alur validasi &amp; persetujuan data resmi (official validation workflow).
          </p>
        </div>
        <button
          type="button"
          onClick={() => setCreating(true)}
          className="inline-flex items-center gap-2 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" /> Buat Submission
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-2 rounded-xl border border-slate-200 bg-white p-3">
        <Search className="h-4 w-4 text-slate-400" />
        <select
          className="rounded-md border border-slate-300 px-2 py-1.5 text-sm"
          value={status}
          onChange={(e) => setStatus(e.target.value)}
        >
          <option value="">Semua Status</option>
          {STATUSES.map((s) => (
            <option key={s} value={s}>{GOVERNANCE_STATUS_LABELS[s]}</option>
          ))}
        </select>
        <select
          className="rounded-md border border-slate-300 px-2 py-1.5 text-sm"
          value={dataset}
          onChange={(e) => setDataset(e.target.value)}
        >
          <option value="">Semua Dataset</option>
          {DATASETS.map((d) => (
            <option key={d} value={d}>{GOVERNANCE_DATASET_LABELS[d]}</option>
          ))}
        </select>
        <select
          className="rounded-md border border-slate-300 px-2 py-1.5 text-sm"
          value={source}
          onChange={(e) => setSource(e.target.value)}
        >
          <option value="">Semua Sumber</option>
          {SOURCES.map((s) => (
            <option key={s} value={s}>{GOVERNANCE_SOURCE_LABELS[s]}</option>
          ))}
        </select>
        <input
          type="number"
          placeholder="Tahun"
          className="w-24 rounded-md border border-slate-300 px-2 py-1.5 text-sm"
          value={year}
          onChange={(e) => setYear(e.target.value)}
        />
        <button
          type="button"
          onClick={() => refetch()}
          className="ml-auto inline-flex items-center gap-1 rounded-md border border-slate-300 px-2.5 py-1.5 text-xs text-slate-600 hover:bg-slate-50"
        >
          <RefreshCw className="h-3.5 w-3.5" /> Muat Ulang
        </button>
      </div>

      {/* Submission list */}
      <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
        <div className="border-b border-slate-100 px-5 py-3">
          <h2 className="text-sm font-semibold text-slate-700">
            Daftar Submission{" "}
            <span className="ml-1 rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-500">
              {data?.meta?.total ?? submissions.length}
            </span>
          </h2>
        </div>

        {isLoading && (
          <div className="flex items-center justify-center py-16 text-slate-400">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        )}
        {error && (
          <div className="flex flex-col items-center gap-2 py-16 text-sm text-red-600">
            <AlertTriangle className="h-6 w-6" />
            Gagal memuat submission.
          </div>
        )}
        {!isLoading && !error && submissions.length === 0 && (
          <div className="flex flex-col items-center gap-2 py-16 text-sm text-slate-400">
            <FileCheck2 className="h-8 w-8" />
            Belum ada submission. Buat submission baru untuk memulai alur validasi.
          </div>
        )}
        {!isLoading && !error && submissions.length > 0 && (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[760px] text-left text-sm">
              <thead className="bg-slate-50 text-xs text-slate-500">
                <tr>
                  <th className="px-4 py-2.5 font-medium">Dataset</th>
                  <th className="px-4 py-2.5 font-medium">Sumber</th>
                  <th className="px-4 py-2.5 font-medium">Periode</th>
                  <th className="px-4 py-2.5 font-medium">Status</th>
                  <th className="px-4 py-2.5 font-medium">Diajukan</th>
                  <th className="px-4 py-2.5 font-medium"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {submissions.map((s: GovernanceSubmission) => (
                  <tr key={s.id} className="hover:bg-slate-50/60">
                    <td className="px-4 py-3 font-medium text-slate-700">
                      {GOVERNANCE_DATASET_LABELS[s.dataset_type] ?? s.dataset_type}
                    </td>
                    <td className="px-4 py-3 text-slate-600">
                      {GOVERNANCE_SOURCE_LABELS[s.source_type] ?? s.source_type}
                    </td>
                    <td className="px-4 py-3 text-slate-600">
                      {s.period_month ? `${s.period_month}/${s.period_year}` : s.period_year}
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge status={s.status} />
                    </td>
                    <td className="px-4 py-3 text-xs text-slate-500">{fmtDate(s.submitted_at || s.created_at)}</td>
                    <td className="px-4 py-3 text-right">
                      <button
                        type="button"
                        className="inline-flex items-center gap-1 rounded-md border border-slate-300 px-2 py-1 text-xs text-slate-600 hover:bg-slate-50"
                        onClick={() => openDetail(s.id)}
                      >
                        {loadingDetail ? <Loader2 className="h-3 w-3 animate-spin" /> : <Eye className="h-3 w-3" />}
                        Detail
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Lock period panel */}
      <LockPeriodPanel />

      {creating && <CreateSubmissionModal onClose={() => setCreating(false)} />}
      {detail && <SubmissionDetailModal submission={detail} onClose={() => setDetail(null)} />}
    </div>
    </DashboardLayout>
  );
}
