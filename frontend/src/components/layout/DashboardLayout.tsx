"use client";

import { useState } from "react";
import { Sidebar } from "./Sidebar";
import { TopBar } from "./TopBar";
import { cn } from "@/lib/utils";

interface DashboardLayoutProps {
  title: string;
  children: React.ReactNode;
}

export function DashboardLayout({ title, children }: DashboardLayoutProps) {
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const [desktopNavOpen, setDesktopNavOpen] = useState(true);

  return (
    <div className="min-h-screen bg-[#f4f7fb]">
      <Sidebar
        mobileOpen={mobileNavOpen}
        onClose={() => setMobileNavOpen(false)}
        desktopOpen={desktopNavOpen}
      />
      <div
        className={cn(
          "flex min-h-screen flex-col transition-[margin] duration-200 ease-out",
          desktopNavOpen ? "lg:ml-60" : "lg:ml-0"
        )}
      >
        <TopBar
          title={title}
          onOpenMenu={() => setMobileNavOpen(true)}
          desktopNavOpen={desktopNavOpen}
          onToggleDesktopNav={() => setDesktopNavOpen((open) => !open)}
        />
        <main className="mx-auto w-full max-w-[2400px] flex-1 px-4 py-5 sm:px-6 lg:px-7">
          {children}
        </main>
      </div>
    </div>
  );
}
