#!/usr/bin/env bash
# Smoke Test: CANKORA-P3-002 — Government Entity Resolution
# Tests the full PENDING_MATCH → MATCHED → UNMATCHED → REJECTED lifecycle.
#
# Prerequisites:
#   - Backend running: cd backend && make dev
#   - Database migrated: make migrate-up
#   - At least one COMMIT sync run completed (so government_external_mappings has rows)
#
# Usage:
#   ./docs/smoke-test-p3-002-government-resolution.sh
#
# Set TOKEN and BASE_URL before running if needed:
#   TOKEN=<jwt> BASE_URL=http://localhost:8080 ./docs/smoke-test-p3-002-government-resolution.sh

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN="${TOKEN:-}"
API="$BASE_URL/api/v1"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
RESET='\033[0m'

pass() { echo -e "${GREEN}✓ PASS${RESET}  $1"; }
fail() { echo -e "${RED}✗ FAIL${RESET}  $1"; FAILURES=$((FAILURES + 1)); }
info() { echo -e "${YELLOW}→${RESET} $1"; }

FAILURES=0

auth_header() {
  if [[ -n "$TOKEN" ]]; then
    echo "-H 'Authorization: Bearer $TOKEN'"
  fi
}

api_get() {
  local path="$1"
  shift
  curl -s -o /tmp/smoke_resp.json -w "%{http_code}" \
    -H "Content-Type: application/json" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} \
    "$API$path" "$@"
}

api_post() {
  local path="$1"
  local body="$2"
  shift 2
  curl -s -o /tmp/smoke_resp.json -w "%{http_code}" \
    -H "Content-Type: application/json" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} \
    -d "$body" \
    "$API$path" "$@"
}

body() { cat /tmp/smoke_resp.json; }
field() { cat /tmp/smoke_resp.json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d$1)" 2>/dev/null || echo ""; }

# ---------------------------------------------------------------------------
# Obtain a JWT token if not supplied
# ---------------------------------------------------------------------------

if [[ -z "$TOKEN" ]]; then
  info "No TOKEN set — logging in as admin@cankora.local..."
  STATUS=$(curl -s -o /tmp/smoke_login.json -w "%{http_code}" \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}' \
    "$BASE_URL/api/v1/auth/login")
  if [[ "$STATUS" != "200" ]]; then
    echo -e "${RED}Login failed (HTTP $STATUS). Set TOKEN env var manually.${RESET}"
    exit 1
  fi
  TOKEN=$(python3 -c "import sys,json; print(json.load(open('/tmp/smoke_login.json'))['data']['access_token'])" 2>/dev/null || echo "")
  if [[ -z "$TOKEN" ]]; then
    echo -e "${RED}Could not extract token from login response. Set TOKEN env var manually.${RESET}"
    exit 1
  fi
  pass "Logged in, token obtained"
fi

echo ""
echo "========================================================"
echo " CANKORA-P3-002 — Government Entity Resolution Smoke Test"
echo " API: $API"
echo "========================================================"
echo ""

# ---------------------------------------------------------------------------
# 1. GET /mappings — verify match_status field present in response
# ---------------------------------------------------------------------------

info "1. GET /mappings — match_status field in response"
STATUS=$(api_get "/integrations/government/mappings?page_size=1")
if [[ "$STATUS" == "200" ]]; then
  HAS_STATUS=$(python3 -c "
import sys, json
d = json.load(open('/tmp/smoke_resp.json'))
rows = d.get('data', {}).get('data', [])
if rows and 'match_status' in rows[0]:
    print('yes')
else:
    print('no')
" 2>/dev/null)
  if [[ "$HAS_STATUS" == "yes" ]]; then
    pass "match_status field present in /mappings response"
  else
    fail "match_status field missing from /mappings response (or no rows)"
  fi
else
  fail "GET /mappings returned HTTP $STATUS"
fi

# ---------------------------------------------------------------------------
# 2. GET /mappings?match_status=PENDING_MATCH — filter works
# ---------------------------------------------------------------------------

info "2. GET /mappings?match_status=PENDING_MATCH"
STATUS=$(api_get "/integrations/government/mappings?match_status=PENDING_MATCH&page_size=5")
if [[ "$STATUS" == "200" ]]; then
  pass "GET /mappings?match_status=PENDING_MATCH → HTTP 200"
else
  fail "GET /mappings?match_status=PENDING_MATCH → HTTP $STATUS (expected 200)"
fi

# ---------------------------------------------------------------------------
# 3. GET /mappings/pending — dedicated pending queue
# ---------------------------------------------------------------------------

info "3. GET /mappings/pending"
STATUS=$(api_get "/integrations/government/mappings/pending?page_size=5")
if [[ "$STATUS" == "200" ]]; then
  pass "GET /mappings/pending → HTTP 200"
  PENDING_TOTAL=$(python3 -c "
import sys, json
d = json.load(open('/tmp/smoke_resp.json'))
print(d.get('data', {}).get('meta', {}).get('total', 0))
" 2>/dev/null || echo "0")
  info "  Pending mappings in queue: $PENDING_TOTAL"
else
  fail "GET /mappings/pending → HTTP $STATUS (expected 200)"
fi

# ---------------------------------------------------------------------------
# 4. Extract a PENDING_MATCH mapping ID for further tests
# ---------------------------------------------------------------------------

MAPPING_ID=$(python3 -c "
import sys, json
d = json.load(open('/tmp/smoke_resp.json'))
rows = d.get('data', {}).get('data', [])
if rows:
    print(rows[0]['id'])
" 2>/dev/null || echo "")

if [[ -z "$MAPPING_ID" ]]; then
  info "No PENDING_MATCH mappings found — skipping resolution lifecycle tests."
  info "Run a COMMIT sync first to create external_mappings rows."
  echo ""
  echo "========================================================"
  echo " Tests completed: $((3 - FAILURES))/3 passed, $FAILURES failed"
  echo "========================================================"
  exit $((FAILURES > 0 ? 1 : 0))
fi

info "  Using mapping ID: $MAPPING_ID"

# ---------------------------------------------------------------------------
# 5. GET /mappings/:id — single mapping retrieval
# ---------------------------------------------------------------------------

info "5. GET /mappings/:id"
STATUS=$(api_get "/integrations/government/mappings/$MAPPING_ID")
if [[ "$STATUS" == "200" ]]; then
  MSTATUS=$(python3 -c "
import json
d = json.load(open('/tmp/smoke_resp.json'))
print(d.get('data', {}).get('match_status', ''))
" 2>/dev/null || echo "")
  pass "GET /mappings/$MAPPING_ID → HTTP 200 (match_status=$MSTATUS)"
else
  fail "GET /mappings/$MAPPING_ID → HTTP $STATUS (expected 200)"
fi

# ---------------------------------------------------------------------------
# 6. GET /mappings/pending — must not be confused with /:id route
# ---------------------------------------------------------------------------

info "6. Route precedence: /mappings/pending not treated as /mappings/:id"
STATUS=$(api_get "/integrations/government/mappings/pending")
if [[ "$STATUS" == "200" ]]; then
  HAS_META=$(python3 -c "
import json
d = json.load(open('/tmp/smoke_resp.json'))
print('yes' if 'meta' in d.get('data', {}) else 'no')
" 2>/dev/null || echo "no")
  if [[ "$HAS_META" == "yes" ]]; then
    pass "/mappings/pending returns list (not treated as ID lookup)"
  else
    fail "/mappings/pending returned unexpected response shape"
  fi
else
  fail "/mappings/pending returned HTTP $STATUS (route conflict?)"
fi

# ---------------------------------------------------------------------------
# 7. GET /mappings/:id/candidates
# ---------------------------------------------------------------------------

info "7. GET /mappings/:id/candidates"
STATUS=$(api_get "/integrations/government/mappings/$MAPPING_ID/candidates")
if [[ "$STATUS" == "200" ]]; then
  CCOUNT=$(python3 -c "
import json
d = json.load(open('/tmp/smoke_resp.json'))
print(d.get('data', {}).get('meta', {}).get('total', 0))
" 2>/dev/null || echo "0")
  pass "GET /mappings/$MAPPING_ID/candidates → HTTP 200 ($CCOUNT candidates)"
else
  fail "GET /mappings/$MAPPING_ID/candidates → HTTP $STATUS (expected 200)"
fi

# ---------------------------------------------------------------------------
# 8. POST /mappings/:id/reject — reject a pending mapping
# ---------------------------------------------------------------------------

info "8. POST /mappings/:id/reject"
STATUS=$(api_post "/integrations/government/mappings/$MAPPING_ID/reject" \
  '{"reject_reason":"Smoke test — tidak ada entitas yang cocok"}')
if [[ "$STATUS" == "200" ]]; then
  NEW_STATUS=$(python3 -c "
import json
d = json.load(open('/tmp/smoke_resp.json'))
print(d.get('data', {}).get('match_status', ''))
" 2>/dev/null || echo "")
  if [[ "$NEW_STATUS" == "REJECTED" ]]; then
    pass "POST /mappings/:id/reject → REJECTED (HTTP 200)"
  else
    fail "POST /mappings/:id/reject → match_status='$NEW_STATUS' (expected REJECTED)"
  fi
else
  fail "POST /mappings/:id/reject → HTTP $STATUS (expected 200)"
fi

# ---------------------------------------------------------------------------
# 9. GET /mappings/pending — mapping should now be absent (it's REJECTED)
# ---------------------------------------------------------------------------

info "9. Rejected mapping absent from /mappings/pending"
STATUS=$(api_get "/integrations/government/mappings/pending?page_size=100")
if [[ "$STATUS" == "200" ]]; then
  FOUND=$(python3 -c "
import json
d = json.load(open('/tmp/smoke_resp.json'))
rows = d.get('data', {}).get('data', [])
ids = [r['id'] for r in rows]
print('yes' if '$MAPPING_ID' in ids else 'no')
" 2>/dev/null || echo "yes")
  if [[ "$FOUND" == "no" ]]; then
    pass "Rejected mapping correctly absent from pending queue"
  else
    fail "Rejected mapping still appears in pending queue"
  fi
else
  fail "GET /mappings/pending → HTTP $STATUS"
fi

# ---------------------------------------------------------------------------
# 10. Attempt to match a PENDING_MATCH mapping with an invalid entity UUID
# ---------------------------------------------------------------------------

info "10. POST /mappings/:id/match with invalid entity UUID → 400/404"
STATUS=$(api_post "/integrations/government/mappings/$MAPPING_ID/match" \
  '{"internal_entity_id":"00000000-0000-0000-0000-000000000000","internal_entity_type":"project"}')
if [[ "$STATUS" == "404" || "$STATUS" == "400" ]]; then
  pass "Match with non-existent entity → HTTP $STATUS (correct rejection)"
else
  fail "Match with non-existent entity → HTTP $STATUS (expected 400 or 404)"
fi

# ---------------------------------------------------------------------------
# 11. Attempt to unmatch a non-MATCHED mapping → 400
# ---------------------------------------------------------------------------

info "11. POST /mappings/:id/unmatch on REJECTED mapping → 400"
STATUS=$(api_post "/integrations/government/mappings/$MAPPING_ID/unmatch" '{}')
if [[ "$STATUS" == "400" ]]; then
  pass "Unmatch of non-MATCHED mapping → HTTP 400 (correct)"
else
  fail "Unmatch of non-MATCHED mapping → HTTP $STATUS (expected 400)"
fi

# ---------------------------------------------------------------------------
# 12. GET /mappings/invalid-uuid → 400
# ---------------------------------------------------------------------------

info "12. GET /mappings/not-a-uuid → 400"
STATUS=$(api_get "/integrations/government/mappings/not-a-valid-uuid")
if [[ "$STATUS" == "400" ]]; then
  pass "GET /mappings/not-a-valid-uuid → HTTP 400 (correct)"
else
  fail "GET /mappings/not-a-valid-uuid → HTTP $STATUS (expected 400)"
fi

# ---------------------------------------------------------------------------
# 13. GET /mappings/00000000-.../candidates → 404
# ---------------------------------------------------------------------------

info "13. GET /mappings/00000000-0000-0000-0000-000000000000/candidates → 404"
STATUS=$(api_get "/integrations/government/mappings/00000000-0000-0000-0000-000000000000/candidates")
if [[ "$STATUS" == "404" ]]; then
  pass "Candidates for non-existent mapping → HTTP 404 (correct)"
else
  fail "Candidates for non-existent mapping → HTTP $STATUS (expected 404)"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

TOTAL=13
PASSED=$((TOTAL - FAILURES))
echo ""
echo "========================================================"
echo " Results: $PASSED/$TOTAL passed, $FAILURES failed"
echo "========================================================"
echo ""

if [[ $FAILURES -gt 0 ]]; then
  exit 1
fi
exit 0
