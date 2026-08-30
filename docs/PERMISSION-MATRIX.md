# Permission Matrix
# CANKORA — Enterprise Operations Platform

**Versi**: 0.2.4  
**Tanggal**: 2026-08-29  
**Dibuat oleh**: Cankonix  
**Status**: Draft — target permission model

> **Planning note 2026-08-29 — CANKORA-DASH-003: BALAI column**
> Penambahan kolom `BALAI` di Dashboard "10 Proyek Prioritas" tidak membutuhkan resource/action RBAC baru. Data berasal dari `projects.org_unit_id` → `org_units.name`; akses mengikuti existing Project view dan Org Unit view. Jika nama org unit ikut di-embed pada response project, tetap tenant-scoped dan read-only.

> **Update 2026-08-29 — CANKORA-DASH-002: ResourcePeriodicReport**
> `ResourcePeriodicReport` (`periodic_report`) ditambahkan ke `resources` slice di `cmd/seed/main.go` dan ke `constants.go`. Endpoints `GET/POST/GET/:id/PUT/:id/DELETE/:id` di `/api/v1/projects/:id/periodic-reports` dilindungi `RequirePermission(ResourcePeriodicReport, ActionView/Create/Update/Delete)`. Tenant guard via parent project ownership (`organization_id` + `project_id`). Duplicate period → 409. Soft delete → 204. Role yang mendapat akses secara default (via AllPerms/broad-resource): SUPER_ADMIN, ADMIN, PMO (all), PROJECT_MANAGER (semua action terkait project), PROJECT_OFFICER (view + create + update). EXECUTIVE_VIEWER dan AUDITOR hanya view.

> **Update 2026-08-29 — UAT-007/010 Corrective Action Permission & Final Acceptance**
> `ResourceCorrectiveAction` (`corrective_action`) ditambahkan ke `resources` slice di `cmd/seed/main.go` (UAT-007). ADMIN/PMO/SUPER_ADMIN kini mendapat permission `corrective_action:view/create/update/delete` di DB. Project Manager tidak mendapat corrective_action permission secara default. CA transition endpoint `POST /projects/:id/corrective-actions/:caID/transition` menggunakan `RequirePermission(ResourceCorrectiveActions, ActionUpdate)` — sesuai server.go. Status UAT-010: **Non-IoT UAT Candidate**. Semua permission resources terverifikasi via smoke UAT-007 86/86 PASS + UAT-006 65/65 PASS.

> **Update 2026-08-29 — UAT-006 Role-Based Permission Hardening**
> (1) Spatial routes `/sectors`, `/regions`, `/river-basins` kini dilindungi `RequirePermission` per-route di `server.go` (sebelumnya hanya `AuthRequired`). Action granular: view/create/update/delete sesuai resource. (2) `ResourceSector`, `ResourceRegion`, `ResourceRiverBasin` ditambahkan ke `resources` slice di `cmd/seed/main.go` — ADMIN/PMO sekarang mendapat permission spatial di DB. (3) Demo users seeded idempotent: `pmo@cankora.local`, `pm@cankora.local`, `officer@cankora.local`, `executive@cankora.local`, `auditor@cankora.local` (semua `Demo@Cankora2024!`). (4) Frontend `Sidebar.tsx`: `NavItem.roles` filter via `hasRole` (UX-only; backend adalah sumber kebenaran authorization). (5) Smoke `docs/smoke-test-uat-006-role-permissions.sh` **65/65 PASS** memverifikasi: no-token→401, fake-token→401, OFFICER GET /audit-logs→403, EXECUTIVE POST /governance→403, OFFICER POST /sectors→403, EXECUTIVE GET /sectors→403, PMO GET semua resource→200, AUDITOR GET /audit-logs→200.  
> `ResourceNotification` seeded (UAT-004): semua role yang punya resource list mendapat `notification:view`. Enforcement: setiap user hanya melihat notifikasi miliknya (`recipient_user_id = caller.user_id`); admin/PMO dapat akses via `/notifications/admin` (semua dalam org). No-token → 401. Endpoint `POST /notifications/test` hanya untuk admin/PMO (permission view sufficient; untuk prod pertimbangkan permission create terpisah). SMTP delivery aktif jika `SMTP_HOST` dikonfigurasi; jika tidak, `NoopProvider` mencatat log tanpa crash — tidak ada downtime akibat SMTP kosong.

> **Update 2026-08-29 — UAT-003 Audit Log Viewer Enforcement**  
> `ResourceAuditLog` (alias `ResourceAuditLogs`) sudah ada di seed sejak awal. UAT-003 menambah read-only API `/api/v1/audit-logs` yang sepenuhnya dijaga oleh `RequirePermission(ResourceAuditLogs, ActionView)`. Role yang berhak: SUPER_ADMIN, ADMIN, PMO (via AllPerms/all resources), AUDITOR (view+export only pada ResourceAuditLogs). EXECUTIVE_VIEWER **tidak** mendapat audit_log secara default (tidak ada di resource list-nya). PROJECT_MANAGER dan PROJECT_OFFICER tidak mendapat akses audit log. Export CSV (`/audit-logs/export`) menggunakan `ActionView` (bukan ActionExport terpisah) karena export ringan inline.

> **Update 2026-08-29 — REL-001 Governance Permission Audit**  
> `ResourceDataGovernance` seeded (P3-003): create/submit/review/approve/reject/lock/cancel submission; create/lock lock-period. Rules: hanya MATCHED gov mapping valid di submission item. UPDATE/DELETE/UPSERT item wajib `entity_id`; CREATE & VALIDATE_ONLY boleh tanpa. RBAC guard aktif di semua 12 route `/api/v1/governance/`. No-token request → 401 (verified di regression smoke REL-001).

> **Update 2026-08-21 — Command Center UI Access**  
> Menu dan route `/command-center` kini memakai backend permission `ResourceCommandCenter`, tenant-scoped aggregate API, persisted escalation, executive decision follow-up, dan audit. Sidebar visibility tetap bukan authorization; backend guard adalah sumber kebenaran.

---

## Keterangan

### Status Implementasi RBAC

| Area | Status Saat Ini | Catatan |
|------|-----------------|---------|
| Default roles seed | Verified (2026-08-20) | Tujuh role awal tersedia; admin memiliki `SUPER_ADMIN` setelah `make seed` |
| Default permissions seed | Verified (2026-08-20) | Resource/action matrix dasar dan seed idempotent terverifikasi |
| User role assignment | Implemented/Basic | Backend dan UI assignment `role_ids` tersedia; scope-management lanjutan belum lengkap |
| Permission middleware | Partial | Middleware ada, tetapi belum semua route menerapkan permission granular |
| Scope-based access | Conceptual/Partial | Scope model ada; filtering per role/scope belum lengkap di semua fitur |
| Project Officer restrictions | Missing/Partial | Assigned-only task access belum fully enforced |
| Executive dashboard scope | Partial | Dashboard masih aggregation dasar |
| PMO Command Center route | Implemented/Basic | Protected by auth plus `ResourceCommandCenter`; validation/action/decision guards tenant-scoped |
| Auditor read-only access | **Implemented** | `/audit-logs` API live (UAT-003): list, getById, summary, export CSV — all tenant-scoped, permission-guarded |

| Simbol | Makna |
|--------|-------|
| ✅ | Diizinkan |
| ❌ | Tidak diizinkan |
| ⚠️ | Terbatas (lihat catatan) |
| — | Tidak relevan |

**Actions**:
- **view** — Melihat/membaca data
- **create** — Membuat data baru
- **update** — Mengubah data yang ada
- **delete** — Menghapus (soft delete)
- **approve** — Menyetujui / menolak workflow
- **export** — Export data ke file (CSV, Excel, PDF)

**Default Roles** (seed awal — administrator bisa tambah role baru):

| Role | Kode | Deskripsi |
|------|------|-----------|
| Super Admin | `SUPER_ADMIN` | Akses penuh ke semua fitur dan konfigurasi sistem |
| Admin | `ADMIN` | Kelola user, role, organisasi. Tidak ada akses data bisnis |
| PMO | `PMO` | Monitor semua project di unit scope-nya |
| Project Manager | `PROJECT_MANAGER` | Kelola project yang di-assign |
| Project Officer | `PROJECT_OFFICER` | Update task & progress di project yang diikuti |
| Executive Viewer | `EXECUTIVE_VIEWER` | Dashboard & laporan read-only seluruh project |
| Auditor | `AUDITOR` | Read-only semua data + akses audit trail |

**Planned Roles untuk Control Tower**:

| Role | Kode | Deskripsi |
|------|------|-----------|
| Data Validator | `DATA_VALIDATOR` | Memeriksa completeness/freshness dan menerima atau menolak submission sesuai scope |
| Field Inspector | `FIELD_INSPECTOR` | Mencatat inspeksi dan evidence pada proyek yang di-assign |
| Integration Admin | `INTEGRATION_ADMIN` | Mengelola connector non-secret, mapping, import/sync run, dan refresh monitoring |

Role planned bersifat additive. `EXECUTIVE_VIEWER` mewakili pengguna Level 1, `PMO` mewakili Level 2, sedangkan `PROJECT_MANAGER`/`PROJECT_OFFICER` mewakili Level 3. Scope aktual tetap ditentukan oleh user scope, org unit, program, dan project assignment.

---

## 1. Core Platform — User Management

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **User** | view | ✅ | ✅ | ⚠️ unit | ⚠️ tim project | ❌ | ❌ | ✅ |
| **User** | create | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **User** | update | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **User** | delete | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **User** | export | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |

> ⚠️ PMO hanya bisa melihat user di org unit scope-nya  
> ⚠️ Project Manager hanya bisa melihat user yang ada di tim project-nya (untuk keperluan assign)

---

## 2. Core Platform — Organization Management

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Organization** | view | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Organization** | create | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Organization** | update | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Organization** | delete | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Org Unit** | view | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Org Unit** | create | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Org Unit** | update | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Org Unit** | delete | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

---

## 3. Core Platform — RBAC Management

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Role** | view | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **Role** | create | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Role** | update | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Role** | delete | ✅ | ⚠️ non-system | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Permission** | view | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **Permission** | create | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Permission** | update | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Group** | view | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **Group** | create | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Group** | update | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Group** | delete | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **User Role Assign** | create | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **User Role Assign** | delete | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

> ⚠️ Admin hanya bisa delete role yang bukan system role (`is_system = false`)  
> Permission definition hanya bisa diubah oleh Super Admin (perubahan sangat berisiko)

---

## 4. Core Platform — Audit Trail

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Audit Log** | view | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **Audit Log** | export | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **Audit Log** | create | — | — | — | — | — | — | — |
| **Audit Log** | update | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Audit Log** | delete | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

> Audit log ditulis otomatis oleh sistem (tidak bisa manual create/update/delete oleh siapapun)

---

## 5. Project Management — Project

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Project** | view | ✅ | ❌ | ⚠️ unit scope | ⚠️ assigned | ⚠️ member | ✅ | ✅ |
| **Project** | create | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Project** | update | ✅ | ❌ | ⚠️ unit scope | ⚠️ assigned | ❌ | ❌ | ❌ |
| **Project** | delete | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Project** | approve | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Project** | export | ✅ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ |

> ⚠️ PMO: bisa lihat & update semua project di org unit scope-nya  
> ⚠️ Project Manager: hanya project yang dia jadi PM atau di-assign  
> ⚠️ Project Officer: hanya project di mana dia menjadi anggota tim

---

## 6. Project Management — Project Team

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Project Team** | view | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Project Team** | create | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Project Team** | update | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Project Team** | delete | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |

---

## 7. Project Management — Milestone

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Milestone** | view | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Milestone** | create | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Milestone** | update | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Milestone** | delete | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Milestone** | approve | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |

---

## 8. Project Management — Task (WBS)

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Task** | view | ✅ | ❌ | ✅ | ✅ | ⚠️ assigned | ✅ | ✅ |
| **Task** | create | ✅ | ❌ | ✅ | ✅ | ⚠️ sub-task | ❌ | ❌ |
| **Task** | update | ✅ | ❌ | ✅ | ✅ | ⚠️ assigned | ❌ | ❌ |
| **Task** | delete | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Task** | assign | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Task Comment** | view | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ |
| **Task Comment** | create | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Task Comment** | update | ✅ | ❌ | ⚠️ own | ⚠️ own | ⚠️ own | ❌ | ❌ |
| **Task Comment** | delete | ✅ | ❌ | ✅ | ✅ | ⚠️ own | ❌ | ❌ |

> ⚠️ Project Officer (view task): hanya task yang di-assign ke dia  
> ⚠️ Project Officer (create task): hanya bisa buat sub-task di bawah task yang ada, bukan task level 1  
> ⚠️ Project Officer (update task): hanya task yang di-assign ke dia (update progress, status)  
> ⚠️ own: hanya bisa update/delete comment miliknya sendiri

---

## 9. Project Management — Issue

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Issue** | view | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Issue** | create | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Issue** | update | ✅ | ❌ | ✅ | ✅ | ⚠️ own/assigned | ❌ | ❌ |
| **Issue** | delete | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Issue** | approve | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |

> ⚠️ Project Officer: hanya bisa update issue yang dia buat atau dia menjadi PIC

---

## 10. Project Management — Risk

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Risk** | view | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Risk** | create | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Risk** | update | ✅ | ❌ | ✅ | ✅ | ⚠️ own/assigned | ❌ | ❌ |
| **Risk** | delete | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |

---

## 11. Project Management — Budget

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Budget** | view | ✅ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ |
| **Budget** | create | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Budget** | update | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Budget** | delete | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Budget** | export | ✅ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ |

> Budget tidak bisa dilihat oleh Project Officer (informasi finansial terbatas)

---

## 12. Project Management — Document

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Document** | view | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Document** | create | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Document** | update | ✅ | ❌ | ✅ | ✅ | ⚠️ own | ❌ | ❌ |
| **Document** | delete | ✅ | ❌ | ✅ | ✅ | ⚠️ own | ❌ | ❌ |
| **Document** | export | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## 13. Project Management — Meeting

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Meeting** | view | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ |
| **Meeting** | create | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Meeting** | update | ✅ | ❌ | ✅ | ✅ | ⚠️ own | ❌ | ❌ |
| **Meeting** | delete | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |

---

## 14. Project Management — Dashboard & Report

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Dashboard** | view | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Report** | view | ✅ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ |
| **Report** | export | ✅ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ |

> Dashboard yang dilihat masing-masing role disesuaikan dengan scope datanya:
> - Executive Viewer & PMO: semua project sesuai scope
> - Project Manager: hanya project yang dia manage
> - Project Officer: hanya project yang dia ikuti
> - Auditor: semua project (read-only)

---

## 15. Workflow & Approval

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|
| **Workflow Definition** | view | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| **Workflow Definition** | create | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Workflow Definition** | update | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Approval Request** | view | ✅ | ❌ | ✅ | ⚠️ own | ❌ | ❌ | ✅ |
| **Approval Request** | approve | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |

---

## 16. Matriks Ringkas — Semua Role

Ringkasan kapabilitas per role (✅ = punya akses, ❌ = tidak punya, ⚠️ = terbatas):

| Kapabilitas | Super Admin | Admin | PMO | PM | Officer | Exec Viewer | Auditor |
|------------|:-----------:|:-----:|:---:|:--:|:-------:|:-----------:|:-------:|
| Kelola User & Org | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Kelola Role & Permission | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Lihat Audit Trail | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Lihat semua Project | ✅ | ❌ | ⚠️ | ⚠️ | ⚠️ | ✅ | ✅ |
| Buat Project | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| Approve Project | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Kelola Task | ✅ | ❌ | ✅ | ✅ | ⚠️ | ❌ | ❌ |
| Update Progress Task | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Lihat Budget | ✅ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ |
| Kelola Budget | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| Upload Dokumen | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Export Laporan | ✅ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ |
| Executive Dashboard | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## 17. Scope Matrix — Batasan Data per Role

| Role | Scope Default | Data yang Bisa Diakses |
|------|--------------|----------------------|
| Super Admin | ALL | Semua data tanpa batasan |
| Admin | ALL | Semua data user/org/role (bukan data project) |
| Executive Viewer | ALL | Semua project (read-only) |
| Auditor | ALL | Semua data (read-only + audit trail) |
| PMO | ORG_UNIT | Project & data di org unit yang di-assign beserta turunannya |
| Project Manager | ASSIGNED_PROJECT | Project yang dia menjadi PM atau di-assign sebagai PM |
| Project Officer | MEMBER_PROJECT | Project di mana dia adalah anggota tim |

---

## 18. Catatan Implementasi

### 18.1 Effective Permission Calculation

```
Effective Permission = UNION(
    permissions dari direct user_roles,
    permissions dari group_roles via user_groups
)

Effective Scope = scope yang paling luas dari semua role yang dimiliki user
(ALL > ORG_UNIT > ASSIGNED_PROJECT > MEMBER_PROJECT)
```

### 18.2 Permission Check di Backend

```go
// Middleware check urutan:
// 1. Apakah JWT valid dan user aktif?
// 2. Apakah user memiliki permission resource.action?
// 3. Apakah resource yang diminta berada dalam scope user?

// Contoh:
router.GET("/projects/:id",
    auth.RequireAuth(),
    rbac.RequirePermission("project", "view"),
    rbac.RequireScope("project", ":id"),  // cek scope
    handler.GetProject,
)
```

### 18.3 Scope Resolution

```
Scope ALL       → query tanpa filter organization/project
Scope ORG_UNIT  → WHERE org_unit_id IN (unit + semua turunannya)
Scope ASSIGNED_PROJECT → WHERE pm_user_id = current_user_id
                          OR id IN (SELECT project_id FROM project_teams WHERE user_id = ...)
Scope MEMBER_PROJECT   → WHERE id IN (SELECT project_id FROM project_teams WHERE user_id = ...)
```

### 18.4 Super Admin Bypass

Super Admin **tidak diperiksa per-permission** — langsung diizinkan untuk semua resource dan action. Ini diimplementasikan di awal permission check middleware:

```go
if user.HasRole("SUPER_ADMIN") {
    return next(ctx)  // bypass semua permission check
}
```

### 18.5 Permission yang Dicatat di Audit Trail

Setiap perubahan role/permission **wajib** menghasilkan audit log:

| Event | Detail yang Dicatat |
|-------|---------------------|
| Role dibuat | Nama role, permission list |
| Role diubah | Permission yang ditambah/dihapus (diff) |
| Role di-assign ke user | User, role, scope |
| Role di-revoke dari user | User, role |
| Group diubah | Role yang ditambah/dihapus dari group |
| User masuk/keluar group | User, group |

---

## 19. PMO Control Tower - Dashboard Access

| Resource | Action | Super Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor | Data Validator | Field Inspector | Integration Admin |
|----------|--------|:-----------:|:---:|:---------------:|:---------------:|:----------------:|:-------:|:--------------:|:---------------:|:-----------------:|
| National Dashboard (Level 1) | view | ✅ | ⚠️ scope | ⚠️ assigned summary | ⚠️ member summary | ✅ | ✅ | ⚠️ quality-only | ❌ | ⚠️ refresh-only |
| Program Dashboard (Level 2) | view | ✅ | ✅ | ⚠️ assigned | ⚠️ member | ✅ | ✅ | ⚠️ quality-only | ❌ | ⚠️ refresh-only |
| Project Control (Level 3) | view | ✅ | ✅ | ⚠️ assigned | ⚠️ member | ✅ | ✅ | ⚠️ validation scope | ⚠️ assigned | ⚠️ integration diagnostics |
| Dashboard Dataset | export | ✅ | ✅ | ⚠️ assigned | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ |

`view` tidak berarti melihat seluruh tenant. Semua query, map layer, chart, export, dan Power BI RLS harus memakai effective organization/org-unit/program/project scope. Summary yang berpotensi mengungkap data proyek di luar scope juga dilarang.

## 20. Data Validation and Quality

| Resource | Action | Super Admin | PMO | Project Manager | Project Officer | Data Validator | Auditor | Integration Admin |
|----------|--------|:-----------:|:---:|:---------------:|:---------------:|:--------------:|:-------:|:-----------------:|
| Data Submission | view | ✅ | ✅ | ⚠️ assigned | ⚠️ own/member | ✅ | ✅ | ✅ |
| Data Submission | create | ✅ | ✅ | ✅ | ✅ | ⚠️ correction | ❌ | ✅ |
| Data Submission | update | ✅ | ⚠️ draft | ⚠️ own draft | ⚠️ own draft | ⚠️ correction | ❌ | ⚠️ import metadata |
| Data Submission | approve/reject | ✅ | ⚠️ policy | ❌ | ❌ | ✅ | ❌ | ❌ |
| Data Quality Result | view | ✅ | ✅ | ⚠️ assigned | ⚠️ member | ✅ | ✅ | ✅ |
| Validation Queue | view | ✅ | ✅ | ⚠️ assigned | ❌ | ✅ | ✅ | ⚠️ integration failures |

Creator dan validator tidak boleh menjadi orang yang sama bila segregation-of-duties policy mengharuskannya. Rejection reason wajib; published snapshot hanya memakai submission `VALID`.

## 20A. Official Data Governance (P3-003)

> **P3-003 Done** — `ResourceDataGovernance` (`data_governance`) tersedia di seed. 12 endpoint aktif di `/api/v1/governance/*`; audit `governance.*` tercatat pada tiap transisi; approval menyimpan actor + timestamp.

| Resource | Action | Super Admin | Admin | PMO | Data Validator | Project Manager | Project Officer | Auditor | Integration Admin |
|----------|--------|:-----------:|:-----:|:---:|:--------------:|:---------------:|:---------------:|:-------:|:-----------------:|
| Governance Submission | view | ✅ | ✅ | ✅ | ✅ | ⚠️ assigned | ⚠️ own/member | ✅ | ✅ |
| Governance Submission | create (DRAFT) | ✅ | ✅ | ✅ | ✅ | ⚠️ assigned | ⚠️ own | ❌ | ⚠️ ingest metadata |
| Governance Submission | submit (DRAFT→SUBMITTED) | ✅ | ✅ | ✅ | ✅ | ⚠️ own | ⚠️ own | ❌ | ⚠️ own |
| Governance Submission | review (SUBMITTED→IN_REVIEW) | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Governance Submission | approve (IN_REVIEW→APPROVED) | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Governance Submission | reject (IN_REVIEW→REJECTED) | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Governance Submission | lock (APPROVED→LOCKED) | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Governance Submission | cancel | ✅ | ✅ | ✅ | ✅ | ⚠️ own draft | ⚠️ own draft | ❌ | ⚠️ own |
| Lock Period | view | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Lock Period | create | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ⚠️ ingest config |
| Lock Period | lock (OPEN→LOCKED) | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |

Aturan kunci:

- Transisi FSM ketat: `DRAFT→SUBMITTED→IN_REVIEW→APPROVED→LOCKED`; `IN_REVIEW→REJECTED` (wajib `rejection_reason`, kembali ke DRAFT); `DRAFT/SUBMITTED→CANCELLED`. Transisi invalid → HTTP 409.
- Review/approve/reject/lock hanya oleh role dengan permission `data_governance` (approver dicatat di `approved_by`/`reviewed_by`/`locked_by` + timestamp).
- Per-item validation menolak entitas soft-deleted dan mapping pemerintah `PENDING_MATCH` (belum bisa jadi official).
- Lock period unik per (org, dataset, year, month); periode LOCKED memblokir submission baru untuk dataset+periode tersebut → 409.
- Data yang di-import/synced (CSV, Primavera, government, BIM) **tidak otomatis approved** — harus melewati submission resmi dan review manusia.
- Creator dan approver tidak boleh sama bila segregation-of-duties policy diaktifkan.
- Semua query tenant-scoped (`organization_id`); snapshot `LOCKED` menjadi dasar data official (immutable).

## 21. Health Score, Alert, and Decision

| Resource | Action | Super Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor | Data Validator |
|----------|--------|:-----------:|:---:|:---------------:|:---------------:|:----------------:|:-------:|:--------------:|
| Health Formula | view | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Health Formula | create/update | ✅ | ⚠️ draft | ❌ | ❌ | ❌ | ❌ | ❌ |
| Health Formula | approve/retire | ✅ | ⚠️ designated approver | ❌ | ❌ | ❌ | ❌ | ❌ |
| Health Snapshot | view | ✅ | ✅ | ⚠️ assigned | ⚠️ member | ✅ | ✅ | ✅ |
| Alert | view | ✅ | ✅ | ⚠️ assigned | ⚠️ member/PIC | ✅ | ✅ | ⚠️ quality alert |
| Alert | update/assign/escalate | ✅ | ✅ | ⚠️ assigned | ⚠️ PIC status | ❌ | ❌ | ⚠️ quality alert |
| Corrective Action | create/update | ✅ | ✅ | ⚠️ assigned | ⚠️ PIC | ❌ | ❌ | ⚠️ data correction |
| Executive Decision | view | ✅ | ✅ | ⚠️ assigned | ⚠️ actionable excerpt | ✅ | ✅ | ❌ |
| Executive Decision | create/approve | ✅ | ⚠️ secretariat/approver | ❌ | ❌ | ⚠️ designated decision maker | ❌ | ❌ |
| Executive Decision | update follow-up | ✅ | ✅ | ⚠️ owner | ⚠️ PIC | ❌ | ❌ | ❌ |

Formula approval, manual override, score recalculation, escalation, keputusan, dan penyelesaian corrective action wajib menghasilkan audit log dengan before/after atau reference snapshot.

## 22. Contract, Finance, GIS, Field, and Benefit

| Resource | Action | Super Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor | Field Inspector |
|----------|--------|:-----------:|:---:|:---------------:|:---------------:|:----------------:|:-------:|:---------------:|
| Contract/Party | view | ✅ | ✅ | ⚠️ assigned | ⚠️ limited | ✅ | ✅ | ⚠️ field context |
| Contract/Party | create/update | ✅ | ✅ | ⚠️ assigned | ❌ | ❌ | ❌ | ❌ |
| Financial Realization | view | ✅ | ✅ | ⚠️ assigned | ❌ | ✅ | ✅ | ❌ |
| Financial Realization | create/update | ✅ | ✅ | ⚠️ assigned | ❌ | ❌ | ❌ | ❌ |
| GIS Project Layer | view | ✅ | ✅ | ⚠️ assigned | ⚠️ member | ✅ | ✅ | ⚠️ assigned |
| Field Inspection | create | ✅ | ✅ | ✅ | ⚠️ assigned | ❌ | ❌ | ✅ |
| Field Inspection | validate | ✅ | ✅ | ⚠️ reviewer | ❌ | ❌ | ❌ | ❌ |
| Field Evidence | create/delete | ✅ | ✅ | ✅ | ⚠️ own | ❌ | ❌ | ⚠️ own |
| Benefit Indicator | view | ✅ | ✅ | ✅ | ⚠️ member | ✅ | ✅ | ⚠️ field context |
| Benefit Indicator | create/update | ✅ | ✅ | ⚠️ assigned | ❌ | ❌ | ❌ | ❌ |

Metadata lokasi sensitif dan evidence asli harus mengikuti classification policy. Delete evidence adalah soft delete metadata; object retention mengikuti kebijakan dokumen/audit.

## 23. Integration and Power BI

| Resource | Action | Super Admin | PMO | Executive Viewer | Auditor | Integration Admin |
|----------|--------|:-----------:|:---:|:----------------:|:-------:|:-----------------:|
| Integration Source | view | ✅ | ⚠️ status | ❌ | ✅ | ✅ |
| Integration Source | create/update/deactivate | ✅ | ❌ | ❌ | ❌ | ✅ |
| Integration Run | execute/retry/cancel | ✅ | ❌ | ❌ | ❌ | ✅ |
| Integration Run | view | ✅ | ⚠️ summary | ❌ | ✅ | ✅ |
| Field Mapping | create/update | ✅ | ⚠️ business mapping approval | ❌ | ✅ | ✅ |
| Power BI Dataset Contract | view | ✅ | ✅ | ✅ | ✅ | ✅ |
| Power BI Refresh | execute/retry | ✅ | ❌ | ❌ | ❌ | ✅ |
| Power BI Refresh | view | ✅ | ✅ | ⚠️ freshness only | ✅ | ✅ |

Secrets tidak pernah dikembalikan oleh API atau ditampilkan di UI. Integration admin hanya mengelola secret reference/rotation state melalui mekanisme secret store yang disetujui.

## 24. Data Import (CSV/Excel)

> **P2-001 Done** — `ResourceImport` (`import`) tersedia di seed. 8 endpoint aktif, audit trail tiap lifecycle event.

| Resource | Action | Super Admin | Admin | PMO | Project Manager | Project Officer | Executive Viewer | Auditor | Integration Admin |
|----------|--------|:-----------:|:-----:|:---:|:---------------:|:---------------:|:----------------:|:-------:|:-----------------:|
| Import Template | view | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ |
| Import Job | view | ✅ | ❌ | ✅ | ⚠️ own | ⚠️ own | ❌ | ✅ | ✅ |
| Import Job | create (upload) | ✅ | ❌ | ✅ | ✅ | ⚠️ limited types | ❌ | ❌ | ✅ |
| Import Job | validate | ✅ | ❌ | ✅ | ✅ | ⚠️ own | ❌ | ❌ | ✅ |
| Import Job | commit | ✅ | ❌ | ✅ | ⚠️ assigned | ❌ | ❌ | ❌ | ✅ |
| Import Job | cancel | ✅ | ❌ | ✅ | ⚠️ own | ⚠️ own | ❌ | ❌ | ✅ |
| Import Row | view | ✅ | ❌ | ✅ | ⚠️ own job | ⚠️ own job | ❌ | ✅ | ✅ |

Commit adalah operasi destructive-additive: data target yang sudah ada bisa diubah. Wajib ada review validate-then-commit; satu orang tidak boleh upload dan commit job yang sama bila segregation-of-duties policy diterapkan. Semua lifecycle event (upload, validate, commit, cancel) dicatat di audit log dengan `organization_id` dan `uploaded_by`.

---

## 25. Additional Scope Dimensions

Scope Control Tower menambah dimensi berikut pada model akses:

```text
NATIONAL       -> seluruh data dalam organization yang diizinkan
PROGRAM        -> program_id tertentu
SECTOR         -> sector_id tertentu
REGION         -> region_id dan turunannya
ORG_UNIT       -> org_unit_id dan turunannya
ASSIGNED_PROJECT / MEMBER_PROJECT
```

Jika user memiliki beberapa scope, hasil adalah union hanya setelah setiap scope diverifikasi masih dalam `organization_id` yang sama. `SUPER_ADMIN` bypass tidak menghapus kewajiban audit, dan tidak boleh diterapkan pada koneksi Power BI/service account tanpa kontrol khusus.
