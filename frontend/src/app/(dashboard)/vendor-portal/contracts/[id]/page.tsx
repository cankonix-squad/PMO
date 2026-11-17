"use client";

import { useRef, useState, type FormEvent } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, Download, Upload } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { apiErrorMessage, vendorPortalService } from "@/services/vendor-portal.service";
import { useAuthStore } from "@/store/auth.store";
import { ChangePasswordCard } from "@/features/vendor-portal/ChangePasswordCard";
import {
  btnPrimary,
  Field,
  fmtBytes,
  fmtDate,
  fmtPeriod,
  fmtRupiahShort,
  inputCls,
  MONTH_OPTIONS,
  StatusBadge,
  textareaCls,
} from "@/features/vendor-portal/ui";
import { VENDOR_DOC_CATEGORY_LABEL, type VendorDocCategory, type VendorDocument } from "@/types/vendor-portal";

const now = new Date();
const YEARS = [now.getFullYear(), now.getFullYear() - 1, now.getFullYear() - 2];

export default function VendorContractPage() {
  const { id } = useParams<{ id: string }>();
  const user = useAuthStore((s) => s.user);
  const mustChange = user?.must_change_pwd ?? false;

  const { data: contract, isLoading, isError } = useQuery({
    queryKey: ["vendor-portal", "contract", id],
    queryFn: () => vendorPortalService.myContract(id).then((r) => r.data),
    enabled: !!id && !!user && !mustChange,
  });

  return (
    <DashboardLayout title="Portal Penyedia">
      <div className="space-y-5">
        <Link href="/vendor-portal" className="inline-flex items-center gap-1 text-sm text-blue-700 hover:underline">
          <ArrowLeft className="h-4 w-4" /> Kontrak Saya
        </Link>
        {mustChange ? (
          <ChangePasswordCard />
        ) : isLoading ? (
          <p className="rounded-xl border border-gray-100 bg-white p-6 text-center text-sm text-gray-400">Memuat kontrak...</p>
        ) : isError || !contract ? (
          <p className="rounded-xl border border-gray-100 bg-white p-6 text-center text-sm text-red-600">Kontrak tidak ditemukan.</p>
        ) : (
          <>
            <div className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
              <p className="text-xs text-gray-400">{contract.contract_number}</p>
              <h1 className="text-xl font-bold text-gray-900">{contract.title}</h1>
              <p className="text-sm text-gray-600">{contract.project_code} · {contract.project_name}</p>
              <div className="mt-3 grid grid-cols-2 gap-3 text-sm md:grid-cols-4">
                <Info label="Nilai Kontrak" value={fmtRupiahShort(contract.contract_value)} />
                <Info label="Mulai" value={fmtDate(contract.start_date)} />
                <Info label="Selesai" value={fmtDate(contract.end_date)} />
                <Info
                  label="Progres Terverifikasi"
                  value={contract.latest_verified_progress_pct != null ? `${contract.latest_verified_progress_pct.toFixed(2)}%` : "—"}
                />
              </div>
            </div>
            <ProgressSection contractId={contract.id} />
            <DocumentSection contractId={contract.id} />
          </>
        )}
      </div>
    </DashboardLayout>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0">
      <p className="text-xs text-gray-500">{label}</p>
      <p className="truncate font-medium text-gray-900">{value}</p>
    </div>
  );
}

function ProgressSection({ contractId }: { contractId: string }) {
  const qc = useQueryClient();
  const [year, setYear] = useState(now.getFullYear());
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [actual, setActual] = useState("");
  const [planned, setPlanned] = useState("");
  const [notes, setNotes] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const { data: reports = [], isLoading } = useQuery({
    queryKey: ["vendor-portal", "reports", contractId],
    queryFn: () => vendorPortalService.listReports(contractId).then((r) => r.data ?? []),
  });

  const create = useMutation({
    mutationFn: () =>
      vendorPortalService.createReport(contractId, {
        period_year: year,
        period_month: month,
        physical_progress_pct: Number(actual),
        planned_progress_pct: planned === "" ? null : Number(planned),
        notes,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["vendor-portal"] });
      setActual("");
      setPlanned("");
      setNotes("");
      setSuccess("Laporan terkirim dan menunggu verifikasi Balai.");
    },
    onError: (err) => setError(apiErrorMessage(err, "Gagal mengirim laporan.")),
  });

  const submit = (e: FormEvent) => {
    e.preventDefault();
    setSuccess(null);
    const a = Number(actual);
    if (actual === "" || !Number.isFinite(a) || a < 0 || a > 100) return setError("Progres fisik harus 0–100%.");
    const p = Number(planned);
    if (planned !== "" && (!Number.isFinite(p) || p < 0 || p > 100)) return setError("Rencana progres harus 0–100%.");
    if (year === now.getFullYear() && month > now.getMonth() + 1) return setError("Periode tidak boleh di masa depan.");
    setError(null);
    create.mutate();
  };

  return (
    <div className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
      <h2 className="mb-3 text-sm font-semibold text-gray-700">Laporan Progres Bulanan</h2>
      <form onSubmit={submit} className="mb-5 grid grid-cols-2 gap-3 rounded-lg bg-gray-50 p-3 md:grid-cols-5 md:items-end">
        <Field label="Bulan" required>
          <select className={inputCls} value={month} onChange={(e) => setMonth(Number(e.target.value))}>
            {MONTH_OPTIONS.map((m) => <option key={m.value} value={m.value}>{m.label}</option>)}
          </select>
        </Field>
        <Field label="Tahun" required>
          <select className={inputCls} value={year} onChange={(e) => setYear(Number(e.target.value))}>
            {YEARS.map((y) => <option key={y} value={y}>{y}</option>)}
          </select>
        </Field>
        <Field label="Progres Fisik (%)" required>
          <input className={inputCls} type="number" min={0} max={100} step="0.01" value={actual} onChange={(e) => setActual(e.target.value)} placeholder="cth. 42.5" />
        </Field>
        <Field label="Rencana (%)">
          <input className={inputCls} type="number" min={0} max={100} step="0.01" value={planned} onChange={(e) => setPlanned(e.target.value)} placeholder="sesuai kurva S" />
        </Field>
        <button type="submit" className={btnPrimary} disabled={create.isPending}>
          {create.isPending ? "Mengirim..." : "Kirim Laporan"}
        </button>
        <Field label="Catatan" className="col-span-2 md:col-span-5">
          <textarea className={textareaCls} rows={2} value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="Pekerjaan yang diselesaikan, kendala, dll." />
        </Field>
        {error && <p className="col-span-2 text-sm text-red-600 md:col-span-5">{error}</p>}
        {success && <p className="col-span-2 text-sm text-green-700 md:col-span-5">{success}</p>}
      </form>

      {isLoading ? (
        <p className="text-sm text-gray-400">Memuat laporan...</p>
      ) : reports.length === 0 ? (
        <p className="text-sm text-gray-500">Belum ada laporan progres.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full min-w-[560px] text-sm">
            <thead className="border-b border-gray-100 text-left text-xs uppercase text-gray-500">
              <tr>
                <th className="py-2 pr-3">Periode</th>
                <th className="py-2 pr-3 text-right">Fisik</th>
                <th className="py-2 pr-3 text-right">Rencana</th>
                <th className="py-2 pr-3">Status</th>
                <th className="py-2">Catatan Balai</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {reports.map((r) => (
                <tr key={r.id}>
                  <td className="py-2 pr-3 font-medium text-gray-800">{fmtPeriod(r.period_year, r.period_month)}</td>
                  <td className="py-2 pr-3 text-right">{r.physical_progress_pct.toFixed(2)}%</td>
                  <td className="py-2 pr-3 text-right text-gray-500">{r.planned_progress_pct != null ? `${r.planned_progress_pct.toFixed(2)}%` : "—"}</td>
                  <td className="py-2 pr-3"><StatusBadge status={r.status} /></td>
                  <td className="py-2 text-xs text-gray-600">{r.review_notes || "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function DocumentSection({ contractId }: { contractId: string }) {
  const qc = useQueryClient();
  const fileRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [category, setCategory] = useState<VendorDocCategory>("PROGRESS_REPORT");
  const [name, setName] = useState("");
  const [withPeriod, setWithPeriod] = useState(true);
  const [year, setYear] = useState(now.getFullYear());
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [notes, setNotes] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const { data: docs = [], isLoading } = useQuery({
    queryKey: ["vendor-portal", "documents", contractId],
    queryFn: () => vendorPortalService.listDocuments(contractId).then((r) => r.data ?? []),
  });

  const upload = useMutation({
    mutationFn: (f: File) =>
      vendorPortalService.uploadDocument(contractId, {
        file: f,
        category,
        name,
        notes,
        ...(withPeriod ? { period_year: year, period_month: month } : {}),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["vendor-portal"] });
      setFile(null);
      setName("");
      setNotes("");
      if (fileRef.current) fileRef.current.value = "";
      setSuccess("Dokumen terkirim dan menunggu verifikasi Balai.");
    },
    onError: (err) => setError(apiErrorMessage(err, "Gagal mengunggah dokumen.")),
  });

  const submit = (e: FormEvent) => {
    e.preventDefault();
    setSuccess(null);
    if (!file) return setError("Pilih file terlebih dahulu.");
    if (file.size > 20 * 1024 * 1024) return setError("Ukuran file maksimal 20 MB.");
    setError(null);
    upload.mutate(file);
  };

  const download = async (d: VendorDocument) => {
    try {
      await vendorPortalService.downloadOwnDocument(d);
    } catch (err) {
      setError(apiErrorMessage(err, "Gagal mengunduh dokumen."));
    }
  };

  return (
    <div className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
      <h2 className="mb-3 text-sm font-semibold text-gray-700">Dokumen</h2>
      <form onSubmit={submit} className="mb-5 grid grid-cols-1 gap-3 rounded-lg bg-gray-50 p-3 md:grid-cols-2">
        <Field label="File" required>
          <input
            ref={fileRef}
            type="file"
            className="block w-full text-sm text-gray-700 file:mr-3 file:rounded-md file:border-0 file:bg-blue-50 file:px-3 file:py-1.5 file:text-sm file:text-blue-700"
            accept=".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.png,.jpg,.jpeg,.txt,.csv"
            onChange={(e) => setFile(e.target.files?.[0] ?? null)}
          />
          <span className="text-xs text-gray-400">PDF, Office, gambar, CSV · maks 20 MB</span>
        </Field>
        <Field label="Jenis Dokumen" required>
          <select className={inputCls} value={category} onChange={(e) => setCategory(e.target.value as VendorDocCategory)}>
            {Object.entries(VENDOR_DOC_CATEGORY_LABEL).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
          </select>
        </Field>
        <Field label="Judul Dokumen">
          <input className={inputCls} value={name} onChange={(e) => setName(e.target.value)} placeholder="kosongkan untuk memakai nama file" maxLength={500} />
        </Field>
        <div className="grid grid-cols-[auto_1fr_1fr] items-end gap-2">
          <label className="flex h-9 items-center gap-2 text-sm text-gray-700">
            <input type="checkbox" checked={withPeriod} onChange={(e) => setWithPeriod(e.target.checked)} /> Periode
          </label>
          <select className={inputCls} value={month} disabled={!withPeriod} onChange={(e) => setMonth(Number(e.target.value))} aria-label="Bulan dokumen">
            {MONTH_OPTIONS.map((m) => <option key={m.value} value={m.value}>{m.label}</option>)}
          </select>
          <select className={inputCls} value={year} disabled={!withPeriod} onChange={(e) => setYear(Number(e.target.value))} aria-label="Tahun dokumen">
            {YEARS.map((y) => <option key={y} value={y}>{y}</option>)}
          </select>
        </div>
        <Field label="Catatan" className="md:col-span-2">
          <textarea className={textareaCls} rows={2} value={notes} onChange={(e) => setNotes(e.target.value)} />
        </Field>
        {error && <p className="text-sm text-red-600 md:col-span-2">{error}</p>}
        {success && <p className="text-sm text-green-700 md:col-span-2">{success}</p>}
        <div className="md:col-span-2">
          <button type="submit" className={btnPrimary} disabled={upload.isPending}>
            <Upload className="h-4 w-4" /> {upload.isPending ? "Mengunggah..." : "Unggah Dokumen"}
          </button>
        </div>
      </form>

      {isLoading ? (
        <p className="text-sm text-gray-400">Memuat dokumen...</p>
      ) : docs.length === 0 ? (
        <p className="text-sm text-gray-500">Belum ada dokumen.</p>
      ) : (
        <ul className="divide-y divide-gray-50">
          {docs.map((d) => (
            <li key={d.id} className="flex flex-wrap items-start justify-between gap-2 py-2.5">
              <div className="min-w-0">
                <p className="truncate text-sm font-medium text-gray-800">{d.name}</p>
                <p className="text-xs text-gray-500">
                  {VENDOR_DOC_CATEGORY_LABEL[d.category] ?? d.category} · {fmtPeriod(d.period_year, d.period_month)} · {fmtBytes(d.file_size)} · dikirim {fmtDate(d.submitted_at)}
                </p>
                {d.review_notes && <p className="text-xs text-gray-600">Catatan Balai: {d.review_notes}</p>}
              </div>
              <div className="flex items-center gap-2">
                <StatusBadge status={d.status} />
                <button className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-700" onClick={() => download(d)} aria-label={`Unduh ${d.name}`}>
                  <Download className="h-4 w-4" />
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
