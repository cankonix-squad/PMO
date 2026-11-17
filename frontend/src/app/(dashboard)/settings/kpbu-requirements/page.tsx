"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Pencil, Plus } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { kpbuErrorMessage, kpbuService } from "@/services/kpbu.service";
import { btnPrimary, btnSecondary, Field, inputCls, Modal, textareaCls } from "@/features/kpbu/ui";
import { cn } from "@/lib/utils";
import {
  KPBU_APPLIES_LABEL,
  KPBU_DOC_CATEGORY_LABEL,
  KPBU_MAIN_STAGES,
  KPBU_SOURCE_LABEL,
  KPBU_STAGE_LABEL,
  type KpbuAppliesTo,
  type KpbuDocCategory,
  type KpbuDocumentRequirement,
  type KpbuSourceStatus,
  type KpbuStage,
} from "@/types/kpbu";

const GATE_STAGES: KpbuStage[] = [...KPBU_MAIN_STAGES, "CLOSED"];

export default function KpbuRequirementsPage() {
  const [editing, setEditing] = useState<KpbuDocumentRequirement | "new" | null>(null);
  const { data = [], isLoading, isError, refetch } = useQuery({
    queryKey: ["kpbu", "requirements", "all"],
    queryFn: () => kpbuService.listRequirements(true).then((r) => r.data),
  });

  return (
    <DashboardLayout title="Checklist Dokumen KPBU">
      <div className="space-y-4">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <p className="max-w-3xl text-sm text-gray-600">
            Dokumen yang wajib dilengkapi badan usaha per tahap KPBU. Dokumen wajib dengan &ldquo;syarat masuk tahap&rdquo; harus disetujui
            sebelum proyek dapat dipindah ke tahap tersebut. Baris &ldquo;Belum terverifikasi&rdquo; perlu dicek terhadap Peraturan LKPP 1/2025.
          </p>
          <button className={btnPrimary} onClick={() => setEditing("new")}><Plus className="h-4 w-4" /> Tambah</button>
        </div>

        <div className="overflow-x-auto rounded-xl border border-gray-100 bg-white shadow-sm">
          {isLoading ? (
            <p className="p-6 text-center text-sm text-gray-400">Memuat...</p>
          ) : isError ? (
            <p className="p-6 text-center text-sm text-red-600">Gagal memuat. <button className="underline" onClick={() => refetch()}>Coba lagi</button></p>
          ) : data.length === 0 ? (
            <p className="p-6 text-center text-sm text-gray-400">Belum ada checklist. Jalankan seed untuk memuat matriks bawaan.</p>
          ) : (
            <table className="w-full min-w-[860px] text-sm">
              <thead className="border-b border-gray-100 text-left text-xs uppercase text-gray-500">
                <tr>
                  <th className="px-3 py-2">Kode</th>
                  <th className="px-3 py-2">Dokumen</th>
                  <th className="px-3 py-2">Berlaku</th>
                  <th className="px-3 py-2">Syarat masuk tahap</th>
                  <th className="px-3 py-2">Dasar / Sumber</th>
                  <th className="px-3 py-2" />
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {data.map((r) => (
                  <tr key={r.id} className={cn(!r.is_active && "opacity-50")}>
                    <td className="px-3 py-2 font-semibold text-gray-500">{r.code}</td>
                    <td className="px-3 py-2">
                      <p className="text-gray-800">{r.name}</p>
                      <p className="text-xs text-gray-400">
                        {KPBU_DOC_CATEGORY_LABEL[r.category]} · {r.is_mandatory ? "Wajib" : "Opsional"}{!r.is_active && " · Nonaktif"}
                      </p>
                    </td>
                    <td className="px-3 py-2 text-xs text-gray-600">{KPBU_APPLIES_LABEL[r.applies_to]}</td>
                    <td className="px-3 py-2 text-xs text-gray-600">{r.required_before_stage ? KPBU_STAGE_LABEL[r.required_before_stage] : "—"}</td>
                    <td className="px-3 py-2 text-xs">
                      <p className="text-gray-600">{r.legal_basis || "—"}</p>
                      <p className={cn(r.source_status === "VERIFIED" ? "text-green-700" : r.source_status === "UNVERIFIED" ? "text-orange-700" : "text-sky-700")}>
                        {KPBU_SOURCE_LABEL[r.source_status]}
                      </p>
                    </td>
                    <td className="px-3 py-2 text-right">
                      <button className="rounded p-1 text-gray-400 hover:bg-gray-100" onClick={() => setEditing(r)} aria-label={`Edit ${r.code}`}><Pencil className="h-3.5 w-3.5" /></button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
      {editing && <RequirementModal req={editing === "new" ? null : editing} onClose={() => setEditing(null)} />}
    </DashboardLayout>
  );
}

function RequirementModal({ req, onClose }: { req: KpbuDocumentRequirement | null; onClose: () => void }) {
  const qc = useQueryClient();
  const [v, setV] = useState({
    code: req?.code ?? "",
    category: req?.category ?? "PRAKUALIFIKASI",
    name: req?.name ?? "",
    description: req?.description ?? "",
    applies_to: req?.applies_to ?? "ALL",
    required_before_stage: req?.required_before_stage ?? "",
    is_mandatory: req?.is_mandatory ?? true,
    legal_basis: req?.legal_basis ?? "",
    source_status: req?.source_status ?? "UNVERIFIED",
    sort_order: String(req?.sort_order ?? 0),
    is_active: req?.is_active ?? true,
  });
  const [error, setError] = useState<string | null>(null);
  const set = (k: keyof typeof v) => (e: { target: { value: string } }) => setV((s) => ({ ...s, [k]: e.target.value }));

  const invalidate = () => qc.invalidateQueries({ queryKey: ["kpbu"] });
  const save = useMutation({
    mutationFn: async (): Promise<void> => {
      const payload = {
        code: v.code.trim(),
        category: v.category as KpbuDocCategory,
        name: v.name.trim(),
        description: v.description,
        applies_to: v.applies_to as KpbuAppliesTo,
        required_before_stage: (v.required_before_stage || null) as KpbuStage | null,
        is_mandatory: v.is_mandatory,
        legal_basis: v.legal_basis,
        source_status: v.source_status as KpbuSourceStatus,
        sort_order: Number(v.sort_order) || 0,
        is_active: v.is_active,
      };
      if (req) await kpbuService.updateRequirement(req.id, payload);
      else await kpbuService.createRequirement(payload);
    },
    onSuccess: () => {
      invalidate();
      onClose();
    },
    onError: (e) => setError(kpbuErrorMessage(e, "Gagal menyimpan.")),
  });

  return (
    <Modal title={req ? `Edit ${req.code}` : "Tambah Dokumen Checklist"} onClose={onClose} wide>
      <form
        className="grid grid-cols-1 gap-3 sm:grid-cols-2"
        onSubmit={(e) => {
          e.preventDefault();
          if (!v.code.trim() || !v.name.trim()) {
            setError("Kode dan nama wajib diisi.");
            return;
          }
          save.mutate();
        }}
      >
        <Field label="Kode" required><input className={inputCls} value={v.code} onChange={set("code")} maxLength={30} /></Field>
        <Field label="Kategori" required>
          <select className={inputCls} value={v.category} onChange={set("category")}>
            {Object.entries(KPBU_DOC_CATEGORY_LABEL).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
          </select>
        </Field>
        <Field label="Nama dokumen" required className="sm:col-span-2"><input className={inputCls} value={v.name} onChange={set("name")} maxLength={300} /></Field>
        <Field label="Keterangan" className="sm:col-span-2"><textarea className={textareaCls} rows={2} value={v.description} onChange={set("description")} /></Field>
        <Field label="Berlaku untuk">
          <select className={inputCls} value={v.applies_to} onChange={set("applies_to")}>
            {Object.entries(KPBU_APPLIES_LABEL).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
          </select>
        </Field>
        <Field label="Syarat masuk tahap">
          <select className={inputCls} value={v.required_before_stage} onChange={set("required_before_stage")}>
            <option value="">— Tidak menahan tahap —</option>
            {GATE_STAGES.map((s) => <option key={s} value={s}>{KPBU_STAGE_LABEL[s]}</option>)}
          </select>
        </Field>
        <Field label="Dasar hukum" className="sm:col-span-2"><input className={inputCls} value={v.legal_basis} onChange={set("legal_basis")} maxLength={500} /></Field>
        <Field label="Status sumber">
          <select className={inputCls} value={v.source_status} onChange={set("source_status")}>
            {Object.entries(KPBU_SOURCE_LABEL).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
          </select>
        </Field>
        <Field label="Urutan"><input className={inputCls} type="number" value={v.sort_order} onChange={set("sort_order")} /></Field>
        <label className="flex items-center gap-2 text-sm text-gray-700">
          <input type="checkbox" checked={v.is_mandatory} onChange={(e) => setV((s) => ({ ...s, is_mandatory: e.target.checked }))} /> Wajib
        </label>
        <label className="flex items-center gap-2 text-sm text-gray-700">
          <input type="checkbox" checked={v.is_active} onChange={(e) => setV((s) => ({ ...s, is_active: e.target.checked }))} /> Aktif
        </label>
        {error && <p className="text-sm text-red-600 sm:col-span-2">{error}</p>}
        <div className="flex justify-end gap-2 sm:col-span-2">
          <button type="button" className={btnSecondary} onClick={onClose}>Batal</button>
          <button type="submit" className={btnPrimary} disabled={save.isPending}>{save.isPending ? "Menyimpan..." : "Simpan"}</button>
        </div>
      </form>
    </Modal>
  );
}
