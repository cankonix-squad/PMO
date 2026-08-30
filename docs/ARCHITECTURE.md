# Technical Architecture
# CANKORA — Enterprise Operations Platform

**Versi**: 0.2.4  
**Tanggal**: 2026-08-29  
**Dibuat oleh**: Cankonix  
**Status**: Draft — target architecture, not full implementation proof

> **Update 2026-08-29 — CANKORA-DASH-002: Periodic Reports Data Source**
> Dashboard trend `/api/v1/dashboard/trend` data source diprioritaskan dari `project_periodic_reports` (migration `000038`). Logika: (1) Jika ada data periodik dalam 12 bulan terakhir → `data_type: "PERIODIC_REPORT"` — aggregate per bulan: `AVG(physical_progress_pct)` across active projects, `SUM(financial_actual)/SUM(financial_planned)*100`; (2) Fallback ke `project_progress_history` + `project_budgets` (`data_type: "OPERATIONAL"`); (3) Fallback terakhir: operational ramp dari current progress/budget aggregate. Tenant-safe: semua query JOIN ke `projects p ON p.organization_id = ? AND p.deleted_at IS NULL`. Frontend disclaimer berbeda per `data_type`: biru untuk PERIODIC_REPORT, amber untuk OPERATIONAL. Handler pattern baru: `Handler.WithDB(db *gorm.DB) *Handler` untuk akses DB langsung di project package tanpa bloating Repository interface.

> **Update 2026-08-29 — UAT-002 Report Export Real File**  
> Export request queue sekarang menghasilkan file nyata (CSV & XLSX). `ProcessExportRequest()` dijalankan synchronous setelah `CreateExportRequest` — status langsung COMPLETED (bukan async queue). File disimpan di `backend/storage/reports/<orgID>/<date>/<filename>` (relative `storage_key`, path-traversal safe). Kolom baru `file_name`, `storage_key`, `mime_type`, `file_size`, `generated_at` ditambahkan ke `report_export_requests` via migration `000036`. Download endpoint `GET /api/v1/analytics/reports/export/requests/:id/download` tenant-safe (org guard), authenticated, streaming file dengan `Content-Type` dan `Content-Disposition`. Audit events: `report.export.requested`, `report.export.completed`, `report.export.failed`, `report.export.downloaded`. XLSX via `github.com/xuri/excelize/v2 v2.8.1`. 6 dataset: executive-summary, project-performance, risk-issue, budget, benefits, priority. Smoke `docs/smoke-test-uat-002-report-export.sh` 44/45 PASS (1 SKIP: full cross-tenant test memerlukan 2 org).

> **Update 2026-08-29 — REL-001 Governance & Read-Model Boundaries**  
> Data governance module (`/api/v1/governance/`) menyediakan official validation workflow (DRAFT→SUBMITTED→IN_REVIEW→APPROVED→LOCKED) dengan per-item validation tenant-scoped dan lock-period. Read-model boundaries: `/dashboard` & `/command-center` = operational current-state; `/executive` (Level 1) & `/programs` (Level 2) = operational aggregate; `/reports/analytics` = read-model; `/governance` = official governed data. Power BI hanya membaca read model tenant-safe. Regression gate: `docs/smoke-test-rel-001-regression.sh`.

> **Update 2026-08-18 — Architecture Sync Note**  
> Dokumen ini masih berisi arsitektur target. Beberapa bagian belum sama dengan kode aktual atau belum usable end-to-end. Lihat [Implementation Gap Analysis](./IMPLEMENTATION-GAP-ANALYSIS.md) sebelum memakai dokumen ini sebagai status delivery.

---

## Daftar Isi

1. [Architectural Decisions](#1-architectural-decisions)
2. [High-Level Architecture](#2-high-level-architecture)
3. [Backend — Modular Monolith](#3-backend--modular-monolith)
4. [Frontend — Next.js 14 App Router](#4-frontend--nextjs-14-app-router)
5. [Database Architecture](#5-database-architecture)
6. [Authentication & Authorization](#6-authentication--authorization)
7. [File Storage](#7-file-storage)
8. [Caching & Background Processing](#8-caching--background-processing)
9. [Notification Architecture](#9-notification-architecture)
10. [Audit Trail Architecture](#10-audit-trail-architecture)
11. [Workflow Engine Architecture](#11-workflow-engine-architecture)
12. [Deployment Architecture](#12-deployment-architecture)
13. [Environment Strategy](#13-environment-strategy)
14. [Future Scalability Path](#14-future-scalability-path)
15. [PMO National Project Control Tower Architecture](#15-pmo-national-project-control-tower-architecture)

---

## 1. Architectural Decisions

### 1.0 Status Implementasi Saat Ini

| Area | Status 2026-08-20 | Catatan |
|------|-------------------|---------|
| Modular Monolith | Implemented | Struktur backend mengikuti modular monolith |
| Repository + Service Layer | Partial | Masih ada beberapa service/handler yang perlu dirapikan agar tidak bypass boundary |
| Auth JWT | Verified/Basic | Login/refresh/logout/me dan frontend persistence/redirect terverifikasi P0 |
| RBAC | Verified/Basic | Seed admin `SUPER_ADMIN` dan middleware pada route aktif terverifikasi; scope granular lanjutan belum lengkap |
| Workflow FSM | Partial | FSM ada; perlu diselaraskan terus dengan status di UI/SRS |
| Audit Trail | Partial/Conceptual | Model/repository ada, tetapi interceptor mutasi belum terbukti aktif end-to-end |
| Notification | Interface-ready | Provider ada, event/reminder belum usable |
| Document Storage | Conceptual/Partial | MinIO-ready di desain, belum document module usable |
| Dashboard | Verified/Basic | Endpoint dan frontend live plus early warning dasar ada; Level 1-3/Power BI/GIS belum ada |
| Project child lifecycle | Verified | P0-014: child task/milestone/issue/risk/budget/dokumen/team ikut soft delete; dashboard exclude child dari project yang dihapus |
| Vendor & Contract (P1-004) | Verified/Basic | `vendors` master tenant-scoped + `contracts` per project; CRUD end-to-end, unique contract_number per org, validasi nilai/tanggal/enum, audit, tenant/project guard; UI project detail panel \"Kontrak\". Contract value operational input — belum otomatis ubah `project.budget_total` dan belum validated/published (menunggu P1-011/P1-012) |
| Reporting | Verified | UAT-002 2026-08-29: export request menghasilkan file nyata CSV+XLSX; download endpoint tenant-safe; audit events report.export.*; file metadata di DB; 6 dataset tersedia |
| Integration | Missing | Belum ada adapter untuk SIRUP/SIMPONI/OM-SPAN atau import data |
| Data Governance (P3-003) | Verified | 2026-08-28: module `governance` (model/service/handler + unit test) — submission resmi (DRAFT→SUBMITTED→IN_REVIEW→APPROVED→LOCKED, +REJECTED/CANCELLED), per-item validation, lock periods, audit `governance.*`, 12 routes `/api/v1/governance/`, `ResourceDataGovernance`, frontend `/governance`; smoke 17/17 PASS |
| Government Entity Resolution (P3-002) | Verified | 2026-08-28: resolver PENDING_MATCH→MATCHED, migration `000029`, 6 routes, frontend tab "Resolusi Entitas" |
| BIM Integration (P3-001) | Verified | 2026-08-28: module `integration/bim`, 10 routes, frontend `/integrations/bim` |

### 1.1 Modular Monolith (bukan Microservices)

**Keputusan**: CANKORA dibangun sebagai **Modular Monolith**.

**Rationale**:

| Dimensi | Modular Monolith | Microservices |
|---------|-----------------|---------------|
| Complexity awal | Rendah | Tinggi |
| Infrastructure cost | Rendah (1-2 server) | Tinggi (orchestrator, service mesh) |
| Developer onboarding | Mudah | Kompleks |
| Debugging & tracing | Mudah (single process) | Kompleks (distributed tracing) |
| Database transaction | Sederhana (ACID native) | Kompleks (saga, 2PC) |
| Deployment | Simple Docker | Kubernetes atau ekuivalen |
| Migration ke microservices | Bisa dilakukan bertahap | N/A |

**Prinsip Modular Monolith yang harus dijaga**:
- Setiap modul hanya boleh mengakses modul lain melalui **interface yang terdefinisi**, bukan langsung import struct internal
- Tidak ada cross-module database query yang bypass service layer
- Setiap modul memiliki model/entity-nya sendiri, tidak berbagi model

### 1.2 Repository Pattern + Service Layer

```
HTTP Request
    ↓
Handler (transport layer)
    ↓
Service (business logic)
    ↓
Repository (data access)
    ↓
Database (PostgreSQL)
```

Tidak ada business logic di Handler.  
Tidak ada query langsung di Handler atau Service.  
Repository hanya bertanggung jawab terhadap data access.

### 1.3 Interface-First untuk External Dependencies

Semua external dependency (email, storage, notification, cache) dibungkus dalam interface:

```go
type StorageProvider interface {
    Upload(ctx context.Context, key string, content io.Reader) (string, error)
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    GetURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}
```

Ini memungkinkan:
- Swap implementasi tanpa ubah business code
- Mock mudah saat testing
- Keycloak/SSO bisa diganti tanpa ubah auth logic

---

## 2. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT LAYER                             │
│                                                                 │
│  Browser (Chrome/Edge/Firefox)                                  │
│  Next.js 14 App Router (SSR + CSR hybrid)                       │
└─────────────────────────────┬───────────────────────────────────┘
                              │ HTTPS
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      REVERSE PROXY                              │
│                    Nginx / Traefik                               │
│                                                                 │
│  /          → Frontend (Next.js :3000)                          │
│  /api/v1/*  → Backend (Go :8080)                                │
│  /files/*   → MinIO (S3-compatible :9000)                       │
└──────────────┬──────────────────────────────┬───────────────────┘
               │                              │
               ▼                              ▼
┌──────────────────────────┐  ┌───────────────────────────────────┐
│  FRONTEND                │  │  BACKEND                          │
│  Next.js 14              │  │  Go (Gin) — Modular Monolith      │
│  TypeScript              │  │                                   │
│  TailwindCSS             │  │  ┌─────────────────────────────┐  │
│  shadcn/ui               │  │  │  Core Platform              │  │
│  TanStack Query          │  │  │  ├── auth                   │  │
│  Zustand                 │  │  │  ├── user                   │  │
│  React Hook Form + Zod   │  │  │  ├── organization           │  │
│                          │  │  │  ├── rbac                   │  │
└──────────────────────────┘  │  │  ├── workflow               │  │
                              │  │  ├── notification (iface)   │  │
                              │  │  └── audit                  │  │
                              │  │                             │  │
                              │  │  Business Modules           │  │
                              │  │  ├── project                │  │
                              │  │  ├── asset (placeholder)    │  │
                              │  │  ├── inventory (placeholder)│  │
                              │  │  └── cmms (placeholder)     │  │
                              │  │                             │  │
                              │  │  Platform                   │  │
                              │  │  ├── config                 │  │
                              │  │  ├── database               │  │
                              │  │  ├── middleware              │  │
                              │  │  ├── response               │  │
                              │  │  ├── validator              │  │
                              │  │  └── server                 │  │
                              │  └─────────────────────────────┘  │
                              └────────────────┬──────────────────┘
                                               │
                    ┌──────────────────────────┼───────────────┐
                    │                          │               │
                    ▼                          ▼               ▼
          ┌─────────────────┐    ┌─────────────────┐  ┌──────────────┐
          │  PostgreSQL 15  │    │  Redis 7        │  │  MinIO       │
          │  Primary DB     │    │  Cache/Session  │  │  Object      │
          │                 │    │  (Phase 2+)     │  │  Storage     │
          └─────────────────┘    └─────────────────┘  └──────────────┘
```

---

## 3. Backend — Modular Monolith

### 3.1 Folder Structure

```
backend/
├── cmd/
│   ├── api/
│   │   └── main.go              # Entry point: load config, init DB, wire modules, start server
│   ├── migrate/
│   │   └── main.go              # CLI: run/rollback database migrations
│   └── seed/
│       └── main.go              # Seed: default admin, roles, sample organization
│
├── internal/
│   │
│   ├── core/                    # Core Platform — shared services
│   │   ├── auth/
│   │   │   ├── handler.go       # POST /auth/login, /auth/logout, /auth/refresh, /auth/me
│   │   │   │                    # POST /auth/forgot-password, /auth/reset-password
│   │   │   ├── service.go       # AuthService interface + implementation
│   │   │   ├── repository.go    # UserRepository (auth operations only)
│   │   │   ├── model.go         # LoginRequest, TokenPair, Claims
│   │   │   └── middleware.go    # JWT validation middleware
│   │   │
│   │   ├── user/
│   │   │   ├── handler.go       # CRUD user, change password, reset password
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── model.go         # User entity, CreateUserRequest, UpdateUserRequest
│   │   │
│   │   ├── organization/
│   │   │   ├── handler.go       # CRUD org units, tree view
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── model.go         # Organization, OrgUnit entities
│   │   │
│   │   ├── rbac/
│   │   │   ├── handler.go       # CRUD roles, permissions, groups, user-role assignment
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   ├── model.go         # Role, Permission, Group, UserGroup, UserRole, UserScope
│   │   │   └── middleware.go    # RequirePermission(resource, action) middleware
│   │   │
│   │   ├── workflow/
│   │   │   ├── engine.go        # FSM engine: GetAllowedTransitions, Transition
│   │   │   ├── repository.go    # WorkflowDefinition, WorkflowTransition CRUD
│   │   │   └── model.go         # WorkflowDefinition, WorkflowTransition, ApprovalRequest
│   │   │
│   │   ├── notification/
│   │   │   ├── provider.go      # NotificationProvider interface
│   │   │   ├── email.go         # SMTP implementation
│   │   │   └── noop.go          # No-op implementation (development)
│   │   │
│   │   └── audit/
│   │       ├── writer.go        # AuditWriter: Write(ctx, entry AuditEntry)
│   │       ├── repository.go    # Immutable write, read with filters
│   │       ├── handler.go       # GET /audit-logs (admin/auditor only)
│   │       └── model.go         # AuditLog entity (no soft delete)
│   │
│   ├── modules/                 # Business Modules
│   │   ├── project/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── model.go
│   │   ├── governance/          # P3-003 Official Data Validation & Approval (Data Governance)
│   │   │   ├── model.go         # DataSubmission, DataSubmissionItem, DataLockPeriod + constants/DTOs
│   │   │   ├── service.go       # FSM transisi, per-item validation, lock period rules
│   │   │   ├── handler.go       # 12 routes /api/v1/governance/*
│   │   │   └── service_test.go  # Unit test FSM + helpers
│   │   ├── asset/
│   │   │   └── placeholder.go   # package asset — reserved for Phase 6
│   │   ├── inventory/
│   │   │   └── placeholder.go   # package inventory — reserved for Phase 7
│   │   └── cmms/
│   │       └── placeholder.go   # package cmms — reserved for Phase 8
│   │
│   ├── platform/                # Infrastructure & Cross-cutting concerns
│   │   ├── config/
│   │   │   └── config.go        # Env vars loader (godotenv), Config struct
│   │   ├── database/
│   │   │   ├── database.go      # PostgreSQL connection (GORM)
│   │   │   └── migrate.go       # Migration runner (golang-migrate)
│   │   ├── middleware/
│   │   │   ├── cors.go          # CORS configuration
│   │   │   ├── ratelimit.go     # Rate limiting (auth endpoints)
│   │   │   ├── requestid.go     # X-Request-ID injection
│   │   │   └── logger.go        # Structured request logging (zap)
│   │   ├── response/
│   │   │   └── response.go      # Standard API response envelope
│   │   ├── validator/
│   │   │   └── validator.go     # Request validation helpers
│   │   └── server/
│   │       └── server.go        # HTTP server, route registration, graceful shutdown
│   │
│   └── shared/                  # Shared types & utilities
│       ├── types/
│       │   ├── pagination.go    # PaginationRequest, PaginationResponse
│       │   ├── flextime.go      # FlexTime: JSON-flexible time.Time wrapper
│       │   └── uuid.go          # UUID helpers
│       ├── utils/
│       │   ├── crypto.go        # bcrypt hash, compare
│       │   ├── token.go         # JWT helpers
│       │   └── slug.go          # URL-safe slug generator
│       └── constants/
│           └── constants.go     # App-wide constants
│
├── migrations/
│   ├── 000001_create_organizations.up.sql
│   ├── 000001_create_organizations.down.sql
│   ├── 000002_create_users_auth.up.sql
│   ├── 000002_create_users_auth.down.sql
│   ├── 000003_create_rbac.up.sql
│   ├── 000003_create_rbac.down.sql
│   ├── 000004_create_audit_logs.up.sql
│   ├── 000004_create_audit_logs.down.sql
│   ├── 000005_create_projects_core.up.sql
│   └── 000005_create_projects_core.down.sql
│
├── go.mod                       # module github.com/harmanto-49/cankora
├── go.sum
├── .env.example
├── Makefile
└── Dockerfile
```

### 3.2 Standard API Response Envelope

Semua response API menggunakan format yang konsisten:

```json
// Success
{
  "success": true,
  "message": "Project berhasil dibuat",
  "data": { ... },
  "meta": {
    "page": 1,
    "page_size": 20,
    "total": 150,
    "total_pages": 8
  }
}

// Error
{
  "success": false,
  "message": "Validation failed",
  "errors": [
    { "field": "name", "message": "Name is required" },
    { "field": "start_date", "message": "Start date must be before end date" }
  ]
}
```

### 3.3 Route Naming Convention

```
GET    /api/v1/projects              # List
POST   /api/v1/projects              # Create
GET    /api/v1/projects/:id          # Get by ID
PUT    /api/v1/projects/:id          # Update
DELETE /api/v1/projects/:id          # Delete (soft)

GET    /api/v1/projects/:id/tasks    # Nested resource
POST   /api/v1/projects/:id/tasks

GET    /api/v1/projects/:id/export   # Special action
POST   /api/v1/projects/:id/submit   # Workflow action

# Vendor/party master (top-level, tenant-scoped)
GET    /api/v1/vendors?type=VENDOR|CONSULTANT   # List (filter type/search/is_active)
POST   /api/v1/vendors                          # Create
GET    /api/v1/vendors/:vendorID                # Get
PUT    /api/v1/vendors/:vendorID                # Update
DELETE /api/v1/vendors/:vendorID                # Delete (soft; 409 bila in-use)

# Project contract (nested under project, tenant via project ownership)
GET    /api/v1/projects/:id/contracts           # List
POST   /api/v1/projects/:id/contracts           # Create
GET    /api/v1/projects/:id/contracts/:contractID
PUT    /api/v1/projects/:id/contracts/:contractID
DELETE /api/v1/projects/:id/contracts/:contractID   # Delete (soft)

# Official data governance (P3-003) — /api/v1/governance/*
GET    /api/v1/governance/submissions            # List submissions (filter status/dataset/source/year)
POST   /api/v1/governance/submissions            # Create DRAFT submission (items: entity_type/entity_id/action/payload)
GET    /api/v1/governance/submissions/:id        # Get detail (items + timeline + validation errors)
POST   /api/v1/governance/submissions/:id/submit # DRAFT → SUBMITTED
POST   /api/v1/governance/submissions/:id/review # SUBMITTED → IN_REVIEW
POST   /api/v1/governance/submissions/:id/approve# IN_REVIEW → APPROVED (reject bila ada item INVALID)
POST   /api/v1/governance/submissions/:id/reject # IN_REVIEW → REJECTED (wajib rejection_reason; kembali ke DRAFT)
POST   /api/v1/governance/submissions/:id/lock    # APPROVED → LOCKED
POST   /api/v1/governance/submissions/:id/cancel  # DRAFT/SUBMITTED → CANCELLED
GET    /api/v1/governance/lock-periods            # List lock periods
POST   /api/v1/governance/lock-periods            # Create lock period (OPEN)
POST   /api/v1/governance/lock-periods/:id/lock   # OPEN → LOCKED (blokir penulisan periode tsb)
```

**Governance flow (P3-003):** submission resmi berisi dataset (project progress/budget/risk/issue/benefit/location/contract/document/other) dengan item action `CREATE/UPDATE/DELETE/UPSERT/VALIDATE_ONLY`. Transisi FSM ketat (DRAFT→SUBMITTED→IN_REVIEW→APPROVED→LOCKED; REJECTED wajib alasan; CANCELLED). Per-item validation menolak entitas soft-deleted dan mapping pemerintah `PENDING_MATCH` (tidak bisa dijadikan official). Lock period unik per (org, dataset, year, month) memblokir penulisan pada periode terkunci. Semua transisi/approval tercatat di audit `governance.*`; approval menyimpan actor + timestamp. Data yang di-import/synced tidak otomatis approved — harus melewati submission resmi.

**Operational contract input flow (P1-004):** PM/PMO membuka project detail → memilih/menambah vendor (`VENDOR`) atau konsultan (`CONSULTANT`) dari master `/vendors` → mencatat kontrak di `/projects/:id/contracts` (nomor, title, vendor, konsultan opsional, nilai, currency, tanggal, status, scope). UI menampilkan contract summary (nilai total, jumlah active, periode, main vendor). Contract value menjadi sumber kontraktual awal tetapi **tidak otomatis** mengganti `project.budget_total` — desain ini sengaja menunggu snapshot/validasi P1-011; data kontrak diberi label "operational contract input, not yet validated/published" dan belum dipakai untuk klaim contract health di dashboard (P1-014).

### 3.4 Dependency Versions (Pinned)

```
github.com/gin-gonic/gin              v1.9.1
gorm.io/gorm                          v1.25.5
gorm.io/driver/postgres               v1.5.4
github.com/golang-jwt/jwt/v5          v5.2.1
github.com/golang-migrate/migrate/v4  v4.17.0
github.com/joho/godotenv              v1.5.1
golang.org/x/crypto                   v0.25.0
github.com/google/uuid                v1.6.0
go.uber.org/zap                       v1.27.0
github.com/redis/go-redis/v9          v9.5.1
```

---

## 4. Frontend — Next.js 14 App Router

> **Implemented UI sync 2026-08-26:** route aktual mencakup `/login`, `/dashboard`, `/command-center`, `/projects`, `/projects/[id]`, `/validation`, `/reports`, `/benefits`, dan `/users`. Dashboard dan Command Center memakai `DashboardLayout`, `Sidebar`, `TopBar`, TanStack Query, serta service layer. `/command-center` kini memakai aggregate backend tenant-scoped untuk alert/action/validation/escalation/decision; GIS aktual dan Level 1/2 resmi tetap menunggu tiket P2.

> **Implemented UI sync 2026-08-28 (P3-003):** route baru `/governance` — Data Governance. Halaman menyediakan list submission + filter (status/dataset/source/year), modal create submission dengan dynamic items (entity_type/entity_id/action/payload), modal detail dengan status timeline + tabel items + tombol aksi state-aware (submit/review/approve/reject/lock/cancel), dan panel lock period (list/create/lock). Service `frontend/src/services/governance.service.ts` (12 method, unwrap double envelope) + types `frontend/src/types/governance.ts`; item Sidebar \"Data Governance\" (ikon ShieldCheck).

### 4.1 Folder Structure

```
frontend/
├── src/
│   ├── app/                         # Next.js App Router
│   │   ├── (auth)/                  # Route group: unauthenticated pages
│   │   │   └── login/
│   │   │       └── page.tsx
│   │   ├── (dashboard)/             # Route group: authenticated pages
│   │   │   ├── dashboard/page.tsx   # Executive dashboard (/dashboard)
│   │   │   ├── command-center/
│   │   │   │   └── page.tsx         # PMO Command Center frontend shell
│   │   │   ├── projects/
│   │   │   │   ├── page.tsx         # Project list
│   │   │   │   └── [id]/
│   │   │   │       └── page.tsx     # Project detail
│   │   │   └── users/page.tsx       # User management
│   │   ├── layout.tsx               # Root layout (providers)
│   │   └── globals.css
│   │
│   ├── modules/                     # Feature modules
│   │   ├── core/
│   │   │   ├── auth/
│   │   │   │   ├── components/      # LoginForm
│   │   │   │   ├── hooks/           # useLogin, useLogout
│   │   │   │   ├── services/        # authService (API calls)
│   │   │   │   └── types/           # LoginRequest, TokenPair, etc.
│   │   │   ├── users/
│   │   │   ├── organizations/
│   │   │   └── rbac/
│   │   └── project/
│   │       ├── components/
│   │       │   ├── ProjectCard.tsx
│   │       │   ├── ProjectTable.tsx
│   │       │   ├── ProjectHealthBadge.tsx
│   │       │   ├── ProjectForm.tsx
│   │       │   └── dashboard/
│   │       │       ├── ExecutiveDashboard.tsx
│   │       │       ├── ProjectHealthMatrix.tsx
│   │       │       ├── ProgressChart.tsx
│   │       │       └── UpcomingMilestones.tsx
│   │       ├── hooks/
│   │       │   ├── useProjects.ts   # TanStack Query hooks
│   │       │   └── useProject.ts
│   │       ├── services/
│   │       │   └── projectService.ts
│   │       └── types/
│   │           └── index.ts         # Project, Task, Milestone, Issue, Risk types
│   │
│   ├── components/                  # Shared components
│   │   ├── ui/                      # shadcn/ui re-exports
│   │   ├── layout/
│   │   │   ├── Sidebar.tsx
│   │   │   ├── Topbar.tsx
│   │   │   └── Breadcrumb.tsx
│   │   └── common/
│   │       ├── DataTable.tsx        # Generic table with pagination, sort, filter
│   │       ├── StatusBadge.tsx      # Reusable status badge
│   │       ├── ConfirmDialog.tsx    # Delete/destructive action confirmation
│   │       ├── LoadingSpinner.tsx
│   │       ├── EmptyState.tsx
│   │       └── ErrorBoundary.tsx
│   │
│   ├── lib/
│   │   ├── api-client.ts            # Axios instance, interceptors, error handling
│   │   ├── auth.ts                  # Token helpers (get, set, clear, decode)
│   │   └── utils.ts                 # cn(), formatDate(), formatIDR(), etc.
│   │
│   ├── store/
│   │   ├── index.ts                 # Zustand store root
│   │   └── auth.store.ts            # Auth state: user, token, permissions, scopes
│   │
│   ├── hooks/
│   │   ├── usePermission.ts         # canDo(resource, action): boolean
│   │   └── usePagination.ts         # Pagination state management
│   │
│   └── types/
│       ├── api.ts                   # ApiResponse<T>, PaginatedResponse<T>
│       ├── auth.ts                  # User, Permission, Scope types
│       └── common.ts                # SelectOption, etc.
│
├── public/
│   └── logo/
├── package.json
├── tailwind.config.ts
├── tsconfig.json
├── next.config.ts
├── .env.local.example
└── Dockerfile
```

### 4.1.1 Next.js Runtime and Asset Isolation

- `next dev` dan `next build` tidak boleh berjalan bersamaan dengan direktori output `.next` yang sama.
- Untuk production build verification: hentikan semua dev server, jalankan type-check/lint/build, lalu hentikan artefak build sebelum kembali ke satu dev server bersih.
- Gejala gambar `fill` memenuhi viewport atau HTML tanpa styling harus ditangani sebagai CSS/build asset mismatch sebelum mengubah layout.
- Recovery lokal: hentikan seluruh proses Next.js CANKORA, hapus hanya generated `frontend/.next`, jalankan satu `npm run dev`, lalu pastikan `/_next/static/css/app/layout.css` merespons HTTP 200.
- Komponen `next/image` dengan `fill` wajib memiliki parent `position: relative` dan dimensi stabil (`height`, `min-height`, atau `aspect-ratio`).
- Visual gate wajib menguji minimal 1366x768, 1920x1080, 2560x1440, dan mobile tanpa overlap/text clipping.

### 4.2 State Management Strategy

```
Server State    → TanStack Query (useQuery, useMutation)
Auth State      → Zustand (auth.store.ts) — persisted ke localStorage
UI State        → React local state (useState, useReducer)
Form State      → React Hook Form + Zod
```

Tidak menggunakan Redux. Tidak ada global store untuk data yang harusnya di-cache oleh TanStack Query.

### 4.3 API Client Pattern

```typescript
// lib/api-client.ts
// Axios instance dengan:
// - Base URL dari env
// - Authorization header otomatis dari Zustand auth store
// - Request interceptor: inject token
// - Response interceptor: 401 → redirect ke login, format error
```

### 4.4 Permission Check di Frontend

```typescript
// hooks/usePermission.ts
const { canDo } = usePermission()

// Dalam komponen:
{canDo('project', 'create') && <Button>Buat Project</Button>}
{canDo('budget', 'view') && <BudgetTab />}
```

Frontend permission check hanya untuk UX (sembunyikan/tampilkan elemen).  
**Backend selalu re-validate permission** — frontend check bukan pengganti server-side enforcement.

### 4.5 Dependency Versions (Pinned)

```
next                          14.2.5
react                         18.3.1
typescript                    5.5.3
tailwindcss                   3.4.6
@tanstack/react-query         5.51.1
react-hook-form               7.52.1
@hookform/resolvers           3.9.0
zod                           3.23.8
axios                         1.7.2
zustand                       4.5.4
lucide-react                  0.408.0
date-fns                      3.6.0
recharts                      2.12.7
class-variance-authority      0.7.0
clsx                          2.1.1
tailwind-merge                2.4.0
```

---

## 5. Database Architecture

### 5.1 PostgreSQL 15

Semua data disimpan di PostgreSQL 15. Tidak ada polyglot persistence pada Phase 1.

### 5.2 Design Principles

**Soft Delete**: Semua tabel menggunakan kolom `deleted_at TIMESTAMPTZ` (GORM soft delete) — **kecuali `audit_logs`** yang immutable.

**Timestamps**: Semua tabel wajib punya:
```sql
created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
deleted_at  TIMESTAMPTZ          -- NULL = active, populated = deleted
```

**UUID sebagai Primary Key**: Semua tabel menggunakan `UUID` sebagai PK, bukan integer sequence. Ini memudahkan distribusi data di masa depan dan mencegah enumeration attack.

**Organization ID**: Semua tabel data utama memiliki `organization_id UUID` untuk multi-tenancy readiness.

**Created/Updated By**: Tabel yang penting memiliki `created_by UUID` dan `updated_by UUID` yang referensi ke `users.id`.

### 5.3 Naming Convention

```sql
-- Tabel: snake_case, plural
users, org_units, role_permissions, project_teams

-- Kolom: snake_case
first_name, org_unit_id, created_at

-- Index: idx_{table}_{columns}
idx_users_email, idx_projects_status_org_unit_id

-- Foreign Key: fk_{table}_{referenced_table}
fk_users_org_unit, fk_projects_organization
```

### 5.4 Migration Strategy

Menggunakan `golang-migrate` dengan file SQL terpisah (up/down).

```
migrations/
├── 000001_create_organizations.up.sql    # Tabel baru, index
├── 000001_create_organizations.down.sql  # DROP TABLE
```

Aturan:
- **Tidak boleh mengedit migration yang sudah di-apply** di environment manapun
- Selalu buat migration baru untuk perubahan schema
- Setiap migration harus bisa di-rollback (down.sql wajib)
- Test down migration sebelum push

---

## 6. Authentication & Authorization

### 6.1 Authentication Flow

> **Implementasi frontend 2026-08-18:** Guard Next.js middleware membaca cookie `cankora-auth`, sementara client memakai Zustand persist. Karena middleware tidak bisa membaca localStorage/sessionStorage, login client harus menjaga cookie auth tetap sinkron dengan state storage. `Remember me` menentukan apakah state disimpan persistent atau hanya session.

```
1. POST /api/v1/auth/login
   → Validasi email + password (bcrypt compare)
   → Generate Access Token (JWT, 15 menit) + Refresh Token (JWT, 7 hari)
   → Simpan refresh token JTI ke DB (user_sessions table)
   → Return TokenPair

2. Setiap API Request
   → Extract Bearer token dari Authorization header
   → Validate JWT signature + expiry
   → Check JTI tidak di-blacklist (redis/in-memory)
   → Load user + effective permissions
   → Inject ke context

3. POST /api/v1/auth/refresh
   → Validate refresh token
   → Generate Access Token baru
   → Return new TokenPair

4. POST /api/v1/auth/logout
   → Blacklist current Access Token JTI
   → Invalidate Refresh Token di DB
```

### 6.2 JWT Claims

```json
{
  "jti": "unique-token-id",
  "sub": "user-uuid",
  "email": "user@example.com",
  "org_id": "org-uuid",
  "org_unit_id": "unit-uuid",
  "roles": ["role-uuid-1", "role-uuid-2"],
  "iat": 1723564800,
  "exp": 1723565700
}
```

Permission di-load dari DB pada setiap request (bukan di-embed di JWT) untuk memastikan perubahan permission efektif real-time.

### 6.3 Authorization Middleware Stack

```go
// Route dengan auth + permission check:
router.GET("/projects",
    authMiddleware.RequireAuth(),          // JWT valid
    rbacMiddleware.RequirePermission("project", "view"),  // permission check
    projectHandler.List,
)
```

### 6.4 Keycloak Readiness

Auth dibungkus dalam interface `AuthProvider`:

```go
type AuthProvider interface {
    ValidateToken(ctx context.Context, token string) (*Claims, error)
    GenerateTokenPair(ctx context.Context, userID uuid.UUID) (*TokenPair, error)
    RevokeToken(ctx context.Context, jti string) error
}
```

Implementasi default: `JWTAuthProvider` (internal).  
Saat kementerian memiliki SSO: `KeycloakAuthProvider` bisa dibuat sebagai drop-in replacement.

---

## 7. File Storage

### 7.1 MinIO (On-Premise S3-Compatible)

Semua file disimpan di MinIO. Tidak ada file yang disimpan di filesystem server atau database.

### 7.2 Bucket Structure

```
cankora-documents/
├── projects/{project-id}/
│   ├── contracts/
│   ├── reports/
│   ├── photos/
│   └── misc/
├── tasks/{task-id}/
│   └── attachments/
└── users/{user-id}/
    └── avatars/
```

### 7.3 Access Pattern

File tidak pernah diakses langsung oleh browser ke MinIO.  
Semua akses melalui Backend yang memvalidasi permission terlebih dahulu:

```
Browser
  → GET /api/v1/documents/:id/download
  → Backend validate permission
  → Backend generate pre-signed URL (berlaku 15 menit)
  → Redirect ke MinIO pre-signed URL
```

### 7.4 StorageProvider Interface

```go
type StorageProvider interface {
    Upload(ctx context.Context, bucket, key string, r io.Reader, size int64, contentType string) error
    GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
    Delete(ctx context.Context, bucket, key string) error
}
```

Phase 1: implementasi dengan MinIO SDK.  
Future: swap ke AWS S3 atau GCS tanpa ubah business code.

---

## 8. Caching & Background Processing

### 8.1 Redis (Phase 2+)

Redis disiapkan di infrastruktur tapi **belum aktif di Phase 1**.

Use case yang direncanakan:

| Use Case | Implementasi |
|----------|-------------|
| JWT blacklist (logout) | `SET jti:{jti} 1 EX {ttl}` |
| Rate limiting | Sliding window counter |
| Session cache | User permissions cache (TTL 5 menit) |
| Dashboard cache | Executive dashboard (TTL 1 menit) |
| Background jobs | Queue untuk email, notifikasi |

Phase 1: JWT blacklist menggunakan in-memory `sync.Map` dengan cleanup goroutine.

### 8.2 Background Processing

Phase 1: Email dikirim via goroutine (non-blocking, fire-and-forget).

Phase 2+: Queue-based menggunakan Redis atau RabbitMQ untuk:
- Notifikasi email massal
- Export laporan besar
- Scheduled reminders (milestone approaching)

---

## 9. Notification Architecture

### 9.1 NotificationProvider Interface

```go
type NotificationProvider interface {
    Send(ctx context.Context, n Notification) error
    SendBulk(ctx context.Context, ns []Notification) error
}

type Notification struct {
    To       []string
    Subject  string
    Template string
    Data     map[string]any
    Channel  NotificationChannel // EMAIL, WHATSAPP, IN_APP
}
```

### 9.2 Phase 1 Implementation

- `EmailNotificationProvider` — SMTP (`net/smtp` stdlib)
- `NoopNotificationProvider` — untuk development (tidak kirim, hanya log)

### 9.3 Events yang Trigger Notifikasi (Phase 1)

| Event | Penerima |
|-------|---------|
| Project status berubah | PM + Anggota tim |
| Milestone H-7 dan H-3 | PM |
| Issue HIGH/CRITICAL dibuat | PM + Atasan |
| Risk HIGH/CRITICAL dibuat | PM + Atasan |
| Task overdue | Assignee + PM |
| Approval request | Approver |
| Approval decision | Requester |

---

## 10. Audit Trail Architecture

### 10.1 Design Principles

- Audit log **immutable**: tidak ada UPDATE, tidak ada soft delete, tidak ada DELETE
- Ditulis secara **synchronous** dalam transaction yang sama dengan operasi utama
- Jika audit write gagal → operasi utama juga di-rollback

### 10.2 AuditWriter Interface

```go
type AuditWriter interface {
    Write(ctx context.Context, entry AuditEntry) error
}

type AuditEntry struct {
    UserID       uuid.UUID
    UserName     string
    UserIP       string
    Action       AuditAction   // CREATE, UPDATE, DELETE, APPROVE, REJECT, LOGIN, LOGOUT, EXPORT
    ResourceType string        // "project", "task", "user", "role"
    ResourceID   uuid.UUID
    ResourceName string
    FieldChanged string        // Untuk UPDATE: nama field yang berubah
    OldValue     string        // JSON atau string representation
    NewValue     string
    Metadata     map[string]any
}
```

### 10.3 Automatic Capture

Middleware `AuditInterceptor` otomatis mencatat semua mutasi (POST, PUT, PATCH, DELETE) kecuali yang di-whitelist (mis. `POST /auth/login` di-handle secara manual).

---

## 11. Workflow Engine Architecture

### 11.1 FSM-Based Approach

CANKORA menggunakan **Finite State Machine per entitas** yang konfigurasi transisinya disimpan di database.

```sql
-- workflow_definitions: mendefinisikan FSM untuk setiap tipe entitas
-- workflow_transitions: mendefinisikan transisi yang valid beserta aturannya
```

### 11.2 Transition Execution

```go
// Saat PM submit project:
err = workflowEngine.Transition(ctx, TransitionRequest{
    EntityType:  "project",
    EntityID:    projectID,
    FromStatus:  "DRAFT",
    ToStatus:    "SUBMITTED",
    TriggeredBy: userID,
})

// Engine akan:
// 1. Cek apakah transisi valid (ada di workflow_transitions)
// 2. Cek apakah user memiliki role yang diizinkan
// 3. Jika require_approval = true → buat ApprovalRequest
// 4. Jika require_approval = false → update status langsung
// 5. Kirim notifikasi sesuai konfigurasi
// 6. Catat di audit trail
```

### 11.3 Default Workflows (Seed)

> **Perhatian 2026-08-18:** Daftar workflow di bawah adalah desain target lama dan harus diselaraskan dengan kode aktual serta SRS. Kontrak UI/SRS yang sedang dipakai untuk project adalah `DRAFT → PLANNING → ACTIVE → ON_HOLD → COMPLETED / CANCELLED`. Jangan menambah UI transition berdasarkan status lama tanpa audit kode.

```
Project:  DRAFT → SUBMITTED → REVIEWED → APPROVED → ACTIVE → ON_HOLD → COMPLETED → CLOSED
Task:     BACKLOG → ANALYSIS → DEVELOPMENT → QA → UAT → PRODUCTION → DONE
Issue:    OPEN → IN_PROGRESS → RESOLVED → CLOSED
Risk:     IDENTIFIED → ASSESSED → MITIGATED | ACCEPTED | ESCALATED; MITIGATED → CLOSED; ACCEPTED → CLOSED; ESCALATED → MITIGATED
```

---

## 12. Deployment Architecture

### 12.1 Production Stack

```
Internet
    ↓ HTTPS (443)
Nginx (Reverse Proxy + SSL Termination)
    ├── / → Frontend Container (Next.js :3000)
    ├── /api/* → Backend Container (Go :8080)
    └── /storage/* → MinIO Container (:9000) — private, signed URL only
    
Backend Container
    ├── PostgreSQL Container (:5432) — internal network
    ├── Redis Container (:6379) — internal network
    └── MinIO Container (:9000) — internal network
```

### 12.2 Docker Compose Services

```yaml
services:
  backend:     # Go binary, port 8080
  frontend:    # Next.js, port 3000
  postgres:    # PostgreSQL 15-alpine, port 5432
  redis:       # Redis 7-alpine, port 6379
  minio:       # MinIO latest, port 9000 (API) + 9001 (Console)
```

### 12.3 Networking

Semua service dalam satu Docker network internal: `cankora_network`.  
Hanya `nginx` yang exposed ke luar.  
Database, Redis, MinIO tidak exposed ke internet.

### 12.4 Volume Strategy

```
postgres_data   → /var/lib/postgresql/data    (persisten)
redis_data      → /data                       (persisten)
minio_data      → /data                       (persisten)
```

---

## 13. Environment Strategy

### 13.1 Environments

| Environment | URL | Purpose |
|------------|-----|---------|
| Development | localhost | Local development |
| Staging | staging.cankora.internal | Testing & UAT |
| Production | cankora.kemxxx.go.id | Live |

### 13.2 Environment Variables (Backend)

```env
# Application
APP_ENV=production        # development | staging | production
APP_VERSION=0.1.0
PORT=8080

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=cankora
DB_PASSWORD=<secure-password>
DB_NAME=cankora_db
DB_SSLMODE=disable        # enable untuk production dengan SSL cert

# Auth
JWT_SECRET=<min-32-char-random-string>
JWT_ACCESS_EXP_MINUTES=15
JWT_REFRESH_EXP_DAYS=7

# Redis
REDIS_URL=redis://redis:6379/0

# MinIO / S3
STORAGE_ENDPOINT=minio:9000
STORAGE_ACCESS_KEY=<access-key>
STORAGE_SECRET_KEY=<secret-key>
STORAGE_BUCKET_DOCUMENTS=cankora-documents
STORAGE_USE_SSL=false     # true untuk production dengan cert

# SMTP
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@cankora.kemxxx.go.id
SMTP_PASSWORD=<password>
SMTP_FROM=CANKORA <noreply@cankora.kemxxx.go.id>

# App URL (untuk email link, file URL)
APP_URL=https://cankora.kemxxx.go.id
```

### 13.3 Configuration Validation

Saat startup, aplikasi memvalidasi semua required config. Jika ada yang missing → aplikasi **tidak jalan** dan menampilkan error yang jelas:

```
FATAL: Required environment variable DB_PASSWORD is not set.
       Please check your .env file or environment configuration.
```

---

## 14. Future Scalability Path

### 14.1 Dari Single Tenant ke Multi-Tenant

Karena semua tabel utama sudah memiliki `organization_id`, migrasi ke multi-tenant hanya perlu:
1. Tambah `organizations` table untuk menyimpan konfigurasi per tenant
2. Tambah `tenant_id` resolver di middleware (dari subdomain atau header)
3. Update query untuk filter by `organization_id`

Tidak perlu schema-per-tenant (terlalu kompleks). Row-level tenancy lebih practical.

### 14.2 Dari Modular Monolith ke Microservices

Urutan ekstraksi yang disarankan jika diperlukan:
1. Notification Service (paling mudah, sudah punya interface)
2. Document/Storage Service
3. Audit Service
4. Kemudian modul bisnis (Asset, CMMS) sebagai service baru

Module dalam Monolith yang sudah memiliki interface yang clean bisa di-extract menjadi gRPC service dengan perubahan minimal pada caller code.

### 14.3 Module Expansion Path

```
Phase 1  → Project Management Monitoring operational MVP (P0 verified)
Phase 2  → PMO data foundation and core process
Phase 3  → Project intelligence and Control Tower Level 3
Phase 4  → Control Tower Level 2/Level 1, GIS, and Power BI
Phase 5  → Primavera, mobile inspection, BIM/digital twin, government/entity resolution, and data governance
Final/Future → IoT telemetry foundation, time-series optimization, digital twin overlay, and AI/ML
Phase 6  → Fixed Asset, Inventory, CMMS, and cross-module integration
```

Setiap modul baru menggunakan Core Platform yang sudah ada (auth, rbac, workflow, audit, notification) tanpa membangun ulang.

---

## 15. PMO National Project Control Tower Architecture

> Bagian ini adalah target architecture. Implementasi yang sudah terbukti tetap terbatas pada status di [Implementation Gap Analysis](./IMPLEMENTATION-GAP-ANALYSIS.md). Sumber kebutuhan dan batas interpretasi ada di [PMO Control Tower Analysis](./PMO-CONTROL-TOWER-ANALYSIS.md).

Current-state note (2026-08-26): native executive dashboard, PMO Command Center, validation, Health Score, Level 3 Project Control, dan Benefit Indicator sudah memakai backend module/API tenant-scoped. Keduanya belum menggantikan analytics read model/Power BI, Level 1/2 resmi, atau GIS data penuh; capability tersebut tetap mengikuti P2/P3.

Current-state data-flow note (2026-08-26): dashboard native masih membaca current operational tables melalui `GET /api/v1/dashboard` dan `GET /api/v1/projects` untuk sebagian KPI operasional. Command Center kini memakai aggregate backend dari risk, corrective action, validation queue, escalation, dan executive decision. Level 3 Project Control memakai VALID snapshot pada cut-off. Ini belum menjadi analytics read model resmi untuk Level 1/2 atau Power BI.

Planned metadata enrichment (2026-08-29 — DASH-003): Dashboard "10 Proyek Prioritas" membutuhkan kolom `BALAI` setelah `KODE`. Source of truth adalah relasi existing `projects.org_unit_id` → `org_units.name`; tidak perlu field owner baru. API `/api/v1/projects` sebaiknya mengembalikan `org_unit_name` atau object `org_unit` secara tenant-safe. Frontend boleh menampilkan `-`/`Belum ditentukan` ketika `org_unit_id` kosong. Tidak ada permission baru; akses mengikuti Project view dan Org Unit view yang sudah ada.

### 15.1 Keputusan Arsitektur Hybrid

CANKORA tetap menjadi system of record operasional. Power BI menjadi kanal dashboard eksekutif tahap awal dan hanya membaca analytics read model. Dashboard native CANKORA melayani workflow operasional, detail proyek, validation queue, command center, serta drill-down yang memerlukan transaksi.

```text
Manual Entry / Field Evidence / External Systems
                       |
                       v
          Integration & Ingestion Adapters
                       |
                       v
          Validation + Data Quality Pipeline
                       |
                       v
  CANKORA API -> PostgreSQL Operational Store
       |                   |
       |                   v
       |          Rules / Health Engine
       |                   |
       v                   v
Native Operational UI   Analytics Read Model
                              |
                    Power BI Semantic Dataset
```

Power BI tidak boleh mengakses tabel operasional dengan kredensial write. View/materialized view atau schema analytics harus mempunyai metric contract, period cut-off, tenant key, dan row-level security.

### 15.2 Logical Components

| Component | Tanggung Jawab |
|-----------|----------------|
| Operational API | CRUD/workflow project, progress, issue, risk, contract, evidence, validation, dan decision |
| Operational PostgreSQL | System of record tenant-scoped; histori, audit, dan soft delete |
| Integration Adapter Layer | Import file dan connector SIRUP/SIMPONI/OM-SPAN/SAKTI/Primavera melalui interface |
| Validation & Data Quality | Schema validation, completeness, freshness, duplicate detection, approval/rejection, dan lineage |
| Rule/Health Engine | Kalkulasi score versioned, alert generation, explanation, dan recalculation terkontrol |
| Analytics Read Model | Agregasi Level 1-3, trend, ranking, snapshot, dan dataset Power BI |
| GIS Layer | Project point/area, administrative region, DAS, cluster, hotspot, dan spatial filtering |
| Object Storage | Foto inspeksi, dokumen kontrak, laporan, evidence, checksum, dan retention |
| Native Web UI | Operational dashboard, Level 3, command center, validation, dan administration |
| Power BI | Executive Level 1/2 visualization, governed refresh, dan RLS |
| Observability | API metrics, job status, refresh status, failure alert, audit, dan trace correlation |

### 15.3 Data Flow dan Publication Rules

#### Current Operational Dashboard Flow

Alur yang sudah berjalan:

1. User mengelola project, task, milestone, issue, sebagian risk backend, dan progress melalui operational API.
2. Tabel operasional menyimpan current state dengan `organization_id` dan soft delete.
3. `/api/v1/dashboard` menghitung agregat tenant-scoped serta early warning dasar:
   - overdue task;
   - overdue milestone;
   - project near end date dengan progress di bawah threshold;
   - budget usage threshold dari `project_budgets`;
   - `RISK_REGISTER` dari register `risks` (risk score desc, status terbuka, parent project aktif).
4. Frontend dashboard dan Command Center mengambil `/dashboard` dan `/projects`, lalu menghitung presentasi tambahan seperti total nilai portofolio, average progress, health/status chart, watchlist, issue/risk summary, dan data quality heuristic.

Aturan batas current flow:

- Current operational dashboard tidak memiliki `period`, `as_of`, `source`, `submitted_by`, `validated_by`, atau `validation_status`.
- Health/status chart yang sekarang tampil adalah distribusi status operasional, bukan Project Health Score delapan dimensi.
- Risiko pada dashboard kini dibaca dari register `risks` via warning `RISK_REGISTER` (P1-002). Isu belum menjadi ringkasan resmi dari register `issues`; sebagian label isu masih derived dari early warning.
- Budget (P1-003): line item `project_budgets` sudah usable (planned/actual per kategori, currency). `variance`, `usage_pct`, dan `status` (NORMAL/WATCH/RISK/OVERRUN) selalu dihitung backend, bukan frontend. `project_budgets` tidak punya kolom `organization_id`; tenant safety dijamin melalui ownership parent project (`projects.organization_id`). Dashboard `BUDGET_THRESHOLD` warning mengagregat per project dan selalu memakai filter `project_budgets.deleted_at IS NULL` + `projects.deleted_at IS NULL`. Nilai yang tampil adalah operational input yang belum melewati workflow validasi baseline/snapshot (P1-011) dan data quality (P1-012); UI dan label warning menegaskan "operational input, not yet validated". Forecast at completion belum tersedia.
- Data quality/freshness pada Command Center masih heuristic dari field project dan `updated_at`, bukan validation queue.
- GIS masih static image sampai spatial model dan map API tersedia.

#### Target Governed Publication Flow

1. Data masuk sebagai draft/submission dengan source dan period.
2. Pipeline memvalidasi schema, tenant, referensi master, completeness, dan freshness.
3. Validator menerima atau menolak submission dengan alasan.
4. Data valid menjadi sumber kalkulasi KPI/health pada period terkait.
5. Health engine menyimpan score dan component snapshot dengan formula version.
6. Reporting snapshot dipublikasikan secara immutable.
7. Analytics read model diperbarui; Power BI melakukan controlled refresh.

Data `REJECTED`, `DRAFT`, atau `STALE` tidak boleh diam-diam dipakai sebagai data resmi. Dashboard harus menampilkan `as_of`, freshness, dan status publikasi.

### 15.4 Tenant, Security, dan Audit

- Semua tabel bisnis dan row analytics membawa `organization_id`.
- Join antarentity wajib memverifikasi tenant yang sama.
- Power BI RLS memetakan identity ke organization, org unit, program, atau project scope.
- Dataset export tidak boleh membocorkan project di luar effective scope.
- Perubahan formula, validasi, override, keputusan, export, dan refresh failure tercatat.
- Published report/health snapshot immutable; koreksi menghasilkan version/snapshot baru.
- Evidence file menggunakan signed/authorized access; metadata dan checksum disimpan di PostgreSQL.

### 15.5 Planned API and Dataset Families

Nama route final ditetapkan saat tiket implementasi, tetapi kontrak publik harus dikelompokkan secara konsisten:

- `/programs`, `/sectors`, `/regions`, `/watersheds`, dan project location metadata.
- Nested project resources untuk contracts, progress snapshots, inspections/evidence, issues, risks, corrective actions, health, dan benefits.
- `/validation-queue`, `/command-center`, serta executive decisions.
- Dashboard/read endpoints untuk project, program/sector, dan national level.
- Integration/import endpoints dan read-only analytics dataset contract.

Semua endpoint aplikasi tetap memakai `response.*`; frontend memanggil melalui `src/services/`; request/response TypeScript tidak boleh memakai `any`.

### 15.6 GIS and Spatial Strategy

Tahap awal menggunakan koordinat point dan referensi wilayah/DAS. PostgreSQL tetap menjadi sumber metadata; PostGIS ditambahkan ketika kebutuhan spatial query, polygon, atau intersection telah disetujui. API mengirim GeoJSON atau struktur typed yang setara dan selalu menerapkan tenant/scope filter sebelum data dipetakan.

### 15.7 Background Jobs

Background processing diperlukan untuk import besar, data-quality scan, health recalculation, report snapshot, analytics refresh, notification, dan external synchronization. Job harus idempotent, memiliki correlation ID, retry terbatas, dead-letter/failure state, serta status yang dapat dilihat integration admin.

### 15.8 Advanced Integration Guardrails

- Primavera P6: adapter schedule membaca baseline/actual tanpa mengikat domain project pada format vendor.
- Mobile inspection: offline/PWA baru setelah conflict resolution dan upload retry dirancang.
- BIM/digital twin: simpan referensi/version metadata; file besar di object storage/sistem BIM.
- IoT: parked as the final/future advanced integration; tahap awal nanti adalah data gateway + device/source registry + telemetry validation + alert foundation. Telemetry wajib dipisahkan dari tabel transaksi; retention dan sampling harus ditetapkan sebelum implementasi.
- AI/ML: tidak menggantikan rules yang dapat dijelaskan; prediction menyimpan model version, feature timestamp, confidence, dan explanation.

> **Update 2026-08-28 (IoT planning):** IoT telemetry (`CANKORA-P3-004`) dipindah ke urutan paling akhir/future advanced integration. Bentuk awal yang disetujui untuk nanti adalah IoT data gateway + device/source registry + telemetry validation + alert foundation; MQTT, time-series optimization, alert rules engine, dan digital twin overlay baru menyusul setelah foundation stabil. Modul `governance` tetap menjadi prasyarat: telemetry tidak boleh langsung menjadi data official, dapat dibuat sebagai draft submission per device/period, dan dilarang menulis ke dataset yang sudah LOCKED. Jangan mulai IoT sebelum governance hardening, dashboard/GIS regression, reporting/read-model reconciliation, dan Phase 3 non-IoT work stabil.
