"use client";

import { useCallback, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AlertCircle,
  CheckCircle2,
  Download,
  FileImage,
  Loader2,
  MapPin,
  Navigation,
  Plus,
  RefreshCw,
  Trash2,
  XCircle,
} from "lucide-react";
import { fieldService } from "@/services/field.service";
import type { VerificationStatus } from "@/types/field";
import { cn, formatDate } from "@/lib/utils";

// ─── Constants ───────────────────────────────────────────────────────────────

const ALLOWED_TYPES = [
  "image/jpeg",
  "image/png",
  "application/pdf",
  "text/plain",
];
const ALLOWED_EXTS = ".jpg,.jpeg,.png,.pdf,.txt";
const MAX_SIZE_MB = 20;
const MAX_SIZE_BYTES = MAX_SIZE_MB * 1024 * 1024;

const VERIFICATION_BADGE: Record<VerificationStatus, string> = {
  PENDING: "bg-amber-100 text-amber-700 border-amber-200",
  VERIFIED: "bg-emerald-100 text-emerald-700 border-emerald-200",
  REJECTED: "bg-rose-100 text-rose-700 border-rose-200",
};

const VERIFICATION_LABEL: Record<VerificationStatus, string> = {
  PENDING: "Menunggu",
  VERIFIED: "Terverifikasi",
  REJECTED: "Ditolak",
};

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function validateFile(file: File): string | null {
  if (!ALLOWED_TYPES.includes(file.type)) {
    return `Tipe file tidak diizinkan. Gunakan: JPG, PNG, PDF, atau TXT.`;
  }
  if (file.size > MAX_SIZE_BYTES) {
    return `Ukuran file terlalu besar (${formatBytes(file.size)}). Maks ${MAX_SIZE_MB} MB.`;
  }
  return null;
}

// ─── Sub-components ───────────────────────────────────────────────────────────

function StatusBadge({ status }: { status: VerificationStatus }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium",
        VERIFICATION_BADGE[status]
      )}
    >
      {VERIFICATION_LABEL[status]}
    </span>
  );
}

interface FilePickerProps {
  file: File | undefined;
  onFile: (f: File | undefined) => void;
  error: string | null;
  onError: (e: string | null) => void;
}

function FilePicker({ file, onFile, error, onError }: FilePickerProps) {
  const inputRef = useRef<HTMLInputElement>(null);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const picked = e.target.files?.[0];
    if (!picked) return;
    const err = validateFile(picked);
    if (err) {
      onError(err);
      onFile(undefined);
      if (inputRef.current) inputRef.current.value = "";
      return;
    }
    onError(null);
    onFile(picked);
  };

  const clear = () => {
    onFile(undefined);
    onError(null);
    if (inputRef.current) inputRef.current.value = "";
  };

  return (
    <div className="space-y-1.5">
      <label className="block text-sm font-medium text-foreground">
        Foto / Berkas Bukti
        <span className="ml-1 text-xs font-normal text-muted-foreground">
          (opsional, maks {MAX_SIZE_MB} MB)
        </span>
      </label>

      {!file ? (
        <button
          type="button"
          onClick={() => inputRef.current?.click()}
          className={cn(
            "flex w-full items-center justify-center gap-2 rounded-lg border-2 border-dashed p-4 text-sm transition-colors",
            error
              ? "border-rose-300 bg-rose-50 text-rose-600"
              : "border-border text-muted-foreground hover:border-primary/50 hover:bg-muted/30"
          )}
        >
          <FileImage className="h-4 w-4 shrink-0" />
          <span>Pilih file</span>
          <span className="text-xs opacity-70">{ALLOWED_EXTS}</span>
        </button>
      ) : (
        <div className="flex items-center gap-2 rounded-lg border bg-muted/30 px-3 py-2">
          <FileImage className="h-4 w-4 shrink-0 text-primary" />
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium">{file.name}</p>
            <p className="text-xs text-muted-foreground">
              {formatBytes(file.size)} · {file.type || "unknown"}
            </p>
          </div>
          <button
            type="button"
            onClick={clear}
            aria-label="Hapus file yang dipilih"
            className="shrink-0 rounded p-0.5 hover:bg-destructive/10 hover:text-destructive"
          >
            <XCircle className="h-4 w-4" />
          </button>
        </div>
      )}

      {error && (
        <p className="flex items-center gap-1 text-xs text-rose-600">
          <AlertCircle className="h-3.5 w-3.5 shrink-0" />
          {error}
        </p>
      )}

      <input
        ref={inputRef}
        type="file"
        accept={ALLOWED_EXTS}
        onChange={handleChange}
        className="sr-only"
        aria-label="Upload file bukti"
      />
    </div>
  );
}

interface GeoInputProps {
  lat: string;
  lon: string;
  onLat: (v: string) => void;
  onLon: (v: string) => void;
  geoState: "idle" | "loading" | "error";
  geoError: string | null;
  onUseLocation: () => void;
}

function GeoInput({
  lat,
  lon,
  onLat,
  onLon,
  geoState,
  geoError,
  onUseLocation,
}: GeoInputProps) {
  return (
    <div className="space-y-1.5">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <label className="block text-sm font-medium text-foreground">
          Koordinat GPS
          <span className="ml-1 text-xs font-normal text-muted-foreground">
            (opsional)
          </span>
        </label>
        {"geolocation" in navigator && (
          <button
            type="button"
            onClick={onUseLocation}
            disabled={geoState === "loading"}
            className={cn(
              "inline-flex items-center gap-1.5 rounded-md border px-2.5 py-1 text-xs font-medium transition-colors",
              geoState === "loading"
                ? "cursor-not-allowed opacity-60"
                : "hover:bg-muted active:bg-muted/80"
            )}
          >
            {geoState === "loading" ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Navigation className="h-3.5 w-3.5" />
            )}
            {geoState === "loading" ? "Mengambil..." : "Gunakan Lokasi Saat Ini"}
          </button>
        )}
      </div>

      <div className="grid grid-cols-2 gap-2">
        <div>
          <label className="mb-1 block text-xs text-muted-foreground">
            Latitude
          </label>
          <input
            type="number"
            step="any"
            value={lat}
            onChange={(e) => onLat(e.target.value)}
            placeholder="-6.2088"
            className="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-muted-foreground">
            Longitude
          </label>
          <input
            type="number"
            step="any"
            value={lon}
            onChange={(e) => onLon(e.target.value)}
            placeholder="106.8456"
            className="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>
      </div>

      {geoState === "error" && geoError && (
        <p className="flex items-center gap-1 text-xs text-amber-600">
          <AlertCircle className="h-3.5 w-3.5 shrink-0" />
          {geoError} — isi koordinat secara manual.
        </p>
      )}
    </div>
  );
}

// ─── Main Panel ───────────────────────────────────────────────────────────────

export function FieldInspectionPanel({ projectId }: { projectId: string }) {
  const qc = useQueryClient();

  // Form state
  const [open, setOpen] = useState(false);
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 16));
  const [notes, setNotes] = useState("");
  const [file, setFile] = useState<File | undefined>();
  const [fileError, setFileError] = useState<string | null>(null);
  const [lat, setLat] = useState("");
  const [lon, setLon] = useState("");
  const [formError, setFormError] = useState<string | null>(null);

  // Geolocation state
  const [geoState, setGeoState] = useState<"idle" | "loading" | "error">("idle");
  const [geoError, setGeoError] = useState<string | null>(null);

  // Download in-flight tracking
  const [downloadingId, setDownloadingId] = useState<string | null>(null);

  // Data query
  const query = useQuery({
    queryKey: ["projects", projectId, "inspections"],
    queryFn: () => fieldService.list(projectId),
  });

  const invalidate = useCallback(() => {
    void qc.invalidateQueries({ queryKey: ["projects", projectId, "inspections"] });
  }, [qc, projectId]);

  // Create mutation
  const create = useMutation({
    mutationFn: () => {
      const latN = lat !== "" ? parseFloat(lat) : undefined;
      const lonN = lon !== "" ? parseFloat(lon) : undefined;
      return fieldService.create(projectId, {
        inspected_at: new Date(date).toISOString(),
        notes: notes.trim() || undefined,
        latitude: latN,
        longitude: lonN,
        file,
      });
    },
    onSuccess: () => { setOpen(false); resetForm(); invalidate(); },
    onError: (e: Error) => setFormError(e.message),
  });

  const verify = useMutation({
    mutationFn: ({ id, status }: { id: string; status: VerificationStatus }) =>
      fieldService.verify(projectId, id, status),
    onSuccess: invalidate,
  });

  const remove = useMutation({
    mutationFn: (id: string) => fieldService.remove(projectId, id),
    onSuccess: invalidate,
  });

  const handleDownload = useCallback(
    async (inspectionId: string, evidenceId: string, fileName: string) => {
      setDownloadingId(evidenceId);
      try {
        const blob = await fieldService.download(projectId, inspectionId, evidenceId);
        const url = URL.createObjectURL(blob);
        const anchor = document.createElement("a");
        anchor.href = url;
        anchor.download = fileName || `evidence-${evidenceId}`;
        anchor.click();
        URL.revokeObjectURL(url);
      } catch {
        setFormError("Gagal mengunduh berkas. Coba lagi.");
      } finally {
        setDownloadingId(null);
      }
    },
    [projectId]
  );

  const handleUseLocation = useCallback(() => {
    setGeoState("loading");
    setGeoError(null);
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setLat(pos.coords.latitude.toFixed(6));
        setLon(pos.coords.longitude.toFixed(6));
        setGeoState("idle");
      },
      (err) => {
        setGeoState("error");
        const msgs: Record<number, string> = {
          1: "Izin lokasi ditolak",
          2: "Posisi tidak tersedia",
          3: "Waktu habis",
        };
        setGeoError(msgs[err.code] ?? "Gagal mendapatkan lokasi");
      },
      { timeout: 10000, enableHighAccuracy: true }
    );
  }, []);

  const resetForm = () => {
    setDate(new Date().toISOString().slice(0, 16));
    setNotes("");
    setFile(undefined);
    setFileError(null);
    setLat("");
    setLon("");
    setFormError(null);
    setGeoState("idle");
    setGeoError(null);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!date) { setFormError("Tanggal inspeksi wajib diisi."); return; }
    if (fileError) { setFormError("Perbaiki error file terlebih dahulu."); return; }
    setFormError(null);
    create.mutate();
  };

  const inspections = query.data ?? [];

  return (
    <section className="mt-6" aria-label="Field Inspection &amp; Evidence">
      {/* ── Header ── */}
      <div className="mb-3 flex flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <h3 className="text-sm font-semibold leading-tight sm:text-base">
            Inspeksi Lapangan &amp; Bukti
          </h3>
          <p className="mt-0.5 text-xs text-muted-foreground">
            Foto/berkas tersimpan dengan checksum SHA-256 dan status verifikasi.
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <button
            type="button"
            onClick={() => void query.refetch()}
            disabled={query.isFetching}
            aria-label="Refresh daftar inspeksi"
            className="inline-flex items-center gap-1 rounded-md border px-2.5 py-1.5 text-xs hover:bg-muted disabled:opacity-50"
          >
            <RefreshCw className={cn("h-3.5 w-3.5", query.isFetching && "animate-spin")} />
            <span className="hidden sm:inline">Muat Ulang</span>
          </button>
          <button
            type="button"
            onClick={() => { resetForm(); setOpen(true); }}
            className="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground hover:bg-primary/90 active:bg-primary/80"
          >
            <Plus className="h-3.5 w-3.5" />
            Tambah Inspeksi
          </button>
        </div>
      </div>
      {/* ── Create Form ── */}
      {open && (
        <div className="mb-4 rounded-lg border bg-muted/20 p-4 shadow-sm">
          <div className="mb-3 flex items-center justify-between">
            <h4 className="text-sm font-semibold">Inspeksi Baru</h4>
            <button
              type="button"
              onClick={() => { setOpen(false); resetForm(); }}
              aria-label="Tutup form"
              className="rounded p-1 hover:bg-muted"
            >
              <XCircle className="h-4 w-4 text-muted-foreground" />
            </button>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="fi-date" className="mb-1 block text-sm font-medium">
                Tanggal &amp; Waktu Inspeksi <span className="text-rose-500">*</span>
              </label>
              <input
                id="fi-date"
                type="datetime-local"
                value={date}
                onChange={(e) => setDate(e.target.value)}
                required
                className="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>

            <div>
              <label htmlFor="fi-notes" className="mb-1 block text-sm font-medium">
                Catatan Lapangan
              </label>
              <textarea
                id="fi-notes"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={3}
                placeholder="Kondisi lapangan, temuan, catatan penting..."
                className="w-full resize-none rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>

            <GeoInput
              lat={lat}
              lon={lon}
              onLat={setLat}
              onLon={setLon}
              geoState={geoState}
              geoError={geoError}
              onUseLocation={handleUseLocation}
            />

            <FilePicker
              file={file}
              onFile={setFile}
              error={fileError}
              onError={setFileError}
            />

            {formError && (
              <p className="flex items-center gap-1.5 rounded-md bg-rose-50 px-3 py-2 text-sm text-rose-600">
                <AlertCircle className="h-4 w-4 shrink-0" />
                {formError}
              </p>
            )}

            <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <button
                type="button"
                onClick={() => { setOpen(false); resetForm(); }}
                className="rounded-md border px-4 py-2 text-sm hover:bg-muted"
              >
                Batal
              </button>
              <button
                type="submit"
                disabled={create.isPending}
                className="inline-flex items-center justify-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-60"
              >
                {create.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                {create.isPending ? "Menyimpan..." : "Simpan Inspeksi"}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* ── List States ── */}
      {query.isLoading && (
        <div className="flex items-center justify-center gap-2 py-8 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
          Memuat data inspeksi...
        </div>
      )}

      {query.isError && (
        <div className="flex items-center gap-2 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-600">
          <AlertCircle className="h-4 w-4 shrink-0" />
          Gagal memuat data. Coba refresh.
        </div>
      )}

      {!query.isLoading && !query.isError && inspections.length === 0 && (
        <div className="rounded-lg border border-dashed px-4 py-8 text-center">
          <MapPin className="mx-auto mb-2 h-8 w-8 text-muted-foreground/40" />
          <p className="text-sm font-medium text-muted-foreground">Belum ada inspeksi lapangan</p>
          <p className="mt-1 text-xs text-muted-foreground">
            Klik &quot;Tambah Inspeksi&quot; untuk mencatat kunjungan lapangan pertama.
          </p>
        </div>
      )}
      {/* ── Inspection List ── */}
      {inspections.length > 0 && (
        <div className="space-y-3">
          {inspections.map((inspection) => (
            <div
              key={inspection.id}
              className="overflow-hidden rounded-lg border bg-card shadow-sm"
            >
              {/* Inspection header row */}
              <div className="flex flex-wrap items-start gap-x-3 gap-y-1 border-b bg-muted/20 px-4 py-2.5">
                <div className="min-w-0 flex-1">
                  <p className="text-xs font-semibold text-foreground sm:text-sm">
                    {formatDate(inspection.inspected_at)}
                  </p>
                  {(inspection.latitude != null || inspection.longitude != null) && (
                    <p className="mt-0.5 flex items-center gap-1 text-xs text-muted-foreground">
                      <MapPin className="h-3 w-3 shrink-0" />
                      {inspection.latitude?.toFixed(5)}, {inspection.longitude?.toFixed(5)}
                    </p>
                  )}
                </div>
                <StatusBadge status={inspection.verification_status} />
              </div>

              {/* Notes */}
              {inspection.notes && (
                <div className="px-4 py-2">
                  <p className="line-clamp-3 text-xs text-muted-foreground">
                    {inspection.notes}
                  </p>
                </div>
              )}

              {/* Evidence list */}
              {inspection.evidence && inspection.evidence.length > 0 && (
                <div className="border-t px-4 py-2">
                  <p className="mb-1.5 text-xs font-medium text-muted-foreground">
                    Bukti ({inspection.evidence.length})
                  </p>
                  <div className="space-y-1.5">
                    {inspection.evidence.map((ev) => (
                      <div
                        key={ev.id}
                        className="flex flex-wrap items-center gap-2 rounded-md bg-muted/30 px-3 py-1.5"
                      >
                        <FileImage className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                        <span className="min-w-0 flex-1 truncate text-xs font-medium">
                          {ev.file_name}
                        </span>
                        <span className="shrink-0 text-xs text-muted-foreground">
                          {formatBytes(ev.file_size)}
                        </span>
                        <span
                          className="shrink-0 font-mono text-xs text-muted-foreground"
                          title={`SHA-256: ${ev.checksum_sha256}`}
                        >
                          {ev.checksum_sha256.slice(0, 8)}…
                        </span>
                        {(ev.latitude != null || ev.longitude != null) && (
                          <span className="flex items-center gap-0.5 text-xs text-muted-foreground">
                            <MapPin className="h-3 w-3" />
                            {ev.latitude?.toFixed(4)}, {ev.longitude?.toFixed(4)}
                          </span>
                        )}
                        <button
                          type="button"
                          onClick={() =>
                            void handleDownload(inspection.id, ev.id, ev.file_name)
                          }
                          disabled={downloadingId === ev.id}
                          aria-label={`Unduh ${ev.file_name}`}
                          className="inline-flex items-center gap-1 rounded border px-2 py-0.5 text-xs hover:bg-background disabled:opacity-50"
                        >
                          {downloadingId === ev.id ? (
                            <Loader2 className="h-3.5 w-3.5 animate-spin" />
                          ) : (
                            <Download className="h-3.5 w-3.5" />
                          )}
                          <span className="hidden sm:inline">Unduh</span>
                        </button>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Actions */}
              <div className="flex flex-wrap items-center gap-2 border-t px-4 py-2.5">
                {inspection.verification_status === "PENDING" && (
                  <>
                    <button
                      type="button"
                      onClick={() =>
                        verify.mutate({ id: inspection.id, status: "VERIFIED" })
                      }
                      disabled={verify.isPending}
                      className="inline-flex min-h-[36px] items-center gap-1.5 rounded-md bg-emerald-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-emerald-700 active:bg-emerald-800 disabled:opacity-60"
                    >
                      <CheckCircle2 className="h-3.5 w-3.5" />
                      Verifikasi
                    </button>
                    <button
                      type="button"
                      onClick={() =>
                        verify.mutate({ id: inspection.id, status: "REJECTED" })
                      }
                      disabled={verify.isPending}
                      className="inline-flex min-h-[36px] items-center gap-1.5 rounded-md border border-rose-200 px-3 py-1.5 text-xs font-medium text-rose-700 hover:bg-rose-50 active:bg-rose-100 disabled:opacity-60"
                    >
                      <XCircle className="h-3.5 w-3.5" />
                      Tolak
                    </button>
                  </>
                )}
                <button
                  type="button"
                  onClick={() => {
                    if (window.confirm("Hapus inspeksi ini beserta semua buktinya?")) {
                      remove.mutate(inspection.id);
                    }
                  }}
                  disabled={remove.isPending}
                  className="ml-auto inline-flex min-h-[36px] items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs hover:border-destructive/30 hover:bg-destructive/10 hover:text-destructive disabled:opacity-60"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                  <span className="hidden sm:inline">Hapus</span>
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Global error banner (download errors etc.) */}
      {formError && inspections.length > 0 && (
        <div className="mt-3 flex items-center justify-between gap-2 rounded-md bg-rose-50 px-3 py-2 text-sm text-rose-600">
          <span className="flex items-center gap-1.5">
            <AlertCircle className="h-4 w-4 shrink-0" />
            {formError}
          </span>
          <button
            type="button"
            onClick={() => setFormError(null)}
            className="shrink-0 text-xs underline"
          >
            Tutup
          </button>
        </div>
      )}
    </section>
  );
}
