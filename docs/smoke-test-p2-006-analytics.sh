#!/usr/bin/env bash
# Smoke test P2-006 — Program & Sector Dashboard Analytics
# Usage: bash docs/smoke-test-p2-006-analytics.sh

set -uo pipefail

BASE="http://localhost:8080/api/v1"
PASS=0
FAIL=0
TOTAL=0

green() { printf '\033[32m✓ %s\033[0m\n' "$1"; }
red()   { printf '\033[31m✗ %s\033[0m\n' "$1"; }

pass() { PASS=$((PASS+1)); TOTAL=$((TOTAL+1)); green "$1"; }
fail() { FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1)); red "$1"; }

check_http() {
  local label="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    pass "$label (HTTP $actual)"
  else
    fail "$label (expected $expected, got $actual)"
  fi
}

check_field() {
  local label="$1" value="$2"
  if [ -n "$value" ] && [ "$value" != "null" ]; then
    pass "$label"
  else
    fail "$label (empty or null)"
  fi
}

check_num() {
  local label="$1" value="$2"
  if echo "$value" | grep -qE '^[0-9]+$'; then
    pass "$label ($value)"
  else
    fail "$label (not a number: $value)"
  fi
}

echo "========================================"
echo " CANKORA Smoke Test — P2-006 Analytics"
echo "========================================"
echo ""

# ------------------------------------------------------------------
# 1. Login
# ------------------------------------------------------------------
echo "── 1. Auth ──────────────────────────────"
LOGIN=$(curl -s -w '\n%{http_code}' -X POST "$BASE/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}')
LOGIN_BODY=$(echo "$LOGIN" | sed '$d')
LOGIN_CODE=$(echo "$LOGIN" | tail -1)
check_http "Login" "200" "$LOGIN_CODE"

TOKEN=$(echo "$LOGIN_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('access_token',''))" 2>/dev/null)
if [ -z "$TOKEN" ]; then
  fail "Extract JWT token"
  echo "ABORT: No token, cannot continue."
  exit 1
fi
pass "Extract JWT token"

AUTH="-H 'Authorization: Bearer $TOKEN'"
get() { curl -s -w '\n%{http_code}' -H "Authorization: Bearer $TOKEN" "$BASE$1"; }

# ------------------------------------------------------------------
# 2. List Programs
# ------------------------------------------------------------------
echo ""
echo "── 2. List Programs ─────────────────────"
R=$(get "/analytics/programs")
BODY=$(echo "$R" | sed '$d')
CODE=$(echo "$R" | tail -1)
check_http "GET /analytics/programs" "200" "$CODE"

GROUPS_LEN=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d.get('data',{}).get('groups',[])))" 2>/dev/null)
check_num "programs.groups count" "$GROUPS_LEN"

AS_OF=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('as_of',''))" 2>/dev/null)
check_field "programs.as_of present" "$AS_OF"

# Extract first program id
PROG_ID=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); gs=d.get('data',{}).get('groups',[]); print(gs[0]['group_id'] if gs else '')" 2>/dev/null)
check_field "programs.groups[0].group_id" "$PROG_ID"

PROG_CODE=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); gs=d.get('data',{}).get('groups',[]); print(gs[0].get('group_code','') if gs else '')" 2>/dev/null)
check_field "programs.groups[0].group_code" "$PROG_CODE"

PROG_TOTAL=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); gs=d.get('data',{}).get('groups',[]); print(gs[0].get('total_projects',0) if gs else 0)" 2>/dev/null)
check_num "programs.groups[0].total_projects" "$PROG_TOTAL"

PROG_BUDGET=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); gs=d.get('data',{}).get('groups',[]); v=gs[0].get('total_budget',0) if gs else 0; print(int(v))" 2>/dev/null)
check_num "programs.groups[0].total_budget is numeric" "$PROG_BUDGET"

# ------------------------------------------------------------------
# 3. Get Program Detail
# ------------------------------------------------------------------
echo ""
echo "── 3. Get Program Detail ────────────────"
R=$(get "/analytics/programs/$PROG_ID")
BODY=$(echo "$R" | sed '$d')
CODE=$(echo "$R" | tail -1)
check_http "GET /analytics/programs/:id" "200" "$CODE"

KPI_NAME=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('kpi',{}).get('group_name',''))" 2>/dev/null)
check_field "program.kpi.group_name" "$KPI_NAME"

KPI_TOTAL=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('kpi',{}).get('total_projects',0))" 2>/dev/null)
check_num "program.kpi.total_projects" "$KPI_TOTAL"

PROJECTS_ARR=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d.get('data',{}).get('projects',[])))" 2>/dev/null)
check_num "program.projects count" "$PROJECTS_ARR"

# projects array has required fields
if [ "$PROJECTS_ARR" -gt 0 ] 2>/dev/null; then
  P_CODE=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); ps=d.get('data',{}).get('projects',[]); print(ps[0].get('project_code','') if ps else '')" 2>/dev/null)
  check_field "program.projects[0].project_code" "$P_CODE"
  P_STATUS=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); ps=d.get('data',{}).get('projects',[]); print(ps[0].get('status','') if ps else '')" 2>/dev/null)
  check_field "program.projects[0].status" "$P_STATUS"
fi

TOP_PHYS=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(type(d.get('data',{}).get('top_physical_deviation',None)).__name__)" 2>/dev/null)
if [ "$TOP_PHYS" = "list" ]; then pass "program.top_physical_deviation is array"; else fail "program.top_physical_deviation is array (got $TOP_PHYS)"; fi

TOP_BUDGET=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(type(d.get('data',{}).get('top_budget_deviation',None)).__name__)" 2>/dev/null)
if [ "$TOP_BUDGET" = "list" ]; then pass "program.top_budget_deviation is array"; else fail "program.top_budget_deviation is array (got $TOP_BUDGET)"; fi

HIGH_RISK=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(type(d.get('data',{}).get('high_risk_projects',None)).__name__)" 2>/dev/null)
if [ "$HIGH_RISK" = "list" ]; then pass "program.high_risk_projects is array"; else fail "program.high_risk_projects is array (got $HIGH_RISK)"; fi

# ------------------------------------------------------------------
# 4. List Sectors
# ------------------------------------------------------------------
echo ""
echo "── 4. List Sectors ──────────────────────"
R=$(get "/analytics/sectors")
BODY=$(echo "$R" | sed '$d')
CODE=$(echo "$R" | tail -1)
check_http "GET /analytics/sectors" "200" "$CODE"

SEC_LEN=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d.get('data',{}).get('groups',[])))" 2>/dev/null)
check_num "sectors.groups count" "$SEC_LEN"

SEC_ID=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); gs=d.get('data',{}).get('groups',[]); print(gs[0]['group_id'] if gs else '')" 2>/dev/null)
check_field "sectors.groups[0].group_id" "$SEC_ID"

SEC_TYPE=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); gs=d.get('data',{}).get('groups',[]); print(gs[0].get('group_type','') if gs else '')" 2>/dev/null)
if [ "$SEC_TYPE" = "sector" ]; then pass "sectors.groups[0].group_type == sector"; else fail "sectors.groups[0].group_type == sector (got $SEC_TYPE)"; fi

# ------------------------------------------------------------------
# 5. Get Sector Detail
# ------------------------------------------------------------------
echo ""
echo "── 5. Get Sector Detail ─────────────────"
R=$(get "/analytics/sectors/$SEC_ID")
BODY=$(echo "$R" | sed '$d')
CODE=$(echo "$R" | tail -1)
check_http "GET /analytics/sectors/:id" "200" "$CODE"

SEC_KPI=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('kpi',{}).get('group_name',''))" 2>/dev/null)
check_field "sector.kpi.group_name" "$SEC_KPI"

SEC_PROJS=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d.get('data',{}).get('projects',[])))" 2>/dev/null)
check_num "sector.projects count" "$SEC_PROJS"

# ------------------------------------------------------------------
# 6. Error cases
# ------------------------------------------------------------------
echo ""
echo "── 6. Error cases ───────────────────────"

R=$(get "/analytics/programs/00000000-0000-0000-0000-000000000000")
CODE=$(echo "$R" | tail -1)
check_http "GET /analytics/programs/:id (not found) → 404" "404" "$CODE"

R=$(get "/analytics/sectors/not-a-uuid")
CODE=$(echo "$R" | tail -1)
check_http "GET /analytics/sectors/:id (bad uuid) → 400" "400" "$CODE"

# Unauthenticated
R=$(curl -s -w '\n%{http_code}' "$BASE/analytics/programs")
CODE=$(echo "$R" | tail -1)
check_http "GET /analytics/programs (no auth) → 401" "401" "$CODE"

# ------------------------------------------------------------------
# 7. Period filter
# ------------------------------------------------------------------
echo ""
echo "── 7. Period filter ─────────────────────"
R=$(get "/analytics/programs?year=2025&month=6")
CODE=$(echo "$R" | tail -1)
check_http "GET /analytics/programs?year=2025&month=6" "200" "$CODE"

R=$(get "/analytics/programs/$PROG_ID?year=2025&month=6")
CODE=$(echo "$R" | tail -1)
check_http "GET /analytics/programs/:id?year=2025&month=6" "200" "$CODE"

# ------------------------------------------------------------------
# Summary
# ------------------------------------------------------------------
echo ""
echo "========================================"
printf "  Result: %d/%d passed" "$PASS" "$TOTAL"
if [ "$FAIL" -eq 0 ]; then
  printf " \033[32m— ALL PASS ✓\033[0m\n"
else
  printf " \033[31m— %d FAILED ✗\033[0m\n" "$FAIL"
fi
echo "========================================"

[ "$FAIL" -eq 0 ]
