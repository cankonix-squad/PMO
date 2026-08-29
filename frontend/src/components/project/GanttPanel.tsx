"use client";

/**
 * GanttPanel — P1-009
 *
 * Renders a read-only horizontal bar Gantt chart for tasks and milestones.
 * Uses pure CSS bars (no extra library). Data comes from existing
 * task/milestone queries already loaded in the project detail page.
 *
 * Layout:
 *   - Left column  : item label + status badge + progress %
 *   - Right column : scrollable bar canvas with date axis
 *
 * Rules:
 *   - Tasks without start_date use created_at as fallback start.
 *   - Tasks without due_date are shown as a 1-day point bar.
 *   - Milestones without due_date are hidden.
 *   - Overdue items (due_date < today, status not terminal) get a red border.
 *   - Today marker is drawn as a vertical red dashed line.
 */

import { useMemo } from "react";
import { cn } from "@/lib/utils";
import type { Task, TaskStatus, Milestone, MilestoneStatus } from "@/types/project";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface GanttRow {
  id: string;
  kind: "milestone" | "task";
  label: string;
  wbsCode?: string | null;
  status: string;
  progressPct: number;
  start: Date;
  end: Date;
  isOverdue: boolean;
}

// ---------------------------------------------------------------------------
// Status helpers
// ---------------------------------------------------------------------------

const TASK_TERMINAL: TaskStatus[] = ["DONE"];
const MILESTONE_TERMINAL: MilestoneStatus[] = ["COMPLETED"];

function taskStatusTone(status: TaskStatus): string {
  switch (status) {
    case "DONE":        return "bg-green-500";
    case "IN_REVIEW":   return "bg-blue-400";
    case "IN_PROGRESS": return "bg-blue-500";
    case "BLOCKED":     return "bg-red-500";
    default:            return "bg-slate-400"; // TODO / BACKLOG
  }
}

function milestoneStatusTone(status: MilestoneStatus): string {
  switch (status) {
    case "COMPLETED":   return "bg-green-500";
    case "IN_PROGRESS": return "bg-blue-500";
    case "DELAYED":     return "bg-orange-400";
    default:            return "bg-slate-400"; // PENDING
  }
}

function statusLabel(status: string): string {
  const labels: Record<string, string> = {
    TODO: "Belum mulai",
    IN_PROGRESS: "Berjalan",
    IN_REVIEW: "Review",
    DONE: "Selesai",
    BLOCKED: "Terhambat",
    PENDING: "Menunggu",
    COMPLETED: "Selesai",
    DELAYED: "Terlambat",
  };
  return labels[status] ?? status.replace(/_/g, " ");
}

// ---------------------------------------------------------------------------
// Date helpers
// ---------------------------------------------------------------------------

function parseDate(value: string | null | undefined, fallback: Date): Date {
  if (!value) return fallback;
  const d = new Date(value);
  return isNaN(d.getTime()) ? fallback : d;
}

function addDays(d: Date, n: number): Date {
  const copy = new Date(d);
  copy.setDate(copy.getDate() + n);
  return copy;
}

function startOfDay(d: Date): Date {
  const copy = new Date(d);
  copy.setHours(0, 0, 0, 0);
  return copy;
}

function daysBetween(a: Date, b: Date): number {
  return Math.round((b.getTime() - a.getTime()) / 86_400_000);
}

function formatAxisDate(d: Date): string {
  return d.toLocaleDateString("id-ID", { day: "numeric", month: "short" });
}

// ---------------------------------------------------------------------------
// Row builder
// ---------------------------------------------------------------------------

function buildRows(tasks: Task[], milestones: Milestone[]): GanttRow[] {
  const today = startOfDay(new Date());
  const rows: GanttRow[] = [];

  // Milestones first
  for (const m of milestones) {
    if (!m.due_date) continue;
    const end = startOfDay(parseDate(m.due_date, today));
    const start = addDays(end, -1); // milestones shown as 1-day diamond
    const isTerminal = MILESTONE_TERMINAL.includes(m.status as MilestoneStatus);
    rows.push({
      id: m.id,
      kind: "milestone",
      label: m.title,
      status: m.status,
      progressPct: m.progress_pct,
      start,
      end,
      isOverdue: !isTerminal && end < today,
    });
  }

  // Top-level tasks only (parent_id === null)
  for (const t of tasks) {
    if (t.parent_id !== null) continue;
    const fallbackStart = startOfDay(parseDate(t.created_at, today));
    const start = startOfDay(parseDate(t.start_date, fallbackStart));
    const end = t.due_date
      ? startOfDay(parseDate(t.due_date, addDays(start, 1)))
      : addDays(start, 1);
    const isTerminal = TASK_TERMINAL.includes(t.status as TaskStatus);
    rows.push({
      id: t.id,
      kind: "task",
      label: t.title,
      wbsCode: t.wbs_code,
      status: t.status,
      progressPct: t.progress_pct,
      start,
      end: end > start ? end : addDays(start, 1),
      isOverdue: !isTerminal && end < today,
    });
  }

  // Sort chronologically
  rows.sort((a, b) => a.start.getTime() - b.start.getTime());
  return rows;
}

// ---------------------------------------------------------------------------
// Axis ticks (every ~7 days)
// ---------------------------------------------------------------------------

function buildTicks(minDate: Date, totalDays: number): Date[] {
  const ticks: Date[] = [];
  const interval = totalDays <= 30 ? 7 : totalDays <= 90 ? 14 : 30;
  for (let i = 0; i <= totalDays; i += interval) {
    ticks.push(addDays(minDate, i));
  }
  return ticks;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

const PX_PER_DAY = 28; // pixels per day in the bar area
const LABEL_WIDTH = 220; // px — left label column

interface GanttPanelProps {
  tasks: Task[];
  milestones: Milestone[];
  isLoading?: boolean;
  isError?: boolean;
}

export function GanttPanel({ tasks, milestones, isLoading, isError }: GanttPanelProps) {
  const today = startOfDay(new Date());

  const rows = useMemo(() => buildRows(tasks, milestones), [tasks, milestones]);

  const { minDate, totalDays, ticks, todayOffset } = useMemo(() => {
    if (rows.length === 0) {
      const minDate = addDays(today, -7);
      const totalDays = 30;
      return {
        minDate,
        totalDays,
        ticks: buildTicks(minDate, totalDays),
        todayOffset: daysBetween(minDate, today),
      };
    }

    const earliest = rows.reduce(
      (min, r) => (r.start < min ? r.start : min),
      rows[0].start
    );
    const latest = rows.reduce(
      (max, r) => (r.end > max ? r.end : max),
      rows[0].end
    );

    const minDate = addDays(earliest, -3);
    const maxDate = addDays(latest, 5);
    const totalDays = Math.max(daysBetween(minDate, maxDate), 14);

    return {
      minDate,
      totalDays,
      ticks: buildTicks(minDate, totalDays),
      todayOffset: daysBetween(minDate, today),
    };
  }, [rows, today]);

  const canvasWidth = totalDays * PX_PER_DAY;
  const todayPx = todayOffset * PX_PER_DAY;

  if (isLoading) {
    return (
      <aside className="rounded-lg border border-border bg-card shadow-sm">
        <SectionHeaderSimple title="Timeline Proyek" />
        <div className="flex items-center justify-center py-10 text-sm text-muted-foreground">
          Memuat timeline...
        </div>
      </aside>
    );
  }

  if (isError) {
    return (
      <aside className="rounded-lg border border-border bg-card shadow-sm">
        <SectionHeaderSimple title="Timeline Proyek" />
        <div className="flex items-center justify-center py-10 text-sm text-destructive">
          Timeline belum dapat dimuat.
        </div>
      </aside>
    );
  }

  if (rows.length === 0) {
    return (
      <aside className="rounded-lg border border-border bg-card shadow-sm">
        <SectionHeaderSimple title="Timeline Proyek" />
        <div className="flex flex-col items-center justify-center px-6 py-10 text-center text-sm text-muted-foreground">
          <p className="font-medium text-foreground">Belum ada item bertanggal.</p>
          <p className="mt-1 max-w-md">
            Tambahkan tanggal mulai atau tanggal target pada task dan milestone agar
            timeline dapat ditampilkan.
          </p>
        </div>
      </aside>
    );
  }

  return (
    <aside className="rounded-lg border border-border bg-card shadow-sm">
      <SectionHeaderSimple title="Timeline Proyek" />

      <div className="overflow-x-auto">
        <div
          className="flex"
          style={{ minWidth: LABEL_WIDTH + canvasWidth }}
        >
          {/* ---- Label column ---- */}
          <div
            className="shrink-0 border-r border-border bg-card"
            style={{ width: LABEL_WIDTH }}
          >
            {/* Header spacer matching axis height */}
            <div className="h-8 border-b border-border" />

            {rows.map((row) => (
              <div
                key={row.id}
                className="flex h-10 items-center gap-2 border-b border-border px-3 last:border-b-0"
              >
                {row.kind === "milestone" && (
                  <span className="text-base leading-none text-amber-500" aria-label="Milestone">◆</span>
                )}
                <div className="min-w-0 flex-1">
                  <p
                    className={cn(
                      "truncate text-xs font-medium leading-tight",
                      row.isOverdue ? "text-destructive" : "text-foreground"
                    )}
                    title={row.label}
                  >
                    {row.wbsCode ? (
                      <span className="mr-1 text-muted-foreground">{row.wbsCode}</span>
                    ) : null}
                    {row.label}
                  </p>
                  <p className="truncate text-[10px] text-muted-foreground">
                    {statusLabel(row.status)}
                    {row.progressPct > 0 ? ` · ${row.progressPct}%` : ""}
                  </p>
                </div>
              </div>
            ))}
          </div>

          {/* ---- Bar canvas ---- */}
          <div className="relative flex-1 overflow-hidden" style={{ width: canvasWidth }}>
            {/* Date axis */}
            <div
              className="relative h-8 border-b border-border bg-muted/30"
              style={{ width: canvasWidth }}
            >
              {ticks.map((tick) => {
                const left = daysBetween(minDate, tick) * PX_PER_DAY;
                return (
                  <span
                    key={tick.toISOString()}
                    className="absolute top-1/2 -translate-x-1/2 -translate-y-1/2 whitespace-nowrap text-[10px] text-muted-foreground"
                    style={{ left }}
                  >
                    {formatAxisDate(tick)}
                  </span>
                );
              })}
            </div>

            {/* Today line */}
            {todayOffset >= 0 && todayOffset <= totalDays && (
              <div
                className="pointer-events-none absolute top-0 z-10 h-full w-px border-l-2 border-dashed border-red-500"
                style={{ left: todayPx }}
                title={`Hari ini: ${formatAxisDate(today)}`}
              />
            )}

            {/* Grid columns */}
            {ticks.map((tick) => {
              const left = daysBetween(minDate, tick) * PX_PER_DAY;
              return (
                <div
                  key={tick.toISOString()}
                  className="pointer-events-none absolute top-8 h-full w-px bg-border/50"
                  style={{ left }}
                />
              );
            })}

            {/* Bars */}
            {rows.map((row, idx) => {
              const barLeft = daysBetween(minDate, row.start) * PX_PER_DAY;
              const barWidth = Math.max(
                daysBetween(row.start, row.end) * PX_PER_DAY,
                row.kind === "milestone" ? 10 : 14
              );
              const barColor =
                row.kind === "milestone"
                  ? milestoneStatusTone(row.status as MilestoneStatus)
                  : taskStatusTone(row.status as TaskStatus);

              return (
                <div
                  key={row.id}
                  className="absolute flex items-center"
                  style={{
                    top: 32 + idx * 40 + 10, // 32px axis + row offset + center
                    left: barLeft,
                    width: barWidth,
                    height: 20,
                  }}
                >
                  {/* Background track */}
                  <div
                    className={cn(
                      "relative h-full w-full overflow-hidden rounded-sm",
                      row.kind === "milestone" ? "rounded-sm" : "rounded",
                      row.isOverdue
                        ? "ring-1 ring-destructive ring-offset-0"
                        : "",
                      "bg-muted"
                    )}
                    title={`${row.label}: ${formatAxisDate(row.start)} → ${formatAxisDate(row.end)}`}
                  >
                    {/* Progress fill */}
                    <div
                      className={cn("h-full transition-all", barColor, "opacity-80")}
                      style={{ width: `${row.progressPct}%` }}
                    />
                    {/* Base color overlay for zero-progress items */}
                    {row.progressPct === 0 && (
                      <div className={cn("absolute inset-0 opacity-40", barColor)} />
                    )}
                    {/* Label inside bar (if wide enough) */}
                    {barWidth > 50 && (
                      <span className="absolute inset-0 flex items-center pl-1 text-[9px] font-medium text-white drop-shadow">
                        {row.progressPct > 0 ? `${row.progressPct}%` : ""}
                      </span>
                    )}
                  </div>

                  {/* Milestone diamond indicator */}
                  {row.kind === "milestone" && (
                    <span className="ml-1 text-[10px] text-amber-500">◆</span>
                  )}
                </div>
              );
            })}

            {/* Row stripe backgrounds */}
            {rows.map((_, idx) => (
              <div
                key={idx}
                className={cn(
                  "pointer-events-none absolute w-full border-b border-border",
                  idx % 2 === 0 ? "bg-transparent" : "bg-muted/20"
                )}
                style={{
                  top: 32 + idx * 40,
                  height: 40,
                  width: canvasWidth,
                }}
              />
            ))}
          </div>
        </div>
      </div>

      {/* Legend */}
      <div className="flex flex-wrap items-center gap-4 border-t border-border px-4 py-2">
        <span className="text-xs text-muted-foreground">Legend:</span>
        {[
          { color: "bg-slate-400", label: "Todo/Backlog" },
          { color: "bg-blue-500", label: "In Progress" },
          { color: "bg-blue-400", label: "In Review" },
          { color: "bg-green-500", label: "Done/Completed" },
          { color: "bg-red-500", label: "Blocked" },
          { color: "bg-orange-400", label: "Delayed" },
        ].map(({ color, label }) => (
          <span key={label} className="flex items-center gap-1 text-xs text-muted-foreground">
            <span className={cn("inline-block h-2.5 w-4 rounded-sm opacity-80", color)} />
            {label}
          </span>
        ))}
        <span className="flex items-center gap-1 text-xs text-muted-foreground">
          <span className="inline-block h-3 w-0.5 border-l-2 border-dashed border-red-500" />
          Today
        </span>
        <span className="flex items-center gap-1 text-xs text-muted-foreground">
          <span className="text-amber-500">◆</span>
          Milestone
        </span>
        <span className="flex items-center gap-1 text-xs text-destructive">
          <span className="inline-block h-2.5 w-4 rounded-sm ring-1 ring-destructive" />
          Overdue
        </span>
      </div>
    </aside>
  );
}

// Simple header without action button
function SectionHeaderSimple({ title }: { title: string }) {
  return (
    <div className="flex items-center justify-between border-b border-border px-4 py-3">
      <h3 className="text-sm font-semibold text-foreground">{title}</h3>
    </div>
  );
}
