# PMO — Project Context for AI Assistants

> **IMPORTANT READING RULE:** When the user says "baca CLAUDE.md", "read CLAUDE.md", "cek CLAUDE.md", or equivalent, automatically read these core docs too before coding: `docs/DEVELOPMENT-BACKLOG.md`, `docs/PHASED-IMPLEMENTATION-PLAN.md`, `docs/SRS.md`, `docs/ARCHITECTURE.md`, `docs/ERD-CONCEPTUAL.md`, `docs/IMPLEMENTATION-GAP-ANALYSIS.md`, `docs/PERMISSION-MATRIX.md`, `docs/UAT-DEMO-GUIDE.md`, `docs/UAT-FINAL-CHECKLIST.md`, and `docs/RELEASE-CANDIDATE-RUNBOOK.md`. Start from each document's newest status/header, identify current project status, last Done ticket, next planned ticket, blockers, explicit prohibitions, and docs read. Conflict priority: `CLAUDE.md` > `docs/DEVELOPMENT-BACKLOG.md` > `docs/PHASED-IMPLEMENTATION-PLAN.md` > domain docs. Do not start coding until the latest status, blockers, known limitations, and rules such as "IoT Future/Last" are understood.

> **STATUS (2026-08-30 — PROJECT-FORM-002 applied):** `PMO-PROJECT-FORM-002` is `Done` (2026-08-30): Project Master Data Input + Rich Text Editor. Backend: migration `000039_create_project_categories` (tenant-scoped, soft-delete, unique code per org); module `internal/modules/projectcategory/` (handler/model/service): full CRUD, `ErrNotFound`/`ErrCodeTaken`, duplicate code → 409; `ResourceProjectCategory` constant seeded to all roles; 7 default SDA categories (BND/IRG/SNP/ABK/ATN/BNJ/OP) + 5 regions + 5 river_basins idempotent seed; `server.go` wires `/api/v1/project-categories` (GET/POST/GET:id/PUT:id/DELETE:id) with RBAC. Frontend: `types/project-category.ts` + `services/project-category.service.ts`; `app/(dashboard)/settings/project-categories/page.tsx` (CRUD modal, toggle active); `components/editor/RichTextEditor.tsx` (TipTap: bold/italic/bulletList/orderedList/hr, paste image, attach file); `ProjectForm` refactored to stateful — description/objectives → RichTextEditor (controlled state), category dropdown from API not hardcoded, `onSubmit` callback changed from FormEvent to `ProjectFormValues`; Sidebar: "Kategori Proyek" (PMO_AND_ABOVE); `onError` handlers show API error in form. Decision: `category` stores `code` string — no `project_category_id` FK added to `projects` (blast radius avoided; documented as known limitation). Rich content stored as HTML string; sanitization deferred. Smoke: CRUD 5/5 PASS, duplicate 409 PASS, project create with rich description PASS. `go build ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK · migration applied · seed OK. Push: `9d18153` on `harmanto-49/PMO` main.

> **STATUS (2026-08-30 — PROJECT-FORM-001 applied):** `PMO-PROJECT-FORM-001` is `Done` (2026-08-30): Add Project mandatory metadata hardening. Frontend `/projects` Add/Edit Project form now treats all visible core fields as mandatory: code, name, description, objectives, priority, category, start/end dates, budget, progress, Balai/Unit Pemilik, Program, Sektor SDA, Wilayah, and DAS. Budget input displays as Rupiah text format (`Rp. 100,000,000,-`) and parses digits back to numeric `budget_total`; progress field shows percent affordance and sends `progress_pct` on create. `Kategori` is currently a controlled UI option list because no true `project_categories` master table exists yet. Backend `CreateProjectRequest` now requires the core project metadata and accepts/persists `program_id`, `sector_id`, `region_id`, `river_basin_id`, and `progress_pct`; create/update validates org unit/program/sector/region/river basin IDs against tenant-scoped active master rows. Seed permissions now include `ResourceProgram` to fix 403 on `/api/v1/programs`; `make seed` rerun and `/api/v1/programs` verified 200 as Super Admin. Verification: `gofmt` OK · `go build ./...` OK · `go test ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK. **Follow-up recommended:** create real `project_categories` master data module/table and replace temporary UI category list with API-backed master data.

> **STATUS (2026-08-29 — DASH-003 applied):** `PMO-DASH-003` is `Done` (2026-08-29): Dashboard Priority Table BALAI Column. Column `BALAI` added to Dashboard → "10 Proyek Prioritas" using `projects.org_unit_id` → `org_units.name`. Backend: `Project` struct has `OrgUnitName string` (gorm:"-", json:"org_unit_name,omitempty"); `postgresRepository.List()` does post-query batch lookup — collects unique `org_unit_id`s, fetches names via `SELECT id, name FROM org_units WHERE id IN ? AND deleted_at IS NULL`, populates each project; `FindByID()` does single lookup if `OrgUnitID != nil`. Seed-demo: `demoOrgUnits` slice (5 BBWS/BWS: BBWS Ciliwung Cisadane, BBWS Citanduy, BWS Sumatera VII, BBWS Serayu Opak, BWS Nusa Tenggara II, level 3); `upsertOrgUnit()` idempotent; `projectSeed.OrgUnitCode` field; all 10 demo projects assigned to a BBWS/BWS. Frontend: `Project` type gets `org_unit_name: string | null`; Dashboard "10 Proyek Prioritas" table gets `BALAI` column after `KODE` with `-` fallback, `TableLoading columns={10}`; form label `Unit Organisasi` → `Balai / Unit Pemilik`. Fix: `ResourceOrgUnit` added to `cmd/seed/main.go` resources slice (was missing — caused 403 on `/settings/org-units`). Verification: `go build ./...` OK · `npx tsc --noEmit` OK · regression **44/44 PASS** · `org_unit_name` verified in API for all 10 demo projects.

> **NEXT PLANNED (2026-08-29):** No next planned ticket specified. Recommended: staging deployment → stakeholder UAT session → bug collection → production hardening. IoT (`PMO-P3-004`) tetap Future/Last.

> **STATUS (2026-08-29 — DASH-002 applied):** `PMO-DASH-002` is `Done` (2026-08-29): Periodic Progress & Financial Realization Input. Dashboard trend `/api/v1/dashboard/trend` now prioritizes data from `project_periodic_reports` table (`data_type: "PERIODIC_REPORT"`) before falling back to progress history + budget aggregate. New table `project_periodic_reports` (migration 000038): organization_id, project_id, period_year, period_month, physical_progress_pct, financial_planned, financial_actual, financial_pct (backend-computed), notes, reported_by, reported_at, soft-delete; unique active index per org+project+year+month. CRUD API `GET/POST/GET/:id/PUT/:id/DELETE/:id` under `/api/v1/projects/:id/periodic-reports` with RBAC (`ResourcePeriodicReport`), tenant guard via parent project, duplicate → 409, validation → 400, soft delete → 204. Handler uses `Handler.WithDB(db)` pattern to avoid bloating the Repository interface. `Service.SyncProjectProgressFromPeriodic()` updates `projects.progress_pct` after create/update. Seed-demo idempotent: 6 months of periodic reports per project × 10 demo projects. Frontend: `types/periodic-report.ts`, `services/periodic-report.service.ts`, `components/project/PeriodicReportPanel.tsx` (inline form, computed preview, sortable table, mobile-friendly), wired to `/projects/[id]/page.tsx`. Dashboard disclaimer shows blue label "Laporan periodik operasional" for PERIODIC_REPORT, amber label for OPERATIONAL fallback. Verification: migration applied · seed-demo 10 projects · `go build ./...` OK · `go test ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK (31 pages) · smoke-test-dash-002 **28 PASS 0 FAIL 1 SKIP** · REL-001 regression **44/44 PASS** · dashboard/trend `data_type: PERIODIC_REPORT` with realistic 6-month data.

> **STATUS (2026-08-29 — UX-004 applied):** `PMO-UX-004` is `Done` (2026-08-29): Dashboard Trend Data & Sidebar Label Fix. Sidebar primary route label restored from `Dasbor` to `Dashboard` for professional app terminology. Dashboard trend panel is hardened: backend `/api/v1/dashboard/trend` now uses stable month buckets from the first day of the month, fills operational fallback points from current project progress/budget aggregates when monthly history is missing, and frontend falls back to project-derived operational trend data if the trend endpoint is unavailable/stale. The existing fullscreen trend modal remains active when data is present. Verification: `go test ./internal/platform/dashboard/...` OK · `go build ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK · API `/api/v1/dashboard/trend` 200 with 12 ordered points · Playwright dashboard smoke OK (`Gagal memuat data tren` absent, `Dashboard` visible, `Dasbor` absent, fullscreen modal opens). Note: Recharts emits development-only `defaultProps` warnings in console; render/build are unaffected.

> **STATUS (2026-08-29 — UX-003 applied):** `PMO-UX-003` is `Done` (2026-08-29): UI/UX Polish Sweep for non-IoT routes. Frontend sweep focused on professional app consistency: sidebar labels normalized (Dashboard, Eksekutif, Peta GIS, Pendukung Keputusan, Manfaat, Connector Pemerintah), topbar title mapping expanded for key modules, dashboard section nav translated (Ringkasan/Portofolio) and refresh action clarified, Projects and Users pages localized to Indonesian wording, project detail recurring section/form/empty/loading/error labels localized, Benefits page labels/empty state polished, Executive duplicated local header reduced, Health/Project Control/Gantt components refactored from dense one-line JSX into maintainable UI with Indonesian states. Verification: `npx tsc --noEmit` OK · `npm run lint` OK · clean `npm run build` OK · clean dev restart · `bash docs/smoke-test-uat-001-ui-full.sh` **101/101 PASS**.

> **STATUS (2026-08-29 — UX-002 applied):** `PMO-UX-002` is `Done` (2026-08-29): Professional App Shell Sidebar Polish. Sidebar sections are now collapsible/expandable per group (Operasional, Control Tower, Data & Pelaporan, Integrasi, Pengaturan), active section opens automatically, the desktop sidebar can be hidden/shown from the topbar for full-width workspace mode, and the bottom-left promotional image card has been removed. Verification: `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK · focused Playwright sidebar check OK (sections collapse, promo absent, hide/show shifts layout) · `bash docs/smoke-test-uat-001-ui-full.sh` **101/101 PASS** after clean dev restart. Rule: do not run `npm run build` while `next dev` is active; stop dev, build, remove `.next`, restart dev.

> **STATUS (2026-08-29 — UX-001 applied):** `PMO-UX-001` is `Done` (2026-08-29): Sidebar IA Cleanup & Dashboard Section Navigation. Frontend IA fix: sidebar no longer exposes `/dashboard#portfolio`, `/dashboard#monitoring`, `/dashboard#risiko`, `/dashboard#isu`, `/dashboard#tindak-lanjut`, or `/dashboard#keputusan` as top-level modules. Sidebar is now grouped into professional module sections: Operasional, Control Tower, Data & Pelaporan, Integrasi, and Pengaturan. `/dashboard` keeps anchors as an in-page horizontal section nav (Overview, Portfolio, Monitoring, Risiko, Isu, Tindak Lanjut, Keputusan). Command Center panel actions no longer route to dashboard anchors; project-related actions go to `/projects`, map action goes to `/gis`. Additional FE hardening: Sidebar waits for client mount before role-filtered nav expansion, fixing hydration mismatch caused by persisted auth roles changing icon order between SSR and client render; `/audit-logs` and `/notifications` are wrapped in `DashboardLayout` so they have sidebar/topbar like other app pages. Verification: `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK · `bash docs/smoke-test-uat-001-ui-full.sh` **101/101 PASS**. Dev server restarted clean after build (`.next` removed first); `/login` 200, `/dashboard` 307. Rule: dashboard anchors are internal page navigation, not sidebar modules.

> **STATUS (2026-08-29 — UAT-010 applied):** `PMO-UAT-010` is `Done` (2026-08-29): Final Acceptance Review & Demo Polish. Status: **Non-IoT UAT Candidate**. Deliverables: (1) `scripts/` directory created — `uat-env-check.sh` (15 PASS 2 WARN 0 FAIL), `uat-start.sh` (migrate+seed+start backend+frontend), `uat-status.sh` (health+login check), `uat-stop.sh` (PID-safe stop, no external process kill); (2) `docs/smoke-test-uat-008-release-candidate.sh` — RC orchestrator: 7 PASS + 1 WARN (UAT-001 non-blocking: Playwright login timeout pre-existing) + 0 FAIL → **RELEASE CANDIDATE: READY**; (3) `docs/UAT-DEPLOYMENT-GUIDE.md` — full deployment guide (setup, accounts, migrations, troubleshooting, known limitations); (4) `docs/RELEASE-CANDIDATE-RUNBOOK.md` — RC gate runbook (go/no-go criteria, exec steps, hotfix procedure, RC sign-off checklist, post-RC next steps); (5) `docs/UAT-FINAL-CHECKLIST.md` — pre/during/post demo checklist + go/no-go criteria + known limitations; (6) `backend/.gitignore` created (.env, build outputs, tmp, .DS_Store); (7) Docs synced: `PHASED-IMPLEMENTATION-PLAN.md` phase table all ✅ Done + Phase 8 UAT added; `IMPLEMENTATION-GAP-ANALYSIS.md` v0.2.8 + UAT-006/007/010 update blocks; `PERMISSION-MATRIX.md` v0.2.3 + `ResourceCorrectiveAction` update block; `DEVELOPMENT-BACKLOG.md` UAT-010 Done entry. Security: no production secrets found in docs/scripts/env; `backend/.env` = local dev values only (intentional); `backend/.env.example` = CHANGE_ME placeholders. Known limitations: IoT Future/Last, gov connector sandbox/mock, BIM metadata-only, storage local filesystem, SMTP NoopProvider if unconfigured, UAT-001 Playwright login timeout pre-existing (non-blocking). Smoke: `uat-env-check.sh` 15/0 PASS/FAIL · `uat-status.sh` UP+LOGIN OK · `smoke-test-uat-007-business-flow.sh` **86/86 PASS** · `smoke-test-uat-006-role-permissions.sh` **65/65 PASS** · `smoke-test-rel-001-regression.sh` **44/44 PASS** · `smoke-test-uat-008-release-candidate.sh` **7 PASS 1 WARN 0 FAIL RC READY** · `gofmt` clean · `go build ./...` OK · `go test ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK. **GO recommendation: Non-IoT UAT Candidate ready for stakeholder demo.** **Next (non-IoT):** staging deployment → stakeholder UAT session → bug collection → production hardening. IoT (`PMO-P3-004`) tetap Future/Last.

> **STATUS (2026-08-29 — UAT-007 applied):** `PMO-UAT-007` is `Done` (2026-08-29): End-to-End Business Process UAT Scenarios & Data Consistency Gate. Deliverables: (1) `docs/UAT-BUSINESS-FLOW-SCENARIOS.md` — comprehensive business flow UAT doc with 11 flows (A-K): Project Setup & Onboarding, Vendor/Contract/Budget, Risk & Issue Register, Corrective Action, Field Evidence & Document, Governance Official Approval, Command Center & Escalation, Analytics & Report Export, RBAC & Audit Trail, Notifications, Dashboard Data Consistency; (2) `docs/smoke-test-uat-007-business-flow.sh` **86/86 PASS, 0 FAIL, 1 SKIP** — full end-to-end flows A→L incl. FSM transitions (project DRAFT→PLANNING→ACTIVE, risk/issue via PUT, CA DRAFT→SUBMITTED→IN_PROGRESS, governance DRAFT→SUBMITTED→IN_REVIEW→APPROVED→LOCKED), budget NORMAL→RISK→OVERRUN thresholds, governance duplicate period 409, evidence download (correct evidence ID from `data.evidence[0].id`), escalation+decision command center, cleanup soft-delete cascade; (3) Backend fixes: `cmd/seed/main.go` — `ResourceCorrectiveAction` added to `resources` slice (PMO/ADMIN now have corrective_action permission); `gofmt -w` applied to `internal/core/notification/service.go` and `internal/platform/server/server.go`. Key API contract corrections discovered: project transition uses `to_status` (not `status`); CA FSM is DRAFT→SUBMITTED→IN_PROGRESS→COMPLETED (not OPEN→IN_PROGRESS); inspection create uses multipart form binding (not JSON), field `inspected_at`; governance `dataset_type` must be `PROJECT_PROGRESS`/`BUDGET`/etc (not `"project"`); decision status valid values are `IN_PROGRESS`/`COMPLETED`/`CANCELLED` (not `DECIDED`); evidence download URL uses evidence ID from `data.evidence[0].id` (not parent inspection ID); `AddEvidence` response returns parent Inspection object. UAT-007 smoke **86/86 PASS**. Regression `docs/smoke-test-rel-001-regression.sh` **44/44 PASS**. UAT-006 **65/65 PASS**. `gofmt` clean · `go build ./...` OK · `go test ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK. **Next (non-IoT):** non-IoT Phase 3 lain. IoT (`PMO-P3-004`) tetap Future/Last.

> **STATUS (2026-08-29 — UAT-006 applied):** `PMO-UAT-006` is `Done` (2026-08-29): Role-Based UAT & Permission Scope Hardening. (1) Demo users seeded idempotent (5 users per role): `pmo@cankora.local`, `pm@cankora.local`, `officer@cankora.local`, `executive@cankora.local`, `auditor@cankora.local` — all password `Demo@Cankora2024!`; `auth.User` uses `FirstName`/`LastName`. (2) `cmd/seed/main.go` — `resources` slice extended with `ResourceSector`, `ResourceRegion`, `ResourceRiverBasin`; `demoUserDef` struct corrected to `FirstName`/`LastName`; demo users section idempotent via `FirstOrCreate`. (3) Backend permission guard audit: spatial routes `/sectors`, `/regions`, `/river-basins` previously only had `AuthRequired` — `RequirePermission` per-route added in `server.go` with granular actions (view/create/update/delete). (4) Frontend `Sidebar.tsx` — role constants, role sets (`ADMIN_ROLES`, `PMO_AND_ABOVE`, `PROJECT_ROLES`, `REPORTING_ROLES`, `EXECUTIVE_ROLES`), `NavItem.roles` filter via `hasRole`, render loop uses `visibleNavItems`/`visibleSettingsItems`; backend remains the authoritative source of authorization. (5) `docs/smoke-test-uat-006-role-permissions.sh` **65/65 PASS** — 12 sections: health, no-token 401 (8 paths), demo login (6 roles), ADMIN full access, PMO broad access, PM allowed+forbidden, OFFICER allowed+forbidden, EXECUTIVE analytics read-only + write 403, AUDITOR logs+reports + forbidden, spatial permission guards, fake token 401, 401 vs 403 semantics. Regression `docs/smoke-test-rel-001-regression.sh` **44/44 PASS**. `gofmt` clean · `go build ./...` OK · `go test ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK (30 pages). **Next (non-IoT):** non-IoT Phase 3 lain. IoT (`PMO-P3-004`) tetap Future/Last.

> **STATUS (2026-08-29 — UAT-005 applied):** `PMO-UAT-005` is `Done` (2026-08-29): Field Inspection Mobile View. (1) `frontend/src/components/project/FieldInspectionPanel.tsx` — full mobile-first rewrite: sub-komponen `StatusBadge`, `FilePicker` (preview nama/ukuran/tipe, clear button, validasi client-side type/size), `GeoInput` (grid 2-col lat/lon + tombol "Gunakan Lokasi Saat Ini" via `navigator.geolocation`, graceful fallback + error state jika izin ditolak); form upload evidence dengan upload state jelas (file terpilih visible, size/type, error, success); evidence list padat dengan `formatBytes`, checksum 8-char, geotag per evidence, download per item dengan loading state; action Verifikasi/Tolak/Hapus touch-friendly `min-h-[36px]`; tidak ada horizontal overflow di 390px; `flex-wrap`, `truncate`, `line-clamp`, `overflow-hidden`. (2) Route baru `frontend/src/app/(dashboard)/projects/[id]/inspections/page.tsx` — dedicated mobile-friendly inspection page, `DashboardLayout` + back nav + project title via `projectService.get`. (3) Backend: endpoint baru `POST /api/v1/projects/:id/inspections/:inspectionID/evidence` (`AddEvidence` handler + public `Service.AddEvidence`) — upload evidence ke inspeksi yang sudah ada; per-evidence lat/lon dari form; audit `field_inspection.evidence_added`; route registered `ResourceFieldEvidence:create`. (4) `frontend/src/services/field.service.ts`: method `addEvidence(projectId, inspectionId, file, lat?, lon?)`. (5) `gofmt -w` diterapkan ke semua modul. Smoke `docs/smoke-test-uat-005-field-mobile.sh` **28/28 PASS** (health, login, auth guard 401, list, create+evidence, evidence in list, download authorized 200+unauthorized 401, POST /:inspectionID/evidence, invalid type 400, verify VERIFIED+REJECTED, soft-delete+DB confirmed, cross-tenant 404, frontend 307, Playwright mobile 390x844 no overflow + desktop 1366x768). Regression `docs/smoke-test-rel-001-regression.sh` **44/44 PASS** (section 16 field inspection API: GET/POST/PATCH/DELETE, 5 checks). UAT-001 UI gate: 20 → 21 routes. `gofmt` clean · `go build ./...` OK · `go test ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK (30 pages). **Next (non-IoT):** non-IoT Phase 3 lain. IoT (`PMO-P3-004`) tetap Future/Last.

> **STATUS (2026-08-29 — UAT-004 applied):** `PMO-UAT-004` is `Done` (2026-08-29): Notification Delivery Foundation. Backend: migration `000037_notifications` (notifications table: id, organization_id, recipient_user_id, channel IN_APP/EMAIL, status PENDING/SENT/FAILED/READ, priority LOW/NORMAL/HIGH/URGENT, subject, body, source_type, source_id, error_message, sent_at, read_at, created_at, updated_at, deleted_at). `backend/internal/core/notification/` extended: `Notification` model, `Repository` interface (Create/GetByID/List/UpdateStatus/MarkAllRead/Summary), `Service` (Enqueue/EnqueueAndReturn/MarkRead/MarkAllRead/Retry), `provider.go` (NoopProvider falls back silently when SMTP env empty), `smtp.go` (SMTP delivery). `backend/internal/modules/notifications/handler.go` — new module; 8 routes: `GET /api/v1/notifications`, `GET /api/v1/notifications/summary`, `POST /api/v1/notifications/test`, `PATCH /api/v1/notifications/read-all`, `GET /api/v1/notifications/admin`, `GET /api/v1/notifications/:id`, `PATCH /api/v1/notifications/:id/read`, `POST /api/v1/notifications/:id/retry`; all tenant-scoped; permission guard `ResourceNotification:view`; `EnqueueAndReturn` returns full created Notification with `id` in response. Frontend: `frontend/src/types/notification.ts` (Notification, NotificationSummary, NotificationFilter, NotificationStatus, NotificationChannel, NotificationPriority); `frontend/src/services/notification.service.ts` (list/getById/summary/markRead/markAllRead/retry/createTest); page `frontend/src/app/(dashboard)/notifications/page.tsx` — summary cards (total/unread/pending/failed), filter bar (status/channel/priority/unread_only), paginated list with channel+status+priority badges, detail drawer (body/source/timestamps), mark-read, retry FAILED, create test button, loading/error/empty states; Sidebar: `Bell` icon + `/notifications` nav item. Smoke `docs/smoke-test-uat-004-notifications.sh` **25/25 PASS, 1 SKIP** (frontend 307 redirect = expected). Regression gate `docs/smoke-test-rel-001-regression.sh` **39/39 PASS** (4 notification checks added). UAT-001 UI gate updated to 20 routes. `gofmt` clean · `go build ./...` OK · `go test ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK (29 pages, /notifications 7.45 kB). **Next recommended step (non-IoT):** non-IoT Phase 3 lain. IoT (`PMO-P3-004`) tetap Future/Last.

> **STATUS (2026-08-29 — UAT-003 applied):**** `PMO-UAT-003` is `Done` (2026-08-29): Audit Log Viewer UI. Read-only audit log viewer for governance/compliance UAT/demo. Backend: `backend/internal/modules/auditlog/handler.go` — new module wired via `server.go`; routes `GET /api/v1/audit-logs`, `GET /api/v1/audit-logs/summary`, `GET /api/v1/audit-logs/export` (CSV, max 1000 rows), `GET /api/v1/audit-logs/:id`; all tenant-scoped by `organization_id`; permission guard `ResourceAuditLogs` + `ActionView` on all routes; `audit.ListFilter` extended with `Search` (ILIKE on actor_email/action/entity_type/entity_label); `audit.Repository` extended with `GetByID()` and `Summary()` returning `total_events`, `unique_actors`, `top_actions[5]`, `top_entities[5]`; `ErrNotFound` sentinel added to `audit/model.go`. Frontend: `frontend/src/types/audit-log.ts`, `frontend/src/services/audit-log.service.ts` (list/getById/summary/exportCsv); page `frontend/src/app/(dashboard)/audit-logs/page.tsx` — summary cards (total events, unique actors, top action/entity), filter bar (search/action/entity/date range), paginated table with expandable rows (old_values/new_values JSON, IP, UA, request_id), action/entity badges, loading/error/empty states, Export CSV button; Sidebar: `Activity` icon + `/audit-logs` nav item. Smoke `docs/smoke-test-uat-003-audit-logs.sh` **26/26 PASS, 1 SKIP** (frontend redirect to login — expected). Regression gate `docs/smoke-test-rel-001-regression.sh` **35/35 PASS** (5 new audit-logs checks added). UAT-001 UI full gate updated to 19 routes (added `/audit-logs`). `gofmt` clean · `go build ./...` OK · `go test ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK (28 pages, /audit-logs 5.46 kB). **Next recommended step (non-IoT):** notification delivery, atau non-IoT Phase 3 lain. IoT (`PMO-P3-004`) tetap Future/Last.

> **STATUS (2026-08-29 — UAT-002 applied):**** `PMO-UAT-002` is `Done` (2026-08-29): Report Export Real File. Export request queue sekarang menghasilkan file nyata (CSV & XLSX). Implementasi: (1) `export_generator.go` — CSV via `encoding/csv`, XLSX via `github.com/xuri/excelize/v2 v2.8.1`; storage lokal `backend/storage/reports/<orgID>/<date>/<filename>` (relative `storage_key`, path-traversal safe); 6 dataset: executive-summary, project-performance, risk-issue, budget, benefits, priority. (2) `service.go` — `ProcessExportRequest()` sync (status langsung COMPLETED); `DownloadExportFile()` tenant-safe; `auditLog()` helper menulis ke `audit_logs` (actor_id, entity_type, new_values). (3) `handler.go` — `ProcessExportRequest` dipanggil inline setelah create; download route baru `GET /api/v1/analytics/reports/export/requests/:id/download` (auth + org guard, `Content-Type` + `Content-Disposition`). (4) Migration `000036_report_export_file_metadata` — kolom baru `file_name`, `storage_key`, `mime_type`, `file_size`, `generated_at` di `report_export_requests` (idempotent `ADD COLUMN IF NOT EXISTS`). (5) Frontend: `ReportExportRequest` type diperluas (file_name, storage_key, mime_type, file_size, generated_at); `reportingService.downloadExportFile()` — authenticated blob download; `ExportsTab` diupdate: badge status PENDING/PROCESSING/COMPLETED/FAILED, file_name, file_size, generated_at, tombol **Unduh** untuk COMPLETED request. Audit events: `report.export.requested`, `report.export.completed`, `report.export.failed`, `report.export.downloaded`. Smoke `docs/smoke-test-uat-002-report-export.sh` **44/45 PASS** (1 SKIP: full cross-tenant test memerlukan 2 org). `go test ./...` OK · `go build ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK. Regression gate `docs/smoke-test-rel-001-regression.sh` **30/30 PASS**. P2-009 smoke **38/38 PASS** (check 9-3 diupdate: PENDING|PROCESSING|COMPLETED semua valid). **Next recommended step (non-IoT):** notification delivery, audit log viewer UI, atau non-IoT Phase 3 lain. Jangan mulai PMO-P3-004 (IoT).

> **STATUS (2026-08-29 — UAT-001 applied):** `PMO-UAT-001` is `Done` (2026-08-29): Demo/UAT Packaging + Full UI Regression Gate. Deliverables: (1) `docs/smoke-test-uat-001-ui-full.sh` — full UI regression gate 18 routes (login, dashboard, dashboard#portfolio, dashboard#monitoring, command-center, projects, executive, programs, reports/analytics, gis, governance, integrations/government, integrations/bim, integrations/primavera, decision-support, benefits, imports, validation); Playwright full checks (HTTP, non-blank, sidebar/topbar present, no horizontal overflow @1366×768, map not full-screen) with curl-only fallback and frontend auto-start; (2) `STRICT_FRONTEND=1` mode added to `docs/smoke-test-rel-001-regression.sh` (frontend reachability is FAIL instead of SKIP when not running); (3) **Layout bug fixed**: `/governance`, `/integrations/bim`, `/integrations/government`, `/integrations/primavera` were missing `DashboardLayout` shell (no sidebar/topbar) — wrapped, same class of bug as `/gis` earlier; (4) `docs/UAT-DEMO-GUIDE.md` new (clean start, demo account, 15–30min demo script, implemented/basic vs future features, known limitations, smoke UAT run, recovery steps); (5) docs updated: CLAUDE.md, DEVELOPMENT-BACKLOG.md, PHASED-IMPLEMENTATION-PLAN.md. IoT (`PMO-P3-004`) remains Future/Last. Seed-demo verified idempotent (10 DEMO projects, 5 programs, 5 sectors, GIS lat/lng, 6 priority scores, gov mappings MATCHED/PENDING/REJECTED, governance submissions, BIM/Primavera runs). `go test ./...` OK · `go build ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK. **Next recommended step (non-IoT):** report export real file, notification delivery, audit log viewer UI, or other non-IoT Phase 3 item; jangan mulai PMO-P3-004 (IoT).

> **STATUS (2026-08-29 — REL-001 applied):** `PMO-REL-001` is `Done` (2026-08-29): Non-IoT Stabilization, Docs Reconciliation, and Regression Gate. Fixes: (1) `actionRequiresEntityID()` helper — UPDATE/DELETE/UPSERT wajib `entity_id` (→ 400 jika kosong), CREATE & VALIDATE_ONLY boleh omit; rule diterapkan di `CreateSubmission` item validation dan `validateItem()`; (2) `ErrEntityIDRequired` error baru; (3) `TestActionRequiresEntityID` unit test (6/6 governance tests PASS); (4) smoke harden diupdate: section 2b setup fixture idempotent (PENDING_MATCH/REJECTED/MATCHED via `INSERT ... ON CONFLICT DO NOTHING`), check 7 (MATCHED→VALID) tidak lagi SKIP (sekarang PASS/FAIL dengan fixture yang dijamin ada), check 13 baru (5 sub-checks entity_id rules CREATE/UPDATE/DELETE/UPSERT/VALIDATE_ONLY); (5) MATCHED fixture di-seed idempotent ke DB (b690b84d-... → project 0782c39f-...); (6) `docs/smoke-test-rel-001-regression.sh` regression gate terpadu (15 kelompok check: health, auth guard, dashboard, projects CRUD, risks/budgets sub-resource, command center, analytics programs/executive/sectors, frontend reachability, governance, government integration, BIM, Primavera, audit integrity, data model integrity). All P0–P3 non-IoT verified stable. Docs di-sync: CLAUDE.md, DEVELOPMENT-BACKLOG.md, PHASED-IMPLEMENTATION-PLAN.md, IMPLEMENTATION-GAP-ANALYSIS.md, PERMISSION-MATRIX.md. Dashboard/reporting label audit: /executive=Level1, /programs=Level2, /command-center=operational, /reports/analytics=read-model, /governance=official-governed-data — labels sudah tepat, tidak ada mixing tanpa disclaimer. `go build ./...` OK · `go test ./internal/modules/governance/...` 6/6 OK · `gofmt` clean. **Next recommended ticket:** PMO-P3-004 (IoT Telemetry) adalah future/last item — jangan dimulai. Setelah REL-001, sistem siap UAT/demo packaging atau penambahan non-IoT feature lanjutan.

> **STATUS (2026-08-28 — P3-003-HARDEN applied):** `PMO-P3-003-HARDEN` is `Done` (2026-08-28): Data Governance Official Validation Hardening. (1) Validasi `government_mapping`: `.Row().Scan()` menggantikan `Raw().Scan()` untuk deteksi not-found yang benar; hanya `MATCHED` yang VALID — `PENDING_MATCH`→`ErrPendingMatchNotReady`, `REJECTED`→`ErrMappingRejected`, empty/unknown→`ErrUnknownMatchStatus`, not-found/cross-tenant→`ErrEntityNotFound`. (2) Validasi `entity_type`: `normalizeEntityType()` menolak free-text/typo dengan `ErrInvalidEntityType` — hanya `project/projects`, `vendor/vendors`, `government_mapping` yang valid. (3) Validasi `source_entity_id` tenant-scoped di `CreateSubmission`/`Submit`/`StartReview`. (4) `Approve()` re-validates semua items FRESH (bukan percaya stale status); persist hasil terbaru ke DB; gagal dengan `ErrInvalidItems` jika ada yang berubah jadi INVALID. (5) `Lock()` memanggil `ensureLockPeriod()` — buat/kunci lock-period row sehingga `CreateSubmission` di periode yang sama → 409. (6) Migration `000034`: expression unique index `COALESCE(period_month,0)` menggantikan NULL-distinct unique index — full-year lock tidak bisa duplicate. (7) Migration `000035`: `created_by` column di `data_submissions` — diset saat DRAFT create; `submitted_by` TIDAK diset saat create (hanya saat Submit). (8) Handler `bindOptionalJSON()`: empty body diizinkan, malformed JSON → 400 (bukan diabaikan). (9) Semua `_ = c.ShouldBindJSON()` di optional actions diganti `bindOptionalJSON`. `go mod tidy` mempromosikan `pgx/v5` ke direct dep. `gofmt` clean · `go build ./...` OK · `go test ./internal/modules/governance/...` 5/5 OK (TestTransitionAllowed, TestValidItemAction, TestNormalizeEntityType, TestIsUniqueViolation, TestMonthHelpers) · `npx tsc --noEmit` OK · `npm run lint` OK. Smoke P3-003 lama 17/17 PASS (diupdate pakai `SMOKE_YEAR` random agar idempotent). Smoke hardening `docs/smoke-test-p3-003-harden.sh` 19/19 PASS. **Next recommended ticket:** audit backlog Phase 3 non-IoT yang tersisa.

> **STATUS (2026-08-28 — UI-REGRESSION-001 fixed):** Dashboard/GIS UI regression selesai. Root cause: `.next` directory corrupt (production build artifacts tercampur dengan dev server) menyebabkan `/_next/static/css/app/layout.css` → 404 → Tailwind classes hilang → `next/image fill` di dashboard "Peta Sebaran Proyek" tampil full-screen tanpa sidebar/topbar. Fix: hapus `.next` + restart dev server bersih. Juga ditemukan bug pre-existing di `/gis`: (1) `DashboardLayout` tidak di-wrap → no sidebar/topbar → FIXED di `gis/page.tsx`; (2) `leaflet/dist/leaflet.css` tidak di-import → map styling rusak → FIXED di `GISMap.tsx`; (3) map height tidak terkontrol → FIXED dengan `h-[420px] lg:h-[calc(100vh-280px)]` pada wrapper; (4) legend bar `flex` tidak wrap di mobile → horizontal overflow → FIXED dengan `flex-wrap gap-x-4`; (5) SummaryCard tidak punya `min-w-0` → teks nilai melebar di grid 2-kolom mobile → FIXED; (6) input filter tidak `w-full sm:w-auto` → melebar di viewport sempit → FIXED; (7) Leaflet tile transform menyebabkan scrollWidth > clientWidth via Chromium quirk → FIXED dengan `overflow-x-hidden` di wrapper page GIS. Verifikasi: `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK (semua halaman compile termasuk /gis 6.9kB). Smoke script: `docs/smoke-test-ui-regression-dashboard-gis.sh`. **RULE: Jangan jalankan `npm run build` saat dev server aktif dan memakai `.next` yang sama. Sebelum build: stop dev server; setelah build: hapus `.next`, restart dev server.** Next recommended ticket: audit backlog Phase 3 non-IoT yang tersisa.

> **STATUS (2026-08-28 — P3-003 applied):** `PMO-P3-003` is `Done` (2026-08-28): Official Data Validation & Approval Workflow — Data Governance. Backend module `backend/internal/modules/governance/` (model/service/handler + `service_test.go` unit tests for FSM/helpers). Lifecycle: DRAFT→SUBMITTED→IN_REVIEW→APPROVED→LOCKED (+REJECTED with required reason, CANCELLED, REJECTED→DRAFT); per-item validation (tenant ownership, soft-delete, government `PENDING_MATCH` cannot be official); lock periods (OPEN/LOCKED, unique per org+dataset+period, blocks writes in locked period); audit `governance.submission.created/submitted/review_started/approved/rejected/locked/cancelled` + `governance.lock_period.created/locked`. 12 routes at `/api/v1/governance/`; `ResourceDataGovernance` seeded. Fixes applied: duplicate `package governance` (build was broken), `data_submissions` FK/status-check relaxation via migrations `000031_governance_relax_fk`, `000032_governance_period_month_nullable`, `000033_governance_drop_legacy_checks` (`project_id`/`snapshot_id`/`period_month` nullable; legacy `data_submissions_status_check` dropped — it blocked IN_REVIEW/APPROVED/LOCKED/CANCELLED with SQLSTATE 23514); `dataquality` List/Get now filter `snapshot_id IS NOT NULL` so governance rows never leak into the validation queue. Frontend: `frontend/src/types/governance.ts`, `frontend/src/services/governance.service.ts`, page `/governance` (list+filter, create submission modal with dynamic items, detail modal with status timeline + items table + state-aware action buttons, lock period panel), Sidebar item "Data Governance". Smoke test `docs/smoke-test-p3-003-data-governance.sh` 17/17 PASS. `go build ./...` OK · `go test ./...` OK · `npx tsc --noEmit` OK · `npm run lint` OK · `npm run build` OK. **Planning update:** IoT telemetry is no longer the immediate next ticket. `PMO-P3-004` is parked as the final/future advanced-integration item after governance hardening, UI regression fixes, reporting/read-model reconciliation, and other non-IoT Phase 3 work are stable.

> **STATUS (2026-08-28 — P3-002 applied):** `PMO-P3-002` is `Done` (2026-08-28): Government Entity Resolution — PENDING_MATCH → MATCHED end-to-end. Migration `000029_government_resolution_metadata` adds `match_confidence SMALLINT`, `match_reason VARCHAR(50)`, `matched_by UUID`, `matched_at TIMESTAMPTZ`, `rejected_by UUID`, `rejected_at TIMESTAMPTZ`, `reject_reason TEXT` to `government_external_mappings`. New `resolver.go` in `government` module: `GetMapping`, `ListPendingMappings`, `GetCandidates` (project by code/name, vendor by NPWP/name), `MatchMapping`, `UnmatchMapping`, `RejectMapping` — all tenant-scoped with audit. 6 new routes: `GET /mappings/pending`, `GET /mappings/:id`, `GET /mappings/:id/candidates`, `POST /mappings/:id/match`, `POST /mappings/:id/unmatch`, `POST /mappings/:id/reject`. `ListMappings` updated to accept `match_status` filter. Frontend: `ExternalMapping` types extended (`match_status`, `match_confidence`, `match_reason`, `matched_by/at`, `rejected_by/at`, `reject_reason`, `internal_entity_id: string | null`); `MatchStatus`, `MatchConfidence`, `ResolutionCandidate`, `MatchMappingRequest`, `RejectMappingRequest` types added; `government.service.ts` extended with 6 new methods; government page gains "Resolusi Entitas" tab with `ResolutionTab`, `MappingDrawer`, `CandidateList`, `MatchStatusBadge`, `ConfidenceBadge` components. Smoke test `docs/smoke-test-p3-002-government-resolution.sh` 13 checks. `go build ./...` OK · `npx tsc --noEmit` OK.

> **STATUS (2026-08-28 — HARDEN-001 applied):** `PMO-HARDEN-001` is `Done` (2026-08-28): BIM/Government integration clean code & data integrity hardening. Fixes: (1) duplicate `package bim` removed from `bim/model.go` — build was broken; (2) `LinkProject` now validates `project_id` belongs to caller's org and is not deleted (`ErrProjectNotFound` → 404); (3) all BIM audit actions renamed to specific events: `bim.model.created`, `bim.model.updated`, `bim.model.deleted`, `bim.version.created`, `bim.project.linked`, `bim.project.unlinked`; (4) `upsertMapping` in government ingestor no longer writes fake `uuid.New()` placeholder — new mappings created with `internal_entity_id = NULL` and `match_status = PENDING_MATCH`; (5) migration `000028_government_external_mapping_match_status` makes `internal_entity_id` nullable and adds `match_status VARCHAR(30) NOT NULL DEFAULT 'PENDING_MATCH'`; smoke test `docs/smoke-test-harden-001-integrations.sh`. `go build ./...` OK · `gofmt` clean · `go test ./...` OK.

> **STATUS (2026-08-28):** P0-001 s.d. P0-014 are `Done` and locally verified; P1-001 through P1-016 are `Done` and verified. P2-003 is `Done`: tenant-scoped Level 3 Project Control API and panel reconcile contract, VALID snapshot, variance, health, evidence, issue/risk, and corrective action. P2-010 is `Done`: benefit indicators and measurements are tenant-scoped, auditable, aggregated by compatible unit/method, and usable from `/benefits`. P2-004 is `Done`: priority scoring with configurable formula (DRAFT/ACTIVE/ARCHIVED versioning, 7-component weighted scoring, explainability breakdown, batch calculate, ranking) available at `/decision-support`; smoke test at `docs/smoke-test-p2-004-priority.sh`. P2-006 is `Done`: Level 2 program dashboard — analytics/program module aggregates KPI per program/sector (budget, physical progress, health distribution, risk/issue counts, priority score, benefit indicators) at `GET /analytics/programs`, `GET /analytics/programs/:id`, `GET /analytics/sectors`, `GET /analytics/sectors/:id`; 5 demo programs + 5 sectors seeded, all 15 projects linked; frontend `/programs` dashboard with KPI cards, drill-down detail panel, top deviations, health distribution, project table; smoke test `docs/smoke-test-p2-006-analytics.sh` 30/30 PASS. P2-007 is `Done`: Level 1 executive dashboard — `analytics/executive` module (model/service/handler) at `GET /api/v1/analytics/executive`; national summary (total/active projects, budget, health distribution, risks, issues, escalations, pending decisions, benefit indicators), critical projects LATERAL JOIN, open escalations, pending decisions, program comparison, benefit summary; `ResourceExecutiveDashboard` seeded; frontend `/executive` page with KPI cards, health bar, critical projects table, escalation list, decision queue, program comparison, benefit indicators, period filter; smoke test `docs/smoke-test-p2-007-executive.sh` 30/30 PASS. P2-002 is `Done`: Government Connector Foundation — mock connectors for SIRUP (projects), OM-SPAN (budget_allocations), location reference, and vendor reference; migration 000026; module `internal/modules/integration/government` (model, config, registry, ingestor, service, handler); 9 endpoints under `/api/v1/integrations/government/`; SAMPLE/DRY_RUN/COMMIT modes; `ResourceGovernmentConnector` seeded; frontend `/integrations/government` page with connector cards, sync run history, external mappings tab; smoke test `docs/smoke-test-p2-002-government-connectors.sh` 22/22 PASS. Historical note only; current resume guidance is below and IoT is parked as final/future.

### Build & Smoke Status (2026-08-25 — verified, P1-002 & P1-003 done; P1-004 verified 2026-08-26)

- Backend: `gofmt` clean · `go build ./...` OK · `go test ./...` OK (no test files present; build+smoke is the gate) · `make seed` OK (idempotent)
- Backend runtime fix (P1-016): dashboard `/api/v1/dashboard` formerly returned HTTP 500 (`column pb.organization_id does not exist`) because `project_budgets` has no `organization_id` column. Patched `budgetThresholdWarnings` in `backend/internal/platform/dashboard/handler.go`: org scope enforced via `p.organization_id` in the JOIN; soft-delete filter `pb.deleted_at IS NULL` retained (valid column). Dashboard API now returns 200 with live budget warnings (`Anggaran Berisiko` rendered).
- Risk (P1-002): `risks` table has `organization_id` + `probability`/`impact` INT 1-5 + `risk_score` + `severity` (migration 000009). Dashboard `/api/v1/dashboard` now emits `RISK_REGISTER` early warnings from the risk register (score desc, open status, active parent project). Dashboard "Risiko Utama" and Command Center "Risiko Prioritas" read risk-register data when available; fallback label clearly states it is derived from early warning otherwise. No Health Score claim (P1-014 pending).
- Budget (P1-003): budget line CRUD tenant-safe via parent project ownership (`project_budgets` has NO `organization_id` — org scope enforced through `projects.organization_id`). `variance`, `usage_pct`, and `status` (NORMAL <80 / WATCH ≥80 / RISK ≥90 / OVERRUN ≥100) are always computed backend-side from planned/actual (round 2). Routes `GET/POST /api/v1/projects/:id/budgets` + `GET/PUT/DELETE /api/v1/projects/:id/budgets/:budgetID` with `RequirePermission(ResourceBudgets, ...)`; audit `budget.created`/`updated`/`deleted`. Dashboard `BUDGET_THRESHOLD` warning aggregates per project with `pb.deleted_at IS NULL` + `p.deleted_at IS NULL` and message clarifies "operational input, not yet validated". Frontend Budget panel in project detail (totals, status badges, create/edit/delete, empty/loading/error states).
- Vendor/Contract (P1-004): `vendors` master (tenant-scoped, type `VENDOR`/`CONSULTANT`) + `contracts` project contract (tenant via `contracts.organization_id` AND parent `projects.organization_id`). Migration `000010_vendors_contracts`; partial unique index `uq_contracts_org_number` (contract_number unique per org on non-deleted rows). Routes `GET/POST /api/v1/vendors` + `GET/PUT/DELETE /api/v1/vendors/:vendorID` (`RequirePermission(ResourceVendors, ...)`), `GET/POST /api/v1/projects/:id/contracts` + `GET/PUT/DELETE /api/v1/projects/:id/contracts/:contractID` (`RequirePermission(ResourceContracts, ...)`). tenant/project ownership selalu diverifikasi (contract of wrong/missing project → 404). Audit `vendor.created`/`updated`/`deleted`, `contract.created`/`updated`/`deleted`. Vendor yang direferensikan contract aktif tidak bisa dihapus → 409. Frontend panel "Kontrak" in project detail (summary totals, vendor/consultant selectors, create/edit/delete, empty/loading/error states, footer note "operational contract input — belum validated/published").
- Document (P1-005): `project_documents` metadata (name, category, version, mime, size, uploaded_by, file_url RELATIVE storage key) — TIDAK punya `organization_id`, tenant-safe via parent project ownership. Migration `000011_documents_version` adds `version` + `updated_at`. Storage Phase 1 = local filesystem `backend/storage/documents` (env `STORAGE_LOCAL_PATH`, `STORAGE_MAX_SIZE_BYTES` default 20MB); helpers in `document_storage.go` designed for later MinIO/S3 swap. Routes `GET/POST /api/v1/projects/:id/documents` + `GET/PUT/DELETE /api/v1/projects/:id/documents/:documentID` + `GET /:documentID/download` (`RequirePermission(ResourceDocuments, ...)`). Upload multipart; extension allowlist pdf/doc/docx/xls/xlsx/ppt/pptx/png/jpg/jpeg/txt/csv + magic-bytes sniff; invalid type → 400, >20MB → 400, wrong/missing project → 404. Delete soft-deletes row AND removes physical file (path-escape guarded). Audit `document.uploaded`/`updated`/`downloaded`/`deleted`. Frontend panel "Dokumen" in project detail (Total Files / Total Size summary, category badge, upload form with file+category+version, list with download/edit/delete, empty/loading/error states, footer note "operational evidence input — belum validated/published menunggu P1-012").
- Command Center (P1-015): `GET /api/v1/command-center` aggregates risk watchlist, corrective action follow-up, validation SLA, persisted escalations, and executive decisions. `POST/PATCH /api/v1/command-center/escalations` and `/decisions` are protected by `ResourceCommandCenter`; create validates tenant ownership for project/source/escalation/user references; status updates set acknowledged/decided metadata and audit `command_escalation.*` / `executive_decision.*`.
- Benefit (P2-010): `benefit_indicators` + `benefit_measurements` are tenant-scoped with optional project link, unit, baseline/target/actual, period, owner/source, validation status, and aggregation method (`SUM`/`AVERAGE`/`LATEST`). API supports list/get/create/update/delete indicators, list/create/update/delete measurements, deterministic aggregate from VALID measurements, and `/benefits/summary` grouped by compatible unit+method so incompatible units are never summed. Frontend `/benefits` provides summary cards, indicator creation, measurement input, drill-down list, delete actions, and loading/error/empty states.
- Frontend: `npm run type-check` OK · `npm run lint` OK · clean `npm run build` OK (9 routes + middleware); dev server recovered after removing `.next`; `/_next/static/css/app/layout.css` HTTP 200
- Viewport smoke (P1-016): `/login`, `/dashboard`, `/command-center`, `/projects`, `/users` — no horizontal overflow and no unpositioned `next/image fill` at 1366×768, 1920×1080, 2560×1440, and mobile 390×844; mobile sidebar hamburger ("Buka/Tutup navigasi") present; no console errors on fresh reload (dashboard seeded with 10 projects / 6 active / 17 early warnings).
- Smoke API: 33/33 PASS — runnable artifact `docs/smoke-test-p0-013.sh` (regression). Plus risk lifecycle smoke 19/19 PASS (P1-002).
- Smoke UI: login → `/dashboard` tanpa redirect loop · redirect unauth → `/login?from=` · dashboard live · project CRUD + transition · task CRUD + kanban move · milestone CRUD + transition · issue CRUD + FSM (P1-001) · risk CRUD + FSM + score recompute (P1-002) · users list · CSS termuat
- Issue management (P1-001, Done): API smoke lifecycle PASS (create/list/FSM transition/delete/404) · UI smoke PASS (create via form → FSM OPEN→IN_PROGRESS→RESOLVED → edit resolution → delete) · audit `issue.created`/`updated`/`deleted` tercatat · migration `000008_issues_tenant_escalation` (organization_id + escalation)
- Known fix: create project dengan `code` duplikat kini return HTTP 409 `"Project code already in use"` (dulu 500). Implementasi: `isUniqueViolation()` (pgconn SQLSTATE 23505) di `backend/internal/modules/project/service.go` → `ErrCodeTaken` → `response.Conflict`.
- Known hardening (P0-014, Done): `Repository.DeleteCascade` soft-deletes project + all children (tasks, milestones, issues, risks, budgets, documents, team, task_assignments) in one transaction; dashboard task/milestone/overdue counts `JOIN projects ... deleted_at IS NULL`; migration `000007_child_soft_delete` adds `deleted_at` to `project_teams`/`task_assignments`/`project_budgets`; smoke cleanup deletes sub-task before parent; orphan smoke rows cleaned (3 tasks, 1 budget, 9 teams); SQL audit = 0 orphan child on soft-deleted projects.

### API Contract Notes (aktual — beda dari ERD konsep)

- Project: `code`, `name`, `start_date`/`end_date` (bukan `planned_start`/`planned_end`), `progress_pct`. FSM: DRAFT→PLANNING→ACTIVE→ON_HOLD→COMPLETED/CANCELLED (`ACTIVE→CANCELLED` valid).
- Milestone: `title` + `due_date` (bukan `name`/`target_date`). FSM: PENDING→IN_PROGRESS→COMPLETED/DELAYED.
- Task: `title`, `progress_pct` (bukan `progress`). FSM: TODO→IN_PROGRESS→IN_REVIEW→DONE/BLOCKED.
- Issue (P1-001): `title`, `status`, `severity` (LOW/MEDIUM/HIGH/CRITICAL), `escalation` (NONE/PROJECT_MANAGER/PROGRAM_MANAGER/EXECUTIVE), `assigned_to`/`reported_by` (UUID), `due_date`, `resolution`. FSM: OPEN→IN_PROGRESS→RESOLVED→CLOSED (RESOLVED→OPEN, CLOSED→OPEN valid). Routes: `GET/POST /api/v1/projects/:id/issues`, `GET/PUT/DELETE /api/v1/projects/:id/issues/:issueId`. Wajib `organization_id` (migration 000008).
- Risk (P1-002): `title`, `description`, `status` (IDENTIFIED/ASSESSED/MITIGATED/ACCEPTED/ESCALATED/CLOSED), `probability`/`impact` (INT 1-5), `risk_score` (= probability × impact, dihitung backend), `severity` (LOW/MEDIUM/HIGH/CRITICAL dari score: 1-4 LOW, 5-9 MEDIUM, 10-15 HIGH, 16-25 CRITICAL), `mitigation`, `owned_by`/`created_by` (UUID), `due_date`. FSM: IDENTIFIED→ASSESSED→MITIGATED|ACCEPTED|ESCALATED; MITIGATED→CLOSED; ACCEPTED→CLOSED; ESCALATED→MITIGATED. Routes: `GET/POST /api/v1/projects/:id/risks`, `GET/PUT/DELETE /api/v1/projects/:id/risks/:riskID` — semua `RequirePermission(ResourceRisks, ...)`. Wajib `organization_id` (migration 000009). Dashboard `/api/v1/dashboard` menambah `RISK_REGISTER` warnings dari risk register (open risk, project aktif, score desc).
- Budget (P1-003): `category`, `description`, `planned`, `actual`, `currency` (default IDR, uppercase). `variance` (planned − actual), `usage_pct` (actual/planned ×100), dan `status` (NORMAL <80 / WATCH ≥80 / RISK ≥90 / OVERRUN ≥100) SELALU dihitung backend (round 2). `project_budgets` TIDAK punya `organization_id` — tenant-safe via ownership parent project (`projects.organization_id`). Routes: `GET/POST /api/v1/projects/:id/budgets`, `GET/PUT/DELETE /api/v1/projects/:id/budgets/:budgetID` — `RequirePermission(ResourceBudgets, ...)`. Partial update pakai `*float64` (planned/actual nil = tidak diubah, 0 eksplisit = set 0). Dashboard `BUDGET_THRESHOLD` warning agregat per project (`pb.deleted_at IS NULL` + `p.deleted_at IS NULL`), message menandai "operational input, not yet validated". Audit `budget.created`/`updated`/`deleted`.
- Vendor (P1-004): `name`, `type` (`VENDOR`/`CONSULTANT`), `legal_name`, `tax_id`, `contact_person`, `email`, `phone`, `address`, `is_active`. Routes: `GET/POST /api/v1/vendors` (filter `type`/`search`/`is_active`), `GET/PUT/DELETE /api/v1/vendors/:vendorID` — `RequirePermission(ResourceVendors, ...)`. Wajib `organization_id` (migration 000010). Vendor yang direferensikan contract aktif tak bisa dihapus (409 `vendor is referenced by contracts and cannot be deleted`); setelah contract dihapus baru bisa (204).
- Contract (P1-004): `contract_number` (wajib, unique per org via `uq_contracts_org_number` WHERE deleted_at IS NULL), `title`, `vendor_id` (wajib), `consultant_id` (opsional), `contract_value` (tidak negatif), `currency` (default IDR uppercase), `signed_date`, `start_date` (tidak setelah `end_date`), `end_date`, `status` (DRAFT/ACTIVE/AMENDED/COMPLETED/TERMINATED), `scope_of_work`. Routes: `GET/POST /api/v1/projects/:id/contracts`, `GET/PUT/DELETE /api/v1/projects/:id/contracts/:contractID` — `RequirePermission(ResourceContracts, ...)`. Tenant & parent project ownership selalu diverifikasi (contract di project salah → 404); `organization_id` wajib. Contract value belakangan menjadi sumber kontraktual input; BELUM otomatis mengubah `project.budget_total` (menunggu P1-011 snapshot/validasi). Audit `contract.created`/`updated`/`deleted`. Dashboard belum mengklaim contract health.
- Document (P1-005): `name`, `category` (CONTRACT/REPORT/EVIDENCE/PHOTO/BAST/TOR_KAK/OTHER), `version`, `file_size`, `mime_type`, `uploaded_by`, `file_url` (RELATIVE storage key `documents/{org}/{project}/{doc}/{safe_filename}`). Routes: `GET/POST /api/v1/projects/:id/documents`, `GET/PUT/DELETE /api/v1/projects/:id/documents/:documentID`, `GET /:documentID/download` — `RequirePermission(ResourceDocuments, ...)`. `project_documents` TIDAK punya `organization_id` — tenant-safe via ownership parent project (`projects.organization_id`), diverifikasi service sebelum akses. Upload = multipart `file` field (+ form `name`/`category`/`version`); download set `Content-Type` dari mime + `Content-Disposition` RFC5987 (`filename*`). Storage Phase 1 local disk `backend/storage/documents`; file fisik dihapus saat delete record (path traversal ditolak). Audit `document.uploaded`/`updated`/`downloaded`/`deleted` — `organization_id` audit dari claims JWT.
- Create user: pakai `role_ids` (UUID array). Deactivate response tanpa field `data` (hanya `message`).
- DELETE endpoints → HTTP 204 No Content (tanpa body).

## Overview

PMO (Enterprise Operations Platform by Cankonix) adalah modular monolith untuk operasi PMO. Operational MVP mencakup Project Management Monitoring; target produk berikutnya adalah PMO National Project Control Tower Level 1-3 dengan strategi hybrid native PMO + Power BI.

## Key Facts for AI Assistants

- **Go module**: `github.com/harmanto-49/cankora`
- **Backend port**: 8080
- **Frontend port**: 3000
- **Default admin**: admin@cankora.local / Admin@Cankora2024!
- **Architecture**: Modular Monolith (NOT microservices)
- **Multi-tenant**: organization_id on every table — enforce on all queries

## Development Source of Truth

Read these files in order before coding:

1. `CLAUDE.md` — project rules, stack, conventions, and local run context.
2. `docs/DEVELOPMENT-BACKLOG.md` — operational task tickets for Codex/Claude. Start here for execution.
3. `docs/PHASED-IMPLEMENTATION-PLAN.md` — phase order, entry criteria, phase gates, and release milestones.
4. `docs/IMPLEMENTATION-GAP-ANALYSIS.md` — dual baseline: local MVP versus full Control Tower.
5. `docs/PMO-CONTROL-TOWER-ANALYSIS.md` — interpretation of new materials, capability map, and roadmap boundaries.
6. `docs/SRS.md` — functional/non-functional requirements and MVP/Control Tower scope.
7. `docs/ARCHITECTURE.md` — architecture, auth flow, route patterns, workflow, analytics, GIS, and Power BI strategy.
8. `docs/ERD-CONCEPTUAL.md` — conceptual data model vs implemented data areas.
9. `docs/PERMISSION-MATRIX.md` — intended RBAC permission and scope model.

Operational rule: P0 is complete (P0-001..P0-014 `Done`). Start P1 per ticket dependency. Do not call PMO Control Tower feature-complete until the relevant P1/P2/P3 tickets are `Done`; a mockup or updated requirement document is not implementation evidence.

Current phase: **Phase 4 - Project Intelligence & Command** (P2 complete) + **Phase 7 - Advanced Integrations** (P3 in progress). All P2 tickets (P2-001 through P2-011) are `Done`. `PMO-P3-001` is `Done`: BIM/Digital Twin Integration Foundation — migration `000027_bim_models`, backend module `integration/bim` (model/service/handler), 10 routes at `/api/v1/integrations/bim/models`, `ResourceBIMIntegration` seeded, frontend page `/integrations/bim`, smoke test `docs/smoke-test-p3-001-bim.sh` 23/23 PASS (2026-08-28). `PMO-P3-002` is `Done` (2026-08-28): Government Entity Resolution — PENDING_MATCH → MATCHED end-to-end (migration `000029`, `resolver.go`, 6 routes, frontend "Resolusi Entitas" tab, smoke 13 checks). `PMO-P3-003` is `Done` (2026-08-28): Official Data Validation & Approval Workflow — Data Governance (module `governance/`, migrations `000030`–`000033`, 12 routes, frontend `/governance`, smoke 17/17 PASS). **Do not start IoT next.** `PMO-P3-004` (IoT telemetry ingestion) is intentionally moved to the last/future advanced-integration slot. Before IoT, finish governance review blockers, dashboard/GIS regression fixes, read-model/reporting reconciliation, and any remaining non-IoT Phase 3 tickets selected by backlog audit. Do not advance to IoT or a later phase unless dependency and phase gate in `docs/PHASED-IMPLEMENTATION-PLAN.md` are satisfied.

Planning note:
- Kanban task board minimal is part of `PMO-P0-010` and should live in the project detail task UI.
- Gantt/timeline is tracked separately as `PMO-P1-009`; implement a simple read-only task/milestone timeline first, with drag date editing deferred until the timeline is stable.

## PMO National Project Control Tower Target

- **Level 1**: national executive KPI, health distribution, national map, trend, critical project ranking, decision-required queue, and benefit summary.
- **Level 2**: program/sector control for bendungan, irigasi, pengairan/pengendalian banjir, air baku, and pertanian.
- **Level 3**: project contract/location/DAS, physical-financial-schedule variance, health explanation, evidence, issue/risk, and follow-up.
- **Hybrid BI**: PMO is the operational system of record; Power BI reads a governed analytics model and never writes operational tables.
- **Health classes**: `GREEN`, `YELLOW`, `RED`, `CRITICAL`; eight target dimensions are schedule, physical, financial, contract, risk, issue, quality, procurement.
- Health weights/thresholds are business-owned, configurable, versioned, and approved. Never infer them from mockup visuals.
- Values, dates, project names, and people shown in supplied mockups are illustrative only.
- BP3R slides are generic PMO references, not SDA domain requirements.

## Project Location

```
/Users/harmanto/Documents/Code/DEV/Project Management/PMO/
```

## Running the Project

```bash
# Backend (from backend/ directory)
cd backend && make dev

# Optional idempotent SDA dashboard demo dataset
cd backend && make seed-demo

# Frontend
cd frontend && npm run dev

# Full stack
docker-compose up -d
```

Demo data note:
- `make seed` remains limited to organization, RBAC, and default admin data.
- `make seed-demo` idempotently creates 10 `DEMO-SDA-*` projects plus tasks, milestones, issues, risks, budgets, and progress history for local dashboard evaluation.

Frontend runtime rule:
- Never run `npm run build` while `npm run dev` is using the same `frontend/.next` directory.
- Stop all Next.js dev processes before a production build. After build verification, remove generated `.next` if switching back to dev and start exactly one dev server.
- If the page renders as a full-screen image or unstyled HTML, treat it as missing/mismatched CSS: stop duplicate servers, clear `.next`, restart dev, and verify `/_next/static/css/app/layout.css` returns HTTP 200 before editing layout code.

## Tech Stack

| Component | Choice | Version |
|-----------|--------|---------|
| Backend language | Go | 1.23 |
| HTTP framework | Gin | 1.9.1 |
| ORM | GORM | 1.25.5 |
| Database | PostgreSQL | 15 |
| JWT | golang-jwt/v5 | 5.2.1 |
| Migrations | golang-migrate | 4.17.0 |
| UUID | google/uuid | 1.6.0 |
| Logger | uber/zap | 1.27.0 |
| Frontend | Next.js App Router | 14.2.5 |
| UI kit | shadcn/ui + Radix | — |
| State | Zustand | 4.5.4 |
| Data fetching | TanStack Query | 5.51.1 |
| Forms | React Hook Form + Zod | 7.52.1 / 3.23.8 |

## Directory Structure

```
backend/
  cmd/api/main.go              # Entry point
  cmd/migrate/main.go          # Migration runner
  cmd/seed/main.go             # Database seeder
  internal/
    platform/
      config/config.go         # Config from env vars
      database/database.go     # GORM + PostgreSQL connection
      response/response.go     # Standard API envelope
      middleware/
        auth.go                # JWT middleware + ClaimsFromContext
        rbac.go                # RequirePermission middleware
        cors.go                # CORS + security headers
        logger.go              # Structured request logging
        requestid.go           # X-Request-ID injection
      server/server.go         # DI wiring + route registration + graceful shutdown
    shared/
      constants/constants.go   # Role codes, permission resources/actions, status constants
      types/
        flextime.go            # FlexTime: accepts RFC3339 and YYYY-MM-DD
        pagination.go          # Pagination request/response
      utils/
        crypto.go              # bcrypt hash/verify (cost 12)
    core/
      auth/                    # User, UserSession, token service, login/logout/refresh
      audit/                   # Immutable audit log + async Writer
      organization/            # Organization + OrgUnit (5-level hierarchy)
      rbac/                    # Role, Permission, UserRole — full CRUD
      notification/            # Email interface, NoopProvider, SMTP
      user/                    # User management service
      workflow/                # FSM-based status transitions (Project/Task/Milestone/Issue/Risk/CorrectiveAction)
    modules/
      project/                 # Project, Task, Milestone, Issue, Risk, Budget, Document
      asset/                   # Phase 2 placeholder
      inventory/               # Phase 3 placeholder
      cmms/                    # Phase 4 placeholder
  migrations/                  # 000001..000008+ up/down SQL files

frontend/
  src/
    app/                       # Next.js App Router
      (auth)/login/            # Login page (public)
      (dashboard)/             # Protected dashboard layout
        dashboard/             # Dashboard overview
        command-center/        # PMO Command Center frontend shell
        projects/              # Projects CRUD
    components/
      ui/                      # shadcn/ui base components
      layout/                  # Sidebar, TopBar, DashboardLayout
    lib/
      axios.ts                 # Axios instance with interceptors + token refresh
      queryClient.ts           # TanStack Query setup
    store/
      auth.store.ts            # Zustand auth store (user, tokens, login/logout)
    types/
      api.ts                   # API envelope types
      auth.ts                  # Auth types
      project.ts               # Project types
    services/
      auth.service.ts          # Auth API calls
      project.service.ts       # Project API calls
```

## API Response Envelope

All backend responses follow this shape:
```json
{
  "success": true,
  "message": "...",
  "data": { ... },
  "meta": { "page": 1, "page_size": 20, "total": 100, "total_pages": 5 },
  "errors": null
}
```

## Auth Flow

1. `POST /api/v1/auth/login` → response envelope with `data: { access_token, refresh_token, expires_at, user }`
2. Access token: 15min JWT, passed as `Authorization: Bearer <token>`
3. Refresh token: 7-day JWT, stored by frontend auth store according to Remember me behavior
4. `POST /api/v1/auth/refresh` → rotates both tokens
5. `POST /api/v1/auth/logout` → revokes session by JTI

Frontend auth note:
- Login must redirect to `from` if present, otherwise `/dashboard`.
- Middleware reads the `cankora-auth` cookie, so Zustand auth persistence must keep cookie/session state in sync.
- Remember me checked means persistent browser session; unchecked means session-only.
- Never send credentials in query string.

## Permission System

Resources: projects, tasks, milestones, issues, risks, budgets, documents, team, users, roles, organizations, audit_logs, reports

Actions: view, create, update, delete, approve, export

Roles (system):
- SUPER_ADMIN — all permissions
- ADMIN — all permissions
- PMO — all resources, all actions
- PROJECT_MANAGER — project ops + team (no system admin)
- PROJECT_OFFICER — task execution (no approve/delete)
- EXECUTIVE_VIEWER — view + export only
- AUDITOR — audit_logs + reports, view + export only

## Status Transitions (FSM)

**Project**: DRAFT → PLANNING → ACTIVE → ON_HOLD → COMPLETED / CANCELLED

**Task**: TODO → IN_PROGRESS → IN_REVIEW → DONE / BLOCKED

**Milestone**: PENDING → IN_PROGRESS → COMPLETED / SKIPPED

**Issue**: OPEN → IN_PROGRESS → RESOLVED → CLOSED / WONT_FIX

**Risk**: IDENTIFIED → ASSESSED → MITIGATING → RESOLVED / ACCEPTED

## Coding Rules

### Backend
- Always check organization_id on queries — never allow cross-tenant data
- Soft-delete via GORM `DeletedAt` — never hard-delete business data
- Use `response.*` helpers, never write raw `c.JSON`
- Log errors with zap, never `fmt.Println`
- Config always from `config.Load()` — never hardcode values
- Run `go run cmd/api/main.go` from `backend/` directory (godotenv reads .env relative to CWD)

### Frontend
- All API calls through service layer in `src/services/`
- Use TanStack Query for server state, Zustand only for auth/UI state
- Forms: React Hook Form + Zod schemas — no manual validation
- `cn()` helper (clsx + tailwind-merge) for conditional classNames
- Never use `any` in TypeScript

## Database Conventions

- All PKs: `UUID DEFAULT gen_random_uuid()`
- All timestamps: `TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Soft delete column: `deleted_at TIMESTAMPTZ` (nullable)
- All business tables have `organization_id UUID NOT NULL REFERENCES organizations(id)`
- Money: `NUMERIC(20,2)` with `currency VARCHAR(10) DEFAULT 'IDR'`

## Environment Variables (Required)

```
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE
JWT_SECRET (min 32 chars)
APP_ENV (development | staging | production)
PORT (default 8080)
```

## Common Mistakes to Avoid

- Don't forget `organization_id` filter on every GORM query
- Don't skip FSM — call `workflow.CanTransition()` before any status update
- Don't hardcode UUIDs in SQL — always use `gen_random_uuid()`
- Don't use `gin.Default()` — use `gin.New()` + explicit middlewares (already done in server.go)
- Don't import `lib/pq` directly — it's a side-effect import only (`_ "github.com/lib/pq"`)

## Current Build Status & Next Steps

> **Agent berikutnya: lanjut ke `PMO-P2-002` (connector pemerintah) atau phase P3 di `docs/DEVELOPMENT-BACKLOG.md`. P0 dan P1 selesai; P2-001, P2-003, P2-004, P2-005, P2-006, P2-007, P2-008, P2-009, P2-010, P2-011 sudah Done dan verified lokal. P2-011: Primavera P6 Adapter — migration 000025 (primavera_sync_runs + primavera_activity_mappings), module `backend/internal/modules/integration/primavera` (model, parser XER+PMXML, service, handler), 6 endpoints `/api/v1/integrations/primavera/runs/...`, `ResourcePrimaveraSync` permission seeded, frontend `/integrations/primavera` page dengan upload + run history + mappings, smoke test `docs/smoke-test-p2-011-primavera.sh` 25/25 PASS; go build + tsc clean.**

### P0 Progress — updated 2026-08-21

- `PMO-P0-001` Done: task/milestone handlers compile, use current service signatures, and no longer use raw `c.JSON` for create/delete.
- `PMO-P0-002` Done: user service uses the real user repository, user update/deactivate are tenant-safe, and user create uses `response.Created`.
- `PMO-P0-003` Done: seed is repeatable, admin login returns `roles:["SUPER_ADMIN"]` and `must_change_pwd:false`, and RBAC middleware is mounted on protected business routes.
- `PMO-P0-004` Done: dashboard route no longer crashes on the local schema, auth refresh/me roles are verified, and backend smoke passed for auth, dashboard, projects, transition, progress history, tasks, milestones, and users.
- `PMO-P0-005` Done: frontend login redirects safely to `from` or `/dashboard`, auth cookie/storage stay aligned for middleware, Remember me remains persistent/session scoped, and password visibility toggle is present.
- `PMO-P0-006` Done: project service now targets `/progress-history`, includes typed progress history responses, and exposes project task/milestone nested CRUD methods without `any`.
- `PMO-P0-007` Done: project/task/milestone repository lookups and deletes are scoped by `organization_id` and `project_id`, parent project ownership is verified before nested access, and soft-delete smoke returns 404 after delete.
- `PMO-P0-008` Done: dashboard frontend uses typed live stats from `GET /api/v1/dashboard`, waits for auth hydration before fetching, and includes explicit loading, error, and empty states.
- `PMO-P0-009` Done: projects page provides list/search, create/edit form validation, selected project detail, status transitions, and progress history via `projectService`.
- `PMO-P0-010` Done: project detail route provides task/milestone list/create/edit/status move/delete flows plus a minimal Kanban board grouped by task status.
- `PMO-P0-011` Done: users page provides list/search, create/edit profile form, deactivate flow, selected user detail, and optional RBAC role ID assignment through `userService`.
- `PMO-P0-012` Done: dashboard response now includes tenant-scoped early warnings for overdue tasks, overdue milestones, low progress near end date, and project budget usage thresholds; frontend renders the warning panel with loading and empty states.
- `PMO-P0-013` Done: full local build and API/UI smoke verification documented in `docs/smoke-test-p0-013.sh`.
- `PMO-P0-014` Done: `DeleteCascade` soft-deletes project + all children in one transaction; dashboard excludes child of deleted projects; migration `000007` adds `deleted_at` to `project_teams`/`task_assignments`/`project_budgets`; smoke cleanup order fixed; orphan smoke rows soft-cleaned (3 tasks, 1 budget, 9 teams); SQL audit 0 orphan; smoke 33/33 PASS.

### P1 Progress — updated 2026-08-26

- `PMO-P1-001` Done: Issue management end-to-end. Backend: model Issue + `organization_id` + `escalation` (migration `000008_issues_tenant_escalation`); repo/service org-scoped + project-scoped; FSM issue (OPEN→IN_PROGRESS→RESOLVED→CLOSED); routes `GET/POST /projects/:id/issues` + `GET/PUT/DELETE /projects/:id/issues/:issueId` dengan `RequirePermission(ResourceIssues, ...)`; audit `issue.created`/`updated`/`deleted`. Frontend: types + `projectService.listIssues/getIssue/createIssue/updateIssue/deleteIssue`; UI project detail: tombol Issue, form create/edit (severity, escalation, due date, resolution), panel Issues dengan FSM move + edit/delete + states. Verifikasi: `go build ./...` + `go test ./...` OK, `npm run type-check` + `npm run build` OK, smoke API issue lifecycle PASS, smoke UI PASS, regression smoke P0-013 33/33 PASS.
- `PMO-P1-002` Done: Risk management end-to-end. Backend: model Risk + `organization_id` (migration `000009_risks_tenant_score`: probability/impact INT 1-5, risk_score, severity, FK + index); repo/service org-scoped + project-scoped; `risk_score`/`severity` selalu dihitung backend (probability × impact); FSM risk (IDENTIFIED→ASSESSED→MITIGATED|ACCEPTED|ESCALATED; MITIGATED→CLOSED; ACCEPTED→CLOSED; ESCALATED→MITIGATED); routes `GET/POST /projects/:id/risks` + `GET/PUT/DELETE /projects/:id/risks/:riskID` dengan `RequirePermission(ResourceRisks, ...)`; audit `risk.created`/`updated`/`deleted`. Dashboard `/api/v1/dashboard` menambah `RISK_REGISTER` warnings dari risk register (score desc, open status, project aktif). Frontend: types + `projectService.listRisks/getRisk/createRisk/updateRisk/deleteRisk`; UI project detail: tombol Risk, form create/edit (probability/impact/due date/description/mitigation), panel Risks dengan score + severity badge + FSM move + edit/delete + loading/error/empty state; dashboard "Risiko Utama" & Command Center "Risiko Prioritas" membaca risk register saat data tersedia, fallback early warning berlabel jelas. Verifikasi: `go build ./...` + `go test ./...` OK, `npm run type-check` + `npm run lint` + `npm run build` OK, smoke API risk lifecycle 19/19 PASS, smoke UI PASS (create → FSM → edit score recompute → delete → empty), regression smoke P0-013 33/33 PASS.
- `PMO-P1-003` Done: Budget monitoring end-to-end. Backend: model `ProjectBudget` + DTOs (`CreateBudgetRequest`/`UpdateBudgetRequest` dengan `*float64` untuk partial update); repo/service org-scoped via ownership parent project (`project_budgets` TIDAK punya `organization_id` — scope lewat `projects.organization_id`); `variance` (planned − actual), `usage_pct` (actual/planned ×100), dan `status` (NORMAL <80 / WATCH ≥80 / RISK ≥90 / OVERRUN ≥100) selalu dihitung backend (round 2); routes `GET/POST /projects/:id/budgets` + `GET/PUT/DELETE /projects/:id/budgets/:budgetID` dengan `RequirePermission(ResourceBudgets, ...)`; audit `budget.created`/`updated`/`deleted`. Dashboard `BUDGET_THRESHOLD` warning agregat per project (`pb.deleted_at IS NULL` + `p.deleted_at IS NULL`), message menegaskan "operational input, not yet validated". Frontend: types (`BudgetLine`, `BudgetStatus`, request/response) + `projectService.listBudgets/getBudget/createBudget/updateBudget/deleteBudget`; UI project detail: tombol Budget, form create/edit (category/planned/actual/currency/description), panel Budget dengan ringkasan Total Planned/Actual/Variance/Usage + badge status + edit/delete + loading/error/empty state + footer note menunggu P1-011/P1-012; dashboard "Anggaran Berisiko" & Command Center membaca data budget. Verifikasi: `go build ./...` + `go test ./...` OK, `npm run type-check` + `npm run lint` + `npm run build` OK, smoke API budget lifecycle 17/17 PASS + boundary 4/4 PASS (WATCH @80, OVERRUN @100, negatif ditolak, currency uppercase) + audit count via SQL, smoke UI PASS (create → edit recompute RISK → delete → empty), 0 orphan budget aktif di project soft-deleted.
- `PMO-P1-004` Done: vendor/contract end-to-end (verified 2026-08-26). Backend: `vendors` master tenant-scoped (type `VENDOR`/`CONSULTANT`) + `contracts` (`organization_id` + parent `projects.organization_id`, partial unique `uq_contracts_org_number`); routes `GET/POST /vendors` + `GET/PUT/DELETE /vendors/:vendorID` (`RequirePermission(ResourceVendors, ...)`), `GET/POST /projects/:id/contracts` + `GET/PUT/DELETE /projects/:id/contracts/:contractID` (`RequirePermission(ResourceContracts, ...)`); ownership selalu diverifikasi (contract di project salah → 404); vendor in-use tidak bisa dihapus (409); audit `vendor.*`/`contract.*`. Frontend: types + `projectService.listVendors/getVendor/createVendor/updateVendor/deleteVendor` + `listContracts/getContract/createContract/updateContract/deleteContract`; UI project detail: tombol Kontrak, form create/edit (vendor/consultant selector dari master, value, currency, signed/start/end, status, scope), panel Kontrak (Total Value / Active / Period / Main Vendor Value + badge status + edit/delete + footer note menunggu P1-011/P1-012). Verifikasi: build backend + type-check/lint/build frontend OK; smoke API 16 PASS; smoke UI PASS.
- `PMO-P1-005` Done: document management end-to-end (verified 2026-08-26). Backend: `project_documents` metadata (name, category CONTRACT/REPORT/EVIDENCE/PHOTO/BAST/TOR_KAK/OTHER, version, file_size, mime_type, uploaded_by, file_url RELATIVE key; migration `000011_documents_version` adds `version`+`updated_at`); storage Phase 1 local disk `backend/storage/documents` (env `STORAGE_LOCAL_PATH`/`STORAGE_MAX_SIZE_BYTES`; helper `document_storage.go` MinIO/S3 swap-ready); routes `GET/POST /projects/:id/documents` + `GET/PUT/DELETE /projects/:id/documents/:documentID` + `GET /:documentID/download` (`RequirePermission(ResourceDocuments, ...)`); upload multipart, extension allowlist + magic bytes, oversize → 400; delete = soft row + hapus file fisik (path-escape guard); audit `document.uploaded/updated/downloaded/deleted` (orgID dari claims). Frontend: types (`ProjectDocument`, `DocumentCategory`, `UploadDocumentRequest`/`UpdateDocumentRequest`) + `projectService.listDocuments/getDocument/uploadDocument/updateDocument/deleteDocument/downloadDocument`; UI project detail: tombol Dokumen, form upload (file + name + category selector + version), panel Dokumen (Total Files / Total Size, badge kategori, download blob + edit metadata + delete confirm, loading/error/empty state, footer note menunggu P1-012). Verifikasi: `go build ./...` + `go test ./...` OK, migration 11 applied, smoke API lifecycle + edge 18/18 PASS, smoke UI PASS (upload → download → edit → delete → empty), 0 file fisik tersisa, audit count via SQL.

### Go Binary Location
```bash
/opt/homebrew/bin/go   # bukan /usr/local/go/bin/go
go version go1.26.2 darwin/arm64
```

### Backend Build Status — 2026-08-14 Historical Clean

**`go mod tidy`**: ✅ Berhasil
**`go build ./...`**: ✅ EXIT:0 — BERSIH, zero errors

#### Bug yang Sudah Diperbaiki (jangan duplikat fix ini):
File-file berikut punya stray `package main` atau duplicate package declaration. Semua sudah di-fix:
- `internal/core/rbac/model.go` → `package rbac`
- `internal/platform/server/server.go` → `package server`
- `internal/shared/utils/crypto.go` → `package utils`
- `internal/shared/types/pagination.go` → `package types`
- `internal/shared/types/flextime.go` → `package types`
- `internal/platform/middleware/auth.go` → `package middleware`
- `internal/platform/middleware/logger.go` → `package middleware`
- `internal/platform/middleware/rbac.go` → `package middleware`
- `internal/platform/middleware/requestid.go` → `package middleware`
- `internal/modules/project/model.go` → `package project`
- `internal/modules/project/service.go` → `package project`
- `internal/core/user/service.go` → `package user`
- `cmd/migrate/main.go` → duplicate `package main` dihapus
- `internal/core/auth/model.go` → duplicate `package auth` dihapus
- `internal/core/auth/service.go` → duplicate `package auth` dihapus
- `internal/core/auth/repository.go` → duplicate `package auth` dihapus
- `internal/core/auth/token_service.go` → duplicate `package auth` dihapus
- `internal/core/audit/repository.go` → duplicate `package audit` dihapus
- `internal/shared/constants/constants.go` → duplicate `package constants` dihapus

#### Bug Logic yang Sudah Diperbaiki:
- `internal/modules/project/handler.go` line 62: argument order `types.NewPaginationMeta(total, page, pageSize)` → `(page, pageSize, total)` ✅
- `internal/shared/constants/constants.go`: tambah alias plural (`ResourceProjects`, `ResourceTasks`, dll) yang dipakai `cmd/seed/main.go` ✅

### Frontend Build Status — 2026-08-14 Historical Clean

**`npm install`**: ✅ Berhasil (499 packages)
**`npm run build`**: ✅ EXIT:0 — BERSIH, zero errors

Routes ter-generate:
- `/` (redirect ke dashboard), `/login`, `/dashboard`, `/projects`, `/_not-found`
- Middleware: 26.9 kB

#### Bug yang Sudah Diperbaiki (frontend):
- `package.json`: hapus `@radix-ui/react-badge@1.0.0` — package ini tidak ada di npm registry
- `next.config.ts` → rename ke `next.config.mjs` + hapus TypeScript import type (Next.js 14 tidak support `.ts` config)
- `postcss.config.js`: `export default` → `module.exports` (Next.js 14 expects CommonJS)
- `.eslintrc.json`: hapus `"next/typescript"` extend — tidak tersedia di Next.js 14
- `src/app/(auth)/login/page.tsx` line 59: escape tanda kutip `"` → `&ldquo;` dan `&rdquo;`

### Current Operational Queue — 2026-08-26

Use `docs/DEVELOPMENT-BACKLOG.md` as the executable task queue.

Immediate order:
1. P0 hardening complete (`PMO-P0-014` Done): project-child soft delete cascade, dashboard exclusion, smoke cleanup, orphan soft cleanup — all verified.
2. `PMO-P1-001` (Issue management) Done & verified.
3. `PMO-P1-016` visual/runtime stabilization Done & verified (2026-08-25).
4. `PMO-P1-002` Risk Management Done & verified (2026-08-25): tenant-scoped register, probability/impact, score/severity recompute, mitigation, owner, lifecycle, dashboard/Command Center source mapping.
5. `PMO-P1-003` Budget Monitoring Done & verified (2026-08-25): budget line CRUD tenant-safe, variance/usage_pct/status backend-computed, dashboard BUDGET_THRESHOLD warning reads active lines only (`pb.deleted_at IS NULL`), audit `budget.created/updated/deleted`.
6. `PMO-P1-004` vendor/contract Done & verified (2026-08-26): vendor/consultant master tenant-scoped, contract CRUD end-to-end di project detail, validation (enum, unique per org, nilai/date), audit, dan smoke API/UI lulus. Contract value belum otomatis mengubah `project.budget_total` (menunggu P1-011 snapshot/validasi).
7. `PMO-P1-005` document management Done & verified (2026-08-26): upload/list/get/download/update/delete soft metadata; local disk storage `backend/storage/documents` (RELATIVE key di DB, MinIO/S3 swap-ready); extension allowlist + magic bytes + ukuran; delete menghapus record + file fisik; audit `document.uploaded/updated/downloaded/deleted`; frontend panel "Dokumen" di project detail (summary totals, upload form, list, download/edit/delete, states, footer note menunggu P1-012).
8. Next: `PMO-P1-006` corrective action (deviation, root cause, recommendation, PIC, target date, follow-up status, evidence).
9. Continue P1 core process and data foundation according to ticket dependency.
10. Build validated progress snapshots before full health score and Control Tower aggregate.
11. Build Level 3, then Level 2, then Level 1/Power BI/GIS.
12. Keep SRS, architecture, ERD, permission, gap analysis, backlog, and this file synchronized.

Baseline local verification:
```bash
# Setup env
cp '/Users/harmanto/Documents/Code/DEV/Project Management/PMO/backend/.env.example' \
   '/Users/harmanto/Documents/Code/DEV/Project Management/PMO/backend/.env'
# Edit .env: DB_HOST=localhost, DB_PASSWORD sesuai setup lokal

cd '/Users/harmanto/Documents/Code/DEV/Project Management/PMO/backend'

# Jalankan migrasi
make migrate-up

# Seed data awal (admin@cankora.local / Admin@Cankora2024! + 7 roles)
make seed

# Jalankan server
make dev   # server di :8080

# Cek health
curl http://localhost:8080/health
# Expected: {"status":"ok","version":"0.1.0"}

# Test login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}'
```

---

## Changelog

### v0.1.0 (scaffold — 2026-08-14)
- Initial backend scaffold: platform layer, core modules, project module
- All 5 SQL migrations (organizations, users/auth, rbac, audit_logs, projects_core)
- Frontend Next.js 14 App Router scaffold (login, dashboard, projects pages)
- docker-compose.yml dengan PostgreSQL 15 + Redis + MinIO + backend + frontend
- backend/Dockerfile (multi-stage Go build)
- frontend/Dockerfile (Next.js standalone output)
- Root README.md, CLAUDE.md, .gitignore
- backend/Makefile, backend/.env.example
- **Historical issue resolved**: stray `package main` files were fixed; see Current Build Status.
