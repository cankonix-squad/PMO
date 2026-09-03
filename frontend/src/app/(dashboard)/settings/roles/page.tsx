"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import rbacService from "@/services/rbac.service";
import { Role, CreateRoleRequest, UpdateRoleRequest } from "@/types/rbac";
import { cn } from "@/lib/utils";
import { ShieldCheck, Plus, Pencil, Trash2, Lock } from "lucide-react";

// ── helpers ────────────────────────────────────────────────────────────────

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

// ── SystemBadge ────────────────────────────────────────────────────────────

function SystemBadge({ isSystem }: { isSystem: boolean }) {
  if (!isSystem) return null;
  return (
    <span className="inline-flex items-center gap-1 rounded px-2 py-0.5 text-xs font-medium bg-amber-50 text-amber-700 border border-amber-200">
      <Lock className="h-3 w-3" />
      Sistem
    </span>
  );
}

// ── RoleModal ──────────────────────────────────────────────────────────────

interface RoleModalProps {
  open: boolean;
  onClose: () => void;
  editing: Role | null;
}

function RoleModal({ open, onClose, editing }: RoleModalProps) {
  const qc = useQueryClient();
  const isEdit = editing !== null;

  const [code, setCode] = useState(editing?.code ?? "");
  const [name, setName] = useState(editing?.name ?? "");
  const [description, setDescription] = useState(editing?.description ?? "");
  const [error, setError] = useState("");

  // Reset form when modal opens
  useState(() => {
    setCode(editing?.code ?? "");
    setName(editing?.name ?? "");
    setDescription(editing?.description ?? "");
    setError("");
  });

  const createMutation = useMutation({
    mutationFn: (req: CreateRoleRequest) => rbacService.createRole(req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["roles"] });
      onClose();
    },
    onError: (err: { response?: { data?: { message?: string } } }) => {
      setError(err.response?.data?.message ?? "Gagal menyimpan role");
    },
  });

  const updateMutation = useMutation({
    mutationFn: (req: UpdateRoleRequest) =>
      rbacService.updateRole(editing!.id, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["roles"] });
      onClose();
    },
    onError: (err: { response?: { data?: { message?: string } } }) => {
      setError(err.response?.data?.message ?? "Gagal menyimpan role");
    },
  });

  if (!open) return null;

  const isPending = createMutation.isPending || updateMutation.isPending;

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!name.trim()) {
      setError("Nama role wajib diisi");
      return;
    }
    if (isEdit) {
      updateMutation.mutate({ name: name.trim(), description: description.trim() });
    } else {
      if (!code.trim()) {
        setError("Kode role wajib diisi");
        return;
      }
      createMutation.mutate({
        code: code.trim().toUpperCase(),
        name: name.trim(),
        description: description.trim(),
      });
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4">
      <div className="w-full max-w-md rounded-xl bg-white shadow-xl">
        <div className="flex items-center justify-between border-b px-6 py-4">
          <h2 className="text-base font-semibold text-gray-900">
            {isEdit ? "Edit Role" : "Tambah Role Baru"}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            ✕
          </button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4 px-6 py-5">
          {!isEdit && (
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                Kode Role <span className="text-red-500">*</span>
              </label>
              <input
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm uppercase focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Contoh: REVIEWER"
                value={code}
                onChange={(e) => setCode(e.target.value.toUpperCase())}
              />
              <p className="mt-1 text-xs text-gray-500">
                Kode unik, huruf kapital. Tidak bisa diubah setelah dibuat.
              </p>
            </div>
          )}
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              Nama Role <span className="text-red-500">*</span>
            </label>
            <input
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Contoh: Reviewer Anggaran"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              Deskripsi
            </label>
            <textarea
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Opsional — jelaskan fungsi role ini"
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          {error && (
            <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">
              {error}
            </p>
          )}
          <div className="flex justify-end gap-3 pt-1">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={isPending}
              className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60"
            >
              {isPending ? "Menyimpan..." : isEdit ? "Simpan Perubahan" : "Tambah Role"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ── DeleteDialog ───────────────────────────────────────────────────────────

interface DeleteDialogProps {
  role: Role | null;
  onClose: () => void;
}

function DeleteDialog({ role, onClose }: DeleteDialogProps) {
  const qc = useQueryClient();
  const [error, setError] = useState("");

  const deleteMutation = useMutation({
    mutationFn: () => rbacService.deleteRole(role!.id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["roles"] });
      onClose();
    },
    onError: (err: { response?: { data?: { message?: string } } }) => {
      setError(err.response?.data?.message ?? "Gagal menghapus role");
    },
  });

  if (!role) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4">
      <div className="w-full max-w-sm rounded-xl bg-white shadow-xl">
        <div className="px-6 py-5">
          <h2 className="text-base font-semibold text-gray-900">Hapus Role</h2>
          <p className="mt-2 text-sm text-gray-600">
            Yakin ingin menghapus role{" "}
            <span className="font-medium text-gray-900">{role.name}</span>?
            Tindakan ini tidak dapat dibatalkan.
          </p>
          {error && (
            <p className="mt-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">
              {error}
            </p>
          )}
          <div className="mt-5 flex justify-end gap-3">
            <button
              onClick={onClose}
              className="rounded-lg border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
            >
              Batal
            </button>
            <button
              onClick={() => deleteMutation.mutate()}
              disabled={deleteMutation.isPending}
              className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-60"
            >
              {deleteMutation.isPending ? "Menghapus..." : "Hapus"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

// ── Page ───────────────────────────────────────────────────────────────────

export default function RolesPage() {
  const [search, setSearch] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Role | null>(null);
  const [deleting, setDeleting] = useState<Role | null>(null);

  const { data: roles = [], isLoading, isError } = useQuery({
    queryKey: ["roles"],
    queryFn: () => rbacService.listRoles(),
  });

  const filtered = roles.filter(
    (r) =>
      r.name.toLowerCase().includes(search.toLowerCase()) ||
      r.code.toLowerCase().includes(search.toLowerCase()) ||
      (r.description ?? "").toLowerCase().includes(search.toLowerCase())
  );

  function openCreate() {
    setEditing(null);
    setModalOpen(true);
  }

  function openEdit(role: Role) {
    setEditing(role);
    setModalOpen(true);
  }

  return (
    <DashboardLayout title="Manajemen Role">
      <div className="space-y-6 p-6">
        {/* Header */}
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-xl font-semibold text-gray-900">Manajemen Role</h1>
            <p className="mt-0.5 text-sm text-gray-500">
              Kelola role dan hak akses pengguna dalam organisasi
            </p>
          </div>
          <button
            onClick={openCreate}
            className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
          >
            <Plus className="h-4 w-4" />
            Tambah Role
          </button>
        </div>

        {/* Search */}
        <div className="relative max-w-sm">
          <input
            className="w-full rounded-lg border border-gray-300 py-2 pl-9 pr-3 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Cari nama atau kode role..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <span className="absolute left-3 top-2.5 text-gray-400 text-xs">🔍</span>
        </div>

        {/* Table */}
        <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
          {isLoading ? (
            <div className="py-16 text-center text-sm text-gray-500">Memuat data role...</div>
          ) : isError ? (
            <div className="py-16 text-center text-sm text-red-500">Gagal memuat data. Coba refresh halaman.</div>
          ) : filtered.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-gray-400">
              <ShieldCheck className="mb-3 h-10 w-10 opacity-30" />
              <p className="text-sm">
                {search ? "Tidak ada role yang cocok" : "Belum ada role"}
              </p>
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
                  <th className="px-4 py-3">Kode</th>
                  <th className="px-4 py-3">Nama Role</th>
                  <th className="hidden px-4 py-3 md:table-cell">Deskripsi</th>
                  <th className="px-4 py-3">Tipe</th>
                  <th className="hidden px-4 py-3 lg:table-cell">Dibuat</th>
                  <th className="px-4 py-3 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filtered.map((role) => (
                  <tr key={role.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <code className="rounded bg-gray-100 px-2 py-0.5 text-xs font-mono text-gray-700">
                        {role.code}
                      </code>
                    </td>
                    <td className="px-4 py-3 font-medium text-gray-900">
                      {role.name}
                    </td>
                    <td className="hidden px-4 py-3 text-gray-500 md:table-cell max-w-xs">
                      <span className="line-clamp-2">
                        {role.description || <span className="italic text-gray-300">—</span>}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <SystemBadge isSystem={role.is_system} />
                      {!role.is_system && (
                        <span className="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium bg-blue-50 text-blue-700 border border-blue-200">
                          Kustom
                        </span>
                      )}
                    </td>
                    <td className="hidden px-4 py-3 text-gray-500 lg:table-cell">
                      {formatDate(role.created_at)}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex justify-end gap-1">
                        <button
                          onClick={() => openEdit(role)}
                          disabled={role.is_system}
                          title={role.is_system ? "Role sistem tidak bisa diubah" : "Edit role"}
                          className={cn(
                            "rounded-lg p-1.5 text-gray-500 hover:bg-gray-100 hover:text-blue-600",
                            role.is_system && "cursor-not-allowed opacity-30"
                          )}
                        >
                          <Pencil className="h-4 w-4" />
                        </button>
                        <button
                          onClick={() => setDeleting(role)}
                          disabled={role.is_system}
                          title={role.is_system ? "Role sistem tidak bisa dihapus" : "Hapus role"}
                          className={cn(
                            "rounded-lg p-1.5 text-gray-500 hover:bg-gray-100 hover:text-red-600",
                            role.is_system && "cursor-not-allowed opacity-30"
                          )}
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        <p className="text-xs text-gray-400">
          Role dengan label <span className="font-medium text-amber-600">Sistem</span> adalah bawaan aplikasi dan tidak dapat diubah atau dihapus.
        </p>
      </div>

      {/* Modals */}
      <RoleModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        editing={editing}
      />
      <DeleteDialog
        role={deleting}
        onClose={() => setDeleting(null)}
      />
    </DashboardLayout>
  );
}
