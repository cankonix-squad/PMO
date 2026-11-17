"use client";

import { useMemo, useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Copy, KeyRound, UserPlus } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { projectService } from "@/services/project.service";
import { apiErrorMessage, vendorAccountService } from "@/services/vendor-portal.service";
import { btnPrimary, btnSecondary, Field, fmtDate, inputCls, Modal } from "@/features/vendor-portal/ui";
import { cn } from "@/lib/utils";
import type { CreatedVendorAccount } from "@/types/vendor-portal";

export default function VendorAccountsPage() {
  const [vendorId, setVendorId] = useState("");
  const [search, setSearch] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [created, setCreated] = useState<CreatedVendorAccount | null>(null);

  const { data: vendors = [], isLoading: loadingVendors } = useQuery({
    queryKey: ["vendors", "all"],
    queryFn: () => projectService.listVendors({ page_size: 500 }).then((r) => r.data.data ?? []),
  });
  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return vendors.filter((v) => !q || v.name.toLowerCase().includes(q));
  }, [vendors, search]);
  const selected = vendors.find((v) => v.id === vendorId);

  const { data: accounts = [], isLoading } = useQuery({
    queryKey: ["vendor-accounts", vendorId],
    queryFn: () => vendorAccountService.list(vendorId).then((r) => r.data ?? []),
    enabled: !!vendorId,
  });

  return (
    <DashboardLayout title="Akun Penyedia">
      <div className="space-y-5">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Akun Penyedia</h1>
          <p className="text-sm text-gray-500">
            Buat akun login portal untuk kontraktor. Akun hanya bisa melihat kontrak milik penyedianya dan wajib mengganti password saat pertama login.
          </p>
        </div>

        <div className="grid grid-cols-1 gap-4 lg:grid-cols-[20rem_1fr]">
          <div className="rounded-xl border border-gray-100 bg-white p-3 shadow-sm">
            <input className={inputCls} placeholder="Cari penyedia..." value={search} onChange={(e) => setSearch(e.target.value)} aria-label="Cari penyedia" />
            <ul className="mt-2 max-h-[60vh] overflow-y-auto">
              {loadingVendors ? (
                <li className="p-3 text-sm text-gray-400">Memuat penyedia...</li>
              ) : filtered.length === 0 ? (
                <li className="p-3 text-sm text-gray-500">Belum ada penyedia. Tambahkan lewat panel Kontrak di detail proyek.</li>
              ) : (
                filtered.map((v) => (
                  <li key={v.id}>
                    <button
                      onClick={() => setVendorId(v.id)}
                      className={cn("w-full rounded-md px-3 py-2 text-left text-sm hover:bg-gray-50", v.id === vendorId && "bg-blue-50 text-blue-800")}
                    >
                      <span className="block truncate font-medium">{v.name}</span>
                      <span className="text-xs text-gray-400">{v.type === "CONSULTANT" ? "Konsultan" : "Kontraktor"}</span>
                    </button>
                  </li>
                ))
              )}
            </ul>
          </div>

          <div className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
            {!selected ? (
              <div className="py-10 text-center">
                <KeyRound className="mx-auto mb-2 h-8 w-8 text-gray-300" />
                <p className="text-sm text-gray-500">Pilih penyedia untuk melihat atau membuat akun.</p>
              </div>
            ) : (
              <>
                <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
                  <h2 className="font-semibold text-gray-900">{selected.name}</h2>
                  <button className={btnPrimary} onClick={() => setShowForm(true)}>
                    <UserPlus className="h-4 w-4" /> Buat Akun
                  </button>
                </div>
                {isLoading ? (
                  <p className="text-sm text-gray-400">Memuat akun...</p>
                ) : accounts.length === 0 ? (
                  <p className="text-sm text-gray-500">Belum ada akun untuk penyedia ini.</p>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="w-full min-w-[560px] text-sm">
                      <thead className="border-b border-gray-100 text-left text-xs uppercase text-gray-500">
                        <tr>
                          <th className="py-2 pr-3">Nama</th>
                          <th className="py-2 pr-3">Email</th>
                          <th className="py-2 pr-3">Status</th>
                          <th className="py-2">Login terakhir</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-50">
                        {accounts.map((a) => (
                          <tr key={a.id}>
                            <td className="py-2 pr-3 font-medium text-gray-800">{a.first_name} {a.last_name}</td>
                            <td className="py-2 pr-3 text-gray-600">{a.email}</td>
                            <td className="py-2 pr-3 text-xs">
                              {!a.is_active ? (
                                <span className="text-gray-400">Nonaktif</span>
                              ) : a.must_change_pwd ? (
                                <span className="text-amber-700">Belum ganti password</span>
                              ) : (
                                <span className="text-green-700">Aktif</span>
                              )}
                            </td>
                            <td className="py-2 text-xs text-gray-600">{fmtDate(a.last_login_at)}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                    <p className="mt-3 text-xs text-gray-400">Untuk menonaktifkan akun, gunakan menu Pengguna.</p>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      </div>

      {showForm && selected && (
        <CreateAccountModal
          vendorId={selected.id}
          vendorName={selected.name}
          onClose={() => setShowForm(false)}
          onCreated={(c) => {
            setShowForm(false);
            setCreated(c);
          }}
        />
      )}
      {created && <CreatedModal created={created} onClose={() => setCreated(null)} />}
    </DashboardLayout>
  );
}

function CreateAccountModal({
  vendorId,
  vendorName,
  onClose,
  onCreated,
}: {
  vendorId: string;
  vendorName: string;
  onClose: () => void;
  onCreated: (c: CreatedVendorAccount) => void;
}) {
  const qc = useQueryClient();
  const [v, setV] = useState({ first_name: "", last_name: "", email: "", phone: "" });
  const [error, setError] = useState<string | null>(null);
  const set = (k: keyof typeof v) => (e: { target: { value: string } }) => setV((s) => ({ ...s, [k]: e.target.value }));

  const create = useMutation({
    mutationFn: () => vendorAccountService.create(vendorId, v),
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ["vendor-accounts", vendorId] });
      onCreated(res.data);
    },
    onError: (err) => setError(apiErrorMessage(err, "Gagal membuat akun.")),
  });

  const submit = (e: FormEvent) => {
    e.preventDefault();
    if (!v.first_name.trim()) return setError("Nama wajib diisi.");
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v.email.trim())) return setError("Email tidak valid.");
    setError(null);
    create.mutate();
  };

  return (
    <Modal title={`Buat Akun — ${vendorName}`} onClose={onClose}>
      <form onSubmit={submit} className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <Field label="Nama depan" required><input className={inputCls} value={v.first_name} onChange={set("first_name")} maxLength={100} /></Field>
        <Field label="Nama belakang"><input className={inputCls} value={v.last_name} onChange={set("last_name")} maxLength={100} /></Field>
        <Field label="Email" required className="sm:col-span-2"><input className={inputCls} type="email" value={v.email} onChange={set("email")} maxLength={200} /></Field>
        <Field label="Telepon" className="sm:col-span-2"><input className={inputCls} value={v.phone} onChange={set("phone")} maxLength={50} /></Field>
        {error && <p className="text-sm text-red-600 sm:col-span-2">{error}</p>}
        <div className="flex justify-end gap-2 sm:col-span-2">
          <button type="button" className={btnSecondary} onClick={onClose}>Batal</button>
          <button type="submit" className={btnPrimary} disabled={create.isPending}>{create.isPending ? "Membuat..." : "Buat Akun"}</button>
        </div>
      </form>
    </Modal>
  );
}

function CreatedModal({ created, onClose }: { created: CreatedVendorAccount; onClose: () => void }) {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(created.temporary_password);
      setCopied(true);
    } catch {
      setCopied(false);
    }
  };
  return (
    <Modal title="Akun Dibuat" onClose={onClose}>
      <div className="space-y-3 text-sm">
        <p>Akun <strong>{created.account.email}</strong> berhasil dibuat.</p>
        <div className="rounded-lg border border-amber-200 bg-amber-50 p-3">
          <p className="text-xs text-amber-800">Password sementara (hanya ditampilkan sekali):</p>
          <div className="mt-1 flex items-center gap-2">
            <code className="flex-1 select-all rounded bg-white px-2 py-1 font-mono text-base">{created.temporary_password}</code>
            <button className={btnSecondary} onClick={copy}><Copy className="h-4 w-4" /> {copied ? "Tersalin" : "Salin"}</button>
          </div>
        </div>
        <p className="text-xs text-gray-500">
          Sampaikan password ini ke penyedia melalui saluran resmi. Penyedia wajib menggantinya saat pertama login di portal.
        </p>
        <div className="flex justify-end"><button className={btnPrimary} onClick={onClose}>Selesai</button></div>
      </div>
    </Modal>
  );
}
