"use client";

import { formatDate, formatIDR } from "@/lib/utils";
import type { Project, ProjectStatus } from "@/types/project";
import { CheckCircle2, Edit3, History } from "lucide-react";
import { NEXT_STATUS, PRIORITY_LABELS, STATUS_LABELS } from "../schemas";
import { sanitizeRichText } from "../utils";
import { StatusBadge } from "./StatusBadge";

export function ProjectDetail({
  project,
  isLoading,
  history,
  historyLoading,
  transitionPending,
  onEdit,
  onTransition,
}: {
  project: Project | null;
  isLoading: boolean;
  history: Array<{
    id: string;
    progress_pct: number;
    notes: string | null;
    recorded_at: string;
  }>;
  historyLoading: boolean;
  transitionPending: boolean;
  onEdit: (project: Project) => void;
  onTransition: (project: Project, status: ProjectStatus) => void;
}) {
  if (isLoading) {
    return (
      <aside className="rounded-lg border border-border bg-card p-5 shadow-sm">
        <div className="h-5 w-32 rounded-md bg-muted" />
        <div className="mt-4 h-24 rounded-md bg-muted" />
      </aside>
    );
  }

  if (!project) {
    return (
      <aside className="rounded-lg border border-border bg-card p-5 shadow-sm">
        <p className="text-sm text-muted-foreground">Pilih proyek untuk melihat detail.</p>
      </aside>
    );
  }

  return (
    <aside className="space-y-6 rounded-lg border border-border bg-card p-5 shadow-sm">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-medium uppercase text-muted-foreground">{project.code}</p>
          <h3 className="mt-1 text-lg font-semibold text-foreground">{project.name}</h3>
          <div className="mt-2 flex flex-wrap gap-2">
            <StatusBadge status={project.status} />
            <span className="inline-flex rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-muted-foreground">
              {PRIORITY_LABELS[project.priority]}
            </span>
          </div>
        </div>
        <button
          type="button"
          onClick={() => onEdit(project)}
          className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          aria-label="Edit proyek"
        >
          <Edit3 className="h-4 w-4" aria-hidden="true" />
        </button>
      </div>

      <div>
        <div className="mb-2 flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Progres</span>
          <span className="font-medium text-foreground">{project.progress_pct}%</span>
        </div>
        <div className="h-2 overflow-hidden rounded-full bg-muted">
          <div
            className="h-full rounded-full bg-primary"
            style={{ width: `${project.progress_pct}%` }}
          />
        </div>
      </div>

      <dl className="grid grid-cols-2 gap-4 text-sm">
        <div>
          <dt className="text-muted-foreground">Anggaran</dt>
          <dd className="font-medium text-foreground">{formatIDR(project.budget_total)}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Tanggal Selesai</dt>
          <dd className="font-medium text-foreground">{formatDate(project.end_date)}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Kategori</dt>
          <dd className="font-medium text-foreground">{project.category || "-"}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Diperbarui</dt>
          <dd className="font-medium text-foreground">{formatDate(project.updated_at)}</dd>
        </div>
      </dl>

      {project.description && (
        <div
          className="prose prose-sm max-w-none text-muted-foreground [&_ul]:list-disc [&_ul]:pl-4 [&_ol]:list-decimal [&_ol]:pl-4 [&_hr]:my-2"
          dangerouslySetInnerHTML={{ __html: sanitizeRichText(project.description) }}
        />
      )}

      {project.objectives && (
        <div>
          <h4 className="mb-1 text-sm font-semibold text-foreground">Tujuan</h4>
          <div
            className="prose prose-sm max-w-none text-muted-foreground [&_ul]:list-disc [&_ul]:pl-4 [&_ol]:list-decimal [&_ol]:pl-4 [&_hr]:my-2"
            dangerouslySetInnerHTML={{ __html: sanitizeRichText(project.objectives) }}
          />
        </div>
      )}

      <div>
        <h4 className="mb-2 text-sm font-semibold text-foreground">Transisi Status</h4>
        <div className="flex flex-wrap gap-2">
          {NEXT_STATUS[project.status].length > 0 ? (
            NEXT_STATUS[project.status].map((status) => (
              <button
                key={status}
                type="button"
                disabled={transitionPending}
                onClick={() => onTransition(project, status)}
                className="inline-flex h-8 items-center justify-center rounded-md border border-input px-3 text-xs font-medium hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
              >
                {STATUS_LABELS[status]}
              </button>
            ))
          ) : (
            <p className="text-sm text-muted-foreground">Tidak ada transisi lanjutan.</p>
          )}
        </div>
      </div>

      <div>
        <div className="mb-3 flex items-center gap-2">
          <History className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
          <h4 className="text-sm font-semibold text-foreground">Riwayat Progres</h4>
        </div>
        {historyLoading ? (
          <p className="text-sm text-muted-foreground">Memuat riwayat...</p>
        ) : history.length === 0 ? (
          <div className="flex items-start gap-3 rounded-md bg-muted/40 p-3">
            <CheckCircle2 className="mt-0.5 h-4 w-4 text-green-700" aria-hidden="true" />
            <p className="text-sm text-muted-foreground">Belum ada riwayat progres.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {history.map((item) => (
              <div key={item.id} className="rounded-md border border-border p-3">
                <div className="flex items-center justify-between text-sm">
                  <span className="font-medium text-foreground">{item.progress_pct}%</span>
                  <span className="text-xs text-muted-foreground">
                    {formatDate(item.recorded_at)}
                  </span>
                </div>
                {item.notes && (
                  <p className="mt-1 text-sm text-muted-foreground">{item.notes}</p>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </aside>
  );
}
