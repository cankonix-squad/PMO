"use client";

import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { Sidebar } from "./Sidebar";
import { TopBar } from "./TopBar";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/store/auth.store";
import { isVendorUser, VENDOR_ALLOWED_PATHS, VENDOR_HOME } from "@/lib/roles";

interface DashboardLayoutProps {
  title: string;
  children: React.ReactNode;
}

export function DashboardLayout({ title, children }: DashboardLayoutProps) {
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const [desktopNavOpen, setDesktopNavOpen] = useState(true);
  const user = useAuthStore((s) => s.user);
  const pathname = usePathname();
  const router = useRouter();
  const vendor = isVendorUser(user);

  // Contractor accounts are confined to the contractor portal. The backend
  // enforces this with permissions; this only avoids showing 403 pages.
  useEffect(() => {
    if (vendor && !VENDOR_ALLOWED_PATHS.some((p) => pathname.startsWith(p))) {
      router.replace(VENDOR_HOME);
    }
  }, [vendor, pathname, router]);

  return (
    <div className="pmo-dashboard min-h-screen bg-[#F4F6F8]">
      <Sidebar
        mobileOpen={mobileNavOpen}
        onClose={() => setMobileNavOpen(false)}
        desktopOpen={desktopNavOpen}
        onToggleDesktop={() => setDesktopNavOpen(false)}
      />
      <div
        className={cn(
          "flex min-h-screen flex-col transition-[margin] duration-200 ease-out",
          desktopNavOpen ? "lg:ml-64" : "lg:ml-0"
        )}
      >
        <TopBar
          title={title}
          onOpenMenu={() => setMobileNavOpen(true)}
          desktopNavOpen={desktopNavOpen}
          onToggleDesktopNav={() => setDesktopNavOpen((open) => !open)}
        />
        <main className="mx-auto w-full max-w-[2400px] flex-1 px-4 py-5 sm:px-6 lg:px-7 xl:py-7">
          {children}
        </main>
      </div>
    </div>
  );
}
