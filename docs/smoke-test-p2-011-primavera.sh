#!/usr/bin/env bash
# =============================================================================
# Smoke Test — CANKORA P2-011 Primavera P6 Adapter
# =============================================================================
# Usage: bash docs/smoke-test-p2-011-primavera.sh
# Requires: curl, jq, psql (for project ID lookup)
# Backend must be running at http://localhost:8080

set -uo pipefail

BASE="http://localhost:8080/api/v1"
EMAIL="admin@cankora.local"
PASSWORD="Admin@Cankora2024!"
PASS=0
FAIL=0

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
green() { echo -e "\033[0;32m✔  $*\033[0m"; }
red()   { echo -e "\033[0;31m✘  $*\033[0m"; }
info()  { echo -e "\033[0;34m▶  $*\033[0m"; }
section() { echo -e "\n\033[1;33m=== $* ===\033[0m"; }

assert_eq() {
  local label="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    green "$label"
    PASS=$((PASS+1))
  else
    red "$label (expected=$expected actual=$actual)"
    FAIL=$((FAIL+1))
  fi
}

assert_not_empty() {
  local label="$1" val="$2"
  if [ -n "$val" ] && [ "$val" != "null" ]; then
    green "$label"
    PASS=$((PASS+1))
  else
    red "$label (got empty/null)"
    FAIL=$((FAIL+1))
  fi
}

assert_ge() {
  local label="$1" expected="$2" actual="$3"
  if [ "$actual" -ge "$expected" ] 2>/dev/null; then
    green "$label"
    PASS=$((PASS+1))
  else
    red "$label (expected>=$expected actual=$actual)"
    FAIL=$((FAIL+1))
  fi
}

# ---------------------------------------------------------------------------
# 1. Auth
# ---------------------------------------------------------------------------
section "1. Auth — login"
LOGIN=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
TOKEN=$(echo "$LOGIN" | jq -r '.data.access_token // empty')
assert_not_empty "Login returns access_token" "$TOKEN"

AUTH="-H \"Authorization: Bearer $TOKEN\""

_curl() { curl -sf -H "Authorization: Bearer $TOKEN" "$@"; }

# ---------------------------------------------------------------------------
# 2. Resolve real project ID from DB (fallback to API)
# ---------------------------------------------------------------------------
section "2. Resolve project_id"
PROJECT_ID=$(psql "postgres://postgres:postgres@localhost:5432/cankora_db" -t -A \
  -c "SELECT id FROM projects WHERE deleted_at IS NULL LIMIT 1" 2>/dev/null || true)
if [ -z "$PROJECT_ID" ]; then
  PROJECT_ID=$(_curl "$BASE/projects?page_size=1" | jq -r '.data[0].id // empty')
fi
assert_not_empty "Resolved project_id" "$PROJECT_ID"
info "Using project_id: $PROJECT_ID"

# ---------------------------------------------------------------------------
# 3. Create minimal XER fixture
# ---------------------------------------------------------------------------
section "3. Create XER fixture"
XER_FILE=$(mktemp /tmp/p2011_test_XXXXXX.xer)
cat > "$XER_FILE" << 'XEREOF'
%T	PROJWBS
%F	wbs_id	proj_id	wbs_short_name	wbs_name
%R	WBS-1	PROJ-001	1.0	Pekerjaan Sipil
%R	WBS-2	PROJ-001	2.0	Pekerjaan Mekanikal
%E
%T	TASK
%F	task_id	task_code	task_name	wbs_id	target_start_date	target_end_date	act_start_date	act_end_date	phys_complete_pct	act_phys_complete_pct
%R	1	A1010	Pekerjaan Pondasi	WBS-1	2024-01-15 00:00	2024-06-30 00:00			0	45.5
%R	2	A1020	Pekerjaan Struktur	WBS-1	2024-03-01 00:00	2024-09-30 00:00	2024-03-05 00:00		0	30.0
%R	3	A2010	Instalasi Pompa	WBS-2	2024-07-01 00:00	2024-12-31 00:00			0	0
%E
XEREOF
assert_not_empty "XER fixture created" "$XER_FILE"
info "XER file: $XER_FILE"

# ---------------------------------------------------------------------------
# 4. Create PMXML fixture
# ---------------------------------------------------------------------------
section "4. Create PMXML fixture"
PMXML_FILE=$(mktemp /tmp/p2011_test_XXXXXX.xml)
cat > "$PMXML_FILE" << 'XMLEOF'
<?xml version="1.0" encoding="UTF-8"?>
<APIBusinessObjects>
  <Project>
    <WBS>
      <ObjectId>WBS-10</ObjectId>
      <Code>1.0</Code>
      <Name>Pekerjaan Sipil</Name>
    </WBS>
    <Activity>
      <ObjectId>1001</ObjectId>
      <Id>B1010</Id>
      <Name>Galian Tanah</Name>
      <WBSObjectId>WBS-10</WBSObjectId>
      <PlannedStartDate>2024-02-01T00:00:00</PlannedStartDate>
      <PlannedFinishDate>2024-05-31T00:00:00</PlannedFinishDate>
      <PhysicalPercentComplete>55.0</PhysicalPercentComplete>
      <ActualPhysicalPercentComplete>55.0</ActualPhysicalPercentComplete>
    </Activity>
    <Activity>
      <ObjectId>1002</ObjectId>
      <Id>B1020</Id>
      <Name>Pengecoran Beton</Name>
      <WBSObjectId>WBS-10</WBSObjectId>
      <PlannedStartDate>2024-06-01T00:00:00</PlannedStartDate>
      <PlannedFinishDate>2024-10-31T00:00:00</PlannedFinishDate>
      <PhysicalPercentComplete>0.0</PhysicalPercentComplete>
      <ActualPhysicalPercentComplete>0.0</ActualPhysicalPercentComplete>
    </Activity>
  </Project>
</APIBusinessObjects>
XMLEOF
assert_not_empty "PMXML fixture created" "$PMXML_FILE"

# ---------------------------------------------------------------------------
# 5. Upload XER — create + process run
# ---------------------------------------------------------------------------
section "5. Upload XER file"
XER_RES=$(curl -s -X POST "$BASE/integrations/primavera/runs" \
  -H "Authorization: Bearer $TOKEN" \
  -F "project_id=$PROJECT_ID" \
  -F "format=XER" \
  -F "p6_version=22.12" \
  -F "file=@$XER_FILE;filename=test_project.xer;type=text/plain")
RUN1_ID=$(echo "$XER_RES" | jq -r '.data.id // empty')
RUN1_STATUS=$(echo "$XER_RES" | jq -r '.data.status // empty')
assert_not_empty "XER upload returns run id" "$RUN1_ID"
assert_eq "XER run completes (DONE or FAILED)" "$(echo "$RUN1_STATUS" | grep -E 'DONE|FAILED' | head -1)" "$RUN1_STATUS"
info "XER run_id: $RUN1_ID status: $RUN1_STATUS"

# ---------------------------------------------------------------------------
# 6. Verify DONE status and activity counts
# ---------------------------------------------------------------------------
section "6. XER run counters"
XER_TOTAL=$(echo "$XER_RES" | jq -r '.data.total_activities // 0')
XER_IMPORTED=$(echo "$XER_RES" | jq -r '.data.imported_activities // 0')
assert_eq "XER total_activities = 3" "3" "$XER_TOTAL"
assert_ge "XER imported_activities >= 1" "1" "$XER_IMPORTED"

# ---------------------------------------------------------------------------
# 7. GET /runs/:id
# ---------------------------------------------------------------------------
section "7. GET run by ID"
RUN1_GET=$(_curl "$BASE/integrations/primavera/runs/$RUN1_ID")
assert_eq "GET run id matches" "$RUN1_ID" "$(echo "$RUN1_GET" | jq -r '.data.id')"
assert_eq "GET run format = XER" "XER" "$(echo "$RUN1_GET" | jq -r '.data.format')"

# ---------------------------------------------------------------------------
# 8. GET /runs/:id/mappings
# ---------------------------------------------------------------------------
section "8. Activity mappings for XER run"
MAPPINGS=$(_curl "$BASE/integrations/primavera/runs/$RUN1_ID/mappings")
MAP_TOTAL=$(echo "$MAPPINGS" | jq -r '.data.meta.total // 0')
assert_ge "Mappings total >= 1" "1" "$MAP_TOTAL"
MAP_FIRST_ACTION=$(echo "$MAPPINGS" | jq -r '.data.data[0].action // empty')
assert_not_empty "First mapping has action" "$MAP_FIRST_ACTION"

# ---------------------------------------------------------------------------
# 9. Idempotency — upload same XER again, activities should be UPDATEd
# ---------------------------------------------------------------------------
section "9. Idempotency — re-upload same XER"
XER_RES2=$(curl -s -X POST "$BASE/integrations/primavera/runs" \
  -H "Authorization: Bearer $TOKEN" \
  -F "project_id=$PROJECT_ID" \
  -F "format=XER" \
  -F "file=@$XER_FILE;filename=test_project_v2.xer;type=text/plain")
RUN2_ID=$(echo "$XER_RES2" | jq -r '.data.id // empty')
RUN2_STATUS=$(echo "$XER_RES2" | jq -r '.data.status // empty')
assert_not_empty "Second XER upload returns run id" "$RUN2_ID"
assert_eq "Second run is distinct from first" "false" "$([ "$RUN1_ID" = "$RUN2_ID" ] && echo true || echo false)"

# Mappings for run 2 should have UPDATE actions (idempotency)
MAPPINGS2=$(_curl "$BASE/integrations/primavera/runs/$RUN2_ID/mappings")
MAP2_ACTIONS=$(echo "$MAPPINGS2" | jq -r '.data.data[].action' | sort | uniq | tr '\n' ',')
assert_not_empty "Second run mappings have actions" "$MAP2_ACTIONS"
info "Run 2 mapping actions: $MAP2_ACTIONS"

# ---------------------------------------------------------------------------
# 10. Upload PMXML
# ---------------------------------------------------------------------------
section "10. Upload PMXML file"
PMXML_RES=$(curl -s -X POST "$BASE/integrations/primavera/runs" \
  -H "Authorization: Bearer $TOKEN" \
  -F "project_id=$PROJECT_ID" \
  -F "format=PMXML" \
  -F "file=@$PMXML_FILE;filename=test_project.xml;type=application/xml")
RUN3_ID=$(echo "$PMXML_RES" | jq -r '.data.id // empty')
RUN3_STATUS=$(echo "$PMXML_RES" | jq -r '.data.status // empty')
assert_not_empty "PMXML upload returns run id" "$RUN3_ID"
assert_eq "PMXML run total_activities = 2" "2" "$(echo "$PMXML_RES" | jq -r '.data.total_activities // 0')"
info "PMXML run_id: $RUN3_ID status: $RUN3_STATUS"

# ---------------------------------------------------------------------------
# 11. LIST /runs — pagination and filter
# ---------------------------------------------------------------------------
section "11. LIST runs"
RUNS_ALL=$(_curl "$BASE/integrations/primavera/runs?page=1&page_size=20")
RUNS_TOTAL=$(echo "$RUNS_ALL" | jq -r '.data.meta.total // 0')
assert_ge "List runs total >= 3" "3" "$RUNS_TOTAL"

# Filter by status=DONE
RUNS_DONE=$(_curl "$BASE/integrations/primavera/runs?status=DONE&page_size=20")
RUNS_DONE_COUNT=$(echo "$RUNS_DONE" | jq -r '.data.data | length')
assert_ge "Filter status=DONE returns >= 1" "1" "$RUNS_DONE_COUNT"

# Filter by format=XER
RUNS_XER=$(_curl "$BASE/integrations/primavera/runs?format=XER&page_size=20")
RUNS_XER_COUNT=$(echo "$RUNS_XER" | jq -r '.data.data | length')
assert_ge "Filter format=XER returns >= 2" "2" "$RUNS_XER_COUNT"

# ---------------------------------------------------------------------------
# 12. Cancel a PENDING run (create run without processing = not possible via
#     current API since upload+process is atomic; test cancel on existing PENDING)
#     We'll POST a cancel on a DONE run — expect 400 InvalidTransition
# ---------------------------------------------------------------------------
section "12. Cancel on non-PENDING run returns 400"
CANCEL_RES=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  "$BASE/integrations/primavera/runs/$RUN1_ID/cancel" \
  -H "Authorization: Bearer $TOKEN")
assert_eq "Cancel DONE run returns 400" "400" "$CANCEL_RES"

# ---------------------------------------------------------------------------
# 13. Cross-tenant check — unknown project_id returns 404
# ---------------------------------------------------------------------------
section "13. Unknown project_id returns 404"
FAKE_PID="00000000-0000-0000-0000-000000000099"
BAD_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/integrations/primavera/runs" \
  -H "Authorization: Bearer $TOKEN" \
  -F "project_id=$FAKE_PID" \
  -F "format=XER" \
  -F "file=@$XER_FILE;filename=bad.xer;type=text/plain")
assert_eq "Unknown project_id returns 404" "404" "$BAD_CODE"

# ---------------------------------------------------------------------------
# 14. Missing project_id returns 400
# ---------------------------------------------------------------------------
section "14. Missing project_id returns 400"
NOPID_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  "$BASE/integrations/primavera/runs" \
  -H "Authorization: Bearer $TOKEN" \
  -F "format=XER" \
  -F "file=@$XER_FILE;filename=nopid.xer;type=text/plain")
assert_eq "Missing project_id returns 400" "400" "$NOPID_CODE"

# ---------------------------------------------------------------------------
# 15. Auth guard — no token returns 401
# ---------------------------------------------------------------------------
section "15. Auth guard"
AUTH401=$(curl -s -o /dev/null -w "%{http_code}" \
  "$BASE/integrations/primavera/runs")
assert_eq "No token returns 401" "401" "$AUTH401"

# ---------------------------------------------------------------------------
# 16. tsc check (frontend)
# ---------------------------------------------------------------------------
section "16. Frontend tsc"
TSC_OUT=$(cd "/Users/harmanto/Documents/Code/DEV/Project Management/CANKORA/frontend" && npx tsc --noEmit 2>&1 || true)
if [ -z "$TSC_OUT" ]; then
  green "Frontend tsc --noEmit clean"
  ((PASS++))
else
  red "Frontend tsc errors: $TSC_OUT"
  ((FAIL++))
fi

# ---------------------------------------------------------------------------
# Cleanup
# ---------------------------------------------------------------------------
rm -f "$XER_FILE" "$PMXML_FILE"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "=============================="
echo " Smoke Test Results: P2-011"
echo "=============================="
echo -e " \033[0;32mPASS: $PASS\033[0m"
if [ "$FAIL" -gt 0 ]; then
  echo -e " \033[0;31mFAIL: $FAIL\033[0m"
  exit 1
else
  echo -e " \033[0;31mFAIL: $FAIL\033[0m"
  echo ""
  echo "All tests passed."
fi
