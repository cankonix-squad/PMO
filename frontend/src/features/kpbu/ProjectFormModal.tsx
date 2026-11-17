"use client";

import { useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { z } from "zod";
import { kpbuErrorMessage, kpbuService } from "@/services/kpbu.service";
import { orgUnitService } from "@/services/org-unit.service";
import {
  KPBU_INITIATIVE_LABEL,
  KPBU_RETURN_SCHEME_LABEL,
  KPBU_SECTOR_LABEL,
  type KpbuInitiative,
  type KpbuProject,
  type KpbuProjectRequest,
  type KpbuReturnScheme,
  type KpbuSector,
} from "@/types/kpbu";
import { btnPrimary, btnSecondary, Field, inputCls, Modal, textareaCls, toDateInput } from "./ui";

const optionalNumber = (min: number, max: number, msg: string) =>
  z
    .string()
    .trim()
    .refine((v) => v === "" || (Number.isFinite(Number(v)) && Number(v) >= min && Number(v) <= max), msg);

const schema = z
  .object({
    code: z.string().trim().min(1, "Kode wajib diisi").max(50),
    name: z.string().trim().min(1, "Nama wajib diisi").max(300),
    description: z.string(),
    sector: z.string().min(1, "Sektor wajib dipilih"),
    initiative: z.string(),
    return_scheme: z.string(),
    pjpk_name: z.string().trim().max(200),
    pjpk_unit: z.string().trim().max(200),
    investment_value: z
      .string()
      .trim()
      .refine((v) => v !== "" && Number.isFinite(Number(v)) && Number(v) >= 0, "Nilai investasi wajib, tidak boleh negatif"),
    concession_years: optionalNumber(1, 99, "Masa konsesi 1–99 tahun"),
    concession_start: z.string(),
    concession_end: z.string(),
    financial_close_target: z.string(),
    cod_target: z.string(),
    org_unit_id: z.string(),
    province: z.string().trim().max(100),
    latitude: optionalNumber(-90, 90, "Latitude harus antara -90 dan 90"),
    longitude: optionalNumber(-180, 180, "Longitude harus antara -180 dan 180"),
  })
  .refine((v) => (v.latitude === "") === (v.longitude === ""), {
    message: "Latitude dan longitude harus diisi berpasangan",
  })
  .refine((v) => !v.concession_start || !v.concession_end || v.concession_end >= v.concession_start, {
    message: "Akhir konsesi tidak boleh sebelum awal konsesi",
  });

type FormValues = z.infer<typeof schema>;

function initial(p?: KpbuProject): FormValues {
  return {
    code: p?.code ?? "",
    name: p?.name ?? "",
    description: p?.description ?? "",
    sector: p?.sector ?? "SDA",
    initiative: p?.initiative ?? "SOLICITED",
    return_scheme: p?.return_scheme ?? "AVAILABILITY_PAYMENT",
    pjpk_name: p?.pjpk_name ?? "",
    pjpk_unit: p?.pjpk_unit ?? "",
    investment_value: p ? String(p.investment_value) : "",
    concession_years: p?.concession_years != null ? String(p.concession_years) : "",
    concession_start: toDateInput(p?.concession_start),
    concession_end: toDateInput(p?.concession_end),
    financial_close_target: toDateInput(p?.financial_close_target),
    cod_target: toDateInput(p?.cod_target),
    org_unit_id: p?.org_unit_id ?? "",
    province: p?.province ?? "",
    latitude: p?.latitude != null ? String(p.latitude) : "",
    longitude: p?.longitude != null ? String(p.longitude) : "",
  };
}

const numOrNull = (v: string) => (v.trim() === "" ? null : Number(v));
const strOrNull = (v: string) => (v.trim() === "" ? null : v);

function toPayload(v: FormValues): KpbuProjectRequest {
  return {
    code: v.code.trim(),
    name: v.name.trim(),
    description: v.description,
    sector: v.sector as KpbuSector,
    initiative: v.initiative as KpbuInitiative,
    return_scheme: v.return_scheme as KpbuReturnScheme,
    pjpk_name: v.pjpk_name,
    pjpk_unit: v.pjpk_unit,
    investment_value: Number(v.investment_value),
    currency: "IDR",
    concession_years: numOrNull(v.concession_years),
    concession_start: strOrNull(v.concession_start),
    concession_end: strOrNull(v.concession_end),
    financial_close_target: strOrNull(v.financial_close_target),
    cod_target: strOrNull(v.cod_target),
    org_unit_id: strOrNull(v.org_unit_id),
    province: v.province,
    latitude: numOrNull(v.latitude),
    longitude: numOrNull(v.longitude),
  };
}

export function ProjectFormModal({ project, onClose, onSaved }: { project?: KpbuProject; onClose: () => void; onSaved?: (p: KpbuProject) => void }) {
  const qc = useQueryClient();
  const [values, setValues] = useState<FormValues>(() => initial(project));
  const [error, setError] = useState<string | null>(null);

  const { data: orgUnits = [] } = useQuery({
    queryKey: ["org-units"],
    queryFn: () => orgUnitService.listOrgUnits(),
  });

  const mutation = useMutation({
    mutationFn: (payload: KpbuProjectRequest) =>
      project ? kpbuService.update(project.id, payload) : kpbuService.create(payload),
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ["kpbu"] });
      onSaved?.(res.data);
      onClose();
    },
    onError: (err) => setError(kpbuErrorMessage(err, "Gagal menyimpan proyek KPBU.")),
  });

  const set = (key: keyof FormValues) => (e: { target: { value: string } }) =>
    setValues((v) => ({ ...v, [key]: e.target.value }));

  const submit = (e: FormEvent) => {
    e.preventDefault();
    const parsed = schema.safeParse(values);
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? "Data belum valid.");
      return;
    }
    setError(null);
    mutation.mutate(toPayload(parsed.data));
  };

  return (
    <Modal title={project ? "Edit Proyek KPBU" : "Proyek KPBU Baru"} onClose={onClose} wide>
      <form onSubmit={submit} className="space-y-5">
        <section className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <Field label="Kode" required>
            <input className={inputCls} value={values.code} onChange={set("code")} placeholder="cth. KPBU-SDA-001" maxLength={50} />
          </Field>
          <Field label="Sektor" required>
            <select className={inputCls} value={values.sector} onChange={set("sector")}>
              {Object.entries(KPBU_SECTOR_LABEL).map(([k, l]) => (
                <option key={k} value={k}>{l}</option>
              ))}
            </select>
          </Field>
          <Field label="Nama Proyek" required className="sm:col-span-2">
            <input className={inputCls} value={values.name} onChange={set("name")} placeholder="cth. KPBU Bendungan ..." maxLength={300} />
          </Field>
          <Field label="Deskripsi" className="sm:col-span-2">
            <textarea className={textareaCls} rows={3} value={values.description} onChange={set("description")} />
          </Field>
        </section>

        <section className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <h3 className="text-sm font-semibold text-gray-800 sm:col-span-2">Skema Kerja Sama</h3>
          <Field label="Prakarsa">
            <select className={inputCls} value={values.initiative} onChange={set("initiative")}>
              {Object.entries(KPBU_INITIATIVE_LABEL).map(([k, l]) => (
                <option key={k} value={k}>{l}</option>
              ))}
            </select>
          </Field>
          <Field label="Skema Pengembalian Investasi">
            <select className={inputCls} value={values.return_scheme} onChange={set("return_scheme")}>
              {Object.entries(KPBU_RETURN_SCHEME_LABEL).map(([k, l]) => (
                <option key={k} value={k}>{l}</option>
              ))}
            </select>
          </Field>
          <Field label="Nama PJPK">
            <input className={inputCls} value={values.pjpk_name} onChange={set("pjpk_name")} placeholder="cth. Menteri Pekerjaan Umum" maxLength={200} />
          </Field>
          <Field label="Unit PJPK / Pelaksana">
            <input className={inputCls} value={values.pjpk_unit} onChange={set("pjpk_unit")} placeholder="cth. Ditjen SDA" maxLength={200} />
          </Field>
          <Field label="Nilai Investasi (Rp)" required>
            <input className={inputCls} type="number" min={0} step="any" value={values.investment_value} onChange={set("investment_value")} placeholder="cth. 1500000000000" />
          </Field>
          <Field label="Masa Konsesi (tahun)">
            <input className={inputCls} type="number" min={1} max={99} value={values.concession_years} onChange={set("concession_years")} />
          </Field>
          <Field label="Awal Konsesi">
            <input className={inputCls} type="date" value={values.concession_start} onChange={set("concession_start")} />
          </Field>
          <Field label="Akhir Konsesi">
            <input className={inputCls} type="date" value={values.concession_end} onChange={set("concession_end")} />
          </Field>
          <Field label="Target Financial Close">
            <input className={inputCls} type="date" value={values.financial_close_target} onChange={set("financial_close_target")} />
          </Field>
          <Field label="Target COD (mulai operasi)">
            <input className={inputCls} type="date" value={values.cod_target} onChange={set("cod_target")} />
          </Field>
        </section>

        <section className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <h3 className="text-sm font-semibold text-gray-800 sm:col-span-2">Lokasi</h3>
          <Field label="Balai / Unit Pemilik">
            <select className={inputCls} value={values.org_unit_id} onChange={set("org_unit_id")}>
              <option value="">— Tidak ada —</option>
              {orgUnits.map((u) => (
                <option key={u.id} value={u.id}>{u.name}</option>
              ))}
            </select>
          </Field>
          <Field label="Provinsi">
            <input className={inputCls} value={values.province} onChange={set("province")} maxLength={100} />
          </Field>
          <Field label="Latitude">
            <input className={inputCls} type="number" step="any" value={values.latitude} onChange={set("latitude")} placeholder="cth. -6.6612" />
          </Field>
          <Field label="Longitude">
            <input className={inputCls} type="number" step="any" value={values.longitude} onChange={set("longitude")} placeholder="cth. 106.8530" />
          </Field>
        </section>

        {error && <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p>}

        <div className="flex justify-end gap-2 border-t border-gray-100 pt-4">
          <button type="button" className={btnSecondary} onClick={onClose}>Batal</button>
          <button type="submit" className={btnPrimary} disabled={mutation.isPending}>
            {mutation.isPending ? "Menyimpan..." : "Simpan"}
          </button>
        </div>
      </form>
    </Modal>
  );
}
