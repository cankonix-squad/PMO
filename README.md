# PMO — Enterprise Operations Platform by Cankonix

> Modular enterprise platform for government operations management. Phase 1: Project Management Monitoring.

> **Implementation note (2026-08-18):** Current code is an MVP foundation and does not yet satisfy the full PMO Grand Design requirements. See [docs/planning/IMPLEMENTATION-GAP-ANALYSIS.md](docs/planning/IMPLEMENTATION-GAP-ANALYSIS.md) before treating the SRS/architecture documents as completed scope.

---

## Modules

| Module | Status | Description |
|--------|--------|-------------|
| Project Management | In progress — Phase 1 MVP foundation | Projects, Tasks, Milestones, Issues, Risks, Budget |
| Fixed Asset Management | Planned — Phase 2 | Asset register, depreciation, maintenance |
| Inventory Management | Planned — Phase 3 | Stock, procurement, reorder |
| CMMS | Planned — Phase 4 | Preventive maintenance, work orders |

---

## Documentation

See the [documentation index](docs/README.md) for the complete folder guide.

| Document | Purpose |
|----------|---------|
| [SRS](docs/architecture/SRS.md) | Target product requirements for PMO |
| [Architecture](docs/architecture/ARCHITECTURE.md) | Target technical architecture and implementation notes |
| [Conceptual ERD](docs/architecture/ERD-CONCEPTUAL.md) | Conceptual data model |
| [Permission Matrix](docs/architecture/PERMISSION-MATRIX.md) | Target role/resource/action model |
| [Implementation Gap Analysis](docs/planning/IMPLEMENTATION-GAP-ANALYSIS.md) | Current alignment against the PMO Grand Design and remaining work |

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.23, Gin, GORM, PostgreSQL 15 |
| Frontend | Next.js 14 (App Router), TypeScript, TailwindCSS, shadcn/ui |
| Auth | JWT (RS256-ready), Keycloak-ready interface |
| Cache | Redis (Phase 2) |
| Storage | MinIO (Phase 2) |
| Queue | — (Phase 2) |

---

## Quick Start (Development)

### Prerequisites
- Go 1.23+
- Node.js 20+
- PostgreSQL 15
- Docker & Docker Compose (for full stack)

### Local Setup

```bash
# 1. Clone
git clone https://github.com/harmanto-49/PMO.git
cd PMO

# 2. Backend
cd backend
cp .env.example .env          # edit values
make tidy
make migrate-up
make seed
make dev                      # runs on :8083

# 3. Frontend (new terminal)
cd frontend
cp .env.local.example .env.local
npm install
npm run dev                   # runs on :3003
```

### Docker (Full Stack)

```bash
# From project root
cp backend/.env.example backend/.env   # edit at minimum JWT_SECRET
docker-compose up --build -d

# Run migrations + seed inside container
docker-compose exec backend ./cankora migrate up
docker-compose exec backend ./cankora seed
```

---

## Default Credentials (after seed)

| Role | Email | Password |
|------|-------|----------|
| Super Admin | admin@cankora.local | Admin@Cankora2024! |

> Change this immediately in production.

> Demo accounts (`pmo@`, `pm@`, `officer@`, `executive@`, `auditor@`) are
> optional and only created by `make seed-demo`. After
> `make reset-local-admin-only` the Super Admin above is the only account.

---

## Reset Local Data (Super Admin only)

Wipes demo/transactional data from the **local** database and leaves only the
default organization, RBAC/permissions/roles, master data, and the Super Admin
user. The schema and the `schema_migrations` bookkeeping table are preserved, and
`seed-demo` is **not** run.

```bash
# From the project root — stops the app, migrates, truncates, seeds, verifies
bash scripts/reset-local-admin-only.sh

# Non-interactive
bash scripts/reset-local-admin-only.sh --yes

# Skip stopping the app / skip migrations
bash scripts/reset-local-admin-only.sh --keep-running
bash scripts/reset-local-admin-only.sh --no-migrate
```

Backend Makefile targets:

```bash
cd backend
make reset-local-admin-only   # truncate + seed Super Admin only + verify
make reset-local-verify       # read-only verification (11 checks)
make reset-local              # truncate only (no re-seed)
```

**Removed:** demo users, projects `DEMO-SDA-*` and all child data (tasks,
milestones, issues, risks, budgets, contracts, documents, periodic reports,
inspections, benefits, snapshots, audit logs, notifications, integration runs…).
Also leaves `projects` empty.

**Kept:** `schema_migrations`, and after re-seed the organization, 7 roles, 246
permissions, 862 role-permission mappings, programs/sectors/regions/river
basins/org units/project categories, and `admin@cankora.local`.

Safety guards (both the script and `cmd/reset-local`):

- refuses `APP_ENV=production` and managed/cloud DB hosts
- refuses non-local `DB_HOST`
- requires `DB_NAME` to match `RESET_LOCAL_EXPECTED_DB` (default `cankora_db`)
- `cmd/reset-local` requires `--yes`; override with `ALLOW_PRODUCTION_RESET=true`
  only if you are absolutely sure

> Prefer this over `make migrate-drop`: `migrate-drop` destroys the whole schema
> and requires a full migration replay. The reset above only deletes application
> rows, so the schema stays as-is.

---


## Architecture

```
PMO/
├── backend/
│   ├── cmd/                # API, migration, and seed entry points
│   ├── internal/
│   │   ├── bootstrap/      # Dependency wiring and application lifecycle
│   │   ├── identity/       # Auth, users, RBAC, organizations
│   │   ├── project/        # Project services, repositories, related modules
│   │   ├── governance/     # Audit, workflow, validation
│   │   ├── analytics/      # Dashboards, reporting, read models
│   │   ├── integrations/   # Imports, government, BIM, Primavera
│   │   ├── notifications/  # Notification API and delivery
│   │   ├── platform/       # HTTP router/middleware, config, database
│   │   ├── modules/        # Future asset/inventory/CMMS stubs
│   │   └── shared/         # Shared constants, types, utilities
│   ├── tests/              # Contract, security, permission, tenant tests
│   └── migrations/         # Existing SQL migrations
├── frontend/src/
│   ├── app/                # Next.js routes
│   ├── features/projects/  # Project layouts, components, schemas, hooks
│   ├── components/         # Shared application UI
│   ├── services/           # Shared API layer
│   ├── store/              # Authentication/UI state
│   ├── lib/                # Shared infrastructure/utilities
│   └── types/              # API contracts
├── docs/
│   ├── architecture/
│   ├── planning/
│   ├── uat/
│   └── references/
├── scripts/
│   └── smoke/              # Existing smoke/UAT scripts
├── archives/
├── compose.yml
└── docker-compose.yml
```

---

## Makefile Commands

```bash
make build          # Build production binary
make dev            # Run with go run
make test           # Run tests with race detector
make lint           # Run golangci-lint
make migrate-up     # Apply pending migrations
make migrate-down   # Revert migrations
make seed           # Seed default data
make docker-up      # Start all Docker services
```

---

## API Health Check

```bash
curl http://localhost:8083/health
# {"status":"ok","version":"0.1.0"}
```

---

## Contributing

See `CLAUDE.md` for project context used by AI assistants.

## License

Proprietary — Cankonix. All rights reserved.
