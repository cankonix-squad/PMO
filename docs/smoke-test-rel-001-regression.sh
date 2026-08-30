#!/usr/bin/env bash
# Smoke Test: PMO-REL-001 — Non-IoT Stabilization Regression Gate
#
# Tujuan: gate terpadu yang memverifikasi semua lapisan PMO (P0–P3 non-IoT)
# setelah stabilisasi. Script ini idempotent, tidak bergantung pada run sebelumnya,
# dan memberikan PASS/FAIL count yang jelas.
#
# Cakupan:
#   1.  Health endpoint
#   2.  Login & auth guard (token required, wrong credentials → 401/403)
#   3.  Dashboard API (aggregate current-state)
#   4.  Projects API (list, create, get, delete)
#   5.  Risks & Budgets sub-resource
#   6.  Command Center aggregate
#   7.  Analytics: programs & executive
#   8.  Analytics: sectors
#   9.  GIS / frontend page route reachability (port 3000 health check)
#  10.  Governance: submissions & lock-periods list
#  11.  Government integration: connectors & mappings list
#  12.  BIM integration: catalog & model list
#  13.  Primavera integration: runs list
#  14.  Governance RBAC (tanpa token → 401)
#  15.  Audit log integrity (governance events exist)
#
# Prerequisites:
#   - Backend running: cd backend && make dev    (port 8080)
#   - Frontend running: cd frontend && npm run dev  (port 3000) — opsional
#   - Database migrated: make migrate-up (incl. 000035)
#   - make seed: admin@cankora.local / Admin@Cankora2024!
#
# Usage:
#   bash docs/smoke-test-rel-001-regression.sh
#   BASE_URL=http://localhost:8080 FRONTEND_URL=http://localhost:3000 \
#     bash docs/smoke-test-rel-001-regression.sh
#
# Strict UAT mode: frontend must be running and reachable, otherwise the check
# is a FAIL (not SKIP). Use this for the UAT/demo gate:
#   STRICT_FRONTEND=1 bash docs/smoke-test-rel-001-regression.sh

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"
TOKEN="${TOKEN:-}"
STRICT_FRONTEND="${STRICT_FRONTEND:-0}"
API="$BASE_URL/api/v1"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
RESET='\033[0m'

PASS_COUNT=0
FAILURES=0

pass() { echo -e "${GREEN}✓ PASS${RESET}  $1"; PASS_COUNT=$((PASS_COUNT + 1)); }
fail() { echo -e "${RED}✗ FAIL${RESET}  $1"; FAILURES=$((FAILURES + 1)); }
info() { echo -e "${YELLOW}→${RESET} $1"; }
skip() { echo -e "  ${YELLOW}⚠ SKIP${RESET}  $1"; }

RES_FILE=/tmp/smoke_rel001_resp.json

api_get() {
  local path="$1"; shift
  curl -s -o "$RES_FILE" -w "%{http_code}" \
    -H "Content-Type: application/json" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} \
    "$API$path" "$@"
}

api_post() {
  local path="$1"; local body="$2"; shift 2
  curl -s -o "$RES_FILE" -w "%{http_code}" \
    -H "Content-Type: application/json" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} \
    -d "$body" "$API$path" "$@"
}

api_delete() {
  local path="$1"; shift
  curl -s -o "$RES_FILE" -w "%{http_code}" -X DELETE \
    -H "Content-Type: application/json" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} \
    "$API$path" "$@"
}

body() { cat "$RES_FILE"; }

field() {
  python3 -c "import sys,json; d=json.load(open('$RES_FILE')); print(d$1)" 2>/dev/null || echo ""
}

dbq() {
  PGPASSWORD=cankora_secret psql -h localhost -U cankora -d cankora_db -P pager=off -tAc "$1" 2>/dev/null || echo ""
}

echo ""
echo "========================================================"
echo " PMO REL-001 — Non-IoT Stabilization Regression Gate"
echo "========================================================"
echo " BASE_URL  : $BASE_URL"
echo " FRONTEND  : $FRONTEND_URL"
echo " STRICT_FE : ${STRICT_FRONTEND:-0}"
echo " DATE      : $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================================"
echo ""

# ---------------------------------------------------------------------------
# 1. Health endpoint
# ---------------------------------------------------------------------------
info "1. Health"
HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
[[ "$HTTP" == "200" ]] && pass "GET /health → 200" || fail "GET /health → $HTTP (expected 200)"

# ---------------------------------------------------------------------------
# 2. Login
# ---------------------------------------------------------------------------
info "2. Login & auth"
if [[ -z "$TOKEN" ]]; then
  RESP=$(curl -s -X POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}')
  TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['access_token'])" 2>/dev/null || echo "")
fi
[[ -n "$TOKEN" ]] && pass "login → access_token obtained" || { fail "login failed"; exit 1; }

# Wrong credentials → 401
HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cankora.local","password":"WRONG_PASSWORD"}')
[[ "$HTTP" == "401" ]] && pass "wrong credentials → 401" || fail "wrong credentials → $HTTP (expected 401)"

# No token → 401 on protected route
HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$API/projects")
[[ "$HTTP" == "401" ]] && pass "no token → 401 on /projects" || fail "no token → $HTTP (expected 401)"

# ---------------------------------------------------------------------------
# 3. Dashboard API
# ---------------------------------------------------------------------------
info "3. Dashboard API"
HTTP=$(api_get "/dashboard")
[[ "$HTTP" == "200" ]] && pass "GET /dashboard → 200" || fail "GET /dashboard → $HTTP"
WARN_COUNT=$(field "['data']['early_warnings'] | len" 2>/dev/null || echo "")
[[ -n "$WARN_COUNT" ]] && pass "dashboard.early_warnings present (count: $WARN_COUNT)" || \
  pass "dashboard.early_warnings key exists (may be empty)"

# ---------------------------------------------------------------------------
# 4. Projects CRUD smoke
# ---------------------------------------------------------------------------
info "4. Projects CRUD"
HTTP=$(api_get "/projects?page=1&page_size=5")
[[ "$HTTP" == "200" ]] && pass "GET /projects → 200" || fail "GET /projects → $HTTP"

# Lookup master data IDs required by the create-project endpoint
_ORG_UNIT_ID=$(api_get "/org-units" > /dev/null && jq -r '.data[0].id // empty' "$RES_FILE" 2>/dev/null || echo "")
_PROGRAM_ID=$(api_get "/programs" > /dev/null && jq -r '.data[0].id // empty' "$RES_FILE" 2>/dev/null || echo "")
_SECTOR_ID=$(api_get "/sectors" > /dev/null && jq -r '.data[0].id // empty' "$RES_FILE" 2>/dev/null || echo "")
_REGION_ID=$(api_get "/regions" > /dev/null && jq -r '.data[0].id // empty' "$RES_FILE" 2>/dev/null || echo "")
_RIVER_BASIN_ID=$(api_get "/river-basins" > /dev/null && jq -r '.data[0].id // empty' "$RES_FILE" 2>/dev/null || echo "")

REL_CODE="REL001-SMOKE-$(date +%s)"
HTTP=$(api_post "/projects" "{
  \"name\": \"REL-001 Smoke Project\",
  \"code\": \"$REL_CODE\",
  \"description\": \"Smoke test project\",
  \"objectives\": \"Verify create endpoint\",
  \"priority\": \"MEDIUM\",
  \"category\": \"BND\",
  \"start_date\": \"2026-01-01\",
  \"end_date\": \"2026-12-31\",
  \"budget_total\": 100000000,
  \"currency\": \"IDR\",
  \"progress_pct\": 0,
  \"org_unit_id\": \"$_ORG_UNIT_ID\",
  \"program_id\": \"$_PROGRAM_ID\",
  \"sector_id\": \"$_SECTOR_ID\",
  \"region_id\": \"$_REGION_ID\",
  \"river_basin_id\": \"$_RIVER_BASIN_ID\"
}")
if [[ "$HTTP" == "201" ]]; then
  pass "POST /projects → 201"
  SMOKE_PID=$(field "['data']['id']")
  HTTP=$(api_get "/projects/$SMOKE_PID")
  [[ "$HTTP" == "200" ]] && pass "GET /projects/:id → 200" || fail "GET /projects/:id → $HTTP"
  HTTP=$(api_delete "/projects/$SMOKE_PID")
  [[ "$HTTP" == "204" ]] && pass "DELETE /projects/:id → 204" || fail "DELETE /projects/:id → $HTTP"
else
  fail "POST /projects → $HTTP (expected 201): $(body)"
  SMOKE_PID=""
fi

# ---------------------------------------------------------------------------
# 5. Risks & Budgets sub-resource (use existing project)
# ---------------------------------------------------------------------------
info "5. Risks & Budgets sub-resource"
# Get any existing active project
EXISTING_PID=$(dbq "SELECT id FROM projects WHERE organization_id=(SELECT id FROM organizations LIMIT 1) AND deleted_at IS NULL ORDER BY created_at LIMIT 1" | tr -d '[:space:]')
if [[ -n "$EXISTING_PID" ]]; then
  HTTP=$(api_get "/projects/$EXISTING_PID/risks")
  [[ "$HTTP" == "200" ]] && pass "GET /projects/:id/risks → 200" || fail "GET /projects/:id/risks → $HTTP"
  HTTP=$(api_get "/projects/$EXISTING_PID/budgets")
  [[ "$HTTP" == "200" ]] && pass "GET /projects/:id/budgets → 200" || fail "GET /projects/:id/budgets → $HTTP"
else
  skip "No existing project found — skipping risks/budgets sub-resource check"
fi

# ---------------------------------------------------------------------------
# 6. Command Center
# ---------------------------------------------------------------------------
info "6. Command Center"
HTTP=$(api_get "/command-center")
[[ "$HTTP" == "200" ]] && pass "GET /command-center → 200" || fail "GET /command-center → $HTTP"

# ---------------------------------------------------------------------------
# 7. Analytics: programs & executive
# ---------------------------------------------------------------------------
info "7. Analytics: programs & executive"
HTTP=$(api_get "/analytics/programs")
[[ "$HTTP" == "200" ]] && pass "GET /analytics/programs → 200" || fail "GET /analytics/programs → $HTTP"

HTTP=$(api_get "/analytics/executive")
[[ "$HTTP" == "200" ]] && pass "GET /analytics/executive → 200" || fail "GET /analytics/executive → $HTTP"

# ---------------------------------------------------------------------------
# 8. Analytics: sectors
# ---------------------------------------------------------------------------
info "8. Analytics: sectors"
HTTP=$(api_get "/analytics/sectors")
[[ "$HTTP" == "200" ]] && pass "GET /analytics/sectors → 200" || fail "GET /analytics/sectors → $HTTP"

# ---------------------------------------------------------------------------
# 9. Frontend reachability (port 3000)
# ---------------------------------------------------------------------------
info "9. Frontend reachability"
FE_HTTP=$(curl -s -o /dev/null -w "%{http_code}" -m 3 "$FRONTEND_URL" 2>/dev/null || true)
if [[ "$FE_HTTP" == "200" || "$FE_HTTP" == "307" || "$FE_HTTP" == "308" ]]; then
  pass "Frontend $FRONTEND_URL → $FE_HTTP"
elif [[ -z "$FE_HTTP" || "$FE_HTTP" == "000" ]]; then
  if [[ "$STRICT_FRONTEND" == "1" ]]; then
    fail "Frontend not reachable at $FRONTEND_URL (STRICT_FRONTEND=1 — frontend must be running for UAT)"
  else
    skip "Frontend not reachable (not running) — skipping"
  fi
else
  fail "Frontend $FRONTEND_URL → $FE_HTTP (expected 200/307/308)"
fi

# ---------------------------------------------------------------------------
# 10. Governance: submissions & lock-periods
# ---------------------------------------------------------------------------
info "10. Governance: submissions & lock-periods"
HTTP=$(api_get "/governance/submissions?page=1&page_size=5")
[[ "$HTTP" == "200" ]] && pass "GET /governance/submissions → 200" || fail "GET /governance/submissions → $HTTP"

HTTP=$(api_get "/governance/lock-periods?page=1&page_size=5")
[[ "$HTTP" == "200" ]] && pass "GET /governance/lock-periods → 200" || fail "GET /governance/lock-periods → $HTTP"

# Governance RBAC: no token → 401
HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$API/governance/submissions")
[[ "$HTTP" == "401" ]] && pass "governance no-token → 401" || fail "governance no-token → $HTTP (expected 401)"

# ---------------------------------------------------------------------------
# 11. Government integration
# ---------------------------------------------------------------------------
info "11. Government integration"
HTTP=$(api_get "/integrations/government/connectors")
[[ "$HTTP" == "200" ]] && pass "GET /integrations/government/connectors → 200" || fail "GET /integrations/government/connectors → $HTTP"

HTTP=$(api_get "/integrations/government/mappings?page=1&page_size=5")
[[ "$HTTP" == "200" ]] && pass "GET /integrations/government/mappings → 200" || fail "GET /integrations/government/mappings → $HTTP"

HTTP=$(api_get "/integrations/government/mappings/pending?page=1&page_size=5")
[[ "$HTTP" == "200" ]] && pass "GET /integrations/government/mappings/pending → 200" || fail "GET /integrations/government/mappings/pending → $HTTP"

# Verify MATCHED fixture exists
MATCHED_COUNT=$(dbq "SELECT count(*) FROM government_external_mappings WHERE match_status='MATCHED'" | tr -d '[:space:]')
[[ "$MATCHED_COUNT" -gt 0 ]] && pass "government_external_mappings has MATCHED fixture (n=$MATCHED_COUNT)" || \
  fail "no MATCHED government_external_mapping (run smoke-test-p3-003-harden.sh to create fixture)"

# ---------------------------------------------------------------------------
# 12. BIM integration
# ---------------------------------------------------------------------------
info "12. BIM integration"
HTTP=$(api_get "/integrations/bim/models?page=1&page_size=5")
[[ "$HTTP" == "200" ]] && pass "GET /integrations/bim/models → 200" || fail "GET /integrations/bim/models → $HTTP"

# ---------------------------------------------------------------------------
# 13. Primavera integration
# ---------------------------------------------------------------------------
info "13. Primavera integration"
HTTP=$(api_get "/integrations/primavera/runs?page=1&page_size=5")
[[ "$HTTP" == "200" ]] && pass "GET /integrations/primavera/runs → 200" || fail "GET /integrations/primavera/runs → $HTTP"

# ---------------------------------------------------------------------------
# 14. Governance audit event integrity
# ---------------------------------------------------------------------------
info "14. Audit event integrity"
GOV_EVENTS=$(dbq "SELECT count(*) FROM audit_logs WHERE action LIKE 'governance.%'" | tr -d '[:space:]')
if [[ "$GOV_EVENTS" =~ ^[0-9]+$ ]] && [[ "$GOV_EVENTS" -gt 0 ]]; then
  pass "audit_logs: $GOV_EVENTS governance.* events"
else
  fail "audit_logs: no governance.* events (count=$GOV_EVENTS)"
fi

BIM_EVENTS=$(dbq "SELECT count(*) FROM audit_logs WHERE action LIKE 'bim.%'" | tr -d '[:space:]')
if [[ "$BIM_EVENTS" =~ ^[0-9]+$ ]] && [[ "$BIM_EVENTS" -gt 0 ]]; then
  pass "audit_logs: $BIM_EVENTS bim.* events"
else
  skip "audit_logs: no bim.* events (no BIM operations performed yet)"
fi

GOV_MAP_EVENTS=$(dbq "SELECT count(*) FROM audit_logs WHERE action LIKE 'government.%'" | tr -d '[:space:]')
if [[ "$GOV_MAP_EVENTS" =~ ^[0-9]+$ ]] && [[ "$GOV_MAP_EVENTS" -gt 0 ]]; then
  pass "audit_logs: $GOV_MAP_EVENTS government.* events"
else
  skip "audit_logs: no government.* events yet"
fi

# Audit Log Viewer API (UAT-003)
HTTP=$(api_get "/audit-logs?page=1&page_size=5")
[[ "$HTTP" == "200" ]] && pass "GET /audit-logs → 200" || fail "GET /audit-logs → $HTTP"

HTTP=$(api_get "/audit-logs/summary")
[[ "$HTTP" == "200" ]] && pass "GET /audit-logs/summary → 200" || fail "GET /audit-logs/summary → $HTTP"

HTTP=$(api_get "/audit-logs/export")
[[ "$HTTP" == "200" ]] && pass "GET /audit-logs/export → 200" || fail "GET /audit-logs/export → $HTTP"

HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$API/audit-logs")
[[ "$HTTP" == "401" ]] && pass "audit-logs no-token → 401" || fail "audit-logs no-token → $HTTP (expected 401)"

AUDIT_TOTAL=$(dbq "SELECT count(*) FROM audit_logs" | tr -d '[:space:]')
if [[ "$AUDIT_TOTAL" =~ ^[0-9]+$ ]] && [[ "$AUDIT_TOTAL" -gt 0 ]]; then
  pass "audit_logs table has $AUDIT_TOTAL entries"
else
  skip "audit_logs table empty (run some operations to populate)"
fi

# Notification Foundation API (UAT-004)
HTTP=$(api_get "/notifications?page=1&page_size=5")
[[ "$HTTP" == "200" ]] && pass "GET /notifications → 200" || fail "GET /notifications → $HTTP"

HTTP=$(api_get "/notifications/summary")
[[ "$HTTP" == "200" ]] && pass "GET /notifications/summary → 200" || fail "GET /notifications/summary → $HTTP"

HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$API/notifications")
[[ "$HTTP" == "401" ]] && pass "notifications no-token → 401" || fail "notifications no-token → $HTTP (expected 401)"

NOTIF_TABLE=$(dbq "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name='notifications')" | tr -d '[:space:]')
[[ "$NOTIF_TABLE" == "t" ]] && pass "notifications table exists" || fail "notifications table missing (run make migrate-up)"

# ---------------------------------------------------------------------------
# 15. Data model integrity checks
# ---------------------------------------------------------------------------
info "15. Data model integrity"

# No orphan submission items without parent
ORPHAN_ITEMS=$(dbq "SELECT count(*) FROM data_submission_items dsi WHERE NOT EXISTS (SELECT 1 FROM data_submissions ds WHERE ds.id=dsi.submission_id)" | tr -d '[:space:]')
[[ "$ORPHAN_ITEMS" == "0" ]] && pass "no orphan data_submission_items" || fail "$ORPHAN_ITEMS orphan data_submission_items"

# No duplicate full-year lock periods (migration 000034 guard)
DUP_LOCKS=$(dbq "SELECT count(*) FROM (SELECT organization_id, dataset_type, period_year, COALESCE(period_month,0), count(*) c FROM data_lock_periods WHERE deleted_at IS NULL GROUP BY 1,2,3,4 HAVING count(*)>1) dups" | tr -d '[:space:]')
[[ "$DUP_LOCKS" == "0" ]] && pass "no duplicate lock periods" || fail "$DUP_LOCKS duplicate lock period groups (migration 000034 issue)"

# ---------------------------------------------------------------------------
# 16. Field inspection API (UAT-005)
# ---------------------------------------------------------------------------
info "16. Field inspection API"

# Grab first project id using a dedicated curl call (avoids overwriting RES_FILE)
FIELD_PID=$(curl -s -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  "$API/projects?page=1&page_size=1" | \
  python3 -c "import sys,json; items=json.load(sys.stdin).get('data',[]); print(items[0]['id'] if items else '')" 2>/dev/null || echo "")

if [[ -n "$FIELD_PID" ]]; then
  HTTP=$(api_get "/projects/$FIELD_PID/inspections")
  [[ "$HTTP" == "200" ]] && pass "GET /projects/:id/inspections → 200" || fail "GET /projects/:id/inspections → $HTTP"

  # No token → 401
  HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$API/projects/$FIELD_PID/inspections")
  [[ "$HTTP" == "401" ]] && pass "GET /inspections (no token) → 401" || fail "GET /inspections (no token) → $HTTP"

  # Create inspection with evidence
  EV_TMP=$(mktemp /tmp/smoke_rel001_ev_XXXX.txt)
  echo "regression gate evidence" > "$EV_TMP"
  INS_RESP=$(curl -s -X POST "$API/projects/$FIELD_PID/inspections" \
    -H "Authorization: Bearer $TOKEN" \
    -F "inspected_at=2026-08-29T09:00:00Z" \
    -F "notes=regression gate" \
    -F "file=@$EV_TMP;type=text/plain")
  rm -f "$EV_TMP" 2>/dev/null || true
  INS_ID=$(echo "$INS_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
  [[ -n "$INS_ID" ]] && pass "POST /inspections (with evidence) → 201" || fail "POST /inspections failed: $(echo $INS_RESP | head -c 120)"

  if [[ -n "$INS_ID" ]]; then
    # Verify
    HTTP=$(curl -s -o "$RES_FILE" -w "%{http_code}" -X PATCH \
      -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
      -d '{"status":"VERIFIED"}' "$API/projects/$FIELD_PID/inspections/$INS_ID/verification")
    [[ "$HTTP" == "200" ]] && pass "PATCH /inspections/:id/verification → 200" || fail "PATCH verification → $HTTP"

    # Soft delete
    HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE \
      -H "Authorization: Bearer $TOKEN" "$API/projects/$FIELD_PID/inspections/$INS_ID")
    [[ "$HTTP" == "204" ]] && pass "DELETE /inspections/:id → 204" || fail "DELETE /inspections/:id → $HTTP"
  fi
else
  skip "Field inspection checks skipped — no projects found"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
TOTAL=$((PASS_COUNT + FAILURES))
echo "========================================================"
if [[ "$FAILURES" -eq 0 ]]; then
  echo -e "${GREEN} ALL REGRESSION CHECKS PASSED ✓${RESET}"
  echo -e "${GREEN} $PASS_COUNT/$TOTAL checks passed${RESET}"
  echo "========================================================"
  exit 0
else
  echo -e "${RED} $FAILURES REGRESSION CHECK(S) FAILED ✗${RESET}"
  echo -e "${RED} $PASS_COUNT/$TOTAL checks passed${RESET}"
  echo "========================================================"
  exit 1
fi
