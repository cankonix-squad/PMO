"use client";

import { useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, Pencil, Trash2 } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { kpbuErrorMessage, kpbuService } from "@/services/kpbu.service";
import { useAuthStore } from "@/store/auth.store";
import { ProjectFormModal } from "@/features/kpbu/ProjectFormModal";
import { StagePanel } from "@/features/kpbu/StagePanel";
import { PartiesPanel } from "@/features/kpbu/PartiesPanel";
import { DocumentsPanel, FinancialCloseBanner } from "@/features/kpbu/DocumentsPanel";
import { GuaranteesPanel } from "@/features/kpbu/GuaranteesPanel";
import { FinancePanel } from "@/features/kpbu/FinancePanel";
import { btnSecondary, fmtDate, fmtRupiahShort, StageBadge } from "@/features/kpbu/ui";
import { KPBU_INITIATIVE_LABEL, KPBU_PROCUREMENT_LABEL, KPBU_RETURN_SCHEME_LABEL, KPBU_SECTOR_LABEL } from "@/types/kpbu";

const WRITE_ROLES = ["SUPER_ADMIN", "ADMIN", "PMO", "PROJECT_MANAGER"];
const DELETE_ROLES = ["SUPER_ADMIN", "ADMIN", "PMO"];

export default function KpbuDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const qc = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const roles = user?.roles ?? [];
  const canWrite = roles.some((r) => WRITE_ROLES.includes(r));
  const canDelete = roles.some((r) => DELETE_ROLES.includes(r));
  const [editing, setEditing] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["kpbu", "detail", id],
    queryFn: () => kpbuService.get(id).then((r) => r.data),
    enabled: !!id,
  });

  const del = useMutation({
    mutationFn: () => kpbuService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["kpbu"] });
      router.push("/kpbu");
    },
    onError: (err) => setDeleteError(kpbuErrorMessage(err, "Gagal menghapus proyek.")),
  });

  const p = data?.project;

  return (
    <DashboardLayout title="Detail Proyek KPBU">
      <div className="space-y-5">
        <Link href="/kpbu" className="inline-flex items-center gap-1 text-sm text-blue-700 hover:underline">
          <ArrowLeft className="h-4 w-4" /> Portofolio KPBU
        </Link>

        {isLoading ? (
          <p className="rounded-xl border border-gray-100 bg-white p-6 text-center text-sm text-gray-400">Memuat proyek...</p>
        ) : isError || !data || !p ? (
          <div className="rounded-xl border border-gray-100 bg-white p-6 text-center text-sm text-red-600">
            Proyek tidak ditemukan atau gagal dimuat. <button className="underline" onClick={() => refetch()}>Coba lagi</button>
          </div>
        ) : (
          <>
            <div className="flex flex-wrap items-start justify-between gap-3 rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
              <div className="min-w-0">
                <p className="text-xs text-gray-400">{p.code}</p>
                <h1 className="text-xl font-bold text-gray-900">{p.name}</h1>
                <div className="mt-2 flex flex-wrap items-center gap-2">
                  <StageBadge stage={p.stage} />
                  {p.stage_is_delayed && <span className="text-xs font-medium text-red-600">Tahap terlambat {p.stage_delay_days} hari</span>}
                </div>
                {p.description && <p className="mt-3 max-w-3xl whitespace-pre-line text-sm text-gray-600">{p.description}</p>}
              </div>
              <div className="flex items-center gap-2">
                {canWrite && <button className={btnSecondary} onClick={() => setEditing(true)}><Pencil className="h-4 w-4" /> Edit</button>}
                {canDelete && !confirmDelete && (
                  <button className={btnSecondary} onClick={() => setConfirmDelete(true)}><Trash2 className="h-4 w-4 text-red-600" /> Hapus</button>
                )}
                {confirmDelete && (
                  <div className="flex items-center gap-2 rounded-lg bg-red-50 px-3 py-1.5 text-sm">
                    <span className="text-red-700">Hapus proyek ini?</span>
                    <button className="font-medium text-red-700 underline" onClick={() => del.mutate()} disabled={del.isPending}>Ya, hapus</button>
                    <button className="text-gray-600" onClick={() => setConfirmDelete(false)}>Batal</button>
                  </div>
                )}
              </div>
              {deleteError && <p className="w-full text-sm text-red-600">{deleteError}</p>}
            </div>

            <FinancialCloseBanner status={p.financial_close_status} deadline={p.financial_close_deadline} daysLeft={p.financial_close_days_left} />

            <div className="grid grid-cols-2 gap-3 rounded-xl border border-gray-100 bg-white p-4 shadow-sm md:grid-cols-4">
              <Info label="Sektor" value={KPBU_SECTOR_LABEL[p.sector] ?? p.sector} />
              <Info label="Nilai Investasi" value={fmtRupiahShort(p.investment_value)} />
              <Info label="Skema Pengembalian" value={KPBU_RETURN_SCHEME_LABEL[p.return_scheme] ?? p.return_scheme} />
              <Info label="Prakarsa" value={KPBU_INITIATIVE_LABEL[p.initiative] ?? p.initiative} />
              <Info label="PJPK" value={p.pjpk_name || "—"} sub={p.pjpk_unit} />
              <Info label="Masa Konsesi" value={p.concession_years ? `${p.concession_years} tahun` : "—"} sub={p.concession_start ? `${fmtDate(p.concession_start)} – ${fmtDate(p.concession_end)}` : undefined} />
              <Info label="Metode Pengadaan" value={p.procurement_method ? KPBU_PROCUREMENT_LABEL[p.procurement_method] : "—"} sub={p.procurement_reason || undefined} />
              <Info label="Perjanjian KPBU Ditandatangani" value={fmtDate(p.agreement_signed_date)} sub={p.financial_close_deadline ? `Batas financial close ${fmtDate(p.financial_close_deadline)}` : undefined} />
              <Info label="Financial Close" value={p.financial_close_actual ? fmtDate(p.financial_close_actual) : "Belum"} sub={`Target ${fmtDate(p.financial_close_target)}`} />
              <Info label="COD (mulai operasi)" value={p.cod_actual ? fmtDate(p.cod_actual) : "Belum"} sub={`Target ${fmtDate(p.cod_target)}`} />
              <Info label="Balai / Unit" value={p.org_unit_name || "—"} />
              <Info label="Provinsi" value={p.province || "—"} />
              <Info label="Koordinat" value={p.latitude != null ? `${p.latitude}, ${p.longitude}` : "—"} />
            </div>

            <StagePanel detail={data} canWrite={canWrite} />
            <DocumentsPanel projectId={p.id} canWrite={canWrite} canReview={canWrite} canDelete={canWrite} />
            <GuaranteesPanel projectId={p.id} canWrite={canWrite} canDelete={canDelete} />
            <FinancePanel projectId={p.id} returnScheme={p.return_scheme} canWrite={canWrite} canDelete={canDelete} />
            <PartiesPanel projectId={p.id} parties={data.parties} canWrite={canWrite} />

            <p className="text-xs text-gray-400">
              Modul KPBU fase K1–K2: data proyek, tahapan, para pihak, checklist dokumen badan usaha, jaminan, struktur pendanaan, dukungan pemerintah, dan pembayaran ketersediaan layanan. KPI layanan menyusul di fase K3.
            </p>
          </>
        )}
      </div>
      {editing && p && <ProjectFormModal project={p} onClose={() => setEditing(false)} />}
    </DashboardLayout>
  );
}

function Info({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div className="min-w-0">
      <p className="text-xs text-gray-500">{label}</p>
      <p className="truncate text-sm font-medium text-gray-900">{value}</p>
      {sub && <p className="truncate text-xs text-gray-400">{sub}</p>}
    </div>
  );
}
