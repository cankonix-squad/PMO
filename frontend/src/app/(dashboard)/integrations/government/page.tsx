"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Building2,
  ChevronDown,
  ChevronRight,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Loader2,
  XCircle,
  RefreshCw,
  Play,
  Ban,
  Link2,
  Link2Off,
  XOctagon,
  GitMerge,
} from "lucide-react";
import { governmentService } from "@/services/government.service";
import type {
  ConnectorDefinition,
  SyncRun,
  SyncRecord,
  ExternalMapping,
  ResolutionCandidate,
  MatchStatus,
  RunStatus,
  SyncMode,
  CreateRunRequest,
} from "@/types/government";
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

// ---------------------------------------------------------------------------
// Status components
// ---------------------------------------------------------------------------

function StatusBadge({ status }: { status: RunStatus }) {
  const map: Record<RunStatus, { label: string; cls: string }> = {
    PENDING:   { label: "Pending",    cls: "bg-slate-100 text-slate-600" },
    RUNNING:   { label: "Berjalan",   cls: "bg-blue-100 text-blue-700" },
    SUCCEEDED: { label: "Berhasil",   cls: "bg-green-100 text-green-700" },
    FAILED:    { label: "Gagal",      cls: "bg-red-100 text-red-700" },
    CANCELLED: { label: "Dibatalkan", cls: "bg-yellow-100 text-yellow-700" },
  };
  const { label, cls } = map[status] ?? { label: status, cls: "bg-slate-100 text-slate-600" };
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${cls}`}>
      {label}
    </span>
  );
}

function StatusIcon({ status }: { status: RunStatus }) {
  if (status === "SUCCEEDED") return <CheckCircle2 className="h-4 w-4 text-green-500" />;
  if (status === "FAILED")    return <AlertTriangle className="h-4 w-4 text-red-500" />;
  if (status === "RUNNING")   return <Loader2 className="h-4 w-4 animate-spin text-blue-500" />;
  if (status === "CANCELLED") return <XCircle className="h-4 w-4 text-yellow-500" />;
  return <Clock className="h-4 w-4 text-slate-400" />;
}

function ModeBadge({ mode }: { mode: SyncMode }) {
  const map: Record<SyncMode, { label: string; cls: string }> = {
    SAMPLE:  { label: "Sampel",   cls: "bg-purple-100 text-purple-700" },
    DRY_RUN: { label: "Dry Run",  cls: "bg-orange-100 text-orange-700" },
    COMMIT:  { label: "Commit",   cls: "bg-teal-100 text-teal-700" },
  };
  const { label, cls } = map[mode] ?? { label: mode, cls: "bg-slate-100 text-slate-600" };
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${cls}`}>
      {label}
    </span>
  );
}

// ---------------------------------------------------------------------------
// RunRecords — records log for a specific run
// ---------------------------------------------------------------------------

function RunRecords({ runID }: { runID: string }) {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useQuery({
    queryKey: ["gov-records", runID, page],
    queryFn: () => governmentService.listRecords(runID, { page, page_size: 10 }),
  });

  if (isLoading) return <p className="py-4 text-center text-sm text-slate-500">Memuat records…</p>;
  if (!data?.data?.length) return <p className="py-4 text-center text-sm text-slate-400">Tidak ada records.</p>;

  return (
    <div className="mt-3">
      <div className="overflow-x-auto rounded-md border border-slate-200">
        <table className="min-w-full text-xs">
          <thead className="bg-slate-50">
            <tr>
              {["External ID", "Dataset", "Status", "Aksi", "Dibuat"].map((h) => (
                <th key={h} className="px-3 py-2 text-left font-medium text-slate-600">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {data.data.map((rec: SyncRecord) => (
              <tr key={rec.id} className="hover:bg-slate-50">
                <td className="px-3 py-2 font-mono text-slate-700">{rec.external_id}</td>
                <td className="px-3 py-2 text-slate-600">{rec.dataset_type}</td>
                <td className="px-3 py-2">
                  <span className={`inline-flex items-center rounded-full px-2 py-0.5 font-medium ${
                    rec.status === "ACCEPTED" ? "bg-green-100 text-green-700" :
                    rec.status === "REJECTED" ? "bg-red-100 text-red-700" :
                    "bg-slate-100 text-slate-600"
                  }`}>{rec.status}</span>
                </td>
                <td className="px-3 py-2 text-slate-600">{rec.action}</td>
                <td className="px-3 py-2 text-slate-500">{fmtDate(rec.created_at)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {data.meta.total > 10 && (
        <div className="mt-2 flex items-center justify-between text-xs text-slate-500">
          <span>Total {data.meta.total} records</span>
          <div className="flex gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="rounded px-2 py-1 hover:bg-slate-100 disabled:opacity-40"
            >← Prev</button>
            <span>Hal. {page}</span>
            <button
              onClick={() => setPage((p) => p + 1)}
              disabled={page * 10 >= data.meta.total}
              className="rounded px-2 py-1 hover:bg-slate-100 disabled:opacity-40"
            >Next →</button>
          </div>
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// RunCard
// ---------------------------------------------------------------------------

function RunCard({ run }: { run: SyncRun }) {
  const [expanded, setExpanded] = useState(false);
  const queryClient = useQueryClient();

  const cancelMutation = useMutation({
    mutationFn: () => governmentService.cancelRun(run.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["gov-runs"] });
    },
  });

  const canCancel = run.status === "PENDING";

  return (
    <div className="rounded-lg border border-slate-200 bg-white shadow-sm">
      <div
        className="flex cursor-pointer items-center gap-3 px-4 py-3"
        onClick={() => setExpanded((v) => !v)}
        role="button"
        aria-expanded={expanded}
        tabIndex={0}
        onKeyDown={(e) => e.key === "Enter" && setExpanded((v) => !v)}
      >
        <StatusIcon status={run.status} />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-medium text-slate-800 text-sm truncate">{run.connector_key}</span>
            <span className="text-slate-400 text-xs">·</span>
            <span className="text-slate-600 text-xs">{run.dataset_type}</span>
            <StatusBadge status={run.status} />
            <ModeBadge mode={run.mode} />
          </div>
          <div className="mt-0.5 flex flex-wrap gap-3 text-xs text-slate-500">
            <span>Diterima: <strong className="text-green-700">{run.accepted_records}</strong></span>
            <span>Ditolak: <strong className="text-red-600">{run.rejected_records}</strong></span>
            <span>Total: {run.total_records}</span>
            <span>Mulai: {fmtDate(run.started_at)}</span>
            {run.finished_at && <span>Selesai: {fmtDate(run.finished_at)}</span>}
          </div>
        </div>
        <div className="flex items-center gap-2">
          {canCancel && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                cancelMutation.mutate();
              }}
              disabled={cancelMutation.isPending}
              title="Batalkan run"
              className="rounded p-1 text-slate-400 hover:bg-red-50 hover:text-red-600 disabled:opacity-50"
              aria-label="Batalkan run"
            >
              {cancelMutation.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Ban className="h-4 w-4" />
              )}
            </button>
          )}
          {expanded ? (
            <ChevronDown className="h-4 w-4 text-slate-400" />
          ) : (
            <ChevronRight className="h-4 w-4 text-slate-400" />
          )}
        </div>
      </div>

      {expanded && (
        <div className="border-t border-slate-100 px-4 pb-4 pt-3">
          <p className="mb-1 text-xs font-semibold uppercase text-slate-500 tracking-wide">Records Log</p>
          <RunRecords runID={run.id} />
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// CreateRunModal
// ---------------------------------------------------------------------------

interface CreateRunModalProps {
  connector: ConnectorDefinition;
  onClose: () => void;
}

function CreateRunModal({ connector, onClose }: CreateRunModalProps) {
  const queryClient = useQueryClient();
  const [datasetType, setDatasetType] = useState(connector.dataset_types[0] ?? "");
  const [mode, setMode] = useState<SyncMode>("SAMPLE");

  const mutation = useMutation({
    mutationFn: (req: CreateRunRequest) => governmentService.createRun(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["gov-runs"] });
      onClose();
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    mutation.mutate({ connector_key: connector.key, dataset_type: datasetType, mode });
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50"
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
    >
      <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
        <h2 id="modal-title" className="mb-4 text-lg font-semibold text-slate-800">
          Buat Sync Run — {connector.name}
        </h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="dataset-type" className="mb-1 block text-sm font-medium text-slate-700">
              Dataset Type
            </label>
            <select
              id="dataset-type"
              value={datasetType}
              onChange={(e) => setDatasetType(e.target.value)}
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            >
              {connector.dataset_types.map((dt) => (
                <option key={dt} value={dt}>{dt}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Mode</label>
            <div className="flex gap-3">
              {(["SAMPLE", "DRY_RUN", "COMMIT"] as SyncMode[]).map((m) => (
                <label key={m} className="flex cursor-pointer items-center gap-2 text-sm">
                  <input
                    type="radio"
                    name="mode"
                    value={m}
                    checked={mode === m}
                    onChange={() => setMode(m)}
                    className="accent-blue-600"
                  />
                  <span className={
                    m === "SAMPLE"  ? "text-purple-700" :
                    m === "DRY_RUN" ? "text-orange-700" :
                    "text-teal-700"
                  }>
                    {m === "SAMPLE" ? "Sampel" : m === "DRY_RUN" ? "Dry Run" : "Commit"}
                  </span>
                </label>
              ))}
            </div>
            <p className="mt-1 text-xs text-slate-500">
              {mode === "SAMPLE"  && "Ambil 5 contoh data dari sumber, tanpa menulis ke database."}
              {mode === "DRY_RUN" && "Validasi semua data, tampilkan preview, tanpa menyimpan."}
              {mode === "COMMIT"  && "Validasi dan simpan semua data ke database."}
            </p>
          </div>

          {mutation.isError && (
            <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-600">
              Gagal membuat run. Coba lagi.
            </p>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md px-4 py-2 text-sm text-slate-600 hover:bg-slate-100"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={mutation.isPending}
              className="flex items-center gap-2 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60"
            >
              {mutation.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
              Jalankan
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// ConnectorCard
// ---------------------------------------------------------------------------

function ConnectorCard({ connector }: { connector: ConnectorDefinition }) {
  const [showModal, setShowModal] = useState(false);
  const stateMap: Record<string, { label: string; cls: string }> = {
    NOT_CONFIGURED: { label: "Tidak Dikonfigurasi", cls: "bg-slate-100 text-slate-500" },
    SANDBOX_SAMPLE: { label: "Sandbox",             cls: "bg-purple-100 text-purple-700" },
    ACTIVE:         { label: "Aktif",               cls: "bg-green-100 text-green-700" },
  };
  const stateInfo = stateMap[connector.state] ?? { label: connector.state, cls: "bg-slate-100 text-slate-500" };

  return (
    <>
      <div className="flex items-start justify-between rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-semibold text-slate-800">{connector.name}</span>
            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${stateInfo.cls}`}>
              {stateInfo.label}
            </span>
          </div>
          <p className="mt-1 text-sm text-slate-500">{connector.description}</p>
          <div className="mt-2 flex flex-wrap gap-1.5">
            {connector.dataset_types.map((dt) => (
              <span
                key={dt}
                className="rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700"
              >
                {dt}
              </span>
            ))}
          </div>
        </div>
        <button
          onClick={() => setShowModal(true)}
          title="Buat sync run baru"
          className="ml-4 flex items-center gap-1.5 rounded-md bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700 shrink-0"
          aria-label={`Buat sync run untuk ${connector.name}`}
        >
          <Play className="h-3.5 w-3.5" />
          Sync
        </button>
      </div>
      {showModal && (
        <CreateRunModal connector={connector} onClose={() => setShowModal(false)} />
      )}
    </>
  );
}

// ---------------------------------------------------------------------------
// MatchStatusBadge
// ---------------------------------------------------------------------------

function MatchStatusBadge({ status }: { status: MatchStatus }) {
  const map: Record<MatchStatus, { label: string; cls: string }> = {
    PENDING_MATCH: { label: "Menunggu",  cls: "bg-yellow-100 text-yellow-700" },
    MATCHED:       { label: "Terhubung", cls: "bg-green-100 text-green-700" },
    REJECTED:      { label: "Ditolak",   cls: "bg-red-100 text-red-700" },
  };
  const { label, cls } = map[status] ?? { label: status, cls: "bg-slate-100 text-slate-600" };
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${cls}`}>
      {label}
    </span>
  );
}

// ---------------------------------------------------------------------------
// ConfidenceBadge
// ---------------------------------------------------------------------------

function ConfidenceBadge({ confidence, reason }: { confidence: number; reason: string }) {
  const cls =
    confidence >= 90 ? "bg-green-100 text-green-700" :
    confidence >= 60 ? "bg-blue-100 text-blue-700" :
    "bg-slate-100 text-slate-500";
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${cls}`}>
      {reason.replace(/_/g, " ")} · {confidence}%
    </span>
  );
}

// ---------------------------------------------------------------------------
// CandidateList — shown inside the detail drawer
// ---------------------------------------------------------------------------

interface CandidateListProps {
  mappingID: string;
  datasetType: string;
  onMatched: () => void;
}

function CandidateList({ mappingID, datasetType, onMatched }: CandidateListProps) {
  const queryClient = useQueryClient();

  const { data, isLoading, isError } = useQuery({
    queryKey: ["gov-candidates", mappingID],
    queryFn: () => governmentService.getMappingCandidates(mappingID),
  });

  const matchMutation = useMutation({
    mutationFn: (candidate: ResolutionCandidate) =>
      governmentService.matchMapping(mappingID, {
        internal_entity_id:   candidate.entity_id,
        internal_entity_type: candidate.entity_type,
        match_reason:         candidate.reason,
        match_confidence:     candidate.confidence,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["gov-pending"] });
      queryClient.invalidateQueries({ queryKey: ["gov-mapping", mappingID] });
      onMatched();
    },
  });

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-4 text-sm text-slate-500">
        <Loader2 className="h-4 w-4 animate-spin" /> Mencari kandidat...
      </div>
    );
  }

  if (isError) {
    return <p className="py-2 text-xs text-red-500">Gagal memuat kandidat.</p>;
  }

  const candidates = data?.data ?? [];

  if (candidates.length === 0) {
    return (
      <p className="py-3 text-xs text-slate-400">
        Tidak ada kandidat ditemukan untuk dataset <strong>{datasetType}</strong>.
      </p>
    );
  }

  return (
    <div className="mt-2 flex flex-col gap-2">
      {candidates.map((c) => (
        <div
          key={c.entity_id}
          className="flex items-center justify-between rounded-md border border-slate-200 bg-slate-50 px-3 py-2"
        >
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium text-slate-800">{c.name}</p>
            {c.code && (
              <p className="mt-0.5 font-mono text-xs text-slate-500">{c.code}</p>
            )}
            <div className="mt-1">
              <ConfidenceBadge confidence={c.confidence} reason={c.reason} />
            </div>
          </div>
          <button
            onClick={() => matchMutation.mutate(c)}
            disabled={matchMutation.isPending}
            className="ml-3 flex items-center gap-1 rounded-md bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            aria-label={`Hubungkan ke ${c.name}`}
          >
            {matchMutation.isPending ? (
              <Loader2 className="h-3 w-3 animate-spin" />
            ) : (
              <Link2 className="h-3 w-3" />
            )}
            Hubungkan
          </button>
        </div>
      ))}
    </div>
  );
}

// ---------------------------------------------------------------------------
// MappingDrawer — detail panel for a single mapping
// ---------------------------------------------------------------------------

interface MappingDrawerProps {
  mapping: ExternalMapping;
  onClose: () => void;
  onAction: () => void;
}

function MappingDrawer({ mapping, onClose, onAction }: MappingDrawerProps) {
  const queryClient = useQueryClient();
  const [rejectReason, setRejectReason] = useState("");
  const [showRejectForm, setShowRejectForm] = useState(false);

  const unmatchMutation = useMutation({
    mutationFn: () => governmentService.unmatchMapping(mapping.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["gov-pending"] });
      queryClient.invalidateQueries({ queryKey: ["gov-mapping", mapping.id] });
      onAction();
    },
  });

  const rejectMutation = useMutation({
    mutationFn: () => governmentService.rejectMapping(mapping.id, { reject_reason: rejectReason }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["gov-pending"] });
      onAction();
    },
  });

  return (
    <div
      className="fixed inset-y-0 right-0 z-40 flex w-full max-w-lg flex-col bg-white shadow-xl"
      role="dialog"
      aria-modal="true"
      aria-label="Detail mapping"
    >
      {/* Header */}
      <div className="flex items-center justify-between border-b border-slate-200 px-5 py-4">
        <div className="flex items-center gap-2">
          <GitMerge className="h-4 w-4 text-blue-600" />
          <span className="font-semibold text-slate-800 text-sm">Detail Mapping</span>
        </div>
        <button
          onClick={onClose}
          className="rounded p-1 text-slate-400 hover:bg-slate-100"
          aria-label="Tutup"
        >
          <XCircle className="h-5 w-5" />
        </button>
      </div>

      {/* Body — scrollable */}
      <div className="flex-1 overflow-y-auto px-5 py-4 space-y-5">
        {/* Identity */}
        <div className="space-y-1.5">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Identitas Eksternal</p>
          <div className="rounded-md bg-slate-50 px-3 py-2 space-y-1">
            <div className="flex justify-between text-xs">
              <span className="text-slate-500">External ID</span>
              <span className="font-mono text-slate-800 truncate max-w-[60%]">{mapping.external_id}</span>
            </div>
            <div className="flex justify-between text-xs">
              <span className="text-slate-500">Connector</span>
              <span className="text-slate-700">{mapping.connector_key}</span>
            </div>
            <div className="flex justify-between text-xs">
              <span className="text-slate-500">Dataset Type</span>
              <span className="text-slate-700">{mapping.dataset_type}</span>
            </div>
            <div className="flex justify-between text-xs">
              <span className="text-slate-500">Entity Type</span>
              <span className="text-slate-700">{mapping.internal_entity_type}</span>
            </div>
          </div>
        </div>

        {/* Match status */}
        <div className="space-y-1.5">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Status Resolusi</p>
          <div className="flex items-center gap-2">
            <MatchStatusBadge status={mapping.match_status} />
            {mapping.match_confidence != null && mapping.match_reason && (
              <ConfidenceBadge confidence={mapping.match_confidence} reason={mapping.match_reason} />
            )}
          </div>
          {mapping.matched_at && (
            <p className="text-xs text-slate-500">
              Dihubungkan: {fmtDate(mapping.matched_at)}
            </p>
          )}
          {mapping.rejected_at && (
            <p className="text-xs text-slate-500">
              Ditolak: {fmtDate(mapping.rejected_at)}
              {mapping.reject_reason && ` — ${mapping.reject_reason}`}
            </p>
          )}
          {mapping.internal_entity_id && (
            <p className="text-xs text-slate-500 font-mono break-all">
              Internal ID: {mapping.internal_entity_id}
            </p>
          )}
        </div>

        {/* Candidates — only for PENDING_MATCH and REJECTED */}
        {(mapping.match_status === "PENDING_MATCH" || mapping.match_status === "REJECTED") && (
          <div className="space-y-1.5">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Kandidat Internal
            </p>
            <CandidateList
              mappingID={mapping.id}
              datasetType={mapping.dataset_type}
              onMatched={onAction}
            />
          </div>
        )}

        {/* Timestamps */}
        <div className="space-y-1">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Waktu</p>
          <div className="rounded-md bg-slate-50 px-3 py-2 space-y-1">
            <div className="flex justify-between text-xs">
              <span className="text-slate-500">Dibuat</span>
              <span className="text-slate-700">{fmtDate(mapping.created_at)}</span>
            </div>
            <div className="flex justify-between text-xs">
              <span className="text-slate-500">Terakhir Dilihat</span>
              <span className="text-slate-700">{fmtDate(mapping.last_seen_at)}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Footer actions */}
      <div className="border-t border-slate-200 px-5 py-3 space-y-2">
        {/* Unmatch — only if MATCHED */}
        {mapping.match_status === "MATCHED" && (
          <button
            onClick={() => unmatchMutation.mutate()}
            disabled={unmatchMutation.isPending}
            className="flex w-full items-center justify-center gap-2 rounded-md border border-slate-200 py-2 text-sm text-slate-600 hover:bg-slate-50 disabled:opacity-50"
          >
            {unmatchMutation.isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Link2Off className="h-4 w-4" />
            )}
            Lepas Hubungan
          </button>
        )}

        {/* Reject — only for PENDING_MATCH */}
        {mapping.match_status === "PENDING_MATCH" && (
          <>
            {showRejectForm ? (
              <div className="space-y-2">
                <input
                  type="text"
                  value={rejectReason}
                  onChange={(e) => setRejectReason(e.target.value)}
                  placeholder="Alasan penolakan (opsional)"
                  className="w-full rounded-md border border-slate-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-red-400"
                />
                <div className="flex gap-2">
                  <button
                    onClick={() => rejectMutation.mutate()}
                    disabled={rejectMutation.isPending}
                    className="flex flex-1 items-center justify-center gap-1 rounded-md bg-red-600 py-1.5 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
                  >
                    {rejectMutation.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : <XOctagon className="h-3 w-3" />}
                    Konfirmasi Tolak
                  </button>
                  <button
                    onClick={() => setShowRejectForm(false)}
                    className="rounded-md border border-slate-200 px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-50"
                  >
                    Batal
                  </button>
                </div>
              </div>
            ) : (
              <button
                onClick={() => setShowRejectForm(true)}
                className="flex w-full items-center justify-center gap-2 rounded-md border border-red-200 py-2 text-sm text-red-600 hover:bg-red-50"
              >
                <XOctagon className="h-4 w-4" />
                Tolak Mapping
              </button>
            )}
          </>
        )}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// ResolutionTab
// ---------------------------------------------------------------------------

function ResolutionTab() {
  const [page, setPage]                   = useState(1);
  const [filterConnector, setFilterConnector] = useState("");
  const [filterDataset, setFilterDataset]     = useState("");
  const [selectedMapping, setSelectedMapping] = useState<ExternalMapping | null>(null);

  const queryClient = useQueryClient();

  const { data, isLoading, isError } = useQuery({
    queryKey: ["gov-pending", page, filterConnector, filterDataset],
    queryFn: () =>
      governmentService.listPendingMappings({
        page,
        page_size: 20,
        connector_key: filterConnector || undefined,
        dataset_type:  filterDataset  || undefined,
      }),
  });

  const handleAction = () => {
    setSelectedMapping(null);
    queryClient.invalidateQueries({ queryKey: ["gov-pending"] });
  };

  return (
    <div>
      {/* Filter bar */}
      <div className="mb-4 flex flex-wrap items-center gap-3">
        <input
          type="text"
          value={filterConnector}
          onChange={(e) => { setFilterConnector(e.target.value); setPage(1); }}
          placeholder="Filter connector key..."
          className="rounded-md border border-slate-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
        />
        <input
          type="text"
          value={filterDataset}
          onChange={(e) => { setFilterDataset(e.target.value); setPage(1); }}
          placeholder="Filter dataset type..."
          className="rounded-md border border-slate-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
        />
        <button
          onClick={() => queryClient.invalidateQueries({ queryKey: ["gov-pending"] })}
          className="flex items-center gap-1 rounded-md border border-slate-200 px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-50"
        >
          <RefreshCw className="h-3.5 w-3.5" />
          Muat Ulang
        </button>
        {data && (
          <span className="text-sm text-slate-500">
            {data.meta.total} mapping menunggu resolusi
          </span>
        )}
      </div>

      {/* Table */}
      {isLoading ? (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-6 w-6 animate-spin text-blue-500" />
        </div>
      ) : isError ? (
        <div className="flex flex-col items-center gap-2 py-16 text-slate-400">
          <AlertTriangle className="h-8 w-8 text-red-400" />
          <p className="text-sm">Gagal memuat data resolusi.</p>
        </div>
      ) : !data?.data?.length ? (
        <div className="flex flex-col items-center gap-3 py-16 text-slate-400">
          <CheckCircle2 className="h-8 w-8 text-green-400" />
          <p className="text-sm">Tidak ada mapping yang menunggu resolusi.</p>
        </div>
      ) : (
        <>
          <div className="overflow-x-auto rounded-md border border-slate-200">
            <table className="min-w-full text-xs">
              <thead className="bg-slate-50">
                <tr>
                  {["External ID", "Connector", "Dataset", "Entity Type", "Status", "Terakhir Dilihat", ""].map((h) => (
                    <th key={h} className="px-3 py-2 text-left font-medium text-slate-600">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {data.data.map((m: ExternalMapping) => (
                  <tr
                    key={m.id}
                    className="cursor-pointer hover:bg-slate-50"
                    onClick={() => setSelectedMapping(m)}
                  >
                    <td className="px-3 py-2 font-mono text-slate-700 max-w-[160px] truncate">
                      {m.external_id}
                    </td>
                    <td className="px-3 py-2 text-slate-600">{m.connector_key}</td>
                    <td className="px-3 py-2 text-slate-600">{m.dataset_type}</td>
                    <td className="px-3 py-2 text-slate-600">{m.internal_entity_type}</td>
                    <td className="px-3 py-2">
                      <MatchStatusBadge status={m.match_status} />
                    </td>
                    <td className="px-3 py-2 text-slate-500">{fmtDate(m.last_seen_at)}</td>
                    <td className="px-3 py-2">
                      <button
                        onClick={(e) => { e.stopPropagation(); setSelectedMapping(m); }}
                        className="rounded px-2 py-1 text-blue-600 hover:bg-blue-50 text-xs font-medium"
                        aria-label="Buka detail"
                      >
                        Resolusi →
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {data.meta.total > 20 && (
            <div className="mt-3 flex items-center justify-between text-sm text-slate-500">
              <span>Total {data.meta.total} mapping</span>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="rounded px-3 py-1 hover:bg-slate-100 disabled:opacity-40"
                >← Prev</button>
                <span>Hal. {page}</span>
                <button
                  onClick={() => setPage((p) => p + 1)}
                  disabled={page * 20 >= data.meta.total}
                  className="rounded px-3 py-1 hover:bg-slate-100 disabled:opacity-40"
                >Next →</button>
              </div>
            </div>
          )}
        </>
      )}

      {/* Drawer overlay */}
      {selectedMapping && (
        <>
          <div
            className="fixed inset-0 z-30 bg-slate-900/30"
            onClick={() => setSelectedMapping(null)}
            aria-hidden="true"
          />
          <MappingDrawer
            mapping={selectedMapping}
            onClose={() => setSelectedMapping(null)}
            onAction={handleAction}
          />
        </>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// MappingsTab
// ---------------------------------------------------------------------------

function MappingsTab() {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useQuery({
    queryKey: ["gov-mappings", page],
    queryFn: () => governmentService.listMappings({ page, page_size: 15 }),
  });

  if (isLoading) return <p className="py-8 text-center text-sm text-slate-500">Memuat mappings…</p>;
  if (!data?.data?.length) return <p className="py-8 text-center text-sm text-slate-400">Belum ada external mappings.</p>;

  return (
    <div>
      <div className="overflow-x-auto rounded-lg border border-slate-200">
        <table className="min-w-full text-sm">
          <thead className="bg-slate-50">
            <tr>
              {["Connector", "Dataset", "External ID", "Entity Type", "Terakhir Dilihat"].map((h) => (
                <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wide">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {data.data.map((m: ExternalMapping) => (
              <tr key={m.id} className="hover:bg-slate-50">
                <td className="px-4 py-2.5 text-slate-700">{m.connector_key}</td>
                <td className="px-4 py-2.5 text-slate-600">{m.dataset_type}</td>
                <td className="px-4 py-2.5 font-mono text-xs text-slate-700">{m.external_id}</td>
                <td className="px-4 py-2.5 text-slate-600">{m.internal_entity_type}</td>
                <td className="px-4 py-2.5 text-slate-500">{fmtDate(m.last_seen_at)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {data.meta.total > 15 && (
        <div className="mt-3 flex items-center justify-between text-xs text-slate-500">
          <span>Total {data.meta.total} mappings</span>
          <div className="flex gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="rounded px-2 py-1 hover:bg-slate-100 disabled:opacity-40"
            >← Prev</button>
            <span>Hal. {page}</span>
            <button
              onClick={() => setPage((p) => p + 1)}
              disabled={page * 15 >= data.meta.total}
              className="rounded px-2 py-1 hover:bg-slate-100 disabled:opacity-40"
            >Next →</button>
          </div>
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Page
// ---------------------------------------------------------------------------

type Tab = "connectors" | "runs" | "mappings" | "resolution";

export default function GovernmentConnectorsPage() {
  const [activeTab, setActiveTab] = useState<Tab>("connectors");
  const [runsPage, setRunsPage] = useState(1);
  const queryClient = useQueryClient();

  const { data: connectors, isLoading: connectorsLoading } = useQuery({
    queryKey: ["gov-connectors"],
    queryFn: () => governmentService.listConnectors(),
  });

  const { data: runsData, isLoading: runsLoading } = useQuery({
    queryKey: ["gov-runs", runsPage],
    queryFn: () => governmentService.listRuns({ page: runsPage, page_size: 10 }),
    enabled: activeTab === "runs",
  });

  const tabs: { key: Tab; label: string }[] = [
    { key: "connectors", label: "Connectors" },
    { key: "runs",       label: "Riwayat Sync" },
    { key: "mappings",   label: "Mappings" },
    { key: "resolution", label: "Resolusi Entitas" },
  ];

  return (
    <DashboardLayout title="Government Connectors">
    <div className="mx-auto max-w-5xl">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100">
            <Building2 className="h-5 w-5 text-blue-700" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-slate-900">Government Connectors</h1>
            <p className="text-sm text-slate-500">Sinkronisasi data dari SIRUP, OM-SPAN, dan referensi pemerintah</p>
          </div>
        </div>
        <button
          onClick={() => queryClient.invalidateQueries({ queryKey: ["gov-connectors", "gov-runs", "gov-mappings"] })}
          title="Muat ulang"
          className="flex items-center gap-1.5 rounded-md border border-slate-200 px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-50"
          aria-label="Muat ulang data"
        >
          <RefreshCw className="h-4 w-4" />
          Muat Ulang
        </button>
      </div>

      {/* Tabs */}
      <div className="mb-6 border-b border-slate-200">
        <nav className="-mb-px flex gap-1" aria-label="Tab navigasi">
          {tabs.map(({ key, label }) => (
            <button
              key={key}
              onClick={() => setActiveTab(key)}
              className={`px-4 py-2.5 text-sm font-medium transition-colors ${
                activeTab === key
                  ? "border-b-2 border-blue-600 text-blue-700"
                  : "text-slate-500 hover:text-slate-800"
              }`}
              aria-selected={activeTab === key}
              role="tab"
            >
              {label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab: Connectors */}
      {activeTab === "connectors" && (
        <div>
          {connectorsLoading ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="h-6 w-6 animate-spin text-blue-500" />
            </div>
          ) : !connectors?.length ? (
            <p className="py-12 text-center text-sm text-slate-400">Tidak ada connector tersedia.</p>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2">
              {connectors.map((c) => (
                <ConnectorCard key={c.key} connector={c} />
              ))}
            </div>
          )}
        </div>
      )}

      {/* Tab: Runs */}
      {activeTab === "runs" && (
        <div>
          {runsLoading ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="h-6 w-6 animate-spin text-blue-500" />
            </div>
          ) : !runsData?.data?.length ? (
            <div className="flex flex-col items-center gap-3 py-16 text-slate-400">
              <Clock className="h-8 w-8" />
              <p className="text-sm">Belum ada sync run. Buat run baru dari tab Connectors.</p>
            </div>
          ) : (
            <>
              <div className="flex flex-col gap-3">
                {runsData.data.map((run) => (
                  <RunCard key={run.id} run={run} />
                ))}
              </div>
              {runsData.meta.total > 10 && (
                <div className="mt-4 flex items-center justify-between text-sm text-slate-500">
                  <span>Total {runsData.meta.total} runs</span>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setRunsPage((p) => Math.max(1, p - 1))}
                      disabled={runsPage === 1}
                      className="rounded px-3 py-1 hover:bg-slate-100 disabled:opacity-40"
                    >← Prev</button>
                    <span>Hal. {runsPage}</span>
                    <button
                      onClick={() => setRunsPage((p) => p + 1)}
                      disabled={runsPage * 10 >= runsData.meta.total}
                      className="rounded px-3 py-1 hover:bg-slate-100 disabled:opacity-40"
                    >Next →</button>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      )}

      {/* Tab: Mappings */}
      {activeTab === "mappings" && <MappingsTab />}

      {/* Tab: Resolution */}
      {activeTab === "resolution" && <ResolutionTab />}
    </div>
    </DashboardLayout>
  );
}
