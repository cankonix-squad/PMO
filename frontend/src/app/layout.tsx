import type { Metadata } from "next";
import "./globals.css";
import { Providers } from "./providers";

export const metadata: Metadata = {
  title: "PMO — Kementerian Pekerjaan Umum",
  description: "Platform monitoring dan pengendalian proyek Kementerian Pekerjaan Umum",
  icons: {
    icon: "/images/logo-kemenpu.png",
    apple: "/images/logo-kemenpu.png",
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="id" suppressHydrationWarning>
      <body className="font-sans">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
