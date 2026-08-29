#!/usr/bin/env bash
# Smoke test P2-002 — Government Connector Foundation
# Usage: bash docs/smoke-test-p2-002-government-connectors.sh
# Requires: curl, jq; backend running at localhost:8080

set -uo pipefail

BASE="http://localhost:8080/api/v1"
PASS=0
FAIL=0
RUN_ID=""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { echo -e "${GREEN}[PASS]${NC} $1"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}[FAIL]${NC} $1"; FAIL=$((FAIL+1)); }
info() { echo -e "${YELLOW}[INFO]${NC} $1"; }

# ---------------------------------------------------------------------------
# 0. Health check
# ---------------------------------------------------------------------------
info "0. Health check"
r=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/health")
if [ "$r" = "200" ]; then pass "Health check => 200"; else fail "Health check => $r (expected 200)"; fi

# ---------------------------------------------------------------------------
# 1. Login
# ---------------------------------------------------------------------------
info "1. Login"
login=$(curl -s -X POST "${BASE}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}')
TOKEN=$(echo "$login" | jq -r '.data.access_token // empty')
if [ -n "$TOKEN" ]; then
  pass "Login => token diterima"
else
  fail "Login => token tidak ditemukan"
  echo "Response: $login"
  exit 1
fi
AUTH="Authorization: Bearer $TOKEN"

# ---------------------------------------------------------------------------
# 2. List connectors
# ---------------------------------------------------------------------------
info "2. List connectors"
r=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH" "${BASE}/integrations/government/connectors")
if [ "$r" = "200" ]; then
  pass "GET /connectors => 200"
else
  fail "GET /connectors => $r (expected 200)"
fi

resp=$(curl -s -H "$AUTH" "${BASE}/integrations/government/connectors")
count=$(echo "$resp" | jq '.data.data | length')
if [ "$count" = "4" ]; then
  pass "List connectors => 4 connectors"
else
  fail "List connectors => $count connectors (expected 4)"
fi

# ---------------------------------------------------------------------------
# 3. Get single connector
# ---------------------------------------------------------------------------
info "3. Get connector — government_project_registry"
r=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH" \
  "${BASE}/integrations/government/connectors/government_project_registry")
if [ "$r" = "200" ]; then pass "GET /connectors/:key => 200"; else fail "GET /connectors/:key => $r"; fi

# ---------------------------------------------------------------------------
# 4. Get connector — invalid key (expect 404)
# ---------------------------------------------------------------------------
info "4. Get connector — invalid key"
r=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH" \
  "${BASE}/integrations/government/connectors/invalid_connector_xyz")
if [ "$r" = "404" ]; then pass "GET /connectors/invalid => 404"; else fail "GET /connectors/invalid => $r (expected 404)"; fi

# ---------------------------------------------------------------------------
# 5. Get config
# ---------------------------------------------------------------------------
info "5. Get config"
r=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH" "${BASE}/integrations/government/config")
if [ "$r" = "200" ]; then pass "GET /config => 200"; else fail "GET /config => $r (expected 200)"; fi

# ---------------------------------------------------------------------------
# 6. Create run — SAMPLE
# ---------------------------------------------------------------------------
info "6. Create run — SAMPLE"
resp=$(curl -s -X POST "${BASE}/integrations/government/runs" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"connector_key":"government_project_registry","dataset_type":"projects","mode":"SAMPLE"}')
http_status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE}/integrations/government/runs" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"connector_key":"government_project_registry","dataset_type":"projects","mode":"SAMPLE","idempotency_key":"smoke-sample-001"}')
if [ "$http_status" = "200" ] || [ "$http_status" = "201" ]; then
  pass "POST /runs (SAMPLE) => $http_status"
else
  fail "POST /runs (SAMPLE) => $http_status (expected 200/201)"
fi

RUN_SAMPLE=$(curl -s -X POST "${BASE}/integrations/government/runs" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"connector_key":"government_project_registry","dataset_type":"projects","mode":"SAMPLE","idempotency_key":"smoke-sample-002"}')
RUN_ID=$(echo "$RUN_SAMPLE" | jq -r '.data.id // empty')
if [ -n "$RUN_ID" ]; then
  pass "SAMPLE run ID: $RUN_ID"
else
  fail "SAMPLE run ID tidak ditemukan"
  echo "Response: $RUN_SAMPLE"
fi

# ---------------------------------------------------------------------------
# 7. Get run
# ---------------------------------------------------------------------------
info "7. Get run"
if [ -n "$RUN_ID" ]; then
  r=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH" \
    "${BASE}/integrations/government/runs/${RUN_ID}")
  if [ "$r" = "200" ]; then pass "GET /runs/:id => 200"; else fail "GET /runs/:id => $r"; fi

  run_status=$(curl -s -H "$AUTH" "${BASE}/integrations/government/runs/${RUN_ID}" | jq -r '.data.status')
  info "Run status: $run_status"
  if [ "$run_status" = "SUCCEEDED" ] || [ "$run_status" = "FAILED" ] || [ "$run_status" = "PENDING" ]; then
    pass "Run status valid: $run_status"
  else
    fail "Run status tidak valid: $run_status"
  fi
else
  fail "Skip get run — RUN_ID kosong"
fi

# ---------------------------------------------------------------------------
# 8. Create run — DRY_RUN
# ---------------------------------------------------------------------------
info "8. Create run — DRY_RUN"
resp_dry=$(curl -s -X POST "${BASE}/integrations/government/runs" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"connector_key":"government_budget_reference","dataset_type":"budget_allocations","mode":"DRY_RUN","idempotency_key":"smoke-dry-001"}')
dry_id=$(echo "$resp_dry" | jq -r '.data.id // empty')
dry_status_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE}/integrations/government/runs" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"connector_key":"government_budget_reference","dataset_type":"budget_allocations","mode":"DRY_RUN","idempotency_key":"smoke-dry-002"}')
if [ "$dry_status_code" = "200" ] || [ "$dry_status_code" = "201" ]; then
  pass "POST /runs (DRY_RUN) => $dry_status_code"
else
  fail "POST /runs (DRY_RUN) => $dry_status_code (expected 200/201)"
fi

# ---------------------------------------------------------------------------
# 9. Create run — COMMIT
# ---------------------------------------------------------------------------
info "9. Create run — COMMIT"
resp_commit=$(curl -s -X POST "${BASE}/integrations/government/runs" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"connector_key":"government_location_reference","dataset_type":"locations","mode":"COMMIT","idempotency_key":"smoke-commit-001"}')
commit_id=$(echo "$resp_commit" | jq -r '.data.id // empty')
commit_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE}/integrations/government/runs" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"connector_key":"government_location_reference","dataset_type":"locations","mode":"COMMIT","idempotency_key":"smoke-commit-002"}')
if [ "$commit_code" = "200" ] || [ "$commit_code" = "201" ]; then
  pass "POST /runs (COMMIT) => $commit_code"
else
  fail "POST /runs (COMMIT) => $commit_code (expected 200/201)"
fi

# ---------------------------------------------------------------------------
# 10. List runs
# ---------------------------------------------------------------------------
info "10. List runs"
r=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH" "${BASE}/integrations/government/runs")
if [ "$r" = "200" ]; then pass "GET /runs => 200"; else fail "GET /runs => $r"; fi

run_count=$(curl -s -H "$AUTH" "${BASE}/integrations/government/runs" | jq '.data | length')
if [ "$run_count" -ge "1" ] 2>/dev/null; then
  pass "List runs => $run_count run(s)"
else
  fail "List runs => $run_count run(s) (expected >= 1)"
fi

# ---------------------------------------------------------------------------
# 11. List records for SAMPLE run
# ---------------------------------------------------------------------------
info "11. List records"
if [ -n "$RUN_ID" ]; then
  r=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH" \
    "${BASE}/integrations/government/runs/${RUN_ID}/records")
  if [ "$r" = "200" ]; then pass "GET /runs/:id/records => 200"; else fail "GET /runs/:id/records => $r"; fi
else
  fail "Skip list records — RUN_ID kosong"
fi

# ---------------------------------------------------------------------------
# 12. List mappings
# ---------------------------------------------------------------------------
info "12. List mappings"
r=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH" "${BASE}/integrations/government/mappings")
if [ "$r" = "200" ]; then pass "GET /mappings => 200"; else fail "GET /mappings => $r"; fi

# ---------------------------------------------------------------------------
# 13. Cancel run yang sudah SUCCEEDED (expect 400 atau 422)
# ---------------------------------------------------------------------------
info "13. Cancel completed run (expect 400/422)"
if [ -n "$RUN_ID" ]; then
  r=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "$AUTH" \
    "${BASE}/integrations/government/runs/${RUN_ID}/cancel")
  if [ "$r" = "400" ] || [ "$r" = "422" ] || [ "$r" = "409" ]; then
    pass "Cancel completed run => $r (transisi tidak valid, diharapkan)"
  else
    fail "Cancel completed run => $r (expected 400/422/409)"
  fi
else
  fail "Skip cancel — RUN_ID kosong"
fi

# ---------------------------------------------------------------------------
# 14. Unauthenticated checks (expect 401)
# ---------------------------------------------------------------------------
info "14. Unauthenticated checks"
for endpoint in \
  "GET ${BASE}/integrations/government/connectors" \
  "GET ${BASE}/integrations/government/runs" \
  "GET ${BASE}/integrations/government/mappings"; do
  method=$(echo "$endpoint" | awk '{print $1}')
  url=$(echo "$endpoint" | awk '{print $2}')
  r=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "$url")
  if [ "$r" = "401" ]; then
    pass "Unauth $method ${url##*/} => 401"
  else
    fail "Unauth $method ${url##*/} => $r (expected 401)"
  fi
done

# ---------------------------------------------------------------------------
# 15. Invalid connector key in create run (expect 400)
# ---------------------------------------------------------------------------
info "15. Invalid connector key in create run"
r=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE}/integrations/government/runs" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"connector_key":"invalid_connector","dataset_type":"projects","mode":"SAMPLE"}')
if [ "$r" = "400" ] || [ "$r" = "422" ]; then
  pass "Create run invalid connector => $r"
else
  fail "Create run invalid connector => $r (expected 400/422)"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "========================================"
echo " Smoke Test P2-002 — Government Connectors"
echo "========================================"
echo -e " ${GREEN}PASS: $PASS${NC}"
echo -e " ${RED}FAIL: $FAIL${NC}"
echo "========================================"
if [ "$FAIL" -eq 0 ]; then
  echo -e " ${GREEN}ALL TESTS PASSED ✓${NC}"
  exit 0
else
  echo -e " ${RED}$FAIL TEST(S) FAILED ✗${NC}"
  exit 1
fi
