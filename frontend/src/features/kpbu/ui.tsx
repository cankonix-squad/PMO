"use client";

import type { ReactNode } from "react";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";
import { KPBU_STAGE_LABEL, type KpbuStage } from "@/types/kpbu";

export const inputCls =
  "h-9 w-full rounded-md border border-gray-200 bg-white px-3 text-sm placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500";
export const textareaCls =
  "w-full rounded-md border border-gray-200 bg-white px-3 py-2 text-sm placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500";
export const btnPrimary =
  "inline-flex items-center gap-2 rounded-lg bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50";
export const btnSecondary =
  "inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50";

export function Field({ label, required, children, className }: { label: string; required?: boolean; children: ReactNode; className?: string }) {
  return (
    <label className={cn("block space-y-1", className)}>
      <span className="text-xs font-medium text-gray-600">
        {label}
        {required && <span className="text-red-500"> *</span>}
      </span>
      {children}
    </label>
  );
}

export function Modal({ title, onClose, children, wide }: { title: string; onClose: () => void; children: ReactNode; wide?: boolean }) {
  return (
    <div className="fixed inset-0 z-[2000] flex items-start justify-center overflow-y-auto bg-black/40 p-4" role="dialog" aria-modal="true" aria-label={title}>
      <div className={cn("my-8 w-full rounded-xl bg-white shadow-xl", wide ? "max-w-3xl" : "max-w-lg")}>
        <div className="flex items-center justify-between border-b border-gray-100 px-5 py-3">
          <h2 className="text-base font-semibold text-gray-900">{title}</h2>
          <button onClick={onClose} className="rounded p-1 text-gray-400 hover:bg-gray-100" aria-label="Tutup">
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="p-5">{children}</div>
      </div>
    </div>
  );
}

const STAGE_TONE: Partial<Record<KpbuStage, string>> = {
  IDENTIFIED: "bg-slate-100 text-slate-700",
  PLANNING: "bg-slate-100 text-slate-700",
  PREPARATION_OBC: "bg-indigo-50 text-indigo-700",
  PREPARATION_FBC: "bg-indigo-50 text-indigo-700",
  TRANSACTION: "bg-violet-50 text-violet-700",
  CONTRACT_SIGNED: "bg-blue-50 text-blue-700",
  FINANCIAL_CLOSE: "bg-blue-50 text-blue-700",
  CONSTRUCTION: "bg-amber-50 text-amber-700",
  OPERATION: "bg-green-50 text-green-700",
  HANDOVER: "bg-teal-50 text-teal-700",
  CLOSED: "bg-gray-100 text-gray-600",
  ON_HOLD: "bg-yellow-50 text-yellow-800",
  CANCELLED: "bg-red-50 text-red-700",
  TERMINATED: "bg-red-50 text-red-700",
};

export function StageBadge({ stage }: { stage: KpbuStage }) {
  return (
    <span className={cn("inline-flex whitespace-nowrap rounded-full px-2 py-0.5 text-xs font-medium", STAGE_TONE[stage] ?? "bg-gray-100 text-gray-700")}>
      {KPBU_STAGE_LABEL[stage] ?? stage}
    </span>
  );
}

/** Formats an ISO date as DD MMM YYYY (Indonesian). */
export function fmtDate(iso: string | null | undefined): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
}

/** ISO → YYYY-MM-DD for <input type="date">. */
export function toDateInput(iso: string | null | undefined): string {
  return iso ? iso.slice(0, 10) : "";
}

/** Compact Rupiah: Rp 1,2 T / Rp 850 M. */
export function fmtRupiahShort(n: number): string {
  if (n >= 1e12) return `Rp ${(n / 1e12).toLocaleString("id-ID", { maximumFractionDigits: 2 })} T`;
  if (n >= 1e9) return `Rp ${(n / 1e9).toLocaleString("id-ID", { maximumFractionDigits: 1 })} M`;
  return `Rp ${n.toLocaleString("id-ID")}`;
}
