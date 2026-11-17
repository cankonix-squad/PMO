"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { projectService } from "@/services/project.service";
import { Pencil, Plus, Trash2, Users } from "lucide-react";
import { kpbuErrorMessage, kpbuService } from "@/services/kpbu.service";
import { KPBU_PARTY_ROLE_LABEL, type KpbuParty, type KpbuPartyRole } from "@/types/kpbu";
import { btnPrimary, btnSecondary, Field, inputCls, Modal, textareaCls } from "./ui";

export function PartiesPanel({ projectId, parties, canWrite }: { projectId: string; parties: KpbuParty[]; canWrite: boolean }) {
  const qc = useQueryClient();
  const [editing, setEditing] = useState<KpbuParty | "new" | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null);

  const del = useMutation({
    mutationFn: (id: string) => kpbuService.deleteParty(projectId, id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["kpbu"] });
      setConfirmDelete(null);
    },
  });

  return (
    <div className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-gray-700">Para Pihak</h2>
        {canWrite && (
          <button className={btnSecondary} onClick={() => setEditing("new")}>
            <Plus className="h-4 w-4" /> Tambah Pihak
          </button>
        )}
      </div>
      {parties.length === 0 ? (
        <div className="py-6 text-center">
          <Users className="mx-auto mb-2 h-7 w-7 text-gray-300" />
          <p className="text-sm text-gray-500">Belum ada pihak tercatat (PJPK, BUP, sponsor, lender, dll.).</p>
        </div>
      ) : (
        <ul className="divide-y divide-gray-50">
          {parties.map((p) => (
            <li key={p.id} className="flex flex-wrap items-start justify-between gap-2 py-2.5">
              <div className="min-w-0">
                <p className="text-sm font-medium text-gray-800">
                  {p.name}
                  {p.share_pct != null && <span className="ml-2 text-xs font-normal text-gray-500">{p.share_pct}%</span>}
                </p>
                <p className="text-xs text-blue-700">{KPBU_PARTY_ROLE_LABEL[p.role] ?? p.role}</p>
                {(p.contact_person || p.email || p.phone) && (
                  <p className="text-xs text-gray-500">{[p.contact_person, p.email, p.phone].filter(Boolean).join(" · ")}</p>
                )}
              </div>
              {canWrite && (
                <div className="flex items-center gap-1">
                  {confirmDelete === p.id ? (
                    <>
                      <span className="text-xs text-gray-600">Hapus?</span>
                      <button className="rounded px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50" onClick={() => del.mutate(p.id)} disabled={del.isPending}>Ya</button>
                      <button className="rounded px-2 py-1 text-xs text-gray-600 hover:bg-gray-100" onClick={() => setConfirmDelete(null)}>Batal</button>
                    </>
                  ) : (
                    <>
                      <button className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-700" onClick={() => setEditing(p)} aria-label={`Edit ${p.name}`}><Pencil className="h-3.5 w-3.5" /></button>
                      <button className="rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-600" onClick={() => setConfirmDelete(p.id)} aria-label={`Hapus ${p.name}`}><Trash2 className="h-3.5 w-3.5" /></button>
                    </>
                  )}
                </div>
              )}
            </li>
          ))}
        </ul>
      )}
      {editing && <PartyModal projectId={projectId} party={editing === "new" ? undefined : editing} onClose={() => setEditing(null)} />}
    </div>
  );
}

function PartyModal({ projectId, party, onClose }: { projectId: string; party?: KpbuParty; onClose: () => void }) {
  const qc = useQueryClient();
  const [v, setV] = useState({
    role: (party?.role ?? "BUP") as KpbuPartyRole,
    name: party?.name ?? "",
    share_pct: party?.share_pct != null ? String(party.share_pct) : "",
    contact_person: party?.contact_person ?? "",
    email: party?.email ?? "",
    phone: party?.phone ?? "",
    notes: party?.notes ?? "",
    vendor_id: party?.vendor_id ?? "",
  });
  const vendors = useQuery({
    queryKey: ["vendors", "kpbu-party"],
    queryFn: () => projectService.listVendors({ is_active: true }).then((r) => r.data.data ?? []),
  });
  const [error, setError] = useState<string | null>(null);
  const set = (k: keyof typeof v) => (e: { target: { value: string } }) => setV((s) => ({ ...s, [k]: e.target.value }));

  const save = useMutation({
    mutationFn: () => {
      const payload = { ...v, share_pct: v.share_pct === "" ? null : Number(v.share_pct), vendor_id: v.vendor_id || null };
      return party ? kpbuService.updateParty(projectId, party.id, payload) : kpbuService.createParty(projectId, payload);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["kpbu"] });
      onClose();
    },
    onError: (err) => setError(kpbuErrorMessage(err, "Gagal menyimpan pihak.")),
  });

  const submit = () => {
    if (!v.name.trim()) return setError("Nama wajib diisi.");
    const share = Number(v.share_pct);
    if (v.share_pct !== "" && (!Number.isFinite(share) || share < 0 || share > 100)) return setError("Porsi kepemilikan 0–100%.");
    setError(null);
    save.mutate();
  };

  return (
    <Modal title={party ? "Edit Pihak" : "Tambah Pihak"} onClose={onClose}>
      <form onSubmit={(e) => { e.preventDefault(); submit(); }} className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <Field label="Peran" required>
          <select className={inputCls} value={v.role} onChange={set("role")}>
            {Object.entries(KPBU_PARTY_ROLE_LABEL).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
          </select>
        </Field>
        <Field label="Porsi Kepemilikan (%)">
          <input className={inputCls} type="number" min={0} max={100} step="any" value={v.share_pct} onChange={set("share_pct")} placeholder="untuk BUP/sponsor" />
        </Field>
        <Field label="Nama" required className="sm:col-span-2">
          <input className={inputCls} value={v.name} onChange={set("name")} maxLength={300} />
        </Field>
        <Field label="Data penyedia (untuk akses portal BUP)" className="sm:col-span-2">
          <select className={inputCls} value={v.vendor_id} onChange={set("vendor_id")}>
            <option value="">— Tidak terhubung —</option>
            {(vendors.data ?? []).map((vd) => <option key={vd.id} value={vd.id}>{vd.name}</option>)}
          </select>
          <span className="text-xs text-gray-400">Bila peran BUP dan terhubung ke data penyedia, akun penyedia tersebut dapat membuka proyek ini di portal.</span>
        </Field>
        <Field label="Kontak"><input className={inputCls} value={v.contact_person} onChange={set("contact_person")} maxLength={200} /></Field>
        <Field label="Telepon"><input className={inputCls} value={v.phone} onChange={set("phone")} maxLength={50} /></Field>
        <Field label="Email" className="sm:col-span-2"><input className={inputCls} type="email" value={v.email} onChange={set("email")} maxLength={200} /></Field>
        <Field label="Catatan" className="sm:col-span-2"><textarea className={textareaCls} rows={2} value={v.notes} onChange={set("notes")} /></Field>
        {error && <p className="text-sm text-red-600 sm:col-span-2">{error}</p>}
        <div className="flex justify-end gap-2 sm:col-span-2">
          <button type="button" className={btnSecondary} onClick={onClose}>Batal</button>
          <button type="submit" className={btnPrimary} disabled={save.isPending}>{save.isPending ? "Menyimpan..." : "Simpan"}</button>
        </div>
      </form>
    </Modal>
  );
}
