'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Pencil, Trash2, ToggleLeft, ToggleRight } from 'lucide-react';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { regionService } from '@/services/spatial.service';
import { cn } from '@/lib/utils';
import type { Region, CreateRegionRequest, UpdateRegionRequest } from '@/types/spatial';

const inputCls = 'h-9 w-full rounded-md border border-input bg-background px-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring';
const textareaCls = 'w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring';
const selectCls = 'h-9 w-full rounded-md border border-input bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring';

const LEVEL_LABELS: Record<number, string> = { 1: 'Nasional', 2: 'Provinsi', 3: 'Kabupaten/Kota', 4: 'Kecamatan' };
const EMPTY: CreateRegionRequest & { is_active?: boolean } = { code: '', name: '', level: 1, parent_id: undefined, description: '', sort_order: 0 };

export default function RegionsPage() {
  const qc = useQueryClient();
  const [showModal, setShowModal] = useState(false);
  const [editing, setEditing] = useState<Region | null>(null);
  const [form, setForm] = useState<typeof EMPTY>({ ...EMPTY });
  const [includeInactive, setIncludeInactive] = useState(false);

  const { data: regions = [], isLoading } = useQuery({ queryKey: ['regions', includeInactive], queryFn: () => regionService.list(includeInactive) });

  const createMut = useMutation({ mutationFn: (p: CreateRegionRequest) => regionService.create(p), onSuccess: () => { qc.invalidateQueries({ queryKey: ['regions'] }); closeModal(); } });
  const updateMut = useMutation({ mutationFn: ({ id, p }: { id: string; p: UpdateRegionRequest }) => regionService.update(id, p), onSuccess: () => { qc.invalidateQueries({ queryKey: ['regions'] }); closeModal(); } });
  const deleteMut = useMutation({ mutationFn: (id: string) => regionService.delete(id), onSuccess: () => qc.invalidateQueries({ queryKey: ['regions'] }) });

  function openCreate() { setEditing(null); setForm({ ...EMPTY }); setShowModal(true); }
  function openEdit(r: Region) { setEditing(r); setForm({ code: r.code, name: r.name, level: r.level, parent_id: r.parent_id, description: r.description ?? '', sort_order: r.sort_order, is_active: r.is_active }); setShowModal(true); }
  function closeModal() { setShowModal(false); setEditing(null); setForm({ ...EMPTY }); }
  function submit() { editing ? updateMut.mutate({ id: editing.id, p: form }) : createMut.mutate(form); }
  function toggle(r: Region) { updateMut.mutate({ id: r.id, p: { is_active: !r.is_active } }); }

  const isPending = createMut.isPending || updateMut.isPending;
  const parentOptions = regions.filter(r => r.level < (form.level ?? 1));

  function indent(level: number) { return level > 1 ? { paddingLeft: `${(level - 1) * 16}px` } : undefined; }

  return (
    <DashboardLayout title="Master Wilayah">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold text-foreground">Wilayah</h1>
            <p className="text-sm text-muted-foreground">Kelola hierarki wilayah administratif untuk lokasi proyek.</p>
          </div>
          <div className="flex items-center gap-3">
            <label className="flex items-center gap-2 text-sm cursor-pointer select-none text-foreground">
              <input type="checkbox" checked={includeInactive} onChange={e => setIncludeInactive(e.target.checked)} className="rounded" />
              Tampilkan non-aktif
            </label>
            <button onClick={openCreate} className="inline-flex h-9 items-center gap-1.5 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary/90">
              <Plus className="h-4 w-4" /> Tambah Wilayah
            </button>
          </div>
        </div>
        <div className="rounded-md border border-border bg-card">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border bg-muted/50">
                <th className="px-4 py-3 text-left font-medium text-foreground">Kode</th>
                <th className="px-4 py-3 text-left font-medium text-foreground">Nama</th>
                <th className="px-4 py-3 text-left font-medium text-foreground">Level</th>
                <th className="px-4 py-3 text-left font-medium text-foreground">Status</th>
                <th className="px-4 py-3 text-right font-medium text-foreground">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (<tr><td colSpan={5} className="px-4 py-8 text-center text-muted-foreground">Memuat...</td></tr>)
              : regions.length === 0 ? (<tr><td colSpan={5} className="px-4 py-8 text-center text-muted-foreground">Belum ada wilayah.</td></tr>)
              : regions.map(r => (
                <tr key={r.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                  <td className="px-4 py-3 font-mono text-xs">{r.code}</td>
                  <td className="px-4 py-3 font-medium text-foreground"><span style={indent(r.level)}>{r.name}</span></td>
                  <td className="px-4 py-3 text-muted-foreground">{LEVEL_LABELS[r.level] ?? `Level ${r.level}`}</td>
                  <td className="px-4 py-3"><span className={cn('inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium', r.is_active ? 'bg-green-100 text-green-700' : 'bg-muted text-muted-foreground')}>{r.is_active ? 'Aktif' : 'Non-aktif'}</span></td>
                  <td className="px-4 py-3"><div className="flex justify-end gap-1">
                    <button onClick={() => toggle(r)} className="rounded-md p-1.5 hover:bg-accent">{r.is_active ? <ToggleRight className="h-4 w-4 text-green-600" /> : <ToggleLeft className="h-4 w-4 text-muted-foreground" />}</button>
                    <button onClick={() => openEdit(r)} className="rounded-md p-1.5 text-muted-foreground hover:bg-accent"><Pencil className="h-4 w-4" /></button>
                    <button onClick={() => { if (confirm(`Hapus wilayah "${r.name}"?`)) deleteMut.mutate(r.id); }} className="rounded-md p-1.5 text-destructive hover:bg-destructive/10"><Trash2 className="h-4 w-4" /></button>
                  </div></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg bg-card border border-border shadow-xl p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">{editing ? 'Edit Wilayah' : 'Tambah Wilayah'}</h2>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <label className="block"><span className="mb-1 block text-sm font-medium text-foreground">Kode <span className="text-destructive">*</span></span><input className={inputCls} value={form.code} onChange={e => setForm(f => ({ ...f, code: e.target.value }))} placeholder="JB" /></label>
                <label className="block"><span className="mb-1 block text-sm font-medium text-foreground">Level <span className="text-destructive">*</span></span>
                  <select className={selectCls} value={form.level} onChange={e => setForm(f => ({ ...f, level: Number(e.target.value), parent_id: undefined }))}>
                    {Object.entries(LEVEL_LABELS).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
                  </select>
                </label>
              </div>
              <label className="block"><span className="mb-1 block text-sm font-medium text-foreground">Nama <span className="text-destructive">*</span></span><input className={inputCls} value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="Nama wilayah" /></label>
              {parentOptions.length > 0 && (
                <label className="block"><span className="mb-1 block text-sm font-medium text-foreground">Induk</span>
                  <select className={selectCls} value={form.parent_id ?? ''} onChange={e => setForm(f => ({ ...f, parent_id: e.target.value || undefined }))}>
                    <option value="">— Tanpa induk —</option>
                    {parentOptions.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                  </select>
                </label>
              )}
              <label className="block"><span className="mb-1 block text-sm font-medium text-foreground">Deskripsi</span><textarea className={textareaCls} rows={2} value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} /></label>
              <label className="block"><span className="mb-1 block text-sm font-medium text-foreground">Urutan</span><input className={inputCls} type="number" value={form.sort_order} onChange={e => setForm(f => ({ ...f, sort_order: Number(e.target.value) }))} /></label>
            </div>
            <div className="mt-6 flex justify-end gap-2">
              <button onClick={closeModal} className="inline-flex h-9 items-center rounded-md border border-input px-3 text-sm font-medium hover:bg-accent">Batal</button>
              <button onClick={submit} disabled={!form.code || !form.name || isPending} className="inline-flex h-9 items-center rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed">{isPending ? 'Menyimpan...' : editing ? 'Simpan' : 'Tambah'}</button>
            </div>
          </div>
        </div>
      )}
    </DashboardLayout>
  );
}
