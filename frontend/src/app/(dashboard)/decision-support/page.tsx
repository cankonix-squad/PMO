"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AlertTriangle,
  BarChart3,
  ChevronDown,
  ChevronRight,
  Info,
  RefreshCw,
  TrendingUp,
  Zap,
} from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { priorityService } from "@/services/priority.service";
import { cn } from "@/lib/utils";
import type {
  Formula,
  ProjectScoreSummary,
  ScoreCategory,
  ScoreComponent,
} from "@/types/priority";
import {
  COMPONENT_KEY_LABELS,
  SCORE_CATEGORY_COLOR,
} from "@/types/priority";

// ── Category badge ────────────────────────────────────────────────────────────

function CategoryBadge({ category }: { category: ScoreCategory }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold",
        SCORE_CATEGORY_COLOR[category]
      )}
    >
      {category}
    </span>
  );
}

// ── Score bar ─────────────────────────────────────────────────────────────────

function ScoreBar({ value }: { value: number }) {
  const color =
    value >= 75
      ? "bg-rose-500"
      : value >= 50
        ? "bg-orange-400"
        : value >= 25
          ? "bg-yellow-400"
          : "bg-emerald-500";
  return (
    <div className="flex items-center gap-2">
      <div className="h-2 w-24 overflow-hidden rounded-full bg-slate-200">
        <div
          className={cn("h-full rounded-full transition-all", color)}
          style={{ width: `${Math.min(100, value)}%` }}
        />
      </div>
      <span className="w-10 text-right text-xs tabular-nums text-slate-700">
        {value.toFixed(1)}
      </span>
    </div>
  );
}

// ── Component explain row ─────────────────────────────────────────────────────

function ExplainRow({ comp }: { comp: ScoreComponent }) {
  return (
    <tr className={cn("text-sm", !comp.available && "opacity-50")}>
      <td className="py-1.5 pr-4 font-medium text-slate-700">
        {COMPONENT_KEY_LABELS[comp.component_key] ?? comp.component_key}
        {!comp.available && (
          <span className="ml-1 text-xs text-slate-400">(N/A)</span>
        )}
      </td>
      <td className="py-1.5 pr-4 tabular-nums text-slate-600">
        {comp.available && comp.raw_value != null ? `${comp.raw_value.toFixed(2)} ${comp.raw_unit ?? ""}` : "–"}
      </td>
      <td className="py-1.5 pr-4">
        <ScoreBar value={comp.normalized_score ?? 0} />
      </td>
      <td className="py-1.5 pr-4 tabular-nums text-slate-600">
        {(comp.weight * 100).toFixed(0)}%
      </td>
      <td className="py-1.5 tabular-nums font-semibold text-slate-800">
        {comp.weighted_score.toFixed(2)}
      </td>
      {comp.note && (
        <td className="py-1.5 pl-2 text-xs text-slate-400">{comp.note}</td>
      )}
    </tr>
  );
}

// ── Explain panel ─────────────────────────────────────────────────────────────

function ExplainPanel({ projectId }: { projectId: string }) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["priority-explain", projectId],
    queryFn: () => priorityService.explainProjectScore(projectId),
  });

  if (isLoading)
    return (
      <div className="px-4 py-3 text-sm text-muted-foreground">
        Memuat detail komponen…
      </div>
    );
  if (isError || !data)
    return (
      <div className="px-4 py-3 text-sm text-rose-600">
        Gagal memuat detail skor.
      </div>
    );

  const comps = data.components ?? [];
  return (
    <div className="overflow-x-auto px-4 py-3">
      <table className="w-full min-w-[560px]">
        <thead>
          <tr className="border-b text-xs font-semibold uppercase tracking-wide text-slate-500">
            <th className="pb-1.5 pr-4 text-left">Komponen</th>
            <th className="pb-1.5 pr-4 text-left">Nilai Mentah</th>
            <th className="pb-1.5 pr-4 text-left">Score Normal.</th>
            <th className="pb-1.5 pr-4 text-left">Bobot</th>
            <th className="pb-1.5 text-left">Weighted</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100">
          {comps.map((c) => (
            <ExplainRow key={c.component_key} comp={c} />
          ))}
        </tbody>
        <tfoot>
          <tr className="border-t font-bold text-slate-800">
            <td colSpan={4} className="pt-2 pr-4">
              Total Score
            </td>
            <td className="pt-2">{data.total_score.toFixed(2)}</td>
          </tr>
        </tfoot>
      </table>
    </div>
  );
}

// ── Project ranking row ───────────────────────────────────────────────────────

function ProjectRow({
  item,
  expanded,
  onToggle,
  onRecalculate,
  isRecalculating,
}: {
  item: ProjectScoreSummary;
  expanded: boolean;
  onToggle: () => void;
  onRecalculate: (projectId: string) => void;
  isRecalculating: boolean;
}) {
  return (
    <>
      <tr
        className="cursor-pointer border-b transition-colors hover:bg-slate-50"
        onClick={onToggle}
      >
        <td className="py-3 pl-4 pr-2 text-sm font-semibold tabular-nums text-slate-500">
          #{item.rank_in_org}
        </td>
        <td className="py-3 pr-3">
          <div className="flex items-center gap-2">
            {expanded ? (
              <ChevronDown className="h-4 w-4 shrink-0 text-slate-400" />
            ) : (
              <ChevronRight className="h-4 w-4 shrink-0 text-slate-400" />
            )}
            <span className="text-sm font-medium text-slate-800">
              {item.project_name}
            </span>
            <span className="text-xs text-slate-400">{item.project_code}</span>
          </div>
        </td>
        <td className="py-3 pr-4">
          <ScoreBar value={item.total_score} />
        </td>
        <td className="py-3 pr-4">
          <CategoryBadge category={item.score_category} />
        </td>
        <td className="py-3 pr-4 text-xs text-slate-400">
          {new Date(item.calculated_at).toLocaleDateString("id-ID", {
            day: "2-digit",
            month: "short",
            year: "numeric",
          })}
        </td>
        <td className="py-3 pr-4">
          <button
            type="button"
            aria-label="Hitung ulang skor"
            disabled={isRecalculating}
            onClick={(e) => {
              e.stopPropagation();
              onRecalculate(item.project_id);
            }}
            className="inline-flex items-center gap-1 rounded px-2 py-1 text-xs text-slate-500 hover:bg-slate-100 disabled:opacity-50"
          >
            <RefreshCw className={cn("h-3 w-3", isRecalculating && "animate-spin")} />
            Hitung Ulang
          </button>
        </td>
      </tr>
      {expanded && (
        <tr className="border-b bg-slate-50">
          <td colSpan={6} className="p-0">
            <ExplainPanel projectId={item.project_id} />
          </td>
        </tr>
      )}
    </>
  );
}

// ── Summary cards ─────────────────────────────────────────────────────────────

const CATEGORIES: ScoreCategory[] = ["CRITICAL", "HIGH", "MEDIUM", "LOW"];
const CATEGORY_ICONS: Record<ScoreCategory, typeof Zap> = {
  CRITICAL: Zap,
  HIGH: AlertTriangle,
  MEDIUM: TrendingUp,
  LOW: BarChart3,
};

function SummaryCards({ scores }: { scores: ProjectScoreSummary[] }) {
  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
      {CATEGORIES.map((cat) => {
        const count = scores.filter((s) => s.score_category === cat).length;
        const Icon = CATEGORY_ICONS[cat];
        return (
          <div
            key={cat}
            className="flex items-center gap-3 rounded-lg border bg-white px-4 py-3 shadow-sm"
          >
            <div
              className={cn(
                "grid h-9 w-9 shrink-0 place-items-center rounded-full",
                SCORE_CATEGORY_COLOR[cat]
              )}
            >
              <Icon className="h-4 w-4" />
            </div>
            <div>
              <p className="text-2xl font-bold tabular-nums leading-none text-slate-800">
                {count}
              </p>
              <p className="mt-0.5 text-xs text-slate-500">{cat}</p>
            </div>
          </div>
        );
      })}
    </div>
  );
}

// ── Formula selector ──────────────────────────────────────────────────────────

function FormulaSelector({
  formulas,
  selected,
  onChange,
}: {
  formulas: Formula[];
  selected: string;
  onChange: (id: string) => void;
}) {
  return (
    <div className="flex items-center gap-2">
      <label htmlFor="formula-select" className="text-sm text-slate-600">
        Formula:
      </label>
      <select
        id="formula-select"
        value={selected}
        onChange={(e) => onChange(e.target.value)}
        className="rounded-md border bg-white px-3 py-1.5 text-sm shadow-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
      >
        <option value="">— Aktif saat ini —</option>
        {formulas.map((f) => (
          <option key={f.id} value={f.id}>
            {f.name} v{f.version} [{f.status}]
          </option>
        ))}
      </select>
    </div>
  );
}

// ── Main page ─────────────────────────────────────────────────────────────────

export default function DecisionSupportPage() {
  const queryClient = useQueryClient();
  const [selectedFormula, setSelectedFormula] = useState("");
  const [expandedRow, setExpandedRow] = useState<string | null>(null);
  const [recalculatingId, setRecalculatingId] = useState<string | null>(null);

  const formulasQuery = useQuery({
    queryKey: ["priority-formulas"],
    queryFn: priorityService.listFormulas,
  });

  const rankingQuery = useQuery({
    queryKey: ["priority-ranking", selectedFormula],
    queryFn: () => priorityService.listRanking(selectedFormula || undefined),
  });

  const batchCalc = useMutation({
    mutationFn: () =>
      priorityService.batchCalculate({
        formula_id: selectedFormula || undefined,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["priority-ranking"] });
      await queryClient.invalidateQueries({ queryKey: ["priority-explain"] });
    },
  });

  const singleCalc = useMutation({
    mutationFn: (projectId: string) =>
      priorityService.calculate({
        project_id: projectId,
        formula_id: selectedFormula || undefined,
      }),
    onSuccess: async (_, projectId) => {
      await queryClient.invalidateQueries({ queryKey: ["priority-ranking"] });
      await queryClient.invalidateQueries({
        queryKey: ["priority-explain", projectId],
      });
      setRecalculatingId(null);
    },
    onError: () => setRecalculatingId(null),
  });

  const scores = rankingQuery.data?.projects ?? [];
  const formulas = formulasQuery.data ?? [];

  function handleRecalculate(projectId: string) {
    setRecalculatingId(projectId);
    singleCalc.mutate(projectId);
  }

  function toggleRow(id: string) {
    setExpandedRow((prev) => (prev === id ? null : id));
  }

  return (
    <DashboardLayout title="Decision Support — Priority Scoring">
      {/* Header actions */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">
          Peringkat prioritas proyek berdasarkan formula skor multi-kriteria yang dapat dikonfigurasi.
        </p>
        <div className="flex items-center gap-2">
          <FormulaSelector
            formulas={formulas}
            selected={selectedFormula}
            onChange={setSelectedFormula}
          />
          <button
            type="button"
            onClick={() => batchCalc.mutate()}
            disabled={batchCalc.isPending}
            className="inline-flex items-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-60"
          >
            <RefreshCw
              className={cn("h-4 w-4", batchCalc.isPending && "animate-spin")}
            />
            Hitung Semua
          </button>
        </div>
      </div>

      {/* Batch summary */}
      {batchCalc.isSuccess && (
        <div className="mt-3 rounded-md border border-blue-100 bg-blue-50 px-4 py-2 text-sm text-blue-800">
          Dihitung: <strong>{batchCalc.data.calculated}</strong> proyek
          {batchCalc.data.skipped > 0 && (
            <span className="ml-2 text-yellow-700">· {batchCalc.data.skipped} dilewati</span>
          )}
        </div>
      )}

      {/* Summary cards */}
      {scores.length > 0 && (
        <div className="mt-5">
          <SummaryCards scores={scores} />
        </div>
      )}

      {/* Formula info — derived from first project in ranking */}
      {scores.length > 0 && (
        <div className="mt-4 flex items-center gap-2 text-xs text-slate-500">
          <Info className="h-3.5 w-3.5" />
          Formula:{" "}
          <span className="font-medium text-slate-700">
            {scores[0].formula_name}
          </span>
          · v{scores[0].formula_version}
        </div>
      )}

      {/* Ranking table */}
      <div className="mt-4 overflow-x-auto rounded-lg border bg-white shadow-sm">
        {rankingQuery.isLoading && (
          <p className="p-6 text-center text-sm text-muted-foreground">
            Memuat ranking…
          </p>
        )}
        {rankingQuery.isError && (
          <p className="p-6 text-center text-sm text-rose-600">
            Gagal memuat data ranking. Pastikan sudah ada formula aktif dan minimal satu proyek.
          </p>
        )}
        {rankingQuery.isSuccess && scores.length === 0 && (
          <p className="rounded-md border border-dashed p-8 text-center text-sm text-muted-foreground">
            Belum ada skor prioritas. Klik{" "}
            <strong>Hitung Semua</strong> untuk memulai.
          </p>
        )}
        {scores.length > 0 && (
          <table className="w-full min-w-[640px]">
            <thead className="border-b bg-slate-50">
              <tr className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                <th className="py-2.5 pl-4 pr-2 text-left">Rank</th>
                <th className="py-2.5 pr-3 text-left">Proyek</th>
                <th className="py-2.5 pr-4 text-left">Score</th>
                <th className="py-2.5 pr-4 text-left">Kategori</th>
                <th className="py-2.5 pr-4 text-left">Dihitung</th>
                <th className="py-2.5 pr-4 text-left">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {scores.map((item) => (
                <ProjectRow
                  key={item.project_id}
                  item={item}
                  expanded={expandedRow === item.project_id}
                  onToggle={() => toggleRow(item.project_id)}
                  onRecalculate={handleRecalculate}
                  isRecalculating={
                    recalculatingId === item.project_id && singleCalc.isPending
                  }
                />
              ))}
            </tbody>
          </table>
        )}
      </div>
    </DashboardLayout>
  );
}
