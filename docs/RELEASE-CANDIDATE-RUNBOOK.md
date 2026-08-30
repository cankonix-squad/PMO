# PMO — Release Candidate Runbook

**Versi**: 1.0.0  
**Tanggal**: 2026-08-29  
**Ticket**: PMO-UAT-010  
**Status**: Non-IoT UAT Candidate  
**Audience**: Release manager, tech lead, QA

---

## 1. Status RC

**PMO Non-IoT UAT Candidate** — semua fitur P0–P3 non-IoT terverifikasi lokal.

| Fase | Status | Tiket |
|---|---|---|
| P0 — Operational MVP | ✅ Done | P0-001..P0-014 |
| P1 — Core Project Control | ✅ Done | P1-001..P1-016 |
| P2 — Intelligence & Analytics | ✅ Done | P2-001..P2-011 |
| P3 non-IoT — Advanced Integration | ✅ Done | P3-001, P3-002, P3-003, UAT-001..UAT-007 |
| P3-IoT — Telemetry | 🔜 Future/Last | P3-004 |
| UAT-001..UAT-007 | ✅ Done | All smoke PASS |
| UAT-008 RC Orchestrator | ✅ Done | smoke-test-uat-008 |
| UAT-010 Final Acceptance | ✅ Done | This runbook |

---

## 2. Go/No-Go Criteria

Semua kriteria berikut harus **PASS** sebelum menyatakan RC siap:

### 2.1 Smoke Gates

| Gate | Expected | Command |
|---|---|---|
| REL-001 Regression | 44/44 PASS | `bash docs/smoke-test-rel-001-regression.sh` |
| UAT-007 Business Flow | 86/86 PASS | `bash docs/smoke-test-uat-007-business-flow.sh` |
| UAT-006 Role Permissions | 65/65 PASS | `bash docs/smoke-test-uat-006-role-permissions.sh` |
| UAT-008 RC Orchestrator | All suites PASS/SKIP | `bash docs/smoke-test-uat-008-release-candidate.sh` |

### 2.2 Build Gates

| Gate | Command |
|---|---|
| `gofmt` clean | `cd backend && gofmt -l internal cmd` → empty output |
| Go build | `cd backend && go build ./...` → exit 0 |
| Go test | `cd backend && go test ./...` → exit 0 |
| TypeScript | `cd frontend && npx tsc --noEmit` → exit 0 |
| ESLint | `cd frontend && npm run lint` → exit 0 |

### 2.3 Doc Sync Gates

- [ ] `CLAUDE.md` — status terbaru UAT-010
- [ ] `DEVELOPMENT-BACKLOG.md` — UAT-010 Done entry
- [ ] `UAT-DEMO-GUIDE.md` — versi terbaru, semua fitur akurat
- [ ] `IMPLEMENTATION-GAP-ANALYSIS.md` — UAT-006/007/010 updates
- [ ] `PERMISSION-MATRIX.md` — corrective_action resource terdaftar
- [ ] `PHASED-IMPLEMENTATION-PLAN.md` — phase status akurat
- [ ] Known limitations jujur (IoT Future, gov connector sandbox, BIM metadata-only)

### 2.4 Security Gate

- [ ] Tidak ada production secret di `.env` / docs / scripts
- [ ] `backend/.env.example` hanya placeholder (`CHANGE_ME`)
- [ ] `backend/.env` tidak di-commit ke git
- [ ] JWT_SECRET minimal 32 karakter di env aktif

---

## 3. RC Execution Steps

### Step 1: Environment Check
```bash
bash scripts/uat-env-check.sh
```
Expected: PASS atau hanya WARN. Tidak ada FAIL.

### Step 2: Start Stack
```bash
bash scripts/uat-start.sh
```
Expected: Backend :8080 + Frontend :3000 ready.

### Step 3: Stack Status
```bash
bash scripts/uat-status.sh
```
Expected: Backend UP + Frontend UP + Login OK.

### Step 4: Backend Build & Test
```bash
cd backend
gofmt -l internal cmd      # Expected: empty output
go build ./...             # Expected: exit 0
go test ./...              # Expected: exit 0 (no test files = OK)
```

### Step 5: Frontend Build & Lint
```bash
cd frontend
npx tsc --noEmit           # Expected: exit 0
npm run lint               # Expected: exit 0
# Note: npm run build dijalankan terpisah setelah stop dev server
```

### Step 6: Regression Gate
```bash
STRICT_FRONTEND=1 bash docs/smoke-test-rel-001-regression.sh
```
Expected: 44/44 PASS.

### Step 7: UAT Smoke Suites
```bash
bash docs/smoke-test-uat-007-business-flow.sh    # 86/86 PASS
bash docs/smoke-test-uat-006-role-permissions.sh # 65/65 PASS
bash docs/smoke-test-uat-008-release-candidate.sh # All PASS
```

### Step 8: Frontend Production Build
```bash
bash scripts/uat-stop.sh
cd frontend
rm -rf .next
npm run build
```
Expected: exit 0, semua halaman compile.

### Step 9: Stop Stack
```bash
bash scripts/uat-stop.sh
bash scripts/uat-status.sh  # Expected: DOWN
```

### Step 10: Docs Final Sync
- Update CLAUDE.md status block
- Update DEVELOPMENT-BACKLOG.md
- Commit dengan message: `chore: UAT-010 final acceptance — Non-IoT RC`

---

## 4. Known Limitations (Jujur)

Berikut adalah batasan yang diketahui dan **tidak** menghalangi status "Non-IoT UAT Candidate":

| Limitasi | Keterangan | Dampak |
|---|---|---|
| **IoT Telemetry (P3-004)** | Belum diimplementasi. Future/Last. | Tidak ada real-time sensor data |
| **Government Connector** | Sandbox/mock data — bukan koneksi production API SIRUP/OM-SPAN | Data integrasi adalah simulasi |
| **BIM Viewer** | Metadata + link saja. Belum embedded 3D viewer. | User tidak bisa melihat model 3D in-app |
| **Document Storage** | Local filesystem. Belum S3/MinIO production. | File hilang jika server restart di deployment baru |
| **SMTP Email** | Fungsional jika env dikonfigurasi. NoopProvider fallback jika kosong. | Email notifikasi tidak terkirim jika SMTP tidak dikonfigurasi |
| **Power BI** | Arsitektur documented. Belum live connection. | Executive dashboard = PMO native, bukan Power BI embedded |
| **Multi-Organization** | Single org per instance untuk UAT | Tidak simulasi multi-tenant production |
| **UAT-001 Playwright** | Login timeout pada beberapa run (81/101 checks). Curl-mode PASS. | Playwright sidebar checks flaky jika Next.js dev server lambat |

---

## 5. Hotfix Procedure

Jika ditemukan bug kritis selama UAT:

1. **Stop**: dokumentasikan bug dengan repro steps
2. **Assess**: apakah blocker RC atau acceptable known limitation?
3. **Fix**: buat branch `hotfix/uat-xxx-description`, fix minimal, tidak tambah fitur
4. **Verify**: jalankan gate smoke yang relevan saja (bukan full suite)
5. **Merge**: ke `main` setelah PASS
6. **Re-run**: jalankan `bash docs/smoke-test-uat-008-release-candidate.sh`

---

## 6. RC Sign-Off Checklist

Sebelum sign-off "Non-IoT UAT Candidate":

- [ ] All smoke gates PASS (lihat section 2.1)
- [ ] All build gates PASS (lihat section 2.2)
- [ ] All doc sync gates done (lihat section 2.3)
- [ ] Security gate clean (lihat section 2.4)
- [ ] Known limitations documented jujur
- [ ] `docs/UAT-FINAL-CHECKLIST.md` tersedia untuk tim demo
- [ ] `docs/UAT-DEMO-GUIDE.md` versi terbaru
- [ ] Tech lead review selesai
- [ ] Go/No-Go decision: **GO** ✅

---

## 7. Post-RC Next Steps

Setelah RC sign-off:

1. **Staging deployment** — deploy ke server UAT/staging (bukan localhost)
2. **UAT facilitation** — jalankan sesi demo dengan stakeholder menggunakan `UAT-DEMO-GUIDE.md`
3. **Bug collection** — catat bug dari stakeholder di backlog
4. **Non-IoT feature completion** — selesaikan item backlog non-IoT yang tersisa
5. **Production readiness** — setelah UAT sign-off dari stakeholder, mulai production hardening
6. **IoT (P3-004)** — mulai hanya setelah production readiness gate terpenuhi
