"use client";

import { useState, type FormEvent } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check, Pencil } from "lucide-react";
import { kpbuErrorMessage, kpbuService } from "@/services/kpbu.service";
import { cn } from "@/lib/utils";
import {
  KPBU_MAIN_STAGES,
  KPBU_STAGE_LABEL,
  type KpbuProjectDetail,
  type KpbuStage,
  type KpbuStageMilestone,
} from "@/types/kpbu";
import { btnPrimary, btnSecondary, Field, fmtDate, inputCls, Modal, textareaCls, toDateInput } from "./ui";

const TERMINAL_EXITS = new Set<KpbuStage>(["ON_HOLD", "CANCELLED", "TERMINATED"]);

export function StagePanel({ detail, canWrite }: { detail: KpbuProjectDetail; canWrite: boolean }) {
  const { project, stage_milestones: milestones, allowed_next_stages: allowed } = detail;
  const qc = useQueryClient();
  const [toStage, setToStage] = useState<KpbuStage | "">("");
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [notes, setNotes] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [editing, setEditing] = useState<KpbuStageMilestone | null>(null);

  const currentIdx = (KPBU_MAIN_STAGES as readonly string[]).indexOf(
    project.stage === "ON_HOLD" && project.previous_stage ? project.previous_stage : project.stage,
  );
  const finished = project.stage === "CLOSED";

  const transition = useMutation({
    mutationFn: () => kpbuService.transition(project.id, { to_stage: toStage as KpbuStage, date, notes }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["kpbu"] });
      setToStage("");
      setNotes("");
      setError(null);
    },
    onError: (err) => setError(kpbuErrorMessage(err, "Gagal mengubah tahap.")),
  });

  const submit = (e: FormEvent) => {
    e.preventDefault();
    if (!toStage) return;
    transition.mutate();
  };

  return (
    <div className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
      <h2 className="mb-4 text-sm font-semibold text-gray-700">Tahapan KPBU</h2>

      <ol className="mb-5 grid grid-cols-2 gap-2 sm:grid-cols-5">
        {KPBU_MAIN_STAGES.map((s, i) => {
          const done = finished || i < currentIdx;
          const current = !finished && i === currentIdx;
          return (
            <li
              key={s}
              className={cn(
                "flex items-center gap-2 rounded-lg border px-2 py-1.5 text-xs",
                done && "border-green-200 bg-green-50 text-green-800",
                current && "border-blue-300 bg-blue-50 font-semibold text-blue-800",
                !done && !current && "border-gray-100 text-gray-400",
              )}
            >
              <span className={cn("flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-[10px]", done ? "bg-green-600 text-white" : current ? "bg-blue-600 text-white" : "bg-gray-100")}>
                {done ? <Check className="h-3 w-3" /> : i + 1}
              </span>
              <span className="truncate">{KPBU_STAGE_LABEL[s]}</span>
            </li>
          );
        })}
      </ol>

      {canWrite && allowed.length > 0 && (
        <form onSubmit={submit} className="mb-5 grid grid-cols-1 gap-3 rounded-lg bg-gray-50 p-3 sm:grid-cols-[1fr_10rem_1fr_auto] sm:items-end">
          <Field label="Pindah ke tahap">
            <select className={inputCls} value={toStage} onChange={(e) => setToStage(e.target.value as KpbuStage)} aria-label="Tahap tujuan">
              <option value="">— Pilih —</option>
              {allowed.map((s) => (
                <option key={s} value={s}>
                  {TERMINAL_EXITS.has(s) ? `⚠ ${KPBU_STAGE_LABEL[s]}` : KPBU_STAGE_LABEL[s]}
                </option>
              ))}
            </select>
          </Field>
          <Field label="Tanggal">
            <input className={inputCls} type="date" value={date} onChange={(e) => setDate(e.target.value)} />
          </Field>
          <Field label="Catatan">
            <input className={inputCls} value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="cth. OBC disetujui" />
          </Field>
          <button type="submit" className={btnPrimary} disabled={!toStage || transition.isPending}>
            {transition.isPending ? "Menyimpan..." : "Ubah Tahap"}
          </button>
          {error && <p className="text-sm text-red-600 sm:col-span-4">{error}</p>}
        </form>
      )}

      <div className="overflow-x-auto">
        <table className="w-full min-w-[640px] text-sm">
          <thead className="border-b border-gray-100 text-left text-xs uppercase text-gray-500">
            <tr>
              <th className="py-2 pr-3">Tahap</th>
              <th className="py-2 pr-3">Rencana</th>
              <th className="py-2 pr-3">Realisasi</th>
              <th className="py-2 pr-3">Status</th>
              {canWrite && <th className="py-2" />}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {milestones.map((m) => (
              <tr key={m.stage}>
                <td className="py-2 pr-3 font-medium text-gray-800">
                  {KPBU_STAGE_LABEL[m.stage]}
                  {m.notes && <p className="text-xs font-normal text-gray-400">{m.notes}</p>}
                </td>
                <td className="py-2 pr-3 text-xs text-gray-600">{fmtDate(m.planned_start)} – {fmtDate(m.planned_end)}</td>
                <td className="py-2 pr-3 text-xs text-gray-600">{fmtDate(m.actual_start)} – {fmtDate(m.actual_end)}</td>
                <td className="py-2 pr-3 text-xs">
                  {m.is_delayed ? (
                    <span className="font-medium text-red-600">Terlambat {m.delay_days} hari</span>
                  ) : m.actual_end ? (
                    <span className="text-green-700">Selesai</span>
                  ) : m.actual_start ? (
                    <span className="text-blue-700">Berjalan</span>
                  ) : (
                    <span className="text-gray-400">—</span>
                  )}
                </td>
                {canWrite && (
                  <td className="py-2 text-right">
                    <button className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-700" onClick={() => setEditing(m)} aria-label={`Edit target ${KPBU_STAGE_LABEL[m.stage]}`}>
                      <Pencil className="h-3.5 w-3.5" />
                    </button>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {editing && <MilestoneModal projectId={project.id} milestone={editing} onClose={() => setEditing(null)} />}
    </div>
  );
}

function MilestoneModal({ projectId, milestone, onClose }: { projectId: string; milestone: KpbuStageMilestone; onClose: () => void }) {
  const qc = useQueryClient();
  const [v, setV] = useState({
    planned_start: toDateInput(milestone.planned_start),
    planned_end: toDateInput(milestone.planned_end),
    actual_start: toDateInput(milestone.actual_start),
    actual_end: toDateInput(milestone.actual_end),
    notes: milestone.notes ?? "",
  });
  const [error, setError] = useState<string | null>(null);
  const set = (k: keyof typeof v) => (e: { target: { value: string } }) => setV((s) => ({ ...s, [k]: e.target.value }));
  const orNull = (s: string) => (s === "" ? null : s);

  const save = useMutation({
    mutationFn: () =>
      kpbuService.upsertStageMilestone(projectId, {
        stage: milestone.stage,
        planned_start: orNull(v.planned_start),
        planned_end: orNull(v.planned_end),
        actual_start: orNull(v.actual_start),
        actual_end: orNull(v.actual_end),
        notes: v.notes,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["kpbu"] });
      onClose();
    },
    onError: (err) => setError(kpbuErrorMessage(err, "Gagal menyimpan target tahap.")),
  });

  return (
    <Modal title={`Target Tahap — ${KPBU_STAGE_LABEL[milestone.stage]}`} onClose={onClose}>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          save.mutate();
        }}
        className="grid grid-cols-2 gap-3"
      >
        <Field label="Rencana Mulai"><input className={inputCls} type="date" value={v.planned_start} onChange={set("planned_start")} /></Field>
        <Field label="Rencana Selesai"><input className={inputCls} type="date" value={v.planned_end} onChange={set("planned_end")} /></Field>
        <Field label="Realisasi Mulai"><input className={inputCls} type="date" value={v.actual_start} onChange={set("actual_start")} /></Field>
        <Field label="Realisasi Selesai"><input className={inputCls} type="date" value={v.actual_end} onChange={set("actual_end")} /></Field>
        <Field label="Catatan" className="col-span-2"><textarea className={textareaCls} rows={2} value={v.notes} onChange={set("notes")} /></Field>
        {error && <p className="col-span-2 text-sm text-red-600">{error}</p>}
        <div className="col-span-2 flex justify-end gap-2">
          <button type="button" className={btnSecondary} onClick={onClose}>Batal</button>
          <button type="submit" className={btnPrimary} disabled={save.isPending}>{save.isPending ? "Menyimpan..." : "Simpan"}</button>
        </div>
      </form>
    </Modal>
  );
}
