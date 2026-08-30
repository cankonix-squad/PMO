# PMO — UAT Final Checklist

**Versi**: 1.0.0  
**Tanggal**: 2026-08-29  
**Ticket**: PMO-UAT-010  
**Status**: Non-IoT UAT Candidate  
**Audience**: Demo facilitator, QA, product owner, stakeholder

---

## 1. Pre-Demo Checklist

Jalankan semua item ini **sebelum** sesi demo dimulai (rekomendasi: 1 jam sebelum).

### 1.1 Environment

- [ ] Laptop/server dengan Go 1.23+, Node.js 20 LTS, PostgreSQL 15 tersedia
- [ ] Koneksi internet tidak diperlukan (demo berjalan lokal)
- [ ] Cukup RAM untuk backend + frontend + DB (rekomendasi: 8GB+)
- [ ] Browser modern tersedia (Chrome/Firefox terbaru)

### 1.2 Stack Startup

```bash
bash scripts/uat-env-check.sh   # Harus: PASS atau WARN saja, tidak ada FAIL
bash scripts/uat-start.sh       # Start backend + frontend
bash scripts/uat-status.sh      # Harus: Backend UP + Frontend UP + Login OK
```

- [ ] `uat-env-check.sh` → no FAIL
- [ ] `uat-start.sh` → backend :8080 ready + frontend :3000 ready
- [ ] `uat-status.sh` → UP + Login OK
- [ ] Buka http://localhost:3000 → tampil halaman login
- [ ] Login sebagai `admin@cankora.local` / `Admin@Cankora2024!` → redirect ke `/dashboard`
- [ ] Dashboard menampilkan data (proyek, warnings, stats) — bukan blank

### 1.3 Smoke Quick Check

```bash
# Quick regression (2 menit)
bash docs/smoke-test-rel-001-regression.sh   # Expected: 44/44 PASS
```

- [ ] Regression gate PASS
- [ ] Tidak ada error merah di terminal

### 1.4 Demo Data

- [ ] 10 proyek DEMO-SDA-* tersedia (`make seed-demo` sudah dijalankan)
- [ ] Minimal 1 proyek dengan status ACTIVE
- [ ] Minimal 1 proyek dengan risiko terdaftar
- [ ] Dashboard menampilkan early warnings
- [ ] Demo users login:
  - [ ] `pmo@cankora.local` → Dashboard, Projects, Command Center terlihat
  - [ ] `executive@cankora.local` → Analytics visible, Create Project hidden (403)

### 1.5 Browser Setup

- [ ] Browser zoom 100%
- [ ] Browser fullscreen atau 1366×768+ window
- [ ] Private/incognito mode untuk akun demo (agar tidak tercampur sesi admin)
- [ ] Tab siap: `/dashboard`, `/projects`, `/command-center`, `/governance`

---

## 2. During-Demo Checklist

### 2.1 Per Skenario (Centang saat demo berlangsung)

#### Skenario 1: Project Lifecycle (5 menit)
- [ ] Buat proyek baru → DRAFT status
- [ ] Transisi: DRAFT → PLANNING → ACTIVE
- [ ] Tambah milestone + task
- [ ] Tampilkan detail project control

#### Skenario 2: Risk & Issue (3 menit)
- [ ] Buat risk (probability × impact = score CRITICAL)
- [ ] Transisi risk IDENTIFIED → ASSESSED
- [ ] Buat issue → transisi IN_PROGRESS → RESOLVED
- [ ] Tampilkan di Command Center

#### Skenario 3: Budget & Vendor (3 menit)
- [ ] Buat budget line → status NORMAL
- [ ] Update actual > 90% → status RISK
- [ ] Buat vendor + contract
- [ ] Dashboard menampilkan budget warning

#### Skenario 4: Field Evidence (3 menit)
- [ ] Upload dokumen (PDF/image)
- [ ] Buat inspeksi lapangan (form)
- [ ] Upload evidence ke inspeksi

#### Skenario 5: Governance & Analytics (5 menit)
- [ ] Tampilkan `/governance` — data submission
- [ ] Tampilkan `/executive` — KPI nasional
- [ ] Tampilkan `/programs` — aggregasi per program
- [ ] Tampilkan `/reports/analytics` — export

#### Skenario 6: RBAC Demo (2 menit)
- [ ] Login sebagai `officer@cankora.local` → menu terbatas
- [ ] Coba akses admin area → 403 (expected)
- [ ] Login sebagai `executive@cankora.local` → analytics visible, no create

#### Skenario 7: Notifications & Audit (2 menit)
- [ ] Tampilkan `/notifications`
- [ ] Tampilkan `/audit-logs` — riwayat aktivitas

### 2.2 Contingency (jika ada masalah)

| Gejala | Quick Fix |
|---|---|
| Sidebar/topbar hilang | Tab baru → login ulang |
| Map full-screen | `rm -rf frontend/.next` lalu restart frontend |
| Backend 500 | `bash scripts/uat-status.sh` → restart jika perlu |
| Data kosong | `cd backend && make seed-demo` |
| Login gagal | Cek backend status: `curl http://localhost:8080/health` |

---

## 3. Post-Demo Checklist

### 3.1 Feedback Collection

- [ ] Catat semua pertanyaan/masukan stakeholder
- [ ] Identifikasi: bug report vs feature request vs known limitation
- [ ] Known limitation yang valid → arahkan ke `docs/UAT-DEMO-GUIDE.md` section Known Limitations
- [ ] Bug kritis → buat tiket di backlog segera

### 3.2 Cleanup

```bash
bash scripts/uat-stop.sh
bash scripts/uat-status.sh   # Expected: DOWN
```

- [ ] Stack stopped bersih
- [ ] PID file dihapus

### 3.3 Bug Triage

Untuk setiap bug yang ditemukan:

| Bug | Severity | Blocker? | Tiket |
|---|---|---|---|
| | | | |

Severity:
- **P0 — Blocker**: crash, data loss, auth bypass — must fix before next demo
- **P1 — Major**: fitur utama tidak berjalan — fix dalam sprint berikutnya
- **P2 — Minor**: UI glitch, copy salah, edge case — fix di iterasi berikutnya
- **P3 — Nice to have**: enhancement request — masuk backlog

---

## 4. Go/No-Go Criteria

### 4.1 Go (UAT Candidate siap)

Semua kriteria berikut harus terpenuhi:

| Kriteria | Status |
|---|---|
| Smoke regression 44/44 PASS | ✅ |
| Business flow smoke 86/86 PASS | ✅ |
| RBAC smoke 65/65 PASS | ✅ |
| Go build clean | ✅ |
| TypeScript check clean | ✅ |
| ESLint clean | ✅ |
| No production secrets in repo | ✅ |
| Known limitations documented | ✅ |
| Demo guide updated | ✅ |
| Stack start/stop repeatable | ✅ |
| IoT tidak diklaim sebagai Done | ✅ |

### 4.2 No-Go (blokir UAT)

Kondisi berikut men-blokir UAT:

- ❌ Backend crash pada operasi dasar (login, list projects, create)
- ❌ Auth bypass (akses resource tanpa token, cross-tenant data leak)
- ❌ Data loss pada CRUD operasi normal
- ❌ Smoke regression < 40/44 PASS
- ❌ Build compile error
- ❌ Production secret nyata terexpose di kode/docs
- ❌ IoT diklaim implemented (P3-004 harus tetap Future/Last)

---

## 5. Known Limitations (Jujur untuk Stakeholder)

Sampaikan limitasi berikut dengan jelas kepada stakeholder sebelum demo:

| Limitasi | Penjelasan |
|---|---|
| **IoT Telemetry** | Belum diimplementasi. Data sensor real-time tidak tersedia. Future/Last. |
| **Government Connector** | Sandbox/mock. Bukan koneksi live ke SIRUP/OM-SPAN. |
| **BIM Viewer** | Metadata + link saja. Belum embedded 3D viewer in-app. |
| **Document Storage** | Local filesystem. File tidak persisten di re-deploy baru. |
| **Email Notifikasi** | Hanya berfungsi jika SMTP dikonfigurasi. Default: NoopProvider (tidak kirim). |
| **Power BI** | Arsitektur planned. Belum live integration. |
| **Multi-Org** | Single org per instance untuk UAT ini. |
| **UAT-001 Playwright** | Login timeout kadang terjadi di lingkungan lambat. Hasil curl-mode tetap PASS. |

---

## 6. Version Record

| Item | Value |
|---|---|
| PMO Version | 0.1.0-uat-candidate |
| Tanggal UAT | 2026-08-29 |
| Smoke UAT-007 | 86/86 PASS |
| Smoke REL-001 | 44/44 PASS |
| Smoke UAT-006 | 65/65 PASS |
| Build | Go 1.23 + Next.js 14 |
| DB | PostgreSQL 15 |
| Status | **Non-IoT UAT Candidate** |
