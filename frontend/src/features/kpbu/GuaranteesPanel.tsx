"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Pencil, Plus, Trash2 } from "lucide-react";
import { kpbuErrorMessage, kpbuService } from "@/services/kpbu.service";
import { cn } from "@/lib/utils";
import {
  KPBU_GUARANTEE_STATUS_LABEL,
  KPBU_GUARANTEE_TYPE_LABEL,
  type KpbuGuarantee,
  type KpbuGuaranteeStatus,
  type KpbuGuaranteeType,
} from "@/types/kpbu";
import { btnPrimary, btnSecondary, Field, fmtDate, fmtRupiahShort, inputCls, Modal, textareaCls, toDateInput } from "./ui";

export function GuaranteesPanel({ projectId, canWrite, canDelete }: { projectId: string; canWrite: boolean; canDelete: boolean }) {
  const qc = useQueryClient();
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["kpbu", "guarantees", projectId],
    queryFn: () => kpbuService.listGuarantees(projectId).then((r) => r.data),
  });
  const [editing, setEditing] = useState<KpbuGuarantee | "new" | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const del = useMutation({
    mutationFn: (gid: string) => kpbuService.deleteGuarantee(projectId, gid),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["kpbu", "guarantees", projectId] }),
    onError: (e) => setErr(kpbuErrorMessage(e, "Gagal menghapus jaminan.")),
  });

  return (
    <div className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-gray-700">Jaminan</h2>
        {canWrite && <button className={btnSecondary} onClick={() => setEditing("new")}><Plus className="h-4 w-4" /> Tambah</button>}
      </div>
      {isLoading ? (
        <p className="py-4 text-center text-sm text-gray-400">Memuat jaminan...</p>
      ) : isError ? (
        <p className="py-4 text-center text-sm text-red-600">Gagal memuat. <button className="underline" onClick={() => refetch()}>Coba lagi</button></p>
      ) : !data?.length ? (
        <p className="py-4 text-center text-sm text-gray-400">Belum ada jaminan penawaran, pelaksanaan, atau penjaminan pemerintah.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full min-w-[640px] text-sm">
            <thead className="border-b border-gray-100 text-left text-xs uppercase text-gray-500">
              <tr>
                <th className="py-2 pr-3">Jenis</th>
                <th className="py-2 pr-3">Penerbit / Nomor</th>
                <th className="py-2 pr-3">Nilai</th>
                <th className="py-2 pr-3">Berlaku s.d.</th>
                <th className="py-2 pr-3">Status</th>
                {(canWrite || canDelete) && <th className="py-2" />}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {data.map((g) => (
                <tr key={g.id}>
                  <td className="py-2 pr-3 font-medium text-gray-800">{KPBU_GUARANTEE_TYPE_LABEL[g.type]}</td>
                  <td className="py-2 pr-3 text-xs text-gray-600">{g.issuer || "—"}<br /><span className="text-gray-400">{g.guarantee_number || "—"}</span></td>
                  <td className="py-2 pr-3 text-xs">
                    {fmtRupiahShort(g.amount)}
                    {g.pct_of_investment != null && (
                      <span className={cn("block", g.exceeds_limit ? "font-medium text-red-600" : "text-gray-400")}>
                        {g.pct_of_investment.toLocaleString("id-ID")}% nilai investasi{g.exceeds_limit && " — melebihi batas 5%"}
                      </span>
                    )}
                  </td>
                  <td className="py-2 pr-3 text-xs">
                    {fmtDate(g.expiry_date)}
                    {g.status === "ACTIVE" && g.days_to_expiry != null && g.days_to_expiry <= 30 && (
                      <span className="block font-medium text-red-600">{g.days_to_expiry < 0 ? "Sudah lewat" : `${g.days_to_expiry} hari lagi`}</span>
                    )}
                  </td>
                  <td className="py-2 pr-3 text-xs">{KPBU_GUARANTEE_STATUS_LABEL[g.status]}</td>
                  {(canWrite || canDelete) && (
                    <td className="py-2 text-right">
                      {canWrite && <button className="rounded p-1 text-gray-400 hover:bg-gray-100" onClick={() => setEditing(g)} aria-label="Edit"><Pencil className="h-3.5 w-3.5" /></button>}
                      {canDelete && <button className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-red-600" onClick={() => del.mutate(g.id)} aria-label="Hapus"><Trash2 className="h-3.5 w-3.5" /></button>}
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {err && <p className="mt-2 text-sm text-red-600">{err}</p>}
      <p className="mt-3 text-xs text-gray-400">Jaminan pelaksanaan paling banyak 5% dari nilai investasi (Perpres 38/2015).</p>
      {editing && <GuaranteeModal projectId={projectId} guarantee={editing === "new" ? null : editing} onClose={() => setEditing(null)} />}
    </div>
  );
}

function GuaranteeModal({ projectId, guarantee, onClose }: { projectId: string; guarantee: KpbuGuarantee | null; onClose: () => void }) {
  const qc = useQueryClient();
  const [v, setV] = useState({
    type: guarantee?.type ?? "PERFORMANCE",
    issuer: guarantee?.issuer ?? "",
    guarantee_number: guarantee?.guarantee_number ?? "",
    amount: guarantee ? String(guarantee.amount) : "",
    issued_date: toDateInput(guarantee?.issued_date),
    expiry_date: toDateInput(guarantee?.expiry_date),
    status: guarantee?.status ?? "ACTIVE",
    notes: guarantee?.notes ?? "",
  });
  const [error, setError] = useState<string | null>(null);
  const set = (k: keyof typeof v) => (e: { target: { value: string } }) => setV((s) => ({ ...s, [k]: e.target.value }));

  const save = useMutation({
    mutationFn: async (): Promise<void> => {
      const payload = {
        type: v.type as KpbuGuaranteeType,
        issuer: v.issuer,
        guarantee_number: v.guarantee_number,
        amount: Number(v.amount),
        currency: "IDR",
        issued_date: v.issued_date || null,
        expiry_date: v.expiry_date || null,
        status: v.status as KpbuGuaranteeStatus,
        notes: v.notes,
      };
      if (guarantee) await kpbuService.updateGuarantee(projectId, guarantee.id, payload);
      else await kpbuService.createGuarantee(projectId, payload);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["kpbu", "guarantees", projectId] });
      onClose();
    },
    onError: (e) => setError(kpbuErrorMessage(e, "Gagal menyimpan jaminan.")),
  });

  return (
    <Modal title={guarantee ? "Edit Jaminan" : "Tambah Jaminan"} onClose={onClose}>
      <form
        className="grid grid-cols-2 gap-3"
        onSubmit={(e) => {
          e.preventDefault();
          if (v.amount === "" || !Number.isFinite(Number(v.amount)) || Number(v.amount) < 0) {
            setError("Nilai jaminan wajib dan tidak boleh negatif.");
            return;
          }
          if (v.issued_date && v.expiry_date && v.expiry_date < v.issued_date) {
            setError("Tanggal berakhir tidak boleh sebelum tanggal terbit.");
            return;
          }
          save.mutate();
        }}
      >
        <Field label="Jenis" required>
          <select className={inputCls} value={v.type} onChange={set("type")}>
            {Object.entries(KPBU_GUARANTEE_TYPE_LABEL).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
          </select>
        </Field>
        <Field label="Status">
          <select className={inputCls} value={v.status} onChange={set("status")}>
            {Object.entries(KPBU_GUARANTEE_STATUS_LABEL).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
          </select>
        </Field>
        <Field label="Penerbit (bank/asuransi/PT PII)"><input className={inputCls} value={v.issuer} onChange={set("issuer")} maxLength={300} /></Field>
        <Field label="Nomor jaminan"><input className={inputCls} value={v.guarantee_number} onChange={set("guarantee_number")} maxLength={200} /></Field>
        <Field label="Nilai (Rp)" required className="col-span-2"><input className={inputCls} type="number" min={0} step="any" value={v.amount} onChange={set("amount")} /></Field>
        <Field label="Tanggal terbit"><input className={inputCls} type="date" value={v.issued_date} onChange={set("issued_date")} /></Field>
        <Field label="Berlaku sampai"><input className={inputCls} type="date" value={v.expiry_date} onChange={set("expiry_date")} /></Field>
        <Field label="Catatan" className="col-span-2"><textarea className={textareaCls} rows={2} value={v.notes} onChange={set("notes")} /></Field>
        {error && <p className="col-span-2 text-sm text-red-600">{error}</p>}
        <div className="col-span-2 flex justify-end gap-2">
          <button type="button" className={btnSecondary} onClick={onClose}>Batal</button>
          <button type="submit" className={btnPrimary} disabled={save.isPending}>{save.isPending ? "Menyimpan..." : "Simpan"}</button>
        </div>
      </form>
    </Modal>
  );
}
