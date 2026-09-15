#!/usr/bin/env bash
# ============================================================
# PMO — Smoke Test: Local Reset (Super Admin only)
# ============================================================
# Memverifikasi bahwa database lokal sudah di-reset ke kondisi
# "Super Admin only":
#   - hanya 1 user: admin@cankora.local
#   - demo users (pmo@/pm@/officer@/executive@/auditor@) tidak ada
#   - project DEMO-SDA-* dan seluruh child data project kosong
#   - organisasi default + RBAC + master data tetap ada
#   - schema_migrations tetap utuh
#
# Usage:
#   bash scripts/smoke/smoke-test-reset-local-admin-only.sh
#   BASE_URL=http://localhost:8083 bash scripts/smoke/smoke-test-reset-local-admin-only.sh
#
# Requirements:
#   - psql (Postgres.app / Homebrew) untuk DB checks
#   - backend optional; jika tidak berjalan, API checks di-SKIP
# ============================================================

set -uo pipefail

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="$BASE_DIR/backend/.env"
BASE_URL="${BASE_URL:-http://localhost:8083}"
API="${BASE_URL}/api/v1"
ADMIN_EMAIL="admin@cankora.local"
ADMIN_PASSWORD="Admin@Cankora2024!"
DEMO_PASSWORD="Demo@Cankora2024!"

PASS=0; FAIL=0; SKIP=0

green()  { echo -e "\033[32m[PASS]\033[0m $*"; }
red()    { echo -e "\033[31m[FAIL]\033[0m $*"; }
yellow() { echo -e "\033[33m[SKIP]\033[0m $*"; }
header() { echo -e "\n\033[1;34m$*\033[0m"; }

pass() { green  "$1"; PASS=$((PASS+1)); }
fail() { red    "$1"; FAIL=$((FAIL+1)); }
skip() { yellow "$1"; SKIP=$((SKIP+1)); }

json_field() { echo "$1" | jq -r "$2" 2>/dev/null; }

# ── 0. Locate psql ────────────────────────────────────────────────────────────
PSQL=""
for candidate in \
  "$(command -v psql 2>/dev/null || true)" \
  /Applications/Postgres.app/Contents/Versions/15/bin/psql \
  /Applications/Postgres.app/Contents/Versions/latest/bin/psql \
  /opt/homebrew/opt/postgresql@15/bin/psql \
  /usr/local/opt/postgresql@15/bin/psql
do
  if [ -n "$candidate" ] && [ -x "$candidate" ]; then PSQL="$candidate"; break; fi
done

# ── 1. Read DB config from backend/.env ───────────────────────────────────────
env_value() {
  local key="$1"
  grep -E "^[[:space:]]*${key}=" "$ENV_FILE" 2>/dev/null | tail -1 | cut -d= -f2- \
    | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' -e 's/^"//' -e 's/"$//' -e "s/^'//" -e "s/'$//"
}

if [ -f "$ENV_FILE" ]; then
  DB_HOST="$(env_value DB_HOST)"; DB_HOST="${DB_HOST:-localhost}"
  DB_PORT="$(env_value DB_PORT)"; DB_PORT="${DB_PORT:-5432}"
  DB_NAME="$(env_value DB_NAME)"; DB_NAME="${DB_NAME:-cankora_db}"
  DB_USER="$(env_value DB_USER)"; DB_USER="${DB_USER:-cankora}"
  DB_PASSWORD="$(env_value DB_PASSWORD)"
else
  DB_HOST=localhost; DB_PORT=5432; DB_NAME=cankora_db; DB_USER=cankora; DB_PASSWORD=""
fi

export PGPASSWORD="$DB_PASSWORD"

q() {
  "$PSQL" -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    -P pager=off -tAc "$1" 2>/dev/null | tr -d '[:space:]'
}

# ── 2. Database checks ────────────────────────────────────────────────────────
header "1. Database checks ($DB_NAME @ $DB_HOST:$DB_PORT)"

if [ -z "$PSQL" ]; then
  skip "psql tidak ditemukan — lewati DB checks (jalankan: cd backend && make reset-local-verify)"
else
  if ! "$PSQL" -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" >/dev/null 2>&1; then
    skip "Tidak bisa konek ke database — lewati DB checks"
    PSQL=""
  else
    # -- users -------------------------------------------------------------
    USERS="$(q "SELECT count(*) FROM users")"
    if [ "$USERS" = "1" ]; then
      pass "users = 1"
    else
      fail "users = $USERS (expected 1)"
    fi

    ADMIN_EXISTS="$(q "SELECT count(*) FROM users WHERE email='admin@cankora.local' AND deleted_at IS NULL")"
    if [ "$ADMIN_EXISTS" = "1" ]; then
      pass "admin@cankora.local present"
    else
      fail "admin@cankora.local missing (found=$ADMIN_EXISTS)"
    fi

    DEMO_USERS="$(q "SELECT count(*) FROM users WHERE email IN ('pmo@cankora.local','pm@cankora.local','officer@cankora.local','executive@cankora.local','auditor@cankora.local')")"
    if [ "$DEMO_USERS" = "0" ]; then
      pass "no demo users"
    else
      fail "demo users still present = $DEMO_USERS"
    fi

    ADMIN_ROLE="$(q "SELECT count(*) FROM user_roles ur JOIN users u ON u.id=ur.user_id JOIN roles r ON r.id=ur.role_id WHERE u.email='admin@cankora.local' AND r.code='SUPER_ADMIN'")"
    if [ "${ADMIN_ROLE:-0}" -ge 1 ] 2>/dev/null; then
      pass "admin has SUPER_ADMIN role"
    else
      fail "admin SUPER_ADMIN role mapping missing"
    fi

    # -- baseline config ---------------------------------------------------
    ORGS="$(q "SELECT count(*) FROM organizations WHERE deleted_at IS NULL")"
    [ "${ORGS:-0}" -ge 1 ] 2>/dev/null && pass "organizations = $ORGS" || fail "organizations missing"

    ROLES="$(q "SELECT count(*) FROM roles WHERE deleted_at IS NULL")"
    [ "$ROLES" = "7" ] && pass "roles = 7" || fail "roles = $ROLES (expected 7)"

    PERMS="$(q "SELECT count(*) FROM permissions")"
    [ "${PERMS:-0}" -gt 0 ] 2>/dev/null && pass "permissions = $PERMS" || fail "permissions empty"

    RPERMS="$(q "SELECT count(*) FROM role_permissions")"
    [ "${RPERMS:-0}" -gt 0 ] 2>/dev/null && pass "role_permissions = $RPERMS" || fail "role_permissions empty"

    # -- project data ------------------------------------------------------
    PROJECTS="$(q "SELECT count(*) FROM projects")"
    [ "$PROJECTS" = "0" ] && pass "projects = 0" || fail "projects = $PROJECTS (expected 0)"

    DEMO_PROJECTS="$(q "SELECT count(*) FROM projects WHERE code LIKE 'DEMO-SDA-%'")"
    [ "$DEMO_PROJECTS" = "0" ] && pass "no DEMO-SDA-* projects" || fail "DEMO-SDA-* projects = $DEMO_PROJECTS"

    CHILD_DIRTY=""
    for t in tasks milestones issues risks project_budgets contracts project_documents \
             project_teams project_progress_history project_periodic_reports \
             corrective_actions field_inspections field_evidence benefit_indicators \
             project_baselines project_snapshots health_snapshots project_priority_scores \
             report_export_requests data_submissions import_jobs government_sync_runs \
             primavera_sync_runs bim_models vendors audit_logs notifications \
             command_escalations executive_decisions approval_requests; do
      n="$(q "SELECT count(*) FROM $t")"
      if [ -n "$n" ] && [ "$n" != "0" ]; then
        CHILD_DIRTY="$CHILD_DIRTY $t=$n"
      fi
    done
    if [ -z "$CHILD_DIRTY" ]; then
      pass "project child data empty"
    else
      fail "project child data not empty:$CHILD_DIRTY"
    fi

    # -- master data preserved --------------------------------------------
    PROGRAMS="$(q "SELECT count(*) FROM programs WHERE deleted_at IS NULL")"
    [ "${PROGRAMS:-0}" -ge 1 ] 2>/dev/null && pass "master data preserved (programs=$PROGRAMS)" || fail "programs missing"

    # -- schema untouched --------------------------------------------------
    MIGRATIONS="$(q "SELECT count(*) FROM schema_migrations")"
    [ "$MIGRATIONS" = "1" ] && pass "schema_migrations preserved" || fail "schema_migrations = $MIGRATIONS (expected 1)"
  fi
fi

# ── 3. API checks (optional — backend may be stopped after reset) ─────────────
header "2. API checks ($BASE_URL)"

if ! curl -sf -o /dev/null "$BASE_URL/health" 2>/dev/null; then
  skip "backend tidak berjalan di $BASE_URL — lewati API checks"
  skip "  (jalankan: bash scripts/uat-start.sh lalu ulangi smoke ini)"
else
  pass "backend health OK"

  # -- admin login -------------------------------------------------------
  LOGIN="$(curl -s -X POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}")"
  TOKEN="$(json_field "$LOGIN" '.data.access_token')"
  LOGIN_ROLES="$(json_field "$LOGIN" '.data.user.roles | join(",")')"

  if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    pass "admin login OK (roles=$LOGIN_ROLES)"
  else
    fail "admin login failed"
  fi

  # -- demo logins must fail --------------------------------------------
  for u in pmo pm officer executive auditor; do
    CODE="$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API/auth/login" \
      -H "Content-Type: application/json" \
      -d "{\"email\":\"$u@cankora.local\",\"password\":\"$DEMO_PASSWORD\"}")"
    if [ "$CODE" = "401" ]; then
      pass "demo login rejected: $u@cankora.local (401)"
    else
      fail "demo login $u@cankora.local returned $CODE (expected 401)"
    fi
  done

  # -- project list empty ------------------------------------------------
  if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    PROJ="$(curl -s -H "Authorization: Bearer $TOKEN" "$API/projects")"
    TOTAL="$(json_field "$PROJ" '.meta.total')"
    if [ "$TOTAL" = "0" ]; then
      pass "GET /projects total = 0"
    else
      fail "GET /projects total = $TOTAL (expected 0)"
    fi
  else
    skip "project list check skipped (no token)"
  fi
fi

# ── Summary ───────────────────────────────────────────────────────────────────
echo
echo "═══════════════════════════════════════════"
echo " Reset Local (Super Admin only) Smoke"
echo "═══════════════════════════════════════════"
echo " PASS: $PASS  FAIL: $FAIL  SKIP: $SKIP"
echo "═══════════════════════════════════════════"
if [ "$FAIL" -gt 0 ]; then
  echo " ✗ RESET VERIFICATION FAILED"
  exit 1
fi
echo " ✓ RESET VERIFIED — Super Admin only"
exit 0
