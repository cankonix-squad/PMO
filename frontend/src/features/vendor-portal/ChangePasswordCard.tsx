"use client";

import { useState, type FormEvent } from "react";
import { KeyRound } from "lucide-react";
import { authService } from "@/services/auth.service";
import { apiErrorMessage } from "@/services/vendor-portal.service";
import { useAuthStore } from "@/store/auth.store";
import { btnPrimary, Field, inputCls } from "./ui";

/** Shown to contractor accounts until the temporary password is changed. */
export function ChangePasswordCard() {
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    if (next.length < 8) return setError("Password baru minimal 8 karakter.");
    if (!/[A-Z]/.test(next) || !/[a-z]/.test(next) || !/[0-9]/.test(next)) {
      return setError("Password baru harus mengandung huruf besar, huruf kecil, dan angka.");
    }
    if (next !== confirm) return setError("Konfirmasi password tidak sama.");
    if (next === current) return setError("Password baru harus berbeda dari password sementara.");
    setSaving(true);
    setError(null);
    try {
      await authService.changePassword({ current_password: current, new_password: next });
      const user = useAuthStore.getState().user;
      if (user) useAuthStore.setState({ user: { ...user, must_change_pwd: false } });
    } catch (err) {
      setError(apiErrorMessage(err, "Gagal mengganti password."));
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="mx-auto max-w-md rounded-xl border border-amber-200 bg-white p-5 shadow-sm">
      <div className="mb-4 flex items-center gap-2">
        <KeyRound className="h-5 w-5 text-amber-600" />
        <h2 className="text-base font-semibold text-gray-900">Ganti Password Sementara</h2>
      </div>
      <p className="mb-4 text-sm text-gray-600">
        Akun Anda dibuat oleh admin dengan password sementara. Ganti password sebelum menggunakan portal.
      </p>
      <form onSubmit={submit} className="space-y-3">
        <Field label="Password sementara" required>
          <input className={inputCls} type="password" autoComplete="current-password" value={current} onChange={(e) => setCurrent(e.target.value)} />
        </Field>
        <Field label="Password baru" required>
          <input className={inputCls} type="password" autoComplete="new-password" value={next} onChange={(e) => setNext(e.target.value)} />
        </Field>
        <Field label="Ulangi password baru" required>
          <input className={inputCls} type="password" autoComplete="new-password" value={confirm} onChange={(e) => setConfirm(e.target.value)} />
        </Field>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <button type="submit" className={btnPrimary} disabled={saving || !current}>
          {saving ? "Menyimpan..." : "Simpan Password"}
        </button>
      </form>
    </div>
  );
}
