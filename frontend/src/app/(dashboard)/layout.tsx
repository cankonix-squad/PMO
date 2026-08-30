import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "PMO",
};

export default function DashboardGroupLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  // Auth guard is handled by middleware.ts
  return <>{children}</>;
}
