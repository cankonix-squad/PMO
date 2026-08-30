"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Box,
  Plus,
  Loader2,
  Trash2,
  RefreshCw,
  ChevronDown,
  ChevronRight,
  Link2,
  Unlink,
  History,
  ExternalLink,
  AlertTriangle,
} from "lucide-react";
import { bimService } from "@/services/bim.service";
import type {
  BIMModel,
  BIMModelStatus,
  BIMDiscipline,
  BIMProvider,
  BIMModelRole,
  CreateBIMModelRequest,
  CreateVersionRequest,
  LinkProjectRequest,
} from "@/types/bim";
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

function fmtBytes(n: number): string {
  if (n === 0) return "—";
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

// ---------------------------------------------------------------------------
// Status / discipline badge
// ---------------------------------------------------------------------------

function StatusBadge({ status }: { status: BIMModelStatus }) {
  const map: Record<BIMModelStatus, { label: string; cls: string }> = {
    DRAFT:    { label: "Draft",    cls: "bg-slate-100 text-slate-600" },
    ACTIVE:   { label: "Aktif",   cls: "bg-green-100 text-green-700" },
    ARCHIVED: { label: "Arsip",   cls: "bg-yellow-100 text-yellow-700" },
  };
  const { label, cls } = map[status] ?? { label: status, cls: "bg-slate-100 text-slate-600" };
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${cls}`}>
      {label}
    </span>
  );
}

function DisciplineBadge({ discipline }: { discipline: BIMDiscipline }) {
  const colours: Record<BIMDiscipline, string> = {
    ARCHITECTURAL: "bg-purple-100 text-purple-700",
    STRUCTURAL:    "bg-blue-100 text-blue-700",
    MEP:           "bg-orange-100 text-orange-700",
    CIVIL:         "bg-cyan-100 text-cyan-700",
    LANDSCAPE:     "bg-green-100 text-green-700",
    OTHER:         "bg-slate-100 text-slate-600",
  };
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${colours[discipline] ?? "bg-slate-100 text-slate-600"}`}>
      {discipline}
    </span>
  );
}

// ---------------------------------------------------------------------------
// Register BIM Model modal
// ---------------------------------------------------------------------------

const DISCIPLINES: BIMDiscipline[] = [
  "ARCHITECTURAL", "STRUCTURAL", "MEP", "CIVIL", "LANDSCAPE", "OTHER",
];
const PROVIDERS: BIMProvider[] = [
  "autodesk_bim360", "trimble_connect", "bentley_projectwise", "local", "other",
];

function RegisterModelModal({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient();
  const [form, setForm] = useState<CreateBIMModelRequest>({
    name: "",
    description: "",
    discipline: "OTHER",
    provider: "other",
    external_model_id: "",
    viewer_url: "",
  });

  const mutation = useMutation({
    mutationFn: () => bimService.createModel(form),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["bim-models"] });
      onClose();
    },
  });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="w-full max-w-lg rounded-xl bg-white p-6 shadow-2xl">
        <h2 className="mb-4 text-lg font-semibold text-slate-800">Daftarkan BIM Model</h2>

        <div className="space-y-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Nama *</label>
            <input
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="Nama model BIM"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Deskripsi</label>
            <textarea
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
              rows={2}
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              placeholder="Deskripsi singkat model"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-700">Disiplin *</label>
              <select
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
                value={form.discipline}
                onChange={(e) => setForm({ ...form, discipline: e.target.value as BIMDiscipline })}
              >
                {DISCIPLINES.map((d) => (
                  <option key={d} value={d}>{d}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="mb-1 block text-sm font-medium text-slate-700">Provider *</label>
              <select
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
                value={form.provider}
                onChange={(e) => setForm({ ...form, provider: e.target.value as BIMProvider })}
              >
                {PROVIDERS.map((p) => (
                  <option key={p} value={p}>{p}</option>
                ))}
              </select>
            </div>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">External Model ID *</label>
            <input
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
              value={form.external_model_id}
              onChange={(e) => setForm({ ...form, external_model_id: e.target.value })}
              placeholder="ID model di sistem BIM provider"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Viewer URL</label>
            <input
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
              value={form.viewer_url}
              onChange={(e) => setForm({ ...form, viewer_url: e.target.value })}
              placeholder="https://..."
            />
            <p className="mt-1 text-xs text-slate-500">Link viewer (bukan link download file)</p>
          </div>
        </div>

        {mutation.isError && (
          <div className="mt-3 flex items-center gap-2 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">
            <AlertTriangle className="h-4 w-4 shrink-0" />
            Gagal mendaftarkan model. Coba lagi.
          </div>
        )}

        <div className="mt-5 flex justify-end gap-2">
          <button
            className="rounded-lg border border-slate-300 px-4 py-2 text-sm text-slate-700 hover:bg-slate-50"
            onClick={onClose}
          >
            Batal
          </button>
          <button
            className="flex items-center gap-1.5 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
            disabled={mutation.isPending || !form.name || !form.external_model_id}
            onClick={() => mutation.mutate()}
          >
            {mutation.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
            Daftarkan
          </button>
        </div>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Version panel (inline)
// ---------------------------------------------------------------------------

function VersionPanel({ modelId }: { modelId: string }) {
  const qc = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState<CreateVersionRequest>({ version_label: "" });

  const { data, isLoading } = useQuery({
    queryKey: ["bim-versions", modelId],
    queryFn: () => bimService.listVersions(modelId),
  });

  const addMutation = useMutation({
    mutationFn: () => bimService.addVersion(modelId, form),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["bim-versions", modelId] });
      setShowForm(false);
      setForm({ version_label: "" });
    },
  });

  return (
    <div className="mt-3 rounded-lg border border-slate-200 bg-slate-50 p-3">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-wide text-slate-500">
          Riwayat Versi
        </span>
        <button
          className="flex items-center gap-1 rounded-md bg-white border border-slate-300 px-2 py-1 text-xs text-slate-700 hover:bg-slate-50"
          onClick={() => setShowForm(!showForm)}
        >
          <Plus className="h-3 w-3" /> Tambah
        </button>
      </div>

      {showForm && (
        <div className="mb-3 grid grid-cols-3 gap-2 rounded-lg border border-indigo-200 bg-white p-3">
          <input
            className="col-span-2 rounded border border-slate-300 px-2 py-1 text-xs focus:outline-none focus:border-indigo-500"
            placeholder="Label versi, mis. v1.0 atau Rev A"
            value={form.version_label}
            onChange={(e) => setForm({ ...form, version_label: e.target.value })}
          />
          <input
            className="rounded border border-slate-300 px-2 py-1 text-xs focus:outline-none focus:border-indigo-500"
            placeholder="Ringkasan perubahan"
            value={form.change_summary ?? ""}
            onChange={(e) => setForm({ ...form, change_summary: e.target.value })}
          />
          <button
            className="col-span-3 flex items-center justify-center gap-1 rounded bg-indigo-600 px-2 py-1 text-xs font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
            disabled={addMutation.isPending || !form.version_label}
            onClick={() => addMutation.mutate()}
          >
            {addMutation.isPending && <Loader2 className="h-3 w-3 animate-spin" />}
            Simpan Versi
          </button>
        </div>
      )}

      {isLoading ? (
        <div className="flex justify-center py-3">
          <Loader2 className="h-4 w-4 animate-spin text-slate-400" />
        </div>
      ) : (data?.data ?? []).length === 0 ? (
        <p className="text-center text-xs text-slate-400 py-2">Belum ada versi.</p>
      ) : (
        <div className="space-y-1.5">
          {(data?.data ?? []).map((v) => (
            <div key={v.id} className="flex items-center justify-between rounded bg-white border border-slate-200 px-2.5 py-1.5 text-xs">
              <div>
                <span className="font-medium text-slate-700">{v.version_label}</span>
                {v.change_summary && (
                  <span className="ml-2 text-slate-500">{v.change_summary}</span>
                )}
              </div>
              <div className="flex items-center gap-3 text-slate-400">
                {v.file_size_bytes > 0 && <span>{fmtBytes(v.file_size_bytes)}</span>}
                <span>{fmtDate(v.created_at)}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Model row (expandable)
// ---------------------------------------------------------------------------

const ROLES: BIMModelRole[] = ["PRIMARY", "REFERENCE", "ASBUILT", "OTHER"];

function ModelRow({ model }: { model: BIMModel }) {
  const qc = useQueryClient();
  const [expanded, setExpanded] = useState(false);
  const [tab, setTab] = useState<"versions" | "mappings">("versions");
  const [showLinkForm, setShowLinkForm] = useState(false);
  const [linkForm, setLinkForm] = useState<LinkProjectRequest>({
    project_id: "",
    model_role: "REFERENCE",
    notes: "",
  });

  const { data: mappingsData } = useQuery({
    queryKey: ["bim-mappings", model.id],
    queryFn: () => bimService.listMappings(model.id),
    enabled: expanded && tab === "mappings",
  });

  const deleteMutation = useMutation({
    mutationFn: () => bimService.deleteModel(model.id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["bim-models"] }),
  });

  const linkMutation = useMutation({
    mutationFn: () => bimService.linkProject(model.id, linkForm),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["bim-mappings", model.id] });
      setShowLinkForm(false);
      setLinkForm({ project_id: "", model_role: "REFERENCE", notes: "" });
    },
  });

  const unlinkMutation = useMutation({
    mutationFn: (projectId: string) => bimService.unlinkProject(model.id, projectId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["bim-mappings", model.id] }),
  });

  return (
    <div className="rounded-xl border border-slate-200 bg-white shadow-sm">
      {/* Header row */}
      <div className="flex items-center gap-3 p-4">
        <button
          className="text-slate-400 hover:text-slate-600"
          onClick={() => setExpanded(!expanded)}
          aria-label={expanded ? "Tutup detail" : "Buka detail"}
        >
          {expanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
        </button>

        <Box className="h-5 w-5 text-indigo-500 shrink-0" />

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="font-medium text-slate-800 truncate">{model.name}</span>
            <StatusBadge status={model.status} />
            <DisciplineBadge discipline={model.discipline} />
            <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-500">
              {model.provider}
            </span>
          </div>
          {model.description && (
            <p className="mt-0.5 text-xs text-slate-500 truncate">{model.description}</p>
          )}
        </div>

        <div className="flex items-center gap-2 shrink-0">
          {model.viewer_url && (
            <a
              href={model.viewer_url}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1 rounded-md border border-slate-300 px-2 py-1 text-xs text-slate-600 hover:bg-slate-50"
              aria-label="Buka viewer"
            >
              <ExternalLink className="h-3 w-3" />
              Viewer
            </a>
          )}
          <button
            className="rounded-md border border-red-200 p-1.5 text-red-400 hover:bg-red-50 disabled:opacity-40"
            onClick={() => deleteMutation.mutate()}
            disabled={deleteMutation.isPending}
            aria-label="Hapus model"
          >
            {deleteMutation.isPending
              ? <Loader2 className="h-3.5 w-3.5 animate-spin" />
              : <Trash2 className="h-3.5 w-3.5" />}
          </button>
        </div>
      </div>

      {/* Expanded detail */}
      {expanded && (
        <div className="border-t border-slate-100 px-4 pb-4">
          {/* Tab bar */}
          <div className="mt-3 flex gap-1 border-b border-slate-200">
            {(["versions", "mappings"] as const).map((t) => (
              <button
                key={t}
                className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium border-b-2 -mb-px transition-colors ${
                  tab === t
                    ? "border-indigo-500 text-indigo-600"
                    : "border-transparent text-slate-500 hover:text-slate-700"
                }`}
                onClick={() => setTab(t)}
              >
                {t === "versions" ? <History className="h-3.5 w-3.5" /> : <Link2 className="h-3.5 w-3.5" />}
                {t === "versions" ? "Versi" : "Proyek Terhubung"}
              </button>
            ))}
          </div>

          {/* Tab content */}
          {tab === "versions" && <VersionPanel modelId={model.id} />}

          {tab === "mappings" && (
            <div className="mt-3 rounded-lg border border-slate-200 bg-slate-50 p-3">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                  Proyek Terhubung
                </span>
                <button
                  className="flex items-center gap-1 rounded-md bg-white border border-slate-300 px-2 py-1 text-xs text-slate-700 hover:bg-slate-50"
                  onClick={() => setShowLinkForm(!showLinkForm)}
                >
                  <Link2 className="h-3 w-3" /> Hubungkan Proyek
                </button>
              </div>

              {showLinkForm && (
                <div className="mb-3 grid grid-cols-3 gap-2 rounded-lg border border-indigo-200 bg-white p-3">
                  <input
                    className="col-span-2 rounded border border-slate-300 px-2 py-1 text-xs focus:outline-none focus:border-indigo-500"
                    placeholder="UUID Proyek"
                    value={linkForm.project_id}
                    onChange={(e) => setLinkForm({ ...linkForm, project_id: e.target.value })}
                  />
                  <select
                    className="rounded border border-slate-300 px-2 py-1 text-xs focus:outline-none focus:border-indigo-500"
                    value={linkForm.model_role}
                    onChange={(e) => setLinkForm({ ...linkForm, model_role: e.target.value as BIMModelRole })}
                  >
                    {ROLES.map((r) => <option key={r} value={r}>{r}</option>)}
                  </select>
                  <input
                    className="col-span-3 rounded border border-slate-300 px-2 py-1 text-xs focus:outline-none focus:border-indigo-500"
                    placeholder="Catatan (opsional)"
                    value={linkForm.notes ?? ""}
                    onChange={(e) => setLinkForm({ ...linkForm, notes: e.target.value })}
                  />
                  <button
                    className="col-span-3 flex items-center justify-center gap-1 rounded bg-indigo-600 px-2 py-1 text-xs font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
                    disabled={linkMutation.isPending || !linkForm.project_id}
                    onClick={() => linkMutation.mutate()}
                  >
                    {linkMutation.isPending && <Loader2 className="h-3 w-3 animate-spin" />}
                    Hubungkan
                  </button>
                </div>
              )}

              {(mappingsData?.data ?? []).length === 0 ? (
                <p className="text-center text-xs text-slate-400 py-2">Belum ada proyek terhubung.</p>
              ) : (
                <div className="space-y-1.5">
                  {(mappingsData?.data ?? []).map((m) => (
                    <div key={m.id} className="flex items-center justify-between rounded bg-white border border-slate-200 px-2.5 py-1.5 text-xs">
                      <div>
                        <span className="font-mono text-slate-600">{m.project_id}</span>
                        <span className={`ml-2 rounded-full px-1.5 py-0.5 text-xs font-medium ${
                          m.model_role === "PRIMARY" ? "bg-indigo-100 text-indigo-700" : "bg-slate-100 text-slate-600"
                        }`}>{m.model_role}</span>
                        {m.notes && <span className="ml-2 text-slate-400">{m.notes}</span>}
                      </div>
                      <button
                        className="ml-2 rounded p-1 text-red-400 hover:bg-red-50 disabled:opacity-40"
                        onClick={() => unlinkMutation.mutate(m.project_id)}
                        disabled={unlinkMutation.isPending}
                        aria-label="Putuskan koneksi proyek"
                      >
                        {unlinkMutation.isPending
                          ? <Loader2 className="h-3 w-3 animate-spin" />
                          : <Unlink className="h-3 w-3" />}
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main page
// ---------------------------------------------------------------------------

export default function BIMIntegrationPage() {
  const [showRegister, setShowRegister] = useState(false);

  const { data, isLoading, isError, refetch, isFetching } = useQuery({
    queryKey: ["bim-models"],
    queryFn: () => bimService.listModels(),
  });

  const models = data?.data ?? [];
  const total = data?.meta.total ?? 0;

  return (
    <DashboardLayout title="BIM / Digital Twin">
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">BIM / Digital Twin</h1>
          <p className="mt-1 text-sm text-slate-500">
            Kelola referensi model BIM eksternal dan pemetaan ke proyek PMO.
            File tidak disimpan di database — hanya metadata dan tautan viewer.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            className="flex items-center gap-1.5 rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-700 hover:bg-slate-100 disabled:opacity-50"
            onClick={() => refetch()}
            disabled={isFetching}
            aria-label="Refresh daftar model"
          >
            <RefreshCw className={`h-4 w-4 ${isFetching ? "animate-spin" : ""}`} />
            Muat Ulang
          </button>
          <button
            className="flex items-center gap-1.5 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
            onClick={() => setShowRegister(true)}
          >
            <Plus className="h-4 w-4" />
            Daftarkan Model
          </button>
        </div>
      </div>

      {/* Summary card */}
      <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
        {[
          { label: "Total Model", value: total },
          { label: "Aktif",   value: models.filter((m) => m.status === "ACTIVE").length },
          { label: "Draft",    value: models.filter((m) => m.status === "DRAFT").length },
          { label: "Arsip",    value: models.filter((m) => m.status === "ARCHIVED").length },
        ].map(({ label, value }) => (
          <div key={label} className="rounded-xl bg-white border border-slate-200 p-4 shadow-sm">
            <p className="text-xs font-medium uppercase tracking-wide text-slate-500">{label}</p>
            <p className="mt-1 text-2xl font-bold text-slate-800">{value}</p>
          </div>
        ))}
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="flex items-center justify-center py-20 text-slate-400">
          <Loader2 className="h-6 w-6 animate-spin" />
          <span className="ml-2 text-sm">Memuat model BIM…</span>
        </div>
      ) : isError ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-red-200 bg-red-50 py-12 text-center">
          <AlertTriangle className="h-8 w-8 text-red-400" />
          <p className="mt-2 text-sm font-medium text-red-700">Gagal memuat data BIM</p>
          <button
            className="mt-3 rounded-lg border border-red-300 px-4 py-1.5 text-sm text-red-600 hover:bg-red-100"
            onClick={() => refetch()}
          >
            Coba Lagi
          </button>
        </div>
      ) : models.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-slate-300 bg-white py-16 text-center">
          <Box className="h-10 w-10 text-slate-300" />
          <p className="mt-3 text-sm font-medium text-slate-500">Belum ada model BIM terdaftar</p>
          <p className="mt-1 text-xs text-slate-400">
            Klik &quot;Daftarkan Model&quot; untuk menambahkan referensi model BIM pertama.
          </p>
          <button
            className="mt-4 flex items-center gap-1.5 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
            onClick={() => setShowRegister(true)}
          >
            <Plus className="h-4 w-4" />
            Daftarkan Model
          </button>
        </div>
      ) : (
        <div className="space-y-3">
          {models.map((model) => (
            <ModelRow key={model.id} model={model} />
          ))}
        </div>
      )}

      {/* Modal */}
      {showRegister && <RegisterModelModal onClose={() => setShowRegister(false)} />}
    </div>
    </DashboardLayout>
  );
}
