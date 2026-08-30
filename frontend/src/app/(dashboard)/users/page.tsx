"use client";

import { type FormEvent, useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AlertCircle,
  CheckCircle2,
  Edit3,
  Plus,
  Save,
  Search,
  UserMinus,
  X,
} from "lucide-react";
import { z } from "zod";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { userService } from "@/services/user.service";
import { cn, formatDate } from "@/lib/utils";
import type {
  CreateUserRequest,
  UpdateUserRequest,
  UserProfile,
} from "@/types/user";

const userFormSchema = z.object({
  first_name: z.string().trim().min(1, "Nama depan wajib diisi").max(100),
  last_name: z.string().trim().max(100).optional(),
  email: z.string().trim().email("Email belum valid"),
  password: z.string().min(8, "Password minimal 8 karakter").optional(),
  phone: z.string().trim().max(50).optional(),
  job_title: z.string().trim().max(200).optional(),
  employee_id: z.string().trim().optional(),
  is_active: z.boolean(),
  role_ids: z.string().trim().optional(),
});

type UserFormValues = z.infer<typeof userFormSchema>;
type FormMode = "create" | "edit";

export default function UsersPage() {
  const qc = useQueryClient();
  const [search, setSearch] = useState("");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [formMode, setFormMode] = useState<FormMode>("create");
  const [formError, setFormError] = useState<string | null>(null);

  const usersQuery = useQuery({
    queryKey: ["users", { search }],
    queryFn: () =>
      userService.list({ search: search || undefined, page: 1, page_size: 100 }),
    select: (res) => res.data.data ?? [],
  });

  const users = useMemo(() => usersQuery.data ?? [], [usersQuery.data]);

  useEffect(() => {
    if (!selectedId && users.length > 0) {
      setSelectedId(users[0].id);
    }
    if (selectedId && users.length > 0 && !users.some((user) => user.id === selectedId)) {
      setSelectedId(users[0].id);
    }
  }, [selectedId, users]);

  const detailQuery = useQuery({
    queryKey: ["users", selectedId],
    queryFn: () => userService.get(selectedId ?? ""),
    select: (res) => res.data.data,
    enabled: Boolean(selectedId),
  });

  const selectedUser = detailQuery.data ?? users.find((user) => user.id === selectedId) ?? null;

  const createMutation = useMutation({
    mutationFn: (payload: CreateUserRequest) => userService.create(payload),
    onSuccess: (res) => {
      setSelectedId(res.data.data.id);
      setShowForm(false);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["users"] });
      void qc.invalidateQueries({ queryKey: ["dashboard", "stats"] });
    },
    onError: (error: unknown) => setFormError(readMutationError(error)),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateUserRequest }) =>
      userService.update(id, payload),
    onSuccess: (res) => {
      const user = res.data.data;
      setSelectedId(user.id);
      setShowForm(false);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["users"] });
      void qc.invalidateQueries({ queryKey: ["users", user.id] });
      void qc.invalidateQueries({ queryKey: ["dashboard", "stats"] });
    },
    onError: (error: unknown) => setFormError(readMutationError(error)),
  });

  const deactivateMutation = useMutation({
    mutationFn: (id: string) => userService.deactivate(id),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["users"] });
      void qc.invalidateQueries({ queryKey: ["users", selectedId] });
      void qc.invalidateQueries({ queryKey: ["dashboard", "stats"] });
    },
  });

  function startCreate() {
    setFormMode("create");
    setShowForm(true);
    setFormError(null);
  }

  function startEdit(user: UserProfile) {
    setSelectedId(user.id);
    setFormMode("edit");
    setShowForm(true);
    setFormError(null);
  }

  function submitUser(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);

    const parsed = userFormSchema.safeParse(readUserForm(event.currentTarget));
    if (!parsed.success) {
      setFormError(parsed.error.issues[0]?.message ?? "Data pengguna belum valid.");
      return;
    }

    if (formMode === "create") {
      if (!parsed.data.password) {
        setFormError("Password wajib diisi untuk pengguna baru.");
        return;
      }
      createMutation.mutate(toCreatePayload(parsed.data));
      return;
    }

    if (!selectedUser) {
      setFormError("Pilih pengguna sebelum menyimpan perubahan.");
      return;
    }
    updateMutation.mutate({
      id: selectedUser.id,
      payload: toUpdatePayload(parsed.data),
    });
  }

  const isSaving = createMutation.isPending || updateMutation.isPending;

  return (
    <DashboardLayout title="Pengguna">
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-lg font-semibold text-foreground">Manajemen Pengguna</h2>
          <p className="text-sm text-muted-foreground">
            {users.length} pengguna dalam organisasi ini
          </p>
        </div>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              type="search"
              placeholder="Cari pengguna"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              className={inputClassNameWithIcon}
            />
          </div>
          <button type="button" onClick={startCreate} className={primaryButton}>
            <Plus className="h-4 w-4" aria-hidden="true" />
            Pengguna Baru
          </button>
        </div>
      </div>

      {showForm && (
        <UserForm
          mode={formMode}
          user={formMode === "edit" ? selectedUser : null}
          error={formError}
          isSaving={isSaving}
          onSubmit={submitUser}
          onCancel={() => {
            setShowForm(false);
            setFormError(null);
          }}
        />
      )}

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.25fr)_minmax(360px,0.75fr)]">
        <section className="rounded-lg border border-border bg-card shadow-sm">
          <div className="border-b border-border px-4 py-3">
            <h3 className="text-sm font-semibold text-foreground">Daftar Pengguna</h3>
          </div>
          {usersQuery.isLoading ? (
            <LoadingState label="Memuat pengguna..." />
          ) : usersQuery.isError ? (
            <ErrorState label="Daftar pengguna belum dapat dimuat." />
          ) : users.length === 0 ? (
            <EmptyState
              label={search ? "Tidak ada pengguna yang cocok." : "Belum ada pengguna."}
              actionLabel={!search ? "Tambah pengguna" : undefined}
              onAction={!search ? startCreate : undefined}
            />
          ) : (
            <div className="overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border bg-muted/40">
                    <th className="px-4 py-3 text-left font-medium text-muted-foreground">
                      Pengguna
                    </th>
                    <th className="hidden px-4 py-3 text-left font-medium text-muted-foreground md:table-cell">
                      Role
                    </th>
                    <th className="hidden px-4 py-3 text-left font-medium text-muted-foreground sm:table-cell">
                      Status
                    </th>
                    <th className="hidden px-4 py-3 text-left font-medium text-muted-foreground lg:table-cell">
                      Dibuat
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {users.map((user) => (
                    <tr
                      key={user.id}
                      onClick={() => setSelectedId(user.id)}
                      className={cn(
                        "cursor-pointer transition-colors hover:bg-muted/30",
                        selectedId === user.id && "bg-muted/40"
                      )}
                    >
                      <td className="px-4 py-3">
                        <div>
                          <p className="font-medium text-foreground">{displayName(user)}</p>
                          <p className="text-xs text-muted-foreground">{user.email}</p>
                        </div>
                      </td>
                      <td className="hidden px-4 py-3 md:table-cell">
                        <RolePills roles={user.roles ?? []} compact />
                      </td>
                      <td className="hidden px-4 py-3 sm:table-cell">
                        <StatusBadge active={user.is_active} />
                      </td>
                      <td className="hidden px-4 py-3 text-muted-foreground lg:table-cell">
                        {formatDate(user.created_at)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>

        <UserDetail
          user={selectedUser}
          isLoading={detailQuery.isLoading}
          onEdit={startEdit}
          onDeactivate={(user) => {
            if (window.confirm(`Nonaktifkan ${displayName(user)}?`)) {
              deactivateMutation.mutate(user.id);
            }
          }}
        />
      </div>
    </DashboardLayout>
  );
}

function UserForm({
  mode,
  user,
  error,
  isSaving,
  onSubmit,
  onCancel,
}: {
  mode: FormMode;
  user: UserProfile | null;
  error: string | null;
  isSaving: boolean;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onCancel: () => void;
}) {
  return (
    <form onSubmit={onSubmit} className="mb-6 rounded-lg border border-border bg-card p-5 shadow-sm">
      <div className="mb-4 flex items-center justify-between gap-3">
        <div>
          <h3 className="text-sm font-semibold text-foreground">
            {mode === "create" ? "Tambah Pengguna" : "Edit Pengguna"}
          </h3>
          <p className="text-xs text-muted-foreground">
            Role ID opsional dan akan disinkronkan dengan RBAC backend jika diisi.
          </p>
        </div>
        <button
          type="button"
          onClick={onCancel}
          className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          aria-label="Tutup form pengguna"
        >
          <X className="h-4 w-4" aria-hidden="true" />
        </button>
      </div>

      {error && (
        <div className="mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-4">
        <Field label="Nama Depan">
          <input name="first_name" defaultValue={user?.first_name ?? ""} className={inputClassName} />
        </Field>
        <Field label="Nama Belakang">
          <input name="last_name" defaultValue={user?.last_name ?? ""} className={inputClassName} />
        </Field>
        <Field label="Email">
          <input
            name="email"
            type="email"
            defaultValue={user?.email ?? ""}
            disabled={mode === "edit"}
            className={inputClassName}
          />
        </Field>
        <Field label="Password">
          <input
            name="password"
            type="password"
            placeholder={mode === "edit" ? "Tidak diubah" : "Minimal 8 karakter"}
            className={inputClassName}
          />
        </Field>
        <Field label="NIP / ID Pegawai">
          <input name="employee_id" defaultValue={user?.employee_id ?? ""} className={inputClassName} />
        </Field>
        <Field label="Telepon">
          <input name="phone" defaultValue={user?.phone ?? ""} className={inputClassName} />
        </Field>
        <Field label="Jabatan">
          <input name="job_title" defaultValue={user?.job_title ?? ""} className={inputClassName} />
        </Field>
        <label className="flex items-end gap-2 pb-2 text-sm text-muted-foreground">
          <input
            name="is_active"
            type="checkbox"
            defaultChecked={user?.is_active ?? true}
            className="h-4 w-4 rounded border-input text-primary focus:ring-ring"
          />
          Aktif
        </label>
      </div>

      <div className="mt-4">
        <Field label="Role IDs">
          <input
            name="role_ids"
            defaultValue={(user?.roles ?? []).map((role) => role.id).join(", ")}
            className={inputClassName}
            placeholder="uuid-1, uuid-2"
          />
        </Field>
      </div>

      <div className="mt-4 flex justify-end gap-2">
        <button
          type="button"
          onClick={onCancel}
          className="inline-flex h-9 items-center rounded-md border border-input px-3 text-sm hover:bg-accent"
        >
          Batal
        </button>
        <button type="submit" disabled={isSaving} className={primaryButton}>
          <Save className="h-4 w-4" aria-hidden="true" />
          {isSaving ? "Menyimpan..." : "Simpan"}
        </button>
      </div>
    </form>
  );
}

function UserDetail({
  user,
  isLoading,
  onEdit,
  onDeactivate,
}: {
  user: UserProfile | null;
  isLoading: boolean;
  onEdit: (user: UserProfile) => void;
  onDeactivate: (user: UserProfile) => void;
}) {
  if (isLoading) {
    return (
      <aside className="rounded-lg border border-border bg-card p-5 shadow-sm">
        <div className="h-5 w-32 rounded-md bg-muted" />
        <div className="mt-4 h-24 rounded-md bg-muted" />
      </aside>
    );
  }

  if (!user) {
    return (
      <aside className="rounded-lg border border-border bg-card p-5 shadow-sm">
        <p className="text-sm text-muted-foreground">Pilih pengguna untuk melihat detail.</p>
      </aside>
    );
  }

  return (
    <aside className="space-y-6 rounded-lg border border-border bg-card p-5 shadow-sm">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-medium uppercase text-muted-foreground">
            {user.employee_id || "Belum ada ID pegawai"}
          </p>
          <h3 className="mt-1 text-lg font-semibold text-foreground">{displayName(user)}</h3>
          <p className="mt-1 text-sm text-muted-foreground">{user.email}</p>
        </div>
        <div className="flex gap-1">
          <button
            type="button"
            onClick={() => onEdit(user)}
            className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
            aria-label="Edit pengguna"
          >
            <Edit3 className="h-4 w-4" aria-hidden="true" />
          </button>
          <button
            type="button"
            onClick={() => onDeactivate(user)}
            disabled={!user.is_active}
            className="rounded-md p-2 text-muted-foreground hover:bg-destructive/10 hover:text-destructive disabled:cursor-not-allowed disabled:opacity-50"
            aria-label="Nonaktifkan pengguna"
          >
            <UserMinus className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
      </div>

      <div className="flex flex-wrap gap-2">
        <StatusBadge active={user.is_active} />
        {user.must_change_pwd && (
          <span className="inline-flex rounded-full bg-yellow-100 px-2.5 py-0.5 text-xs font-medium text-yellow-700">
            Wajib ganti password
          </span>
        )}
      </div>

      <dl className="grid grid-cols-2 gap-4 text-sm">
        <div>
          <dt className="text-muted-foreground">Jabatan</dt>
          <dd className="font-medium text-foreground">{user.job_title || "-"}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Telepon</dt>
          <dd className="font-medium text-foreground">{user.phone || "-"}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Dibuat</dt>
          <dd className="font-medium text-foreground">{formatDate(user.created_at)}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Login Terakhir</dt>
          <dd className="font-medium text-foreground">{formatDate(user.last_login_at)}</dd>
        </div>
      </dl>

      <div>
        <h4 className="mb-2 text-sm font-semibold text-foreground">Roles</h4>
        <RolePills roles={user.roles ?? []} />
      </div>
    </aside>
  );
}

function RolePills({ roles, compact = false }: { roles: UserProfile["roles"]; compact?: boolean }) {
  if (!roles || roles.length === 0) {
    return <span className="text-sm text-muted-foreground">Belum ada role</span>;
  }

  const visible = compact ? roles.slice(0, 2) : roles;
  return (
    <div className="flex flex-wrap gap-1.5">
      {visible.map((role) => (
        <span
          key={role.id}
          className="inline-flex rounded-full bg-blue-50 px-2.5 py-0.5 text-xs font-medium text-blue-700"
          title={role.name}
        >
          {role.code}
        </span>
      ))}
      {compact && roles.length > visible.length && (
        <span className="text-xs text-muted-foreground">+{roles.length - visible.length}</span>
      )}
    </div>
  );
}

function StatusBadge({ active }: { active: boolean }) {
  return (
    <span
      className={cn(
        "inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium",
        active ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-700"
      )}
    >
      {active ? "Aktif" : "Nonaktif"}
    </span>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="mb-1 block text-sm font-medium text-foreground">{label}</span>
      {children}
    </label>
  );
}

function EmptyState({
  label,
  actionLabel,
  onAction,
}: {
  label: string;
  actionLabel?: string;
  onAction?: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <CheckCircle2 className="mb-3 h-6 w-6 text-green-700" aria-hidden="true" />
      <p className="text-sm text-muted-foreground">{label}</p>
      {actionLabel && onAction && (
        <button type="button" onClick={onAction} className={cn(primaryButton, "mt-4")}>
          <Plus className="h-4 w-4" aria-hidden="true" />
          {actionLabel}
        </button>
      )}
    </div>
  );
}

function ErrorState({ label }: { label: string }) {
  return (
    <div className="flex items-center gap-3 p-6 text-sm text-destructive">
      <AlertCircle className="h-4 w-4" aria-hidden="true" />
      {label}
    </div>
  );
}

function LoadingState({ label }: { label: string }) {
  return <div className="p-6 text-sm text-muted-foreground">{label}</div>;
}

function readUserForm(form: HTMLFormElement): UserFormValues {
  const data = new FormData(form);
  return {
    first_name: getString(data, "first_name"),
    last_name: getString(data, "last_name"),
    email: getString(data, "email"),
    password: getString(data, "password") || undefined,
    phone: getString(data, "phone"),
    job_title: getString(data, "job_title"),
    employee_id: getString(data, "employee_id"),
    is_active: data.get("is_active") === "on",
    role_ids: getString(data, "role_ids"),
  };
}

function toCreatePayload(values: UserFormValues): CreateUserRequest {
  return {
    first_name: values.first_name,
    last_name: optionalString(values.last_name),
    email: values.email,
    password: values.password ?? "",
    phone: optionalString(values.phone),
    job_title: optionalString(values.job_title),
    employee_id: optionalString(values.employee_id),
    is_active: values.is_active,
    role_ids: parseRoleIDs(values.role_ids),
  };
}

function toUpdatePayload(values: UserFormValues): UpdateUserRequest {
  return {
    first_name: values.first_name,
    last_name: optionalString(values.last_name),
    phone: optionalString(values.phone),
    job_title: optionalString(values.job_title),
    employee_id: optionalString(values.employee_id),
    is_active: values.is_active,
    role_ids: parseRoleIDs(values.role_ids),
  };
}

function parseRoleIDs(value: string | undefined) {
  const ids = value
    ?.split(",")
    .map((item) => item.trim())
    .filter(Boolean);
  return ids && ids.length > 0 ? ids : undefined;
}

function getString(data: FormData, key: string) {
  const value = data.get(key);
  return typeof value === "string" ? value : "";
}

function optionalString(value: string | undefined) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

function displayName(user: UserProfile) {
  return user.full_name.trim() || `${user.first_name} ${user.last_name}`.trim() || user.email;
}

function readMutationError(error: unknown) {
  const maybeError = error as { response?: { data?: { message?: string } } };
  return maybeError.response?.data?.message ?? "Pengguna belum dapat disimpan.";
}

const inputClassName = cn(
  "h-9 w-full rounded-md border border-input bg-background px-3 text-sm",
  "placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring",
  "disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground"
);

const inputClassNameWithIcon = cn(inputClassName, "pl-9 sm:w-56");

const primaryButton = cn(
  "inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground",
  "hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
);
