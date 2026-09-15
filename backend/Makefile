# =============================================================================
# PMO Backend — Makefile
# =============================================================================

APP_NAME   := cankora
BUILD_DIR  := ./bin
MAIN       := ./cmd/api/main.go
MIGRATE    := ./cmd/migrate/main.go
SEED       := ./cmd/seed/main.go
SEED_DEMO  := ./cmd/seed-demo/main.go
RESET      := ./cmd/reset-local/main.go

# Detect OS for binary extension
ifeq ($(OS),Windows_NT)
    BINARY_EXT := .exe
else
    BINARY_EXT :=
endif

BINARY := $(BUILD_DIR)/$(APP_NAME)$(BINARY_EXT)

.PHONY: all build run dev clean test lint tidy \
        migrate-up migrate-down migrate-drop seed seed-demo \
        reset-local reset-local-admin-only reset-local-verify \
        docker-build docker-up docker-down

# -----------------------------------------------------------------------------
# Build
# -----------------------------------------------------------------------------

all: tidy build

build:
	@echo "→ Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BINARY) $(MAIN)
	@echo "✓ Built: $(BINARY)"

run: build
	@echo "→ Running $(APP_NAME)..."
	$(BINARY)

dev:
	@echo "→ Starting dev server (hot reload not configured — use air for that)..."
	go run $(MAIN)

clean:
	@echo "→ Cleaning..."
	rm -rf $(BUILD_DIR)

# -----------------------------------------------------------------------------
# Test & Lint
# -----------------------------------------------------------------------------

test:
	go test ./... -v -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	@command -v golangci-lint >/dev/null 2>&1 || \
		(echo "golangci-lint not found — install: brew install golangci-lint" && exit 1)
	golangci-lint run ./...

tidy:
	go mod tidy
	go mod verify

# -----------------------------------------------------------------------------
# Database migrations
# -----------------------------------------------------------------------------

migrate-up:
	@echo "→ Running migrations UP..."
	go run $(MIGRATE) up

migrate-down:
	@echo "→ Running migrations DOWN..."
	go run $(MIGRATE) down

migrate-drop:
	@echo "→ Dropping all tables..."
	go run $(MIGRATE) drop

# -----------------------------------------------------------------------------
# Seed
# -----------------------------------------------------------------------------

seed:
	@echo "→ Seeding database..."
	go run $(SEED)

seed-demo:
	@echo "→ Seeding idempotent SDA demo data..."
	go run $(SEED_DEMO)

# -----------------------------------------------------------------------------
# Local data reset
# -----------------------------------------------------------------------------
#
# Wipes demo/transactional data from the LOCAL database and re-seeds the
# baseline (organization, RBAC, master data, default Super Admin) WITHOUT
# demo users. The schema and the golang-migrate bookkeeping table are kept.
#
# Safety: cmd/reset-local refuses to run against production/cloud hosts and
# requires DB_NAME to match RESET_LOCAL_EXPECTED_DB (default cankora_db).
# Prefer scripts/reset-local-admin-only.sh at the repo root, which also stops
# the running app first.

reset-local:
	@echo "→ Truncating local application data (schema preserved)..."
	go run $(RESET) --yes

reset-local-admin-only: reset-local
	@echo "→ Re-seeding baseline with Super Admin only (no demo users)..."
	SEED_SKIP_DEMO_USERS=true go run $(SEED)
	@echo "→ Verifying reset result..."
	go run $(RESET) --verify

reset-local-verify:
	@echo "→ Verifying local reset result (read-only)..."
	go run $(RESET) --verify

# -----------------------------------------------------------------------------
# Docker
# -----------------------------------------------------------------------------

docker-build:
	docker build -t $(APP_NAME):latest .

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

# -----------------------------------------------------------------------------
# Help
# -----------------------------------------------------------------------------

help:
	@echo ""
	@echo "PMO Backend — Available targets:"
	@echo "  make build          Build production binary"
	@echo "  make run            Build and run"
	@echo "  make dev            Run with go run (development)"
	@echo "  make test           Run tests with race detector"
	@echo "  make lint           Run golangci-lint"
	@echo "  make tidy           go mod tidy + verify"
	@echo "  make migrate-up     Apply all pending migrations"
	@echo "  make migrate-down   Revert all migrations"
	@echo "  make migrate-drop   Drop all tables (DESTRUCTIVE)"
	@echo "  make seed           Seed database with default data"
	@echo "  make seed-demo      Seed idempotent SDA dashboard demo data"
	@echo "  make reset-local    Truncate local app data (schema preserved)"
	@echo "  make reset-local-admin-only"
	@echo "                      Reset local DB + seed Super Admin only (no demo)"
	@echo "  make reset-local-verify"
	@echo "                      Verify a local reset (read-only)"
	@echo "  make docker-build   Build Docker image"
	@echo "  make docker-up      Start all services via docker-compose"
	@echo "  make docker-down    Stop all services"
	@echo ""
