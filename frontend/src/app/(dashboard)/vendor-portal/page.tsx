"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { Building, FileText, HardHat } from "lucide-react";
import { bupPortalService } from "@/services/kpbu.service";
import { StageBadge } from "@/features/kpbu/ui";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { vendorPortalService } from "@/services/vendor-portal.service";
import { useAuthStore } from "@/store/auth.store";
import { ChangePasswordCard } from "@/features/vendor-portal/ChangePasswordCard";
import { fmtDate, fmtRupiahShort } from "@/features/vendor-portal/ui";

export default function VendorPortalPage() {
  const user = useAuthStore((s) => s.user);
  const mustChange = user?.must_change_pwd ?? false;

  const { data: contracts = [], isLoading, isError, refetch } = useQuery({
    queryKey: ["vendor-portal", "contracts"],
    queryFn: () => vendorPortalService.myContracts().then((r) => r.data ?? []),
    enabled: !!user && !mustChange,
  });
  const { data: kpbuProjects = [] } = useQuery({
    queryKey: ["bup-portal", "projects"],
    queryFn: () => bupPortalService.list().then((r) => r.data ?? []),
    enabled: !!user && !mustChange,
  });

  return (
    <DashboardLayout title="Portal Penyedia">
      <div className="space-y-5">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Kontrak Saya</h1>
          <p className="text-sm text-gray-500">
            Laporkan kemajuan pekerjaan dan unggah dokumen. Semua kiriman diverifikasi oleh Balai sebelum dipakai.
          </p>
        </div>

        {mustChange ? (
          <ChangePasswordCard />
        ) : isLoading ? (
          <p className="rounded-xl border border-gray-100 bg-white p-6 text-center text-sm text-gray-400">Memuat kontrak...</p>
        ) : isError ? (
          <div className="rounded-xl border border-gray-100 bg-white p-6 text-center text-sm text-red-600">
            Gagal memuat kontrak. <button className="underline" onClick={() => refetch()}>Coba lagi</button>
          </div>
        ) : contracts.length === 0 ? (
          <div className="rounded-xl border border-gray-100 bg-white p-8 text-center">
            <HardHat className="mx-auto mb-2 h-8 w-8 text-gray-300" />
            <p className="text-sm text-gray-500">Belum ada kontrak yang terhubung dengan perusahaan Anda. Hubungi Balai terkait.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            {contracts.map((c) => (
              <Link
                key={c.id}
                href={`/vendor-portal/contracts/${c.id}`}
                className="block rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-colors hover:border-blue-200"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="text-xs text-gray-400">{c.contract_number}</p>
                    <h2 className="truncate font-semibold text-gray-900">{c.title}</h2>
                    <p className="truncate text-sm text-gray-600">{c.project_code} · {c.project_name}</p>
                  </div>
                  <FileText className="h-5 w-5 shrink-0 text-blue-500" />
                </div>
                <div className="mt-3 grid grid-cols-3 gap-2 text-xs">
                  <div>
                    <p className="text-gray-400">Nilai</p>
                    <p className="font-medium text-gray-800">{fmtRupiahShort(c.contract_value)}</p>
                  </div>
                  <div>
                    <p className="text-gray-400">Periode</p>
                    <p className="font-medium text-gray-800">{fmtDate(c.start_date)} – {fmtDate(c.end_date)}</p>
                  </div>
                  <div>
                    <p className="text-gray-400">Progres terverifikasi</p>
                    <p className="font-medium text-gray-800">
                      {c.latest_verified_progress_pct != null ? `${c.latest_verified_progress_pct.toFixed(2)}%` : "—"}
                    </p>
                  </div>
                </div>
                <div className="mt-3 flex flex-wrap gap-2 text-xs">
                  {c.pending_reports + c.pending_documents > 0 && (
                    <span className="rounded-full bg-amber-50 px-2 py-0.5 text-amber-800">
                      {c.pending_reports + c.pending_documents} menunggu verifikasi
                    </span>
                  )}
                  {c.rejected_submissions > 0 && (
                    <span className="rounded-full bg-red-50 px-2 py-0.5 text-red-700">{c.rejected_submissions} ditolak</span>
                  )}
                </div>
              </Link>
            ))}
          </div>
        )}

        {!mustChange && kpbuProjects.length > 0 && (
          <div className="space-y-3">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">Proyek KPBU</h2>
              <p className="text-sm text-gray-500">Proyek tempat perusahaan Anda tercatat sebagai Badan Usaha Pelaksana (BUP).</p>
            </div>
            <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
              {kpbuProjects.map((p) => (
                <Link key={p.id} href={`/vendor-portal/kpbu/${p.id}`} className="block rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-colors hover:border-blue-200">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="text-xs text-gray-400">{p.code}</p>
                      <h3 className="truncate font-semibold text-gray-900">{p.name}</h3>
                      <p className="truncate text-sm text-gray-600">PJPK: {p.pjpk_name || "—"}</p>
                    </div>
                    <Building className="h-5 w-5 shrink-0 text-blue-500" />
                  </div>
                  <div className="mt-3 flex items-center gap-3 text-xs text-gray-600">
                    <StageBadge stage={p.stage} />
                    <span>Investasi {fmtRupiahShort(p.investment_value)}</span>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}
