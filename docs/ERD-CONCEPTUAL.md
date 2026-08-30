# ERD Konseptual
# PMO — Enterprise Operations Platform

**Versi**: 0.2.4  
**Tanggal**: 2026-08-29  
**Dibuat oleh**: Cankonix  
**Status**: Draft — conceptual data model

> **Planning note 2026-08-29 — PMO-DASH-003: BALAI display source**
> Kolom UI `BALAI` pada Dashboard "10 Proyek Prioritas" harus berasal dari relasi existing `projects.org_unit_id` → `org_units.id/name`. Tidak perlu migration/kolom database baru kecuali nanti business owner dipisahkan dari struktur org unit. `created_by` dan `manager_id` bukan sumber yang tepat untuk `BALAI`.

> **Update 2026-08-29 — PMO-DASH-002: project_periodic_reports**
> Tabel baru `project_periodic_reports` (migration 000038): `id UUID PK`, `organization_id UUID NOT NULL REFERENCES organizations(id)`, `project_id UUID NOT NULL REFERENCES projects(id)`, `period_year SMALLINT NOT NULL CHECK(2000..2100)`, `period_month SMALLINT NOT NULL CHECK(1..12)`, `physical_progress_pct NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK(0..100)`, `financial_planned NUMERIC(20,2) NOT NULL DEFAULT 0 CHECK(>=0)`, `financial_actual NUMERIC(20,2) NOT NULL DEFAULT 0 CHECK(>=0)`, `financial_pct NUMERIC(8,4) NOT NULL DEFAULT 0` (backend-computed: `financial_actual/financial_planned*100`, 0 jika planned=0), `notes TEXT`, `reported_by UUID REFERENCES users(id)`, `reported_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `created_at`, `updated_at`, `deleted_at TIMESTAMPTZ`. Unique partial index `uq_periodic_report_active` pada `(organization_id, project_id, period_year, period_month) WHERE deleted_at IS NULL`. Index performa: `idx_periodic_report_org_period (organization_id, period_year, period_month) WHERE deleted_at IS NULL`, `idx_periodic_report_project (project_id, period_year, period_month) WHERE deleted_at IS NULL`. Klasifikasi data: **OPERATIONAL** (laporan periodik operasional — belum official-governed/validated). Dashboard trend `/api/v1/dashboard/trend` memprioritaskan data dari tabel ini (`data_type: "PERIODIC_REPORT"`) sebelum fallback ke `project_progress_history` + `project_budgets`. Tidak ada relasi FK ke `data_submissions` atau `data_lock_periods` — tabel ini berdiri sendiri sebagai input operasional.

> **Update 2026-08-29 — REL-001 Data Governance tables**  
> `data_submission_items` (`entity_id` nullable; action enum CREATE/UPDATE/DELETE/UPSERT/VALIDATE_ONLY; `payload_before`/`payload_after` JSONB; `validation_status` PENDING/VALID/INVALID) dan `data_lock_periods` (`period_month` NULL = full-year; expression unique index `COALESCE(period_month,0)`; `status` OPEN/LOCKED). `data_submissions` diperluas dengan kolom official-approval (P3-003) + `created_by` (P3-003-HARDEN). Rule: UPDATE/DELETE/UPSERT item wajib `entity_id`; CREATE/VALIDATE_ONLY boleh NULL.

> **Update 2026-08-28 — HARDEN-001: `government_external_mappings` schema change**
> `internal_entity_id UUID` diubah dari `NOT NULL` menjadi **nullable** (migration 000028).
> Ditambahkan kolom `match_status VARCHAR(30) NOT NULL DEFAULT 'PENDING_MATCH'`.
> Nilai: `PENDING_MATCH` (external record belum di-resolve ke internal entity),
> `MATCHED` (sudah di-resolve — belum ada mekanisme otomatis, harus lewat reconciliation step).
> Index `idx_gov_map_entity` diubah menjadi partial index `WHERE internal_entity_id IS NOT NULL`.
> Index baru `idx_gov_map_match_status` ditambahkan untuk query PENDING_MATCH yang efisien.
> **Jangan tulis `uuid.New()` sebagai placeholder di `internal_entity_id` — selalu NULL sampai benar-benar resolved.**

> **Update 2026-08-18 — Conceptual vs Implemented**  
> ERD ini adalah model konseptual target. Tidak semua entity di dokumen ini sudah tersedia pada migration/backend/frontend. Gunakan [Implementation Gap Analysis](./IMPLEMENTATION-GAP-ANALYSIS.md) untuk memetakan gap terhadap kebutuhan PMO dari lampiran.

> **Update 2026-08-20 — Control Tower Extension**  
> Model konseptual diperluas untuk PMO National Project Control Tower. Entitas tambahan pada bagian 8 seluruhnya berstatus planned dan belum boleh dianggap sudah dimigrasikan. Lihat [PMO Control Tower Analysis](./PMO-CONTROL-TOWER-ANALYSIS.md).

> **Update 2026-08-21 — UI Shell Does Not Change Data Model**  
> **Update 2026-08-26 — Conceptual vs Implemented (P1-004)**
> P1-004 menambahkan dua tabel implemented: `vendors` (tenant-scoped master party) dan `contracts` (project contract). Nama aktual `vendors`/`contracts` menggantikan konsep `business_parties`/`project_contracts` untuk menghindari over-engineering; relasi dan field inti sama (organization, project, vendor/consultant, contract_number, value, currency, dates, status, scope). Contract value adalah operational input yang belum otomatis mengubah `project.budget_total` dan belum validated/published (menunggu P1-011/P1-012). Dashboard belum mengklaim contract health.

> **Update 2026-08-28 — P3-003 Data Governance (Official Validation & Approval Workflow)**
> Tiga area tabel implemented baru:
> 1. `data_submissions` (migration `000016`/P1-012, diperluas `000030`-`000033`/P3-003): kolom `dataset_type`, `source_type`, `source_entity_type`, `source_entity_id`, `reviewed_by/at`, `review_notes`, `approved_by/at`, `locked_by/at`; `project_id`/`snapshot_id`/`period_month` dibuat nullable (migration `000031`/`000032`) sehingga jalur official tidak wajib terikat project/snapshot dan bisa full-year; constraint status diperluas `DRAFT/SUBMITTED/IN_REVIEW/APPROVED/REJECTED/LOCKED/CANCELLED/VALID/STALE` (legacy `data_submissions_status_check` yang membatasi `DRAFT/SUBMITTED/VALID/REJECTED/STALE` di-drop di `000033` — sebelumnya memblokir IN_REVIEW/APPROVED/LOCKED/CANCELLED dengan SQLSTATE 23514).
> 2. `data_submission_items` (baru, migration `000030`): line item submission — entity_type, entity_id, action (CREATE/UPDATE/DELETE/UPSERT/VALIDATE_ONLY), payload_before/payload_after (JSONB), validation_status (PENDING/VALID/INVALID), validation_errors (JSONB).
> 3. `data_lock_periods` (baru, migration `000030`): period lock — organization_id, dataset_type, period_year, period_month (NULL = full-year), status (OPEN/LOCKED), locked_by/at, lock_reason; unique partial `uq_data_lock_periods_org_dataset` per (org, dataset, year, month) WHERE deleted_at IS NULL.
> Semua query governance wajib tenant-scoped (`organization_id`); entitas soft-deleted dan mapping pemerintah `PENDING_MATCH` ditolak saat validation; approval menyimpan actor + timestamp; data yang di-import/synced tidak otomatis approved.

Dashboard eksekutif dan `/command-center` frontend shell memakai entity/API existing (`projects`, tasks, milestones, issues, risks, budgets, progress history, early warnings). Placeholder validation, Health Score, corrective action, decision, freshness, dan peta ilustratif tidak membuktikan entity konseptual terkait sudah dimigrasikan. Tidak ada perubahan ERD implemented dari pekerjaan UI ini.

---

## Daftar Isi

1. [Konvensi](#1-konvensi)
2. [Core Platform — Entity Diagram](#2-core-platform--entity-diagram)
3. [Project Management — Entity Diagram](#3-project-management--entity-diagram)
4. [Workflow & Approval — Entity Diagram](#4-workflow--approval--entity-diagram)
5. [Definisi Tabel Lengkap](#5-definisi-tabel-lengkap)
6. [Relasi Antar Modul](#6-relasi-antar-modul)
7. [Index Strategy](#7-index-strategy)
8. [PMO Control Tower Extension](#8-pmo-control-tower-extension)

---

## 1. Konvensi

### 1.0 Status Implementasi Data Model

| Domain Data | Status Saat Ini | Catatan |
|-------------|-----------------|---------|
| Organizations & Org Units | Partial | Tabel/model ada; UI/API management org unit belum lengkap |
| Users/Auth/Sessions | Verified/Basic | Login, refresh/me, logout, dan session behavior terverifikasi P0 |
| RBAC Roles/Permissions/User Roles | Verified/Basic | Seed idempotent dan admin `SUPER_ADMIN` terverifikasi; scope detail belum lengkap |
| Projects | Verified/Basic | CRUD, transition, progress history, dan tenant guard dasar terverifikasi |
| Team Members | Partial | Model ada; UI/workload belum lengkap |
| Milestones | Verified/Basic | Model, nested CRUD, transition, dan soft delete dasar terverifikasi |
| Tasks/WBS | Verified/Basic | Model, nested CRUD, subtask dasar, Kanban, dan soft delete dasar terverifikasi; WBS level 3/comments/assignments belum penuh |
| Issues | Implemented/Basic | P1-001 usable: model `issues` + `organization_id` + `escalation`, nested CRUD, FSM, soft delete, audit, tenant guard |
| Risks | Implemented/Basic | P1-002 usable: model `risks` + `organization_id` + probability/impact INT 1-5 + `risk_score` + `severity`, nested CRUD, FSM, soft delete, audit, tenant guard; dashboard/Command Center membaca via `RISK_REGISTER` |
| Budget Lines | Implemented/Basic | P1-003 usable: model `project_budgets` (planned/actual/currency per kategori, soft delete via 000007) + nested CRUD + audit + tenant guard via ownership parent project (`project_budgets` tidak punya `organization_id`); variance/usage_pct/status dihitung backend; dashboard `BUDGET_THRESHOLD` agregat per project |
| Documents | Conceptual/Partial | Model ada sebagian, storage/upload/download belum usable |
| Workflow Definitions/Transitions | Conceptual | Kode saat ini memakai FSM in-memory, bukan full configurable DB workflow |
| Approval Requests | Conceptual/Partial | Model konsep ada; workflow approval belum usable |
| Corrective Action | Missing | Perlu entity baru untuk deviation, root cause, recommendation, PIC, target, follow-up |
| Program/Sector/Location/DAS | Missing | Master data untuk agregasi Level 1-3 belum ada |
| Contract/Vendor/Consultant | Implemented/Basic | P1-004 usable: tabel `vendors` (master party tenant-scoped, type `VENDOR`/`CONSULTANT`) dan `contracts` (per project, `organization_id` + parent project ownership) via migration `000010_vendors_contracts`; nested CRUD contract + master CRUD vendor, contract_number unique per org (partial unique index `uq_contracts_org_number`), validasi enum/nilai/tanggal, soft delete, audit, tenant guard; UI project detail panel "Kontrak" |
| Progress/Financial Snapshots | Missing | Baseline, period snapshot, validation, dan variance belum tersedia |
| Health Score Snapshots | Implemented/Basic | P1-014: `health_formulas` dan `health_snapshots` menyimpan formula version, component score, class, period, explanation, dan audit |
| Validation/Data Quality | Implemented/Basic | P1-012: `data_submissions` tenant-scoped dengan status lifecycle, freshness, completeness, SLA, validator, rejection reason, lineage, dan audit |
| Data Governance (P3-003) | Implemented | 2026-08-28: jalur official di `data_submissions` (status IN_REVIEW/APPROVED/LOCKED/CANCELLED + reviewer/approver/locker audit fields) + `data_submission_items` (line item validation) + `data_lock_periods` (period lock); FSM ketat, lock period unik per org+dataset+period, migration `000030`-`000033` |
| Field Inspection/Evidence | Implemented/Basic | P1-013: `field_inspections`/`field_evidence` dengan koordinat, metadata, checksum, verification, soft delete, dan authorized download |
| Executive Decision/Benefit | Implemented/Basic | `command_escalations`, `executive_decisions`, `benefit_indicators`, dan `benefit_measurements` tersedia tenant-scoped; decision/benefit lifecycle dasar usable; Level 1/2 aggregate resmi masih P2 roadmap |
| Integration Runs | Missing | Source connector, import run, dan error lineage belum tersedia |
| Periodic Reports | Missing | Perlu entity/report snapshot untuk mingguan, bulanan, triwulanan |
| External System Integrations | Missing | Perlu entity/adapter/import log untuk SIRUP/SIMPONI/OM-SPAN atau CSV/Excel import |

| Konvensi | Aturan |
|----------|--------|
| Primary Key | `UUID`, bukan integer/serial |
| Timestamps | Semua tabel: `created_at`, `updated_at`, `deleted_at` (soft delete) |
| Immutable | `audit_logs`: tidak ada soft delete, tidak ada update |
| Organization ID | Semua tabel data utama punya `organization_id` (multi-tenancy ready) |
| Foreign Key | Referensi selalu ke `id` (UUID) tabel lain |
| Naming | snake_case, tabel plural (users, projects), kolom singular (user_id, project_id) |
| Soft Delete | `deleted_at IS NULL` = aktif; `deleted_at IS NOT NULL` = dihapus |

---

## 2. Core Platform — Entity Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  ORGANIZATIONS                                                               │
│  ─────────────                                                               │
│  id              UUID PK                                                     │
│  name            VARCHAR(200) NOT NULL                                       │
│  code            VARCHAR(50) UNIQUE NOT NULL                                 │
│  short_name      VARCHAR(100)                                                │
│  logo_url        TEXT                                                        │
│  address         TEXT                                                        │
│  phone           VARCHAR(50)                                                 │
│  email           VARCHAR(200)                                                │
│  website         VARCHAR(200)                                                │
│  is_active       BOOLEAN DEFAULT TRUE                                        │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  deleted_at      TIMESTAMPTZ                                                 │
└─────────────────────────────────────────────────────────────────────────────┘
                    │
                    │ has many
                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  ORG_UNITS                                                                   │
│  ─────────                                                                   │
│  id              UUID PK                                                     │
│  organization_id UUID FK → organizations.id NOT NULL                        │
│  parent_id       UUID FK → org_units.id (nullable = root unit)              │
│  name            VARCHAR(200) NOT NULL                                       │
│  code            VARCHAR(100) NOT NULL                                       │
│  short_name      VARCHAR(100)                                                │
│  level           INT NOT NULL (1=Kementerian, 2=Ditjen, 3=Dit, 4=Subdit, 5=Unit) │
│  head_user_id    UUID FK → users.id (nullable)                              │
│  description     TEXT                                                        │
│  is_active       BOOLEAN DEFAULT TRUE                                        │
│  sort_order      INT DEFAULT 0                                               │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  deleted_at      TIMESTAMPTZ                                                 │
└─────────────────────────────────────────────────────────────────────────────┘
                    │
                    │ has many
                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  USERS                                                                       │
│  ─────                                                                       │
│  id              UUID PK                                                     │
│  organization_id UUID FK → organizations.id NOT NULL                        │
│  org_unit_id     UUID FK → org_units.id (nullable)                          │
│  employee_id     VARCHAR(100) UNIQUE (NIP untuk ASN)                        │
│  first_name      VARCHAR(100) NOT NULL                                       │
│  last_name       VARCHAR(100)                                                │
│  email           VARCHAR(200) UNIQUE NOT NULL                                │
│  password_hash   VARCHAR(255) NOT NULL                                       │
│  phone           VARCHAR(50)                                                 │
│  job_title       VARCHAR(200)                                                │
│  avatar_url      TEXT                                                        │
│  is_active       BOOLEAN DEFAULT TRUE                                        │
│  must_change_pwd BOOLEAN DEFAULT TRUE                                        │
│  last_login_at   TIMESTAMPTZ                                                 │
│  login_failed    INT DEFAULT 0                                               │
│  locked_until    TIMESTAMPTZ                                                 │
│  created_by      UUID FK → users.id (nullable)                              │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  deleted_at      TIMESTAMPTZ                                                 │
└─────────────────────────────────────────────────────────────────────────────┘
                    │
          ┌─────────┼──────────┐
          ▼         ▼          ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  USER_SESSIONS                                                               │
│  ─────────────                                                               │
│  id              UUID PK                                                     │
│  user_id         UUID FK → users.id NOT NULL                                │
│  jti             VARCHAR(255) UNIQUE NOT NULL (JWT ID)                      │
│  refresh_jti     VARCHAR(255) UNIQUE NOT NULL                               │
│  user_agent      TEXT                                                        │
│  ip_address      VARCHAR(50)                                                 │
│  is_revoked      BOOLEAN DEFAULT FALSE                                       │
│  expires_at      TIMESTAMPTZ NOT NULL                                        │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│  PASSWORD_RESET_TOKENS                                                       │
│  ─────────────────────                                                       │
│  id              UUID PK                                                     │
│  user_id         UUID FK → users.id NOT NULL                                │
│  token_hash      VARCHAR(255) NOT NULL                                       │
│  expires_at      TIMESTAMPTZ NOT NULL                                        │
│  used_at         TIMESTAMPTZ                                                 │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
└─────────────────────────────────────────────────────────────────────────────┘


═══ RBAC ═══

┌─────────────────────────────────────────────────────────────────────────────┐
│  ROLES                                                                       │
│  ─────                                                                       │
│  id              UUID PK                                                     │
│  organization_id UUID FK → organizations.id NOT NULL                        │
│  name            VARCHAR(100) NOT NULL                                       │
│  description     TEXT                                                        │
│  is_system       BOOLEAN DEFAULT FALSE (system role tidak bisa dihapus)     │
│  default_scope   VARCHAR(50) (ALL, ORG_UNIT, ASSIGNED_PROJECT, MEMBER_PROJECT) │
│  is_active       BOOLEAN DEFAULT TRUE                                        │
│  created_by      UUID FK → users.id                                         │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  deleted_at      TIMESTAMPTZ                                                 │
└─────────────────────────────────────────────────────────────────────────────┘
                    │ has many
                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  PERMISSIONS                                                                 │
│  ───────────                                                                 │
│  id              UUID PK                                                     │
│  resource        VARCHAR(100) NOT NULL (project, task, milestone, issue...) │
│  action          VARCHAR(50) NOT NULL (view, create, update, delete, approve, export) │
│  description     TEXT                                                        │
│  UNIQUE(resource, action)                                                    │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│  ROLE_PERMISSIONS (junction)                                                 │
│  ──────────────────                                                          │
│  id              UUID PK                                                     │
│  role_id         UUID FK → roles.id NOT NULL                                │
│  permission_id   UUID FK → permissions.id NOT NULL                          │
│  UNIQUE(role_id, permission_id)                                              │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│  GROUPS                                                                      │
│  ──────                                                                      │
│  id              UUID PK                                                     │
│  organization_id UUID FK → organizations.id NOT NULL                        │
│  name            VARCHAR(100) NOT NULL                                       │
│  description     TEXT                                                        │
│  is_active       BOOLEAN DEFAULT TRUE                                        │
│  created_by      UUID FK → users.id                                         │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  deleted_at      TIMESTAMPTZ                                                 │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│  GROUP_ROLES (junction)                                                      │
│  ──────────────────────                                                      │
│  id              UUID PK                                                     │
│  group_id        UUID FK → groups.id NOT NULL                               │
│  role_id         UUID FK → roles.id NOT NULL                                │
│  UNIQUE(group_id, role_id)                                                   │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│  USER_GROUPS (junction)                                                      │
│  ──────────────────────                                                      │
│  id              UUID PK                                                     │
│  user_id         UUID FK → users.id NOT NULL                                │
│  group_id        UUID FK → groups.id NOT NULL                               │
│  UNIQUE(user_id, group_id)                                                   │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│  USER_ROLES (junction — direct role assignment, tanpa group)                 │
│  ──────────────────────────────────────────────────────────                 │
│  id              UUID PK                                                     │
│  user_id         UUID FK → users.id NOT NULL                                │
│  role_id         UUID FK → roles.id NOT NULL                                │
│  UNIQUE(user_id, role_id)                                                    │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│  USER_SCOPES                                                                 │
│  ──────────                                                                  │
│  id              UUID PK                                                     │
│  user_id         UUID FK → users.id NOT NULL                                │
│  role_id         UUID FK → roles.id NOT NULL                                │
│  scope_type      VARCHAR(50) (ALL, ORG_UNIT, ASSIGNED_PROJECT, MEMBER_PROJECT) │
│  org_unit_id     UUID FK → org_units.id (nullable — diisi jika scope=ORG_UNIT) │
│  UNIQUE(user_id, role_id)                                                    │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()                                   │
└─────────────────────────────────────────────────────────────────────────────┘


═══ AUDIT ═══

┌─────────────────────────────────────────────────────────────────────────────┐
│  AUDIT_LOGS  ← IMMUTABLE: no update, no delete                              │
│  ──────────                                                                  │
│  id              UUID PK                                                     │
│  organization_id UUID NOT NULL                                               │
│  user_id         UUID (nullable — system actions)                           │
│  user_name       VARCHAR(200)                                                │
│  user_ip         VARCHAR(50)                                                 │
│  user_agent      TEXT                                                        │
│  action          VARCHAR(50) NOT NULL                                        │
│                  (CREATE, UPDATE, DELETE, APPROVE, REJECT,                  │
│                   LOGIN, LOGOUT, FAILED_LOGIN, EXPORT,                      │
│                   ROLE_ASSIGN, ROLE_REVOKE, PERMISSION_CHANGE)              │
│  resource_type   VARCHAR(100) NOT NULL                                       │
│  resource_id     UUID                                                        │
│  resource_name   VARCHAR(500)                                                │
│  field_changed   VARCHAR(200)                                                │
│  old_value       TEXT                                                        │
│  new_value       TEXT                                                        │
│  metadata        JSONB                                                       │
│  created_at      TIMESTAMPTZ DEFAULT NOW() ← NO deleted_at                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Project Management — Entity Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  PROJECTS                                                                    │
│  ────────                                                                    │
│  id              UUID PK                                                     │
│  organization_id UUID FK → organizations.id NOT NULL                        │
│  org_unit_id     UUID FK → org_units.id NOT NULL                            │
│  code            VARCHAR(100) UNIQUE NOT NULL                                │
│  name            VARCHAR(500) NOT NULL                                       │
│  description     TEXT                                                        │
│  type            VARCHAR(100) (Software, Infrastruktur, Pengadaan, dll)     │
│  category        VARCHAR(100)                                                │
│  tags            TEXT[] (array of tags)                                      │
│  thumbnail_url   TEXT                                                        │
│  status          VARCHAR(50) NOT NULL DEFAULT 'DRAFT'                        │
│                  (DRAFT, SUBMITTED, REVIEWED, APPROVED,                     │
│                   ACTIVE, ON_HOLD, COMPLETED, CLOSED)                       │
│  health          VARCHAR(20) DEFAULT 'ON_TRACK'                              │
│                  (ON_TRACK, AT_RISK, DELAYED)                               │
│  priority        VARCHAR(20) DEFAULT 'MEDIUM'                                │
│                  (LOW, MEDIUM, HIGH, CRITICAL)                               │
│  planned_start   DATE NOT NULL                                               │
│  planned_end     DATE NOT NULL                                               │
│  actual_start    DATE                                                        │
│  actual_end      DATE                                                        │
│  progress_manual INT DEFAULT 0 (0-100, input manual)                        │
│  progress_auto   INT DEFAULT 0 (0-100, hitung dari task)                    │
│  use_auto_progress BOOLEAN DEFAULT FALSE                                     │
│  client_name     VARCHAR(200)                                                │
│  client_contact  VARCHAR(200)                                                │
│  pm_user_id      UUID FK → users.id (Project Manager)                       │
│  budget_total    DECIMAL(20,2) DEFAULT 0                                     │
│  budget_actual   DECIMAL(20,2) DEFAULT 0                                     │
│  project_value   DECIMAL(20,2) DEFAULT 0                                     │
│  currency        VARCHAR(10) DEFAULT 'IDR'                                   │
│  notes           TEXT                                                        │
│  created_by      UUID FK → users.id NOT NULL                                │
│  updated_by      UUID FK → users.id                                         │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  deleted_at      TIMESTAMPTZ                                                 │
└─────────────────────────────────────────────────────────────────────────────┘
                    │
          ┌─────────┼──────────────────────────────┐
          ▼         ▼         ▼         ▼           ▼
    PROJECT_    MILESTONES  TASKS    ISSUES       RISKS
    TEAMS
    
┌──────────────────────────────────────────────────┐
│  PROJECT_TEAMS                                    │
│  ─────────────                                    │
│  id          UUID PK                              │
│  project_id  UUID FK → projects.id NOT NULL       │
│  user_id     UUID FK → users.id NOT NULL          │
│  role        VARCHAR(50) NOT NULL                 │
│              (PM, OFFICER, QA, REVIEWER, VIEWER)  │
│  joined_at   TIMESTAMPTZ DEFAULT NOW()            │
│  left_at     TIMESTAMPTZ                          │
│  UNIQUE(project_id, user_id)                      │
│  created_at  TIMESTAMPTZ DEFAULT NOW()            │
└──────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────┐
│  MILESTONES                                       │
│  ──────────                                       │
│  id              UUID PK                          │
│  project_id      UUID FK → projects.id NOT NULL   │
│  organization_id UUID FK NOT NULL                 │
│  name            VARCHAR(500) NOT NULL            │
│  description     TEXT                             │
│  target_date     DATE NOT NULL                    │
│  actual_date     DATE                             │
│  status          VARCHAR(50) DEFAULT 'UPCOMING'   │
│                  (UPCOMING, IN_PROGRESS,          │
│                   ACHIEVED, MISSED)               │
│  deliverables    TEXT                             │
│  evidence_url    TEXT                             │
│  sort_order      INT DEFAULT 0                    │
│  created_by      UUID FK → users.id              │
│  created_at      TIMESTAMPTZ DEFAULT NOW()        │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()        │
│  deleted_at      TIMESTAMPTZ                      │
└──────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────┐
│  TASKS (WBS)                                      │
│  ────────────                                     │
│  id              UUID PK                          │
│  project_id      UUID FK → projects.id NOT NULL   │
│  organization_id UUID FK NOT NULL                 │
│  parent_id       UUID FK → tasks.id (nullable)    │
│  wbs_level       INT DEFAULT 1 (1, 2, or 3)       │
│  wbs_code        VARCHAR(50) (1.1, 1.1.2, dll)   │
│  title           VARCHAR(500) NOT NULL            │
│  description     TEXT                             │
│  status          VARCHAR(50) DEFAULT 'BACKLOG'    │
│                  (BACKLOG, ANALYSIS, DEVELOPMENT, │
│                   QA, UAT, PRODUCTION, DONE)      │
│  priority        VARCHAR(20) DEFAULT 'MEDIUM'     │
│  category        VARCHAR(100)                     │
│  planned_start   DATE                             │
│  planned_end     DATE                             │
│  actual_start    DATE                             │
│  actual_end      DATE                             │
│  est_hours       DECIMAL(10,2)                    │
│  actual_hours    DECIMAL(10,2)                    │
│  progress        INT DEFAULT 0 (0-100)            │
│  is_blocked      BOOLEAN DEFAULT FALSE            │
│  blocked_reason  TEXT                             │
│  sort_order      INT DEFAULT 0                    │
│  milestone_id    UUID FK → milestones.id          │
│  created_by      UUID FK → users.id              │
│  updated_by      UUID FK → users.id              │
│  created_at      TIMESTAMPTZ DEFAULT NOW()        │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()        │
│  deleted_at      TIMESTAMPTZ                      │
└──────────────────────────────────────────────────┘
                    │
                    ▼
┌──────────────────────────────────────────────────┐
│  TASK_ASSIGNMENTS                                 │
│  ────────────────                                 │
│  id          UUID PK                              │
│  task_id     UUID FK → tasks.id NOT NULL          │
│  user_id     UUID FK → users.id NOT NULL          │
│  is_lead     BOOLEAN DEFAULT FALSE                │
│  UNIQUE(task_id, user_id)                         │
│  created_at  TIMESTAMPTZ DEFAULT NOW()            │
└──────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────┐
│  TASK_COMMENTS                                    │
│  ─────────────                                    │
│  id          UUID PK                              │
│  task_id     UUID FK → tasks.id NOT NULL          │
│  user_id     UUID FK → users.id NOT NULL          │
│  content     TEXT NOT NULL                        │
│  created_at  TIMESTAMPTZ DEFAULT NOW()            │
│  updated_at  TIMESTAMPTZ DEFAULT NOW()            │
│  deleted_at  TIMESTAMPTZ                          │
└──────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────┐
│  ISSUES                                           │
│  ──────                                           │
│  id              UUID PK                          │
│  project_id      UUID FK → projects.id NOT NULL   │
│  organization_id UUID FK NOT NULL                 │
│  title           VARCHAR(500) NOT NULL            │
│  description     TEXT                             │
│  severity        VARCHAR(20) DEFAULT 'MEDIUM'     │
│                  (LOW, MEDIUM, HIGH, CRITICAL)    │
│  status          VARCHAR(50) DEFAULT 'OPEN'       │
│                  (OPEN, IN_PROGRESS,              │
│                   RESOLVED, CLOSED)               │
│  impact          TEXT                             │
│  pic_user_id     UUID FK → users.id              │
│  due_date        DATE                             │
│  resolved_at     TIMESTAMPTZ                      │
│  resolution_note TEXT                             │
│  is_escalated    BOOLEAN DEFAULT FALSE            │
│  created_by      UUID FK → users.id              │
│  created_at      TIMESTAMPTZ DEFAULT NOW()        │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()        │
│  deleted_at      TIMESTAMPTZ                      │
└──────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────┐
│  RISKS                                            │
│  ─────                                            │
│  id              UUID PK                          │
│  project_id      UUID FK → projects.id NOT NULL   │
│  organization_id UUID FK NOT NULL                 │
│  title           VARCHAR(500) NOT NULL            │
│  description     TEXT                             │
│  probability     INT NOT NULL (1-5)               │
│  impact          INT NOT NULL (1-5)               │
│  risk_score      INT NOT NULL (probability×impact)│
│  severity        VARCHAR(20) (LOW, MEDIUM, HIGH, CRITICAL) │
│  status          VARCHAR(50) DEFAULT 'IDENTIFIED' │
│                  (IDENTIFIED, MONITORED,          │
│                   MITIGATED, CLOSED)              │
│  mitigation_plan TEXT                             │
│  contingency_plan TEXT                            │
│  pic_user_id     UUID FK → users.id              │
│  due_date        DATE                             │
│  mitigated_at    TIMESTAMPTZ                      │
│  created_by      UUID FK → users.id              │
│  created_at      TIMESTAMPTZ DEFAULT NOW()        │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()        │
│  deleted_at      TIMESTAMPTZ                      │
└──────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────┐
│  PROJECT_BUDGETS                                  │
│  ───────────────                                  │
│  id              UUID PK                          │
│  project_id      UUID FK → projects.id NOT NULL   │
│  organization_id UUID FK NOT NULL                 │
│  component       VARCHAR(200) NOT NULL            │
│                  (Tenaga Ahli, Infrastruktur,     │
│                   Operasional, Perjalanan, Lain)  │
│  planned_amount  DECIMAL(20,2) DEFAULT 0          │
│  actual_amount   DECIMAL(20,2) DEFAULT 0          │
│  notes           TEXT                             │
│  created_by      UUID FK → users.id              │
│  created_at      TIMESTAMPTZ DEFAULT NOW()        │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()        │
│  deleted_at      TIMESTAMPTZ                      │
└──────────────────────────────────────────────────┘

NOTES (P1-003, aktual implementasi): tabel `project_budgets` TIDAK punya kolom `organization_id` — tenant safety dijamin melalui ownership parent project (`projects.organization_id`). Kolom aktual: `id`, `project_id`, `category`, `description`, `planned`, `actual`, `currency` (default IDR), `created_by`, `created_at`, `updated_at`, `deleted_at` (dari migration 000007). Tidak ada kolom `component`/`planned_amount`/`actual_amount`/`notes` seperti konsep di atas. Response API menghitung `variance` (planned − actual), `usage_pct` (actual/planned ×100), dan `status` (NORMAL/WATCH/RISK/OVERRUN) selalu di backend (round 2); field ini tidak disimpan di DB.

┌──────────────────────────────────────────────────┐
│  PROJECT_DOCUMENTS                                │
│  ─────────────────                                │
│  id              UUID PK                          │
│  project_id      UUID FK → projects.id NOT NULL   │
│  organization_id UUID FK NOT NULL                 │
│  task_id         UUID FK → tasks.id (nullable)    │
│  name            VARCHAR(500) NOT NULL            │
│  doc_type        VARCHAR(100)                     │
│                  (TOR, KAK, CONTRACT, BAST,       │
│                   REPORT, PHOTO, OTHER)           │
│  file_key        TEXT NOT NULL (MinIO object key) │
│  file_name       VARCHAR(500) NOT NULL            │
│  file_size       BIGINT                           │
│  mime_type       VARCHAR(200)                     │
│  version         VARCHAR(50) DEFAULT '1.0'        │
│  description     TEXT                             │
│  uploaded_by     UUID FK → users.id NOT NULL      │
│  created_at      TIMESTAMPTZ DEFAULT NOW()        │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()        │
│  deleted_at      TIMESTAMPTZ                      │
└──────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────┐
│  PROJECT_MEETINGS                                 │
│  ────────────────                                 │
│  id              UUID PK                          │
│  project_id      UUID FK → projects.id NOT NULL   │
│  organization_id UUID FK NOT NULL                 │
│  title           VARCHAR(500) NOT NULL            │
│  meeting_date    TIMESTAMPTZ NOT NULL             │
│  location        VARCHAR(500)                     │
│  attendees       TEXT[] (user IDs atau nama)      │
│  agenda          TEXT                             │
│  minutes         TEXT                             │
│  action_items    TEXT                             │
│  created_by      UUID FK → users.id              │
│  created_at      TIMESTAMPTZ DEFAULT NOW()        │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()        │
│  deleted_at      TIMESTAMPTZ                      │
└──────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────┐
│  PROJECT_PROGRESS_HISTORY                         │
│  ────────────────────────                         │
│  id              UUID PK                          │
│  project_id      UUID FK → projects.id NOT NULL   │
│  progress        INT NOT NULL                     │
│  health          VARCHAR(20) NOT NULL             │
│  note            TEXT                             │
│  recorded_by     UUID FK → users.id              │
│  recorded_at     TIMESTAMPTZ DEFAULT NOW()        │
└──────────────────────────────────────────────────┘
```

---

## 4. Workflow & Approval — Entity Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  WORKFLOW_DEFINITIONS                                                        │
│  ────────────────────                                                        │
│  id              UUID PK                                                     │
│  organization_id UUID FK NOT NULL                                            │
│  entity_type     VARCHAR(100) NOT NULL (project, task, issue, risk)         │
│  name            VARCHAR(200) NOT NULL                                       │
│  description     TEXT                                                        │
│  is_active       BOOLEAN DEFAULT TRUE                                        │
│  UNIQUE(organization_id, entity_type)                                        │
│  created_at      TIMESTAMPTZ DEFAULT NOW()                                   │
│  updated_at      TIMESTAMPTZ DEFAULT NOW()                                   │
└─────────────────────────────────────────────────────────────────────────────┘
                    │ has many
                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  WORKFLOW_TRANSITIONS                                                        │
│  ────────────────────                                                        │
│  id                  UUID PK                                                 │
│  workflow_id         UUID FK → workflow_definitions.id NOT NULL             │
│  from_status         VARCHAR(50) NOT NULL                                    │
│  to_status           VARCHAR(50) NOT NULL                                    │
│  action_label        VARCHAR(100) NOT NULL (Submit, Approve, Reject, dll)   │
│  allowed_roles       TEXT[] NOT NULL (array of role IDs)                    │
│  require_approval    BOOLEAN DEFAULT FALSE                                   │
│  approver_roles      TEXT[] (array of role IDs yang bisa approve)           │
│  require_comment     BOOLEAN DEFAULT FALSE                                   │
│  notify_roles        TEXT[] (role IDs yang di-notify setelah transisi)      │
│  UNIQUE(workflow_id, from_status, to_status)                                 │
│  created_at          TIMESTAMPTZ DEFAULT NOW()                               │
│  updated_at          TIMESTAMPTZ DEFAULT NOW()                               │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│  APPROVAL_REQUESTS                                                           │
│  ─────────────────                                                           │
│  id                  UUID PK                                                 │
│  organization_id     UUID FK NOT NULL                                        │
│  workflow_id         UUID FK → workflow_definitions.id NOT NULL             │
│  transition_id       UUID FK → workflow_transitions.id NOT NULL             │
│  entity_type         VARCHAR(100) NOT NULL                                   │
│  entity_id           UUID NOT NULL                                           │
│  entity_name         VARCHAR(500)                                            │
│  from_status         VARCHAR(50) NOT NULL                                    │
│  to_status           VARCHAR(50) NOT NULL                                    │
│  requested_by        UUID FK → users.id NOT NULL                            │
│  status              VARCHAR(50) DEFAULT 'PENDING'                           │
│                      (PENDING, APPROVED, REJECTED, CANCELLED)               │
│  comment             TEXT                                                    │
│  decided_by          UUID FK → users.id                                     │
│  decided_at          TIMESTAMPTZ                                             │
│  decision_note       TEXT                                                    │
│  expires_at          TIMESTAMPTZ                                             │
│  created_at          TIMESTAMPTZ DEFAULT NOW()                               │
│  updated_at          TIMESTAMPTZ DEFAULT NOW()                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 5. Definisi Tabel Lengkap

### Ringkasan Tabel

| # | Tabel | Modul | Keterangan |
|---|-------|-------|-----------|
| 1 | organizations | Core | Entitas organisasi (kementerian) |
| 2 | org_units | Core | Hierarki unit organisasi |
| 3 | users | Core | Pengguna sistem |
| 4 | user_sessions | Core | JWT session tracking |
| 5 | password_reset_tokens | Core | Token reset password |
| 6 | roles | Core/RBAC | Role yang bisa dikonfigurasi |
| 7 | permissions | Core/RBAC | Definisi permission per resource+action |
| 8 | role_permissions | Core/RBAC | Junction: role ↔ permission |
| 9 | groups | Core/RBAC | Group user |
| 10 | group_roles | Core/RBAC | Junction: group ↔ role |
| 11 | user_groups | Core/RBAC | Junction: user ↔ group |
| 12 | user_roles | Core/RBAC | Direct role assignment ke user |
| 13 | user_scopes | Core/RBAC | Scope akses per user per role |
| 14 | audit_logs | Core | Immutable audit trail |
| 15 | workflow_definitions | Core/Workflow | Definisi FSM per entity type |
| 16 | workflow_transitions | Core/Workflow | Transisi valid per workflow |
| 17 | approval_requests | Core/Workflow | Request approval dengan status |
| 18 | projects | Project | Entitas project utama |
| 19 | project_teams | Project | Anggota tim project |
| 20 | milestones | Project | Milestone project |
| 21 | tasks | Project | WBS / Task (hirarki 3 level) |
| 22 | task_assignments | Project | Assignment task ke user |
| 23 | task_comments | Project | Komentar di task |
| 24 | issues | Project | Issue/kendala project |
| 25 | risks | Project | Risiko project |
| 26 | project_budgets | Project | Budget per komponen |
| 27 | project_documents | Project | Dokumen project (MinIO) |
| 28 | project_meetings | Project | Meeting & notulen |
| 29 | project_progress_history | Project | Riwayat progress |
| 30 | programs | PMO Control Tower | Master program nasional |
| 31 | sectors | PMO Control Tower | Master sektor SDA |
| 32 | administrative_regions | PMO Control Tower | Provinsi/kabupaten/kota hierarkis |
| 33 | watersheds | PMO Control Tower | Master sungai/DAS |
| 34 | project_locations | Project/Spatial | Lokasi, koordinat, wilayah, dan DAS proyek |
| 35 | business_parties | Project/Contract | Penyedia, kontraktor, dan konsultan |
| 36 | project_parties | Project/Contract | Peran party pada proyek |
| 37 | project_contracts | Project/Contract | Kontrak, nilai, tanggal, status, dan procurement state |
| 38 | progress_baselines | Project/Monitoring | Baseline fisik, keuangan, dan jadwal |
| 39 | progress_snapshots | Project/Monitoring | Snapshot periodik planned versus actual |
| 40 | financial_realizations | Project/Finance | Realisasi keuangan per periode/komponen |
| 41 | health_formula_versions | Analytics | Formula, bobot, threshold, dan approval |
| 42 | project_health_snapshots | Analytics | Hasil health per proyek/periode |
| 43 | health_component_scores | Analytics | Nilai dan explanation per dimensi |
| 44 | alerts | Command Center | Alert actionable dan escalation state |
| 45 | corrective_actions | Command Center | Akar masalah, rekomendasi, PIC, SLA, evidence |
| 46 | executive_decisions | Decision Support | Arahan/keputusan dan tindak lanjut pimpinan |
| 47 | data_submissions | Data Quality | Submission data manual/file/integrasi |
| 48 | data_quality_results | Data Quality | Validation, completeness, freshness, dan rejection |
| 49 | field_inspections | Field | Inspeksi, waktu, lokasi, petugas, dan verifikasi |
| 50 | field_evidence | Field/Document | Foto/dokumen evidence dan checksum |
| 51 | reporting_snapshots | Reporting | Snapshot laporan yang dipublikasikan |
| 52 | project_benefits | Outcome | Definisi target manfaat proyek/program |
| 53 | benefit_measurements | Outcome | Realisasi manfaat per periode |
| 54 | integration_sources | Integration | Konfigurasi sumber dan adapter |
| 55 | integration_runs | Integration | Status import/sync, correlation, dan statistik |
| 56 | data_submission_items | Data Governance (P3-003) | Line item submission resmi: entity + action + payload + validation status |
| 57 | data_lock_periods | Data Governance (P3-003) | Period lock per org+dataset+periode, status OPEN/LOCKED |

> **Tabel implemented aktual P3-003 (migration `000030`-`000033`):** `data_submission_items` (submission_id FK CASCADE, entity_type, entity_id, action CREATE/UPDATE/DELETE/UPSERT/VALIDATE_ONLY, payload_before/payload_after JSONB, validation_status PENDING/VALID/INVALID, validation_errors JSONB) dan `data_lock_periods` (organization_id, dataset_type, period_year, period_month nullable, status OPEN/LOCKED, locked_by/at, lock_reason, created_by; unique partial per (org, dataset, year, month)). `data_submissions` diperluas: `dataset_type`/`source_type` (default MANUAL), `source_entity_type`/`source_entity_id` (link ke submission lain), `reviewed_by/at` + `review_notes`, `approved_by/at`, `locked_by/at`; `project_id`/`snapshot_id`/`period_month` nullable; status check mencakup IN_REVIEW/APPROVED/LOCKED/CANCELLED. Dataset types: PROJECT_PROGRESS, BUDGET, RISK, ISSUE, BENEFIT, LOCATION, CONTRACT, DOCUMENT, OTHER. Sumber: MANUAL, CSV_IMPORT, PRIMAVERA, GOVERNMENT, BIM, API.

> **Tabel implemented aktual P1-004 (migration `000010_vendors_contracts`):** `vendors` dan `contracts` menggantikan konsep `business_parties`/`project_parties`/`project_contracts` untuk kebutuhan MVP kontrak. `vendors` = master party (organization_id, name, type VENDOR/CONSULTANT, legal_name, tax_id, contact_person, email, phone, address, is_active, created_by, deleted_at). `contracts` = kontrak per project (organization_id, project_id, contract_number unique per org via `uq_contracts_org_number`, title, vendor_id wajib, consultant_id opsional, contract_value, currency default IDR, signed_date, start_date, end_date, status DRAFT/ACTIVE/AMENDED/COMPLETED/TERMINATED, scope_of_work, created_by, deleted_at). `task_comments`/`project_meetings` belum dimigrasikan.

---

## 6. Relasi Antar Modul

```
organizations
    ├── org_units (hierarki, parent_id self-referencing)
    ├── users
    │   ├── user_sessions
    │   ├── password_reset_tokens
    │   ├── user_groups → groups → group_roles → roles
    │   ├── user_roles → roles
    │   └── user_scopes
    ├── roles
    │   └── role_permissions → permissions
    ├── groups
    │   └── group_roles → roles
    ├── workflow_definitions
    │   ├── workflow_transitions
    │   └── approval_requests
    └── projects
        ├── project_teams → users
        ├── milestones
        ├── tasks (WBS, recursive parent_id)
        │   ├── task_assignments → users
        │   └── task_comments → users
        ├── issues
        ├── risks
        ├── project_budgets
        ├── project_documents
        ├── project_meetings
        └── project_progress_history
```

---

## 7. Index Strategy

```sql
-- organizations
CREATE UNIQUE INDEX idx_organizations_code ON organizations(code) WHERE deleted_at IS NULL;

-- org_units
CREATE INDEX idx_org_units_organization_id ON org_units(organization_id);
CREATE INDEX idx_org_units_parent_id ON org_units(parent_id);
CREATE INDEX idx_org_units_level ON org_units(organization_id, level);

-- users
CREATE UNIQUE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_organization_id ON users(organization_id);
CREATE INDEX idx_users_org_unit_id ON users(org_unit_id);
CREATE INDEX idx_users_is_active ON users(is_active) WHERE deleted_at IS NULL;

-- user_sessions
CREATE UNIQUE INDEX idx_user_sessions_jti ON user_sessions(jti);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);

-- roles
CREATE INDEX idx_roles_organization_id ON roles(organization_id);

-- user_roles
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);

-- projects
CREATE INDEX idx_projects_organization_id ON projects(organization_id);
CREATE INDEX idx_projects_org_unit_id ON projects(org_unit_id);
CREATE INDEX idx_projects_status ON projects(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_health ON projects(health) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_pm_user_id ON projects(pm_user_id);

-- project_teams
CREATE INDEX idx_project_teams_project_id ON project_teams(project_id);
CREATE INDEX idx_project_teams_user_id ON project_teams(user_id);

-- tasks
CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);
CREATE INDEX idx_tasks_status ON tasks(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_planned_end ON tasks(planned_end) WHERE deleted_at IS NULL;

-- issues
CREATE INDEX idx_issues_project_id ON issues(project_id);
CREATE INDEX idx_issues_severity_status ON issues(severity, status) WHERE deleted_at IS NULL;

-- risks
CREATE INDEX idx_risks_project_id ON risks(project_id);
CREATE INDEX idx_risks_risk_score ON risks(risk_score) WHERE deleted_at IS NULL;

-- audit_logs
CREATE INDEX idx_audit_logs_organization_id ON audit_logs(organization_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
```

---

## 8. PMO Control Tower Extension

### 8.1 Invariant untuk Seluruh Entitas Baru

Seluruh business entity pada bagian ini wajib memiliki:

- `id UUID` sebagai primary key.
- `organization_id UUID NOT NULL` dan foreign key tenant yang sesuai.
- `created_at`, `created_by`, `updated_at`, dan `updated_by` bila entity dapat dimutasi user.
- `deleted_at` untuk soft delete, kecuali snapshot/audit immutable.
- Index tenant-first, misalnya `(organization_id, project_id, period_date)`.
- Tenant consistency pada setiap foreign key/join di repository/service.

Snapshot yang sudah dipublikasikan tidak dihapus atau diubah; koreksi menghasilkan version/snapshot baru dan hubungan supersedes jika diperlukan.

### 8.2 Portfolio, Program, dan Spatial Model

```text
programs ──< projects >── sectors
                    |
                    └──< project_locations >── administrative_regions
                              |
                              └── watersheds
```

`projects` direncanakan memiliki `program_id` dan `sector_id` atau junction bila satu proyek benar-benar lintas program. `project_locations` menyimpan latitude/longitude, alamat, region, watershed, dan geometry reference opsional. PostGIS belum wajib pada tahap point-map awal.

### 8.3 Contract and Party Model

```text
projects ──< project_contracts
    |
    └──< project_parties >── business_parties
```

`business_parties.party_type` membedakan contractor, consultant, supervisor, supplier, atau pihak lain. `project_contracts` menyimpan contract number, title, value, signed/start/end date, status, procurement status, dan primary party. Perubahan nilai/tanggal kontrak harus memiliki audit/change history saat implementasi.

### 8.4 Baseline, Progress, dan Financial Model

```text
projects ──< progress_baselines
    |
    ├──< progress_snapshots
    └──< financial_realizations
```

`progress_snapshots` menyimpan `period_date`, planned/actual physical percentage, planned/actual financial percentage, planned/actual schedule percentage, source submission, validation status, dan calculated variance. Unique key minimal `(organization_id, project_id, period_date, version)`.

Nilai agregat dashboard harus berasal dari snapshot valid pada cut-off yang sama. Nilai terkini di `projects` dapat dipertahankan sebagai operational convenience tetapi bukan satu-satunya sumber histori.

### 8.5 Health and Alert Model

```text
health_formula_versions ──< project_health_snapshots ──< health_component_scores
                                      |
                                      └──< alerts ──< corrective_actions
```

`health_formula_versions` menyimpan status draft/approved/retired, effective period, dimensions, weights, thresholds, missing-data rule, approver, dan approved_at. `project_health_snapshots` menyimpan `GREEN`, `YELLOW`, `RED`, atau `CRITICAL`, score total, explanation, period, dan formula version.

Alert menyimpan type, severity, detected_at, aging basis, source entity, source snapshot, PIC, due date, escalation level, status, dan resolved_at. Corrective action terhubung ke alert/issue/risk dan menyimpan deviation, root cause, recommendation, owner, target date, progress, verification, serta evidence.

### 8.6 Data Quality and Integration Model

```text
integration_sources ──< integration_runs ──< data_submissions ──< data_quality_results
                                                    │
                                                    ├──< data_submission_items   (P3-003: line-item validation)
                                                    └──< data_lock_periods       (P3-003: period lock)
                                                    |
                                                    └── domain records/snapshots
```

`data_submissions` mempunyai source type, source reference, period, payload/file reference, submitter, submitted_at, validation status, validator, validated_at, rejection reason, dan lineage metadata. `data_quality_results` menyimpan rule code/version, dimension, result, severity, measured value, dan message.

> **Status aktual 2026-08-28:** dua jalur memakai `data_submissions`. (a) Jalur validation queue (P1-012, module `dataquality`): `snapshot_id IS NOT NULL`, status DRAFT/SUBMITTED/VALID/REJECTED/STALE, terikat project + snapshot. (b) Jalur official (P3-003, module `governance`): `snapshot_id IS NULL`, status DRAFT→SUBMITTED→IN_REVIEW→APPROVED→LOCKED (+REJECTED/CANCELLED), `dataset_type`/`source_type` mengidentifikasi dataset, items di `data_submission_items` membawa validation per entity, dan `data_lock_periods` membatasi penulisan pada periode yang sudah terkunci. Query `dataquality` selalu menambah `snapshot_id IS NOT NULL` agar row governance tidak bocor ke validation queue.

Raw credential/secrets tidak disimpan pada tabel domain. Konfigurasi rahasia connector memakai environment/secret store.

### 8.7 Field Evidence, Reporting, Decision, dan Benefit

```text
projects ──< field_inspections ──< field_evidence
    |
    ├──< reporting_snapshots
    ├──< executive_decisions
    └──< project_benefits ──< benefit_measurements
```

Evidence menyimpan object key, media type, size, checksum, captured_at, coordinate, uploader, dan verification status. Reporting snapshot menyimpan period/frequency, publication version, KPI payload/reference, approver, published_at, dan source cut-off.

Executive decision menyimpan subject, decision text, owner, due date, status, source escalation/project, decided_by, decided_at, dan audit metadata. Benefit indicator menyimpan unit, aggregation method, owner, source, optional project, dan description; benefit measurement menyimpan baseline, target, actual, period, source, dan validation status. Aggregate hanya memakai measurement `VALID`; unit berbeda tidak dijumlahkan dalam summary.

### 8.8 Planned Indexes and Constraints

- Unique active master code per tenant untuk program, sector, region, watershed, party, dan integration source.
- `(organization_id, program_id, sector_id, status)` untuk portfolio filter.
- `(organization_id, period_date, validation_status)` untuk snapshot/read model.
- `(organization_id, health_class, period_date)` untuk ranking/watchlist.
- `(organization_id, status, severity, due_date)` untuk alert/corrective action.
- `(organization_id, validation_status, submitted_at)` untuk validation queue.
- Spatial index baru ditambahkan bersama PostGIS migration, bukan sebelum kebutuhan polygon/intersection disetujui.
