import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Masuk - CANKORA PMO Digital Platform",
  description: "Masuk ke CANKORA PMO National Project Control Tower Ditjen Sumber Daya Air.",
};

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}
