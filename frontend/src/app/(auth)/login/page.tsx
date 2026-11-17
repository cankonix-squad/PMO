"use client";

import Image from "next/image";
import Link from "next/link";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  BarChart3,
  Eye,
  EyeOff,
  LockKeyhole,
  LogIn,
  ShieldCheck,
  TriangleAlert,
  UserRound,
  UsersRound,
} from "lucide-react";
import { authService } from "@/services/auth.service";
import { useAuthStore } from "@/store/auth.store";
import { isVendorUser, VENDOR_HOME } from "@/lib/roles";
import { cn } from "@/lib/utils";

const loginSchema = z.object({
  email: z.string().email("Masukkan alamat email yang valid"),
  password: z.string().min(1, "Password wajib diisi"),
  rememberMe: z.boolean(),
});

type LoginForm = z.infer<typeof loginSchema>;

const platformBenefits = [
  {
    icon: BarChart3,
    title: "Monitoring Terintegrasi",
    description: "Pantau progres proyek SDA secara real-time dan akurat.",
  },
  {
    icon: TriangleAlert,
    title: "Peringatan Dini",
    description: "Identifikasi risiko lebih awal untuk mitigasi yang tepat.",
  },
  {
    icon: UsersRound,
    title: "Keputusan Eksekutif",
    description: "Dukung keputusan lebih cepat dengan data dan insight.",
  },
];

function getSafeRedirectPath(search: string) {
  const from = new URLSearchParams(search).get("from");
  if (!from || !from.startsWith("/") || from.startsWith("//")) {
    return "/dashboard";
  }
  return from;
}

function BrandMark({ compact = false }: { compact?: boolean }) {
  return (
    <div
      className={cn(
        "grid shrink-0 place-items-center overflow-hidden rounded-md bg-white shadow-sm ring-1 ring-white/20",
        compact ? "h-12 w-12" : "h-14 w-14"
      )}
      aria-hidden="true"
    >
      <Image
        src="/images/logo-kemenpu.png"
        alt="Logo Kementerian Pekerjaan Umum"
        width={compact ? 48 : 56}
        height={compact ? 48 : 56}
        priority
        className="h-full w-full object-contain p-1"
      />
    </div>
  );
}

export default function LoginPage() {
  const router = useRouter();
  const { setAuth } = useAuthStore();
  const [serverError, setServerError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
      rememberMe: true,
    },
  });

  const onSubmit = async (values: LoginForm) => {
    setServerError(null);
    try {
      const { data: res } = await authService.login({
        email: values.email,
        password: values.password,
      });
      if (res.success) {
        const { access_token, refresh_token, user } = res.data;
        setAuth(user, access_token, refresh_token, values.rememberMe);
        const requested =
          typeof window !== "undefined"
            ? getSafeRedirectPath(window.location.search)
            : "/dashboard";
        // Contractor accounts only have the contractor portal.
        const redirectPath =
          isVendorUser(user) && !requested.startsWith(VENDOR_HOME) ? VENDOR_HOME : requested;
        router.replace(redirectPath);
        router.refresh();
      }
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data
          ?.message ?? "Login gagal. Periksa kembali email dan password Anda.";
      setServerError(msg);
    }
  };

  return (
    <main className="min-h-screen bg-[#F4F6F8] lg:grid lg:grid-cols-[minmax(0,1.63fr)_minmax(0,1fr)]">
      <section className="relative hidden min-h-screen overflow-hidden bg-white lg:flex lg:flex-col lg:justify-between">
        <div className="absolute inset-0 bg-white" />
        <svg className="absolute inset-0 h-full w-full opacity-[.075]" viewBox="0 0 1200 1000" preserveAspectRatio="xMidYMid slice" aria-hidden="true">
          <defs>
            <pattern id="draft-hatch" width="12" height="12" patternTransform="rotate(28)" patternUnits="userSpaceOnUse"><path d="M0 0v12" stroke="#aeb6c0" strokeWidth="1" /></pattern>
            <pattern id="draft-hatch-light" width="17" height="17" patternTransform="rotate(-22)" patternUnits="userSpaceOnUse"><path d="M0 0v17" stroke="#c0c6ce" strokeWidth=".8" /></pattern>
          </defs>
          <g fill="none" stroke="#8c97a5" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
            {/* Sparse structural sketch, kept toward the open upper and right areas */}
            <path d="M390 470 490 95l265-35 94 410M422 470 512 108l224-29 79 391M452 470 535 121l187-24 65 373" />
            <path d="M492 100 503 470M574 88 578 470M656 76 650 470M738 92 728 470M820 172 806 470" />
            <path d="m492 172 350-42M476 242l385-47M460 312l420-53M442 386l458-60" />
            {/* Soft hand-drawn contour strokes */}
            <path d="M-30 250q55-26 93 3t77-5q37-22 76 7t72 0M-20 276q58-20 94 7t78-4q39-20 77 8t70 1" />
            <path d="M914 64q32 14 48 1t34 17q19 24 45 8t36 22q14 26 47 12t58 29M904 84q33 12 49 0t35 18q19 23 46 7t36 23q14 25 47 11t59 30" />
            {/* One small hatch detail */}
            <path d="M918 430q46-24 86-5t76-5l-12 38q-53 13-91-8t-64 14z" fill="url(#draft-hatch)" stroke="none" />
            <path d="M112 665h690M112 659v12M802 659v12" strokeDasharray="3 7" />
          </g>
        </svg>
        <div className="absolute inset-x-0 bottom-0 h-[28%] bg-gradient-to-t from-[#17385D]/[.035] to-transparent" />
        <div className="absolute inset-y-0 left-0 w-1.5 bg-[#EBAA16]" />

        <header className="relative z-10 flex items-center justify-between px-10 pt-9 xl:px-14 xl:pt-11">
          <div className="flex items-center gap-4">
            <BrandMark />
            <div className="text-xs font-semibold uppercase leading-snug tracking-[.12em] text-[#17385D] xl:text-sm">
              <p>Kementerian Pekerjaan Umum</p>
            </div>
          </div>
          <div className="rounded-full border border-[#17385D]/10 bg-white/80 px-4 py-2 text-xs font-medium tracking-wide text-[#17385D]/70">
            PROJECT MANAGEMENT OFFICE
          </div>
        </header>

        <div className="relative z-10 max-w-4xl px-10 pb-14 xl:px-14">
          <p className="mb-5 text-sm font-semibold uppercase tracking-[.18em] text-[#A77508]">
            PMO DIGITAL PLATFORM
          </p>
          <h1 className="max-w-3xl text-5xl font-bold leading-[1.04] tracking-tight text-[#17385D] xl:text-6xl 2xl:text-7xl">
            Pantau proyek.<br /><span className="text-[#EBAA16]">Bangun dampak.</span>
          </h1>
          <p className="mt-5 max-w-2xl text-base leading-7 text-slate-600 xl:text-lg">
            PMO Digital Platform menyatukan data, progres, dan risiko proyek dalam satu ruang kendali untuk keputusan yang lebih cepat.
          </p>
          <div className="mt-8 h-1 w-16 rounded-full bg-[#EBAA16]" />
          <div className="mt-7 flex flex-wrap items-center gap-x-8 gap-y-3 text-sm font-medium text-[#17385D]/75">
            <span>Monitoring terintegrasi</span>
            <span>Peringatan dini</span>
            <span>Keputusan berbasis data</span>
          </div>
        </div>

        <div className="relative z-10 grid grid-cols-3 gap-5 border-t border-[#17385D]/10 bg-white/80 px-10 py-6 xl:px-14">
          {platformBenefits.map(({ icon: Icon, title, description }) => (
            <div key={title} className="flex min-w-0 gap-3">
              <div className="grid h-10 w-10 shrink-0 place-items-center rounded-xl border border-[#EBAA16]/50 bg-white text-[#A77508]">
                <Icon className="h-5 w-5" aria-hidden="true" />
              </div>
              <div className="min-w-0">
                <p className="text-sm font-semibold text-[#17385D]">{title}</p>
                <p className="mt-1 text-xs leading-5 text-slate-500">{description}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="relative flex min-h-screen flex-col bg-[#F4F6F8] lg:justify-center lg:px-7 xl:px-10 2xl:px-14">
        <div className="relative h-48 overflow-hidden bg-[#17365b] lg:hidden">
          <div className="absolute -right-8 -top-24 h-64 w-64 rounded-full border border-white/15" />
          <div className="absolute -right-2 -top-16 h-48 w-48 rounded-full border border-[#f4b323]/35" />
          <div className="relative z-10 flex h-full flex-col justify-between p-5">
            <div className="flex items-center gap-3">
              <BrandMark compact />
              <div className="text-xs font-semibold uppercase leading-snug text-white">
                <p>Kementerian Pekerjaan Umum</p>
              </div>
            </div>
            <div>
              <p className="text-2xl font-bold text-white">PMO Digital Platform</p>
              <p className="mt-1 text-sm text-white/75">Kementerian Pekerjaan Umum</p>
            </div>
          </div>
        </div>

        <div className="relative z-10 flex flex-1 items-center justify-center px-5 py-6 lg:px-5 lg:py-4">
          <div className="w-full max-w-[440px] rounded-[24px] border border-white bg-white/90 px-6 py-6 shadow-[0_20px_56px_rgba(23,56,93,0.08)] sm:px-9 sm:py-7">
            <div className="mb-6">
              <div className="mb-6 flex items-center gap-3">
                <div className="grid h-11 w-11 place-items-center overflow-hidden rounded-xl bg-white ring-1 ring-slate-200">
                  <Image src="/images/logo-kemenpu.png" alt="Logo Kementerian Pekerjaan Umum" width={40} height={40} priority className="h-10 w-10 object-contain" />
                </div>
                <div className="text-left">
                  <p className="text-[11px] font-bold uppercase tracking-[.14em] text-[#17385D]">Kementerian PU</p>
                  <p className="mt-0.5 text-xs text-[#64748B]">PMO Digital Platform</p>
                </div>
              </div>
              <p className="mb-2 text-xs font-semibold uppercase tracking-[.13em] text-[#A77508]">Selamat datang kembali</p>
              <h2 className="whitespace-nowrap text-[26px] font-bold tracking-[-.04em] text-[#17385D]">Masuk ke akun Anda</h2>
              <p className="mt-1.5 text-sm leading-5 text-[#64748B]">Gunakan akun PMO untuk melanjutkan.</p>
            </div>

            {serverError && (
              <div
                role="alert"
                className="mb-5 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
              >
                {serverError}
              </div>
            )}

            <form
              onSubmit={handleSubmit(onSubmit)}
              className="space-y-3.5"
              method="post"
              noValidate
            >
              <div>
                <label htmlFor="email" className="mb-2 block text-sm font-semibold text-[#102a52]">
                  Email
                </label>
                <div className="relative">
                  <UserRound className="pointer-events-none absolute left-3.5 top-1/2 h-5 w-5 -translate-y-1/2 text-[#1554b8]" aria-hidden="true" />
                  <input
                    id="email"
                    type="email"
                    autoComplete="email"
                    {...register("email")}
                    className={cn(
                      "h-[52px] w-full rounded-xl border bg-white pl-11 pr-4 text-sm text-slate-900 transition-colors",
                      "placeholder:text-slate-400 focus:outline-none focus:ring-4 focus:ring-[#17385D]/[.08]",
                      errors.email ? "border-red-400" : "border-slate-300 focus:border-[#17385D]/40"
                    )}
                    placeholder="nama@pu.go.id"
                  />
                </div>
                {errors.email && (
                  <p className="mt-1.5 text-xs text-red-600">{errors.email.message}</p>
                )}
              </div>

              <div>
                <label htmlFor="password" className="mb-2 block text-sm font-semibold text-[#102a52]">
                  Password
                </label>
                <div className="relative">
                  <LockKeyhole className="pointer-events-none absolute left-3.5 top-1/2 h-5 w-5 -translate-y-1/2 text-[#1554b8]" aria-hidden="true" />
                  <input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    autoComplete="current-password"
                    {...register("password")}
                    className={cn(
                      "h-[52px] w-full rounded-xl border bg-white pl-11 pr-12 text-sm text-slate-900 transition-colors",
                      "placeholder:text-slate-400 focus:outline-none focus:ring-4 focus:ring-[#17385D]/[.08]",
                      errors.password ? "border-red-400" : "border-slate-300 focus:border-[#17385D]/40"
                    )}
                    placeholder="Masukkan password"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword((current) => !current)}
                    className="absolute right-2 top-1/2 inline-flex h-9 w-9 -translate-y-1/2 items-center justify-center rounded-md text-slate-500 transition-colors hover:bg-slate-100 hover:text-[#082e63] focus:outline-none focus:ring-2 focus:ring-[#1554b8]/30"
                    aria-label={showPassword ? "Sembunyikan password" : "Tampilkan password"}
                    title={showPassword ? "Sembunyikan password" : "Tampilkan password"}
                  >
                    {showPassword ? (
                      <EyeOff className="h-5 w-5" aria-hidden="true" />
                    ) : (
                      <Eye className="h-5 w-5" aria-hidden="true" />
                    )}
                  </button>
                </div>
                {errors.password && (
                  <p className="mt-1.5 text-xs text-red-600">{errors.password.message}</p>
                )}
              </div>

              <div className="flex items-center justify-between gap-4 text-sm">
                <label className="flex cursor-pointer items-center gap-2 text-slate-600">
                  <input
                    type="checkbox"
                    {...register("rememberMe")}
                    className="h-4 w-4 rounded border-slate-300 text-[#1554b8] focus:ring-[#1554b8]"
                  />
                  Ingat saya
                </label>
                <Link href="/forgot-password" className="font-medium text-[#1554b8] hover:underline">
                  Lupa password?
                </Link>
              </div>

              <button
                type="submit"
                disabled={isSubmitting}
                className={cn(
                  "inline-flex h-[52px] w-full items-center justify-center gap-2 rounded-xl bg-[#17385D] px-4 text-sm font-semibold text-white shadow-[0_8px_18px_rgba(23,56,93,0.16)]",
                  "transition-all hover:-translate-y-0.5 hover:bg-[#1E416B] hover:shadow-[0_12px_24px_rgba(23,56,93,0.2)] focus:outline-none focus:ring-4 focus:ring-[#17385D]/20 focus:ring-offset-2",
                  "disabled:cursor-not-allowed disabled:opacity-60"
                )}
              >
                <LogIn className="h-5 w-5" aria-hidden="true" />
                {isSubmitting ? "Memproses..." : "Masuk"}
              </button>
            </form>

            <div className="mt-5 flex items-center justify-center gap-2 border-t border-slate-100 pt-4 text-xs text-slate-500">
              <ShieldCheck className="h-4 w-4 text-slate-600" aria-hidden="true" />
              <span>Akses terbatas untuk pengguna berwenang</span>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
