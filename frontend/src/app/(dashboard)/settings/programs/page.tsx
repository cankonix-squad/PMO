'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Pencil, Trash2, ToggleLeft, ToggleRight } from 'lucide-react';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { programService } from '@/services/portfolio.service';
import { cn } from '@/lib/utils';
import type { Program, CreateProgramRequest, UpdateProgramRequest } from '@/types/portfolio';

const inputCls = 'h-9 w-full rounded-md border border-input bg-background px-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring';
const textareaCls = 'w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring';

const EMPTY: CreateProgramRequest & { is_active?: boolean } = { code: '', name: '', description: '', fiscal_year: undefined, sort_order: 0 };

export default function ProgramsPage() {
  const qc = useQueryClient();
  const [showModal, setShowModal] = useState(false);
  const [editing, setEditing] = useState<Program | null>(null);
  const [form, setForm] = useState<typeof EMPTY>({ ...EMPTY });
  const [includeInactive, setIncludeInactive] = useState(false);

  const { data: programs = [], isLoading } = useQuery({ queryKey: ['programs', includeInactive], queryFn: () => programService.list(includeInactive) });

  const createMut = useMutation({ mutationFn: (p: CreateProgramRequest) => programService.create(p), onSuccess: () => { qc.invalidateQueries({ queryKey: ['programs'] }); closeModal(); } });
  const updateMut = useMutation({ mutationFn: ({ id, p }: { id: string; p: UpdateProgramRequest }) => programService.update(id, p), onSuccess: () => { qc.invalidateQueries({ queryKey: ['programs'] }); closeModal(); } });
  const deleteMut = useMutation({ mutationFn: (id: string) => programService.delete(id), onSuccess: () => qc.invalidateQueries({ queryKey: ['programs'] }) });

  function openCreate() { setEditing(null); setForm({ ...EMPTY }); setShowModal(true); }
  function openEdit(p: Program) { setEditing(p); setForm({ code: p.code, name: p.name, description: p.description ?? '', fiscal_year: p.fiscal_year, sort_order: p.sort_order, is_active: p.is_active }); setShowModal(true); }
  function closeModal() { setShowModal(false); setEditing(null); setForm({ ...EMPTY }); }
  function submit() { editing ? updateMut.mutate({ id: editing.id, p: form }) : createMut.mutate(form); }
  function toggle(p: Program) { updateMut.mutate({ id: p.id, p: { is_active: !p.is_active } }); }

  const isPending = createMut.isPending || updateMut.isPending;

  return (
    <DashboardLayout title="Master Program">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold text-foreground">Program</h1>
            <p className="text-sm text-muted-foreground">Kelola master program sebagai payung pengelompokan proyek.</p>
          </div>
          <div className="flex items-center gap-3">
            <label className="flex items-center gap-2 text-sm cursor-pointer select-none text-foreground">
              <input type="checkbox" checked={includeInactive} onChange={e => setIncludeInactive(e.target.checked)} className="rounded" />
              Tampilkan non-aktif
            </label>
            <button onClick={openCreate} className="inline-flex h-9 items-center gap-1.5 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary/90">
              <Plus className="h-4 w-4" /> Tambah Program
            </button>
          </div>
        </div>

        <div className="rounded-md border border-border bg-card">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border bg-muted/50">
                <th className="px-4 py-3 text-left font-medium text-foreground">Kode</th>
                <th className="px-4 py-3 text-left font-medium text-foreground">Nama</th>
                <th className="px-4 py-3 text-left font-medium text-foreground">Tahun Fiskal</th>
                <th className="px-4 py-3 text-left font-medium text-foreground">Status</th>
                <th className="px-4 py-3 text-right font-medium text-foreground">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr><td colSpan={5} className="px-4 py-8 text-center text-muted-foreground">Memuat...</td></tr>
              ) : programs.length === 0 ? (
                <tr><td colSpan={5} className="px-4 py-8 text-center text-muted-foreground">Belum ada program.</td></tr>
              ) : programs.map(p => (
                <tr key={p.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                  <td className="px-4 py-3 font-mono text-xs text-foreground">{p.code}</td>
                  <td className="px-4 py-3"><div className="font-medium text-foreground">{p.name}</div>{p.description && <div className="text-xs text-muted-foreground line-clamp-1">{p.description}</div>}</td>
                  <td className="px-4 py-3 text-muted-foreground">{p.fiscal_year ?? '—'}</td>
                  <td className="px-4 py-3"><span className={cn('inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium', p.is_active ? 'bg-green-100 text-green-700' : 'bg-muted text-muted-foreground')}>{p.is_active ? 'Aktif' : 'Non-aktif'}</span></td>
                  <td className="px-4 py-3"><div className="flex justify-end gap-1">
                    <button onClick={() => toggle(p)} className="rounded-md p-1.5 text-muted-foreground hover:bg-accent">{p.is_active ? <ToggleRight className="h-4 w-4 text-green-600" /> : <ToggleLeft className="h-4 w-4" />}</button>
                    <button onClick={() => openEdit(p)} className="rounded-md p-1.5 text-muted-foreground hover:bg-accent"><Pencil className="h-4 w-4" /></button>
                    <button onClick={() => { if (confirm(`Hapus program "${p.name}"?`)) deleteMut.mutate(p.id); }} className="rounded-md p-1.5 text-destructive hover:bg-destructive/10"><Trash2 className="h-4 w-4" /></button>
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
            <h2 className="text-lg font-semibold text-foreground mb-4">{editing ? 'Edit Program' : 'Tambah Program'}</h2>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <label className="block"><span className="mb-1 block text-sm font-medium text-foreground">Kode <span className="text-destructive">*</span></span><input className={inputCls} value={form.code} onChange={e => setForm(f => ({ ...f, code: e.target.value }))} placeholder="PRG-001" /></label>
                <label className="block"><span className="mb-1 block text-sm font-medium text-foreground">Tahun Fiskal</span><input className={inputCls} type="number" value={form.fiscal_year ?? ''} onChange={e => setForm(f => ({ ...f, fiscal_year: e.target.value ? Number(e.target.value) : undefined }))} placeholder="2025" /></label>
              </div>
              <label className="block"><span className="mb-1 block text-sm font-medium text-foreground">Nama <span className="text-destructive">*</span></span><input className={inputCls} value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="Nama program" /></label>
              <label className="block"><span className="mb-1 block text-sm font-medium text-foreground">Deskripsi</span><textarea className={textareaCls} rows={3} value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} /></label>
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
