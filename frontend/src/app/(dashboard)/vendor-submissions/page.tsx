"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CheckCircle2, Download, FileCheck2, XCircle } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { apiErrorMessage, vendorSubmissionService } from "@/services/vendor-portal.service";
import {
  btnPrimary,
  btnSecondary,
  Field,
  fmtBytes,
  fmtDate,
  fmtPeriod,
  inputCls,
  Modal,
  StatusBadge,
  textareaCls,
} from "@/features/vendor-portal/ui";
import { cn } from "@/lib/utils";
import {
  SUBMISSION_STATUS_LABEL,
  VENDOR_DOC_CATEGORY_LABEL,
  type VendorDocument,
  type VendorProgressReport,
} from "@/types/vendor-portal";

type Target = { kind: "report"; item: VendorProgressReport } | { kind: "document"; item: VendorDocument };

export default function VendorSubmissionsPage() {
  const [status, setStatus] = useState("SUBMITTED");
  const [tab, setTab] = useState<"report" | "document">("report");
  const [target, setTarget] = useState<Target | null>(null);
  const [error, setError] = useState<string | null>(null);

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["vendor-submissions", status],
    queryFn: () => vendorSubmissionService.list(status ? { status } : {}).then((r) => r.data),
  });
  const reports = data?.progress_reports ?? [];
  const docs = data?.documents ?? [];

  const download = async (d: VendorDocument) => {
    try {
      setError(null);
      await vendorSubmissionService.downloadDocument(d);
    } catch (err) {
      setError(apiErrorMessage(err, "Gagal mengunduh dokumen."));
    }
  };

  return (
    <DashboardLayout title="Verifikasi Penyedia">
      <div className="space-y-5">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Verifikasi Kiriman Penyedia</h1>
          <p className="text-sm text-gray-500">Laporan progres dan dokumen dari kontraktor yang perlu diverifikasi Balai.</p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <div className="inline-flex rounded-lg border border-gray-200 bg-white p-0.5">
            {([["report", `Laporan Progres (${reports.length})`], ["document", `Dokumen (${docs.length})`]] as const).map(([k, l]) => (
              <button key={k} onClick={() => setTab(k)} className={cn("rounded-md px-3 py-1.5 text-sm", tab === k ? "bg-blue-600 text-white" : "text-gray-600 hover:bg-gray-50")}>
                {l}
              </button>
            ))}
          </div>
          <select className={cn(inputCls, "w-auto")} value={status} onChange={(e) => setStatus(e.target.value)} aria-label="Filter status">
            <option value="">Semua status</option>
            {Object.entries(SUBMISSION_STATUS_LABEL).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
          </select>
        </div>
        {error && <p className="text-sm text-red-600">{error}</p>}

        <div className="overflow-x-auto rounded-xl border border-gray-100 bg-white shadow-sm">
          {isLoading ? (
            <p className="p-6 text-center text-sm text-gray-400">Memuat kiriman...</p>
          ) : isError ? (
            <p className="p-6 text-center text-sm text-red-600">Gagal memuat. <button className="underline" onClick={() => refetch()}>Coba lagi</button></p>
          ) : (tab === "report" ? reports.length : docs.length) === 0 ? (
            <div className="p-8 text-center">
              <FileCheck2 className="mx-auto mb-2 h-8 w-8 text-gray-300" />
              <p className="text-sm text-gray-500">Tidak ada kiriman{status === "SUBMITTED" ? " yang menunggu verifikasi" : ""}.</p>
            </div>
          ) : tab === "report" ? (
            <table className="w-full min-w-[820px] text-sm">
              <thead className="border-b border-gray-100 text-left text-xs uppercase text-gray-500">
                <tr>
                  <th className="px-4 py-3">Proyek / Kontrak</th>
                  <th className="px-4 py-3">Penyedia</th>
                  <th className="px-4 py-3">Periode</th>
                  <th className="px-4 py-3 text-right">Fisik</th>
                  <th className="px-4 py-3 text-right">Rencana</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {reports.map((r) => (
                  <tr key={r.id} className="align-top">
                    <td className="px-4 py-3">
                      <p className="font-medium text-gray-800">{r.project_name}</p>
                      <p className="text-xs text-gray-400">{r.project_code} · {r.contract_number}</p>
                      {r.notes && <p className="mt-1 max-w-xs text-xs text-gray-600">{r.notes}</p>}
                    </td>
                    <td className="px-4 py-3 text-gray-600">{r.vendor_name}</td>
                    <td className="px-4 py-3">{fmtPeriod(r.period_year, r.period_month)}</td>
                    <td className="px-4 py-3 text-right font-medium">{r.physical_progress_pct.toFixed(2)}%</td>
                    <td className="px-4 py-3 text-right text-gray-500">{r.planned_progress_pct != null ? `${r.planned_progress_pct.toFixed(2)}%` : "—"}</td>
                    <td className="px-4 py-3">
                      <StatusBadge status={r.status} />
                      {r.review_notes && <p className="mt-1 text-xs text-gray-500">{r.review_notes}</p>}
                    </td>
                    <td className="px-4 py-3 text-right">
                      {r.status === "SUBMITTED" && <button className={btnSecondary} onClick={() => setTarget({ kind: "report", item: r })}>Periksa</button>}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <table className="w-full min-w-[820px] text-sm">
              <thead className="border-b border-gray-100 text-left text-xs uppercase text-gray-500">
                <tr>
                  <th className="px-4 py-3">Dokumen</th>
                  <th className="px-4 py-3">Proyek / Kontrak</th>
                  <th className="px-4 py-3">Penyedia</th>
                  <th className="px-4 py-3">Dikirim</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {docs.map((d) => (
                  <tr key={d.id} className="align-top">
                    <td className="px-4 py-3">
                      <button className="text-left font-medium text-blue-700 hover:underline" onClick={() => download(d)}>{d.name}</button>
                      <p className="text-xs text-gray-400">
                        {VENDOR_DOC_CATEGORY_LABEL[d.category] ?? d.category} · {fmtPeriod(d.period_year, d.period_month)} · {fmtBytes(d.file_size)}
                      </p>
                    </td>
                    <td className="px-4 py-3">
                      <p className="text-gray-800">{d.project_name}</p>
                      <p className="text-xs text-gray-400">{d.project_code} · {d.contract_number}</p>
                    </td>
                    <td className="px-4 py-3 text-gray-600">{d.vendor_name}</td>
                    <td className="px-4 py-3 text-xs text-gray-600">{fmtDate(d.submitted_at)}</td>
                    <td className="px-4 py-3">
                      <StatusBadge status={d.status} />
                      {d.review_notes && <p className="mt-1 text-xs text-gray-500">{d.review_notes}</p>}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex justify-end gap-2">
                        <button className="rounded p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700" onClick={() => download(d)} aria-label={`Unduh ${d.name}`}>
                          <Download className="h-4 w-4" />
                        </button>
                        {d.status === "SUBMITTED" && <button className={btnSecondary} onClick={() => setTarget({ kind: "document", item: d })}>Periksa</button>}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
      {target && <ReviewModal target={target} onClose={() => setTarget(null)} />}
    </DashboardLayout>
  );
}

function ReviewModal({ target, onClose }: { target: Target; onClose: () => void }) {
  const qc = useQueryClient();
  const [notes, setNotes] = useState("");
  const [error, setError] = useState<string | null>(null);

  const review = useMutation({
    mutationFn: async (decision: "VERIFIED" | "REJECTED"): Promise<void> => {
      if (target.kind === "report") {
        await vendorSubmissionService.reviewReport(target.item.id, { decision, notes });
      } else {
        await vendorSubmissionService.reviewDocument(target.item.id, { decision, notes });
      }
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["vendor-submissions"] });
      onClose();
    },
    onError: (err) => setError(apiErrorMessage(err, "Gagal menyimpan verifikasi.")),
  });

  const reject = () => {
    if (!notes.trim()) return setError("Alasan penolakan wajib diisi agar penyedia bisa memperbaiki.");
    review.mutate("REJECTED");
  };

  const item = target.item;
  return (
    <Modal title={target.kind === "report" ? "Verifikasi Laporan Progres" : "Verifikasi Dokumen"} onClose={onClose}>
      <div className="space-y-4">
        <dl className="grid grid-cols-2 gap-3 text-sm">
          <div className="col-span-2"><dt className="text-xs text-gray-500">Proyek</dt><dd className="font-medium">{item.project_code} · {item.project_name}</dd></div>
          <div><dt className="text-xs text-gray-500">Kontrak</dt><dd>{item.contract_number}</dd></div>
          <div><dt className="text-xs text-gray-500">Penyedia</dt><dd>{item.vendor_name}</dd></div>
          <div className="col-span-2"><dt className="text-xs text-gray-500">PPK kontrak</dt><dd>{item.ppk_name || "Belum ditetapkan — semua verifikator berwenang dapat memverifikasi"}</dd></div>
          {target.kind === "report" ? (
            <>
              <div><dt className="text-xs text-gray-500">Periode</dt><dd>{fmtPeriod(target.item.period_year, target.item.period_month)}</dd></div>
              <div><dt className="text-xs text-gray-500">Progres fisik / rencana</dt><dd>{target.item.physical_progress_pct.toFixed(2)}% / {target.item.planned_progress_pct != null ? `${target.item.planned_progress_pct.toFixed(2)}%` : "—"}</dd></div>
            </>
          ) : (
            <>
              <div><dt className="text-xs text-gray-500">Dokumen</dt><dd>{target.item.name}</dd></div>
              <div><dt className="text-xs text-gray-500">Jenis</dt><dd>{VENDOR_DOC_CATEGORY_LABEL[target.item.category] ?? target.item.category}</dd></div>
            </>
          )}
          {item.notes && <div className="col-span-2"><dt className="text-xs text-gray-500">Catatan penyedia</dt><dd className="whitespace-pre-line">{item.notes}</dd></div>}
        </dl>
        {target.kind === "report" && (
          <p className="rounded-lg bg-blue-50 p-3 text-xs text-blue-800">
            Bila diverifikasi, progres ini otomatis mengisi laporan periodik proyek bulan tersebut (progres fisik, rata-rata tertimbang nilai
            kontrak) dan memperbarui progres proyek. Data keuangan di laporan periodik tidak diubah.
          </p>
        )}
        <Field label="Catatan verifikasi (wajib bila ditolak)">
          <textarea className={textareaCls} rows={3} value={notes} onChange={(e) => setNotes(e.target.value)} />
        </Field>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex flex-wrap justify-end gap-2">
          <button className={btnSecondary} onClick={onClose}>Batal</button>
          <button className={cn(btnSecondary, "text-red-700")} onClick={reject} disabled={review.isPending}>
            <XCircle className="h-4 w-4" /> Tolak
          </button>
          <button className={btnPrimary} onClick={() => review.mutate("VERIFIED")} disabled={review.isPending}>
            <CheckCircle2 className="h-4 w-4" /> Verifikasi
          </button>
        </div>
      </div>
    </Modal>
  );
}
