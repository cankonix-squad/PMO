# Regression verification

Run from `backend/`:

```sh
go test -race ./...
go build ./...
```

- `architecture`: retired imports and service/infrastructure dependency guards.
- `integration`: the pre-refactor HTTP route contract and missing/invalid token checks.
- `security`: revoked sessions and access/refresh token separation.
- `authorization`: admin/project route denial and exact resource/action matching.
- `tenant_isolation`: user/project parent guards and tenant/soft-delete SQL predicates.
- `testkit`: isolated HTTP dependencies and narrow repository fakes.

No live database or external service is required. Repository SQL tests run the
actual GORM repository against a recording database driver. They check SQL shape
and bound organization IDs; PostgreSQL behavior and full business flows still
need the existing smoke/UAT scripts against a prepared test environment.

`integration/testdata/routes.txt` is the contract recorded before package migration.
Update it only after reviewing an intentional HTTP API change:

```sh
UPDATE_ROUTE_CONTRACT=1 go test ./tests/integration -run TestRouteContract
```
