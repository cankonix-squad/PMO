# Software Requirement Specification (SRS)
# CANKORA — Enterprise Operations Platform
## Modul: Project Management Monitoring & PMO National Project Control Tower

**Versi**: 0.2.3  
**Tanggal**: 2026-08-29  
**Dibuat oleh**: Cankonix  
**Status**: Draft — perlu sinkronisasi implementasi

> **Update 2026-08-29 — REL-001 Non-IoT Stabilization**  
> Regression gate terpadu `docs/smoke-test-rel-001-regression.sh` tersedia. Data governance official validation (P3-003) diperkuat: UPDATE/DELETE/UPSERT submission item wajib `entity_id`; CREATE & VALIDATE_ONLY boleh tanpa. Semua P0–P3 non-IoT stable; IoT (`CANKORA-P3-004`) tetap Future/Last. Sistem siap UAT/demo packaging.

> **Update 2026-08-18 — Implementation Reality Check**  
> Dokumen ini menggambarkan kebutuhan produk/MVP target untuk PMO. Kondisi kode saat ini belum memenuhi seluruh requirement. Gunakan [Implementation Gap Analysis](./IMPLEMENTATION-GAP-ANALYSIS.md) untuk melihat status `Implemented`, `Partial`, dan `Missing` sebelum menyatakan kebutuhan user sudah terpenuhi.

> **Update 2026-08-20 — P0 Stabilization Verified**  
> Semua tiket P0 (`CANKORA-P0-001` s.d. `CANKORA-P0-013`) sudah `Done` dan terverifikasi lokal end-to-end: migrate, seed, login admin, dashboard live, project CRUD + transition + progress history, task/milestone CRUD nested, user management, kanban task board, dan early warning dasar. Lihat `docs/DEVELOPMENT-BACKLOG.md` dan `CLAUDE.md` untuk status build/smoke terkini. Scope MVP inti (auth, RBAC awal, project monitoring dasar) sudah usable; gap terhadap Grand Design PMO tetap mengacu `IMPLEMENTATION-GAP-ANALYSIS.md`.

> **Update 2026-08-20 — PMO National Project Control Tower Baseline**  
> Materi terbaru memperluas target menjadi Control Tower Level 1 Nasional, Level 2 Program/Sektor, dan Level 3 Detail Proyek dengan strategi hybrid CANKORA + Power BI. Target ini masih roadmap dan tidak mengubah status P0 yang sudah terverifikasi. Lihat [PMO Control Tower Analysis](./PMO-CONTROL-TOWER-ANALYSIS.md).

> **Update 2026-08-21 — Native UI Shell dan Runtime Gate**  
> Login PMO, dashboard eksekutif native, data demo SDA, dan route `/command-center` sudah tersedia. Per 2026-08-26, `CANKORA-P1-015` sudah memenuhi lifecycle API/UI dasar untuk aggregate alert, validation SLA, corrective action follow-up, persisted escalation, executive decision follow-up, tenant guard, dan audit. GIS aktual, Level 1/2 resmi, Power BI, dan advanced integrations tetap mengikuti tiket P2/P3.

> **Update 2026-08-26 — Data Validation Queue**
> `CANKORA-P1-012` sudah terverifikasi: submission snapshot tenant-scoped memiliki status lifecycle, completeness, freshness, SLA aging, lineage, validator, rejection reason, configurable segregation-of-duties, audit, protected API, dan UI `/validation`. Health Score serta Control Tower tetap roadmap.

> **Update 2026-08-26 — Field Inspection and Evidence**
> `CANKORA-P1-013` sudah terverifikasi: inspeksi lapangan tenant/project-scoped menyimpan waktu, koordinat, petugas, catatan, verification status, evidence metadata, checksum SHA-256, soft delete, dan authorized download. GIS map penuh serta offline/mobile workflow tetap roadmap.

> **Update 2026-08-28 — Official Data Validation & Approval Workflow (P3-003)**
> `CANKORA-P3-003` Data Governance sudah terverifikasi: submission resmi (DRAFT→SUBMITTED→IN_REVIEW→APPROVED→LOCKED, +REJECTED dengan alasan wajib, CANCELLED) dengan item-level validation (tenant ownership, soft-delete, government `PENDING_MATCH` tidak bisa jadi official), lock period (OPEN/LOCKED) yang memblokir penulisan pada periode terkunci, audit `governance.*`, dan 12 route `/api/v1/governance/`. Resource RBAC `ResourceDataGovernance` tersedia; frontend `/governance` menampilkan list/filter, create submission modal dengan dynamic items, detail modal dengan status timeline + actions state-aware, dan lock period panel. **Planning update:** IoT telemetry tidak menjadi next ticket; P3-004 diparkir sebagai capability paling akhir/future setelah governance hardening, UI regression, read-model/reporting reconciliation, dan Phase 3 non-IoT stabil.

---

## Daftar Isi

1. [Business Objective](#1-business-objective)
2. [System Scope](#2-system-scope)
3. [Stakeholder](#3-stakeholder)
4. [Organization Hierarchy](#4-organization-hierarchy)
5. [User, Role & Permission Concept](#5-user-role--permission-concept)
6. [Project Management Business Process](#6-project-management-business-process)
7. [Functional Requirements — Core Platform](#7-functional-requirements--core-platform)
8. [Functional Requirements — Project Management](#8-functional-requirements--project-management)
9. [Non-Functional Requirements](#9-non-functional-requirements)
10. [Security Requirements](#10-security-requirements)
11. [Audit Trail Requirements](#11-audit-trail-requirements)
12. [Workflow Requirements](#12-workflow-requirements)
13. [Integration Requirements](#13-integration-requirements)
14. [Constraints & Assumptions](#14-constraints--assumptions)
15. [Glossary](#15-glossary)

---

## 1. Business Objective

### 1.1 Latar Belakang

Pengelolaan project di lingkungan kementerian saat ini menghadapi beberapa tantangan:

- Tidak ada visibilitas terpusat atas status seluruh project yang sedang berjalan
- Progress monitoring masih dilakukan manual via spreadsheet atau dokumen terpisah
- Tidak ada early warning system untuk project yang berisiko terlambat atau melebihi budget
- Data manpower, budget, dan timeline tersebar di berbagai sistem
- Tidak ada audit trail yang memadai untuk pertanggungjawaban

### 1.2 Tujuan Sistem

CANKORA dibangun untuk menjadi **Enterprise Operations Platform** yang:

1. Memberikan visibilitas penuh atas seluruh project secara real-time kepada management
2. Menyediakan dashboard eksekutif yang langsung menjawab pertanyaan manajemen tingkat tinggi
3. Mengintegrasikan monitoring project, manpower, budget, timeline, risiko, dan isu dalam satu platform
4. Menyediakan fondasi untuk ekspansi ke modul Fixed Asset, Inventory, dan CMMS
5. Memenuhi standar tata kelola pemerintahan: audit trail, segregation of duties, dan data governance

### 1.3 Business Goals

| ID | Goal | Ukuran Keberhasilan |
|----|------|---------------------|
| BG-01 | Visibilitas project meningkat | 100% project terpantau dalam satu dashboard |
| BG-02 | Deteksi risiko lebih awal | Early warning tampil min. 2 minggu sebelum deadline terlewat |
| BG-03 | Efisiensi pelaporan | Waktu pembuatan laporan berkurang dari hari ke menit |
| BG-04 | Akuntabilitas meningkat | Semua perubahan data terekam di audit trail |
| BG-05 | Skalabilitas platform | Platform siap menerima 3 modul tambahan tanpa arsitektur ulang |

---

## 2. System Scope

### 2.1 Ruang Lingkup — Fase Pertama (MVP)

> **Status implementasi 2026-08-26:** Operational MVP untuk auth, RBAC awal, project/task/milestone/user, dashboard live, Kanban, early warning dasar, project-child soft-delete cascade, Issue, Risk, Budget, Vendor/Contract, Document, Corrective Action, Reporting, master portfolio/spatial dasar, monitoring snapshot, data validation, field evidence, Health Score, Command Center, Level 3 Project Control, dan Benefit Indicator sudah terverifikasi lokal. Level 1/2 resmi, GIS aktual, Power BI/read model, external integration, dan advanced analytics tetap roadmap.

Platform ini pada fase pertama mencakup:

```
CANKORA v1 — Project Management Monitoring
│
├── Core Platform
│   ├── Identity & Access Management (IAM)
│   ├── Dynamic RBAC + Scope-Based Access Control
│   ├── Organization Hierarchy Management
│   ├── Workflow Engine (FSM-based)
│   ├── Notification (interface-ready)
│   ├── Document Management (MinIO-ready)
│   └── Audit Trail (immutable)
│
└── Project Management Module
    ├── Project Information & Registration
    ├── Project Team & Manpower
    ├── Work Breakdown Structure (WBS) / Task
    ├── Milestone
    ├── Progress Monitoring
    ├── Issue Management
    ├── Risk Management
    ├── Budget & Cost Monitoring
    ├── Meeting & Activity Log
    ├── Document Management
    └── Project Dashboard & Reporting
```

### 2.2 Ruang Lingkup — Fase Berikutnya (Planned)

```
CANKORA v2+ (Roadmap)
├── PMO Data Foundation & Core Process
├── Project Intelligence & Control Tower Level 3
├── Control Tower Level 2/Level 1, GIS & Power BI
├── Primavera, Mobile Inspection, BIM/Digital Twin, IoT & AI/ML
├── Fixed Asset Management
├── Inventory Management
└── CMMS (Computerized Maintenance Management System)
```

### 2.3 Di Luar Ruang Lingkup

Hal-hal berikut **tidak** termasuk dalam scope CANKORA:

- Sistem penggajian atau HR management
- Sistem pengadaan barang/jasa (e-procurement)
- Integrasi langsung dengan SAKTI/SIMDA pada fase operational MVP (tetap roadmap melalui adapter)
- Mobile application native (platform web-first)
- Offline mode / PWA pada fase operational MVP (tetap roadmap mobile inspection)

---

## 3. Stakeholder

### 3.1 Internal Stakeholder

| Role | Jabatan | Kebutuhan Utama |
|------|---------|-----------------|
| Executive User | Sekjen, Dirjen, Kairo | Dashboard eksekutif, health overview seluruh project |
| Project Manager | Kasubdit, Kabid | Kelola project, timeline, tim, risiko, isu |
| Project Officer | Staff, Analis | Update progress, task, upload dokumen |
| PMO (Project Management Office) | Tim PMO | Monitor seluruh project di unit, laporan konsolidasi |
| Auditor | Inspektur, Auditor Internal | Audit trail, laporan, akses read-only |
| System Administrator | IT, Admin Aplikasi | Kelola user, role, permission, organisasi |

### 3.2 External Stakeholder

| Role | Kebutuhan Utama |
|------|-----------------|
| Vendor / Kontraktor | (opsional, fase berikutnya) View progress project yang melibatkan mereka |
| BPK / Auditor Eksternal | (opsional) Akses audit trail terbatas |

---

## 4. Organization Hierarchy

### 4.1 Struktur Hierarki

CANKORA mendukung hierarki organisasi kementerian standar:

```
Level 1: Kementerian
    └── Level 2: Sekretariat Jenderal / Direktorat Jenderal / Inspektorat Jenderal / Badan
            └── Level 3: Biro / Direktorat
                    └── Level 4: Bagian / Subdirektorat
                            └── Level 5: Subbagian / Seksi / Unit Kerja
```

### 4.2 Properti Org Unit

Setiap unit organisasi memiliki:
- Kode unit (contoh: `KEMXXX.DITJEN-A.DIT-1.SUBDIT-2`)
- Nama resmi
- Level hierarki (1-5)
- Parent unit
- Kepala unit (relasi ke User)
- Status aktif/nonaktif

### 4.3 Project & Org Unit

Setiap project harus terikat pada minimal satu org unit. Project dapat melibatkan lebih dari satu org unit (lintas direktorat).

---

## 5. User, Role & Permission Concept

### 5.1 Model Akses

CANKORA menggunakan kombinasi **RBAC (Role-Based Access Control)** dan **Scope-Based Access Control**:

```
USER
 ↓
USER GROUP (opsional)
 ↓
ROLE (dinamis, bisa dibuat admin)
 ↓
PERMISSIONS (per resource, per action)
 ↓
SCOPE (batasan data yang bisa diakses)
```

### 5.2 Dimensi Akses

Akses seorang user ditentukan oleh **tiga dimensi**:

1. **Role** — Apa yang boleh dilakukan (permission)
2. **Organization Scope** — Data unit mana yang bisa dilihat
3. **Project Membership** — Apakah user adalah anggota project tersebut

Contoh rule:
```
(Role = Project Manager) AND (Org Scope = Direktorat Infrastruktur)
→ Bisa mengelola semua project di Direktorat Infrastruktur

(Role = Project Officer) AND (Project Member = TRUE)
→ Hanya bisa update task di project yang dia ikuti

(Role = Executive Viewer) AND (Org Scope = ALL)
→ Bisa lihat semua project di semua unit (read-only)

(Role = Auditor) AND (Access = READ_ONLY)
→ Bisa lihat semua data tapi tidak bisa mengubah apapun
```

### 5.3 Dynamic Role

Role **tidak di-hardcode**. Administrator dapat:
- Membuat role baru dengan nama bebas
- Mendefinisikan permission per resource per action
- Menetapkan scope bawaan untuk role
- Menonaktifkan atau menghapus role (selama tidak ada user yang menggunakannya)

### 5.4 User Group

- User dapat dimasukkan ke dalam satu atau lebih Group
- Group memiliki satu atau lebih Role
- User juga bisa mendapat Role secara langsung (tanpa Group)
- Effective permission = union dari semua permission dari semua Role (langsung + via Group)

### 5.5 Default Roles (Seed)

| Role | Deskripsi | Default Scope |
|------|-----------|---------------|
| Super Admin | Akses penuh ke semua fitur termasuk sistem | ALL |
| Admin | Kelola user, role, organisasi. Tidak akses data project | ALL |
| Executive Viewer | Dashboard & laporan read-only semua project | ALL |
| PMO | Monitor semua project di unit yang ditentukan | ORG_UNIT |
| Project Manager | Kelola project yang di-assign | ASSIGNED_PROJECT |
| Project Officer | Update task & progress di project yang diikuti | MEMBER_PROJECT |
| Auditor | Read-only semua data + audit trail | ALL (READ_ONLY) |

### 5.6 Scope Types

| Scope | Deskripsi |
|-------|-----------|
| `ALL` | Akses ke semua data organisasi |
| `ORG_UNIT` | Akses ke data di org unit tertentu beserta turunannya |
| `ASSIGNED_PROJECT` | Akses ke project yang secara eksplisit di-assign ke user |
| `MEMBER_PROJECT` | Akses ke project di mana user adalah anggota tim |

---

## 6. Project Management Business Process

### 6.1 Siklus Hidup Project

```
DRAFT → SUBMITTED → REVIEWED → APPROVED → ACTIVE → ON_HOLD → COMPLETED → CLOSED
                                                    ↑
                                             (bisa kembali ke ACTIVE dari ON_HOLD)
```

| Status | Deskripsi | Siapa yang Trigger |
|--------|-----------|-------------------|
| DRAFT | Project dibuat, belum diajukan | Project Manager |
| SUBMITTED | Diajukan untuk review | Project Manager |
| REVIEWED | Sudah direviu, menunggu approval | PMO / Reviewer |
| APPROVED | Disetujui, siap berjalan | Kepala Unit / Approver |
| ACTIVE | Project sedang berjalan | Auto setelah APPROVED |
| ON_HOLD | Project ditunda sementara | Project Manager + approval |
| COMPLETED | Semua deliverable selesai | Project Manager + approval |
| CLOSED | Secara administratif ditutup | Admin / PMO |

### 6.2 Proses Update Progress

```
Project Officer / PM
    → Update % progress actual
    → Update status task
    → Upload bukti (dokumen, foto)
    → Catat kendala / isu
    
System
    → Hitung overall project health
    → Bandingkan actual vs planned
    → Trigger alert jika at risk / delayed
    → Catat di audit trail
```

### 6.3 Proses Manajemen Risiko

```
1. Identifikasi risiko (PM / Officer)
2. Penilaian: Probability x Impact = Risk Score
3. Penentuan mitigasi (PIC + Due Date)
4. Monitoring status risiko
5. Eskalasi jika risk score tinggi
```

Severity Matrix:

| | Impact Rendah | Impact Sedang | Impact Tinggi |
|-|---------------|---------------|---------------|
| **Probabilitas Tinggi** | 🟡 Medium | 🔴 High | 🔴 Critical |
| **Probabilitas Sedang** | 🟢 Low | 🟡 Medium | 🔴 High |
| **Probabilitas Rendah** | 🟢 Low | 🟢 Low | 🟡 Medium |

### 6.4 Project Health Calculation

```
Overall Health ditentukan berdasarkan kombinasi:

Schedule Health:
  Actual Progress vs Planned Progress
  Overdue Milestone Count

Budget Health:
  Actual Cost vs Budget
  Forecast at Completion vs Budget

Risk Health:
  Count of HIGH/CRITICAL open risks

Issue Health:
  Count of HIGH severity open issues

Result:
  GREEN     — Proyek berjalan sesuai rencana
  YELLOW    — Perlu perhatian dan mitigasi
  RED       — Perlu tindakan segera
  CRITICAL  — Memerlukan intervensi/keputusan pimpinan
```

Target Control Tower memperluas dimensi health menjadi schedule, physical progress, financial progress, contract, risk, issue, quality, dan procurement. Formula, bobot, threshold, dan version harus configurable serta disetujui PMO; mockup bukan persetujuan formula.

---

## 7. Functional Requirements — Core Platform

### 7.1 Authentication & Session Management

| ID | Requirement |
|----|-------------|
| FR-AUTH-01 | User dapat login dengan email + password |
| FR-AUTH-02 | Password disimpan menggunakan bcrypt (cost ≥ 12) |
| FR-AUTH-03 | Session menggunakan JWT dengan expiry yang dapat dikonfigurasi |
| FR-AUTH-04 | Refresh token untuk perpanjang sesi tanpa re-login |
| FR-AUTH-05 | Logout invalidasi token (blacklist JTI) |
| FR-AUTH-06 | Brute-force protection: lockout setelah N kali gagal login |
| FR-AUTH-07 | Forgot password via email token (berlaku 1 jam) |
| FR-AUTH-08 | Wajib ganti password pada login pertama |
| FR-AUTH-09 | Platform harus siap diintegrasikan dengan SSO/Keycloak (interface-based) |

### 7.2 User Management

| ID | Requirement |
|----|-------------|
| FR-USER-01 | Admin dapat membuat, mengedit, menonaktifkan user |
| FR-USER-02 | User memiliki: nama, email, NIP (opsional), jabatan, unit organisasi, foto profil |
| FR-USER-03 | User tidak dapat dihapus permanen, hanya dinonaktifkan (soft delete) |
| FR-USER-04 | User yang nonaktif tidak dapat login |
| FR-USER-05 | Admin dapat reset password user |
| FR-USER-06 | User dapat ubah password sendiri |

### 7.3 Organization Management

| ID | Requirement |
|----|-------------|
| FR-ORG-01 | Admin dapat membuat dan mengelola hierarki org unit (min. 5 level) |
| FR-ORG-02 | Setiap org unit memiliki kode unik, nama, level, parent |
| FR-ORG-03 | Org unit tidak dapat dihapus jika masih ada user atau project yang terikat |
| FR-ORG-04 | Admin dapat memindah user antar org unit |
| FR-ORG-05 | Sistem mendukung tampilan tree/hierarki org unit |

### 7.4 Dynamic RBAC

| ID | Requirement |
|----|-------------|
| FR-RBAC-01 | Admin dapat membuat Role dengan nama bebas |
| FR-RBAC-02 | Setiap Role memiliki permission per resource per action |
| FR-RBAC-03 | Action standar: view, create, update, delete, approve, export |
| FR-RBAC-04 | Admin dapat membuat Group dan memasukkan Role ke dalam Group |
| FR-RBAC-05 | User dapat di-assign ke Group atau ke Role secara langsung |
| FR-RBAC-06 | Satu user bisa berada di beberapa Group |
| FR-RBAC-07 | Effective permission = union dari semua permission user |
| FR-RBAC-08 | Scope dapat di-set per user per role: ALL / ORG_UNIT / PROJECT |
| FR-RBAC-09 | Role tidak dapat dihapus jika masih digunakan user |
| FR-RBAC-10 | Perubahan role dan permission harus tercatat di audit trail |

### 7.5 Audit Trail

| ID | Requirement |
|----|-------------|
| FR-AUDIT-01 | Semua operasi Create, Update, Delete, Approve, Reject, Export dicatat |
| FR-AUDIT-02 | Setiap record audit berisi: who, what, when, where (IP), old value, new value |
| FR-AUDIT-03 | Audit log bersifat immutable — tidak dapat diubah atau dihapus |
| FR-AUDIT-04 | Admin/Auditor dapat filter audit log: by user, by resource, by action, by date range |
| FR-AUDIT-05 | Audit log dapat di-export (CSV/Excel) |
| FR-AUDIT-06 | Login, logout, dan failed login dicatat |
| FR-AUDIT-07 | Perubahan role, permission, dan group harus dicatat dengan detail (old vs new) |

---

## 8. Functional Requirements — Project Management

> **Catatan status 2026-08-20:** Bagian ini adalah requirement target. Status aktual harus dirujuk ke tabel di bawah dan ke `IMPLEMENTATION-GAP-ANALYSIS.md`.

### 8.0 Status Implementasi Requirement PMO

| Area | Status Saat Ini | Catatan Gap |
|------|-----------------|-------------|
| Project Information & Registration | Implemented/Basic | CRUD, transition, progress history, dan UI dasar terverifikasi; metadata SDA/program/lokasi belum lengkap |
| Project Team & Manpower | Partial | Team member model ada; UI/workload belum usable |
| WBS / Task | Implemented/Basic | CRUD nested dan Kanban minimal terverifikasi; assignment, komentar, attachment, dan WBS level 3 belum lengkap |
| Milestone | Implemented/Basic | CRUD nested terverifikasi; deliverable/evidence/alert lanjut belum lengkap |
| Timeline/Gantt | Missing | Belum ada Gantt/timeline interaktif |
| Progress Monitoring | Implemented/Basic | Progress dan history dasar ada; baseline, snapshot periodik, validasi, dan deviasi belum lengkap |
| Issue Management | Implemented/Basic | P1-001 sudah usable: create/list/update/status/soft delete, severity, escalation, PIC, due date, resolution, audit; dashboard executive issue summary masih belum membaca register issue secara resmi |
| Risk Management | Implemented/Basic | P1-002 usable end-to-end: model + migration `000009` (organization_id, probability/impact INT 1-5, risk_score, severity), repo/service org-scoped + project-scoped, score probability × impact dihitung backend, FSM status, soft delete, audit, routes `RequirePermission(ResourceRisks, ...)`, UI project detail (register, form, status move, edit/delete), dashboard/Command Center membaca risk register via warning `RISK_REGISTER`; Health Score risk dimension masih menunggu P1-014 |
| Budget & Cost Monitoring | Implemented/Basic | P1-003 usable end-to-end: line item budget (planned/actual/currency) CRUD tenant-safe via ownership parent project, variance + usage_pct + status (NORMAL/WATCH/RISK/OVERRUN) selalu dihitung backend, dashboard `BUDGET_THRESHOLD` warning agregat (`pb.deleted_at IS NULL` + `p.deleted_at IS NULL`), audit, UI project detail (panel + create/edit/delete). Belum: forecast at completion, baseline/snapshot validasi (P1-011), dan data quality (P1-012) |
| Contract & Vendor | Implemented/Basic | P1-004 usable end-to-end: master `vendors` (VENDOR/CONSULTANT) tenant-scoped + `contracts` per project tenant-safe (organization_id + parent project ownership), contract_number unique per org, contract_value/currency/status/signed/start/end/scope validation, audit, routes `RequirePermission(ResourceVendors/ResourceContracts, ...)`, UI project detail panel "Kontrak" (summary + list + create/edit/delete). Contract value adalah operational input yang BELUM otomatis mengubah `project.budget_total` dan belum validated/published (menunggu P1-011 snapshot/validasi). Dashboard belum mengklaim contract health.
| Meeting & Activity Log | Missing | Belum ada modul |
| Document Management | Partial/Conceptual | Model ada sebagian; upload/list/download/delete belum usable |
| Dashboard & Reporting | Implemented/Basic | Dashboard live memakai current operational tables dan early warning dasar; sebagian KPI/Command Center masih derived di frontend/placeholder; Control Tower Level 1-3, validation, snapshot periodik, Health Score resmi, Power BI, GIS, reporting periodik, dan export belum tersedia |

### 8.1 Project Information

> **Planning note 2026-08-29 — DASH-003:** Field bisnis `BALAI` untuk tampilan dashboard tidak membutuhkan kolom database baru pada tahap ini. Sumber data resmi yang sudah tersedia adalah `projects.org_unit_id` yang mereferensi `org_units.name`. Dalam konteks CANKORA, org unit level Unit/Satker/Balai/BBWS/BWS berperan sebagai pemilik atau inisiator proyek. `created_by` hanya pembuat record dan `manager_id` adalah PIC personal, sehingga tidak boleh dipakai sebagai `BALAI`.

| ID | Requirement |
|----|-------------|
| FR-PM-01 | User dapat membuat project dengan: nama, kode, deskripsi, client/stakeholder, org unit, tipe, kategori, PM, start date, target end date |
| FR-PM-02 | Project memiliki status workflow: DRAFT → APPROVED → ACTIVE → COMPLETED → CLOSED |
| FR-PM-03 | Project dapat diklasifikasikan: by direktorat, by unit, by tipe, by status |
| FR-PM-04 | Project dapat memiliki tag/label untuk kategorisasi |
| FR-PM-05 | Project memiliki foto/gambar thumbnail (opsional) |

### 8.2 Project Team

| ID | Requirement |
|----|-------------|
| FR-PT-01 | PM dapat menambah/menghapus anggota tim ke project |
| FR-PT-02 | Setiap anggota tim memiliki role dalam project: PM, Officer, QA, Reviewer, Viewer |
| FR-PT-03 | Satu user dapat menjadi anggota di beberapa project dengan role berbeda |
| FR-PT-04 | PM dapat melihat workload anggota tim (jumlah task aktif) |

### 8.3 Work Breakdown Structure (WBS) / Task

| ID | Requirement |
|----|-------------|
| FR-WBS-01 | Project dapat memiliki WBS dengan maksimal 3 level (Task → Subtask → Sub-subtask) |
| FR-WBS-02 | Setiap task memiliki: judul, deskripsi, assignee, start date, due date, estimasi jam, actual jam, priority, status, kategori |
| FR-WBS-03 | Status task: Backlog → Analysis → Development → QA → UAT → Production → Done |
| FR-WBS-04 | Task dapat di-assign ke lebih dari satu user |
| FR-WBS-05 | Task memiliki indikator progress 0-100% |
| FR-WBS-06 | Task dapat memiliki attachment/dokumen |
| FR-WBS-07 | Task dapat memiliki komentar/diskusi |
| FR-WBS-08 | Task dapat ditandai sebagai blocked, dengan penjelasan |
| FR-WBS-09 | Project detail menyediakan Kanban board task minimal berdasarkan status task untuk monitoring dan pemindahan status operasional |

### 8.4 Milestone

| ID | Requirement |
|----|-------------|
| FR-MS-01 | Project dapat memiliki milestone dengan target date dan deliverable |
| FR-MS-02 | Milestone memiliki status: Upcoming, In Progress, Achieved, Missed |
| FR-MS-03 | Sistem otomatis men-trigger alert jika milestone mendekati due date (H-7, H-3) |
| FR-MS-04 | Pencapaian milestone memerlukan bukti (dokumen/komentar) |

### 8.5 Timeline

| ID | Requirement |
|----|-------------|
| FR-TL-01 | Sistem menampilkan Gantt chart sederhana untuk task dan milestone |
| FR-TL-02 | Gantt chart membandingkan planned vs actual |
| FR-TL-03 | Task yang terlambat tampil dengan indikator merah |
| FR-TL-04 | PM dapat mengubah tanggal task langsung dari Gantt chart (drag) |

> **Planning note 2026-08-18:** Kanban minimal dimasukkan ke `CANKORA-P0-010` sebagai bagian dari task UI nested. Gantt/timeline sederhana dimasukkan ke backlog P1 (`CANKORA-P1-009`); drag-edit tanggal tetap target lanjutan setelah timeline read-only stabil.

### 8.6 Progress Monitoring

| ID | Requirement |
|----|-------------|
| FR-PRG-01 | Progress project dapat di-update manual oleh PM/Officer |
| FR-PRG-02 | Sistem juga menghitung auto-progress berdasarkan % task selesai |
| FR-PRG-03 | Tersedia riwayat perubahan progress (history) |
| FR-PRG-04 | Sistem membandingkan actual progress vs planned progress |
| FR-PRG-05 | Sistem menghitung dan menampilkan Schedule Variance (SV) dan Schedule Performance Index (SPI) |

### 8.7 Issue Management

| ID | Requirement |
|----|-------------|
| FR-ISS-01 | User dapat mencatat issue dengan: judul, deskripsi, severity (Low/Medium/High/Critical), status, PIC, due date, dampak |
| FR-ISS-02 | Status issue: Open → In Progress → Resolved → Closed |
| FR-ISS-03 | Issue dapat di-eskalasi ke level manajemen |
| FR-ISS-04 | Issue mendekati due date memicu notifikasi ke PIC |
| FR-ISS-05 | Issue resolved memerlukan catatan resolusi |

### 8.8 Risk Management

| ID | Requirement |
|----|-------------|
| FR-RSK-01 | User dapat mencatat risiko dengan: judul, deskripsi, probabilitas, dampak, risk score, PIC, due date, mitigasi |
| FR-RSK-02 | Risk score dihitung otomatis: Probability × Impact |
| FR-RSK-03 | Status risiko: Identified → Monitored → Mitigated → Closed |
| FR-RSK-04 | Risiko HIGH dan CRITICAL memicu notifikasi ke PM dan atasan |

> **Status implementasi 2026-08-25 (P1-002):** FR-RSK-01/02/03 implemented. Implementasi FSM aktual lebih eksplisit dari konsep: `IDENTIFIED→ASSESSED→MITIGATED|ACCEPTED|ESCALATED`; `MITIGATED→CLOSED`; `ACCEPTED→CLOSED`; `ESCALATED→MITIGATED`. Reject transisi invalid via `workflow.ErrTransitionNotAllowed` (409). FR-RSK-04 (notifikasi HIGH/CRITICAL) belum implemented — menunggu modul notification lanjut.

### 8.9 Budget & Cost

| ID | Requirement |
|----|-------------|
| FR-BDG-01 | Project memiliki: project value, budget/COGS, planned cost per komponen |
| FR-BDG-02 | Actual cost dapat di-input per komponen (tenaga, infrastruktur, operasional, dll) |
| FR-BDG-03 | Sistem menampilkan: budget vs actual, variance, % utilized, forecast at completion |
| FR-BDG-04 | Alert jika actual cost melebihi threshold (mis. 80% dari budget) |

> **Status implementasi 2026-08-25 (P1-003):** FR-BDG-01 (parsial — `project_budgets` line items per kategori dengan planned), FR-BDG-02 (implemented — actual realization per line item), FR-BDG-03 (parsial — variance dan % utilized (`usage_pct`) dihitung backend; forecast at completion belum, menunggu P1-011 baseline/snapshot), FR-BDG-04 (implemented — status otomatis NORMAL <80 / WATCH ≥80 / RISK ≥90 / OVERRUN ≥100 per line dan agregat; dashboard `BUDGET_THRESHOLD` warning memakai filter `project_budgets.deleted_at IS NULL` + `projects.deleted_at IS NULL`). Nilai budget/realisasi adalah operational input yang belum melalui workflow validasi (menunggu P1-011/P1-012), dan label UI/dashboard menegaskan hal itu.

### 8.9 bis Contract & Vendor (P1-004)

| ID | Requirement |
|----|-------------|
| FR-CNT-01 | Master vendor/penyedia (VENDOR) dan konsultan supervisi (CONSULTANT) tersedia dengan CRUD, tenant-scoped |
| FR-CNT-02 | Vendor memiliki: nama (wajib), tipe enum, legal_name, tax_id, contact, email, phone, address, is_active |
| FR-CNT-03 | Kontrak terikat ke project dan organization |
| FR-CNT-04 | Kontrak memiliki: contract_number (unique per org), title, vendor (wajib), consultant (opsional), contract_value, currency (default IDR), signed/start/end date, status (DRAFT/ACTIVE/AMENDED/COMPLETED/TERMINATED), scope_of_work |
| FR-CNT-05 | Validasi: start_date ≤ end_date, contract_value tidak negatif, contract_number unique per org |
| FR-CNT-06 | Vendor yang masih dipakai contract aktif tidak bisa dihapus; semua delete soft delete |
| FR-CNT-07 | Audit `vendor.*` dan `contract.*` |

> **Status implementasi 2026-08-26 (P1-004):** FR-CNT-01 s.d. FR-CNT-07 implemented end-to-end. Migration `000010_vendors_contracts`; routes `/api/v1/vendors` (top-level) dan `/api/v1/projects/:id/contracts` (nested) dengan `RequirePermission(ResourceVendors, ...)` / `RequirePermission(ResourceContracts, ...)`. Tenant/project ownership selalu diverifikasi (contract di project salah atau missing → 404); duplicate contract_number → 409. Vendor yang direferensikan contract aktif → 409; setelah contract dihapus vendor baru bisa dihapus (204). Contract value BELUM otomatis mengubah `project.budget_total` dan belum validated/published (menunggu P1-011/P1-012); UI project detail menampilkan label "operational contract input".

### 8.10 Meeting & Activity Log

| ID | Requirement |
|----|-------------|
| FR-MTG-01 | PM/Officer dapat mencatat meeting: tanggal, topik, peserta, notulen, tindak lanjut |
| FR-MTG-02 | Activity log otomatis mencatat semua perubahan penting di project |

### 8.11 Document Management

| ID | Requirement |
|----|-------------|
| FR-DOC-01 | User dapat upload dokumen ke project: TOR, KAK, Kontrak, BAST, laporan, foto |
| FR-DOC-02 | Dokumen memiliki: nama, tipe, versi, tanggal upload, uploader |
| FR-DOC-03 | Dokumen dapat di-download oleh user yang berhak |
| FR-DOC-04 | Storage menggunakan MinIO (on-premise ready) |

### 8.12 Dashboard & Reporting

**Status sumber data aktual 2026-08-25:** Dashboard yang sudah berjalan membaca `GET /api/v1/dashboard` dan `GET /api/v1/projects`. Backend dashboard menghitung agregat dari current state `projects`, `tasks`, `milestones`, `users`, dan `project_budgets`; frontend menghitung nilai portofolio, rata-rata progress, distribusi status, ranking, dan watchlist dari list project dan early warning. Belum ada period cut-off, validation status, report snapshot immutable, analytics read model, atau KPI dictionary resmi. Karena itu, dashboard saat ini memenuhi monitoring operasional dasar, tetapi belum memenuhi reporting resmi Control Tower.

| ID | Requirement |
|----|-------------|
| FR-RPT-01 | Executive Dashboard menampilkan total project, `GREEN`, `YELLOW`, `RED`, `CRITICAL`, completed, dan average progress sesuai formula version aktif |
| FR-RPT-02 | Dashboard menampilkan: Project Health matrix, Progress per project, Upcoming Milestone, Overdue Task, Top Issues/Risks |
| FR-RPT-03 | Dashboard dapat difilter: by org unit, by status, by periode |
| FR-RPT-04 | Project dapat dilihat per direktorat, per unit, per vendor, per status |
| FR-RPT-05 | Budget vs Actual chart tersedia di dashboard |
| FR-RPT-06 | Laporan project dapat di-export (PDF, Excel) |
| FR-RPT-07 | Report Builder sederhana untuk laporan ad-hoc (fase berikutnya) |

### 8.13 Control Tower Level 1 - Nasional

| ID | Requirement |
|----|-------------|
| FR-CTL1-01 | Dashboard nasional menampilkan total proyek, nilai kontrak/portofolio, progress fisik, progress keuangan, deviasi waktu/biaya, dan distribusi health |
| FR-CTL1-02 | Dashboard menampilkan peta sebaran proyek nasional, trend rencana versus realisasi, proyek kritis, serta antrean keputusan pimpinan |
| FR-CTL1-03 | Setiap KPI dapat di-drill-down sampai program, wilayah, proyek, periode, sumber data, dan bukti pendukung |
| FR-CTL1-04 | Dashboard menampilkan indikator manfaat nasional yang definisi, unit, target, realisasi, periode, dan sumbernya terdokumentasi |
| FR-CTL1-05 | Pengguna dapat memfilter periode, program/sektor, wilayah, org unit, status, dan health sesuai scope akses |

### 8.14 Control Tower Level 2 - Program/Sektor

| ID | Requirement |
|----|-------------|
| FR-CTL2-01 | Sistem menyediakan dashboard per program/sektor: bendungan, irigasi, pengairan dan pengendalian banjir, air baku, dan pertanian |
| FR-CTL2-02 | Dashboard program menampilkan jumlah proyek, nilai kontrak/anggaran, progress fisik/keuangan, distribusi health, dan trend |
| FR-CTL2-03 | Sistem menyediakan perbandingan antarprogram, top deviasi waktu, top deviasi biaya, proyek berisiko tinggi, dan trend realisasi anggaran |
| FR-CTL2-04 | Pengguna dapat drill-down dari program/sektor ke wilayah, org unit, dan proyek |

### 8.15 Control Tower Level 3 - Detail Proyek

| ID | Requirement |
|----|-------------|
| FR-CTL3-01 | Detail proyek menampilkan lokasi, sungai/DAS, program/sektor, nilai kontrak, penyedia, konsultan, tanggal kontrak, target selesai, dan penanggung jawab |
| FR-CTL3-02 | Detail proyek menampilkan progress fisik dan keuangan, deviasi waktu dan biaya, health score beserta alasan, serta jadwal rencana versus realisasi |
| FR-CTL3-03 | Detail proyek menyediakan bukti lapangan, dokumen kontrak, foto progress, hasil inspeksi, isu, risiko, dan tindak lanjut |
| FR-CTL3-04 | Data ringkasan proyek harus konsisten dengan task, milestone, kontrak, snapshot progress, dan data tervalidasi pada periode yang sama |

### 8.16 Project Health Score

| ID | Requirement |
|----|-------------|
| FR-HLT-01 | Health Score menghitung dimensi schedule, physical, financial, contract, risk, issue, quality, dan procurement |
| FR-HLT-02 | Hasil diklasifikasikan menjadi `GREEN`, `YELLOW`, `RED`, atau `CRITICAL` |
| FR-HLT-03 | Formula, bobot, normalisasi, missing-data rule, dan threshold harus configurable, versioned, dan mendapat approval PMO |
| FR-HLT-04 | Setiap hasil menyimpan component score, formula version, period, calculated_at, dan explanation |
| FR-HLT-05 | Perubahan formula tidak mengubah snapshot historis yang sudah dipublikasikan |

### 8.17 Data Quality, Validation, dan Freshness

| ID | Requirement |
|----|-------------|
| FR-DQ-01 | Input manual, file, dan integrasi eksternal masuk validation queue sebelum dipakai pada laporan resmi |
| FR-DQ-02 | Setiap submission menyimpan source, period, submitter, validator, validation status, freshness, completeness, dan rejection reason |
| FR-DQ-03 | Status validasi minimal `DRAFT`, `SUBMITTED`, `VALID`, `REJECTED`, dan `STALE` |
| FR-DQ-04 | Dashboard menampilkan data as-of time dan indikator stale/missing data; missing data tidak dianggap sehat |
| FR-DQ-05 | Koreksi dan validasi data tercatat pada audit trail |

> **Status implementasi 2026-08-28 (P3-003):** FR-DQ-01/02/05 implemented end-to-end untuk jalur *official data*: modul `governance` menyediakan submission resmi berisi items (action CREATE/UPDATE/DELETE/UPSERT/VALIDATE_ONLY), per-item validation (tenant ownership, soft-delete ditolak, mapping pemerintah `PENDING_MATCH` → `INVALID`), status FSM DRAFT→SUBMITTED→IN_REVIEW→APPROVED→LOCKED dengan REJECTED (alasan wajib, kembali ke DRAFT) dan CANCELLED, serta audit `governance.*`. FR-DQ-03 dipenuhi dua jalur: queue data quality memakai `DRAFT/SUBMITTED/VALID/REJECTED/STALE` (P1-012), sedangkan jalur official memakai status governance di atas (P3-003). FR-DQ-04 (dashboard menandai stale/missing dari data official) masih roadmap — snapshot resmi yang sudah LOCKED tersedia sebagai dasar perhitungan. Data yang di-import/synced (P2-001, P2-002, P2-011) tidak otomatis di-approve; harus melalui submission resmi.

### 8.18 GIS dan Bukti Lapangan

| ID | Requirement |
|----|-------------|
| FR-GIS-01 | Proyek dapat dipetakan berdasarkan koordinat, provinsi, kabupaten/kota, org unit, dan sungai/DAS |
| FR-GIS-02 | Peta mendukung cluster, filter health, hotspot, dan drill-down ke detail proyek sesuai scope |
| FR-FLD-01 | Petugas dapat mencatat inspeksi dengan waktu, lokasi, petugas, catatan, status verifikasi, foto, dan dokumen |
| FR-FLD-02 | Evidence menyimpan metadata dan checksum; file di object storage, metadata di PostgreSQL |

### 8.19 Alert, Corrective Action, dan Keputusan

| ID | Requirement |
|----|-------------|
| FR-CMD-01 | Alert memiliki jenis, severity, sumber indikator, aging, PIC, SLA, status, dan link ke bukti |
| FR-CMD-02 | Alert dapat menghasilkan corrective action berisi deviasi, akar masalah, rekomendasi, PIC, target, progress, evidence, dan verifikasi selesai |
| FR-CMD-03 | PMO dapat melakukan escalation dan mencatat keputusan/arahan pimpinan beserta owner, due date, serta tindak lanjut |
| FR-CMD-04 | Command Center menampilkan alert aktif, validasi tertunda, corrective action overdue, top risk/issue, freshness, dan watchlist proyek prioritas |

**Status implementasi 2026-08-26:** `CANKORA-P1-015` Done untuk lifecycle API/UI dasar. Command Center membaca aggregate backend, validation queue, corrective action, risk watchlist, persisted escalation, dan executive decision follow-up. Create escalation/decision memvalidasi ownership source/project/user agar tidak membuat referensi lintas tenant; status action tercatat di audit. Peta masih ilustratif sampai P2-008 dan Level 1/2 resmi tetap menunggu tiket P2.

### 8.20 Power BI dan Analytics Read Model

| ID | Requirement |
|----|-------------|
| FR-BI-01 | CANKORA menyediakan read model/semantic dataset terdokumentasi untuk Power BI tanpa memberi akses tulis ke tabel operasional |
| FR-BI-02 | Dataset menerapkan tenant filter, row-level security, data dictionary, metric definition, refresh schedule, dan refresh monitoring |
| FR-BI-03 | Dashboard native dan Power BI menggunakan definisi KPI serta period cut-off yang sama |
| FR-BI-04 | Kegagalan refresh terlihat oleh integration admin dan tidak boleh menyamarkan data lama sebagai data terkini |

### 8.21 Manfaat dan Outcome

| ID | Requirement |
|----|-------------|
| FR-BNF-01 | Sistem dapat mendefinisikan indikator manfaat per program/proyek dengan unit, target, periode, baseline, realisasi, owner, dan source |
| FR-BNF-02 | Agregasi manfaat hanya dilakukan untuk indikator dengan unit dan metode agregasi yang kompatibel |
| FR-BNF-03 | Angka manfaat pada mockup tidak menjadi seed atau target produksi tanpa data resmi |

---

## 9. Non-Functional Requirements

### 9.1 Performance

| ID | Requirement |
|----|-------------|
| NFR-PERF-01 | Response time API ≤ 500ms untuk 95% request dalam kondisi normal |
| NFR-PERF-02 | Dashboard eksekutif load ≤ 3 detik untuk data 100 project |
| NFR-PERF-03 | Sistem mendukung minimal 200 concurrent user tanpa degradasi signifikan |
| NFR-PERF-04 | Query database yang berat wajib menggunakan index yang sesuai |

### 9.2 Availability

| ID | Requirement |
|----|-------------|
| NFR-AV-01 | Target uptime 99.5% pada jam kerja (07:00-17:00 WIB, hari kerja) |
| NFR-AV-02 | Scheduled maintenance dilakukan di luar jam kerja |
| NFR-AV-03 | Sistem memiliki health check endpoint |

### 9.3 Scalability

| ID | Requirement |
|----|-------------|
| NFR-SC-01 | Arsitektur modular memungkinkan penambahan modul tanpa refactoring besar |
| NFR-SC-02 | Database schema mempertimbangkan multi-tenancy (organization_id di semua tabel utama) |
| NFR-SC-03 | Backend stateless untuk mendukung horizontal scaling di masa depan |

### 9.4 Maintainability

| ID | Requirement |
|----|-------------|
| NFR-MT-01 | Codebase mengikuti konvensi yang terdefinisi di CLAUDE.md |
| NFR-MT-02 | Setiap modul independen dengan interface yang jelas |
| NFR-MT-03 | Database migration menggunakan tool standar (golang-migrate) |
| NFR-MT-04 | Semua konfigurasi dari environment variable, tidak ada hardcode |
| NFR-MT-05 | Structured logging dengan level: DEBUG, INFO, WARN, ERROR |

### 9.5 Usability

| ID | Requirement |
|----|-------------|
| NFR-UX-01 | UI responsif untuk layar minimal 1280px lebar |
| NFR-UX-02 | Loading state dan error state di semua komponen yang fetch data |
| NFR-UX-03 | Pesan error yang informatif (bukan hanya kode HTTP) |
| NFR-UX-04 | Operasi destruktif (delete, reject) memerlukan konfirmasi |

---

## 10. Security Requirements

| ID | Requirement |
|----|-------------|
| SEC-01 | Semua komunikasi menggunakan HTTPS (TLS 1.2+) |
| SEC-02 | Password hashing: bcrypt dengan cost factor ≥ 12 |
| SEC-03 | JWT secret minimal 32 karakter, disimpan di env var |
| SEC-04 | Input validation dan sanitasi di semua endpoint API |
| SEC-05 | SQL injection prevention via parameterized query / ORM |
| SEC-06 | Rate limiting pada endpoint auth (login, forgot password) |
| SEC-07 | CORS dikonfigurasi secara ketat |
| SEC-08 | File upload: validasi tipe file, scan ukuran, nama file di-sanitasi |
| SEC-09 | Principle of least privilege: user hanya dapat akses data sesuai scope |
| SEC-10 | Sensitive data (password, token) tidak boleh muncul di log |
| SEC-11 | API endpoint tidak mengekspos informasi sistem internal di error message |
| SEC-12 | HTTP security headers: HSTS, X-Frame-Options, X-Content-Type-Options, CSP |

---

## 11. Audit Trail Requirements

### 11.1 Cakupan Audit

Semua operasi berikut **wajib** dicatat:

| Kategori | Operasi yang Dicatat |
|----------|---------------------|
| **Auth** | Login, Logout, Failed Login, Password Change, Password Reset |
| **User** | Create, Update, Deactivate, Role Change |
| **Role & Permission** | Create, Update, Delete Role; Add/Remove Permission; Assign/Revoke User Role |
| **Project** | Create, Update, Status Change, Delete |
| **Task** | Create, Update, Status Change, Assign, Delete |
| **Milestone** | Create, Update, Achieve, Miss |
| **Issue** | Create, Update, Escalate, Resolve, Close |
| **Risk** | Create, Update, Mitigate, Close |
| **Budget** | Create, Update Budget Entry |
| **Document** | Upload, Download, Delete |
| **Export** | Setiap export data dicatat (who, what, when) |

### 11.2 Format Record Audit

```json
{
  "id": "uuid",
  "timestamp": "2026-08-14T10:41:00+07:00",
  "user_id": "uuid",
  "user_name": "Budi Santoso",
  "user_ip": "192.168.1.100",
  "action": "UPDATE",
  "resource_type": "project",
  "resource_id": "uuid",
  "resource_name": "Digitalisasi Layanan Publik 2026",
  "field_changed": "progress",
  "old_value": "45",
  "new_value": "60",
  "metadata": {}
}
```

---

## 12. Workflow Requirements

### 12.1 Pendekatan Workflow

CANKORA menggunakan **Finite State Machine (FSM) per entitas** yang transisinya dapat dikonfigurasi melalui tabel `workflow_definitions` dan `workflow_transitions`.

Ini memberikan fleksibilitas konfigurasi tanpa kompleksitas full workflow engine.

### 12.2 Workflow Project

```
DRAFT → SUBMITTED → REVIEWED → APPROVED → ACTIVE → ON_HOLD → COMPLETED → CLOSED
```

Setiap transisi memiliki:
- Role yang diizinkan melakukan transisi
- Apakah memerlukan approval (boolean)
- Role approver (jika diperlukan)
- Notifikasi yang dikirim

### 12.3 Workflow Approval Request

```
PENDING → APPROVED / REJECTED
```

Rejection dapat mengembalikan dokumen ke status sebelumnya (configurable per workflow).

> **Status implementasi 2026-08-28 (P3-003):** Approval resmi data dilayani modul `governance` dengan FSM ketat: `DRAFT→SUBMITTED→IN_REVIEW→APPROVED→LOCKED`; `REJECTED` (wajib `rejection_reason`) hanya dari `IN_REVIEW` dan kembali ke `DRAFT`; `CANCELLED` dari `DRAFT/SUBMITTED`. Transisi invalid → HTTP 409 `ErrInvalidTransition`. Review/approve/lock membutuhkan resource `ResourceDataGovernance` (permission `data_governance`); approver dicatat (actor + timestamp) di submission dan audit.

---

## 13. Integration Requirements

> **Status implementasi 2026-08-20:** Belum ada integration adapter aktif untuk sistem eksternal. Untuk PMO Control Tower, minimal perlu import manual CSV/Excel beserta validation/lineage sebelum klaim "data integration" terpenuhi.

### 13.1 Fase Pertama

| Integrasi | Status | Keterangan |
|-----------|--------|-----------|
| Email (SMTP) | Required | Notifikasi, reset password |
| MinIO / S3 | Required | Document storage |
| Redis | Ready (interface) | Cache, session, rate limiting — aktif fase berikutnya |

### 13.2 Fase Berikutnya (Planned)

| Integrasi | Keterangan |
|-----------|-----------|
| SSO / Keycloak | Single sign-on kementerian |
| SAKTI | Sinkronisasi data keuangan/anggaran |
| SIMPEG | Sinkronisasi data pegawai |
| WhatsApp / Telegram | Notifikasi push |
| Power BI | Dashboard eksekutif melalui analytics read model, semantic dataset, RLS, dan controlled refresh |
| GIS/Spatial | Peta proyek, DAS, wilayah administratif, cluster, dan hotspot |
| SIRUP/SIMPONI/OM-SPAN/SAKTI | Kontrak, pengadaan, anggaran, dan realisasi melalui adapter yang disetujui |
| Primavera P6 | Baseline dan realisasi jadwal proyek |
| Mobile Field Inspection | Capture inspeksi dan evidence lapangan |
| BIM/Digital Twin | Referensi model dan aset digital proyek |
| IoT/Sensor | Telemetry lapangan ditempatkan sebagai final/future advanced integration. Bentuk awal nanti: IoT data gateway + device/source registry + telemetry validation + alert foundation; MQTT, time-series optimization, alert rules engine, dan digital twin overlay menyusul setelah foundation stabil. Jangan mulai P3-004 sebelum governance hardening, UI regression, reporting/read-model reconciliation, dan Phase 3 non-IoT stabil. |
| AI/ML | Early warning prediktif setelah histori data tervalidasi mencukupi |

---

## 14. Constraints & Assumptions

### 14.1 Constraints

- Deployment on-premise (server kementerian), tidak boleh ada data ke cloud publik
- Browser yang didukung: Chrome 110+, Edge 110+, Firefox 115+ (tidak perlu IE)
- Bahasa antarmuka: Bahasa Indonesia (i18n-ready untuk Inggris)

### 14.2 Assumptions

- Kementerian memiliki infrastruktur server yang memadai untuk menjalankan Docker containers
- Tim developer familiar dengan Go dan Next.js/TypeScript
- User memiliki akses browser modern dan koneksi internet/intranet yang stabil
- Admin kementerian akan mengelola data master organisasi (input awal)

---

## 15. Glossary

| Istilah | Definisi |
|---------|----------|
| **RBAC** | Role-Based Access Control — model akses berbasis peran |
| **Scope** | Batasan data yang bisa diakses oleh user berdasarkan dimensi tertentu |
| **WBS** | Work Breakdown Structure — dekomposisi pekerjaan project menjadi task |
| **Milestone** | Titik pencapaian penting dalam project yang memiliki target tanggal |
| **Health** | Indikator kesehatan proyek: `GREEN`, `YELLOW`, `RED`, atau `CRITICAL` berdasarkan formula versioned |
| **SPI** | Schedule Performance Index = Actual Progress / Planned Progress |
| **FSM** | Finite State Machine — mesin status dengan transisi yang terdefinisi |
| **PMO** | Project Management Office — unit yang mengawasi pengelolaan project |
| **Org Unit** | Unit organisasi dalam hierarki kementerian |
| **Tenant** | Organisasi (kementerian) yang menggunakan platform |
| **MinIO** | Object storage open-source kompatibel dengan S3, untuk deployment on-premise |
| **CMMS** | Computerized Maintenance Management System |
| **Control Tower** | Tampilan dan workflow pengendalian berjenjang dari nasional, program/sektor, sampai detail proyek |
| **Data Freshness** | Selisih waktu antara periode/waktu data sumber dan waktu data tersedia untuk konsumsi |
| **DAS** | Daerah Aliran Sungai |
| **RLS** | Row-Level Security untuk membatasi baris data pada dataset analytics/Power BI |
