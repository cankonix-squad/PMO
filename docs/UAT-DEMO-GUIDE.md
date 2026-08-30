# CANKORA — UAT & Demo Guide

**Versi**: 1.6.0  
**Tanggal**: 2026-08-29  
**Ticket**: CANKORA-UAT-001, CANKORA-UAT-002, CANKORA-UAT-003, CANKORA-UAT-004, CANKORA-UAT-005, CANKORA-UAT-006, CANKORA-UAT-007  
**Audience**: QA, stakeholder PMO, product owner, demo facilitator

---

## 1. Prasyarat Lokal

| Komponen | Versi | Status |
|---|---|---|
| Go | 1.23+ | Required |
| Node.js | 20 LTS | Required |
| PostgreSQL | 15 | Required (Homebrew atau Docker) |
| Redis | Opsional | Tidak diperlukan untuk UAT |
| MinIO | Opsional | Tidak diperlukan untuk UAT |

Port default: backend **:8080**, frontend **:3000**.

---

## 2. Cara Start Bersih (Fresh Clean Start)

```bash
# 1. Clone atau buka workspace
cd '/Users/harmanto/Documents/Code/DEV/Project Management/CANKORA'

# 2. Setup backend env (skip jika .env sudah ada)
cp backend/.env.example backend/.env
# Edit backend/.env: pastikan DB_HOST=localhost, DB_PASSWORD=cankora_secret

# 3. Migrasi dan seed
cd backend
make migrate-up      # apply semua migration (s.d. 000036)
make seed            # org + RBAC + admin user (idempotent)
make seed-demo       # 10 proyek SDA demo (idempotent)

# 4. Start backend
make dev             # port 8080
```

```bash
# Terminal baru — Frontend
cd '/Users/harmanto/Documents/Code/DEV/Project Management/CANKORA/frontend'

# PENTING: hapus .next sebelum dev jika ada build lama
rm -rf .next

npm install          # skip jika node_modules sudah ada
npm run dev          # port 3000
```

**Verifikasi**: buka http://localhost:3000 → tampil halaman login.

---

## 3. Akun Demo

| Email | Password | Role | Keterangan |
|---|---|---|---|
| `admin@cankora.local` | `Admin@Cankora2024!` | SUPER_ADMIN | Akses penuh semua fitur |
| `pmo@cankora.local` | `Demo@Cankora2024!` | PMO | PMO Command Center, semua resource |
| `pm@cankora.local` | `Demo@Cankora2024!` | PROJECT_MANAGER | Project ops, tasks, risks, issues |
| `officer@cankora.local` | `Demo@Cankora2024!` | PROJECT_OFFICER | Task execution, milestones, documents |
| `executive@cankora.local` | `Demo@Cankora2024!` | EXECUTIVE_VIEWER | Analytics read-only (executive/program/GIS) |
| `auditor@cankora.local` | `Demo@Cankora2024!` | AUDITOR | Audit logs + reports, view+export only |

Semua demo users di-seed idempotent via `make seed`. Tidak perlu membuat manual.

---

## 4. Data Demo yang Tersedia

Setelah `make seed` + `make seed-demo`:

| Entitas | Jumlah | Keterangan |
|---|---|---|
| Field inspections | Demo sesuai kebutuhan | Via `/projects/:id/inspections` atau panel di project detail |
| Proyek SDA | 10 | DEMO-SDA-001..010, berbagai status/kategori |
| Program | 5 | Bendungan, Sungai & Pantai, Irigasi, Air Baku, Pertanian |
| Sektor | 5 | Irigasi & Rawa, Banjir & Drainase, Air Baku & Sanitasi, Pertanian & Perdesaan, Bendungan & Tampungan |
| Tasks / Milestones | ~30 | Per proyek demo |
| Issues / Risks | ~20 | Per proyek demo |
| Budget lines | ~10 | Per proyek demo |
| GIS koordinat | 10 | Semua proyek DEMO punya lat/lng |
| Priority scores | 6 | Dari formula aktif |
| BIM models | 1+ | Model referensi |
| Primavera runs | 19+ | Riwayat sync import |
| Gov. mappings | 8 | MATCHED/PENDING_MATCH/REJECTED |
| Governance submissions | 107+ | Berbagai status lifecycle |
| Benefit indicators | 2+ | Ha irigasi, kapasitas tampungan |

Semua seed bersifat **idempotent** — aman dijalankan berulang kali tanpa membuat duplikat.

---

## 5. Urutan Demo 15–30 Menit

### Blok A — Login & Overview (3 menit)
1. Buka `http://localhost:3000/login`
2. Login dengan `admin@cankora.local` / `Admin@Cankora2024!`
3. Dashboard tampil: KPI proyek (total/aktif/overrun), distribusi status, early warning.
4. Scroll ke section **Portfolio** (anchor `#portfolio`) → tabel proyek dengan filter.
5. Tunjukkan **Monitoring** (anchor `#monitoring`) → map sebaran proyek Indonesia.

### Blok B — Project Detail & Control (5 menit)
1. Navigasi ke `/projects` → daftar proyek.
2. Klik proyek "Bendungan Marga Tiga" atau salah satu `DEMO-SDA-001`.
3. Tunjukkan tab: **Overview**, **Tasks**, **Milestones**, **Issues**, **Risks**, **Budget**, **Dokumen**, **Kontrak**.
4. Demo transisi status task: TODO → IN_PROGRESS → DONE.
5. Tunjukkan risk register: probability × impact → severity otomatis.
6. Buka `/projects/<id>/inspections` — tunjukkan panel inspeksi lapangan: form tambah inspeksi, upload foto/berkas bukti, geolocation helper, list evidence dengan checksum, tombol Verifikasi/Tolak, tampilan mobile 390px (tidak ada horizontal overflow).

### Blok C — PMO Command Center (3 menit)
1. Buka `/command-center`
2. KPI: Proyek Bermasalah, Validasi Tertunda, Tindak Lanjut Overdue, Risiko Tinggi.
3. Tunjukkan **Alert Center**, **Corrective Action Tracker**, **Watchlist Proyek**.
4. Tunjukkan **Heatmap Proyek** (ilustratif sampai koordinat resmi P2-008).

### Blok D — Analytics & Reporting (5 menit)
1. **Program Dashboard** `/programs`: KPI per program (bendungan vs irigasi vs air baku).
2. **Executive Dashboard** `/executive`: national summary, health distribution bar, critical projects, open escalations.
3. **Reporting Center** `/reports/analytics`: tab Executive Summary → Performance → Risk → Budget → Benefits → Priority.
4. Di salah satu tab dataset, klik **Request Export** → pilih format **CSV** atau **XLSX** → klik **Request Export**.
5. Pindah ke tab **Export Queue** — export status langsung **COMPLETED** (sync processing). Klik tombol **Unduh** untuk download file.
6. File metadata (nama, ukuran, waktu generate) tampil di tabel Export Queue.
4. Tunjukkan export request (generate laporan, cek status PENDING → COMPLETED).

### Blok E — GIS Map (2 menit)
1. Buka `/gis` → peta interaktif Leaflet dengan marker 10 proyek.
2. Tunjukkan filter provinsi/status/kategori.
3. Klik marker → popup detail proyek (nama, progress, budget, status).
4. Tunjukkan bahwa sidebar tetap hadir (bukan full-screen map).

### Blok F — Integrations (5 menit)
1. **Government Connectors** `/integrations/government`:
   - Tab Connectors: SIRUP (projects), OM-SPAN (budget_allocations) — _demo/sandbox mode_.
   - Tab Mappings: entitas MATCHED/PENDING_MATCH/REJECTED.
   - Tab Resolusi Entitas: workflow resolusi PENDING → MATCHED.
2. **BIM / Digital Twin** `/integrations/bim`:
   - Model metadata referensi (file tidak disimpan, hanya link viewer).
3. **Primavera P6** `/integrations/primavera`:
   - Daftar riwayat sync run (19+ entries).
   - Demo upload file XER/PMXML (opsional).

### Blok G — Data Governance (3 menit)
1. Buka `/governance`
2. Tunjukkan daftar submissions dengan berbagai status (DRAFT, SUBMITTED, APPROVED, LOCKED).
3. Buat submission baru: dataset_type, source_type, items dengan action.
4. Tunjukkan lifecycle: DRAFT → SUBMITTED → IN_REVIEW → APPROVED → LOCKED.
5. Lock Period Panel: cegah duplikat submission di periode yang sama.

### Blok H — Decision Support & Benefits (2 menit)
1. **Decision Support** `/decision-support`: ranking proyek berdasarkan weighted priority score.
2. **Benefits** `/benefits`: benefit indicators + measurements + agregasi.

### Blok I — Validation & Imports (2 menit)
1. **Validation Queue** `/validation`: antrian data yang perlu review.
2. **Import Data** `/imports`: upload batch data proyek.

### Blok J — Audit Logs (2 menit)
1. Buka `/audit-logs` di sidebar (icon Activity).
2. Tunjukkan summary cards: Total Events, Unique Actors, Top Action, Top Entity.
3. Filter: action=`login`, entity_type=`project`, date range.
4. Klik **Detail** pada salah satu row → expandable JSON (old_values / new_values / IP / UA).
5. Klik **Export CSV** → file `audit-logs-YYYYMMDD.csv` terunduh.
6. Tunjukkan bahwa `/audit-logs` redirect ke login jika belum auth (403 guard aktif).

### Blok K — Notifikasi (2 menit)
1. Buka `/notifications` di sidebar (icon Bell).
2. Tunjukkan summary cards: Total, Unread, Pending, Failed.
3. Klik **Buat Test** (tombol di kanan atas) → notifikasi baru muncul di list.
4. Klik notifikasi → detail drawer (subject/body/channel/status/timestamps).
5. Klik **Tandai Dibaca** → status berubah ke READ, unread count berkurang.
6. Tunjukkan filter status=PENDING / channel=IN_APP / Hanya Belum Dibaca.
7. Klik **Tandai Semua Dibaca** → semua notifikasi jadi READ.

---

## 6. Fitur yang Sudah Diimplementasi (Functional)

### Core P0–P1 (Full CRUD + FSM)
- ✅ Login / logout / session management
- ✅ Dashboard KPI + early warnings
- ✅ Project management (CRUD, FSM status, progress history)
- ✅ Task management (CRUD, FSM, Kanban board)
- ✅ Milestone management (CRUD, FSM)
- ✅ Issue management (CRUD, FSM, severity, escalation)
- ✅ Risk management (CRUD, FSM, probability × impact → score/severity)
- ✅ Budget monitoring (CRUD, variance/usage computed, status thresholds)
- ✅ Vendor & contract management (CRUD, unique constraint, in-use guard)
- ✅ Document management (upload/download/metadata, local storage)
- ✅ Corrective action tracker
- ✅ User management + RBAC (7 roles, permission matrix)

### Analytics P2 (Read-model / Reporting)
- ✅ PMO Command Center (aggregasi watchlist, escalation, decisions)
- ✅ Program Dashboard (KPI per program/sektor)
- ✅ Executive Dashboard (Level 1 national summary)
- ✅ GIS Map (koordinat proyek, Leaflet interactive)
- ✅ Reporting Center (tabs: executive, performance, risk, budget, benefits, priority, exports)
- ✅ Report Export Real File — CSV & XLSX (UAT-002): export request → file nyata → download authenticated endpoint; 6 dataset; file metadata (file_name, file_size, mime_type, generated_at); audit events report.export.*
- ✅ Audit Log Viewer (UAT-003): `/audit-logs` — summary cards, paginated table, expandable JSON detail, filter (action/entity/search/date), CSV export; read-only, tenant-scoped, permission-guarded (`ResourceAuditLogs:view`).
- ✅ Notification Delivery Foundation (UAT-004): `/notifications` — notification outbox DB (IN_APP + EMAIL/Noop), summary cards (total/unread/pending/failed), filter (status/channel/priority/unread_only), paginated list, detail drawer, mark-read/mark-all-read, retry FAILED, create test notification, CSV export; tenant-scoped, permission-guarded (`ResourceNotification:view`). SMTP delivered when env configured; NoopProvider logs safely when SMTP env empty.
- ✅ Benefits & Outcome Indicators (measurements, aggregasi unit-kompatibel)
- ✅ Decision Support (priority scoring dengan formula configurable, ranking, explainability)

### Integrations P2–P3
- ✅ Data Import (batch upload)
- ✅ Government Connectors (SIRUP, OM-SPAN — sandbox/mock, PENDING→MATCHED resolution)
- ✅ BIM / Digital Twin (model registry, versioning, project linking)
- ✅ Primavera P6 (XER/PMXML import, activity mappings)
- ✅ Data Governance (official validation workflow, lock periods, DRAFT→LOCKED lifecycle)
- ✅ Field Inspection Mobile View (UAT-005): `/projects/:id/inspections` — mobile-first panel, FilePicker + GeoInput, evidence upload per-inspection, download per evidence, verify/reject/delete actions, dedicated inspection route.
- ✅ Role-Based UAT & Permission Hardening (UAT-006): demo users seeded per role (pmo/pm/officer/executive/auditor), spatial routes `/sectors`/`/regions`/`/river-basins` have `RequirePermission` guards, frontend Sidebar role-filtered via `hasRole`, 65/65 smoke PASS.
- ✅ End-to-End Business Process UAT & Data Consistency Gate (UAT-007): `docs/UAT-BUSINESS-FLOW-SCENARIOS.md` (11 flows A-K), `docs/smoke-test-uat-007-business-flow.sh` **86/86 PASS** — full FSM transitions (project/CA/governance), budget thresholds, evidence upload/download, escalation+decision, governance duplicate period guard, cleanup cascade. Backend fix: `ResourceCorrectiveAction` added to seed resources for PMO/ADMIN.

---

## 7. Fitur yang Sengaja Belum Dikerjakan / Future

| Fitur | Status | Keterangan |
|---|---|---|
| **IoT Telemetry** (`CANKORA-P3-004`) | Future/Last | Sengaja diparkirkan. Jangan dimulai sebelum ada explicit sprint decision. |
| Health Score formula aktif (P1-014) | Placeholder | Formula bobot/dimensi memerlukan approval PMO sebelum diaktifkan. |
| Snapshot/validasi kontrak→budget (P1-011) | Partial | Contract value belum otomatis mengubah `project.budget_total`. |
| Gantt timeline interactive (P1-009) | Partial | Timeline read-only sudah ada; drag editing deferred. |
| MinIO / S3 document storage | Local only | Storage lokal `backend/storage/documents`; kode swap-ready. |
| Power BI integration | Future | CANKORA sebagai read-model; koneksi BI belum dikonfigurasi. |
| Field inspection & evidence mobile | Future | P2-003 partial; mobile-first view deferred. |
| GIS koordinat proyek yang tervalidasi (P2-008) | Seed demo | Koordinat demo tersedia; validasi resmi dari sistem referensi menunggu. |
| Notifikasi email/push | Future | Interface ada (`notification/`); hanya NoopProvider aktif. |
| Audit log export | Partial | DB rows ada; UI export belum dibangun. |

---

## 8. Known Limitations

1. **Playwright tidak terinstall di project** — `smoke-test-uat-001-ui-full.sh` akan fallback ke mode curl-only. Install Playwright untuk full layout/overflow checks:
   ```bash
   cd frontend && npm install --save-dev playwright && npx playwright install chromium
   ```

2. **BIM viewer URL** — BIM model menunjuk ke external viewer URL; viewer tidak di-embed di dalam CANKORA (hanya metadata + link eksternal).

3. **Government connectors = sandbox** — SIRUP/OM-SPAN adalah mock connectors; data yang disinkronkan adalah data demo, bukan data pemerintah asli.

4. **Peta monitoring = ilustratif** — Image peta di dashboard (heatmap hotspot) adalah ilustratif. Peta GIS interaktif di `/gis` menggunakan Leaflet dengan koordinat proyek demo yang nyata.

5. **Health Score = placeholder** — Dimensi health score (schedule, physical, financial, contract, risk, issue, quality, procurement) ada di arsitektur tetapi formula bobot belum dikonfigurasi. Dashboard menampilkan `—` atau nilai placeholder untuk health class sampai P1-014 aktif.

6. **Redis/MinIO tidak diperlukan** — Phase 2 features (caching, object storage) belum diaktifkan; tidak perlu dijalankan untuk UAT.

7. **Data governance submissions = seeded** — 107+ submissions di DB berasal dari smoke test sebelumnya, bukan data produksi. Ini normal untuk demo.

---

## 9. Cara Menjalankan Smoke UAT

### Backend + API regression gate
```bash
cd '/Users/harmanto/Documents/Code/DEV/Project Management/CANKORA'

# Standard (frontend optional, SKIP jika tidak jalan)
bash docs/smoke-test-rel-001-regression.sh

# Strict UAT mode (frontend WAJIB running, FAIL jika tidak ada)
STRICT_FRONTEND=1 bash docs/smoke-test-rel-001-regression.sh
```

### Full UI regression gate
```bash
# Frontend harus running, atau script akan auto-start
bash docs/smoke-test-uat-001-ui-full.sh

# Dengan Playwright (full layout/overflow/map checks):
# Install dulu: cd frontend && npm install --save-dev playwright && npx playwright install chromium
# Lalu: bash docs/smoke-test-uat-001-ui-full.sh

# Tanpa auto-start (fail langsung jika frontend tidak running)
NO_AUTO_START=1 bash docs/smoke-test-uat-001-ui-full.sh
```

### Module-specific smokes
```bash
bash docs/smoke-test-p3-003-harden.sh        # governance hardening (19/19)
bash docs/smoke-test-harden-001-integrations.sh  # BIM/gov integration (idempotent)
bash docs/smoke-test-ui-regression-dashboard-gis.sh  # dashboard + GIS layout
bash docs/smoke-test-uat-002-report-export.sh    # UAT-002 report export real file (44/45)
```

---

## 10. Recovery Steps

### `.next` corrupt / CSS hilang / full-screen map
Gejala: sidebar/topbar hilang, halaman tidak ada style, atau map monitoring full-screen.

```bash
# Stop semua proses Next.js
pkill -f "next dev" || true

# Hapus cache
rm -rf frontend/.next

# Restart bersih
cd frontend && npm run dev

# Verifikasi CSS:
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/_next/static/css/app/layout.css
# Harus: 200
```

### Database migration tidak sinkron
```bash
cd backend && make migrate-up
```

### Seed data hilang / terhapus
```bash
cd backend && make seed && make seed-demo
# Keduanya idempotent — aman diulang
```

### Backend tidak start (port sudah terpakai)
```bash
lsof -ti:8080 | xargs kill -9 || true
cd backend && make dev
```

### Frontend tidak start (port sudah terpakai)
```bash
lsof -ti:3000 | xargs kill -9 || true
cd frontend && npm run dev
```

### Login gagal setelah seed ulang
Seed menggunakan `FirstOrCreate` + update password → password selalu di-reset ke `Admin@Cankora2024!`. Jika masih gagal:
```bash
cd backend && make seed  # ulang seed untuk reset password
```

---

## 11. Environment Variables Backend (.env)

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=cankora
DB_PASSWORD=cankora_secret
DB_NAME=cankora_db
DB_SSLMODE=disable

JWT_SECRET=supersecretkey-min32chars-change-in-production

APP_ENV=development
PORT=8080

STORAGE_LOCAL_PATH=./storage/documents
STORAGE_MAX_SIZE_BYTES=20971520
```

---

## 12. Checklist Sebelum UAT

- [ ] Backend running di `:8080` — `curl http://localhost:8080/health` → 200
- [ ] Frontend running di `:3000` — browser buka `http://localhost:3000`
- [ ] Login berhasil dengan `admin@cankora.local` / `Admin@Cankora2024!`
- [ ] Dashboard tampil dengan data (tidak semua 0)
- [ ] `/gis` → peta tampil dengan marker, sidebar dan topbar hadir
- [ ] `/dashboard#monitoring` → map tidak full-screen, sidebar tetap ada
- [ ] `bash docs/smoke-test-rel-001-regression.sh` → semua PASS
- [ ] `bash docs/smoke-test-uat-001-ui-full.sh` → semua PASS atau curl-mode PASS
- [ ] `bash docs/smoke-test-uat-002-report-export.sh` → 44/45 PASS (1 SKIP: cross-tenant test)

---

## 13. Kontak & Support

- Repository: `/Users/harmanto/Documents/Code/DEV/Project Management/CANKORA`
- CLAUDE.md: konteks lengkap untuk AI assistant
- docs/DEVELOPMENT-BACKLOG.md: ticket dan status implementasi
- docs/ARCHITECTURE.md: arsitektur sistem
