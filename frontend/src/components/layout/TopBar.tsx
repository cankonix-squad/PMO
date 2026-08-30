"use client";

import { Bell, Menu, PanelLeftClose, PanelLeftOpen, UserRound } from "lucide-react";
import { useAuthStore } from "@/store/auth.store";

interface TopBarProps {
  title: string;
  onOpenMenu: () => void;
  desktopNavOpen: boolean;
  onToggleDesktopNav: () => void;
}

export function TopBar({
  title,
  onOpenMenu,
  desktopNavOpen,
  onToggleDesktopNav,
}: TopBarProps) {
  const user = useAuthStore((state) => state.user);
  const displayName = user
    ? `${user.first_name} ${user.last_name}`.trim()
    : "Pengguna";
  const role = user?.roles?.[0]?.replaceAll("_", " ") ?? "PMO";
  const dateLabel = new Intl.DateTimeFormat("id-ID", {
    weekday: "long",
    day: "2-digit",
    month: "long",
    year: "numeric",
  }).format(new Date());
  const titleMap: Record<string, { title: string; subtitle: string }> = {
    Dashboard: {
      title: "Dashboard Eksekutif",
      subtitle: "Ringkasan portofolio, monitoring, risiko, isu, dan keputusan",
    },
    Projects: {
      title: "Proyek",
      subtitle: "Daftar dan detail proyek dalam cakupan organisasi",
    },
    Pengguna: {
      title: "Pengguna",
      subtitle: "Manajemen akun, status aktif, dan role akses",
    },
    "Indikator Manfaat": {
      title: "Indikator Manfaat",
      subtitle: "Outcome proyek dan agregasi measurement tervalidasi",
    },
    "Benefit & Outcome Indicators": {
      title: "Indikator Manfaat",
      subtitle: "Outcome proyek dan agregasi measurement tervalidasi",
    },
    "Reports & Analytics": {
      title: "Reporting & Analytics",
      subtitle: "Read model, dataset analitik, dan export laporan",
    },
    "PMO Command Center": {
      title: "Command Center",
      subtitle: "Peringatan, eskalasi, keputusan, dan tindak lanjut prioritas",
    },
    "Executive Dashboard": {
      title: "Dashboard Eksekutif",
      subtitle: "Ringkasan nasional untuk pimpinan",
    },
    "Decision Support — Priority Scoring": {
      title: "Decision Support",
      subtitle: "Ranking prioritas proyek dan penjelasan komponen skor",
    },
    "GIS Map": {
      title: "Peta GIS",
      subtitle: "Sebaran lokasi proyek dan status kesehatan",
    },
    "BIM / Digital Twin": {
      title: "BIM / Digital Twin",
      subtitle: "Metadata model, versi, dan pemetaan proyek",
    },
  };
  const pageMeta = titleMap[title] ?? {
    title,
    subtitle: "Kementerian Pekerjaan Umum · Ditjen Sumber Daya Air",
  };

  return (
    <header className="sticky top-0 z-30 border-b border-slate-200 bg-white">
      <div className="mx-auto flex min-h-20 w-full max-w-[2400px] items-center justify-between gap-4 px-4 py-3 sm:px-6 lg:px-7">
        <div className="flex min-w-0 items-center gap-3">
          <button
            type="button"
            onClick={onOpenMenu}
            className="grid h-10 w-10 shrink-0 place-items-center rounded-md border border-slate-200 text-[#0b4c91] lg:hidden"
            aria-label="Buka navigasi"
            title="Buka navigasi"
          >
            <Menu className="h-5 w-5" aria-hidden="true" />
          </button>
          <button
            type="button"
            onClick={onToggleDesktopNav}
            className="hidden h-10 w-10 shrink-0 place-items-center rounded-md border border-slate-200 text-[#0b4c91] transition-colors hover:bg-slate-50 lg:grid"
            aria-label={desktopNavOpen ? "Sembunyikan navigasi" : "Tampilkan navigasi"}
            title={desktopNavOpen ? "Sembunyikan navigasi" : "Tampilkan navigasi"}
          >
            {desktopNavOpen ? (
              <PanelLeftClose className="h-5 w-5" aria-hidden="true" />
            ) : (
              <PanelLeftOpen className="h-5 w-5" aria-hidden="true" />
            )}
          </button>
          <div className="min-w-0">
            <h1 className="truncate text-lg font-semibold leading-7 text-[#082e63] sm:text-xl">
              {pageMeta.title}
            </h1>
            <p className="mt-0.5 truncate text-xs font-medium leading-5 text-slate-500 sm:text-sm">
              {pageMeta.subtitle}
            </p>
          </div>
        </div>

        <div className="flex shrink-0 items-center gap-3">
          <div className="hidden border-r border-slate-200 pr-4 text-right xl:block">
            <p className="text-xs font-medium text-[#082e63]">{dateLabel}</p>
            <p className="mt-0.5 text-[11px] text-slate-500">Waktu Indonesia Barat</p>
          </div>
          <button
            type="button"
            className="relative grid h-9 w-9 place-items-center rounded-md text-slate-500 hover:bg-slate-100"
            aria-label="Notifikasi"
            title="Notifikasi"
          >
            <Bell className="h-5 w-5" aria-hidden="true" />
          </button>
          <div className="flex items-center gap-2">
            <div className="grid h-9 w-9 place-items-center rounded-full bg-[#0b4c91] text-white">
              <UserRound className="h-5 w-5" aria-hidden="true" />
            </div>
            <div className="hidden text-left sm:block">
              <p className="max-w-36 truncate text-xs font-semibold text-[#082e63]">{displayName}</p>
              <p className="max-w-36 truncate text-[10px] uppercase text-slate-500">{role}</p>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
