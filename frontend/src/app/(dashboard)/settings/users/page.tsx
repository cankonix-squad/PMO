"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { UserPlus, Pencil, UserX, UserCheck, Search, X, ChevronDown } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { userService } from "@/services/user.service";
import rbacService from "@/services/rbac.service";
import { cn } from "@/lib/utils";
import type { UserProfile, CreateUserRequest, UpdateUserRequest } from "@/types/user";
import type { Role } from "@/types/rbac";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function formatDate(iso?: string | null) {
  if (!iso) return "—";
  return new Date(iso).toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

function RoleBadge({ name }: { name: string }) {
  return (
    <span className="inline-flex items-center rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700 ring-1 ring-inset ring-blue-700/10">
      {name}
    </span>
  );
}

function StatusBadge({ active }: { active: boolean }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset",
        active
          ? "bg-green-50 text-green-700 ring-green-600/20"
          : "bg-red-50 text-red-600 ring-red-500/20"
      )}
    >
      {active ? "Aktif" : "Nonaktif"}
    </span>
  );
}

// ---------------------------------------------------------------------------
// Multi-select role picker
// ---------------------------------------------------------------------------

function RolePicker({
  roles,
  selected,
  onChange,
}: {
  roles: Role[];
  selected: string[];
  onChange: (ids: string[]) => void;
}) {
  const [open, setOpen] = useState(false);

  function toggle(id: string) {
    onChange(selected.includes(id) ? selected.filter((x) => x !== id) : [...selected, id]);
  }

  const selectedNames = roles
    .filter((r) => selected.includes(r.id))
    .map((r) => r.name)
    .join(", ");

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center justify-between rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        <span className={cn("truncate", !selectedNames && "text-gray-400")}>
          {selectedNames || "Pilih role…"}
        </span>
        <ChevronDown className="ml-2 h-4 w-4 shrink-0 text-gray-400" />
      </button>
      {open && (
        <div className="absolute z-20 mt-1 w-full rounded-md border border-gray-200 bg-white shadow-lg">
          {roles.length === 0 ? (
            <p className="px-3 py-2 text-sm text-gray-500">Tidak ada role tersedia</p>
          ) : (
            roles.map((r) => (
              <label
                key={r.id}
                className="flex cursor-pointer items-center gap-2 px-3 py-2 text-sm hover:bg-gray-50"
              >
                <input
                  type="checkbox"
                  checked={selected.includes(r.id)}
                  onChange={() => toggle(r.id)}
                  className="h-4 w-4 rounded border-gray-300 text-blue-600"
                />
                <span>{r.name}</span>
                {r.is_system && (
                  <span className="ml-auto text-[10px] text-gray-400">sistem</span>
                )}
              </label>
            ))
          )}
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Create / Edit Modal
// ---------------------------------------------------------------------------

interface ModalProps {
  user: UserProfile | null; // null = create mode
  roles: Role[];
  onClose: () => void;
  onSaved: () => void;
}

function UserModal({ user, roles, onClose, onSaved }: ModalProps) {
  const isEdit = user !== null;

  const [firstName, setFirstName] = useState(user?.first_name ?? "");
  const [lastName, setLastName] = useState(user?.last_name ?? "");
  const [email, setEmail] = useState(user?.email ?? "");
  const [password, setPassword] = useState("");
  const [phone, setPhone] = useState(user?.phone ?? "");
  const [jobTitle, setJobTitle] = useState(user?.job_title ?? "");
  const [employeeId, setEmployeeId] = useState(user?.employee_id ?? "");
  const [roleIds, setRoleIds] = useState<string[]>(
    user?.roles?.map((r) => r.id) ?? []
  );
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSaving(true);
    try {
      if (isEdit) {
        const req: UpdateUserRequest = {
          first_name: firstName || undefined,
          last_name: lastName || undefined,
          phone: phone || undefined,
          job_title: jobTitle || undefined,
          employee_id: employeeId || undefined,
          role_ids: roleIds,
        };
        await userService.update(user!.id, req);
      } else {
        if (!firstName || !email || !password) {
          setError("Nama depan, email, dan password wajib diisi.");
          setSaving(false);
          return;
        }
        const req: CreateUserRequest = {
          first_name: firstName,
          last_name: lastName || undefined,
          email,
          password,
          phone: phone || undefined,
          job_title: jobTitle || undefined,
          employee_id: employeeId || undefined,
          is_active: true,
          role_ids: roleIds,
        };
        await userService.create(req);
      }
      onSaved();
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ??
        "Terjadi kesalahan. Coba lagi.";
      setError(msg);
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-lg rounded-xl bg-white shadow-xl">
        {/* Header */}
        <div className="flex items-center justify-between border-b px-6 py-4">
          <h2 className="text-base font-semibold text-gray-900">
            {isEdit ? "Edit Pengguna" : "Tambah Pengguna"}
          </h2>
          <button onClick={onClose} className="rounded-md p-1 hover:bg-gray-100">
            <X className="h-5 w-5 text-gray-500" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4 px-6 py-5">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                Nama Depan <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                placeholder="Budi"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">Nama Belakang</label>
              <input
                type="text"
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                placeholder="Santoso"
              />
            </div>
          </div>

          {!isEdit && (
            <>
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Email <span className="text-red-500">*</span>
                </label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  placeholder="budi@example.com"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Password <span className="text-red-500">*</span>
                </label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  placeholder="Min. 8 karakter"
                />
              </div>
            </>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">No. Telepon</label>
              <input
                type="tel"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                placeholder="+62812…"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">NIP / ID Pegawai</label>
              <input
                type="text"
                value={employeeId}
                onChange={(e) => setEmployeeId(e.target.value)}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                placeholder="197001011990031001"
              />
            </div>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">Jabatan</label>
            <input
              type="text"
              value={jobTitle}
              onChange={(e) => setJobTitle(e.target.value)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="Kepala Balai"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">Role</label>
            <RolePicker roles={roles} selected={roleIds} onChange={setRoleIds} />
          </div>

          {error && (
            <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-600">{error}</p>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={saving}
              className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60"
            >
              {saving ? "Menyimpan…" : isEdit ? "Simpan Perubahan" : "Tambah Pengguna"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Page
// ---------------------------------------------------------------------------

export default function UsersSettingsPage() {
  const queryClient = useQueryClient();
  const [search, setSearch] = useState("");
  const [filterActive, setFilterActive] = useState<"all" | "active" | "inactive">("all");
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<UserProfile | null>(null);

  // Queries
  const { data: users = [], isLoading, error } = useQuery({
    queryKey: ["users", filterActive],
    queryFn: async () => {
      const params =
        filterActive === "active"
          ? { is_active: true }
          : filterActive === "inactive"
          ? { is_active: false }
          : undefined;
      const res = await userService.list(params);
      return res.data.data ?? [];
    },
  });

  const { data: roles = [] } = useQuery({
    queryKey: ["roles"],
    queryFn: rbacService.listRoles,
  });

  const deactivateMutation = useMutation({
    mutationFn: (id: string) => userService.deactivate(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["users"] }),
  });

  // Filtered list
  const filtered = users.filter((u) => {
    if (!search) return true;
    const q = search.toLowerCase();
    return (
      u.full_name.toLowerCase().includes(q) ||
      u.email.toLowerCase().includes(q) ||
      (u.job_title ?? "").toLowerCase().includes(q)
    );
  });

  function handleSaved() {
    setModalOpen(false);
    setEditing(null);
    queryClient.invalidateQueries({ queryKey: ["users"] });
  }

  function openCreate() {
    setEditing(null);
    setModalOpen(true);
  }

  function openEdit(user: UserProfile) {
    setEditing(user);
    setModalOpen(true);
  }

  return (
    <DashboardLayout title="Manajemen Pengguna">
      <div className="space-y-6 p-6">
        {/* Page header */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Manajemen Pengguna</h1>
            <p className="mt-1 text-sm text-gray-500">
              Kelola akun, role, dan status pengguna dalam organisasi.
            </p>
          </div>
          <button
            onClick={openCreate}
            className="inline-flex items-center gap-2 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700"
          >
            <UserPlus className="h-4 w-4" />
            Tambah Pengguna
          </button>
        </div>

        {/* Filters */}
        <div className="flex flex-col gap-3 sm:flex-row">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Cari nama, email, jabatan…"
              className="w-full rounded-md border border-gray-300 py-2 pl-9 pr-3 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
          <div className="flex gap-1 rounded-md border border-gray-300 bg-white p-1 shadow-sm">
            {(["all", "active", "inactive"] as const).map((opt) => (
              <button
                key={opt}
                onClick={() => setFilterActive(opt)}
                className={cn(
                  "rounded px-3 py-1 text-sm font-medium transition-colors",
                  filterActive === opt
                    ? "bg-blue-600 text-white"
                    : "text-gray-600 hover:text-gray-900"
                )}
              >
                {opt === "all" ? "Semua" : opt === "active" ? "Aktif" : "Nonaktif"}
              </button>
            ))}
          </div>
        </div>

        {/* Table */}
        {isLoading ? (
          <div className="flex h-40 items-center justify-center text-sm text-gray-500">
            Memuat data pengguna…
          </div>
        ) : error ? (
          <div className="flex h-40 items-center justify-center text-sm text-red-500">
            Gagal memuat data. Coba refresh halaman.
          </div>
        ) : filtered.length === 0 ? (
          <div className="flex h-40 items-center justify-center text-sm text-gray-500">
            {search ? "Tidak ada pengguna yang cocok." : "Belum ada pengguna."}
          </div>
        ) : (
          <div className="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm">
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Nama</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Email</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Jabatan</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Role</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Status</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Login Terakhir</th>
                  <th className="px-4 py-3 text-right font-medium text-gray-500">Aksi</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filtered.map((user) => (
                  <tr key={user.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <div className="font-medium text-gray-900">{user.full_name}</div>
                      {user.employee_id && (
                        <div className="text-xs text-gray-400">{user.employee_id}</div>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-600">{user.email}</td>
                    <td className="px-4 py-3 text-gray-600">{user.job_title || "—"}</td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {user.roles && user.roles.length > 0 ? (
                          user.roles.map((r) => <RoleBadge key={r.id} name={r.name} />)
                        ) : (
                          <span className="text-gray-400">—</span>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge active={user.is_active} />
                    </td>
                    <td className="px-4 py-3 text-gray-500">{formatDate(user.last_login_at)}</td>
                    <td className="px-4 py-3">
                      <div className="flex justify-end gap-2">
                        <button
                          onClick={() => openEdit(user)}
                          title="Edit"
                          className="rounded-md p-1.5 text-gray-500 hover:bg-gray-100 hover:text-gray-700"
                        >
                          <Pencil className="h-4 w-4" />
                        </button>
                        {user.is_active ? (
                          <button
                            onClick={() => deactivateMutation.mutate(user.id)}
                            title="Nonaktifkan"
                            disabled={deactivateMutation.isPending}
                            className="rounded-md p-1.5 text-red-500 hover:bg-red-50 hover:text-red-700 disabled:opacity-50"
                          >
                            <UserX className="h-4 w-4" />
                          </button>
                        ) : (
                          <span title="Sudah nonaktif" className="rounded-md p-1.5 text-gray-300">
                            <UserCheck className="h-4 w-4" />
                          </span>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Modal */}
      {modalOpen && (
        <UserModal
          user={editing}
          roles={roles}
          onClose={() => {
            setModalOpen(false);
            setEditing(null);
          }}
          onSaved={handleSaved}
        />
      )}
    </DashboardLayout>
  );
}
