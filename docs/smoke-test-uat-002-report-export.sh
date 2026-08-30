#!/usr/bin/env bash
# smoke-test-uat-002-report-export.sh
# UAT-002 Report Export Real File — smoke tests
# Usage: BASE_URL=http://localhost:8080 bash smoke-test-uat-002-report-export.sh
#
# Requirements: curl, python3, jq (optional but helps)
# Idempotent: safe to re-run multiple times

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
TMPFILE=$(mktemp)
trap 'rm -f "$TMPBODY" "$TMPFILE"' EXIT

pass() { echo -e "${GREEN}PASS${NC} $1"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}FAIL${NC} $1"; FAIL=$((FAIL+1)); }
skip() { echo -e "${YELLOW}SKIP${NC} $1"; SKIP=$((SKIP+1)); }

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

assert_value() {
  local label="$1" expected="$2" body="$3"
  if echo "$body" | grep -q "$expected"; then
    pass "$label (value '$expected' found)"
  else
    fail "$label — value '$expected' not found in: $body"
  fi
}

echo ""
echo "================================================="
echo "  UAT-002 Report Export Real File — Smoke Test"
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
assert_status "[0-1] Admin login" "200" "$HTTP_CODE"

TOKEN=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['access_token'])" 2>/dev/null || true)
if [[ -z "$TOKEN" ]]; then
  echo -e "${RED}FATAL: Could not extract auth token — aborting${NC}"
  exit 1
fi
pass "[0-2] Token extracted"

AUTH_H="Authorization: Bearer $TOKEN"

# ---------------------------------------------------------------------------
# 1. Unauthorized access guard
# ---------------------------------------------------------------------------
echo ""
echo "--- [1] Unauthorized Guard ---"
curl_req GET "$BASE_URL/api/v1/analytics/reports/export/requests"
assert_status "[1-1] List exports no auth → 401" "401" "$HTTP_CODE"

curl_req POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "Content-Type: application/json" \
  -d '{"dataset_key":"budget","format":"CSV"}'
assert_status "[1-2] Create export no auth → 401" "401" "$HTTP_CODE"

# ---------------------------------------------------------------------------
# 2. Create CSV export — executive-summary
# ---------------------------------------------------------------------------
echo ""
echo "--- [2] Create CSV Export (executive-summary) ---"
curl_req POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "Content-Type: application/json" \
  -H "$AUTH_H" \
  -d '{"dataset_key":"executive-summary","format":"CSV"}'
assert_status "[2-1] Create export request → 201" "201" "$HTTP_CODE"

assert_field "[2-2] response has id" "id" "$HTTP_BODY"
assert_field "[2-3] response has status" "status" "$HTTP_BODY"
assert_field "[2-4] response has dataset_key" "dataset_key" "$HTTP_BODY"
assert_field "[2-5] response has format" "format" "$HTTP_BODY"

# Extract request ID
EXPORT_ID=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['id'])" 2>/dev/null || true)
if [[ -z "$EXPORT_ID" ]]; then
  fail "[2-6] Could not extract export request ID"
else
  pass "[2-6] Export request ID extracted: $EXPORT_ID"
fi

# ---------------------------------------------------------------------------
# 3. Verify export COMPLETED with file metadata
# ---------------------------------------------------------------------------
echo ""
echo "--- [3] Verify Export Completed with File Metadata ---"
curl_req GET "$BASE_URL/api/v1/analytics/reports/export/requests/$EXPORT_ID" \
  -H "$AUTH_H"
assert_status "[3-1] Get export request → 200" "200" "$HTTP_CODE"

# Check status is COMPLETED (sync processing)
STATUS=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['status'])" 2>/dev/null || true)
if [[ "$STATUS" == "COMPLETED" ]]; then
  pass "[3-2] Export status = COMPLETED"
elif [[ "$STATUS" == "PENDING" || "$STATUS" == "PROCESSING" ]]; then
  # Allow a brief wait in case of async
  sleep 2
  curl_req GET "$BASE_URL/api/v1/analytics/reports/export/requests/$EXPORT_ID" \
    -H "$AUTH_H"
  STATUS=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['status'])" 2>/dev/null || true)
  if [[ "$STATUS" == "COMPLETED" ]]; then
    pass "[3-2] Export status = COMPLETED (after wait)"
  else
    fail "[3-2] Export status = $STATUS (expected COMPLETED)"
  fi
else
  fail "[3-2] Export status = $STATUS (expected COMPLETED)"
fi

assert_field "[3-3] response has file_name" "file_name" "$HTTP_BODY"
assert_field "[3-4] response has mime_type" "mime_type" "$HTTP_BODY"
assert_field "[3-5] response has file_size" "file_size" "$HTTP_BODY"
assert_field "[3-6] response has generated_at" "generated_at" "$HTTP_BODY"

FILE_NAME=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'].get('file_name',''))" 2>/dev/null || true)
FILE_SIZE=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'].get('file_size',0))" 2>/dev/null || true)
MIME_TYPE=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'].get('mime_type',''))" 2>/dev/null || true)

if [[ "$FILE_NAME" == *.csv ]]; then
  pass "[3-7] file_name ends with .csv ($FILE_NAME)"
else
  fail "[3-7] file_name does not end with .csv: $FILE_NAME"
fi

if [[ -n "$FILE_SIZE" && "$FILE_SIZE" -gt 0 ]] 2>/dev/null; then
  pass "[3-8] file_size > 0 ($FILE_SIZE bytes)"
else
  fail "[3-8] file_size is 0 or missing ($FILE_SIZE)"
fi

if echo "$MIME_TYPE" | grep -qi "text/csv"; then
  pass "[3-9] mime_type is text/csv ($MIME_TYPE)"
else
  fail "[3-9] mime_type not text/csv: $MIME_TYPE"
fi

# ---------------------------------------------------------------------------
# 4. Download file — 200, non-empty, CSV headers
# ---------------------------------------------------------------------------
echo ""
echo "--- [4] Download CSV File ---"
HTTP_CODE=$(curl -s -o "$TMPFILE" -w "%{http_code}" \
  -H "$AUTH_H" \
  "$BASE_URL/api/v1/analytics/reports/export/requests/$EXPORT_ID/download")

if [[ "$HTTP_CODE" == "200" ]]; then
  pass "[4-1] Download → 200"
else
  fail "[4-1] Download → expected 200, got $HTTP_CODE"
fi

# File must not be empty
FILE_BYTES=$(wc -c < "$TMPFILE" 2>/dev/null || echo 0)
if [[ "$FILE_BYTES" -gt 0 ]]; then
  pass "[4-2] Downloaded file is not empty ($FILE_BYTES bytes)"
else
  fail "[4-2] Downloaded file is empty"
fi

# CSV first line must be a header with known columns
FIRST_LINE=$(head -1 "$TMPFILE" 2>/dev/null || true)
if echo "$FIRST_LINE" | grep -q "metric"; then
  pass "[4-3] CSV header contains 'metric'"
elif echo "$FIRST_LINE" | grep -q "project_id\|total_projects\|dataset\|value"; then
  pass "[4-3] CSV header looks valid: $FIRST_LINE"
else
  fail "[4-3] CSV header unexpected: $FIRST_LINE"
fi

# File must have at least 2 lines (header + data)
LINE_COUNT=$(wc -l < "$TMPFILE" 2>/dev/null || echo 0)
if [[ "$LINE_COUNT" -ge 2 ]]; then
  pass "[4-4] CSV has $LINE_COUNT lines (≥2)"
else
  fail "[4-4] CSV has only $LINE_COUNT lines (expected ≥2)"
fi

# ---------------------------------------------------------------------------
# 5. Unauthorized download guard
# ---------------------------------------------------------------------------
echo ""
echo "--- [5] Unauthorized Download Guard ---"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  "$BASE_URL/api/v1/analytics/reports/export/requests/$EXPORT_ID/download")
if [[ "$HTTP_CODE" == "401" ]]; then
  pass "[5-1] Download without auth → 401"
else
  fail "[5-1] Download without auth → expected 401, got $HTTP_CODE"
fi

# ---------------------------------------------------------------------------
# 6. Invalid / non-existent request ID
# ---------------------------------------------------------------------------
echo ""
echo "--- [6] Invalid Request ID ---"
FAKE_ID="00000000-0000-0000-0000-000000000000"
curl_req GET "$BASE_URL/api/v1/analytics/reports/export/requests/$FAKE_ID" \
  -H "$AUTH_H"
assert_status "[6-1] Get non-existent export → 404" "404" "$HTTP_CODE"

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "$AUTH_H" \
  "$BASE_URL/api/v1/analytics/reports/export/requests/$FAKE_ID/download")
if [[ "$HTTP_CODE" == "404" ]]; then
  pass "[6-2] Download non-existent export → 404"
else
  fail "[6-2] Download non-existent export → expected 404, got $HTTP_CODE"
fi

# Invalid UUID format
curl_req GET "$BASE_URL/api/v1/analytics/reports/export/requests/not-a-uuid" \
  -H "$AUTH_H"
assert_status "[6-3] Get invalid UUID → 400" "400" "$HTTP_CODE"

# ---------------------------------------------------------------------------
# 7. Create XLSX export — project-performance
# ---------------------------------------------------------------------------
echo ""
echo "--- [7] Create XLSX Export (project-performance) ---"
curl_req POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "Content-Type: application/json" \
  -H "$AUTH_H" \
  -d '{"dataset_key":"project-performance","format":"XLSX"}'
assert_status "[7-1] Create XLSX export → 201" "201" "$HTTP_CODE"

XLSX_ID=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['id'])" 2>/dev/null || true)
XLSX_STATUS=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['status'])" 2>/dev/null || true)

if [[ "$XLSX_STATUS" == "COMPLETED" ]]; then
  pass "[7-2] XLSX export status = COMPLETED"
elif [[ "$XLSX_STATUS" == "FAILED" ]]; then
  fail "[7-2] XLSX export status = FAILED"
else
  skip "[7-2] XLSX export status = $XLSX_STATUS (async)"
fi

if [[ -n "$XLSX_ID" && "$XLSX_STATUS" == "COMPLETED" ]]; then
  HTTP_CODE=$(curl -s -o "$TMPFILE" -w "%{http_code}" \
    -H "$AUTH_H" \
    "$BASE_URL/api/v1/analytics/reports/export/requests/$XLSX_ID/download")
  if [[ "$HTTP_CODE" == "200" ]]; then
    pass "[7-3] Download XLSX → 200"
    XLSX_BYTES=$(wc -c < "$TMPFILE" 2>/dev/null || echo 0)
    if [[ "$XLSX_BYTES" -gt 0 ]]; then
      pass "[7-4] XLSX file not empty ($XLSX_BYTES bytes)"
    else
      fail "[7-4] XLSX file empty"
    fi
  else
    fail "[7-3] Download XLSX → expected 200, got $HTTP_CODE"
  fi
else
  skip "[7-3] XLSX download skipped (not COMPLETED)"
  skip "[7-4] XLSX size check skipped"
fi

# ---------------------------------------------------------------------------
# 8. All 6 datasets — CSV export smoke
# ---------------------------------------------------------------------------
echo ""
echo "--- [8] All 6 Datasets CSV Export ---"
DATASETS=("executive-summary" "project-performance" "risk-issue" "budget" "benefits" "priority")
for ds in "${DATASETS[@]}"; do
  curl_req POST "$BASE_URL/api/v1/analytics/reports/export/request" \
    -H "Content-Type: application/json" \
    -H "$AUTH_H" \
    -d "{\"dataset_key\":\"$ds\",\"format\":\"CSV\"}"
  if [[ "$HTTP_CODE" == "201" ]]; then
    DS_STATUS=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['status'])" 2>/dev/null || true)
    if [[ "$DS_STATUS" == "COMPLETED" ]]; then
      pass "[8] dataset '$ds' export → COMPLETED"
    else
      fail "[8] dataset '$ds' export → status=$DS_STATUS"
    fi
  else
    fail "[8] dataset '$ds' create export → HTTP $HTTP_CODE"
  fi
done

# ---------------------------------------------------------------------------
# 9. List export requests — pagination & format
# ---------------------------------------------------------------------------
echo ""
echo "--- [9] List Export Requests ---"
curl_req GET "$BASE_URL/api/v1/analytics/reports/export/requests" \
  -H "$AUTH_H"
assert_status "[9-1] List export requests → 200" "200" "$HTTP_CODE"

LIST_COUNT=$(echo "$HTTP_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d.get('data',d.get('data',[]))))" 2>/dev/null || true)
if [[ -n "$LIST_COUNT" && "$LIST_COUNT" -gt 0 ]]; then
  pass "[9-2] List returns $LIST_COUNT export requests"
else
  fail "[9-2] List returned 0 or invalid count: $LIST_COUNT"
fi

# ---------------------------------------------------------------------------
# 10. Tenant guard — second org cannot download first org's export
# ---------------------------------------------------------------------------
echo ""
echo "--- [10] Tenant Guard (cross-org download) ---"
# We test this by using a fake org UUID in the request — the org guard is
# enforced by the service using the caller's claims.OrganizationID.
# A simpler proxy: use a FAKE token with wrong claims can't be created here,
# but we can verify the endpoint returns 401 for a no-auth request (already
# tested in section 5), and 404 for a non-existent ID in section 6.
# Full tenant isolation test requires a second org + second user — skip with note.
skip "[10-1] Full cross-tenant test requires second org (skipped; 401 & 404 guards verified in [5] and [6])"

# ---------------------------------------------------------------------------
# 11. Audit events logged
# ---------------------------------------------------------------------------
echo ""
echo "--- [11] Audit Events ---"
# Query audit_logs via DB to verify events were written
if command -v psql &>/dev/null; then
  DB_HOST="${DB_HOST:-localhost}"
  DB_PORT="${DB_PORT:-5432}"
  DB_USER="${DB_USER:-cankora}"
  DB_NAME="${DB_NAME:-cankora_db}"
  DB_PASSWORD="${DB_PASSWORD:-cankora_secret}"

  AUDIT_COUNT=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    -tAc "SELECT COUNT(*) FROM audit_logs WHERE action LIKE 'report.export.%'" 2>/dev/null || echo "")
  if [[ -n "$AUDIT_COUNT" && "$AUDIT_COUNT" -gt 0 ]]; then
    pass "[11-1] Audit events present: $AUDIT_COUNT rows with action LIKE 'report.export.%'"
  else
    fail "[11-1] No audit events found for report.export.* (count=$AUDIT_COUNT)"
  fi

  # Check specific events
  for action in "report.export.requested" "report.export.completed"; do
    COUNT=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
      -tAc "SELECT COUNT(*) FROM audit_logs WHERE action = '$action'" 2>/dev/null || echo "")
    if [[ -n "$COUNT" && "$COUNT" -gt 0 ]]; then
      pass "[11-2] Audit event '$action' present ($COUNT rows)"
    else
      fail "[11-2] Audit event '$action' missing (count=$COUNT)"
    fi
  done
else
  skip "[11-1] psql not available — skipping audit DB check"
  skip "[11-2] Audit event check skipped"
fi

# ---------------------------------------------------------------------------
# 12. Invalid input validation
# ---------------------------------------------------------------------------
echo ""
echo "--- [12] Input Validation ---"
# Missing dataset_key
curl_req POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "Content-Type: application/json" \
  -H "$AUTH_H" \
  -d '{"format":"CSV"}'
assert_status "[12-1] Missing dataset_key → 400" "400" "$HTTP_CODE"

# Invalid format
curl_req POST "$BASE_URL/api/v1/analytics/reports/export/request" \
  -H "Content-Type: application/json" \
  -H "$AUTH_H" \
  -d '{"dataset_key":"budget","format":"TXT"}'
assert_status "[12-2] Invalid format → 400" "400" "$HTTP_CODE"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "================================================="
echo "  UAT-002 Smoke Test Summary"
echo "================================================="
echo -e "  ${GREEN}PASS${NC}: $PASS"
echo -e "  ${RED}FAIL${NC}: $FAIL"
echo -e "  ${YELLOW}SKIP${NC}: $SKIP"
TOTAL=$((PASS+FAIL+SKIP))
echo "  TOTAL: $TOTAL"
echo "================================================="
echo ""
if [[ "$FAIL" -eq 0 ]]; then
  echo -e "${GREEN}ALL CHECKS PASSED${NC} (${PASS}/${TOTAL} PASS, ${SKIP} SKIP)"
  exit 0
else
  echo -e "${RED}${FAIL} CHECK(S) FAILED${NC}"
  exit 1
fi
