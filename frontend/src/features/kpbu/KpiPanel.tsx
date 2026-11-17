"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CheckCircle2, Plus, XCircle } from "lucide-react";
import { kpbuErrorMessage, kpbuService } from "@/services/kpbu.service";
import { cn } from "@/lib/utils";
import {
  KPBU_KPI_DIRECTION_LABEL,
  KPBU_MEASUREMENT_STATUS_LABEL,
  KPBU_SCORE_STATUS_LABEL,
  type KpbuKpi,
  type KpbuKpiDirection,
  type KpbuKpiMeasurement,
  type KpbuMeasurementRequest,
  type KpbuScoreStatus,
} from "@/types/kpbu";
import { MONTH_OPTIONS, MONTHS, num, RecordSection } from "./FinancePanel";
import { btnPrimary, btnSecondary, Field, fmtRupiahShort, inputCls, Modal, textareaCls } from "./ui";

const SCORE_CLS: Record<KpbuScoreStatus, string> = {
  MET: "text-green-700",
  MISSED: "font-medium text-red-600",
  PENDING: "text-orange-700",
  NO_DATA: "text-gray-400",
};

export const fmtValue = (v: number, unit: string) => `${v.toLocaleString("id-ID")}${unit ? ` ${unit}` : ""}`;
const targetLabel = (k: KpbuKpi) => `${k.direction === "LOWER_BETTER" ? "≤" : "≥"} ${fmtValue(k.target_value, k.unit)}`;

/** Modal for entering a KPI value. Used internally and from the BUP portal. */
export function MeasurementModal({
  kpis,
  onSubmit,
  onDone,
  onClose,
  portal,
}: {
  kpis: KpbuKpi[];
  onSubmit: (payload: KpbuMeasurementRequest) => Promise<unknown>;
  onDone: () => void;
  onClose: () => void;
  portal?: boolean;
}) {
  const now = new Date();
  const [v, setV] = useState({ kpi_id: kpis[0]?.id ?? "", year: String(now.getFullYear()), month: String(now.getMonth() + 1), value: "", notes: "" });
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: () => onSubmit({ kpi_id: v.kpi_id, period_year: Number(v.year), period_month: Number(v.month), actual_value: Number(v.value), notes: v.notes }),
    onSuccess: onDone,
    onError: (e) => setError(kpbuErrorMessage(e, "Gagal menyimpan nilai KPI.")),
  });
  const set = (k: keyof typeof v) => (e: { target: { value: string } }) => setV((s) => ({ ...s, [k]: e.target.value }));
  const kpi = kpis.find((k) => k.id === v.kpi_id);

  return (
    <Modal title="Laporkan Nilai KPI" onClose={onClose}>
      <form
        className="grid grid-cols-2 gap-3"
        onSubmit={(e) => {
          e.preventDefault();
          if (!v.kpi_id) return setError("Pilih KPI.");
          if (v.value === "" || !Number.isFinite(Number(v.value))) return setError("Nilai realisasi wajib berupa angka.");
          const y = Number(v.year);
          if (!Number.isInteger(y) || y < 2000 || y > 2100) return setError("Tahun tidak valid.");
          setError(null);
          save.mutate();
        }}
      >
        <Field label="KPI" required className="col-span-2">
          <select className={inputCls} value={v.kpi_id} onChange={set("kpi_id")}>
            {kpis.map((k) => <option key={k.id} value={k.id}>{k.code} · {k.name}</option>)}
          </select>
        </Field>
        {kpi && <p className="col-span-2 text-xs text-gray-500">Target {targetLabel(kpi)}</p>}
        <Field label="Tahun" required><input className={inputCls} type="number" value={v.year} onChange={set("year")} /></Field>
        <Field label="Bulan" required>
          <select className={inputCls} value={v.month} onChange={set("month")}>
            {Object.entries(MONTH_OPTIONS).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
          </select>
        </Field>
        <Field label={`Realisasi${kpi?.unit ? ` (${kpi.unit})` : ""}`} required className="col-span-2">
          <input className={inputCls} type="number" step="any" value={v.value} onChange={set("value")} />
        </Field>
        <Field label="Catatan / dasar pengukuran" className="col-span-2"><textarea className={textareaCls} rows={2} value={v.notes} onChange={set("notes")} /></Field>
        <p className="col-span-2 text-xs text-gray-400">
          {portal ? "Nilai akan diverifikasi PJPK sebelum dipakai menghitung potongan pembayaran." : "Nilai yang dimasukkan PJPK langsung tercatat terverifikasi."}
        </p>
        {error && <p className="col-span-2 text-sm text-red-600">{error}</p>}
        <div className="col-span-2 flex justify-end gap-2">
          <button type="button" className={btnSecondary} onClick={onClose}>Batal</button>
          <button type="submit" className={btnPrimary} disabled={save.isPending}>{save.isPending ? "Menyimpan..." : "Simpan"}</button>
        </div>
      </form>
    </Modal>
  );
}

function ReviewModal({ projectId, m, onClose }: { projectId: string; m: KpbuKpiMeasurement; onClose: () => void }) {
  const qc = useQueryClient();
  const [notes, setNotes] = useState("");
  const [error, setError] = useState<string | null>(null);
  const review = useMutation({
    mutationFn: (d: "VERIFIED" | "REJECTED") => kpbuService.reviewMeasurement(projectId, m.id, d, notes),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["kpbu", "kpi", projectId] });
      onClose();
    },
    onError: (e) => setError(kpbuErrorMessage(e, "Gagal menyimpan verifikasi.")),
  });
  return (
    <Modal title="Verifikasi Nilai KPI" onClose={onClose}>
      <div className="space-y-3 text-sm">
        <p><span className="font-medium">{m.kpi_code} · {m.kpi_name}</span> — {MONTHS[m.period_month - 1]} {m.period_year}</p>
        <p>Realisasi: <span className="font-medium">{m.actual_value.toLocaleString("id-ID")}</span> {m.met != null && (m.met ? <span className="text-green-700">(memenuhi target)</span> : <span className="text-red-600">(tidak memenuhi target)</span>)}</p>
        {m.notes && <p className="whitespace-pre-line text-gray-600">{m.notes}</p>}
        <Field label="Catatan verifikasi (wajib bila ditolak)"><textarea className={textareaCls} rows={3} value={notes} onChange={(e) => setNotes(e.target.value)} /></Field>
        {error && <p className="text-red-600">{error}</p>}
        <div className="flex justify-end gap-2">
          <button className={btnSecondary} onClick={onClose}>Batal</button>
          <button className={cn(btnSecondary, "text-red-700")} disabled={review.isPending} onClick={() => (notes.trim() ? review.mutate("REJECTED") : setError("Alasan penolakan wajib diisi."))}>
            <XCircle className="h-4 w-4" /> Tolak
          </button>
          <button className={btnPrimary} disabled={review.isPending} onClick={() => review.mutate("VERIFIED")}><CheckCircle2 className="h-4 w-4" /> Verifikasi</button>
        </div>
      </div>
    </Modal>
  );
}

export function KpiPanel({ projectId, canWrite, canDelete }: { projectId: string; canWrite: boolean; canDelete: boolean }) {
  const qc = useQueryClient();
  const now = new Date();
  const [year, setYear] = useState(now.getFullYear());
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [entering, setEntering] = useState(false);
  const [reviewing, setReviewing] = useState<KpbuKpiMeasurement | null>(null);

  const kpis = useQuery({ queryKey: ["kpbu", "kpi", projectId, "defs"], queryFn: () => kpbuService.listKpis(projectId).then((r) => r.data) });
  const score = useQuery({ queryKey: ["kpbu", "kpi", projectId, "score", year, month], queryFn: () => kpbuService.scorecard(projectId, year, month).then((r) => r.data) });
  const pending = useQuery({ queryKey: ["kpbu", "kpi", projectId, "pending"], queryFn: () => kpbuService.listMeasurements(projectId, { status: "SUBMITTED" }).then((r) => r.data) });
  const refresh = () => qc.invalidateQueries({ queryKey: ["kpbu", "kpi", projectId] });
  const activeKpis = (kpis.data ?? []).filter((k) => k.is_active);
  const sc = score.data;

  return (
    <div className="space-y-5 rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
      <div>
        <h2 className="text-sm font-semibold text-gray-700">KPI Layanan</h2>
        <p className="text-xs text-gray-400">Indikator kinerja layanan badan usaha. KPI terverifikasi yang tidak tercapai memberi saran potongan pembayaran ketersediaan layanan; potongan tetap dicatat manual oleh PJPK.</p>
      </div>

      <RecordSection<KpbuKpi>
        projectId={projectId}
        canWrite={canWrite}
        canDelete={canDelete}
        title="Definisi KPI"
        queryKey="kpi-defs"
        empty="Belum ada KPI. Tambahkan sesuai lampiran spesifikasi keluaran di Perjanjian KPBU."
        list={(pid) => kpbuService.listKpis(pid).then((r) => r.data)}
        remove={async (pid, id) => {
          await kpbuService.deleteKpi(pid, id);
          refresh();
        }}
        save={async (pid, id, v) => {
          await kpbuService.saveKpi(pid, id, {
            code: v.code,
            name: v.name,
            description: v.description,
            unit: v.unit,
            target_value: num(v.target_value),
            direction: v.direction as KpbuKpiDirection,
            deduction_pct: num(v.deduction_pct),
            is_active: v.is_active === "true",
            sort_order: num(v.sort_order),
          });
          refresh();
        }}
        toValues={(k) => ({
          code: k?.code ?? "",
          name: k?.name ?? "",
          description: k?.description ?? "",
          unit: k?.unit ?? "",
          target_value: k ? String(k.target_value) : "",
          direction: k?.direction ?? "HIGHER_BETTER",
          deduction_pct: String(k?.deduction_pct ?? 0),
          is_active: String(k?.is_active ?? true),
          sort_order: String(k?.sort_order ?? 0),
        })}
        validate={(v) => (num(v.deduction_pct) > 100 ? "Potongan maksimal 100%." : null)}
        fields={[
          { key: "code", label: "Kode", kind: "text", required: true },
          { key: "name", label: "Nama indikator", kind: "text", required: true },
          { key: "target_value", label: "Target", kind: "number", required: true },
          { key: "unit", label: "Satuan (%, jam, m³/detik, ...)", kind: "text" },
          { key: "direction", label: "Arah target", kind: "select", options: KPBU_KPI_DIRECTION_LABEL },
          { key: "deduction_pct", label: "Potongan bila tidak tercapai (% AP periode)", kind: "number" },
          { key: "is_active", label: "Status", kind: "select", options: { true: "Aktif", false: "Nonaktif" } },
          { key: "sort_order", label: "Urutan", kind: "number" },
          { key: "description", label: "Cara pengukuran", kind: "textarea" },
        ]}
        columns={[
          { label: "Kode", render: (k) => <span className={cn("font-semibold", !k.is_active && "text-gray-400")}>{k.code}</span> },
          { label: "Indikator", render: (k) => <>{k.name}{!k.is_active && <span className="text-gray-400"> · nonaktif</span>}</> },
          { label: "Target", render: targetLabel },
          { label: "Potongan", render: (k) => `${k.deduction_pct.toLocaleString("id-ID")}%` },
        ]}
      />

      {!!pending.data?.length && (
        <div>
          <h3 className="mb-2 text-sm font-semibold text-gray-700">Menunggu Verifikasi ({pending.data.length})</h3>
          <ul className="divide-y divide-gray-50 text-sm">
            {pending.data.map((m) => (
              <li key={m.id} className="flex items-center justify-between gap-2 py-2">
                <span>
                  <span className="font-medium">{m.kpi_code}</span> · {MONTHS[m.period_month - 1]} {m.period_year}: {m.actual_value.toLocaleString("id-ID")}
                  {m.source === "PORTAL" && <span className="ml-1 text-xs text-gray-400">(dari portal BUP)</span>}
                </span>
                {canWrite && <button className={btnSecondary} onClick={() => setReviewing(m)}>Periksa</button>}
              </li>
            ))}
          </ul>
        </div>
      )}

      <div>
        <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
          <h3 className="text-sm font-semibold text-gray-700">Scorecard Periode</h3>
          <div className="flex items-center gap-2">
            <select className={cn(inputCls, "w-auto")} value={month} onChange={(e) => setMonth(Number(e.target.value))} aria-label="Bulan">
              {Object.entries(MONTH_OPTIONS).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
            </select>
            <input className={cn(inputCls, "w-24")} type="number" value={year} onChange={(e) => setYear(Number(e.target.value) || now.getFullYear())} aria-label="Tahun" />
            {canWrite && activeKpis.length > 0 && <button className={btnSecondary} onClick={() => setEntering(true)}><Plus className="h-4 w-4" /> Nilai</button>}
          </div>
        </div>
        {score.isLoading ? (
          <p className="py-3 text-center text-sm text-gray-400">Memuat...</p>
        ) : score.isError ? (
          <p className="py-3 text-center text-sm text-red-600">{kpbuErrorMessage(score.error, "Gagal memuat scorecard.")}</p>
        ) : !sc?.rows.length ? (
          <p className="py-3 text-center text-sm text-gray-400">Belum ada KPI aktif.</p>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full min-w-[560px] text-sm">
                <thead className="border-b border-gray-100 text-left text-xs uppercase text-gray-500">
                  <tr><th className="py-2 pr-3">KPI</th><th className="py-2 pr-3">Target</th><th className="py-2 pr-3">Realisasi</th><th className="py-2 pr-3">Status</th></tr>
                </thead>
                <tbody className="divide-y divide-gray-50">
                  {sc.rows.map((r) => (
                    <tr key={r.kpi.id}>
                      <td className="py-2 pr-3 text-xs"><span className="font-semibold">{r.kpi.code}</span> · {r.kpi.name}</td>
                      <td className="py-2 pr-3 text-xs">{targetLabel(r.kpi)}</td>
                      <td className="py-2 pr-3 text-xs">{r.measurement ? fmtValue(r.measurement.actual_value, r.kpi.unit) : "—"}</td>
                      <td className={cn("py-2 pr-3 text-xs", SCORE_CLS[r.status])}>
                        {KPBU_SCORE_STATUS_LABEL[r.status]}{r.status === "MISSED" && r.kpi.deduction_pct > 0 && ` (−${r.kpi.deduction_pct}%)`}
                        {r.measurement && r.status === "PENDING" && <span className="block text-gray-400">{KPBU_MEASUREMENT_STATUS_LABEL[r.measurement.status]}</span>}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <div className="mt-2 rounded-lg bg-gray-50 p-3 text-xs text-gray-700">
              Tercapai {sc.met} · Tidak tercapai {sc.missed} · Menunggu {sc.pending} · Belum dilaporkan {sc.no_data}. Saran potongan{" "}
              <span className="font-semibold">{sc.suggested_deduction_pct.toLocaleString("id-ID")}%</span>
              {sc.scheduled_payment != null ? (
                <> dari jadwal AP {fmtRupiahShort(sc.scheduled_payment)} = <span className="font-semibold">{fmtRupiahShort(sc.suggested_deduction ?? 0)}</span>; tercatat {fmtRupiahShort(sc.recorded_deduction ?? 0)}.</>
              ) : (
                <> (belum ada jadwal pembayaran ketersediaan layanan untuk periode ini).</>
              )}
            </div>
          </>
        )}
      </div>

      {entering && (
        <MeasurementModal
          kpis={activeKpis}
          onSubmit={(p) => kpbuService.createMeasurement(projectId, p)}
          onDone={() => {
            refresh();
            setEntering(false);
          }}
          onClose={() => setEntering(false)}
        />
      )}
      {reviewing && <ReviewModal projectId={projectId} m={reviewing} onClose={() => setReviewing(null)} />}
    </div>
  );
}
