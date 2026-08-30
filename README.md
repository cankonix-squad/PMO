# PMO — Enterprise Operations Platform by Cankonix

> Modular enterprise platform for government operations management. Phase 1: Project Management Monitoring.

> **Implementation note (2026-08-18):** Current code is an MVP foundation and does not yet satisfy the full PMO Grand Design requirements. See [docs/IMPLEMENTATION-GAP-ANALYSIS.md](docs/IMPLEMENTATION-GAP-ANALYSIS.md) before treating the SRS/architecture documents as completed scope.

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

| Document | Purpose |
|----------|---------|
| [SRS](docs/SRS.md) | Target product requirements for PMO |
| [Architecture](docs/ARCHITECTURE.md) | Target technical architecture and implementation notes |
| [Conceptual ERD](docs/ERD-CONCEPTUAL.md) | Conceptual data model |
| [Permission Matrix](docs/PERMISSION-MATRIX.md) | Target role/resource/action model |
| [Implementation Gap Analysis](docs/IMPLEMENTATION-GAP-ANALYSIS.md) | Current alignment against the PMO Grand Design and remaining work |

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
make dev                      # runs on :8080

# 3. Frontend (new terminal)
cd frontend
cp .env.local.example .env.local
npm install
npm run dev                   # runs on :3000
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

---

## Architecture

```
PMO/
├── backend/               # Go Modular Monolith
│   ├── cmd/
│   │   ├── api/           # HTTP server entry point
│   │   ├── migrate/       # Migration runner
│   │   └── seed/          # Database seeder
│   ├── internal/
│   │   ├── platform/      # config, database, middleware, server, response
│   │   ├── shared/        # constants, types, utils
│   │   ├── core/          # auth, audit, organization, rbac, notification, user, workflow
│   │   └── modules/       # project, asset (stub), inventory (stub), cmms (stub)
│   └── migrations/        # SQL migration files (golang-migrate)
├── frontend/              # Next.js 14 App Router
│   └── src/
│       ├── app/           # Routes (App Router)
│       ├── components/    # Reusable UI components
│       ├── services/      # API service layer
│       ├── store/         # Zustand stores
│       ├── lib/           # Axios instance, query client
│       └── types/         # TypeScript types
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
curl http://localhost:8080/health
# {"status":"ok","version":"0.1.0"}
```

---

## Contributing

See `CLAUDE.md` for project context used by AI assistants.

## License

Proprietary — Cankonix. All rights reserved.
