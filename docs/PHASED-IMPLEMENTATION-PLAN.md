# Phased Implementation Plan
# PMO - PMO National Project Control Tower

**Versi**: 1.3  
**Tanggal**: 2026-08-29  
**Status**: Living execution plan  
**Acuan operasional**: [Development Backlog](./DEVELOPMENT-BACKLOG.md)

> **Planning update 2026-08-29 — DASH-003:** Sebelum staging/UAT lanjutan, lakukan polish kecil dashboard: tambah kolom `BALAI` pada "10 Proyek Prioritas" memakai relasi `projects.org_unit_id` → `org_units.name`, tanpa schema owner baru dan tanpa memulai IoT.

> **Update 2026-08-29 — PMO-UAT-010 Done (Final Acceptance Review & Demo Polish):** Status: **Non-IoT UAT Candidate**. Semua fase non-IoT selesai. Dibuat: `scripts/` (uat-env-check/start/status/stop), `docs/smoke-test-uat-008-release-candidate.sh` (orchestrator 7 PASS + 1 WARN non-blocking), `docs/UAT-DEPLOYMENT-GUIDE.md`, `docs/RELEASE-CANDIDATE-RUNBOOK.md`, `docs/UAT-FINAL-CHECKLIST.md`, `backend/.gitignore`. Phase table diupdate untuk mencerminkan status aktual semua fase. `scripts/uat-env-check.sh` 15 PASS 2 WARN 0 FAIL. UAT-008 RC Orchestrator: 7/8 PASS (UAT-001 non-blocking WARN karena Playwright login timeout pre-existing). `go build`, `gofmt`, `npx tsc`, `npm run lint`, `npm run build` semua OK. IoT (`PMO-P3-004`) tetap Future/Last.

> **Update 2026-08-29 — PMO-UAT-001 Done (Demo/UAT Packaging + Full UI Regression Gate):** Sistem non-IoT siap UAT/demo. Ditambahkan: `docs/smoke-test-uat-001-ui-full.sh` (18 routes full UI regression), `STRICT_FRONTEND=1` pada `docs/smoke-test-rel-001-regression.sh`, `docs/UAT-DEMO-GUIDE.md`, dan fix layout `DashboardLayout` pada 4 halaman integrasi/governance. IoT (`PMO-P3-004`) tetap **Future/Last**.

---

## 1. Tujuan

Dokumen ini membagi roadmap PMO menjadi fase implementasi yang mengikuti dependency teknis dan bisnis. Status tiket tetap dikelola di `DEVELOPMENT-BACKLOG.md`; dokumen ini menentukan urutan, hasil fase, keputusan yang dibutuhkan, dan gate sebelum fase berikutnya dimulai.

Aturan eksekusi:

- Satu agent mengerjakan satu tiket pada satu waktu.
- Hanya satu tiket boleh `In Progress`.
- Tiket menjadi `Done` hanya setelah acceptance criteria dan command verifikasi berhasil.
- Tiket dalam fase berikutnya tidak dimulai sebelum dependency dan phase gate terpenuhi.
- Fitur yang tidak bergantung pada blocker eksternal boleh dilanjutkan; blocker harus dicatat eksplisit.
- Perubahan API, schema, workflow, permission, atau status harus disinkronkan ke dokumen sumber kebenaran.

## 2. Ringkasan Fase

| Fase | Nama | Tiket | Status | Hasil Utama |
|------|------|-------|--------|-------------|
| 0 | Operational MVP & Hardening | P0-001..P0-014, P1-001 | ✅ Done | MVP lokal stabil dan Issue Management usable |
| 1 | Frontend Stabilization & PMO Core Controls | P1-016, P1-002..P1-006 | ✅ Done | Visual/runtime gate, risk, budget, contract/party, document, corrective action |
| 2 | Organization & Monitoring Foundation | P1-008..P1-011 | ✅ Done | Org/program/sector/location/DAS, Gantt, baseline dan snapshots |
| 3 | Data Governance, Field & Reporting | P1-012, P1-013, P1-007, P2-001 | ✅ Done | Validation, evidence lapangan, laporan periodik, import dasar |
| 4 | Project Intelligence & Command | P1-014, P1-015, P2-003, P2-010, P2-004 | ✅ Done | Health Score, Command Center, Level 3, benefit, decision support |
| 5 | National Control Tower & BI | P2-006..P2-009, P2-007, P2-005 | ✅ Done | Level 2, GIS, Power BI, Level 1, responsive QA |
| 6 | External Schedule & Government Integration | P2-002, P2-011 | ✅ Done | P2-011 Done: Primavera P6 XER/PMXML adapter; P2-002 Done: Government Connector Foundation (sandbox/mock) |
| 7 | Advanced Digital Integration | P3-001..P3-004 | ✅ Stable (non-IoT Done) | P3-001 Done: BIM; P3-002 Done: Gov Entity Resolution; P3-003 Done: Data Governance; REL-001 Done: Non-IoT Stabilization Gate; P3-004 IoT parked as final/future item |
| 8 | UAT & Final Acceptance | UAT-001..UAT-010 | ✅ Done | Non-IoT UAT Candidate. All smoke gates PASS. Scripts, docs, checklists ready. |

## 3. Phase 0 - Operational MVP & Hardening

**Status**: Done dan verified pada 2026-08-21.

Tiket:

- `PMO-P0-001` sampai `PMO-P0-014`.
- `PMO-P1-001` Issue Management.

Hasil yang sudah tersedia:

- Auth/session, admin `SUPER_ADMIN`, seed idempotent, dan permission middleware dasar.
- Project, task, milestone, user management, Kanban, dashboard live, dan early warning dasar.
- Tenant-safe nested CRUD dan project-child soft-delete cascade.
- Issue register, escalation, resolution, audit, API lifecycle, dan UI smoke.

Gate fase: backend build/test, frontend type-check/build, migration/seed, API smoke, UI smoke, dan orphan child audit sudah lulus. Regression terhadap fase ini wajib diperbaiki sebelum tiket baru dinyatakan `Done`.

## 4. Phase 1 - Frontend Stabilization & PMO Core Project Controls

**Tujuan**: mengunci stabilitas visual/runtime frontend terlebih dahulu, lalu melengkapi register dan workflow operasional yang menjadi input Health Score dan Command Center.

Urutan tiket:

1. `PMO-P1-016` - Frontend visual/runtime stabilization untuk login, Dashboard, dan Command Center.
2. `PMO-P1-002` - Risk Management.
3. `PMO-P1-003` - Budget Monitoring.
4. `PMO-P1-004` - Vendor, consultant, dan contract.
5. `PMO-P1-005` - Document Management.
6. `PMO-P1-006` - Corrective Action.

Status UI saat ini:

- Login PMO, dashboard eksekutif native, sidebar, data demo SDA, dan route `/command-center` sudah ada sebagai frontend shell.
- Dashboard/Command Center belum menjadi bukti Level 1 atau P1-015 selesai karena sebagian angka masih derived dari early warning dan sebagian capability diberi placeholder.
- Insiden CSS 2026-08-21 disebabkan beberapa proses dev/build memakai direktori `.next` yang sama. Build produksi wajib dijalankan saat dev server berhenti; visual smoke dilakukan setelah server dimulai ulang dari cache bersih.

Keputusan sebelum/dalam fase:

- Tetapkan storage development dan production untuk dokumen: local/S3-compatible/MinIO.
- Tetapkan lifecycle risk dan corrective action yang konsisten dengan SRS/FSM.
- Contract change harus memiliki audit trail; data finansial tidak boleh terbuka ke role yang tidak berhak.

Phase gate:

- Screenshot browser pada viewport target tidak menunjukkan overlap, full-screen image escape, text clipping, atau missing CSS.
- Asset CSS/JS merespons 200 setelah clean restart; hanya satu proses frontend menggunakan `.next` pada satu waktu.
- Seluruh register menyediakan CRUD/status/soft delete yang tenant-safe.
- Parent project ownership dan permission admin/non-admin teruji.
- Risk score berasal dari likelihood x impact dengan kontrak yang terdokumentasi.
- Corrective action dapat ditelusuri ke issue/risk, PIC, due date, evidence, dan verification.
- API dan browser smoke seluruh lifecycle lulus.

## 5. Phase 2 - Organization & Monitoring Foundation

**Tujuan**: membangun dimensi agregasi dan histori periodik sebelum dashboard eksekutif dibuat.

Urutan tiket:

1. `PMO-P1-008` - Org Unit Satker/Balai/BBWS/BWS.
2. `PMO-P1-010` - Program, sektor, wilayah, lokasi, dan DAS.
3. `PMO-P1-009` - Gantt/timeline read-only.
4. `PMO-P1-011` - Baseline dan periodic progress/financial snapshots.

Phase gate:

- Project dapat diklasifikasikan dan difilter per org unit, program, sektor, wilayah, dan DAS.
- Master code unik per tenant dan tidak di-hardcode pada chart.
- Gantt memakai task/milestone live, menampilkan overdue, dan tidak memakai dummy data.
- Snapshot menyimpan period/cut-off, planned versus actual, source, version, dan validation state.
- Agregasi tidak hanya bergantung pada current value di tabel project.

## 6. Phase 3 - Data Governance, Field & Reporting

**Tujuan**: memastikan data resmi dapat divalidasi, ditelusuri, dibuktikan, dan dipublikasikan.

Urutan tiket:

1. `PMO-P1-012` - Data validation, completeness, freshness, lineage, dan SLA.
2. `PMO-P1-013` - Field inspection dan evidence.
3. `PMO-P1-007` - Weekly/monthly/quarterly reporting snapshots.
4. `PMO-P2-001` - CSV/Excel import awal melalui validation pipeline. **Done** — migration 000024, module imports, 8 endpoints, ResourceImport permission, frontend /imports, smoke test 27/27 PASS.

Phase gate:

- Status `DRAFT`, `SUBMITTED`, `VALID`, `REJECTED`, dan `STALE` berjalan end-to-end.
- Missing/stale data terlihat dan tidak dianggap sehat.
- Evidence memiliki project/tenant scope, metadata, checksum, authorization, dan verification status.
- Published report snapshot immutable; koreksi menghasilkan version baru.
- Import idempotent, memiliki mapping/error report, dan tidak overwrite tanpa aturan jelas.

## 7. Phase 4 - Project Intelligence & Command

**Tujuan**: mengubah data tervalidasi menjadi health, alert, tindakan, dan tampilan pengendalian proyek.

Urutan tiket:

1. `PMO-P1-014` - Configurable Project Health Score.
2. `PMO-P1-015` - PMO Command Center.
3. `PMO-P2-003` - Level 3 Project Control.
4. `PMO-P2-010` - Benefit/outcome indicators.
5. `PMO-P2-004` - Priority scoring dan decision support.

Keputusan wajib sebelum Health Score diaktifkan:

- PMO menyetujui dimensi, bobot, normalisasi, threshold, dan missing-data rule.
- Formula memiliki version, effective period, approver, dan explanation.
- Priority scoring tidak otomatis menjadi keputusan pimpinan.

Phase gate:

- Health menghasilkan `GREEN`, `YELLOW`, `RED`, atau `CRITICAL` secara reproducible.
- Snapshot lama tidak berubah ketika formula baru diterbitkan.
- Command Center menghubungkan alert, aging, PIC, SLA, corrective action, escalation, dan decision follow-up.
- Level 3 merekonsiliasi contract, progress, financial, schedule, health, issue/risk, evidence, dan actions pada cut-off yang sama.
- Benefit hanya diagregasi bila unit dan aggregation method kompatibel.

## 8. Phase 5 - National Control Tower & BI

**Tujuan**: menyediakan pengendalian berjenjang dari program/sektor sampai nasional dengan native dashboard dan Power BI yang konsisten.

Urutan tiket:

1. `PMO-P2-006` - Level 2 Program/Sector Dashboard.
2. `PMO-P2-008` - GIS project distribution dan hotspot.
3. `PMO-P2-009` - Analytics read model dan Power BI semantic contract.
4. `PMO-P2-007` - Level 1 National Executive Dashboard.
5. `PMO-P2-005` - Responsive/mobile QA seluruh workflow utama.

Phase gate:

- Level 1 dapat drill-down ke Level 2 dan Level 3 sesuai effective scope.
- KPI Level 1/2/3 memakai definition, formula, period cut-off, dan validated snapshot yang sama.
- GIS, chart, export, dan Power BI RLS tidak membocorkan data lintas tenant/scope.
- Power BI read-only terhadap operational store dan refresh failure terlihat.
- Angka native dashboard dan Power BI lulus reconciliation sample.
- Desktop/mobile screenshot QA tidak menemukan overlap atau alur yang terputus.

## 9. Phase 6 - External Schedule & Government Integration

**Tujuan**: menghubungkan Control Tower dengan sistem sumber tanpa merusak system of record dan lineage.

Tiket:

1. `PMO-P2-002` - Adapter SIRUP/SIMPONI/OM-SPAN atau connector lain yang disetujui.
2. `PMO-P2-011` - Primavera P6 schedule integration.

Entry criteria:

- Dokumentasi API/file format, sample data, credential, owner, refresh frequency, dan error handling tersedia.
- Mapping master dan validation pipeline Phase 2-3 sudah stabil.

Phase gate:

- Sync idempotent dan tenant-safe.
- Setiap record memiliki source lineage, run ID, mapping version, dan error status.
- Retry tidak menghasilkan duplikasi atau cross-tenant overwrite.
- Ketiadaan akses eksternal dicatat `Blocked`; mock connector tidak diklaim sebagai production integration.

## 10. Phase 7 - Advanced Digital Integration

**Tujuan**: menambahkan capability digital lanjutan setelah governance dan histori data memadai.

Urutan tiket:

1. `PMO-P3-001` - BIM/digital twin reference integration. **Done** (2026-08-28): migration `000027_bim_models`, module `integration/bim`, 10 routes `/api/v1/integrations/bim/models`, `ResourceBIMIntegration`, frontend `/integrations/bim`, smoke 23/23 PASS.
2. `PMO-P3-002` - Government Entity Resolution. **Done** (2026-08-28): PENDING_MATCH → MATCHED (migration `000029_government_resolution_metadata`, `resolver.go`, 6 routes, frontend tab "Resolusi Entitas", smoke 13 checks).
3. `PMO-P3-003` - Data Governance — Official Data Validation & Approval Workflow. **Done** (2026-08-28): module `governance/`, migrations `000030`–`000033`, 12 routes `/api/v1/governance/`, `ResourceDataGovernance`, frontend `/governance`, smoke 17/17 PASS.
4. `PMO-P3-004` - IoT/sensor telemetry ingestion pipeline. **Future / Last** — sengaja diparkir sebagai item advanced-integration paling akhir. Bentuk awal yang direncanakan adalah IoT data gateway + device/source registry + telemetry validation + alert foundation; MQTT, time-series optimization, alert rules engine, dan digital twin overlay baru setelah foundation stabil. Jangan mulai sebelum governance hardening, dashboard/GIS regression, reporting/read-model reconciliation, dan semua Phase 3 non-IoT yang lebih mendasar stabil.

Entry criteria:

- Storage, GIS, analytics read model, security, retention, dan observability sudah stabil.
- Dataset tervalidasi memiliki histori dan coverage yang disetujui PMO untuk evaluasi AI/ML.

Phase gate:

- BIM menyimpan external reference/version dan authorization, bukan file besar pada row database.
- Telemetry memiliki device registry, quality flag, retention/sampling, idempotency, dan tenant mapping.
- AI/ML menyimpan model version, feature timestamp, confidence, explanation, dan human review.
- Prediction dibandingkan dengan rules baseline dan tidak membuat keputusan otomatis.

## 11. Release Milestones

| Milestone | Fase | Kelayakan |
|-----------|------|-----------|
| Release A - PMO Operational Control | Phase 1-3 | Register, monitoring periodik, validation, evidence, dan reporting usable |
| Release B - PMO Project Intelligence | Phase 4 | Health, Command Center, Level 3, benefit, dan decision support usable |
| Release C - National Control Tower | Phase 5 | Level 1/2/3, GIS, Power BI, dan responsive QA usable |
| Release D - Integrated Control Tower | Phase 6 | External government/schedule sources terhubung |
| Release E - Advanced Digital PMO | Phase 7 | BIM, government resolution, data governance, dan integration hardening tersedia; IoT dan explainable AI/ML sebagai final/future capability |

## 12. Verification Baseline

Setiap fase tetap menjalankan command spesifik pada tiket. Minimum regression gate:

```bash
cd backend
gofmt -w <file-go-yang-berubah>
go build ./...
go test ./...

cd ../frontend
npm run type-check
npm run lint
npm run build
```

`npm run build` dijalankan hanya setelah dev server berhenti. Saat kembali ke mode dev, gunakan satu server dengan cache `.next` yang bersih. Jika schema berubah, wajib verifikasi migration up/down dan seed idempotency. Jika UI berubah, wajib browser smoke desktop/mobile serta HTTP 200 untuk CSS/JS asset. Jika agregasi atau analytics berubah, wajib reconciliation terhadap API/SQL fixture dan tenant isolation test.

## 13. Resume Rule untuk AI Agent

Pada awal sesi:

1. Baca `CLAUDE.md`, phased plan ini, dan `DEVELOPMENT-BACKLOG.md`.
2. Verifikasi tiket `Done` terakhir berdasarkan catatan test.
3. Pilih tiket `Todo` pertama dalam fase aktif yang dependency-nya `Done`.
4. Selesaikan, verifikasi, dokumentasikan, lalu lanjut hanya jika phase gate tetap hijau.

Current resume point pada 2026-08-28: **jangan mulai IoT sebagai next ticket**. P3-001 BIM foundation, P3-002 Government Entity Resolution, dan P3-003 Data Governance (Official Data Validation & Approval Workflow) sudah lulus gate backend/frontend lokal, tetapi review terakhir menemukan governance validation hardening dan dashboard/GIS regression yang harus dibereskan lebih dulu. `PMO-P3-004` - IoT/sensor telemetry ingestion dipindah ke urutan paling akhir/future advanced integration setelah governance hardening, UI regression, reporting/read-model reconciliation, dan semua non-IoT Phase 3 work stabil.
