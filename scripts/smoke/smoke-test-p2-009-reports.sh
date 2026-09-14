#!/usr/bin/env bash
# smoke-test-p2-009-reports.sh
# P2-009 Reporting Integration — smoke tests
# Usage: BASE_URL=http://localhost:8080 bash smoke-test-p2-009-reports.sh

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@cankora.local}"
ADMIN_PASS="${ADMIN_PASS:-Admin@Cankora2024!}"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0
FAIL=0
SKIP=0

TMPBODY=$(mktemp)
trap 'rm -f "$TMPBODY"' EXIT

pass() { echo -e "${GREEN}PASS${NC} $1"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}FAIL${NC} $1"; FAIL=$((FAIL+1)); }
skip() { echo -e "${YELLOW}SKIP${NC} $1"; SKIP=$((SKIP+1)); }

# curl_req <method> <url> [curl extra args...]
# Sets: HTTP_CODE, HTTP_BODY
curl_req() {
  local method="$1" url="$2"
  shift 2
  HTTP_CODE=$(curl -s -o "$TMPBODY" -w "%{http_code}" -X "$method" "$url" "$@")
  HTTP_BODY=$(cat "$TMPBODY")
}

assert_status() {
  local label="$1" expected="$2" actual="$3"
  if [[ "$actual" == "$expected" ]]; then
    pass "$label (HTTP $actual)"
  else
    fail "$label — expected HTTP $expected, got $actual"
  fi
}

assert_field() {
  local label="$1" field="$2" body="$3"
  if echo "$body" | grep -q "\"$field\""; then
    pass "$label (field '$field' present)"
  else
    fail "$label — field '$field' missing in response"
  fi
}

echo ""
echo "================================================="
echo "  P2-009 Reporting Integration — Smoke Test"
echo "  Target: $BASE_URL"
echo "================================================="
echo ""

# ---------------------------------------------------------------------------
# 0. Auth
# ---------------------------------------------------------------------------
echo "--- [0] Auth ---"
curl_req POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASS\"}"
assert_status "[0-1] Login" "200" "$HTTP_CODE"

TOKEN=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['access_token'])" 2>/dev/null || true)
if [[ -z "$TOKEN" ]]; then
  echo -e "${RED}FATAL: Could not extract auth token — aborting${NC}"
  exit 1
fi
pass "[0-2] Token extracted"

AUTH_H="Authorization: Bearer $TOKEN"

# ---------------------------------------------------------------------------
# 1. Report Catalog
# ---------------------------------------------------------------------------
echo ""
echo "--- [1] Report Catalog ---"

curl_req GET "$BASE_URL/api/v1/analytics/reports/catalog" -H "$AUTH_H"
assert_status "[1-1] GET /analytics/reports/catalog" "200" "$HTTP_CODE"
assert_field  "[1-2] Catalog response has 'data'" "data" "$HTTP_BODY"

# ---------------------------------------------------------------------------
# 2. Dataset: Executive Summary
# ---------------------------------------------------------------------------
echo ""
echo "--- [2] Dataset: Executive Summary ---"

curl_req GET "$BASE_URL/api/v1/analytics/reports/datasets/executive-summary" -H "$AUTH_H"
assert_status "[2-1] GET /datasets/executive-summary" "200" "$HTTP_CODE"
assert_field  "[2-2] Has total_projects"    "total_projects"    "$HTTP_BODY"
assert_field  "[2-3] Has total_budget_plan" "total_budget_plan" "$HTTP_BODY"
assert_field  "[2-4] Has green_health"      "green_health"      "$HTTP_BODY"
assert_field  "[2-5] Has open_risks"        "open_risks"        "$HTTP_BODY"

curl_req GET "$BASE_URL/api/v1/analytics/reports/datasets/executive-summary?status=ACTIVE" -H "$AUTH_H"
assert_status "[2-6] GET /datasets/executive-summary?status=ACTIVE" "200" "$HTTP_CODE"

# ---------------------------------------------------------------------------
# 3. Dataset: Project Performance
# ---------------------------------------------------------------------------
echo ""
echo "--- [3] Dataset: Project Performance ---"

curl_req GET "$BASE_URL/api/v1/analytics/reports/datasets/project-performance" -H "$AUTH_H"
assert_status "[3-1] GET /datasets/project-performance" "200" "$HTTP_CODE"
assert_field  "[3-2] Has data" "data" "$HTTP_BODY"

curl_req GET "$BASE_URL/api/v1/analytics/reports/datasets/project-performance?province=Jawa%20Barat" -H "$AUTH_H"
assert_status "[3-3] GET /datasets/project-performance?province=Jawa+Barat" "200" "$HTTP_CODE"

# ---------------------------------------------------------------------------
# 4. Dataset: Risk & Issue
# ---------------------------------------------------------------------------
echo ""
echo "--- [4] Dataset: Risk & Issue ---"

curl_req GET "$BASE_URL/api/v1/analytics/reports/datasets/risk-issue" -H "$AUTH_H"
assert_status "[4-1] GET /datasets/risk-issue" "200" "$HTTP_CODE"
assert_field  "[4-2] Has data" "data" "$HTTP_BODY"

# ---------------------------------------------------------------------------
# 5. Dataset: Budget
# ---------------------------------------------------------------------------
echo ""
echo "--- [5] Dataset: Budget ---"

curl_req GET "$BASE_URL/api/v1/analytics/reports/datasets/budget" -H "$AUTH_H"
assert_status "[5-1] GET /datasets/budget" "200" "$HTTP_CODE"
assert_field  "[5-2] Has data" "data" "$HTTP_BODY"

curl_req GET "$BASE_URL/api/v1/analytics/reports/datasets/budget?status=ACTIVE" -H "$AUTH_H"
assert_status "[5-3] GET /datasets/budget?status=ACTIVE" "200" "$HTTP_CODE"

# ---------------------------------------------------------------------------
# 6. Dataset: Benefits
# ---------------------------------------------------------------------------
echo ""
echo "--- [6] Dataset: Benefits ---"

curl_req GET "$BASE_URL/api/v1/analytics/reports/datasets/benefits" -H "$AUTH_H"
assert_status "[6-1] GET /datasets/benefits" "200" "$HTTP_CODE"
assert_field  "[6-2] Has data" "data" "$HTTP_BODY"

# ---------------------------------------------------------------------------
# 7. Dataset: Priority
# ---------------------------------------------------------------------------
echo ""
echo "--- [7] Dataset: Priority ---"

curl_req GET "$BASE_URL/api/v1/analytics/reports/datasets/priority" -H "$AUTH_H"
assert_status "[7-1] GET /datasets/priority" "200" "$HTTP_CODE"
assert_field  "[7-2] Has data" "data" "$HTTP_BODY"

# ---------------------------------------------------------------------------
# 8. Power BI Config
# ---------------------------------------------------------------------------
echo ""
echo "--- [8] Power BI Config ---"

curl_req GET "$BASE_URL/api/v1/analytics/reports/powerbi/config" -H "$AUTH_H"
assert_status "[8-1] GET /analytics/reports/powerbi/config" "200" "$HTTP_CODE"
assert_field  "[8-2] Has 'configured' field" "configured" "$HTTP_BODY"

if echo "$HTTP_BODY" | grep -q "client_secret\|access_token\|password"; then
  fail "[8-3] Power BI config must NOT expose secrets"
else
  pass "[8-3] Power BI config does not expose secrets"
fi

# ---------------------------------------------------------------------------
# 9. Export Requests
# ---------------------------------------------------------------------------
echo ""
echo "--- [9] Export Requests ---"

curl_req POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "$AUTH_H" -H "Content-Type: application/json" \
  -d '{"dataset_key":"executive-summary","format":"XLSX","parameters":{}}'
assert_status "[9-1] POST /export/request (XLSX)" "201" "$HTTP_CODE"
assert_field  "[9-2] Has 'data' in response" "data"    "$HTTP_BODY"
# UAT-002: export now processes synchronously — status may be COMPLETED immediately
STATUS_9=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['status'])" 2>/dev/null || true)
if [[ "$STATUS_9" == "PENDING" || "$STATUS_9" == "PROCESSING" || "$STATUS_9" == "COMPLETED" ]]; then
  pass "[9-3] Has valid export status ($STATUS_9)"
else
  fail "[9-3] Unexpected export status: $STATUS_9"
fi
EXP_ID=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['id'])" 2>/dev/null || true)

curl_req POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "$AUTH_H" -H "Content-Type: application/json" \
  -d '{"dataset_key":"project-performance","format":"CSV","parameters":{"status":"ACTIVE"}}'
assert_status "[9-4] POST /export/request (CSV)" "201" "$HTTP_CODE"

curl_req POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "$AUTH_H" -H "Content-Type: application/json" \
  -d '{"dataset_key":"budget","format":"PDF","parameters":{}}'
assert_status "[9-5] POST /export/request (PDF)" "201" "$HTTP_CODE"

curl_req GET "$BASE_URL/api/v1/analytics/reports/export/requests" -H "$AUTH_H"
assert_status "[9-6] GET /export/requests" "200" "$HTTP_CODE"
assert_field  "[9-7] List has 'data'" "data" "$HTTP_BODY"

if [[ -n "$EXP_ID" ]]; then
  curl_req GET "$BASE_URL/api/v1/analytics/reports/export/requests/$EXP_ID" -H "$AUTH_H"
  assert_status "[9-8] GET /export/requests/:id" "200" "$HTTP_CODE"
  assert_field  "[9-9] Export request has dataset_key" "dataset_key" "$HTTP_BODY"
else
  skip "[9-8] GET /export/requests/:id (no ID from create)"
  skip "[9-9] Export request has dataset_key"
fi

curl_req POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "$AUTH_H" -H "Content-Type: application/json" \
  -d '{"dataset_key":"budget","format":"DOCX"}'
assert_status "[9-10] POST /export/request with invalid format → 400" "400" "$HTTP_CODE"

# ---------------------------------------------------------------------------
# 10. Auth guard
# ---------------------------------------------------------------------------
echo ""
echo "--- [10] Auth Guard ---"

curl_req GET "$BASE_URL/api/v1/analytics/reports/datasets/executive-summary"
assert_status "[10-1] No-auth request → 401" "401" "$HTTP_CODE"

curl_req GET "$BASE_URL/api/v1/analytics/reports/catalog"
assert_status "[10-2] No-auth catalog → 401" "401" "$HTTP_CODE"

curl_req GET "$BASE_URL/api/v1/analytics/reports/powerbi/config"
assert_status "[10-3] No-auth powerbi config → 401" "401" "$HTTP_CODE"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "================================================="
TOTAL=$((PASS + FAIL + SKIP))
echo "  Total: $TOTAL  |  PASS: $PASS  |  FAIL: $FAIL  |  SKIP: $SKIP"
echo "================================================="

if [[ $FAIL -gt 0 ]]; then
  exit 1
fi
exit 0

# ---------------------------------------------------------------------------
# 2. Dataset: Executive Summary
# ---------------------------------------------------------------------------
echo ""
echo "--- [2] Dataset: Executive Summary ---"

EXEC_RESP=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/analytics/reports/datasets/executive-summary" \
  -H "$AUTH")
EXEC_BODY=$(echo "$EXEC_RESP" | sed '$d')
EXEC_CODE=$(echo "$EXEC_RESP" | tail -n 1)
assert_status "[2-1] GET /datasets/executive-summary" "200" "$EXEC_CODE"
assert_field "[2-2] Has total_projects"    "total_projects"    "$EXEC_BODY"
assert_field "[2-3] Has total_budget_plan" "total_budget_plan" "$EXEC_BODY"
assert_field "[2-4] Has green_health"      "green_health"      "$EXEC_BODY"
assert_field "[2-5] Has open_risks"        "open_risks"        "$EXEC_BODY"

# With status filter
EXEC_FILTER_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  "$BASE_URL/api/v1/analytics/reports/datasets/executive-summary?status=ACTIVE" \
  -H "$AUTH")
assert_status "[2-6] GET /datasets/executive-summary?status=ACTIVE" "200" "$EXEC_FILTER_CODE"

# ---------------------------------------------------------------------------
# 3. Dataset: Project Performance
# ---------------------------------------------------------------------------
echo ""
echo "--- [3] Dataset: Project Performance ---"

PERF_RESP=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/analytics/reports/datasets/project-performance" \
  -H "$AUTH")
PERF_BODY=$(echo "$PERF_RESP" | sed '$d')
PERF_CODE=$(echo "$PERF_RESP" | tail -n 1)
assert_status "[3-1] GET /datasets/project-performance" "200" "$PERF_CODE"
assert_field "[3-2] Has data" "data" "$PERF_BODY"

# With province filter
PERF_PROV_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  "$BASE_URL/api/v1/analytics/reports/datasets/project-performance?province=Jawa+Barat" \
  -H "$AUTH")
assert_status "[3-3] GET /datasets/project-performance?province=Jawa+Barat" "200" "$PERF_PROV_CODE"

# ---------------------------------------------------------------------------
# 4. Dataset: Risk & Issue
# ---------------------------------------------------------------------------
echo ""
echo "--- [4] Dataset: Risk & Issue ---"

RISK_RESP=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/analytics/reports/datasets/risk-issue" \
  -H "$AUTH")
RISK_BODY=$(echo "$RISK_RESP" | sed '$d')
RISK_CODE=$(echo "$RISK_RESP" | tail -n 1)
assert_status "[4-1] GET /datasets/risk-issue" "200" "$RISK_CODE"
assert_field "[4-2] Has data" "data" "$RISK_BODY"

# ---------------------------------------------------------------------------
# 5. Dataset: Budget
# ---------------------------------------------------------------------------
echo ""
echo "--- [5] Dataset: Budget ---"

BUDGET_RESP=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/analytics/reports/datasets/budget" \
  -H "$AUTH")
BUDGET_BODY=$(echo "$BUDGET_RESP" | sed '$d')
BUDGET_CODE=$(echo "$BUDGET_RESP" | tail -n 1)
assert_status "[5-1] GET /datasets/budget" "200" "$BUDGET_CODE"
assert_field "[5-2] Has data" "data" "$BUDGET_BODY"

# With status filter
BUDGET_FILTER_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  "$BASE_URL/api/v1/analytics/reports/datasets/budget?status=ACTIVE" \
  -H "$AUTH")
assert_status "[5-3] GET /datasets/budget?status=ACTIVE" "200" "$BUDGET_FILTER_CODE"

# ---------------------------------------------------------------------------
# 6. Dataset: Benefits
# ---------------------------------------------------------------------------
echo ""
echo "--- [6] Dataset: Benefits ---"

BEN_RESP=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/analytics/reports/datasets/benefits" \
  -H "$AUTH")
BEN_BODY=$(echo "$BEN_RESP" | sed '$d')
BEN_CODE=$(echo "$BEN_RESP" | tail -n 1)
assert_status "[6-1] GET /datasets/benefits" "200" "$BEN_CODE"
assert_field "[6-2] Has data" "data" "$BEN_BODY"

# ---------------------------------------------------------------------------
# 7. Dataset: Priority
# ---------------------------------------------------------------------------
echo ""
echo "--- [7] Dataset: Priority ---"

PRIO_RESP=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/analytics/reports/datasets/priority" \
  -H "$AUTH")
PRIO_BODY=$(echo "$PRIO_RESP" | sed '$d')
PRIO_CODE=$(echo "$PRIO_RESP" | tail -n 1)
assert_status "[7-1] GET /datasets/priority" "200" "$PRIO_CODE"
assert_field "[7-2] Has data" "data" "$PRIO_BODY"

# ---------------------------------------------------------------------------
# 8. Power BI Config
# ---------------------------------------------------------------------------
echo ""
echo "--- [8] Power BI Config ---"

POWERBI_RESP=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/analytics/reports/powerbi/config" \
  -H "$AUTH")
POWERBI_BODY=$(echo "$POWERBI_RESP" | sed '$d')
POWERBI_CODE=$(echo "$POWERBI_RESP" | tail -n 1)
assert_status "[8-1] GET /analytics/reports/powerbi/config" "200" "$POWERBI_CODE"
assert_field "[8-2] Has 'configured' field"  "configured"  "$POWERBI_BODY"

# Should not expose secrets (client_secret, client_id raw token)
if echo "$POWERBI_BODY" | grep -q "client_secret\|access_token\|password"; then
  fail "[8-3] Power BI config must NOT expose secrets"
else
  pass "[8-3] Power BI config does not expose secrets"
fi

# ---------------------------------------------------------------------------
# 9. Export Requests
# ---------------------------------------------------------------------------
echo ""
echo "--- [9] Export Requests ---"

# Create export request
EXP_RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"dataset_key":"executive-summary","format":"XLSX","parameters":{}}')
EXP_BODY=$(echo "$EXP_RESP" | sed '$d')
EXP_CODE=$(echo "$EXP_RESP" | tail -n 1)
assert_status "[9-1] POST /export/request (XLSX)" "201" "$EXP_CODE"
assert_field "[9-2] Has 'data' in response"   "data"   "$EXP_BODY"
assert_field "[9-3] Has status PENDING"        "PENDING" "$EXP_BODY"

EXP_ID=$(echo "$EXP_BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

# Create CSV export
EXP_CSV_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"dataset_key":"project-performance","format":"CSV","parameters":{"status":"ACTIVE"}}')
assert_status "[9-4] POST /export/request (CSV)" "201" "$EXP_CSV_CODE"

# Create PDF export
EXP_PDF_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"dataset_key":"budget","format":"PDF","parameters":{}}')
assert_status "[9-5] POST /export/request (PDF)" "201" "$EXP_PDF_CODE"

# List export requests
LIST_RESP=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/analytics/reports/export/requests" \
  -H "$AUTH")
LIST_BODY=$(echo "$LIST_RESP" | sed '$d')
LIST_CODE=$(echo "$LIST_RESP" | tail -n 1)
assert_status "[9-6] GET /export/requests" "200" "$LIST_CODE"
assert_field "[9-7] List has 'data'" "data" "$LIST_BODY"

# Get specific export request
if [[ -n "$EXP_ID" ]]; then
  GET_EXP_RESP=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/analytics/reports/export/requests/$EXP_ID" \
    -H "$AUTH")
  GET_EXP_BODY=$(echo "$GET_EXP_RESP" | sed '$d')
  GET_EXP_CODE=$(echo "$GET_EXP_RESP" | tail -n 1)
  assert_status "[9-8] GET /export/requests/:id" "200" "$GET_EXP_CODE"
  assert_field "[9-9] Export request has dataset_key" "dataset_key" "$GET_EXP_BODY"
else
  skip "[9-8] GET /export/requests/:id (no ID from create)"
  skip "[9-9] Export request has dataset_key"
fi

# Invalid format should fail
BAD_FORMAT_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"dataset_key":"budget","format":"DOCX"}')
assert_status "[9-10] POST /export/request with invalid format → 400" "400" "$BAD_FORMAT_CODE"

# ---------------------------------------------------------------------------
# 10. Auth guard
# ---------------------------------------------------------------------------
echo ""
echo "--- [10] Auth Guard ---"

NO_AUTH_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  "$BASE_URL/api/v1/analytics/reports/datasets/executive-summary")
assert_status "[10-1] No-auth request → 401" "401" "$NO_AUTH_CODE"

NO_AUTH_CATALOG=$(curl -s -o /dev/null -w "%{http_code}" \
  "$BASE_URL/api/v1/analytics/reports/catalog")
assert_status "[10-2] No-auth catalog → 401" "401" "$NO_AUTH_CATALOG"

NO_AUTH_POWERBI=$(curl -s -o /dev/null -w "%{http_code}" \
  "$BASE_URL/api/v1/analytics/reports/powerbi/config")
assert_status "[10-3] No-auth powerbi config → 401" "401" "$NO_AUTH_POWERBI"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "================================================="
TOTAL=$((PASS + FAIL + SKIP))
echo "  Total: $TOTAL  |  PASS: $PASS  |  FAIL: $FAIL  |  SKIP: $SKIP"
echo "================================================="

if [[ $FAIL -gt 0 ]]; then
  exit 1
fi
exit 0
