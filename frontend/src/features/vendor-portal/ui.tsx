"use client";

import { cn } from "@/lib/utils";
import { SUBMISSION_STATUS_LABEL, type SubmissionStatus } from "@/types/vendor-portal";

export { Field, Modal, inputCls, textareaCls, btnPrimary, btnSecondary, fmtDate, fmtRupiahShort } from "@/features/kpbu/ui";

const TONE: Record<SubmissionStatus, string> = {
  SUBMITTED: "bg-amber-50 text-amber-800",
  VERIFIED: "bg-green-50 text-green-700",
  REJECTED: "bg-red-50 text-red-700",
};

export function StatusBadge({ status }: { status: SubmissionStatus }) {
  return (
    <span className={cn("inline-flex whitespace-nowrap rounded-full px-2 py-0.5 text-xs font-medium", TONE[status])}>
      {SUBMISSION_STATUS_LABEL[status] ?? status}
    </span>
  );
}

const MONTHS = ["Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"];

export function fmtPeriod(year: number | null | undefined, month: number | null | undefined): string {
  if (!year || !month) return "—";
  return `${MONTHS[month - 1]} ${year}`;
}

export function fmtBytes(n: number): string {
  if (n >= 1 << 20) return `${(n / (1 << 20)).toFixed(1)} MB`;
  if (n >= 1 << 10) return `${Math.round(n / (1 << 10))} KB`;
  return `${n} B`;
}

export const MONTH_OPTIONS = MONTHS.map((label, i) => ({ value: i + 1, label }));
