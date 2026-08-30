# =============================================================================
# CANKORA Backend — Makefile
# =============================================================================

APP_NAME   := cankora
BUILD_DIR  := ./bin
MAIN       := ./cmd/api/main.go
MIGRATE    := ./cmd/migrate/main.go
SEED       := ./cmd/seed/main.go
SEED_DEMO  := ./cmd/seed-demo/main.go

# Detect OS for binary extension
ifeq ($(OS),Windows_NT)
    BINARY_EXT := .exe
else
    BINARY_EXT :=
endif

BINARY := $(BUILD_DIR)/$(APP_NAME)$(BINARY_EXT)

.PHONY: all build run dev clean test lint tidy \
        migrate-up migrate-down migrate-drop seed seed-demo \
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
	@echo "CANKORA Backend — Available targets:"
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
	@echo "  make docker-build   Build Docker image"
	@echo "  make docker-up      Start all services via docker-compose"
	@echo "  make docker-down    Stop all services"
	@echo ""
