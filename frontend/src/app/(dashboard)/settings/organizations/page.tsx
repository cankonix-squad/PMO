"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Building2, Plus, Pencil, X, Search } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import organizationService from "@/services/organization.service";
import { cn } from "@/lib/utils";
import type {
  Organization,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
} from "@/types/organization";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

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

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

// ---------------------------------------------------------------------------
// Create / Edit Modal
// ---------------------------------------------------------------------------

interface ModalProps {
  org: Organization | null; // null = create mode
  onClose: () => void;
  onSaved: () => void;
}

function OrgModal({ org, onClose, onSaved }: ModalProps) {
  const isEdit = org !== null;

  const [code, setCode] = useState(org?.code ?? "");
  const [name, setName] = useState(org?.name ?? "");
  const [shortName, setShortName] = useState(org?.short_name ?? "");
  const [address, setAddress] = useState(org?.address ?? "");
  const [website, setWebsite] = useState(org?.website ?? "");
  const [isActive, setIsActive] = useState(org?.is_active ?? true);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    if (!name.trim()) {
      setError("Nama organisasi wajib diisi.");
      return;
    }
    if (!isEdit && !code.trim()) {
      setError("Kode organisasi wajib diisi.");
      return;
    }

    setSaving(true);
    try {
      if (isEdit) {
        const req: UpdateOrganizationRequest = {
          name: name.trim() || undefined,
          short_name: shortName.trim() || undefined,
          address: address.trim() || undefined,
          website: website.trim() || undefined,
          is_active: isActive,
        };
        await organizationService.update(org!.id, req);
      } else {
        const req: CreateOrganizationRequest = {
          code: code.trim(),
          name: name.trim(),
          short_name: shortName.trim() || undefined,
          address: address.trim() || undefined,
          website: website.trim() || undefined,
        };
        await organizationService.create(req);
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
            {isEdit ? "Edit Organisasi" : "Tambah Organisasi"}
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
                Kode {!isEdit && <span className="text-red-500">*</span>}
              </label>
              <input
                type="text"
                value={code}
                onChange={(e) => setCode(e.target.value.toUpperCase())}
                disabled={isEdit}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-400"
                placeholder="BWSNT2"
              />
              {isEdit && (
                <p className="mt-1 text-xs text-gray-400">Kode tidak dapat diubah.</p>
              )}
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                Nama Singkat
              </label>
              <input
                type="text"
                value={shortName}
                onChange={(e) => setShortName(e.target.value)}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                placeholder="BWS NT II"
              />
            </div>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              Nama Resmi <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="Balai Wilayah Sungai Nusa Tenggara II"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">Alamat</label>
            <textarea
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              rows={2}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="Jl. …"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">Website</label>
            <input
              type="url"
              value={website}
              onChange={(e) => setWebsite(e.target.value)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="https://bwsnt2.pu.go.id"
            />
          </div>

          {isEdit && (
            <div className="flex items-center gap-3">
              <button
                type="button"
                role="switch"
                aria-checked={isActive}
                onClick={() => setIsActive((v) => !v)}
                className={cn(
                  "relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors",
                  isActive ? "bg-blue-600" : "bg-gray-200"
                )}
              >
                <span
                  className={cn(
                    "pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition-transform",
                    isActive ? "translate-x-5" : "translate-x-0"
                  )}
                />
              </button>
              <span className="text-sm text-gray-700">
                {isActive ? "Aktif" : "Nonaktif"}
              </span>
            </div>
          )}

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
              {saving ? "Menyimpan…" : isEdit ? "Simpan Perubahan" : "Tambah Organisasi"}
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

export default function OrganizationsSettingsPage() {
  const queryClient = useQueryClient();
  const [search, setSearch] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Organization | null>(null);

  const { data: orgs = [], isLoading, error } = useQuery({
    queryKey: ["organizations"],
    queryFn: organizationService.list,
  });

  // Toggle active status inline
  const toggleMutation = useMutation({
    mutationFn: (org: Organization) =>
      organizationService.update(org.id, { is_active: !org.is_active }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["organizations"] }),
  });

  const filtered = orgs.filter((o) => {
    if (!search) return true;
    const q = search.toLowerCase();
    return (
      o.name.toLowerCase().includes(q) ||
      o.code.toLowerCase().includes(q) ||
      (o.short_name ?? "").toLowerCase().includes(q)
    );
  });

  function handleSaved() {
    setModalOpen(false);
    setEditing(null);
    queryClient.invalidateQueries({ queryKey: ["organizations"] });
  }

  return (
    <DashboardLayout title="Manajemen Organisasi">
      <div className="space-y-6 p-6">
        {/* Page header */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Manajemen Organisasi</h1>
            <p className="mt-1 text-sm text-gray-500">
              Kelola daftar organisasi yang terdaftar di platform.
            </p>
          </div>
          <button
            onClick={() => {
              setEditing(null);
              setModalOpen(true);
            }}
            className="inline-flex items-center gap-2 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700"
          >
            <Plus className="h-4 w-4" />
            Tambah Organisasi
          </button>
        </div>

        {/* Search */}
        <div className="relative max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari nama atau kode…"
            className="w-full rounded-md border border-gray-300 py-2 pl-9 pr-3 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          />
        </div>

        {/* Content */}
        {isLoading ? (
          <div className="flex h-40 items-center justify-center text-sm text-gray-500">
            Memuat data organisasi…
          </div>
        ) : error ? (
          <div className="flex h-40 items-center justify-center text-sm text-red-500">
            Gagal memuat data. Coba refresh halaman.
          </div>
        ) : filtered.length === 0 ? (
          <div className="flex h-40 flex-col items-center justify-center gap-3 text-sm text-gray-500">
            <Building2 className="h-10 w-10 text-gray-300" />
            {search ? "Tidak ada organisasi yang cocok." : "Belum ada organisasi terdaftar."}
          </div>
        ) : (
          <div className="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm">
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Kode</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Nama</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Nama Singkat</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Website</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Status</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500">Dibuat</th>
                  <th className="px-4 py-3 text-right font-medium text-gray-500">Aksi</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filtered.map((org) => (
                  <tr key={org.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <span className="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-700">
                        {org.code}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="font-medium text-gray-900">{org.name}</div>
                      {org.address && (
                        <div className="max-w-xs truncate text-xs text-gray-400">
                          {org.address}
                        </div>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-600">{org.short_name || "—"}</td>
                    <td className="px-4 py-3">
                      {org.website ? (
                        <a
                          href={org.website}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-blue-600 hover:underline"
                        >
                          {org.website.replace(/^https?:\/\//, "")}
                        </a>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => toggleMutation.mutate(org)}
                        disabled={toggleMutation.isPending}
                        title={org.is_active ? "Klik untuk nonaktifkan" : "Klik untuk aktifkan"}
                        className="cursor-pointer"
                      >
                        <StatusBadge active={org.is_active} />
                      </button>
                    </td>
                    <td className="px-4 py-3 text-gray-500">{formatDate(org.created_at)}</td>
                    <td className="px-4 py-3">
                      <div className="flex justify-end">
                        <button
                          onClick={() => {
                            setEditing(org);
                            setModalOpen(true);
                          }}
                          title="Edit"
                          className="rounded-md p-1.5 text-gray-500 hover:bg-gray-100 hover:text-gray-700"
                        >
                          <Pencil className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {modalOpen && (
        <OrgModal
          org={editing}
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
