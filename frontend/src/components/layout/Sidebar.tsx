"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import {
  Activity,
  ArrowUpFromLine,
  BarChart2,
  BarChart3,
  Bell,
  Building2,
  ChevronLeft,
  ChevronRight,
  Database,
  Plug,
  ClipboardCheck,
  FileText,
  FileCheck2,
  HardHat,
  KeyRound,
  Handshake,
  Crown,
  Gavel,
  Layers3,
  LayoutDashboard,
  LogOut,
  MapPinned,
  Network,
  Box,
  Settings,
  ShieldAlert,
  ShieldCheck,
  Tag,
  Users,
  X,
} from "lucide-react";
import { authService } from "@/services/auth.service";
import { useAuthStore } from "@/store/auth.store";
import { cn } from "@/lib/utils";
import { isVendorUser } from "@/lib/roles";

// Role constants — must match backend constants.go values
const R_SUPER_ADMIN = "SUPER_ADMIN";
const R_ADMIN = "ADMIN";
const R_PMO = "PMO";
const R_PROJECT_MANAGER = "PROJECT_MANAGER";
const R_EXECUTIVE_VIEWER = "EXECUTIVE_VIEWER";
const R_AUDITOR = "AUDITOR";

// Shorthand role sets for nav visibility
const ADMIN_ROLES = [R_SUPER_ADMIN, R_ADMIN];
const PMO_AND_ABOVE = [R_SUPER_ADMIN, R_ADMIN, R_PMO];
const REPORTING_ROLES = [R_SUPER_ADMIN, R_ADMIN, R_PMO, R_EXECUTIVE_VIEWER, R_AUDITOR, R_PROJECT_MANAGER];
const EXECUTIVE_ROLES = [R_SUPER_ADMIN, R_ADMIN, R_PMO, R_EXECUTIVE_VIEWER];

interface SidebarProps {
  mobileOpen: boolean;
  onClose: () => void;
  desktopOpen: boolean;
  onToggleDesktop?: () => void;
}

interface NavItem {
  label: string;
  href: string;
  icon: typeof LayoutDashboard;
  /** Roles allowed to see this item. Undefined = visible to all authenticated users. */
  roles?: string[];
}

interface NavSection {
  label: string;
  icon: typeof LayoutDashboard;
  items: NavItem[];
}

const navSections: NavSection[] = [
  {
    label: "Operasional",
    icon: LayoutDashboard,
    items: [
      { label: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
      { label: "Proyek", href: "/projects", icon: Building2 },
      { label: "Command Center", href: "/command-center", icon: ShieldAlert, roles: PMO_AND_ABOVE },
      { label: "Validasi Data", href: "/validation", icon: ClipboardCheck, roles: PMO_AND_ABOVE },
      { label: "Verifikasi Penyedia", href: "/vendor-submissions", icon: FileCheck2, roles: [...PMO_AND_ABOVE, R_PROJECT_MANAGER] },
      { label: "Notifikasi", href: "/notifications", icon: Bell },
    ],
  },
  {
    label: "Control Tower",
    icon: Crown,
    items: [
      { label: "Eksekutif", href: "/executive", icon: Crown, roles: EXECUTIVE_ROLES },
      { label: "Program", href: "/programs", icon: Layers3, roles: EXECUTIVE_ROLES },
      { label: "Peta GIS", href: "/gis", icon: MapPinned, roles: EXECUTIVE_ROLES },
      { label: "Pendukung Keputusan", href: "/decision-support", icon: Gavel, roles: EXECUTIVE_ROLES },
      { label: "Manfaat", href: "/benefits", icon: BarChart3, roles: EXECUTIVE_ROLES },
    ],
  },
  {
    label: "KPBU",
    icon: Handshake,
    items: [
      { label: "Portofolio KPBU", href: "/kpbu", icon: Handshake, roles: [...EXECUTIVE_ROLES, R_PROJECT_MANAGER] },
    ],
  },
  {
    label: "Data & Pelaporan",
    icon: Database,
    items: [
      { label: "Pusat Reporting", href: "/reports/analytics", icon: BarChart2, roles: REPORTING_ROLES },
      { label: "Laporan Periodik", href: "/reports", icon: FileText, roles: REPORTING_ROLES },
      { label: "Import Data", href: "/imports", icon: ArrowUpFromLine, roles: PMO_AND_ABOVE },
      { label: "Data Governance", href: "/governance", icon: ShieldCheck, roles: PMO_AND_ABOVE },
      { label: "Audit Logs", href: "/audit-logs", icon: Activity, roles: [...PMO_AND_ABOVE, R_AUDITOR] },
    ],
  },
  {
    label: "Integrasi",
    icon: Plug,
    items: [
      { label: "Connector Pemerintah", href: "/integrations/government", icon: Building2, roles: PMO_AND_ABOVE },
      { label: "Primavera P6", href: "/integrations/primavera", icon: Network, roles: PMO_AND_ABOVE },
      { label: "BIM / Digital Twin", href: "/integrations/bim", icon: Box, roles: PMO_AND_ABOVE },
    ],
  },
];

const settingsItems: NavItem[] = [
  { label: "Pengguna", href: "/settings/users", icon: Users, roles: ADMIN_ROLES },
  { label: "Organisasi", href: "/settings/organizations", icon: Building2, roles: [R_SUPER_ADMIN] },
  { label: "Role", href: "/settings/roles", icon: ShieldCheck, roles: ADMIN_ROLES },
  { label: "Org Unit", href: "/settings/org-units", icon: Settings, roles: ADMIN_ROLES },
  { label: "Program", href: "/settings/programs", icon: Layers3, roles: PMO_AND_ABOVE },
  { label: "Sektor SDA", href: "/settings/sectors", icon: BarChart3, roles: PMO_AND_ABOVE },
  { label: "Wilayah", href: "/settings/regions", icon: MapPinned, roles: PMO_AND_ABOVE },
  { label: "DAS", href: "/settings/river-basins", icon: MapPinned, roles: PMO_AND_ABOVE },
  { label: "Kategori Proyek", href: "/settings/project-categories", icon: Tag, roles: PMO_AND_ABOVE },
  { label: "Akun Penyedia", href: "/settings/vendor-accounts", icon: KeyRound, roles: ADMIN_ROLES },
  { label: "Checklist Dokumen KPBU", href: "/settings/kpbu-requirements", icon: Tag, roles: PMO_AND_ABOVE },
];

/** Contractor accounts see only the contractor portal. */
const vendorNavSections: NavSection[] = [
  {
    label: "Portal Penyedia",
    icon: HardHat,
    items: [
      { label: "Kontrak Saya", href: "/vendor-portal", icon: HardHat },
      { label: "Notifikasi", href: "/notifications", icon: Bell },
    ],
  },
];

const defaultOpenSections: Record<string, boolean> = {
  Operasional: true,
};

export function Sidebar({ mobileOpen, onClose, desktopOpen, onToggleDesktop }: SidebarProps) {
  const pathname = usePathname();
  const router = useRouter();
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const hasRole = useAuthStore((state) => state.hasRole);
  const user = useAuthStore((state) => state.user);
  const [isMounted, setIsMounted] = useState(false);
  const [openSections, setOpenSections] = useState<Record<string, boolean>>(defaultOpenSections);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    const allSections = [...navSections, ...vendorNavSections, { label: "Pengaturan", icon: Settings, items: settingsItems }];
    const activeSection = allSections.find((section) =>
      section.items.some((item) => isItemActive(item, pathname))
    );

    if (activeSection) {
      // Close all sections except the one containing the active route
      const next: Record<string, boolean> = {};
      allSections.forEach((s) => {
        next[s.label] = s.label === activeSection.label;
      });
      setOpenSections(next);
    }
  }, [pathname]);

  const vendorUser = isMounted && isVendorUser(user);
  const visibleNavSections = (vendorUser ? vendorNavSections : navSections)
    .map((section) => ({
      ...section,
      items: section.items.filter(
        (item) => !item.roles || (isMounted && item.roles.some((r) => hasRole(r)))
      ),
    }))
    .filter((section) => section.items.length > 0);
  const visibleSettingsItems = vendorUser
    ? []
    : settingsItems.filter((item) => !item.roles || (isMounted && item.roles.some((r) => hasRole(r))));
  const visibleSections = [
    ...visibleNavSections,
    ...(visibleSettingsItems.length > 0
      ? [{ label: "Pengaturan", icon: Settings, items: visibleSettingsItems }]
      : []),
  ];

  const toggleSection = (label: string) => {
    setOpenSections((current) => ({
      ...current,
      [label]: !(current[label] ?? false),
    }));
  };

  const handleLogout = async () => {
    try {
      await authService.logout();
    } catch {
      // Local auth is cleared even when the server session is unavailable.
    } finally {
      clearAuth();
      onClose();
      router.push("/login");
    }
  };

  const primaryRole = isMounted ? user?.roles?.[0] : undefined;

  return (
    <>
      {mobileOpen && (
        <button
          type="button"
          className="fixed inset-0 z-40 bg-slate-950/50 lg:hidden"
          aria-label="Tutup navigasi"
          onClick={onClose}
        />
      )}

      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r border-slate-200 bg-white text-slate-700 transition-transform duration-200 ease-out",
          mobileOpen ? "translate-x-0" : "-translate-x-full",
          desktopOpen ? "lg:translate-x-0" : "lg:-translate-x-full"
        )}
      >
        <div className="flex items-center gap-3 px-5 pb-4 pt-5">
          <div className="grid h-11 w-11 shrink-0 place-items-center overflow-hidden">
            <Image
              src="/images/logo-kemenpu.png"
              alt="Logo Kementerian Pekerjaan Umum"
              width={44}
              height={44}
              className="h-full w-full object-contain"
            />
          </div>
          <div className="min-w-0">
            <p className="truncate text-base font-bold uppercase leading-5 text-slate-800">Kementerian PU</p>
            <p className="mt-0.5 truncate text-xs text-slate-500">PMO Digital Platform</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="ml-auto grid h-8 w-8 place-items-center rounded-md text-slate-500 hover:bg-slate-100 lg:hidden"
            aria-label="Tutup navigasi"
          >
            <X className="h-5 w-5" aria-hidden="true" />
          </button>
        </div>

        {onToggleDesktop && (
          <button
            type="button"
            onClick={onToggleDesktop}
            className="absolute -right-4 top-6 hidden h-8 w-8 place-items-center rounded-full bg-[#17385D] text-white shadow-md ring-4 ring-white transition-colors hover:bg-[#0F2742] lg:grid"
            aria-label="Sembunyikan navigasi"
          >
            <ChevronLeft className="h-4 w-4" aria-hidden="true" />
          </button>
        )}

        <div className="mx-4 mb-3 rounded-2xl border border-[#17385D]/15 bg-[#17385D]/[.05] px-4 py-3">
          <p className="text-[10px] font-medium uppercase tracking-[.12em] text-slate-500">Ruang Kerja Aktif</p>
          <p className="mt-1 truncate text-sm font-semibold text-slate-800">PMO Control Tower</p>
          {primaryRole && (
            <p className="mt-1 flex items-center gap-1.5 truncate text-xs text-[#B7800A]">
              <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-[#EBAA16]" aria-hidden="true" />
              {primaryRole.replace(/_/g, " ")}
            </p>
          )}
        </div>

        <nav className="flex-1 overflow-y-auto px-4 py-1" aria-label="Navigasi utama">
          <ul className="space-y-1" role="list">
            {visibleSections.map((section) =>
              section.items.length === 1 ? (
                <SidebarLink key={section.label} item={section.items[0]} pathname={pathname} onClose={onClose} />
              ) : (
                <SidebarSection
                  key={section.label}
                  section={section}
                  pathname={pathname}
                  open={openSections[section.label] ?? false}
                  onToggle={() => toggleSection(section.label)}
                  onClose={onClose}
                />
              )
            )}
          </ul>
        </nav>

        <div className="border-t border-slate-100 p-4">
          <button
            type="button"
            onClick={handleLogout}
            className="flex h-11 w-full items-center gap-3 rounded-xl px-3 text-sm text-slate-500 transition-colors hover:bg-red-50 hover:text-red-600"
          >
            <LogOut className="h-5 w-5" aria-hidden="true" />
            Keluar
          </button>
        </div>
      </aside>
    </>
  );
}

function SidebarSection({
  section,
  pathname,
  open,
  onToggle,
  onClose,
}: {
  section: NavSection;
  pathname: string;
  open: boolean;
  onToggle: () => void;
  onClose: () => void;
}) {
  const active = section.items.some((item) => isItemActive(item, pathname));
  const Icon = section.icon;

  return (
    <li>
      <button
        type="button"
        onClick={onToggle}
        className={cn(
          "flex h-11 w-full items-center gap-3 rounded-xl px-3 text-left text-sm transition-colors",
          active ? "font-semibold text-[#17385D]" : "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
        )}
        aria-expanded={open}
      >
        <Icon className="h-5 w-5 shrink-0 text-[#17385D]" aria-hidden="true" />
        <span className="min-w-0 flex-1 truncate">{section.label}</span>
        <ChevronRight
          className={cn("h-4 w-4 shrink-0 text-slate-400 transition-transform", open && "rotate-90")}
          aria-hidden="true"
        />
      </button>
      {open && (
        <ul className="mt-1 space-y-1 border-l border-slate-200 pl-3 ml-5" role="list">
          {section.items.map((item) => (
            <SidebarLink key={item.href} item={item} pathname={pathname} onClose={onClose} nested />
          ))}
        </ul>
      )}
    </li>
  );
}

function SidebarLink({
  item,
  pathname,
  onClose,
  nested = false,
}: {
  item: NavItem;
  pathname: string;
  onClose: () => void;
  nested?: boolean;
}) {
  const Icon = item.icon;
  const isActive = isItemActive(item, pathname);

  return (
    <li>
      <Link
        href={item.href}
        onClick={onClose}
        className={cn(
          "flex items-center gap-3 rounded-xl px-3 transition-colors",
          nested ? "h-9 text-[13px]" : "h-11 text-sm",
          isActive
            ? "bg-[#17385D]/[.08] font-semibold text-[#17385D]"
            : "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
        )}
        aria-current={isActive ? "page" : undefined}
      >
        <Icon
          className={cn("shrink-0 text-[#17385D]", nested ? "h-4 w-4" : "h-5 w-5", !isActive && nested && "text-slate-400")}
          aria-hidden="true"
        />
        <span className="min-w-0 flex-1 truncate">{item.label}</span>
      </Link>
    </li>
  );
}

function isItemActive(item: NavItem, pathname: string) {
  return (
    pathname === item.href ||
    (item.href !== "/reports" && pathname.startsWith(`${item.href}/`))
  );
}
