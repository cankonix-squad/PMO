"use client";

import Image from "next/image";
import Link from "next/link";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Eye,
  EyeOff,
  LockKeyhole,
  LogIn,
  ShieldCheck,
  UserRound,
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

function getSafeRedirectPath(search: string) {
  const from = new URLSearchParams(search).get("from");
  if (!from || !from.startsWith("/") || from.startsWith("//")) {
    return "/dashboard";
  }
  return from;
}

function BrandMark({ compact = false, small = false }: { compact?: boolean; small?: boolean }) {
  const size = compact || small ? 48 : 56;
  return (
    <div
      className={cn(
        "grid shrink-0 place-items-center overflow-hidden",
        size === 48 ? "h-12 w-12" : "h-14 w-14"
      )}
      aria-hidden="true"
    >
      <Image
        src="/images/logo-kemenpu.png"
        alt="Logo Kementerian Pekerjaan Umum"
        width={size}
        height={size}
        priority
        className="h-full w-full object-contain"
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
    <main className="relative isolate min-h-screen overflow-x-hidden bg-[#F4F6F8] lg:grid lg:grid-cols-[minmax(0,1fr)_minmax(0,1.63fr)]">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-0 -z-10 bg-[url('/images/pmo-architecture.png')] bg-cover bg-center bg-no-repeat lg:bg-[length:100%_100%]"
      />
      <section className="relative order-2 hidden min-h-screen overflow-hidden bg-transparent lg:order-2 lg:flex lg:flex-col">
        <header className="relative z-10 flex items-center justify-end px-10 pt-9 xl:px-14 xl:pt-11">
          <div className="rounded-full border border-[#17385D]/10 bg-white/85 px-4 py-2 text-xs font-medium tracking-wide text-[#17385D]/70">
            PROJECT MANAGEMENT OFFICE
          </div>
        </header>

        <div className="relative z-10 flex flex-1 flex-col justify-center pb-28 pl-8 pr-10 pt-4 xl:pl-12 xl:pr-14">
          <div className="min-w-0 max-w-[640px]">
            <div className="mb-7 flex items-center gap-4">
              <BrandMark small />
              <div>
                <p className="mb-1 text-xs font-semibold uppercase tracking-[.12em] text-[#17385D] xl:text-sm">
                  Kementerian Pekerjaan Umum
                </p>
                <p className="text-xs font-bold uppercase tracking-[.2em] text-[#A77508] xl:text-sm">
                  PMO DIGITAL PLATFORM
                </p>
              </div>
            </div>
            <div className="relative before:pointer-events-none before:absolute before:-inset-x-6 before:-inset-y-8 before:-z-10 before:bg-[radial-gradient(ellipse_at_38%_50%,rgba(255,255,255,0.96)_0%,rgba(255,255,255,0.88)_45%,rgba(255,255,255,0)_100%)] before:content-['']">
              <h1 className="max-w-[570px] text-4xl font-bold leading-[1.04] tracking-[-.04em] text-[#17385D] xl:text-5xl">
                Pantau proyek.<br /><span className="text-[#EBAA16]">Bangun dampak.</span>
              </h1>
              <p className="mt-5 max-w-[660px] text-base leading-7 text-slate-600 xl:text-lg">
                PMO Digital Platform menyatukan data, progres, dan risiko proyek dalam satu ruang kendali untuk keputusan yang lebih cepat.
              </p>
              <div className="mt-7 h-1 w-14 rounded-full bg-[#EBAA16]" />
            </div>
          </div>
        </div>

      </section>

      <section className="relative order-1 flex min-h-screen flex-col bg-transparent lg:order-1 lg:justify-center lg:pl-7 lg:pr-0 xl:pl-10">
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

        <div className="relative z-10 flex flex-1 items-center justify-center px-4 py-6 sm:px-6 lg:justify-end lg:px-0 lg:pb-16 lg:pt-4">
          <div className="w-full max-w-[440px] rounded-[24px] border border-[#17385D]/15 bg-white/55 px-6 py-6 shadow-[0_20px_56px_rgba(23,56,93,0.06)] backdrop-blur-[2px] sm:px-9 sm:py-7">
            <div className="mb-6">
              <div className="mb-6 flex items-center gap-3">
                <div className="grid h-11 w-11 shrink-0 place-items-center">
                  <Image src="/images/logo-kemenpu.png" alt="Logo Kementerian Pekerjaan Umum" width={44} height={44} priority className="h-11 w-11 object-contain" />
                </div>
                <div className="text-left">
                  <p className="text-[11px] font-bold uppercase tracking-[.14em] text-[#17385D]">Kementerian PU</p>
                  <p className="mt-0.5 text-xs text-[#64748B]">PMO Digital Platform</p>
                </div>
              </div>
              <p className="mb-2 text-xs font-semibold uppercase tracking-[.13em] text-[#A77508]">Selamat datang kembali</p>
              <h2 className="text-2xl sm:text-[26px] font-bold tracking-[-.04em] text-[#17385D]">Masuk ke akun Anda</h2>
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
