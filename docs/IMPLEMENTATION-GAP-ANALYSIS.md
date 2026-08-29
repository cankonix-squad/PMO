# Implementation Gap Analysis
# CANKORA - PMO National Project Control Tower Alignment

**Versi**: 0.2.9  
**Tanggal Audit**: 2026-08-29  
**Sumber Kebutuhan**: Grand Design PMO SDA, Deck PMO SDA, dan Grand Design PMO National Project Control Tower  
**Status**: Living Document

> **Update 2026-08-29 — CANKORA-DASH-003 Planned (BALAI Column):**
> Gap UI ditemukan pada Dashboard "10 Proyek Prioritas": user membutuhkan kolom `BALAI` di sebelah `KODE`. Analisa: sumber data sudah ada (`projects.org_unit_id`), tetapi response/UI belum meng-resolve nama org unit. Implementasi berikutnya harus enrich project response dengan `org_unit_name`/`org_unit`, menampilkan kolom `BALAI`, dan memperjelas label form project menjadi `Balai / Unit Pemilik`. Jangan tambah field owner baru sebelum ada kebutuhan bisnis yang memisahkan owner dari org hierarchy.

> **Update 2026-08-29 — CANKORA-UAT-010 (Final Acceptance Review & Demo Polish):**
> Status: **Non-IoT UAT Candidate**. Deliverables: (1) `scripts/` directory dibuat — `uat-env-check.sh` (15 PASS 2 WARN 0 FAIL), `uat-start.sh`, `uat-status.sh`, `uat-stop.sh`; (2) `docs/smoke-test-uat-008-release-candidate.sh` — RC orchestrator 7 PASS + 1 WARN (UAT-001 non-blocking) + 0 FAIL → RELEASE CANDIDATE: READY; (3) `docs/UAT-DEPLOYMENT-GUIDE.md`, `docs/RELEASE-CANDIDATE-RUNBOOK.md`, `docs/UAT-FINAL-CHECKLIST.md` dibuat; (4) `backend/.gitignore` dibuat; (5) Phase table `PHASED-IMPLEMENTATION-PLAN.md` diupdate — semua fase non-IoT status ✅ Done, Phase 8 UAT baru ditambahkan; (6) `PERMISSION-MATRIX.md` diupdate — `corrective_action` resource ditambahkan; (7) `IMPLEMENTATION-GAP-ANALYSIS.md` versi diupdate ke 0.2.8. Gap yang ditutup: scripts UAT tidak ada → ada; docs deployment/runbook/checklist tidak ada → ada; phase table stale → akurat. Known limitations: IoT Future/Last, gov connector sandbox, BIM metadata-only, storage lokal, SMTP NoopProvider jika unconfigured, UAT-001 Playwright login timeout pre-existing (non-blocking). `gofmt` clean · `go build ./...` OK · `go test ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK.

> **Update 2026-08-29 — CANKORA-UAT-007 (End-to-End Business Flow Smoke):**
> End-to-end business process UAT selesai. Gap yang ditutup: (1) `docs/UAT-BUSINESS-FLOW-SCENARIOS.md` (11 flows A-K); (2) `docs/smoke-test-uat-007-business-flow.sh` **86/86 PASS** — full FSM transitions, budget thresholds, evidence upload/download, escalation+decision, governance duplicate period guard, cleanup cascade; (3) `cmd/seed/main.go` — `ResourceCorrectiveAction` ditambahkan ke resources slice (PMO/ADMIN sekarang punya corrective_action permission). API contracts yang diverifikasi: `TransitionRequest.to_status` (bukan `status`); CA FSM DRAFT→SUBMITTED→IN_PROGRESS; inspection multipart form + `inspected_at`; governance `dataset_type` UPPERCASE; decision status `IN_PROGRESS`/`COMPLETED`/`CANCELLED`; evidence ID di `data.evidence[0].id`.

> **Update 2026-08-29 — CANKORA-UAT-006 (Role-Based UAT & Permission Scope Hardening):**
> RBAC hardening selesai untuk UAT. Gap yang ditutup: (1) demo users seeded per role (pmo/pm/officer/executive/auditor, semua `Demo@Cankora2024!`); (2) spatial routes `/sectors`/`/regions`/`/river-basins` ditambahkan `RequirePermission` guard; (3) `ResourceSector`, `ResourceRegion`, `ResourceRiverBasin` ditambahkan ke seed resources; (4) `Sidebar.tsx` role-filtered via `hasRole`. Smoke **65/65 PASS**.

> **Update 2026-08-29 — CANKORA-UAT-005 (Field Inspection Mobile View):**
> Field inspection/evidence polished untuk UAT/demo mobile. Gap yang ditutup: (1) `FieldInspectionPanel` mobile-first — tidak ada horizontal overflow di 390px, sub-komponen `FilePicker`/`GeoInput`/`StatusBadge`, upload state jelas, geolocation browser optional + graceful fallback, evidence list padat dengan checksum/size/geotag. (2) Route `/projects/:id/inspections` dedicated untuk mobile. (3) Endpoint `POST /:inspectionID/evidence` baru — upload evidence ke inspeksi existing, per-evidence lat/lon, audit `field_inspection.evidence_added`. (4) Regression gate diperluas: section 16 (5 checks), UAT-001 21 routes. Smoke UAT-005 **28/28 PASS** (Playwright mobile 390x844 no overflow + desktop 1366x768). Backend: `gofmt` clean · `go build ./...` OK. Frontend: `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK (30 pages).

> **Update 2026-08-29 — CANKORA-REL-001 (Non-IoT Stabilization):**
> Regression gate terpadu `docs/smoke-test-rel-001-regression.sh` (15 kelompok check) selesai. Gap yang ditutup: (1) `actionRequiresEntityID()` — UPDATE/DELETE/UPSERT wajib `entity_id`, CREATE & VALIDATE_ONLY boleh omit; (2) MATCHED fixture di-seed idempotent; (3) smoke harden check 7 tidak lagi SKIP; (4) entity_id rules 5 sub-checks. Dashboard/reporting label audit: `/executive`=Level 1 (operational aggregate), `/programs`=Level 2 (operational aggregate), `/command-center`=operational current-state, `/reports/analytics`=read-model aggregate, `/governance`=official governed data — tidak ada mixing tanpa disclaimer. Docs di-sync ke versi 2026-08-29. IoT (`CANKORA-P3-004`) tetap Future/Last. Sistem siap UAT/demo packaging.

> **Update 2026-08-28 — CANKORA-P3-003 (Data Governance):**
> Official Data Validation & Approval Workflow selesai end-to-end. Modul `governance` menyediakan submission resmi (DRAFT→SUBMITTED→IN_REVIEW→APPROVED→LOCKED, +REJECTED dengan alasan wajib, CANCELLED), per-item validation (tenant ownership, soft-delete ditolak, mapping pemerintah `PENDING_MATCH` → `INVALID`), dan lock period (OPEN/LOCKED, unik per org+dataset+periode) yang memblokir penulisan pada periode terkunci. 12 route `/api/v1/governance/`, `ResourceDataGovernance`, audit `governance.*`, frontend `/governance`, smoke `docs/smoke-test-p3-003-data-governance.sh` 17/17 PASS. **Gap yang ditutup:** "Gap yang masih terbuka" PENDING_MATCH → MATCHED (dari HARDEN-001) telah diselesaikan oleh `CANKORA-P3-002` (Government Entity Resolution, 2026-08-28). **Planning update:** pipeline IoT telemetry (`CANKORA-P3-004`) dipindah ke urutan paling akhir/future advanced integration; jangan dimulai sebelum governance hardening, dashboard/GIS regression, reporting/read-model reconciliation, dan non-IoT Phase 3 work stabil.

---

## 1. Ringkasan Eksekutif

Status CANKORA harus dinilai menggunakan dua baseline yang berbeda:

| Baseline | Status | Kesimpulan |
|----------|--------|------------|
| Local Operational MVP | P0-001 s.d. P0-014 `Done` dan verified | Auth, RBAC awal, project/task/milestone/user, dashboard live, Kanban, early warning dasar, dan integritas data saat project delete (soft delete cascade) usable di lokal |
| Full PMO National Project Control Tower | Level 3 Basic Implemented; Level 1/2/GIS/BI Planned | Native dashboard, Command Center lifecycle dasar, Level 3 project control, health score, validation, contract/financial, field evidence, reporting, decision follow-up, dan benefit indicator sudah basic usable; Level 1/2 resmi, GIS aktual, Power BI/read model, external integration, dan advanced governance tetap roadmap |

P0 selesai tidak berarti kebutuhan Control Tower selesai. Visual mockup juga tidak menjadi bukti implementasi. Angka, nama proyek, tanggal, nilai, dan persentase pada materi adalah ilustrasi.

Known hardening gap setelah smoke P0-013: task anak dapat tetap aktif ketika project sudah soft-deleted. Gap ini dicatat sebagai `CANKORA-P0-014` dan **sudah diselesaikan pada 2026-08-21** (soft delete cascade transactional, dashboard exclusion, cleanup smoke, orphan soft cleanup — semua verified).

## 2. Interpretasi Materi

- Materi SDA menetapkan kebutuhan normatif: monitoring, corrective action, reporting, Control Tower Level 1-3, health score, GIS, evidence, dan roadmap integrasi.
- Mockup UI menetapkan information architecture, bukan pixel-perfect design atau data produksi.
- Slide BP3R hanya menjadi referensi pola PMO generik; proses dan data perumahan tidak dimasukkan sebagai requirement SDA.
- Strategi yang dipilih adalah hybrid: CANKORA sebagai system of record dan Power BI sebagai dashboard eksekutif tahap awal.
- Formula/bobot Project Health Score belum disetujui dan harus configurable, versioned, explainable, serta approved sebelum aktif.

Detail interpretasi ada di [PMO Control Tower Analysis](./PMO-CONTROL-TOWER-ANALYSIS.md).

## 3. Status Implementasi Teknis Saat Ini

| Area | Status | Evidence/Gap |
|------|--------|--------------|
| Backend build/test | Verified 2026-08-20 | `gofmt`, `go build ./...`, dan `go test ./...` OK pada P0-013 |
| Migration/seed | Verified 2026-08-20 | Migrate dan seed idempotent; admin `SUPER_ADMIN`, `must_change_pwd=false` |
| Frontend build | Verified 2026-08-20 | `npm run type-check` dan `npm run build` OK |
| Auth/session | Verified | Login, refresh/me, redirect, remember me, logout, password eye toggle |
| Project | Verified/Basic | CRUD, FSM, progress history, duplicate code 409 |
| Task/WBS | Verified/Basic | Nested CRUD, subtask dasar, Kanban; comments/assignment penuh belum ada |
| Milestone | Verified/Basic | Nested CRUD dan status transition |
| User/RBAC | Verified/Basic | User CRUD/deactivate dan role assignment; scoped role management belum penuh |
| Dashboard | Verified/Basic | Data live dari current operational tables (`projects`, `tasks`, `milestones`, `users`, `project_budgets`), empty state, tenant-scoped early warning dasar; belum memakai snapshot periodik, validation state, atau analytics read model |
| Frontend visual/runtime | Verified/Basic | Login/dashboard/Command Center tersedia; insiden CSS asset mismatch dari concurrent `next dev`/`next build` sudah direcover; multi-viewport screenshot regression dan clean runtime gate selesai pada `P1-016` |
| PMO Command Center UI | Implemented/Basic | Aggregate API menyediakan alert/watchlist, validation SLA, corrective action PIC/aging, persisted escalation, executive decision follow-up, audit, dan tenant-safe source/project validation; GIS/Level 1-2 resmi masih roadmap |
| Data integrity on project delete | Verified 2026-08-21 | P0-014: `DeleteCascade` soft-deletes project + all children in one transaction; dashboard excludes child of deleted projects; SQL audit 0 orphan; smoke 33/33 PASS |
| Issue/risk/budget/contract | Mostly missing → issue/risk/budget/contract done | Issue (P1-001) usable end-to-end: create/list/update/FSM/soft delete, org-scoped, severity/escalation/due date/resolution/audit. Risk (P1-002) usable: register `risks` + organization_id + probability/impact INT 1-5 + risk_score + severity, nested CRUD/FSM/soft delete/audit, dashboard `RISK_REGISTER`. Budget (P1-003) usable: line item planned/actual/currency CRUD tenant-safe, variance/usage_pct/status computed backend, dashboard `BUDGET_THRESHOLD`. Contract (P1-004) usable: master `vendors` (VENDOR/CONSULTANT) tenant-scoped + `contracts` per project (contract_number unique per org, nilai/tanggal/enum validation, audit, tenant guard), UI panel "Kontrak" di project detail; contract value operational input — belum otomatis mengubah `project.budget_total` dan belum validated/published (menunggu P1-011/P1-012) |
| Corrective action/report/document | Implemented/Basic | Corrective action, reporting snapshot, dan document management sudah usable end-to-end dengan tenant guard, audit, soft delete, dan frontend panel/route dasar |
| Program/sector/location/DAS | Implemented/Basic | Master program/sektor/wilayah/DAS dan project classification tersedia; map spatial penuh masih menunggu P2-008 |
| Progress/financial snapshots | Implemented/Basic | Baseline, cut-off period, validation status, variance, dan immutable VALID snapshot flow tersedia |
| Health Score full | Implemented/Basic | P1-014: formula configurable/versioned, approval/retire, delapan komponen, missing-data rule, explanation, dan immutable project health snapshots tersedia |
| Control Tower Level 1-3 | Basic Level 3 implemented | P2-003 menyediakan tenant-scoped Level 3 control API/panel dengan contract, VALID snapshot, variance, health, evidence, issue/risk, dan corrective action; Level 1/2 dan governed aggregate masih roadmap |
| Validation/freshness | Implemented/Basic | P1-012: tenant-scoped validation queue, SLA aging, completeness, freshness, stale marker, lineage, SoD, audit, dan status lifecycle tersedia |
| GIS/field evidence | Implemented/Basic | P1-013: field inspection/evidence tenant-scoped dengan koordinat, checksum, verification, soft delete, dan authorized download; GIS map tetap belum ada |
| Power BI/read model | Missing | Belum ada semantic dataset, RLS, atau refresh monitoring |
| External integration | Partial | P2-011 Done: Primavera P6 adapter (XER/PMXML), idempotent sync, lineage, conflict/error report, tenant validation |
| Data Governance / Official Approval (P3-003) | Implemented | 2026-08-28: submission resmi DRAFT→SUBMITTED→IN_REVIEW→APPROVED→LOCKED (+REJECTED wajib alasan, CANCELLED), per-item validation (tenant ownership, soft-delete, `PENDING_MATCH` ditolak), lock period per org+dataset+periode, audit `governance.*`, 12 route `/api/v1/governance/`, `ResourceDataGovernance`, frontend `/governance`, unit test FSM, smoke 17/17 PASS |
| Government Entity Resolution (P3-002) | Implemented | 2026-08-28: PENDING_MATCH → MATCHED end-to-end via resolver (candidates by code/NPWP, match/unmatch/reject), migration `000029`, 6 route, frontend tab "Resolusi Entitas", smoke 13 checks |
| BIM Integration (P3-001) | Implemented | 2026-08-28: module `integration/bim` (model/service/handler), migration `000027_bim_models`, 10 route `/api/v1/integrations/bim/models`, `ResourceBIMIntegration`, frontend `/integrations/bim`, smoke 23/23 PASS |

### 3.1 Dashboard Data Source Reality Check

Analisa source code 2026-08-25 menunjukkan dashboard yang sudah dikembangkan saat ini adalah **operational visibility layer**, bukan governed Control Tower analytics.

| Area Dashboard | Sumber Aktual | Catatan Bisnis/Data |
|----------------|---------------|---------------------|
| KPI proyek, task, milestone, user | `GET /api/v1/dashboard` dari tabel `projects`, `tasks`, `milestones`, dan `users` | Menggunakan current state operasional; sudah tenant-scoped dan exclude soft-deleted project/child utama |
| Early warning | Rule backend pada overdue task, overdue milestone, low progress near end date, dan budget threshold | Rule sederhana, belum alert entity dengan PIC, SLA, lifecycle, source evidence, atau acknowledgement |
| Nilai portofolio dan progress nasional | `GET /api/v1/projects` lalu dihitung di frontend dari `budget_total` dan `progress_pct` | Belum memakai cut-off period, submitted/validated snapshot, atau official KPI definition |
| Kesehatan proyek | Distribusi status proyek di frontend | Bukan Health Score delapan dimensi; formula P1-014 belum tersedia |
| Risiko utama dan isu utama di dashboard | Derived dari `early_warnings` | Belum membaca register `risks`/`issues` sebagai executive top risk/top issue resmi |
| Command Center | Kombinasi `/dashboard`, `/projects`, dan derivasi frontend | Placeholder/derived untuk validation, action tracker, data quality, reporting schedule, decision, dan GIS |
| Peta/GIS | Static image frontend | Belum ada koordinat proyek, spatial model, cluster, hotspot, atau tenant-safe map API |

Implikasi:

- Dashboard saat ini cukup untuk monitoring MVP dan demo alur operasional dasar.
- Dashboard belum boleh dipakai sebagai bukti reporting resmi karena belum ada data period, validation queue, immutable report snapshot, formula version, atau analytics lineage.
- Temuan teknis untuk backlog: budget warning harus memastikan `project_budgets.deleted_at IS NULL` agar line item soft-deleted tidak ikut dihitung setelah P0-014 menambahkan soft delete pada tabel tersebut — sudah diterapkan di `budgetThresholdWarnings` (P1-016/P1-003), termasuk JOIN `projects.deleted_at IS NULL` dan scope org via `projects.organization_id`.

## 4. Alignment terhadap Target Control Tower

| Target Capability | Current State | Gap | Ticket Utama |
|-------------------|---------------|-----|--------------|
| Level 1 National Executive | Missing | KPI nasional, map, trend, critical ranking, decision queue | P2-007, P2-008, P2-009 |
| Level 2 Program/Sector | Missing | Master program/sektor, comparison, drill-down | P1-010, P2-006 |
| Level 3 Project Control | Partial | Contract done (P1-004); progress snapshot, variance, health explanation, evidence masih menunggu P1-011, P1-013, P1-014 | P1-004, P1-009, P1-011, P1-013, P1-014 |
| Project Health Score | Early warning dasar | Delapan dimensi, formula approval/version, snapshot/explanation | P1-014 |
| PMO Command Center | Implemented/Basic | Alert/watchlist, validation queue, aging/PIC/SLA, action tracker workflow, escalation, decision follow-up, project drill-down, permission, dan audit tersedia; GIS dan Level 1/2 resmi tetap P2 | P1-006, P1-012, P1-014, P1-015, P1-016 |
| Issue/Risk/Corrective Action | Partial | Issue management usable (P1-001 Done: register, FSM, severity, escalation, resolution, audit). Risk & corrective action masih Missing | P1-002, P1-006 |
| Contract/Financial Monitoring | Partial | Parties/contract done (P1-004: `vendors` + `contracts` CRUD tenant-safe), realization/variance via budget (P1-003); baseline, forecast, validated snapshot (P1-011) masih missing | P1-003, P1-004, P1-011 |
| Schedule Control | Partial | Baseline, planned vs actual, Gantt, Primavera | P1-009, P1-011, P2-011 |
| Field Progress/Evidence | Missing | Inspection, geotag, photo, verification | P1-005, P1-013 |
| Reporting | Missing | Weekly/monthly/quarterly snapshots, archive, export | P1-007 |
| GIS | Missing | Project point, region/DAS, cluster, hotspot | P1-010, P2-008 |
| Power BI | Missing | Read model, data dictionary, RLS, refresh status | P2-009 |
| Benefit/Outcome | Implemented/Basic | Indicator, baseline/target/actual measurement, validation status, source, owner, dan aggregation by compatible unit/method tersedia | P2-010 |
| Data Integration | Partial | P2-001 Done: CSV/Excel upload, 5 dataset types, job lifecycle, row validation, audit trail; P2-011 Done: Primavera P6 XER/PMXML adapter, idempotent sync, lineage, conflict report, tenant guard; connectors/retry (P2-002) still pending | P2-001 Done, P2-011 Done, P2-002 |
| Mobile/PWA | Missing | Responsive field workflow/offline strategy | P2-005 |
| BIM/Digital Twin, IoT, AI/ML | Partial — P3-001/P3-002/P3-003 Done | P3-001 BIM foundation Done (module `integration/bim`, 10 routes, frontend `/integrations/bim`, smoke 23/23); P3-002 Government Entity Resolution Done (PENDING_MATCH→MATCHED, resolver, 6 routes, tab UI, smoke 13); P3-003 Data Governance Done (official validation & approval workflow, 12 routes, `/governance`, smoke 17/17). P3-004 IoT telemetry diparkir sebagai final/future item: data gateway + registry + telemetry validation + alert foundation dulu; MQTT/time-series/overlay setelah foundation stabil. AI/ML menunggu histori data tervalidasi |
| Data Governance / Official Approval | Implemented | P3-003: submission FSM + per-item validation + lock period + audit `governance.*` + `ResourceDataGovernance` + frontend `/governance`; data import/sync tidak otomatis approved |

## 5. Recommended Delivery Sequence

1. **Frontend stability gate**: P1-016 — clean Next.js lifecycle, CSS asset check, bounded image, active navigation, and screenshot QA pada target viewports.
2. **Hardening baseline**: P0-014 selesai — soft delete cascade child, dashboard exclusion, smoke cleanup, dan orphan cleanup.
3. **Data foundation**: org unit, program, sektor, lokasi/DAS, contract/party, baseline, progress/financial snapshots, dan validation.
4. **PMO core process**: issue, risk, budget, document, corrective action, reporting, dan field evidence.
5. **Project intelligence**: Gantt, variance, health score, alert, dan Level 3 project control.
6. **Control Tower**: P1-015 backend/workflow, Level 2, Level 1, GIS, Power BI, benefit, dan decision support.
7. **Advanced integration**: Primavera (P2-011 Done), BIM/digital twin (P3-001 Done), government entity resolution (P3-002 Done), data governance/approval workflow (P3-003 Done), lalu governance hardening/regression/read-model work. IoT telemetry (P3-004) ditempatkan paling akhir sebagai future item; AI/ML setelah histori data tervalidasi mencukupi.

AI/ML tidak boleh dipercepat sebelum data definition, validation, lineage, dan historical coverage memenuhi kriteria yang disetujui PMO.

## 6. Minimum Acceptance untuk Setiap Tahap

### Operational Integrity

- Semua business query dan aggregate memfilter `organization_id`.
- Parent ownership diverifikasi untuk nested resources.
- Project soft delete ikut menonaktifkan child business data secara soft delete atau memastikan child tidak bisa dihitung/diakses.
- Seed/migration idempotent dan smoke test repeatable.

### Data Foundation

- KPI memiliki nama, definisi, unit, formula, period, owner, source, dan validation rule.
- Snapshot menyimpan period/cut-off dan tidak bergantung pada nilai current-state saja.
- Missing/stale data terlihat dan tidak dianggap `GREEN`.
- Data valid dapat ditelusuri ke submission dan source.

### Control Tower

- Level 1 dapat drill-down ke Level 2 dan Level 3 sesuai scope.
- Angka agregat konsisten dengan project snapshot pada cut-off yang sama.
- Health score menyimpan formula version dan component explanation.
- Power BI dan native dashboard memakai metric definition yang sama.
- GIS, export, dan RLS tidak membocorkan data lintas tenant/scope.

### Decision and Corrective Action

- Alert actionable memiliki severity, aging, PIC, due date, recommendation, evidence, dan status.
- Escalation dan keputusan tercatat serta dapat dilacak sampai tindak lanjut selesai.
- Published report/health snapshot dan audit history tidak dimutasi.

## 7. Source-of-Truth Synchronization

- [SRS](./SRS.md): kebutuhan fungsional dan non-fungsional.
- [Architecture](./ARCHITECTURE.md): keputusan teknis dan data flow.
- [ERD Conceptual](./ERD-CONCEPTUAL.md): target data model versus migration aktual.
- [Permission Matrix](./PERMISSION-MATRIX.md): target role, permission, dan scope.
- [Development Backlog](./DEVELOPMENT-BACKLOG.md): tiket operasional dan urutan eksekusi.
- [PMO Control Tower Analysis](./PMO-CONTROL-TOWER-ANALYSIS.md): interpretasi materi dan batas normatif.
- `CLAUDE.md`: context ringkas, aturan wajib, verified state, dan next ticket untuk AI agent.

Status hanya berubah menjadi `Implemented`/`Done` setelah acceptance criteria dan command verifikasi tiket terkait berhasil. Penambahan requirement pada dokumen tidak mengubah status implementasi.
