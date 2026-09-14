# Module boundaries

The package migration reviewed on 2026-09-14 is implemented. This remains a
modular monolith: HTTP routes, JSON contracts, database tables, and migrations
are preserved. Compatibility aliases and their retired package paths are removed.

## Implemented structure

- `bootstrap`: dependency wiring, application startup, graceful shutdown.
- `platform/http/router`: engine, route registration, and HTTP dependencies.
- `platform/http/middleware`: the single implementation of HTTP request controls.
- `platform`: configuration, database connection, and response adapters.
- `identity`: authentication, users, roles/permissions/scopes, organizations.
- `governance`: audit, audit-log HTTP access, workflow, submissions, data quality.
- `project`: project aggregate, with task/milestone/issue/risk/budget/vendor/
  contract/document/corrective-action services and repositories in separate files;
  related category, field, health, monitoring, portfolio, and spatial packages.
- `analytics`: dashboards, program/executive/GIS/project-control read models,
  priority, benefits, reporting, and command center.
- `integrations`: imports, government connectors, BIM, and Primavera.
- `notifications`: notification HTTP API and delivery implementations.
- `modules`: existing future asset/inventory/CMMS stubs only.

Project and user services depend on repository interfaces. Project periodic
progress synchronization also goes through the repository contract; services no
longer assert the concrete PostgreSQL repository. Persistence implementations
live in `*_repository.go`/`repository.go`; HTTP adapters live in handler files.
Existing GORM persistence models and other legacy module adapters remain part of
this modular monolith; this migration does not claim a fully framework-free domain.

## Rules for subsequent changes

- Keep tenant filters and soft-delete behavior in repository operations.
- Do not import retired `core`, `platform/server`, or `platform/middleware` paths.
- Register each route once, with authentication and action-specific permissions.
- Keep HTTP and database dependencies out of extracted project/user service files.
- Change public contracts only through an explicitly scoped behavior change.

The empty `jobs` placeholders were removed. A background worker would change
execution, retries, and delivery semantics; exports and integrations retain their
existing synchronous behavior. Implement durable jobs as a separate feature when
required, rather than leaving empty folders that imply an existing worker.

## Verification

Run `go test -race ./...` and `go build ./...` from `backend/`.
`tests/integration/testdata/routes.txt` captures the routes before migration.
Tests cover route preservation, rejection of missing/invalid/revoked tokens,
permission resource/action matching, tenant/parent service guards, real repository
SQL tenant predicates through a recording driver, and retired dependency guards.
These tests require no live database. They do not replace staging database/UAT tests.
