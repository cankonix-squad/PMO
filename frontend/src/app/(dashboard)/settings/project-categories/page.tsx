'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Pencil, Trash2, ToggleLeft, ToggleRight, Tag } from 'lucide-react';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { projectCategoryService } from '@/services/project-category.service';
import { cn } from '@/lib/utils';
import type {
  ProjectCategory,
  CreateProjectCategoryRequest,
  UpdateProjectCategoryRequest,
} from '@/types/project-category';

const inputCls =
  'h-9 w-full rounded-md border border-input bg-background px-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring';
const textareaCls =
  'w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring';

const EMPTY: CreateProjectCategoryRequest = {
  code: '',
  name: '',
  description: '',
  is_active: true,
  sort_order: 0,
};

export default function ProjectCategoriesPage() {
  const qc = useQueryClient();
  const [showModal, setShowModal] = useState(false);
  const [editing, setEditing] = useState<ProjectCategory | null>(null);
  const [form, setForm] = useState<typeof EMPTY>({ ...EMPTY });
  const [includeInactive, setIncludeInactive] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);

  const { data: categories = [], isLoading } = useQuery({
    queryKey: ['project-categories', includeInactive],
    queryFn: () => projectCategoryService.list(includeInactive).then((r) => r.data ?? []),
  });

  const createMut = useMutation({
    mutationFn: (p: CreateProjectCategoryRequest) => projectCategoryService.create(p),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['project-categories'] });
      closeModal();
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ??
        'Gagal menyimpan kategori.';
      setApiError(msg);
    },
  });

  const updateMut = useMutation({
    mutationFn: ({ id, p }: { id: string; p: UpdateProjectCategoryRequest }) =>
      projectCategoryService.update(id, p),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['project-categories'] });
      closeModal();
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ??
        'Gagal memperbarui kategori.';
      setApiError(msg);
    },
  });

  const deleteMut = useMutation({
    mutationFn: (id: string) => projectCategoryService.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['project-categories'] }),
  });

  function openCreate() {
    setEditing(null);
    setForm({ ...EMPTY });
    setApiError(null);
    setShowModal(true);
  }

  function openEdit(cat: ProjectCategory) {
    setEditing(cat);
    setForm({
      code: cat.code,
      name: cat.name,
      description: cat.description ?? '',
      is_active: cat.is_active,
      sort_order: cat.sort_order,
    });
    setApiError(null);
    setShowModal(true);
  }

  function closeModal() {
    setShowModal(false);
    setEditing(null);
    setForm({ ...EMPTY });
    setApiError(null);
  }

  function submit() {
    setApiError(null);
    if (editing) {
      updateMut.mutate({ id: editing.id, p: form });
    } else {
      createMut.mutate(form);
    }
  }

  function toggle(cat: ProjectCategory) {
    updateMut.mutate({ id: cat.id, p: { is_active: !cat.is_active } });
  }

  const isPending = createMut.isPending || updateMut.isPending;

  return (
    <DashboardLayout title="Kategori Proyek">
      <div className="space-y-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-xl font-semibold text-foreground">Kategori Proyek</h1>
            <p className="text-sm text-muted-foreground">
              Kelola kategori master untuk klasifikasi proyek SDA.
            </p>
          </div>
          <div className="flex items-center gap-3">
            <label className="flex cursor-pointer select-none items-center gap-2 text-sm text-foreground">
              <input
                type="checkbox"
                checked={includeInactive}
                onChange={(e) => setIncludeInactive(e.target.checked)}
                className="rounded"
              />
              Tampilkan non-aktif
            </label>
            <button
              onClick={openCreate}
              className="inline-flex h-9 items-center gap-1.5 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            >
              <Plus className="h-4 w-4" /> Tambah Kategori
            </button>
          </div>
        </div>

        <div className="rounded-md border border-border bg-card">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border bg-muted/50">
                <th className="px-4 py-3 text-left font-medium text-foreground">Kode</th>
                <th className="px-4 py-3 text-left font-medium text-foreground">Nama</th>
                <th className="px-4 py-3 text-left font-medium text-foreground">Deskripsi</th>
                <th className="px-4 py-3 text-center font-medium text-foreground">Urutan</th>
                <th className="px-4 py-3 text-left font-medium text-foreground">Status</th>
                <th className="px-4 py-3 text-right font-medium text-foreground">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-muted-foreground">
                    Memuat...
                  </td>
                </tr>
              ) : categories.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-muted-foreground">
                    Belum ada kategori proyek.
                  </td>
                </tr>
              ) : (
                categories.map((cat) => (
                  <tr
                    key={cat.id}
                    className="border-b border-border last:border-0 hover:bg-muted/30"
                  >
                    <td className="px-4 py-3 font-mono text-xs font-medium text-foreground">
                      {cat.code}
                    </td>
                    <td className="px-4 py-3 font-medium text-foreground">{cat.name}</td>
                    <td className="px-4 py-3 text-muted-foreground">
                      {cat.description || '—'}
                    </td>
                    <td className="px-4 py-3 text-center text-muted-foreground">
                      {cat.sort_order}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={cn(
                          'inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium',
                          cat.is_active
                            ? 'bg-green-100 text-green-700'
                            : 'bg-muted text-muted-foreground'
                        )}
                      >
                        {cat.is_active ? 'Aktif' : 'Non-aktif'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex justify-end gap-1">
                        <button
                          onClick={() => toggle(cat)}
                          className="rounded-md p-1.5 hover:bg-accent"
                          title={cat.is_active ? 'Nonaktifkan' : 'Aktifkan'}
                        >
                          {cat.is_active ? (
                            <ToggleRight className="h-4 w-4 text-green-600" />
                          ) : (
                            <ToggleLeft className="h-4 w-4 text-muted-foreground" />
                          )}
                        </button>
                        <button
                          onClick={() => openEdit(cat)}
                          className="rounded-md p-1.5 text-muted-foreground hover:bg-accent"
                          title="Edit"
                        >
                          <Pencil className="h-4 w-4" />
                        </button>
                        <button
                          onClick={() => {
                            if (confirm(`Hapus kategori "${cat.name}"?`))
                              deleteMut.mutate(cat.id);
                          }}
                          className="rounded-md p-1.5 text-destructive hover:bg-destructive/10"
                          title="Hapus"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg border border-border bg-card p-6 shadow-xl">
            <div className="mb-4 flex items-center gap-2">
              <Tag className="h-5 w-5 text-primary" />
              <h2 className="text-lg font-semibold text-foreground">
                {editing ? 'Edit Kategori' : 'Tambah Kategori Proyek'}
              </h2>
            </div>

            {apiError && (
              <div className="mb-4 rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {apiError}
              </div>
            )}

            <div className="space-y-4">
              <label className="block">
                <span className="mb-1 block text-sm font-medium text-foreground">
                  Kode <span className="text-destructive">*</span>
                </span>
                <input
                  className={inputCls}
                  value={form.code}
                  onChange={(e) => setForm((f) => ({ ...f, code: e.target.value }))}
                  placeholder="BND"
                  maxLength={100}
                />
              </label>
              <label className="block">
                <span className="mb-1 block text-sm font-medium text-foreground">
                  Nama <span className="text-destructive">*</span>
                </span>
                <input
                  className={inputCls}
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  placeholder="Bendungan"
                  maxLength={300}
                />
              </label>
              <label className="block">
                <span className="mb-1 block text-sm font-medium text-foreground">Deskripsi</span>
                <textarea
                  className={textareaCls}
                  rows={3}
                  value={form.description ?? ''}
                  onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                  placeholder="Deskripsi singkat kategori..."
                />
              </label>
              <div className="grid grid-cols-2 gap-3">
                <label className="block">
                  <span className="mb-1 block text-sm font-medium text-foreground">Urutan</span>
                  <input
                    className={inputCls}
                    type="number"
                    min={0}
                    value={form.sort_order ?? 0}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, sort_order: Number(e.target.value) }))
                    }
                  />
                </label>
                <label className="flex cursor-pointer items-center gap-2 pt-6 text-sm text-foreground">
                  <input
                    type="checkbox"
                    checked={form.is_active ?? true}
                    onChange={(e) => setForm((f) => ({ ...f, is_active: e.target.checked }))}
                    className="rounded"
                  />
                  Aktif
                </label>
              </div>
            </div>

            <div className="mt-6 flex justify-end gap-2">
              <button
                onClick={closeModal}
                className="inline-flex h-9 items-center rounded-md border border-input px-3 text-sm font-medium hover:bg-accent"
              >
                Batal
              </button>
              <button
                onClick={submit}
                disabled={!form.code.trim() || !form.name.trim() || isPending}
                className="inline-flex h-9 items-center rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {isPending ? 'Menyimpan...' : editing ? 'Simpan' : 'Tambah'}
              </button>
            </div>
          </div>
        </div>
      )}
    </DashboardLayout>
  );
}
