# PMO — UAT Deployment Guide

**Versi**: 1.0.0  
**Tanggal**: 2026-08-29  
**Ticket**: PMO-UAT-010  
**Status**: Non-IoT UAT Candidate  
**Audience**: QA engineer, DevOps, release manager

---

## 1. Scope

Dokumen ini menjelaskan cara men-deploy PMO untuk sesi UAT (User Acceptance Testing) di lingkungan **lokal** atau **UAT server**. Target status: **Non-IoT UAT Candidate** — semua fitur P0–P3 non-IoT terverifikasi, IoT telemetry tetap Future/Last.

Tidak untuk produksi. Gunakan prosedur ini hanya untuk UAT/demo.

---

## 2. Prasyarat

| Komponen | Versi | Keterangan |
|---|---|---|
| Go | 1.23+ | Backend runtime |
| Node.js | 20 LTS | Frontend build & dev |
| npm | 10+ | Frontend package manager |
| PostgreSQL | 15 | Database (Homebrew atau Docker) |
| Redis | 7 (opsional) | Tidak diperlukan untuk UAT |
| MinIO | (opsional) | Storage lokal cukup untuk UAT |
| curl | any | Smoke test dependency |
| bash | 5+ | Script compatibility |

### Catatan macOS
```bash
# Install Go via Homebrew
brew install go node postgresql@15

# Start PostgreSQL
brew services start postgresql@15
```

---

## 3. Setup Lokal (Quick Start)

```bash
# 1. Clone/buka workspace
cd '/Users/harmanto/Documents/Code/DEV/Project Management/PMO'

# 2. Run environment check
bash scripts/uat-env-check.sh

# 3. Start stack (migrate + seed + backend + frontend)
bash scripts/uat-start.sh

# 4. Verify
bash scripts/uat-status.sh
```

---

## 4. Setup Manual (step-by-step)

### 4.1 Database

```bash
# Buat user dan database (jika belum ada)
psql -U postgres -c "CREATE USER cankora WITH PASSWORD 'cankora_secret';"
psql -U postgres -c "CREATE DATABASE cankora_db OWNER cankora;"
```

### 4.2 Backend Environment

```bash
cd backend
cp .env.example .env
```

Edit `backend/.env`:
```
APP_ENV=development
DB_HOST=localhost
DB_PORT=5432
DB_USER=cankora
DB_PASSWORD=cankora_secret
DB_NAME=cankora_db
DB_SSLMODE=disable
JWT_SECRET=<ganti-dengan-string-32-karakter-minimal>
PORT=8080
APP_URL=http://localhost:3000
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

> **Catatan**: `JWT_SECRET` harus minimal 32 karakter. Untuk UAT lokal, nilai placeholder dari `.env.example` dapat digunakan. Jangan gunakan nilai produksi.

### 4.3 Migration & Seed

```bash
cd backend
make migrate-up    # Apply semua migration (s.d. 000037+)
make seed          # Org + RBAC + admin user (idempotent)
make seed-demo     # 10 proyek SDA demo (idempotent)
```

### 4.4 Backend Start

```bash
cd backend
make dev           # Port 8080
```

Verifikasi:
```bash
curl http://localhost:8080/health
# Expected: {"success":true,"message":"ok","data":{"status":"ok","version":"0.1.0"}}
```

### 4.5 Frontend Start

```bash
cd frontend
rm -rf .next       # PENTING: hapus build lama untuk mencegah CSS corruption
npm install        # Skip jika node_modules sudah ada
npm run dev        # Port 3000
```

> **RULE**: Jangan jalankan `npm run build` saat dev server aktif menggunakan `.next` yang sama. Ini menyebabkan CSS corruption (sidebar/topbar hilang).

---

## 5. Akun UAT

| Email | Password | Role |
|---|---|---|
| admin@cankora.local | Admin@Cankora2024! | SUPER_ADMIN / ADMIN |
| pmo@cankora.local | Demo@Cankora2024! | PMO |
| pm@cankora.local | Demo@Cankora2024! | PROJECT_MANAGER |
| officer@cankora.local | Demo@Cankora2024! | PROJECT_OFFICER |
| executive@cankora.local | Demo@Cankora2024! | EXECUTIVE_VIEWER |
| auditor@cankora.local | Demo@Cankora2024! | AUDITOR |

---

## 6. Migrasi (Urutan)

Semua migration bersifat **idempotent** (aman dijalankan ulang):

| No | File | Deskripsi |
|---|---|---|
| 000001 | create_organizations | Organizations + org units |
| 000002 | create_users_auth | Users, sessions |
| 000003 | create_rbac | Roles, permissions, user_roles |
| 000004 | create_audit_logs | Audit log table |
| 000005 | create_projects_core | Projects, tasks, milestones |
| 000006 | align_milestones_schema | Milestone field alignment |
| 000007 | child_soft_delete | Soft delete pada child tables |
| 000008 | issues_tenant_escalation | Issues + organization_id |
| 000009 | risks_tenant_score | Risks + probability/impact/score |
| 000010 | vendors_contracts | Vendors + contracts |
| 000011 | documents_version | Document version + updated_at |
| 000012–000025 | (P1/P2 features) | Field inspection, analytics, BIM, Primavera, dll |
| 000026 | government_connector | Government integration tables |
| 000027 | bim_models | BIM model tables |
| 000028 | government_mapping_match_status | Match status untuk entity resolution |
| 000029 | government_resolution_metadata | Confidence, reason, audit fields |
| 000030–000035 | governance | Data governance tables + constraints |
| 000036 | report_export_file_metadata | Report export file columns |
| 000037 | notifications | Notifications table |

---

## 7. Known Limitations UAT

Lihat `docs/UAT-DEMO-GUIDE.md` section Known Limitations untuk daftar lengkap. Ringkasan:

| Limitasi | Status |
|---|---|
| IoT Telemetry (P3-004) | Future/Last — belum diimplementasi |
| Government connector | Sandbox/mock — bukan production API |
| BIM viewer | Metadata + link saja — belum embedded viewer |
| Storage dokumen | Lokal filesystem — belum S3/MinIO production |
| SMTP | Fungsional jika env dikonfigurasi — NoopProvider fallback |
| Power BI | Arsitektur documented — belum terkoneksi |
| Multi-org | Single org per instance untuk UAT |

---

## 8. Troubleshooting

| Gejala | Penyebab | Fix |
|---|---|---|
| Dashboard full-screen tanpa sidebar | `.next` corrupt | `rm -rf frontend/.next && restart npm run dev` |
| Backend 500 saat login | DB tidak ready | `brew services start postgresql@15` |
| Migration gagal | Versi migration tidak cocok | `cd backend && make migrate-down && make migrate-up` |
| Port 8080 in use | Backend sudah berjalan | `bash scripts/uat-stop.sh` lalu `bash scripts/uat-start.sh` |
| SMTP email tidak terkirim | SMTP env kosong | Normal — NoopProvider aktif. Set SMTP_HOST/USER/PASSWORD untuk email nyata |
| Seed-demo duplikat | Seed dijalankan ulang | Aman — `make seed-demo` idempotent via `ON CONFLICT DO NOTHING` |

---

## 9. Stop Stack

```bash
bash scripts/uat-stop.sh
```

---

## 10. Smoke Test

Setelah stack berjalan:

```bash
# Quick regression check
bash docs/smoke-test-rel-001-regression.sh

# Full release candidate orchestrator
bash docs/smoke-test-uat-008-release-candidate.sh

# Individual UAT suites
bash docs/smoke-test-uat-007-business-flow.sh  # Business flow E2E (86/86)
bash docs/smoke-test-uat-006-role-permissions.sh  # RBAC (65/65)
```

---

## 11. File yang Relevan

| File | Deskripsi |
|---|---|
| `scripts/uat-env-check.sh` | Pre-flight environment check |
| `scripts/uat-start.sh` | Start backend + frontend |
| `scripts/uat-status.sh` | Check stack status |
| `scripts/uat-stop.sh` | Stop stack |
| `docs/smoke-test-uat-008-release-candidate.sh` | RC smoke orchestrator |
| `docs/UAT-DEMO-GUIDE.md` | Demo guide per role |
| `docs/UAT-FINAL-CHECKLIST.md` | Pre/during/post demo checklist |
| `docs/RELEASE-CANDIDATE-RUNBOOK.md` | RC gate runbook |
| `backend/.env.example` | Environment template |
