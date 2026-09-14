"use client";

import type { Snapshot, SnapshotStatus } from "@/types/monitoring";
import { Edit3, Trash2 } from "lucide-react";

export function SnapshotActions({
  snapshot,
  onEdit,
  onTransition,
  onReject,
  onDelete,
}: {
  snapshot: Snapshot;
  onEdit: (s: Snapshot) => void;
  onTransition: (s: Snapshot, status: SnapshotStatus) => void;
  onReject: (s: Snapshot, reason: string) => void;
  onDelete: (s: Snapshot) => void;
}) {
  const s = snapshot;
  return (
    <div className="flex items-center gap-1.5 flex-wrap">
      {s.status === "DRAFT" && (
        <>
          <button
            onClick={() => onEdit(s)}
            className="text-xs text-muted-foreground hover:text-foreground"
            title="Edit"
          >
            <Edit3 className="h-3.5 w-3.5" />
          </button>
          <button
            onClick={() => onTransition(s, "SUBMITTED")}
            className="rounded px-2 py-0.5 text-xs bg-blue-100 text-blue-700 hover:bg-blue-200"
          >
            Ajukan
          </button>
          <button
            onClick={() => onDelete(s)}
            className="text-muted-foreground hover:text-destructive"
            title="Hapus"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        </>
      )}
      {s.status === "SUBMITTED" && (
        <>
          <button
            onClick={() => onTransition(s, "VALID")}
            className="rounded px-2 py-0.5 text-xs bg-green-100 text-green-700 hover:bg-green-200"
          >
            Validasi
          </button>
          <button
            onClick={() => {
              const reason = window.prompt("Alasan penolakan:");
              if (reason !== null) onReject(s, reason);
            }}
            className="rounded px-2 py-0.5 text-xs bg-red-100 text-red-700 hover:bg-red-200"
          >
            Tolak
          </button>
        </>
      )}
      {(s.status === "VALID" || s.status === "REJECTED") && (
        <button
          onClick={() => onTransition(s, "STALE")}
          className="rounded px-2 py-0.5 text-xs bg-yellow-100 text-yellow-700 hover:bg-yellow-200"
        >
          Kedaluwarsa
        </button>
      )}
    </div>
  );
}
