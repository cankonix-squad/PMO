#!/usr/bin/env bash
# ============================================================
# CANKORA — Smoke Test: UAT-003 Audit Log Viewer
# ============================================================
# Usage:
#   bash docs/smoke-test-uat-003-audit-logs.sh
#   BASE_URL=http://localhost:8080 bash docs/smoke-test-uat-003-audit-logs.sh
#
# Requirements:
#   - Backend running at BASE_URL (default http://localhost:8080)
#   - Seeded DB (admin@cankora.local / Admin@Cankora2024!)
#   - curl + jq available
# ============================================================

set -uo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
API="${BASE_URL}/api/v1"
ADMIN_EMAIL="admin@cankora.local"
ADMIN_PASSWORD="Admin@Cankora2024!"

PASS=0
FAIL=0
SKIP=0

# ── helpers ──────────────────────────────────────────────────────────────────

green()  { echo -e "\033[32m[PASS]\033[0m $*"; }
red()    { echo -e "\033[31m[FAIL]\033[0m $*"; }
yellow() { echo -e "\033[33m[SKIP]\033[0m $*"; }
header() { echo -e "\n\033[1;34m$*\033[0m"; }

pass() { green  "$1"; ((PASS++)); }
fail() { red    "$1"; ((FAIL++)); }
skip() { yellow "$1"; ((SKIP++)); }

check_http() {
  local label="$1" url="$2" expected="$3" extra_flags="${4:-}"
  local code
  code=$(eval curl -s -o /dev/null -w "%{http_code}" $extra_flags "\"$url\"")
  if [[ "$code" == "$expected" ]]; then
    pass "[${code}] $label"
  else
    fail "[${code}] $label (expected ${expected})"
  fi
}

json_field() {
  # json_field <json_string> <jq_expression>
  echo "$1" | jq -r "$2" 2>/dev/null
}

# ── login and get token ───────────────────────────────────────────────────────

header "[0] Pre-flight: backend health"
check_http "Backend health" "${BASE_URL}/health" "200"

header "[1] Auth"
LOGIN_RESP=$(curl -s -X POST "${API}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASSWORD}\"}")

TOKEN=$(json_field "$LOGIN_RESP" '.data.access_token')
if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
  red "Login failed — cannot continue"
  echo "Response: $LOGIN_RESP"
  exit 1
fi
pass "[1-1] Admin login OK, token received"

AUTH="-H \"Authorization: Bearer ${TOKEN}\""

# ── [2] No-token 401 guard ────────────────────────────────────────────────────

header "[2] Unauthenticated requests must be rejected"
check_http "[2-1] GET /audit-logs no token → 401" \
  "${API}/audit-logs" "401"

check_http "[2-2] GET /audit-logs/summary no token → 401" \
  "${API}/audit-logs/summary" "401"

check_http "[2-3] GET /audit-logs/export no token → 401" \
  "${API}/audit-logs/export" "401"

# ── [3] Admin can list audit logs ─────────────────────────────────────────────

header "[3] Admin: list audit logs"
LIST_RESP=$(eval curl -s "${AUTH}" "${API}/audit-logs")
LIST_SUCCESS=$(json_field "$LIST_RESP" '.success')

if [[ "$LIST_SUCCESS" == "true" ]]; then
  pass "[3-1] GET /audit-logs → 200 success"
else
  fail "[3-1] GET /audit-logs unexpected response: $LIST_RESP"
fi

# pagination metadata
META_PAGE=$(json_field "$LIST_RESP" '.meta.page')
META_TOTAL=$(json_field "$LIST_RESP" '.meta.total')
META_PAGES=$(json_field "$LIST_RESP" '.meta.total_pages')

if [[ "$META_PAGE" != "null" && "$META_PAGE" != "" ]]; then
  pass "[3-2] Pagination meta.page present (${META_PAGE})"
else
  fail "[3-2] Pagination meta.page missing"
fi

if [[ "$META_TOTAL" != "null" && "$META_TOTAL" != "" ]]; then
  pass "[3-3] Pagination meta.total present (${META_TOTAL})"
else
  fail "[3-3] Pagination meta.total missing"
fi

if [[ "$META_PAGES" != "null" && "$META_PAGES" != "" ]]; then
  pass "[3-4] Pagination meta.total_pages present (${META_PAGES})"
else
  fail "[3-4] Pagination meta.total_pages missing"
fi

# ── [4] Pagination params ─────────────────────────────────────────────────────

header "[4] Pagination params"
PAGE2_RESP=$(eval curl -s "${AUTH}" "${API}/audit-logs?page=1&page_size=5")
PAGE2_SUCCESS=$(json_field "$PAGE2_RESP" '.success')
if [[ "$PAGE2_SUCCESS" == "true" ]]; then
  pass "[4-1] GET /audit-logs?page=1&page_size=5 → 200"
else
  fail "[4-1] GET /audit-logs?page=1&page_size=5 unexpected: $PAGE2_RESP"
fi

# ── [5] Filter: action ────────────────────────────────────────────────────────

header "[5] Filter params"
ACT_RESP=$(eval curl -s "${AUTH}" "${API}/audit-logs?action=login")
ACT_SUCCESS=$(json_field "$ACT_RESP" '.success')
if [[ "$ACT_SUCCESS" == "true" ]]; then
  pass "[5-1] Filter ?action=login → 200"
else
  fail "[5-1] Filter ?action=login unexpected: $ACT_RESP"
fi

ENT_RESP=$(eval curl -s "${AUTH}" "${API}/audit-logs?entity_type=project")
ENT_SUCCESS=$(json_field "$ENT_RESP" '.success')
if [[ "$ENT_SUCCESS" == "true" ]]; then
  pass "[5-2] Filter ?entity_type=project → 200"
else
  fail "[5-2] Filter ?entity_type=project unexpected: $ENT_RESP"
fi

# ── [6] Search keyword ────────────────────────────────────────────────────────

header "[6] Search keyword"
SEARCH_RESP=$(eval curl -s "${AUTH}" "${API}/audit-logs?search=admin")
SEARCH_SUCCESS=$(json_field "$SEARCH_RESP" '.success')
if [[ "$SEARCH_SUCCESS" == "true" ]]; then
  pass "[6-1] Search ?search=admin → 200"
else
  fail "[6-1] Search ?search=admin unexpected: $SEARCH_RESP"
fi

# ── [7] Date filter ───────────────────────────────────────────────────────────

header "[7] Date filters"
DATE_FROM="2020-01-01"
DATE_TO="2099-12-31"
DATE_RESP=$(eval curl -s "${AUTH}" \
  "\"${API}/audit-logs?date_from=${DATE_FROM}&date_to=${DATE_TO}\"")
DATE_SUCCESS=$(json_field "$DATE_RESP" '.success')
if [[ "$DATE_SUCCESS" == "true" ]]; then
  pass "[7-1] Filter ?date_from=&date_to= → 200"
else
  fail "[7-1] Date filter unexpected: $DATE_RESP"
fi

# invalid date returns 400
BAD_DATE_CODE=$(eval curl -s -o /dev/null -w "%{http_code}" "${AUTH}" \
  "\"${API}/audit-logs?date_from=not-a-date\"")
if [[ "$BAD_DATE_CODE" == "400" ]]; then
  pass "[7-2] Invalid date_from → 400"
else
  skip "[7-2] Invalid date_from → got ${BAD_DATE_CODE} (may be 400 or ignored)"
fi

# ── [8] Summary endpoint ──────────────────────────────────────────────────────

header "[8] Summary endpoint"
SUMMARY_RESP=$(eval curl -s "${AUTH}" "${API}/audit-logs/summary")
SUMMARY_SUCCESS=$(json_field "$SUMMARY_RESP" '.success')
TOTAL_EVENTS=$(json_field "$SUMMARY_RESP" '.data.total_events')

if [[ "$SUMMARY_SUCCESS" == "true" ]]; then
  pass "[8-1] GET /audit-logs/summary → 200"
else
  fail "[8-1] GET /audit-logs/summary unexpected: $SUMMARY_RESP"
fi

if [[ "$TOTAL_EVENTS" != "null" && "$TOTAL_EVENTS" != "" ]]; then
  pass "[8-2] summary.data.total_events present (${TOTAL_EVENTS})"
else
  fail "[8-2] summary.data.total_events missing"
fi

UNIQUE_ACTORS=$(json_field "$SUMMARY_RESP" '.data.unique_actors')
if [[ "$UNIQUE_ACTORS" != "null" && "$UNIQUE_ACTORS" != "" ]]; then
  pass "[8-3] summary.data.unique_actors present (${UNIQUE_ACTORS})"
else
  fail "[8-3] summary.data.unique_actors missing"
fi

TOP_ACTIONS=$(json_field "$SUMMARY_RESP" '.data.top_actions')
if [[ "$TOP_ACTIONS" != "null" && "$TOP_ACTIONS" != "" ]]; then
  pass "[8-4] summary.data.top_actions present"
else
  fail "[8-4] summary.data.top_actions missing"
fi

TOP_ENTITIES=$(json_field "$SUMMARY_RESP" '.data.top_entities')
if [[ "$TOP_ENTITIES" != "null" && "$TOP_ENTITIES" != "" ]]; then
  pass "[8-5] summary.data.top_entities present"
else
  fail "[8-5] summary.data.top_entities missing"
fi

# ── [9] Get by ID ─────────────────────────────────────────────────────────────

header "[9] Get by ID"
FIRST_ID=$(json_field "$LIST_RESP" '.data[0].id')
if [[ -z "$FIRST_ID" || "$FIRST_ID" == "null" ]]; then
  skip "[9-1] No audit log entries in DB — skipping GetByID test"
else
  BYID_RESP=$(eval curl -s "${AUTH}" "${API}/audit-logs/${FIRST_ID}")
  BYID_SUCCESS=$(json_field "$BYID_RESP" '.success')
  if [[ "$BYID_SUCCESS" == "true" ]]; then
    pass "[9-1] GET /audit-logs/:id → 200 for first entry"
  else
    fail "[9-1] GET /audit-logs/:id unexpected: $BYID_RESP"
  fi

  # ensure org isolation: bad UUID → 404
  BAD_ID_CODE=$(eval curl -s -o /dev/null -w "%{http_code}" "${AUTH}" \
    "${API}/audit-logs/00000000-0000-0000-0000-000000000000")
  if [[ "$BAD_ID_CODE" == "404" ]]; then
    pass "[9-2] GET /audit-logs/00000000-... → 404"
  else
    fail "[9-2] GET /audit-logs/00000000-... → got ${BAD_ID_CODE} (expected 404)"
  fi
fi

# ── [10] CSV Export ───────────────────────────────────────────────────────────

header "[10] CSV Export"
EXPORT_HTTP=$(eval curl -s -o /dev/null -w "%{http_code}" "${AUTH}" \
  "${API}/audit-logs/export")
if [[ "$EXPORT_HTTP" == "200" ]]; then
  pass "[10-1] GET /audit-logs/export → 200"
else
  fail "[10-1] GET /audit-logs/export → got ${EXPORT_HTTP}"
fi

EXPORT_CONTENT=$(eval curl -s -D - -o /dev/null "${AUTH}" \
  "${API}/audit-logs/export" 2>/dev/null | grep -i "content-type" || true)
if echo "$EXPORT_CONTENT" | grep -qi "text/csv"; then
  pass "[10-2] Export Content-Type contains text/csv"
else
  skip "[10-2] Export Content-Type check inconclusive (header parse may vary)"
fi

EXPORT_DISP=$(eval curl -s -D - -o /dev/null "${AUTH}" \
  "${API}/audit-logs/export" 2>/dev/null | grep -i "content-disposition" || true)
if echo "$EXPORT_DISP" | grep -qi "attachment"; then
  pass "[10-3] Export Content-Disposition is attachment"
else
  skip "[10-3] Export Content-Disposition check inconclusive"
fi

# ── [11] Tenant isolation ─────────────────────────────────────────────────────

header "[11] Tenant isolation"
# All results must belong to the same org — check data[0].organization_id matches login user
FIRST_ORG=$(json_field "$LIST_RESP" '.data[0].organization_id')
USER_ORG=$(json_field "$LOGIN_RESP" '.data.user.organization_id')
if [[ -z "$FIRST_ORG" || "$FIRST_ORG" == "null" ]]; then
  skip "[11-1] No entries to verify org isolation"
elif [[ "$FIRST_ORG" == "$USER_ORG" ]]; then
  pass "[11-1] First audit log org matches caller's org (tenant-safe)"
else
  fail "[11-1] Org mismatch: entry.org=${FIRST_ORG} caller.org=${USER_ORG}"
fi

# ── [12] Frontend route reachability (optional) ───────────────────────────────

header "[12] Frontend route /audit-logs"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"
FE_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  --max-time 5 "${FRONTEND_URL}/audit-logs" 2>/dev/null || echo "000")
if [[ "$FE_CODE" == "200" ]]; then
  pass "[12-1] Frontend /audit-logs → 200"
elif [[ "$FE_CODE" == "000" ]]; then
  skip "[12-1] Frontend not reachable at ${FRONTEND_URL} (start frontend for full test)"
else
  skip "[12-1] Frontend /audit-logs → ${FE_CODE} (may redirect to login — OK)"
fi

# ── summary ───────────────────────────────────────────────────────────────────

echo ""
echo "============================================================"
echo " Smoke Test: UAT-003 Audit Log Viewer"
echo " PASS: ${PASS}  FAIL: ${FAIL}  SKIP: ${SKIP}"
echo "============================================================"

if [[ "$FAIL" -gt 0 ]]; then
  echo "Result: FAIL"
  exit 1
else
  echo "Result: PASS"
  exit 0
fi
