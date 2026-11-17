"use client";

import { useMemo, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, Check, Download, Trash2, Upload, X } from "lucide-react";
import { kpbuErrorMessage, kpbuService } from "@/services/kpbu.service";
import { cn } from "@/lib/utils";
import {
  KPBU_DOC_CATEGORY_LABEL,
  KPBU_DOC_STATUS_LABEL,
  KPBU_SOURCE_LABEL,
  KPBU_STAGE_LABEL,
  type KpbuChecklistItem,
  type KpbuDocCategory,
  type KpbuDocStatus,
  type KpbuDocument,
  type KpbuDocumentRequirement,
} from "@/types/kpbu";
import { btnPrimary, btnSecondary, Field, fmtDate, inputCls, Modal, textareaCls } from "./ui";

const STATUS_TONE: Record<KpbuDocStatus, string> = {
  MISSING: "bg-gray-100 text-gray-600",
  SUBMITTED: "bg-amber-50 text-amber-800",
  APPROVED: "bg-green-50 text-green-700",
  REJECTED: "bg-red-50 text-red-700",
};

function StatusBadge({ status }: { status: KpbuDocStatus }) {
  return (
    <span className={cn("inline-flex whitespace-nowrap rounded-full px-2 py-0.5 text-xs font-medium", STATUS_TONE[status])}>
      {KPBU_DOC_STATUS_LABEL[status]}
    </span>
  );
}

function SourceBadge({ req }: { req: KpbuDocumentRequirement }) {
  if (req.source_status === "VERIFIED") return null;
  return (
    <span
      title={KPBU_SOURCE_LABEL[req.source_status]}
      className={cn(
        "inline-flex whitespace-nowrap rounded px-1.5 py-0.5 text-[10px] font-medium",
        req.source_status === "UNVERIFIED" ? "bg-orange-50 text-orange-700" : "bg-sky-50 text-sky-700",
      )}
    >
      {req.source_status === "UNVERIFIED" ? "Belum terverifikasi" : "Sumber sekunder"}
    </span>
  );
}

function fmtSize(n: number): string {
  if (n >= 1 << 20) return `${(n / (1 << 20)).toFixed(1)} MB`;
  return `${Math.max(1, Math.round(n / 1024))} KB`;
}

export function DocumentsPanel({ projectId, canWrite, canReview, canDelete }: { projectId: string; canWrite: boolean; canReview: boolean; canDelete: boolean }) {
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["kpbu", "checklist", projectId],
    queryFn: () => kpbuService.checklist(projectId).then((r) => r.data),
  });
  const [uploadFor, setUploadFor] = useState<KpbuDocumentRequirement | "other" | null>(null);
  const [reviewing, setReviewing] = useState<KpbuDocument | null>(null);

  const groups = useMemo(() => {
    const m = new Map<KpbuDocCategory, KpbuChecklistItem[]>();
    for (const it of data?.items ?? []) {
      const list = m.get(it.requirement.category) ?? [];
      list.push(it);
      m.set(it.requirement.category, list);
    }
    return Array.from(m.entries());
  }, [data]);

  if (isLoading) {
    return <Shell><p className="py-6 text-center text-sm text-gray-400">Memuat checklist dokumen...</p></Shell>;
  }
  if (isError || !data) {
    return (
      <Shell>
        <p className="py-6 text-center text-sm text-red-600">
          Gagal memuat checklist. <button className="underline" onClick={() => refetch()}>Coba lagi</button>
        </p>
      </Shell>
    );
  }

  const s = data.summary;
  const pct = s.mandatory_total ? Math.round((s.mandatory_approved / s.mandatory_total) * 100) : 0;

  return (
    <Shell
      action={canWrite ? <button className={btnSecondary} onClick={() => setUploadFor("other")}><Upload className="h-4 w-4" /> Dokumen Lain</button> : undefined}
    >
      <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
        <div className="rounded-lg bg-gray-50 p-3">
          <p className="text-xs text-gray-500">Dokumen wajib disetujui</p>
          <p className="text-lg font-semibold text-gray-900">{s.mandatory_approved} / {s.mandatory_total}</p>
          <div className="mt-1 h-1.5 rounded-full bg-gray-200"><div className="h-1.5 rounded-full bg-green-600" style={{ width: `${pct}%` }} /></div>
        </div>
        <div className="rounded-lg bg-gray-50 p-3">
          <p className="text-xs text-gray-500">Menunggu verifikasi</p>
          <p className="text-lg font-semibold text-amber-700">{s.pending_review}</p>
        </div>
        <div className={cn("rounded-lg p-3", s.blocking_next_stage.length ? "bg-red-50" : "bg-gray-50")}>
          <p className="text-xs text-gray-500">
            Syarat masuk tahap berikutnya{s.next_stage ? ` (${KPBU_STAGE_LABEL[s.next_stage]})` : ""}
          </p>
          {s.blocking_next_stage.length ? (
            <p className="text-sm font-medium text-red-700">Kurang: {s.blocking_next_stage.join(", ")}</p>
          ) : (
            <p className="text-sm font-medium text-green-700">Lengkap</p>
          )}
        </div>
      </div>

      <div className="space-y-5">
        {groups.map(([cat, items]) => (
          <section key={cat}>
            <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">{KPBU_DOC_CATEGORY_LABEL[cat] ?? cat}</h3>
            <ul className="divide-y divide-gray-50 rounded-lg border border-gray-100">
              {items.map((it) => (
                <ChecklistRow
                  key={it.requirement.id}
                  item={it}
                  projectId={projectId}
                  canWrite={canWrite}
                  canReview={canReview}
                  canDelete={canDelete}
                  onUpload={() => setUploadFor(it.requirement)}
                  onReview={setReviewing}
                />
              ))}
            </ul>
          </section>
        ))}
        {data.other_documents.length > 0 && (
          <section>
            <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">Dokumen Lain</h3>
            <ul className="divide-y divide-gray-50 rounded-lg border border-gray-100">
              {data.other_documents.map((d) => (
                <li key={d.id} className="px-3 py-2">
                  <DocLine doc={d} projectId={projectId} canReview={canReview} canDelete={canDelete} onReview={setReviewing} />
                </li>
              ))}
            </ul>
          </section>
        )}
      </div>

      <p className="mt-4 text-xs text-gray-400">
        Checklist berasal dari Matriks Dokumen Badan Usaha (Perpres 38/2015, Peraturan LKPP 1/2025, Permen PPN 7/2023). Baris bertanda
        &ldquo;Belum terverifikasi&rdquo; perlu dicek terhadap teks Peraturan LKPP 1/2025 dan dokumen pengadaan proyek.
      </p>

      {uploadFor && (
        <UploadModal projectId={projectId} requirement={uploadFor === "other" ? null : uploadFor} onClose={() => setUploadFor(null)} />
      )}
      {reviewing && <ReviewModal projectId={projectId} doc={reviewing} onClose={() => setReviewing(null)} />}
    </Shell>
  );
}

function Shell({ children, action }: { children: React.ReactNode; action?: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
      <div className="mb-4 flex items-center justify-between gap-2">
        <h2 className="text-sm font-semibold text-gray-700">Checklist Dokumen Badan Usaha</h2>
        {action}
      </div>
      {children}
    </div>
  );
}

function ChecklistRow({
  item, projectId, canWrite, canReview, canDelete, onUpload, onReview,
}: {
  item: KpbuChecklistItem; projectId: string; canWrite: boolean; canReview: boolean; canDelete: boolean;
  onUpload: () => void; onReview: (d: KpbuDocument) => void;
}) {
  const [open, setOpen] = useState(false);
  const r = item.requirement;
  return (
    <li className="px-3 py-2.5">
      <div className="flex flex-wrap items-start gap-2">
        <span className="mt-0.5 w-7 shrink-0 text-xs font-semibold text-gray-400">{r.code}</span>
        <div className="min-w-0 flex-1">
          <p className="text-sm text-gray-800">
            {r.name}
            {!r.is_mandatory && <span className="ml-1 text-xs text-gray-400">(opsional)</span>}
          </p>
          <div className="mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-400">
            {r.legal_basis && <span>{r.legal_basis}</span>}
            {r.required_before_stage && <span>Syarat masuk: {KPBU_STAGE_LABEL[r.required_before_stage]}</span>}
            <SourceBadge req={r} />
          </div>
        </div>
        <StatusBadge status={item.status} />
        {item.documents.length > 0 && (
          <button className="text-xs text-blue-700 hover:underline" onClick={() => setOpen((v) => !v)}>
            {open ? "Tutup" : `${item.documents.length} file`}
          </button>
        )}
        {canWrite && item.status !== "APPROVED" && (
          <button className="rounded p-1 text-gray-500 hover:bg-gray-100" onClick={onUpload} aria-label={`Unggah ${r.code}`}>
            <Upload className="h-4 w-4" />
          </button>
        )}
      </div>
      {open && (
        <ul className="mt-2 space-y-2 pl-9">
          {item.documents.map((d) => (
            <li key={d.id}><DocLine doc={d} projectId={projectId} canReview={canReview} canDelete={canDelete} onReview={onReview} /></li>
          ))}
        </ul>
      )}
    </li>
  );
}

function DocLine({ doc, projectId, canReview, canDelete, onReview }: { doc: KpbuDocument; projectId: string; canReview: boolean; canDelete: boolean; onReview: (d: KpbuDocument) => void }) {
  const qc = useQueryClient();
  const [err, setErr] = useState<string | null>(null);
  const del = useMutation({
    mutationFn: () => kpbuService.deleteDocument(projectId, doc.id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["kpbu", "checklist", projectId] }),
    onError: (e) => setErr(kpbuErrorMessage(e, "Gagal menghapus dokumen.")),
  });
  return (
    <div className="rounded-md bg-gray-50 px-2 py-1.5 text-xs">
      <div className="flex flex-wrap items-center gap-2">
        <span className="min-w-0 flex-1 truncate font-medium text-gray-800">{doc.name}</span>
        <StatusBadge status={doc.status} />
        <span className="text-gray-400">{fmtSize(doc.file_size)} · {fmtDate(doc.uploaded_at)}</span>
        <button className="rounded p-1 text-gray-500 hover:bg-white" onClick={() => kpbuService.downloadDocument(projectId, doc).catch(() => setErr("Gagal mengunduh."))} aria-label="Unduh">
          <Download className="h-3.5 w-3.5" />
        </button>
        {canReview && doc.status === "SUBMITTED" && (
          <button className="rounded bg-blue-600 px-2 py-0.5 text-white hover:bg-blue-700" onClick={() => onReview(doc)}>Verifikasi</button>
        )}
        {canDelete && doc.status !== "APPROVED" && (
          <button className="rounded p-1 text-gray-400 hover:bg-white hover:text-red-600" onClick={() => del.mutate()} disabled={del.isPending} aria-label="Hapus">
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        )}
      </div>
      {(doc.document_number || doc.document_date) && (
        <p className="mt-0.5 text-gray-500">No. {doc.document_number || "—"} · Tgl. {fmtDate(doc.document_date)}</p>
      )}
      {doc.review_notes && <p className={cn("mt-0.5", doc.status === "REJECTED" ? "text-red-600" : "text-gray-500")}>Catatan verifikasi: {doc.review_notes}</p>}
      {err && <p className="mt-0.5 text-red-600">{err}</p>}
    </div>
  );
}

function UploadModal({ projectId, requirement, onClose }: { projectId: string; requirement: KpbuDocumentRequirement | null; onClose: () => void }) {
  const qc = useQueryClient();
  const fileRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [v, setV] = useState({ name: "", document_number: "", document_date: "", notes: "" });
  const [error, setError] = useState<string | null>(null);
  const set = (k: keyof typeof v) => (e: { target: { value: string } }) => setV((s) => ({ ...s, [k]: e.target.value }));

  const upload = useMutation({
    mutationFn: async (): Promise<void> => {
      if (!file) throw new Error("no file");
      await kpbuService.uploadDocument(projectId, {
        file,
        requirement_id: requirement?.id,
        name: v.name.trim() || undefined,
        document_number: v.document_number.trim() || undefined,
        document_date: v.document_date || undefined,
        notes: v.notes.trim() || undefined,
      });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["kpbu", "checklist", projectId] });
      onClose();
    },
    onError: (e) => setError(kpbuErrorMessage(e, "Gagal mengunggah dokumen.")),
  });

  return (
    <Modal title={requirement ? `Unggah — ${requirement.code}` : "Unggah Dokumen Lain"} onClose={onClose}>
      <form
        className="grid grid-cols-2 gap-3"
        onSubmit={(e) => {
          e.preventDefault();
          if (!file) {
            setError("Pilih file terlebih dahulu.");
            return;
          }
          upload.mutate();
        }}
      >
        {requirement && <p className="col-span-2 rounded-md bg-gray-50 p-2 text-xs text-gray-600">{requirement.name}</p>}
        <Field label="File (PDF, Office, gambar; maks. 20 MB)" required className="col-span-2">
          <input ref={fileRef} type="file" className="block w-full text-sm" onChange={(e) => setFile(e.target.files?.[0] ?? null)} />
        </Field>
        <Field label="Nama dokumen" className="col-span-2">
          <input className={inputCls} value={v.name} onChange={set("name")} placeholder={file?.name ?? "Default: nama file"} maxLength={500} />
        </Field>
        <Field label="Nomor dokumen"><input className={inputCls} value={v.document_number} onChange={set("document_number")} maxLength={200} /></Field>
        <Field label="Tanggal dokumen"><input className={inputCls} type="date" value={v.document_date} onChange={set("document_date")} /></Field>
        <Field label="Catatan" className="col-span-2"><textarea className={textareaCls} rows={2} value={v.notes} onChange={set("notes")} /></Field>
        {error && <p className="col-span-2 text-sm text-red-600">{error}</p>}
        <div className="col-span-2 flex justify-end gap-2">
          <button type="button" className={btnSecondary} onClick={onClose}>Batal</button>
          <button type="submit" className={btnPrimary} disabled={upload.isPending}>{upload.isPending ? "Mengunggah..." : "Unggah"}</button>
        </div>
      </form>
    </Modal>
  );
}

function ReviewModal({ projectId, doc, onClose }: { projectId: string; doc: KpbuDocument; onClose: () => void }) {
  const qc = useQueryClient();
  const [notes, setNotes] = useState("");
  const [error, setError] = useState<string | null>(null);
  const review = useMutation({
    mutationFn: async (decision: "APPROVE" | "REJECT"): Promise<void> => {
      await kpbuService.reviewDocument(projectId, doc.id, decision, notes.trim());
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["kpbu", "checklist", projectId] });
      onClose();
    },
    onError: (e) => setError(kpbuErrorMessage(e, "Gagal menyimpan verifikasi.")),
  });
  return (
    <Modal title="Verifikasi Dokumen" onClose={onClose}>
      <div className="space-y-3">
        <div className="rounded-md bg-gray-50 p-2 text-sm">
          <p className="font-medium text-gray-800">{doc.name}</p>
          {doc.notes && <p className="text-xs text-gray-500">Catatan pengunggah: {doc.notes}</p>}
          <button className="mt-1 inline-flex items-center gap-1 text-xs text-blue-700 hover:underline" onClick={() => kpbuService.downloadDocument(projectId, doc)}>
            <Download className="h-3 w-3" /> Unduh untuk diperiksa
          </button>
        </div>
        <Field label="Catatan verifikasi (wajib bila ditolak)">
          <textarea className={textareaCls} rows={3} value={notes} onChange={(e) => setNotes(e.target.value)} />
        </Field>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex justify-end gap-2">
          <button
            className={btnSecondary}
            disabled={review.isPending}
            onClick={() => {
              if (!notes.trim()) {
                setError("Catatan wajib diisi saat menolak dokumen.");
                return;
              }
              review.mutate("REJECT");
            }}
          >
            <X className="h-4 w-4 text-red-600" /> Tolak
          </button>
          <button className={btnPrimary} disabled={review.isPending} onClick={() => review.mutate("APPROVE")}>
            <Check className="h-4 w-4" /> Setujui
          </button>
        </div>
      </div>
    </Modal>
  );
}

/** Banner for the 12-month financial-close deadline (Perpres 38/2015 Ps.42). */
export function FinancialCloseBanner({ status, deadline, daysLeft }: { status: string; deadline: string | null; daysLeft: number | null }) {
  if (status !== "WARNING" && status !== "OVERDUE") return null;
  const overdue = status === "OVERDUE";
  return (
    <div className={cn("flex items-start gap-2 rounded-xl border p-3 text-sm", overdue ? "border-red-200 bg-red-50 text-red-800" : "border-amber-200 bg-amber-50 text-amber-800")}>
      <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
      <p>
        {overdue
          ? `Financial close melewati batas 12 bulan sejak penandatanganan (batas ${fmtDate(deadline)}, lewat ${Math.abs(daysLeft ?? 0)} hari).`
          : `Batas financial close ${fmtDate(deadline)} — sisa ${daysLeft} hari.`}{" "}
        Perpres 38/2015 Ps.42 mewajibkan BUP memperoleh pembiayaan paling lambat 12 bulan setelah Perjanjian KPBU ditandatangani.
      </p>
    </div>
  );
}
