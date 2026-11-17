"use client";

import { useRef, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, Download, Plus, Upload } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { bupPortalService, kpbuErrorMessage } from "@/services/kpbu.service";
import { MONTHS } from "@/features/kpbu/FinancePanel";
import { fmtValue, MeasurementModal } from "@/features/kpbu/KpiPanel";
import { btnSecondary, fmtDate, fmtRupiahShort, StageBadge } from "@/features/kpbu/ui";
import { cn } from "@/lib/utils";
import {
  KPBU_DOC_STATUS_LABEL,
  KPBU_MEASUREMENT_STATUS_LABEL,
  KPBU_PAYMENT_STATUS_LABEL,
  type KpbuDocStatus,
  type KpbuDocument,
} from "@/types/kpbu";

const DOC_CLS: Record<KpbuDocStatus, string> = {
  MISSING: "text-gray-400",
  SUBMITTED: "text-orange-700",
  APPROVED: "text-green-700",
  REJECTED: "text-red-600",
};

function UploadButton({ projectId, requirementId }: { projectId: string; requirementId?: string }) {
  const qc = useQueryClient();
  const ref = useRef<HTMLInputElement>(null);
  const [err, setErr] = useState<string | null>(null);
  const up = useMutation({
    mutationFn: (file: File) => bupPortalService.uploadDocument(projectId, { file, requirement_id: requirementId }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["bup-portal", projectId, "checklist"] }),
    onError: (e) => setErr(kpbuErrorMessage(e, "Gagal mengunggah.")),
  });
  return (
    <span className="inline-flex flex-col items-end">
      <input
        ref={ref}
        type="file"
        className="hidden"
        onChange={(e) => {
          const f = e.target.files?.[0];
          if (f) {
            setErr(null);
            up.mutate(f);
          }
          e.target.value = "";
        }}
      />
      <button className={btnSecondary} disabled={up.isPending} onClick={() => ref.current?.click()}>
        <Upload className="h-4 w-4" /> {up.isPending ? "Mengunggah..." : "Unggah"}
      </button>
      {err && <span className="mt-1 text-xs text-red-600">{err}</span>}
    </span>
  );
}

export default function BupProjectPage() {
  const { id } = useParams<{ id: string }>();
  const qc = useQueryClient();
  const [entering, setEntering] = useState(false);
  const [dlErr, setDlErr] = useState<string | null>(null);

  const project = useQuery({ queryKey: ["bup-portal", id], queryFn: () => bupPortalService.get(id).then((r) => r.data) });
  const checklist = useQuery({ queryKey: ["bup-portal", id, "checklist"], queryFn: () => bupPortalService.checklist(id).then((r) => r.data), enabled: project.isSuccess });
  const kpis = useQuery({ queryKey: ["bup-portal", id, "kpis"], queryFn: () => bupPortalService.listKpis(id).then((r) => r.data), enabled: project.isSuccess });
  const measures = useQuery({ queryKey: ["bup-portal", id, "measures"], queryFn: () => bupPortalService.listMeasurements(id).then((r) => r.data), enabled: project.isSuccess });
  const p = project.data;
  const usesAP = p?.return_scheme === "AVAILABILITY_PAYMENT" || p?.return_scheme === "HYBRID";
  const payments = useQuery({ queryKey: ["bup-portal", id, "payments"], queryFn: () => bupPortalService.listPayments(id).then((r) => r.data), enabled: project.isSuccess && usesAP });

  const download = async (d: KpbuDocument) => {
    try {
      setDlErr(null);
      await bupPortalService.downloadDocument(id, d);
    } catch (e) {
      setDlErr(kpbuErrorMessage(e, "Gagal mengunduh dokumen."));
    }
  };

  return (
    <DashboardLayout title="Portal Badan Usaha">
      <div className="space-y-5">
        <Link href="/vendor-portal" className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-gray-800"><ArrowLeft className="h-4 w-4" /> Kembali</Link>
        {project.isLoading ? (
          <p className="text-sm text-gray-400">Memuat proyek...</p>
        ) : project.isError || !p ? (
          <p className="text-sm text-red-600">{kpbuErrorMessage(project.error, "Proyek tidak ditemukan atau tidak terhubung dengan badan usaha Anda.")}</p>
        ) : (
          <>
            <div>
              <p className="text-xs text-gray-400">{p.code} · sebagai {p.party_name}</p>
              <h1 className="text-2xl font-bold text-gray-900">{p.name}</h1>
              <div className="mt-1 flex flex-wrap items-center gap-3 text-sm text-gray-600">
                <StageBadge stage={p.stage} />
                <span>PJPK: {p.pjpk_name || "—"}</span>
                <span>Investasi {fmtRupiahShort(p.investment_value)}</span>
              </div>
            </div>

            <section className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
              <h2 className="mb-1 text-sm font-semibold text-gray-700">Checklist Dokumen</h2>
              <p className="mb-3 text-xs text-gray-400">Unggah dokumen per butir. Setiap dokumen diperiksa PJPK; dokumen ditolak dapat diunggah ulang.</p>
              {dlErr && <p className="mb-2 text-sm text-red-600">{dlErr}</p>}
              {checklist.isLoading ? (
                <p className="text-sm text-gray-400">Memuat...</p>
              ) : checklist.isError ? (
                <p className="text-sm text-red-600">Gagal memuat checklist.</p>
              ) : (
                <ul className="divide-y divide-gray-50">
                  {checklist.data?.items.map((it) => (
                    <li key={it.requirement.id} className="flex flex-wrap items-start justify-between gap-3 py-3">
                      <div className="min-w-0 flex-1">
                        <p className="text-sm text-gray-800">
                          <span className="font-semibold text-gray-500">{it.requirement.code}</span> {it.requirement.name}
                          {it.requirement.is_mandatory && <span className="text-red-500"> *</span>}
                        </p>
                        <p className={cn("text-xs", DOC_CLS[it.status])}>{KPBU_DOC_STATUS_LABEL[it.status]}</p>
                        {it.documents.map((d) => (
                          <p key={d.id} className="mt-1 text-xs text-gray-600">
                            <button className="inline-flex items-center gap-1 text-blue-700 hover:underline" onClick={() => download(d)}><Download className="h-3 w-3" />{d.name}</button>
                            {" · "}<span className={DOC_CLS[d.status]}>{KPBU_DOC_STATUS_LABEL[d.status]}</span>
                            {d.review_notes && <span className="text-gray-500"> — {d.review_notes}</span>}
                          </p>
                        ))}
                      </div>
                      {it.status !== "APPROVED" && <UploadButton projectId={id} requirementId={it.requirement.id} />}
                    </li>
                  ))}
                </ul>
              )}
              <div className="mt-3 flex items-center justify-between border-t border-gray-50 pt-3">
                <p className="text-xs text-gray-500">Dokumen lain di luar checklist</p>
                <UploadButton projectId={id} />
              </div>
            </section>

            <section className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
              <div className="mb-3 flex items-center justify-between">
                <div>
                  <h2 className="text-sm font-semibold text-gray-700">Laporan KPI Layanan</h2>
                  <p className="text-xs text-gray-400">Laporkan realisasi KPI per bulan. PJPK memverifikasi sebelum nilai dipakai.</p>
                </div>
                {!!kpis.data?.length && <button className={btnSecondary} onClick={() => setEntering(true)}><Plus className="h-4 w-4" /> Laporkan</button>}
              </div>
              {!kpis.data?.length ? (
                <p className="text-sm text-gray-400">PJPK belum menetapkan KPI untuk proyek ini.</p>
              ) : !measures.data?.length ? (
                <p className="text-sm text-gray-400">Belum ada laporan KPI.</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full min-w-[560px] text-sm">
                    <thead className="border-b border-gray-100 text-left text-xs uppercase text-gray-500">
                      <tr><th className="py-2 pr-3">Periode</th><th className="py-2 pr-3">KPI</th><th className="py-2 pr-3">Realisasi</th><th className="py-2 pr-3">Status</th></tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50">
                      {measures.data.map((m) => {
                        const k = kpis.data?.find((x) => x.id === m.kpi_id);
                        return (
                          <tr key={m.id}>
                            <td className="py-2 pr-3 text-xs">{MONTHS[m.period_month - 1]} {m.period_year}</td>
                            <td className="py-2 pr-3 text-xs"><span className="font-semibold">{m.kpi_code}</span> · {m.kpi_name}</td>
                            <td className="py-2 pr-3 text-xs">{fmtValue(m.actual_value, k?.unit ?? "")}{m.met === false && <span className="block text-red-600">di bawah target</span>}</td>
                            <td className="py-2 pr-3 text-xs">
                              {KPBU_MEASUREMENT_STATUS_LABEL[m.status]}
                              {m.review_notes && <span className="block text-gray-500">{m.review_notes}</span>}
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              )}
            </section>

            {usesAP && (
              <section className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
                <h2 className="mb-3 text-sm font-semibold text-gray-700">Pembayaran Ketersediaan Layanan</h2>
                {!payments.data?.length ? (
                  <p className="text-sm text-gray-400">Belum ada jadwal pembayaran.</p>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="w-full min-w-[560px] text-sm">
                      <thead className="border-b border-gray-100 text-left text-xs uppercase text-gray-500">
                        <tr><th className="py-2 pr-3">Periode</th><th className="py-2 pr-3">Jadwal</th><th className="py-2 pr-3">Potongan</th><th className="py-2 pr-3">Dibayar</th><th className="py-2 pr-3">Status</th></tr>
                      </thead>
                      <tbody className="divide-y divide-gray-50">
                        {payments.data.map((a) => (
                          <tr key={a.id}>
                            <td className="py-2 pr-3 text-xs">{MONTHS[a.period_month - 1]} {a.period_year}<span className="block text-gray-400">jatuh tempo {fmtDate(a.due_date)}</span></td>
                            <td className="py-2 pr-3 text-xs">{fmtRupiahShort(a.scheduled_amount)}</td>
                            <td className="py-2 pr-3 text-xs">{a.deduction_amount > 0 ? <>{fmtRupiahShort(a.deduction_amount)}<span className="block text-gray-500">{a.deduction_reason}</span></> : "—"}</td>
                            <td className="py-2 pr-3 text-xs">{fmtRupiahShort(a.paid_amount)}{a.paid_date && <span className="block text-gray-400">{fmtDate(a.paid_date)}</span>}</td>
                            <td className={cn("py-2 pr-3 text-xs", a.status === "OVERDUE" && "font-medium text-red-600", a.status === "PAID" && "text-green-700")}>{KPBU_PAYMENT_STATUS_LABEL[a.status]}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </section>
            )}
          </>
        )}
      </div>
      {entering && kpis.data && (
        <MeasurementModal
          portal
          kpis={kpis.data}
          onSubmit={(payload) => bupPortalService.createMeasurement(id, payload)}
          onDone={() => {
            qc.invalidateQueries({ queryKey: ["bup-portal", id, "measures"] });
            setEntering(false);
          }}
          onClose={() => setEntering(false)}
        />
      )}
    </DashboardLayout>
  );
}
