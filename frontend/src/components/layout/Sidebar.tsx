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
  ChevronDown,
  ClipboardCheck,
  FileText,
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
  items: NavItem[];
}

const navSections: NavSection[] = [
  {
    label: "Operasional",
    items: [
      { label: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
      { label: "Proyek", href: "/projects", icon: Building2 },
      { label: "Command Center", href: "/command-center", icon: ShieldAlert, roles: PMO_AND_ABOVE },
      { label: "Validasi Data", href: "/validation", icon: ClipboardCheck, roles: PMO_AND_ABOVE },
      { label: "Notifikasi", href: "/notifications", icon: Bell },
    ],
  },
  {
    label: "Control Tower",
    items: [
      { label: "Eksekutif", href: "/executive", icon: Crown, roles: EXECUTIVE_ROLES },
      { label: "Program", href: "/programs", icon: Layers3, roles: EXECUTIVE_ROLES },
      { label: "Peta GIS", href: "/gis", icon: MapPinned, roles: EXECUTIVE_ROLES },
      { label: "Pendukung Keputusan", href: "/decision-support", icon: Gavel, roles: EXECUTIVE_ROLES },
      { label: "Manfaat", href: "/benefits", icon: BarChart3, roles: EXECUTIVE_ROLES },
    ],
  },
  {
    label: "Data & Pelaporan",
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
];

const defaultOpenSections: Record<string, boolean> = {
  Operasional: true,
};

export function Sidebar({ mobileOpen, onClose, desktopOpen }: SidebarProps) {
  const pathname = usePathname();
  const router = useRouter();
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const hasRole = useAuthStore((state) => state.hasRole);
  const [isMounted, setIsMounted] = useState(false);
  const [openSections, setOpenSections] = useState<Record<string, boolean>>(defaultOpenSections);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    const allSections = [...navSections, { label: "Pengaturan", items: settingsItems }];
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

  const visibleNavSections = navSections
    .map((section) => ({
      ...section,
      items: section.items.filter(
        (item) => !item.roles || (isMounted && item.roles.some((r) => hasRole(r)))
      ),
    }))
    .filter((section) => section.items.length > 0);
  const visibleSettingsItems = settingsItems.filter(
    (item) => !item.roles || (isMounted && item.roles.some((r) => hasRole(r)))
  );
  const visibleSections = [
    ...visibleNavSections,
    ...(visibleSettingsItems.length > 0
      ? [{ label: "Pengaturan", items: visibleSettingsItems }]
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
          "fixed inset-y-0 left-0 z-50 flex w-60 flex-col bg-[#052f63] text-white shadow-xl transition-transform duration-200 ease-out",
          mobileOpen ? "translate-x-0" : "-translate-x-full",
          desktopOpen ? "lg:translate-x-0" : "lg:-translate-x-full"
        )}
      >
        <div className="flex min-h-20 items-center gap-3 border-b border-white/15 px-5 py-3">
          <div className="grid h-11 w-11 shrink-0 place-items-center overflow-hidden rounded-md bg-white shadow-sm ring-1 ring-white/20">
            <Image
              src="/images/logo-kemenpu.png"
              alt="Logo Kementerian Pekerjaan Umum"
              width={44}
              height={44}
              className="h-full w-full object-contain p-0.5"
            />
          </div>
          <div className="min-w-0">
            <p className="text-sm font-bold uppercase leading-5">Kementerian PU</p>
            <p className="mt-0.5 text-[10px] uppercase leading-4 text-blue-100">
              Ditjen Sumber Daya Air
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="ml-auto grid h-8 w-8 place-items-center rounded-md text-blue-100 hover:bg-white/10 hover:text-white lg:hidden"
            aria-label="Tutup navigasi"
          >
            <X className="h-5 w-5" aria-hidden="true" />
          </button>
        </div>

        <div className="px-5 pb-3 pt-5">
          <p className="text-[10px] font-semibold uppercase text-blue-300">PMO Control Tower</p>
          <p className="mt-1 text-sm font-semibold">PMO</p>
        </div>

        <nav className="flex-1 overflow-y-auto px-3 py-2" aria-label="Navigasi utama">
          <div className="space-y-2">
            {visibleSections.map((section) => (
              <SidebarSection
                key={section.label}
                section={section}
                pathname={pathname}
                open={openSections[section.label] ?? false}
                onToggle={() => toggleSection(section.label)}
                onClose={onClose}
              />
            ))}
          </div>
        </nav>

        <div className="border-t border-white/15 p-3">
          <button
            type="button"
            onClick={handleLogout}
            className="flex h-10 w-full items-center gap-3 rounded-md px-3 text-sm text-blue-100 transition-colors hover:bg-red-500/15 hover:text-white"
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

  return (
    <section>
      <button
        type="button"
        onClick={onToggle}
        className={cn(
          "flex h-9 w-full items-center justify-between rounded-md px-3 text-left text-[10px] font-semibold uppercase tracking-wide transition-colors",
          active || open
            ? "bg-white/10 text-blue-100"
            : "text-blue-300 hover:bg-white/10 hover:text-blue-100"
        )}
        aria-expanded={open}
      >
        <span className="truncate">{section.label}</span>
        <ChevronDown
          className={cn("h-4 w-4 shrink-0 transition-transform", open ? "rotate-0" : "-rotate-90")}
          aria-hidden="true"
        />
      </button>
      {open && (
        <ul className="mt-1 space-y-1" role="list">
          {section.items.map((item) => (
            <SidebarLink key={item.href} item={item} pathname={pathname} onClose={onClose} />
          ))}
        </ul>
      )}
    </section>
  );
}

function SidebarLink({
  item,
  pathname,
  onClose,
}: {
  item: NavItem;
  pathname: string;
  onClose: () => void;
}) {
  const Icon = item.icon;
  const isActive = isItemActive(item, pathname);

  return (
    <li>
      <Link
        href={item.href}
        onClick={onClose}
        className={cn(
          "group flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium transition-colors",
          isActive
            ? "bg-[#1262b8] text-white shadow-sm"
            : "text-blue-100 hover:bg-white/10 hover:text-white"
        )}
        aria-current={isActive ? "page" : undefined}
      >
        <span
          className={cn(
            "h-5 w-0.5 rounded-full transition-colors",
            isActive ? "bg-white" : "bg-transparent"
          )}
          aria-hidden="true"
        />
        <Icon className="h-4 w-4 shrink-0" aria-hidden="true" />
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
