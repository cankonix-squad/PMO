"use client";

import { cn, formatDate } from "@/lib/utils";
import type { ProjectDocument } from "@/types/project";
import { Download, Edit3, Trash2 } from "lucide-react";
import { documentCategoryTone, formatFileSize } from "../utils";
import { EmptyState } from "./EmptyState";
import { ErrorState } from "./ErrorState";
import { LoadingState } from "./LoadingState";
import { SectionHeader } from "./SectionHeader";

export function DocumentPanel({
  documents,
  totals,
  isLoading,
  isError,
  onCreate,
  onEdit,
  onDownload,
  onDelete,
}: {
  documents: ProjectDocument[];
  totals: { count: number; totalSize: number; byCategory: Record<string, number> };
  isLoading: boolean;
  isError: boolean;
  onCreate: () => void;
  onEdit: (document: ProjectDocument) => void;
  onDownload: (document: ProjectDocument) => void;
  onDelete: (document: ProjectDocument) => void;
}) {
  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeader title="Dokumen" actionLabel="Upload" onAction={onCreate} />
      {isLoading ? (
        <LoadingState label="Memuat dokumen..." />
      ) : isError ? (
        <ErrorState label="Dokumen belum dapat dimuat." />
      ) : documents.length === 0 ? (
        <EmptyState label="Belum ada dokumen." />
      ) : (
        <div>
          <div className="grid grid-cols-2 gap-2 border-b border-border p-4 text-xs">
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Total Files</p>
              <p className="mt-0.5 font-semibold text-foreground">{totals.count}</p>
            </div>
            <div className="rounded-md bg-muted/40 p-2">
              <p className="text-muted-foreground">Total Size</p>
              <p className="mt-0.5 font-semibold text-foreground">
                {formatFileSize(totals.totalSize)}
              </p>
            </div>
          </div>
          <div className="divide-y divide-border">
            {documents.map((doc) => (
              <div key={doc.id} className="p-4">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium text-foreground">{doc.name}</p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {formatFileSize(doc.file_size ?? 0)}
                      {doc.category ? ` · ${doc.category.replace("_", " ")}` : ""}
                      {doc.version ? ` · ${doc.version}` : ""}
                      {" · "}
                      {formatDate(doc.created_at)}
                    </p>
                  </div>
                  <div className="flex flex-col items-end gap-2">
                    <span
                      className={cn(
                        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold",
                        documentCategoryTone(doc.category)
                      )}
                    >
                      {(doc.category ?? "OTHER").replace("_", " ")}
                    </span>
                    <div className="flex gap-1">
                      <button
                        type="button"
                        onClick={() => onDownload(doc)}
                        className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        aria-label="Download"
                      >
                        <Download className="h-4 w-4" aria-hidden="true" />
                      </button>
                      <button
                        type="button"
                        onClick={() => onEdit(doc)}
                        className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        aria-label="Edit"
                      >
                        <Edit3 className="h-4 w-4" aria-hidden="true" />
                      </button>
                      <button
                        type="button"
                        onClick={() => onDelete(doc)}
                        className="rounded-md p-2 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                        aria-label="Delete"
                      >
                        <Trash2 className="h-4 w-4" aria-hidden="true" />
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
      <p className="border-t border-border px-4 py-2 text-[10px] text-muted-foreground">
        File tersimpan lokal (storage/documents). Operational evidence input — belum
        validated/published (menunggu P1-012).
      </p>
    </aside>
  );
}
