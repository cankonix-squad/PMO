# Production module boundaries

This tree is the target boundary for the production refactor. Public HTTP
contracts and database migrations remain compatible during migration.

- `bootstrap`: dependency wiring and application startup.
- `platform`: framework and infrastructure adapters only.
- `identity`: authentication, users, roles, permissions, and scopes.
- `governance`: audit, workflow, and data-quality controls.
- `project`: project-management bounded context split by subdomain.
- `analytics` and `integrations`: read models and external connectors.
- `jobs`: asynchronous work that must not run in an HTTP request.

New code must depend inward on domain interfaces. Domain packages must not
import Gin, GORM, or concrete infrastructure clients.
