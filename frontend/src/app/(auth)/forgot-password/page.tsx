"use client";

import Link from "next/link";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowLeft, Mail, Send, ShieldCheck } from "lucide-react";
import { z } from "zod";
import { authService } from "@/services/auth.service";
import { cn } from "@/lib/utils";

const forgotSchema = z.object({
  email: z.string().email("Masukkan alamat email yang valid"),
});

type ForgotForm = z.infer<typeof forgotSchema>;

export default function ForgotPasswordPage() {
  const [serverError, setServerError] = useState<string | null>(null);
  const [submittedEmail, setSubmittedEmail] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ForgotForm>({
    resolver: zodResolver(forgotSchema),
    defaultValues: { email: "" },
  });

  const onSubmit = async (values: ForgotForm) => {
    setServerError(null);
    setSubmittedEmail(null);
    try {
      await authService.forgotPassword({ email: values.email });
      setSubmittedEmail(values.email);
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data
          ?.message ?? "Permintaan reset password belum dapat diproses.";
      setServerError(msg);
    }
  };

  return (
    <main className="flex min-h-screen items-center justify-center bg-[#f7f9fc] px-5 py-8">
      <div className="w-full max-w-md rounded-lg border border-slate-200 bg-white px-6 py-8 shadow-[0_18px_50px_rgba(15,45,85,0.12)] sm:px-9 sm:py-10">
        <Link
          href="/login"
          className="mb-7 inline-flex items-center gap-2 text-sm font-medium text-[#1554b8] hover:underline"
        >
          <ArrowLeft className="h-4 w-4" aria-hidden="true" />
          Kembali ke login
        </Link>

        <div className="mb-8">
          <div className="mb-4 grid h-12 w-12 place-items-center rounded-md bg-[#eef5ff] text-[#1554b8]">
            <ShieldCheck className="h-6 w-6" aria-hidden="true" />
          </div>
          <h1 className="text-2xl font-bold text-[#082e63]">Reset Password</h1>
          <p className="mt-2 text-sm leading-6 text-slate-500">
            Masukkan email akun CANKORA. Jika akun terdaftar, instruksi reset
            akan dikirim sesuai konfigurasi notifikasi sistem.
          </p>
        </div>

        {submittedEmail && (
          <div
            role="status"
            className="mb-5 rounded-md border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700"
          >
            Permintaan reset untuk {submittedEmail} sudah diterima.
          </div>
        )}

        {serverError && (
          <div
            role="alert"
            className="mb-5 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
          >
            {serverError}
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5" noValidate>
          <div>
            <label htmlFor="email" className="mb-2 block text-sm font-semibold text-[#102a52]">
              Email
            </label>
            <div className="relative">
              <Mail className="pointer-events-none absolute left-3.5 top-1/2 h-5 w-5 -translate-y-1/2 text-[#1554b8]" aria-hidden="true" />
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

          <button
            type="submit"
            disabled={isSubmitting}
            className="inline-flex h-12 w-full items-center justify-center gap-2 rounded-md bg-[#0d4ba8] px-4 text-sm font-semibold text-white transition-colors hover:bg-[#083b86] focus:outline-none focus:ring-2 focus:ring-[#1554b8] focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60"
          >
            <Send className="h-5 w-5" aria-hidden="true" />
            {isSubmitting ? "Mengirim..." : "Kirim Instruksi Reset"}
          </button>
        </form>
      </div>
    </main>
  );
}
