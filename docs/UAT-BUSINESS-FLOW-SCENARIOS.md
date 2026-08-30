# PMO — UAT Business Flow Scenarios
# End-to-End Business Process Validation

**Versi**: 1.0.0
**Tanggal**: 2026-08-29
**Ticket**: PMO-UAT-007
**Audience**: QA, stakeholder PMO, product owner, demo facilitator, tech lead

---

## 1. Tujuan Dokumen

Dokumen ini mendefinisikan alur bisnis end-to-end yang harus diuji selama UAT PMO sebagai PMO National Project Control Tower. Fokus validasi adalah memastikan proses PMO nyambung dari:

**Input Operasional → Validasi/Governance → Dashboard/Reporting → Audit/Notification**

Setiap scenario mencakup:
- Aktor/role yang terlibat
- Data yang dipakai
- Step-by-step flow UAT
- Expected result per step
- Endpoint/page yang terlibat
- Data source (operational vs validated vs official governed vs reporting/read-model)
- Audit events dan notifikasi yang harus muncul
- Known limitations

---

## 2. Data Source Boundaries

| Layer | Deskripsi | Contoh Endpoint/Page | Status |
|---|---|---|---|
| **Operational Input** | Data diinput langsung oleh user — belum divalidasi/dikunci | `/projects`, `/projects/:id/issues`, `/projects/:id/risks`, `/projects/:id/budgets`, `/projects/:id/contracts` | Live (tulis) |
| **Validated Snapshot** | Snapshot progress tervalidasi per periode — dibuat dari operational input | `/projects/:id/data-quality` | Partial (P2-003) |
| **Official Governed Data** | Data yang telah melalui workflow approval formal governance (APPROVED/LOCKED) | `/governance` | Live (P3-003) |
| **Reporting/Read-model** | Agregasi dari operational/governed data — read-only, tidak boleh dicampur | `/analytics/executive`, `/analytics/programs`, `/reports/analytics`, `/command-center` | Live (read-only) |
| **Audit Trail** | Immutable event log untuk semua aksi tulis | `/audit-logs` | Live (append-only) |
| **Notification** | Outbox notifikasi per user/channel | `/notifications` | Live (UAT-004) |

> **Rule penting**: Jangan campur lapisan tanpa disclaimer. Dashboard `/executive` dan `/programs` adalah **operational aggregate** (bukan official governed data). `/governance` adalah **official governed data**. `/reports/analytics` adalah **read-model aggregate**. Pisahkan label ini di UI.

---

## 3. Business Flow Scenarios

### Flow A — Project Setup & Onboarding

**Aktor**: PMO / Admin
**Halaman terlibat**: `/projects`, project detail
**Data source**: Operational Input

| Step | Action | Expected Result | Endpoint |
|---|---|---|---|
| A-1 | Login sebagai `admin@cankora.local` | Redirect ke `/dashboard`, token valid | `POST /auth/login` |
| A-2 | Buat proyek baru dengan `code` unik, `name`, `start_date`, `end_date`, `status=DRAFT` | HTTP 201, proyek tersimpan dengan `organization_id` dari token | `POST /projects` |
| A-3 | Transisi proyek ke PLANNING | HTTP 200, status berubah | `POST /projects/:id/transition` |
| A-4 | Transisi proyek ke ACTIVE | HTTP 200, status berubah | `POST /projects/:id/transition` |
| A-5 | Buat milestone pertama (title, due_date) | HTTP 201, milestone terikat ke proyek | `POST /projects/:id/milestones` |
| A-6 | Buat task (title, due_date, priority) | HTTP 201 | `POST /projects/:id/tasks` |
| A-7 | Coba buat proyek dengan `code` sama | HTTP 409 "Project code already in use" | `POST /projects` |
| A-8 | Cek dashboard menampilkan proyek baru | KPI updated, proyek muncul di list | `GET /dashboard` |

**Audit events**: `project.created`, `project.transition`, `milestone.created`, `task.created`
**Known limitation**: Progress snapshot formal belum otomatis (P1-011); `progress_pct` adalah input manual.

---

### Flow B — Vendor, Contract & Budget Input

**Aktor**: PMO / Project Manager
**Halaman terlibat**: Project detail (panel Kontrak, Budget)
**Data source**: Operational Input

| Step | Action | Expected Result | Endpoint |
|---|---|---|---|
| B-1 | Buat vendor baru (`type=VENDOR`, name, tax_id) | HTTP 201, vendor tersimpan dengan `organization_id` | `POST /vendors` |
| B-2 | Buat konsultan baru (`type=CONSULTANT`) | HTTP 201 | `POST /vendors` |
| B-3 | Buat contract untuk proyek (contract_number unik, vendor_id, value, currency=IDR) | HTTP 201 | `POST /projects/:id/contracts` |
| B-4 | Coba buat contract dengan number sama di org yang sama | HTTP 409 unique constraint | `POST /projects/:id/contracts` |
| B-5 | Buat budget line (category=KONSTRUKSI, planned=5000000, actual=0, currency=IDR) | HTTP 201; `variance`=5000000, `usage_pct`=0, `status`=NORMAL | `POST /projects/:id/budgets` |
| B-6 | Update actual budget (actual=4500000) | `usage_pct`=90, `status`=RISK | `PUT /projects/:id/budgets/:id` |
| B-7 | Update actual budget (actual=5100000) | `usage_pct`=102, `status`=OVERRUN | `PUT /projects/:id/budgets/:id` |
| B-8 | Coba hapus vendor yang masih direferensikan contract | HTTP 409 "vendor is referenced by contracts" | `DELETE /vendors/:id` |
| B-9 | Dashboard menampilkan BUDGET_THRESHOLD warning | Warning muncul dengan status RISK/OVERRUN | `GET /dashboard` |

**Audit events**: `vendor.created`, `contract.created`, `budget.created`, `budget.updated`
**Known limitation**: Contract value belum otomatis mengubah `project.budget_total` (menunggu P1-011/P1-012).

---

### Flow C — Risk & Issue Register

**Aktor**: Project Manager / Project Officer
**Halaman terlibat**: Project detail (panel Risk, Issue)
**Data source**: Operational Input

| Step | Action | Expected Result | Endpoint |
|---|---|---|---|
| C-1 | Buat risk (title, probability=4, impact=5, status=IDENTIFIED) | HTTP 201; `risk_score`=20, `severity`=CRITICAL (backend computed) | `POST /projects/:id/risks` |
| C-2 | Transisi risk ke ASSESSED | HTTP 200 | `POST /projects/:id/risks/:id/transition` |
| C-3 | Transisi risk ke MITIGATED | HTTP 200 | `POST /projects/:id/risks/:id/transition` |
| C-4 | Buat issue (title, severity=HIGH, escalation=PROJECT_MANAGER) | HTTP 201 | `POST /projects/:id/issues` |
| C-5 | Transisi issue ke IN_PROGRESS | HTTP 200 | `POST /projects/:id/issues/:id/transition` |
| C-6 | Transisi issue ke RESOLVED (dengan `resolution` text) | HTTP 200 | `POST /projects/:id/issues/:id/transition` |
| C-7 | Dashboard menampilkan RISK_REGISTER warning | Risk CRITICAL muncul di warnings | `GET /dashboard` |
| C-8 | Command Center menampilkan risk watchlist | Risk prioritas muncul | `GET /command-center` |

**Audit events**: `risk.created`, `risk.updated`, `issue.created`, `issue.updated`
**Known limitation**: Risk mitigation effectiveness tidak terukur secara formal (menunggu P1-014 health score).

---

### Flow D — Corrective Action

**Aktor**: Project Manager / PMO
**Halaman terlibat**: Project detail (panel Corrective Action)
**Data source**: Operational Input

| Step | Action | Expected Result | Endpoint |
|---|---|---|---|
| D-1 | Buat corrective action linked ke issue (title, deviation, root_cause, recommendation, source_type=ISSUE, source_issue_id) | HTTP 201 | `POST /projects/:id/corrective-actions` |
| D-2 | Update corrective action (recommendation text, target_date) | HTTP 200 | `PUT /projects/:id/corrective-actions/:caID` |
| D-3 | Transisi corrective action ke IN_PROGRESS | HTTP 200 | `POST /projects/:id/corrective-actions/:caID/transition` |
| D-4 | List corrective actions per project | HTTP 200, list returned | `GET /projects/:id/corrective-actions` |
| D-5 | Command Center menampilkan follow-up corrective action | Corrective actions muncul di follow-up section | `GET /command-center` |

**Audit events**: `corrective_action.created`, `corrective_action.updated`, `corrective_action.transition`
**Known limitation**: Corrective action belum punya dedicated notification trigger (akan diimplementasikan saat notification trigger rules ditambahkan).

---

### Flow E — Field Evidence & Document

**Aktor**: Project Officer / Project Manager
**Halaman terlibat**: `/projects/:id/inspections`, project detail (panel Dokumen)
**Data source**: Operational Input

| Step | Action | Expected Result | Endpoint |
|---|---|---|---|
| E-1 | Upload dokumen (file txt/pdf, category=EVIDENCE) | HTTP 201, metadata tersimpan, file di storage lokal | `POST /projects/:id/documents` |
| E-2 | Download dokumen | HTTP 200, file binary + correct Content-Type | `GET /projects/:id/documents/:id/download` |
| E-3 | Update metadata dokumen | HTTP 200 | `PUT /projects/:id/documents/:id` |
| E-4 | Buat field inspection (scheduled_date, notes, location) | HTTP 201 | `POST /projects/:id/inspections` |
| E-5 | Upload evidence ke inspeksi (file, optional lat/lon) | HTTP 201 | `POST /projects/:id/inspections/:inspectionID/evidence` |
| E-6 | Download evidence dari inspeksi | HTTP 200 | `GET /projects/:id/inspections/:inspectionID/evidence/:evidenceID/download` |
| E-7 | Verifikasi inspeksi (VERIFIED) | HTTP 200 | `PATCH /projects/:id/inspections/:id/verification` |
| E-8 | Cross-tenant: coba akses inspeksi org lain | HTTP 404 | `GET /projects/:id/inspections` |

**Audit events**: `document.uploaded`, `document.downloaded`, `field_inspection.created`, `field_inspection.evidence_added`, `field_inspection.verified`
**Known limitation**: Attachment ke governance submission manual (user harus ambil file_url dari evidence, lalu input ke submission item).

---

### Flow F — Governance Official Approval

**Aktor**: PMO / Admin
**Halaman terlibat**: `/governance`
**Data source**: Official Governed Data

| Step | Action | Expected Result | Endpoint |
|---|---|---|---|
| F-1 | Buat submission governance (dataset_type=project, period_year=2026) | HTTP 201, status=DRAFT | `POST /governance/submissions` |
| F-2 | Tambah submission item (entity_type=project, entity_id valid, action=CREATE) | HTTP 200/201, item masuk submission | via submission create payload |
| F-3 | Submit (DRAFT → SUBMITTED) | HTTP 200 | `POST /governance/submissions/:id/submit` |
| F-4 | Start review (SUBMITTED → IN_REVIEW) | HTTP 200 | `POST /governance/submissions/:id/review` |
| F-5 | Approve (IN_REVIEW → APPROVED) | HTTP 200; semua item VALID | `POST /governance/submissions/:id/approve` |
| F-6 | Lock (APPROVED → LOCKED) | HTTP 200; lock period dibuat | `POST /governance/submissions/:id/lock` |
| F-7 | Coba buat submission baru untuk periode yang sama setelah LOCKED | HTTP 409 "period is locked" | `POST /governance/submissions` |
| F-8 | Reject scenario: start baru, submit, review, reject dengan alasan | HTTP 200, status=REJECTED, rejection_reason wajib | `POST /governance/submissions/:id/reject` |

**Audit events**: `governance.submission.created`, `governance.submission.submitted`, `governance.submission.review_started`, `governance.submission.approved`, `governance.submission.locked`, `governance.submission.rejected`
**Known limitation**: Governance belum mengirim notifikasi otomatis ke reviewer (notification trigger rules belum terhubung ke governance FSM).

---

### Flow G — Command Center & Escalation

**Aktor**: PMO / Admin
**Halaman terlibat**: `/command-center`
**Data source**: Operational current-state (bukan official governed data)

| Step | Action | Expected Result | Endpoint |
|---|---|---|---|
| G-1 | Baca command center aggregation | HTTP 200; risk watchlist, corrective follow-up, validation SLA, escalations, decisions | `GET /command-center` |
| G-2 | Buat escalation (project_id, description, urgency) | HTTP 201 | `POST /command-center/escalations` |
| G-3 | Update status escalation ke ACKNOWLEDGED | HTTP 200 | `PATCH /command-center/escalations/:id/status` |
| G-4 | Buat decision (project_id, decision_text, decided_by) | HTTP 201 | `POST /command-center/decisions` |
| G-5 | Update status decision ke DECIDED | HTTP 200 | `PATCH /command-center/decisions/:id/status` |
| G-6 | List escalations | HTTP 200 | `GET /command-center/escalations` |

**Audit events**: `command_escalation.created`, `command_escalation.acknowledged`, `executive_decision.created`, `executive_decision.decided`
**Known limitation**: Command Center adalah operational current-state. Data belum dari official governed snapshot.

---

### Flow H — Analytics & Reporting Dashboard

**Aktor**: PMO / Executive Viewer
**Halaman terlibat**: `/executive`, `/programs`, `/reports/analytics`
**Data source**: Reporting/Read-model aggregate (operational aggregate, bukan official governed)

| Step | Action | Expected Result | Endpoint |
|---|---|---|---|
| H-1 | Baca executive dashboard | HTTP 200; national summary (total/active projects, budget, health distribution, risks, issues) | `GET /analytics/executive` |
| H-2 | Baca program dashboard | HTTP 200; KPI per program/sektor | `GET /analytics/programs` |
| H-3 | Baca GIS summary | HTTP 200; project coordinates | `GET /analytics/gis/summary` |
| H-4 | Baca reporting catalog | HTTP 200; list dataset available | `GET /analytics/reports/catalog` |
| H-5 | Request report export (dataset_key=executive-summary, format=CSV) | HTTP 201; export request created, status=COMPLETED setelah diproses | `POST /analytics/reports/export/request` |
| H-6 | Download export file | HTTP 200; file binary (CSV) | `GET /analytics/reports/export/requests/:id/download` |
| H-7 | Cek audit log export | audit event `report.export.requested` + `report.export.completed` | `GET /audit-logs` |

**Audit events**: `report.export.requested`, `report.export.completed`, `report.export.downloaded`
**Known limitation**: Dashboard adalah operational aggregate — belum dari official governed snapshot. Power BI integration masih mock.

---

### Flow I — Audit Log Viewer

**Aktor**: Admin / PMO / Auditor
**Halaman terlibat**: `/audit-logs`
**Data source**: Audit Trail (append-only)

| Step | Action | Expected Result | Endpoint |
|---|---|---|---|
| I-1 | Baca audit log list | HTTP 200; paginated, tenant-scoped | `GET /audit-logs` |
| I-2 | Filter audit log by action (project.created) | HTTP 200; filtered result | `GET /audit-logs?action=project.created` |
| I-3 | Filter audit log by entity_type | HTTP 200 | `GET /audit-logs?entity_type=project` |
| I-4 | Search audit log (actor email) | HTTP 200 | `GET /audit-logs?search=admin@cankora` |
| I-5 | Baca audit log summary | HTTP 200; total events, unique actors, top actions/entities | `GET /audit-logs/summary` |
| I-6 | Export audit log CSV | HTTP 200; file CSV max 1000 rows | `GET /audit-logs/export` |
| I-7 | Coba akses tanpa token | HTTP 401 | `GET /audit-logs` |
| I-8 | Coba akses sebagai EXECUTIVE_VIEWER | HTTP 403 (bukan di resource list) | `GET /audit-logs` |
| I-9 | Akses sebagai AUDITOR | HTTP 200 (ada di resource list) | `GET /audit-logs` |

**Known limitation**: Audit log tidak bisa diedit/dihapus (immutable by design).

---

### Flow J — Notification

**Aktor**: Semua role
**Halaman terlibat**: `/notifications`
**Data source**: Notification outbox

| Step | Action | Expected Result | Endpoint |
|---|---|---|---|
| J-1 | Kirim test notification ke diri sendiri | HTTP 201; notifikasi tersimpan di outbox | `POST /notifications/test` |
| J-2 | List notifikasi | HTTP 200; hanya notifikasi milik user (recipient_user_id = caller) | `GET /notifications` |
| J-3 | Baca summary notifikasi | HTTP 200; total/unread/pending/failed | `GET /notifications/summary` |
| J-4 | Mark notifikasi sebagai read | HTTP 200 | `PATCH /notifications/:id/read` |
| J-5 | Mark all as read | HTTP 200 | `PATCH /notifications/read-all` |
| J-6 | Retry notifikasi FAILED | HTTP 200 | `POST /notifications/:id/retry` |
| J-7 | Admin list semua notifikasi org | HTTP 200 (admin/PMO only) | `GET /notifications/admin` |

**Known limitation**: Notification tidak otomatis ter-trigger oleh business events (governance approval, escalation create, dll.) — masih manual via test endpoint atau future integration.

---

### Flow K — Cross-Cutting: Permission & Tenant Guard

**Aktor**: Semua demo roles
**Tujuan**: Pastikan tidak ada cross-tenant data leak dan permission guard aktif

| Step | Role | Action | Expected |
|---|---|---|---|
| K-1 | officer@cankora.local | GET /audit-logs | 403 (PROJECT_OFFICER tidak punya audit_logs permission) |
| K-2 | executive@cankora.local | POST /governance/submissions | 403 (EXECUTIVE_VIEWER view+export only) |
| K-3 | auditor@cankora.local | GET /audit-logs | 200 (AUDITOR punya audit_logs permission) |
| K-4 | pmo@cankora.local | GET /command-center | 200 (PMO punya semua resource) |
| K-5 | Tanpa token | GET /projects | 401 |
| K-6 | Token valid | GET /projects dari org lain (fake project ID) | 404 (tenant guard) |

---

## 4. End-to-End Flow Sequence (Alur Gabungan)

```
[A] Project Setup
    ↓
[B] Vendor + Contract + Budget
    ↓
[C] Risk + Issue Register
    ↓
[D] Corrective Action
    ↓
[E] Field Evidence + Document
    ↓
[F] Governance Official Approval
    ↓
[G] Command Center Escalation + Decision
    ↓
[H] Analytics + Report Export
    ↓
[I] Audit Log Verification
    ↓
[J] Notification Verification
    ↓
[K] Permission + Tenant Guard
```

---

## 5. Data yang Digunakan dalam UAT

Semua data UAT menggunakan prefix `UAT-BF-007` agar bisa diidentifikasi dan di-cleanup tanpa mengacaukan data demo seed.

| Entitas | Identifier UAT | Nilai |
|---|---|---|
| Project | code=`UAT-BF-007` | name=`Proyek UAT Business Flow 007` |
| Vendor | name=`UAT Vendor 007` | type=VENDOR |
| Contract | contract_number=`CNT-UAT-007` | value=10000000 |
| Budget line | category=`KONSTRUKSI` | planned=5000000 |
| Risk | title=`UAT Risk 007` | probability=4, impact=5 |
| Issue | title=`UAT Issue 007` | severity=HIGH |
| Corrective Action | title=`UAT Corrective 007` | source_type=ISSUE |
| Governance submission | dataset_type=`project`, period_year=9999 (cleanup-safe) | — |

---

## 6. Known Limitations (yang masih Implemented/Basic)

| Area | Status | Catatan |
|---|---|---|
| Progress snapshot formal | Basic/Partial | `progress_pct` adalah input manual; formal validated snapshot belum otomatis dari governance approval (menunggu P1-011/P1-012) |
| Health Score | Basic/Partial | Health score formula bisa dikonfigurasi (P1-014) tapi belum terhubung ke executive dashboard aggregate secara otomatis |
| Notification auto-trigger | Basic | Notifikasi hanya via manual test endpoint; belum terhubung ke business event (governance approve, escalation create, risk critical) |
| Corrective action notification | Missing | Tidak ada auto-notification saat corrective action dibuat/overdue |
| Lock period + budget write block | Implemented | Governance lock period memblokir create submission baru di periode sama; belum memblokir langsung penulisan tabel operasional |
| Official governed snapshot untuk dashboard | Not yet | `/executive` dan `/programs` masih operational aggregate, bukan dari official governed data; disclaimer ada di label |
| Power BI actual integration | Not yet | Embed config tersedia tapi Power BI workspace nyata belum terhubung |
| IoT telemetry | Future/Last | `PMO-P3-004` diparkirkan sebagai item terakhir |
| Assigned-only task access | Partial | PROJECT_OFFICER tidak memiliki filter "hanya task yang di-assign ke dia" |

---

## 7. Smoke Test Coverage

Smoke test untuk scenario ini ada di:
- `docs/smoke-test-uat-007-business-flow.sh` — end-to-end business flow (UAT-007)
- `docs/smoke-test-uat-006-role-permissions.sh` — role permission checks (65/65)
- `docs/smoke-test-rel-001-regression.sh` — regression gate (44/44)
- `docs/smoke-test-uat-001-ui-full.sh` — UI route accessibility

---

## 8. Tanggung Jawab per Aktor

| Role | Flow Utama | Akses Dilarang |
|---|---|---|
| admin / SUPER_ADMIN | Semua flow | — |
| pmo@cankora.local (PMO) | A-K semua | — |
| pm@cankora.local (PROJECT_MANAGER) | A, B, C, D, E, G partial | Audit logs, executive analytics |
| officer@cankora.local (PROJECT_OFFICER) | C partial, D partial, E | Audit logs, executive analytics, command center |
| executive@cankora.local (EXECUTIVE_VIEWER) | H, J | Write operations, audit logs, command center |
| auditor@cankora.local (AUDITOR) | I, H (reports) | Write operations, executive dashboard |
