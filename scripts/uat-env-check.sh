#!/usr/bin/env bash
# =============================================================================
# CANKORA UAT Environment Check
# Verifikasi prasyarat lokal sebelum menjalankan UAT
# Usage: bash scripts/uat-env-check.sh
# =============================================================================
set -uo pipefail

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$BASE_DIR/backend"
FRONTEND_DIR="$BASE_DIR/frontend"

PASS=0; FAIL=0; WARN=0

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
pass()  { echo -e "${GREEN}✓ PASS${NC}  $1"; PASS=$((PASS+1)); }
fail()  { echo -e "${RED}✗ FAIL${NC}  $1"; FAIL=$((FAIL+1)); }
warn()  { echo -e "${YELLOW}~ WARN${NC}  $1"; WARN=$((WARN+1)); }
header(){ echo; echo "── $1 ──"; }

header "1. Required binaries"
command -v go    >/dev/null 2>&1 && pass "go found: $(go version | awk '{print $3}')" || fail "go not found"
command -v node  >/dev/null 2>&1 && pass "node found: $(node --version)" || fail "node not found"
command -v npm   >/dev/null 2>&1 && pass "npm found: $(npm --version)" || fail "npm not found"
command -v psql  >/dev/null 2>&1 && pass "psql found" || warn "psql not found (optional for CLI checks)"
command -v curl  >/dev/null 2>&1 && pass "curl found" || fail "curl not found"

header "2. PostgreSQL connectivity"
PG_HOST="${DB_HOST:-localhost}"
PG_PORT="${DB_PORT:-5432}"
PG_DB="${DB_NAME:-cankora_db}"
PG_USER="${DB_USER:-cankora}"
if nc -z "$PG_HOST" "$PG_PORT" 2>/dev/null; then
  pass "PostgreSQL port $PG_HOST:$PG_PORT reachable"
else
  fail "PostgreSQL port $PG_HOST:$PG_PORT not reachable — start PostgreSQL first"
fi

header "3. Backend .env"
if [ -f "$BACKEND_DIR/.env" ]; then
  pass ".env exists"
  # Check no placeholder still in place
  if grep -q "CHANGE_ME" "$BACKEND_DIR/.env" 2>/dev/null; then
    warn ".env still has CHANGE_ME placeholder values — update before UAT"
  else
    pass ".env has no CHANGE_ME placeholders"
  fi
  # Check JWT_SECRET length
  JWT_SECRET=$(grep "^JWT_SECRET=" "$BACKEND_DIR/.env" 2>/dev/null | cut -d= -f2-)
  if [ ${#JWT_SECRET} -ge 32 ]; then
    pass "JWT_SECRET length >= 32 chars"
  else
    fail "JWT_SECRET too short (< 32 chars) — set a strong secret"
  fi
else
  fail ".env not found — run: cp backend/.env.example backend/.env && edit backend/.env"
fi

header "4. Backend node_modules / go modules"
if [ -d "$BACKEND_DIR/vendor" ] || [ -f "$BACKEND_DIR/go.sum" ]; then
  pass "go.sum exists"
else
  warn "go.sum not found — run: cd backend && go mod tidy"
fi
if [ -d "$FRONTEND_DIR/node_modules" ]; then
  pass "node_modules exists"
else
  fail "node_modules not found — run: cd frontend && npm install"
fi

header "5. Migrations"
LATEST_MIGRATION=$(ls "$BACKEND_DIR/migrations/"*.up.sql 2>/dev/null | wc -l | tr -d ' ')
if [ "$LATEST_MIGRATION" -gt 0 ]; then
  pass "$LATEST_MIGRATION migration files found"
else
  fail "No migration files found in backend/migrations/"
fi

header "6. Ports availability"
if nc -z localhost 8080 2>/dev/null; then
  warn "Port 8080 already in use — backend may already be running (OK if intentional)"
else
  pass "Port 8080 available"
fi
if nc -z localhost 3000 2>/dev/null; then
  warn "Port 3000 already in use — frontend may already be running (OK if intentional)"
else
  pass "Port 3000 available"
fi

header "7. Security sanity"
# Ensure .env is not committed
if [ -f "$BASE_DIR/.gitignore" ] && grep -q "\.env" "$BASE_DIR/.gitignore" 2>/dev/null; then
  pass ".env pattern in .gitignore"
else
  warn ".env not explicitly in root .gitignore — check before committing"
fi
if [ -f "$BACKEND_DIR/.gitignore" ] && grep -q "\.env" "$BACKEND_DIR/.gitignore" 2>/dev/null; then
  pass "backend/.env in backend/.gitignore"
else
  warn "backend/.env not explicitly in backend/.gitignore"
fi
# Check no real production DB strings
if grep -qE "rds\.amazonaws|supabase|neon\.tech|railway\.app" "$BACKEND_DIR/.env" 2>/dev/null; then
  fail "Production DB URL detected in .env — do NOT use production credentials for UAT"
else
  pass "No production cloud DB URL in .env"
fi

echo
echo "═══════════════════════════════════════════"
echo " UAT Environment Check"
echo "═══════════════════════════════════════════"
echo " PASS: $PASS  WARN: $WARN  FAIL: $FAIL"
echo "═══════════════════════════════════════════"
if [ "$FAIL" -gt 0 ]; then
  echo " ✗ ENVIRONMENT NOT READY — fix FAIL items above"
  exit 1
elif [ "$WARN" -gt 0 ]; then
  echo " ~ ENVIRONMENT READY WITH WARNINGS — review WARN items"
  exit 0
else
  echo " ✓ ENVIRONMENT READY"
  exit 0
fi
