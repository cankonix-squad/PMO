"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { AlertTriangle, Handshake, Landmark, Plus, Wallet } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { kpbuService } from "@/services/kpbu.service";
import { useAuthStore } from "@/store/auth.store";
import { ProjectFormModal } from "@/features/kpbu/ProjectFormModal";
import { btnPrimary, fmtRupiahShort, inputCls, StageBadge } from "@/features/kpbu/ui";
import { cn } from "@/lib/utils";
import {
  KPBU_MAIN_STAGES,
  KPBU_RETURN_SCHEME_LABEL,
  KPBU_SECTOR_LABEL,
  KPBU_STAGE_LABEL,
  type KpbuProject,
} from "@/types/kpbu";

const PRE_CONTRACT = new Set(["IDENTIFIED", "PLANNING", "PREPARATION_OBC", "PREPARATION_FBC", "TRANSACTION"]);
const WRITE_ROLES = ["SUPER_ADMIN", "ADMIN", "PMO", "PROJECT_MANAGER"];

export default function KpbuPage() {
  const user = useAuthStore((s) => s.user);
  const canWrite = (user?.roles ?? []).some((r) => WRITE_ROLES.includes(r));
  const [stage, setStage] = useState("");
  const [sector, setSector] = useState("");
  const [search, setSearch] = useState("");
  const [showForm, setShowForm] = useState(false);

  const { data: all = [], isLoading, isError, refetch } = useQuery({
    queryKey: ["kpbu", "list"],
    queryFn: () => kpbuService.list().then((r) => r.data ?? []),
  });

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return all.filter(
      (p) =>
        (!stage || p.stage === stage) &&
        (!sector || p.sector === sector) &&
        (!q || [p.code, p.name, p.pjpk_name].some((s) => s.toLowerCase().includes(q))),
    );
  }, [all, stage, sector, search]);

  const stats = useMemo(() => {
    const active = all.filter((p) => !["CANCELLED", "TERMINATED", "CLOSED"].includes(p.stage));
    return {
      total: all.length,
      investment: active.reduce((s, p) => s + (p.investment_value || 0), 0),
      preContract: all.filter((p) => PRE_CONTRACT.has(p.stage)).length,
      delayed: all.filter((p) => p.stage_is_delayed).length,
      byStage: KPBU_MAIN_STAGES.map((s) => ({ stage: s, count: all.filter((p) => p.stage === s).length })),
      other: all.filter((p) => !(KPBU_MAIN_STAGES as readonly string[]).includes(p.stage)).length,
    };
  }, [all]);
  const maxCount = Math.max(1, ...stats.byStage.map((s) => s.count));

  return (
    <DashboardLayout title="Portofolio KPBU">
      <div className="space-y-5">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Portofolio KPBU</h1>
            <p className="text-sm text-gray-500">Kerjasama Pemerintah dengan Badan Usaha — tahapan, para pihak, dan target</p>
          </div>
          {canWrite && (
            <button className={btnPrimary} onClick={() => setShowForm(true)}>
              <Plus className="h-4 w-4" /> Proyek KPBU Baru
            </button>
          )}
        </div>

        <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          <StatCard icon={<Handshake className="h-5 w-5 text-blue-600" />} label="Total Proyek KPBU" value={String(stats.total)} />
          <StatCard icon={<Wallet className="h-5 w-5 text-emerald-600" />} label="Nilai Investasi Aktif" value={fmtRupiahShort(stats.investment)} />
          <StatCard icon={<Landmark className="h-5 w-5 text-indigo-600" />} label="Pra-Perjanjian" value={String(stats.preContract)} sub="Identifikasi s.d. transaksi" />
          <StatCard icon={<AlertTriangle className="h-5 w-5 text-red-600" />} label="Tahap Terlambat" value={String(stats.delayed)} sub="Lewat target akhir tahap" />
        </div>

        <div className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
          <h2 className="mb-3 text-sm font-semibold text-gray-700">Sebaran per Tahap</h2>
          <div className="space-y-1.5">
            {stats.byStage.map(({ stage: s, count }) => (
              <button
                key={s}
                onClick={() => setStage(stage === s ? "" : s)}
                className={cn("grid w-full grid-cols-[minmax(0,10rem)_1fr_2rem] items-center gap-2 rounded px-1 py-0.5 text-left text-xs hover:bg-gray-50", stage === s && "bg-blue-50")}
              >
                <span className="truncate text-gray-600">{KPBU_STAGE_LABEL[s]}</span>
                <span className="h-2.5 overflow-hidden rounded-full bg-gray-100">
                  <span className="block h-full rounded-full bg-blue-500" style={{ width: `${(count / maxCount) * 100}%` }} />
                </span>
                <span className="text-right font-medium text-gray-800">{count}</span>
              </button>
            ))}
            {stats.other > 0 && (
              <p className="pt-1 text-xs text-gray-400">{stats.other} proyek ditunda/dibatalkan/diakhiri/selesai</p>
            )}
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          <input className={cn(inputCls, "sm:w-64")} placeholder="Cari kode, nama, PJPK..." value={search} onChange={(e) => setSearch(e.target.value)} aria-label="Cari proyek KPBU" />
          <select className={cn(inputCls, "w-auto")} value={stage} onChange={(e) => setStage(e.target.value)} aria-label="Filter tahap">
            <option value="">Semua Tahap</option>
            {Object.entries(KPBU_STAGE_LABEL).map(([k, l]) => (
              <option key={k} value={k}>{l}</option>
            ))}
          </select>
          <select className={cn(inputCls, "w-auto")} value={sector} onChange={(e) => setSector(e.target.value)} aria-label="Filter sektor">
            <option value="">Semua Sektor</option>
            {Object.entries(KPBU_SECTOR_LABEL).map(([k, l]) => (
              <option key={k} value={k}>{l}</option>
            ))}
          </select>
        </div>

        <div className="overflow-x-auto rounded-xl border border-gray-100 bg-white shadow-sm">
          {isLoading ? (
            <p className="p-6 text-center text-sm text-gray-400">Memuat proyek KPBU...</p>
          ) : isError ? (
            <div className="p-6 text-center text-sm text-red-600">
              Gagal memuat data. <button className="underline" onClick={() => refetch()}>Coba lagi</button>
            </div>
          ) : filtered.length === 0 ? (
            <div className="p-8 text-center">
              <Handshake className="mx-auto mb-2 h-8 w-8 text-gray-300" />
              <p className="text-sm text-gray-500">{all.length === 0 ? "Belum ada proyek KPBU." : "Tidak ada proyek yang cocok dengan filter."}</p>
            </div>
          ) : (
            <table className="w-full min-w-[760px] text-sm">
              <thead className="border-b border-gray-100 text-left text-xs uppercase text-gray-500">
                <tr>
                  <th className="px-4 py-3">Proyek</th>
                  <th className="px-4 py-3">Sektor</th>
                  <th className="px-4 py-3">Tahap</th>
                  <th className="px-4 py-3">Skema</th>
                  <th className="px-4 py-3 text-right">Investasi</th>
                  <th className="px-4 py-3">PJPK</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {filtered.map((p) => (
                  <Row key={p.id} p={p} />
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
      {showForm && <ProjectFormModal onClose={() => setShowForm(false)} />}
    </DashboardLayout>
  );
}

function Row({ p }: { p: KpbuProject }) {
  return (
    <tr className="hover:bg-gray-50">
      <td className="px-4 py-3">
        <Link href={`/kpbu/${p.id}`} className="font-medium text-blue-700 hover:underline">{p.name}</Link>
        <p className="text-xs text-gray-400">{p.code}{p.org_unit_name ? ` · ${p.org_unit_name}` : ""}</p>
      </td>
      <td className="px-4 py-3 text-gray-600">{KPBU_SECTOR_LABEL[p.sector] ?? p.sector}</td>
      <td className="px-4 py-3">
        <StageBadge stage={p.stage} />
        {p.stage_is_delayed && <p className="mt-1 text-xs font-medium text-red-600">Terlambat {p.stage_delay_days} hari</p>}
      </td>
      <td className="px-4 py-3 text-xs text-gray-600">{KPBU_RETURN_SCHEME_LABEL[p.return_scheme] ?? p.return_scheme}</td>
      <td className="px-4 py-3 text-right font-medium text-gray-800">{fmtRupiahShort(p.investment_value)}</td>
      <td className="px-4 py-3 text-xs text-gray-600">{p.pjpk_name || "—"}</td>
    </tr>
  );
}

function StatCard({ icon, label, value, sub }: { icon: React.ReactNode; label: string; value: string; sub?: string }) {
  return (
    <div className="flex min-w-0 items-start gap-3 rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
      <div className="rounded-lg bg-gray-50 p-2">{icon}</div>
      <div className="min-w-0">
        <p className="text-xs text-gray-500">{label}</p>
        <p className="truncate text-xl font-bold text-gray-900">{value}</p>
        {sub && <p className="truncate text-xs text-gray-400">{sub}</p>}
      </div>
    </div>
  );
}
