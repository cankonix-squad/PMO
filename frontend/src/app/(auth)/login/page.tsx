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
        const redirectPath =
          typeof window !== "undefined"
            ? getSafeRedirectPath(window.location.search)
            : "/dashboard";
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
    <main className="min-h-screen bg-white lg:grid lg:grid-cols-[minmax(0,1.95fr)_minmax(420px,1fr)]">
      <section className="relative hidden min-h-screen overflow-hidden lg:flex lg:flex-col lg:justify-between">
        <Image
          src="/images/cankora-dam-login.png"
          alt="Bendungan dan waduk di kawasan pegunungan"
          fill
          priority
          sizes="(min-width: 1024px) 66vw, 100vw"
          className="object-cover object-center"
        />
        <div className="absolute inset-0 bg-[#052c58]/65" />
        <div className="absolute inset-y-0 left-0 w-2 bg-[#f6c515]" />

        <header className="relative z-10 flex items-center gap-4 px-12 pt-10 xl:px-16 xl:pt-12">
          <BrandMark />
          <div className="text-sm font-semibold uppercase leading-snug text-white xl:text-base">
            <p>Kementerian Pekerjaan Umum</p>
            <p>Direktorat Jenderal Sumber Daya Air</p>
          </div>
        </header>

        <div className="relative z-10 max-w-4xl px-12 xl:px-16">
          <p className="mb-4 text-sm font-semibold uppercase text-[#f6c515]">
            CANKORA Control Tower
          </p>
          <h1 className="text-5xl font-bold leading-tight text-white xl:text-6xl">
            PMO Digital Platform
          </h1>
          <p className="mt-3 max-w-3xl text-lg text-white/85 xl:text-xl">
            Project Management Office - Direktorat Jenderal Sumber Daya Air
          </p>
          <div className="my-7 h-1 w-20 bg-[#f6c515]" />
          <p className="max-w-xl text-2xl font-semibold leading-snug text-white xl:text-3xl">
            Data akurat. Proyek terkendali. Keputusan lebih cepat.
          </p>
          <p className="mt-4 max-w-xl text-sm leading-6 text-white/75 xl:text-base">
            Platform terintegrasi untuk monitoring proyek, peringatan dini, dan
            pengambilan keputusan strategis pengelolaan sumber daya air.
          </p>
        </div>

        <div className="relative z-10 grid grid-cols-3 gap-5 border-t border-white/20 bg-[#052c58]/55 px-12 py-7 xl:px-16">
          {platformBenefits.map(({ icon: Icon, title, description }) => (
            <div key={title} className="flex min-w-0 gap-3">
              <div className="grid h-10 w-10 shrink-0 place-items-center rounded-md border border-[#f6c515]/70 text-[#f6c515]">
                <Icon className="h-5 w-5" aria-hidden="true" />
              </div>
              <div className="min-w-0">
                <p className="text-sm font-semibold text-white">{title}</p>
                <p className="mt-1 text-xs leading-5 text-white/70">{description}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="flex min-h-screen flex-col bg-[#f7f9fc] lg:justify-center lg:px-8 xl:px-12">
        <div className="relative h-48 overflow-hidden lg:hidden">
          <Image
            src="/images/cankora-dam-login.png"
            alt="Bendungan dan waduk di kawasan pegunungan"
            fill
            priority
            sizes="100vw"
            className="object-cover object-center"
          />
          <div className="absolute inset-0 bg-[#052c58]/70" />
          <div className="relative z-10 flex h-full flex-col justify-between p-5">
            <div className="flex items-center gap-3">
              <BrandMark compact />
              <div className="text-xs font-semibold uppercase leading-snug text-white">
                <p>Kementerian Pekerjaan Umum</p>
                <p>Ditjen Sumber Daya Air</p>
              </div>
            </div>
            <div>
              <p className="text-2xl font-bold text-white">PMO Digital Platform</p>
              <p className="mt-1 text-sm text-white/75">CANKORA National Project Control Tower</p>
            </div>
          </div>
        </div>

        <div className="flex flex-1 items-center justify-center px-5 py-8 lg:flex-none lg:px-0 lg:py-0">
          <div className="w-full max-w-md rounded-lg border border-slate-200 bg-white px-6 py-8 shadow-[0_18px_50px_rgba(15,45,85,0.12)] sm:px-9 sm:py-10">
            <div className="mb-8 text-center">
              <div className="mx-auto mb-5 hidden justify-center lg:flex">
                <div className="overflow-hidden rounded-md bg-white p-2 shadow-sm ring-1 ring-slate-200">
                  <Image
                    src="/images/logo-kemenpu.png"
                    alt="Logo Kementerian Pekerjaan Umum"
                    width={110}
                    height={110}
                    priority
                    className="h-auto w-auto object-contain"
                  />
                </div>
              </div>
              <h2 className="text-2xl font-bold text-[#082e63]">Masuk ke Sistem</h2>
              <div className="mx-auto mt-3 h-1 w-12 bg-[#f6c515]" />
              <p className="mt-4 text-sm text-slate-500">
                Gunakan akun CANKORA Anda untuk melanjutkan
              </p>
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
              className="space-y-5"
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
                      "h-12 w-full rounded-md border bg-white pl-11 pr-4 text-sm text-slate-900",
                      "placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-[#1554b8]/25",
                      errors.email ? "border-red-400" : "border-slate-300 focus:border-[#1554b8]"
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
                      "h-12 w-full rounded-md border bg-white pl-11 pr-12 text-sm text-slate-900",
                      "placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-[#1554b8]/25",
                      errors.password ? "border-red-400" : "border-slate-300 focus:border-[#1554b8]"
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
                  "inline-flex h-12 w-full items-center justify-center gap-2 rounded-md bg-[#0d4ba8] px-4 text-sm font-semibold text-white",
                  "transition-colors hover:bg-[#083b86] focus:outline-none focus:ring-2 focus:ring-[#1554b8] focus:ring-offset-2",
                  "disabled:cursor-not-allowed disabled:opacity-60"
                )}
              >
                <LogIn className="h-5 w-5" aria-hidden="true" />
                {isSubmitting ? "Memproses..." : "Masuk"}
              </button>
            </form>

            <div className="mt-7 flex items-center justify-center gap-2 border-t border-slate-100 pt-5 text-xs text-slate-500">
              <ShieldCheck className="h-4 w-4 text-slate-600" aria-hidden="true" />
              <span>Akses terbatas untuk pengguna berwenang</span>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
