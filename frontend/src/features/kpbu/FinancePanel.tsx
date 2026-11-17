"use client";

import { useState, type ReactNode } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Pencil, Plus, Trash2 } from "lucide-react";
import { kpbuErrorMessage, kpbuService } from "@/services/kpbu.service";
import { cn } from "@/lib/utils";
import {
  KPBU_FINANCING_SOURCE_LABEL,
  KPBU_FINANCING_STATUS_LABEL,
  KPBU_PAYMENT_STATUS_LABEL,
  KPBU_SUPPORT_STATUS_LABEL,
  KPBU_SUPPORT_TYPE_LABEL,
  type KpbuAvailabilityPayment,
  type KpbuAvailabilityPaymentRequest,
  type KpbuFinanceSummary,
  type KpbuFinancing,
  type KpbuFinancingRequest,
  type KpbuFinancingSource,
  type KpbuGovernmentSupport,
  type KpbuGovernmentSupportRequest,
  type KpbuReturnScheme,
} from "@/types/kpbu";
import { btnPrimary, btnSecondary, Field, fmtDate, fmtRupiahShort, inputCls, Modal, textareaCls, toDateInput } from "./ui";

// ── Generic record section (table + create/edit modal) ──────────────────────

export type FieldDef = {
  key: string;
  label: string;
  kind: "text" | "number" | "date" | "select" | "textarea";
  options?: Record<string, string>;
  required?: boolean;
  wide?: boolean;
};
export type Values = Record<string, string>;
export type Column<T> = { label: string; render: (row: T) => ReactNode; className?: string };

interface SectionProps<T extends { id: string }> {
  title: string;
  queryKey: string;
  projectId: string;
  canWrite: boolean;
  canDelete: boolean;
  empty: string;
  note?: ReactNode;
  columns: Column<T>[];
  fields: FieldDef[];
  list: (pid: string) => Promise<T[]>;
  save: (pid: string, id: string | null, v: Values) => Promise<unknown>;
  remove: (pid: string, id: string) => Promise<void>;
  toValues: (row: T | null) => Values;
  validate?: (v: Values) => string | null;
}

export function RecordSection<T extends { id: string }>(props: SectionProps<T>) {
  const qc = useQueryClient();
  const key = ["kpbu", props.queryKey, props.projectId];
  const { data, isLoading, isError, refetch } = useQuery({ queryKey: key, queryFn: () => props.list(props.projectId) });
  const [editing, setEditing] = useState<T | "new" | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const invalidate = () => {
    qc.invalidateQueries({ queryKey: key });
    qc.invalidateQueries({ queryKey: ["kpbu", "finance-summary", props.projectId] });
  };
  const del = useMutation({
    mutationFn: (id: string) => props.remove(props.projectId, id),
    onSuccess: invalidate,
    onError: (e) => setErr(kpbuErrorMessage(e, "Gagal menghapus.")),
  });
  const showActions = props.canWrite || props.canDelete;

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <h3 className="text-sm font-semibold text-gray-700">{props.title}</h3>
        {props.canWrite && <button className={btnSecondary} onClick={() => setEditing("new")}><Plus className="h-4 w-4" /> Tambah</button>}
      </div>
      {isLoading ? (
        <p className="py-3 text-center text-sm text-gray-400">Memuat...</p>
      ) : isError ? (
        <p className="py-3 text-center text-sm text-red-600">Gagal memuat. <button className="underline" onClick={() => refetch()}>Coba lagi</button></p>
      ) : !data?.length ? (
        <p className="py-3 text-center text-sm text-gray-400">{props.empty}</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full min-w-[680px] text-sm">
            <thead className="border-b border-gray-100 text-left text-xs uppercase text-gray-500">
              <tr>
                {props.columns.map((c) => <th key={c.label} className={cn("py-2 pr-3", c.className)}>{c.label}</th>)}
                {showActions && <th className="py-2" />}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {data.map((row) => (
                <tr key={row.id} className="align-top">
                  {props.columns.map((c) => <td key={c.label} className={cn("py-2 pr-3 text-xs text-gray-700", c.className)}>{c.render(row)}</td>)}
                  {showActions && (
                    <td className="whitespace-nowrap py-2 text-right">
                      {props.canWrite && <button className="rounded p-1 text-gray-400 hover:bg-gray-100" onClick={() => setEditing(row)} aria-label="Edit"><Pencil className="h-3.5 w-3.5" /></button>}
                      {props.canDelete && <button className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-red-600" onClick={() => del.mutate(row.id)} aria-label="Hapus"><Trash2 className="h-3.5 w-3.5" /></button>}
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {err && <p className="mt-2 text-sm text-red-600">{err}</p>}
      {props.note && <p className="mt-2 text-xs text-gray-400">{props.note}</p>}
      {editing && (
        <RecordModal
          title={`${editing === "new" ? "Tambah" : "Edit"} ${props.title}`}
          fields={props.fields}
          initial={props.toValues(editing === "new" ? null : editing)}
          validate={props.validate}
          onSave={(v) => props.save(props.projectId, editing === "new" ? null : editing.id, v)}
          onDone={() => {
            invalidate();
            setEditing(null);
          }}
          onClose={() => setEditing(null)}
        />
      )}
    </div>
  );
}

function RecordModal(props: {
  title: string;
  fields: FieldDef[];
  initial: Values;
  validate?: (v: Values) => string | null;
  onSave: (v: Values) => Promise<unknown>;
  onDone: () => void;
  onClose: () => void;
}) {
  const [v, setV] = useState<Values>(props.initial);
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: () => props.onSave(v),
    onSuccess: props.onDone,
    onError: (e) => setError(kpbuErrorMessage(e, "Gagal menyimpan.")),
  });
  const set = (k: string) => (e: { target: { value: string } }) => setV((s) => ({ ...s, [k]: e.target.value }));

  return (
    <Modal title={props.title} onClose={props.onClose} wide>
      <form
        className="grid grid-cols-1 gap-3 sm:grid-cols-2"
        onSubmit={(e) => {
          e.preventDefault();
          const missing = props.fields.find((f) => f.required && !v[f.key]?.trim());
          if (missing) return setError(`${missing.label} wajib diisi.`);
          const bad = props.fields.find((f) => f.kind === "number" && v[f.key] !== "" && (!Number.isFinite(Number(v[f.key])) || Number(v[f.key]) < 0));
          if (bad) return setError(`${bad.label} harus angka dan tidak negatif.`);
          const msg = props.validate?.(v);
          if (msg) return setError(msg);
          setError(null);
          save.mutate();
        }}
      >
        {props.fields.map((f) => (
          <Field key={f.key} label={f.label} required={f.required} className={f.wide || f.kind === "textarea" ? "sm:col-span-2" : undefined}>
            {f.kind === "select" ? (
              <select className={inputCls} value={v[f.key]} onChange={set(f.key)}>
                {Object.entries(f.options ?? {}).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
              </select>
            ) : f.kind === "textarea" ? (
              <textarea className={textareaCls} rows={2} value={v[f.key]} onChange={set(f.key)} />
            ) : (
              <input className={inputCls} type={f.kind} min={f.kind === "number" ? 0 : undefined} step={f.kind === "number" ? "any" : undefined} value={v[f.key]} onChange={set(f.key)} />
            )}
          </Field>
        ))}
        {error && <p className="text-sm text-red-600 sm:col-span-2">{error}</p>}
        <div className="flex justify-end gap-2 sm:col-span-2">
          <button type="button" className={btnSecondary} onClick={props.onClose}>Batal</button>
          <button type="submit" className={btnPrimary} disabled={save.isPending}>{save.isPending ? "Menyimpan..." : "Simpan"}</button>
        </div>
      </form>
    </Modal>
  );
}

export const num = (s: string) => (s === "" ? 0 : Number(s));
const dateOrNull = (s: string) => s || null;
const str = (n: number | null | undefined) => (n == null ? "" : String(n));
export const MONTHS = ["Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"];
export const MONTH_OPTIONS = Object.fromEntries(MONTHS.map((m, i) => [String(i + 1), m]));

// ── Summary ─────────────────────────────────────────────────────────────────

function Stat({ label, value, sub, warn }: { label: string; value: string; sub?: string; warn?: boolean }) {
  return (
    <div className="rounded-lg border border-gray-100 p-3">
      <p className="text-xs text-gray-500">{label}</p>
      <p className={cn("text-base font-semibold", warn ? "text-red-600" : "text-gray-900")}>{value}</p>
      {sub && <p className="text-xs text-gray-400">{sub}</p>}
    </div>
  );
}

function SummaryCards({ s, usesAP }: { s: KpbuFinanceSummary; usesAP: boolean }) {
  const pct = (v: number | null) => (v == null ? "—" : `${v.toLocaleString("id-ID")}%`);
  return (
    <div className="grid grid-cols-2 gap-2 lg:grid-cols-4">
      <Stat label="Total pendanaan" value={fmtRupiahShort(s.financing_total)} sub={`${pct(s.funding_coverage_pct)} dari nilai investasi`} warn={s.funding_coverage_pct != null && s.funding_coverage_pct < 100} />
      <Stat label="Ekuitas : Utang" value={s.debt_to_equity == null ? "—" : `1 : ${s.debt_to_equity.toLocaleString("id-ID")}`} sub={`Ekuitas ${pct(s.equity_pct)} · dicairkan ${fmtRupiahShort(s.disbursed_total)}`} />
      <Stat label="Dukungan pemerintah" value={fmtRupiahShort(s.support_approved)} sub={`disetujui/terealisasi · ${fmtRupiahShort(s.support_proposed)} diusulkan`} />
      {usesAP ? (
        <Stat
          label="Pembayaran ketersediaan layanan"
          value={`${fmtRupiahShort(s.ap_paid)} / ${fmtRupiahShort(s.ap_net)}`}
          sub={s.ap_overdue_count > 0 ? `${s.ap_overdue_count} periode lewat jatuh tempo · tunggakan ${fmtRupiahShort(s.ap_outstanding_due)}` : `Potongan kinerja ${fmtRupiahShort(s.ap_deductions)}`}
          warn={s.ap_overdue_count > 0}
        />
      ) : (
        <Stat label="Pembayaran ketersediaan layanan" value="—" sub="Skema pengembalian bukan AP" />
      )}
    </div>
  );
}

// ── Panel ───────────────────────────────────────────────────────────────────

export function FinancePanel({ projectId, returnScheme, canWrite, canDelete }: { projectId: string; returnScheme: KpbuReturnScheme | string; canWrite: boolean; canDelete: boolean }) {
  const usesAP = returnScheme === "AVAILABILITY_PAYMENT" || returnScheme === "HYBRID";
  const summary = useQuery({
    queryKey: ["kpbu", "finance-summary", projectId],
    queryFn: () => kpbuService.financeSummary(projectId).then((r) => r.data),
  });
  const common = { projectId, canWrite, canDelete };

  return (
    <div className="space-y-5 rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
      <div>
        <h2 className="text-sm font-semibold text-gray-700">Keuangan KPBU</h2>
        <p className="text-xs text-gray-400">Struktur pendanaan badan usaha, dukungan pemerintah, dan pembayaran ketersediaan layanan. Angka adalah input operasional, bukan laporan keuangan teraudit.</p>
      </div>
      {summary.data && <SummaryCards s={summary.data} usesAP={usesAP} />}
      {summary.data && summary.data.skipped_other_currency > 0 && (
        <p className="text-xs text-orange-700">{summary.data.skipped_other_currency} baris bermata uang lain tidak dijumlahkan ke ringkasan.</p>
      )}

      <RecordSection<KpbuFinancing>
        {...common}
        title="Struktur Pendanaan"
        queryKey="financings"
        empty="Belum ada sumber pendanaan (ekuitas, pinjaman, obligasi)."
        list={(pid) => kpbuService.listFinancings(pid).then((r) => r.data)}
        remove={kpbuService.deleteFinancing}
        save={(pid, id, v) => {
          const payload: KpbuFinancingRequest = {
            source: v.source as KpbuFinancingSource,
            provider: v.provider,
            agreement_number: v.agreement_number,
            agreement_date: dateOrNull(v.agreement_date),
            amount: num(v.amount),
            disbursed_amount: num(v.disbursed_amount),
            currency: v.currency || "IDR",
            status: v.status as KpbuFinancingRequest["status"],
            notes: v.notes,
          };
          return kpbuService.saveFinancing(pid, id, payload);
        }}
        toValues={(r) => ({
          source: r?.source ?? "EQUITY",
          provider: r?.provider ?? "",
          agreement_number: r?.agreement_number ?? "",
          agreement_date: toDateInput(r?.agreement_date),
          amount: str(r?.amount),
          disbursed_amount: str(r?.disbursed_amount ?? 0),
          currency: r?.currency ?? "IDR",
          status: r?.status ?? "PLANNED",
          notes: r?.notes ?? "",
        })}
        validate={(v) => (num(v.disbursed_amount) > num(v.amount) ? "Nilai dicairkan tidak boleh melebihi nilai pendanaan." : null)}
        fields={[
          { key: "source", label: "Sumber", kind: "select", options: KPBU_FINANCING_SOURCE_LABEL, required: true },
          { key: "status", label: "Status", kind: "select", options: KPBU_FINANCING_STATUS_LABEL },
          { key: "provider", label: "Pemberi dana (sponsor/bank/investor)", kind: "text", wide: true },
          { key: "agreement_number", label: "No. perjanjian", kind: "text" },
          { key: "agreement_date", label: "Tanggal perjanjian", kind: "date" },
          { key: "amount", label: "Nilai (Rp)", kind: "number", required: true },
          { key: "disbursed_amount", label: "Sudah dicairkan (Rp)", kind: "number" },
          { key: "notes", label: "Catatan", kind: "textarea" },
        ]}
        columns={[
          { label: "Sumber", render: (r) => <span className="font-medium text-gray-800">{KPBU_FINANCING_SOURCE_LABEL[r.source]}</span> },
          { label: "Pemberi dana", render: (r) => <>{r.provider || "—"}<br /><span className="text-gray-400">{r.agreement_number || ""} {r.agreement_date ? `· ${fmtDate(r.agreement_date)}` : ""}</span></> },
          { label: "Nilai", render: (r) => <>{fmtRupiahShort(r.amount)}{r.currency !== "IDR" && ` ${r.currency}`}<br /><span className="text-gray-400">cair {fmtRupiahShort(r.disbursed_amount)}</span></> },
          { label: "Status", render: (r) => KPBU_FINANCING_STATUS_LABEL[r.status] },
        ]}
      />

      <RecordSection<KpbuGovernmentSupport>
        {...common}
        title="Dukungan Pemerintah"
        queryKey="government-supports"
        empty="Belum ada dukungan pemerintah (fiskal maupun non-fiskal)."
        note="Dukungan non-fiskal (mis. perizinan, pengadaan tanah) boleh bernilai 0."
        list={(pid) => kpbuService.listSupports(pid).then((r) => r.data)}
        remove={kpbuService.deleteSupport}
        save={(pid, id, v) => {
          const payload: KpbuGovernmentSupportRequest = {
            type: v.type as KpbuGovernmentSupportRequest["type"],
            provider: v.provider,
            decision_number: v.decision_number,
            decision_date: dateOrNull(v.decision_date),
            amount: num(v.amount),
            currency: v.currency || "IDR",
            status: v.status as KpbuGovernmentSupportRequest["status"],
            notes: v.notes,
          };
          return kpbuService.saveSupport(pid, id, payload);
        }}
        toValues={(r) => ({
          type: r?.type ?? "VGF",
          provider: r?.provider ?? "",
          decision_number: r?.decision_number ?? "",
          decision_date: toDateInput(r?.decision_date),
          amount: str(r?.amount ?? 0),
          currency: r?.currency ?? "IDR",
          status: r?.status ?? "PROPOSED",
          notes: r?.notes ?? "",
        })}
        fields={[
          { key: "type", label: "Jenis dukungan", kind: "select", options: KPBU_SUPPORT_TYPE_LABEL, required: true },
          { key: "status", label: "Status", kind: "select", options: KPBU_SUPPORT_STATUS_LABEL },
          { key: "provider", label: "Pemberi (Kemenkeu / K/L / PT PII / Pemda)", kind: "text", wide: true },
          { key: "decision_number", label: "No. keputusan/persetujuan", kind: "text" },
          { key: "decision_date", label: "Tanggal keputusan", kind: "date" },
          { key: "amount", label: "Nilai (Rp)", kind: "number" },
          { key: "notes", label: "Catatan", kind: "textarea" },
        ]}
        columns={[
          { label: "Jenis", render: (r) => <span className="font-medium text-gray-800">{KPBU_SUPPORT_TYPE_LABEL[r.type]}</span> },
          { label: "Pemberi / Keputusan", render: (r) => <>{r.provider || "—"}<br /><span className="text-gray-400">{r.decision_number || ""} {r.decision_date ? `· ${fmtDate(r.decision_date)}` : ""}</span></> },
          { label: "Nilai", render: (r) => (r.amount > 0 ? fmtRupiahShort(r.amount) : "Non-fiskal") },
          { label: "Status", render: (r) => KPBU_SUPPORT_STATUS_LABEL[r.status] },
        ]}
      />

      {usesAP ? (
        <RecordSection<KpbuAvailabilityPayment>
          {...common}
          title="Pembayaran Ketersediaan Layanan"
          queryKey="availability-payments"
          empty="Belum ada jadwal pembayaran ketersediaan layanan."
          note="Nilai bersih = jadwal − potongan kinerja. Potongan wajib diberi alasan; kaitan otomatis dengan KPI layanan menyusul di K3."
          list={(pid) => kpbuService.listPayments(pid).then((r) => r.data)}
          remove={kpbuService.deletePayment}
          save={(pid, id, v) => {
            const payload: KpbuAvailabilityPaymentRequest = {
              period_year: Number(v.period_year),
              period_month: Number(v.period_month),
              due_date: dateOrNull(v.due_date),
              scheduled_amount: num(v.scheduled_amount),
              deduction_amount: num(v.deduction_amount),
              deduction_reason: v.deduction_reason,
              paid_amount: num(v.paid_amount),
              paid_date: dateOrNull(v.paid_date),
              invoice_number: v.invoice_number,
              currency: v.currency || "IDR",
              notes: v.notes,
            };
            return kpbuService.savePayment(pid, id, payload);
          }}
          toValues={(r) => ({
            period_year: String(r?.period_year ?? new Date().getFullYear()),
            period_month: String(r?.period_month ?? new Date().getMonth() + 1),
            due_date: toDateInput(r?.due_date),
            scheduled_amount: str(r?.scheduled_amount),
            deduction_amount: str(r?.deduction_amount ?? 0),
            deduction_reason: r?.deduction_reason ?? "",
            paid_amount: str(r?.paid_amount ?? 0),
            paid_date: toDateInput(r?.paid_date),
            invoice_number: r?.invoice_number ?? "",
            currency: r?.currency ?? "IDR",
            notes: r?.notes ?? "",
          })}
          validate={(v) => {
            const y = Number(v.period_year);
            if (!Number.isInteger(y) || y < 2000 || y > 2100) return "Tahun periode tidak valid.";
            const net = num(v.scheduled_amount) - num(v.deduction_amount);
            if (net < 0) return "Potongan tidak boleh melebihi nilai jadwal.";
            if (num(v.deduction_amount) > 0 && !v.deduction_reason.trim()) return "Alasan potongan wajib diisi.";
            if (num(v.paid_amount) > net + 0.005) return "Nilai dibayar tidak boleh melebihi nilai bersih.";
            return null;
          }}
          fields={[
            { key: "period_year", label: "Tahun", kind: "number", required: true },
            { key: "period_month", label: "Bulan", kind: "select", options: MONTH_OPTIONS },
            { key: "due_date", label: "Jatuh tempo", kind: "date" },
            { key: "invoice_number", label: "No. tagihan", kind: "text" },
            { key: "scheduled_amount", label: "Nilai jadwal (Rp)", kind: "number", required: true },
            { key: "deduction_amount", label: "Potongan kinerja (Rp)", kind: "number" },
            { key: "deduction_reason", label: "Alasan potongan", kind: "text", wide: true },
            { key: "paid_amount", label: "Dibayar (Rp)", kind: "number" },
            { key: "paid_date", label: "Tanggal bayar", kind: "date" },
            { key: "notes", label: "Catatan", kind: "textarea" },
          ]}
          columns={[
            { label: "Periode", render: (r) => <span className="font-medium text-gray-800">{MONTHS[r.period_month - 1]} {r.period_year}</span> },
            { label: "Jadwal / Potongan", render: (r) => <>{fmtRupiahShort(r.scheduled_amount)}{r.deduction_amount > 0 && <span className="block text-orange-700" title={r.deduction_reason}>− {fmtRupiahShort(r.deduction_amount)}</span>}</> },
            { label: "Bersih / Dibayar", render: (r) => <>{fmtRupiahShort(r.net_amount)}<br /><span className="text-gray-400">{fmtRupiahShort(r.paid_amount)}{r.paid_date ? ` · ${fmtDate(r.paid_date)}` : ""}</span></> },
            { label: "Jatuh tempo", render: (r) => fmtDate(r.due_date) },
            {
              label: "Status",
              render: (r) => (
                <span className={cn(r.status === "OVERDUE" ? "font-medium text-red-600" : r.status === "PAID" ? "text-green-700" : "text-gray-700")}>
                  {KPBU_PAYMENT_STATUS_LABEL[r.status]}{r.status === "OVERDUE" && ` (${r.days_overdue} hari)`}
                </span>
              ),
            },
          ]}
        />
      ) : (
        <p className="text-xs text-gray-400">Pembayaran ketersediaan layanan hanya ditampilkan untuk skema pengembalian Availability Payment atau Hybrid.</p>
      )}
    </div>
  );
}
