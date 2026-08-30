#!/usr/bin/env bash
# Smoke Test: CANKORA-P3-003-HARDEN — Data Governance Official Validation Hardening
#
# Covers:
#   1. unknown entity_type ditolak saat create (400)
#   2. non-existent government_mapping → INVALID setelah review
#   3. PENDING_MATCH government_mapping → INVALID (double-check)
#   4. REJECTED government_mapping → INVALID
#   5. MATCHED government_mapping → VALID (jika ada di DB)
#   6. Lock submission memblokir submission baru di periode sama (409)
#   7. Full-year lock tidak bisa duplicate (409)
#   8. Malformed JSON → 400 di action endpoints
#   9. submitted_by = NULL di DRAFT (set hanya saat Submit)
#  10. created_by diset di DRAFT
#
# Prerequisites:
#   - Backend running: cd backend && make dev
#   - Database migrated: make migrate-up (termasuk 000034, 000035)
#   - PostgreSQL accessible via PGPASSWORD env
#
# Usage:
#   ./docs/smoke-test-p3-003-harden.sh
#   TOKEN=<jwt> BASE_URL=http://localhost:8080 ./docs/smoke-test-p3-003-harden.sh

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN="${TOKEN:-}"
API="$BASE_URL/api/v1"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
RESET='\033[0m'

pass() { echo -e "${GREEN}✓ PASS${RESET}  $1"; PASS_COUNT=$((PASS_COUNT + 1)); }
fail() { echo -e "${RED}✗ FAIL${RESET}  $1"; FAILURES=$((FAILURES + 1)); }
info() { echo -e "${YELLOW}→${RESET} $1"; }
skip() { echo -e "  ${YELLOW}⚠ SKIP${RESET}  $1"; }

FAILURES=0
PASS_COUNT=0

api_get() {
  local path="$1"; shift
  curl -s -o /tmp/smoke_harden_resp.json -w "%{http_code}" \
    -H "Content-Type: application/json" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} \
    "$API$path" "$@"
}

api_post() {
  local path="$1"; local body="$2"; shift 2
  curl -s -o /tmp/smoke_harden_resp.json -w "%{http_code}" \
    -H "Content-Type: application/json" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} \
    -d "$body" "$API$path" "$@"
}

api_post_raw() {
  local path="$1"; local body="$2"; shift 2
  curl -s -o /tmp/smoke_harden_resp.json -w "%{http_code}" \
    -H "Content-Type: application/json" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} \
    --data-raw "$body" "$API$path" "$@"
}

body() { cat /tmp/smoke_harden_resp.json; }
field() {
  cat /tmp/smoke_harden_resp.json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d$1)" 2>/dev/null || echo ""
}
dbq() {
  PGPASSWORD=cankora_secret psql -h localhost -U cankora -d cankora_db -tAc "$1" 2>/dev/null || echo ""
}

# ---------------------------------------------------------------------------
# 1. Health check
# ---------------------------------------------------------------------------
info "Health check"
HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
[[ "$HTTP" == "200" ]] && pass "health endpoint 200" || fail "health endpoint $HTTP"

# ---------------------------------------------------------------------------
# 2. Login
# ---------------------------------------------------------------------------
info "Login"
if [[ -z "$TOKEN" ]]; then
  RESP=$(curl -s -X POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}')
  TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['access_token'])" 2>/dev/null || echo "")
fi
[[ -n "$TOKEN" ]] && pass "login obtained access_token" || { fail "login failed"; exit 1; }

# ---------------------------------------------------------------------------
# 2b. Idempotent fixture setup: PENDING_MATCH / REJECTED / MATCHED
# ---------------------------------------------------------------------------
# Governance item validation depends on government_external_mappings rows.
# Without these fixtures, the smoke test would SKIP (false pass). This block
# guarantees at least one row of each match_status exists so the hardening
# checks below always run. Uses ON CONFLICT-safe SQL (INSERT ... ON CONFLICT
# DO NOTHING on the natural key) so it is idempotent across runs.
info "2b. Fixture setup (idempotent): government mappings"
FIXTURE_ORG=$(dbq "SELECT id FROM organizations ORDER BY created_at LIMIT 1" | tr -d '[:space:]')
FIXTURE_PROJECT=$(dbq "SELECT id FROM projects WHERE organization_id='$FIXTURE_ORG' AND deleted_at IS NULL ORDER BY created_at LIMIT 1" | tr -d '[:space:]')
if [[ -z "$FIXTURE_ORG" || -z "$FIXTURE_PROJECT" ]]; then
  fail "fixture setup: no organization/project found (org='$FIXTURE_ORG' project='$FIXTURE_PROJECT')"
else
  # Create 3 mappings with distinct external_ids under a dedicated connector key.
  dbq "INSERT INTO government_external_mappings (organization_id, connector_key, dataset_type, external_id, internal_entity_type, match_status, source_payload_hash)
       VALUES ('$FIXTURE_ORG', 'REL001_FIXTURE', 'location', 'rel001-pending', 'location', 'PENDING_MATCH', 'rel001-pending')
       ON CONFLICT (organization_id, connector_key, dataset_type, external_id) DO NOTHING;" > /dev/null
  dbq "INSERT INTO government_external_mappings (organization_id, connector_key, dataset_type, external_id, internal_entity_type, match_status, source_payload_hash)
       VALUES ('$FIXTURE_ORG', 'REL001_FIXTURE', 'location', 'rel001-rejected', 'location', 'REJECTED', 'rel001-rejected')
       ON CONFLICT (organization_id, connector_key, dataset_type, external_id) DO NOTHING;" > /dev/null
  dbq "INSERT INTO government_external_mappings (organization_id, connector_key, dataset_type, external_id, internal_entity_type, internal_entity_id, match_status, match_confidence, match_reason, source_payload_hash)
       VALUES ('$FIXTURE_ORG', 'REL001_FIXTURE', 'location', 'rel001-matched', 'project', '$FIXTURE_PROJECT', 'MATCHED', 90, 'rel001 fixture', 'rel001-matched')
       ON CONFLICT (organization_id, connector_key, dataset_type, external_id) DO NOTHING;" > /dev/null
  pass "fixture setup ensured PENDING_MATCH/REJECTED/MATCHED mappings"
fi

# Tahun unik untuk section 3-6 (menghindari lock period dari run sebelumnya)
HARDEN_YEAR=$((2040 + RANDOM % 30))

# ---------------------------------------------------------------------------
# 3. Unknown entity_type ditolak saat create (400)
# ---------------------------------------------------------------------------
info "3. Unknown entity_type rejected at create time"
STATUS=$(api_post "/governance/submissions" "{
  \"dataset_type\": \"PROJECT_PROGRESS\",
  \"source_type\": \"MANUAL\",
  \"period_year\": $HARDEN_YEAR,
  \"items\": [{\"entity_type\": \"UNKNOWN_TYPE\", \"action\": \"VALIDATE_ONLY\", \"payload_after\": {\"x\": 1}}]
}")
if [[ "$STATUS" == "400" ]]; then
  pass "unknown entity_type -> 400"
else
  fail "unknown entity_type returned $STATUS (expected 400): $(body)"
fi

# ---------------------------------------------------------------------------
# 4. Non-existent government_mapping → INVALID after review
# ---------------------------------------------------------------------------
info "4. Non-existent government_mapping → INVALID after review"
NONEXIST_ID="00000000-dead-beef-0000-000000000000"

STATUS=$(api_post "/governance/submissions" "{
  \"dataset_type\": \"PROJECT_PROGRESS\",
  \"source_type\": \"GOVERNMENT\",
  \"period_year\": $HARDEN_YEAR,
  \"items\": [{
    \"entity_type\": \"government_mapping\",
    \"entity_id\": \"$NONEXIST_ID\",
    \"action\": \"VALIDATE_ONLY\",
    \"payload_after\": {\"x\": 1}
  }]
}")
if [[ "$STATUS" == "201" ]]; then
  SUB4_ID=$(field "['data']['id']")
  pass "created submission with gov_mapping item"

  STATUS=$(api_post "/governance/submissions/$SUB4_ID/submit" '{}')
  if [[ "$STATUS" == "200" ]]; then
    pass "submitted -> SUBMITTED"
    STATUS=$(api_post "/governance/submissions/$SUB4_ID/review" '{}')
    if [[ "$STATUS" == "200" ]]; then
      pass "review started -> IN_REVIEW"
      # Check item validation_status
      ITEM_STATUS=$(field "['data']['items'][0]['validation_status']" 2>/dev/null || \
        curl -s -H "Authorization: Bearer $TOKEN" "$API/governance/submissions/$SUB4_ID" | \
        python3 -c "import sys,json; items=json.load(sys.stdin)['data']['items']; print(items[0]['validation_status'] if items else 'NO_ITEMS')" 2>/dev/null || echo "")
      if [[ "$ITEM_STATUS" == "INVALID" ]]; then
        pass "non-existent government_mapping -> INVALID"
      else
        # try getting detail
        DETAIL=$(curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" "$API/governance/submissions/$SUB4_ID")
        ITEM_STATUS2=$(echo "$DETAIL" | python3 -c "import sys,json; d=json.load(sys.stdin); items=d.get('data',{}).get('items',[]); print(items[0].get('validation_status','?') if items else 'NO_ITEMS')" 2>/dev/null || echo "")
        if [[ "$ITEM_STATUS2" == "INVALID" ]]; then
          pass "non-existent government_mapping -> INVALID (from detail)"
        else
          fail "non-existent gov_mapping should be INVALID, got: $ITEM_STATUS2"
        fi
      fi
    else
      fail "review returned $STATUS (expected 200)"
    fi
  else
    fail "submit for non-existent mapping returned $STATUS"
  fi
else
  fail "create submission with gov_mapping returned $STATUS (expected 201)"
fi

# ---------------------------------------------------------------------------
# 5. PENDING_MATCH government mapping → INVALID (DB-based)
# ---------------------------------------------------------------------------
info "5. PENDING_MATCH government_mapping → INVALID"
PENDING_ID=$(dbq "SELECT id FROM government_external_mappings WHERE match_status='PENDING_MATCH' LIMIT 1" 2>/dev/null | tr -d '[:space:]' || echo "")
if [[ -n "$PENDING_ID" && "$PENDING_ID" != "" ]]; then
  STATUS=$(api_post "/governance/submissions" "{
    \"dataset_type\": \"PROJECT_PROGRESS\",
    \"source_type\": \"GOVERNMENT\",
    \"period_year\": $HARDEN_YEAR,
    \"items\": [{
      \"entity_type\": \"government_mapping\",
      \"entity_id\": \"$PENDING_ID\",
      \"action\": \"VALIDATE_ONLY\",
      \"payload_after\": {}
    }]
  }")
  if [[ "$STATUS" == "201" ]]; then
    PENDING_SUB_ID=$(field "['data']['id']")
    api_post "/governance/submissions/$PENDING_SUB_ID/submit" '{}' > /dev/null
    api_post "/governance/submissions/$PENDING_SUB_ID/review" '{}' > /dev/null
    DETAIL=$(curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" "$API/governance/submissions/$PENDING_SUB_ID")
    ITEM_VS=$(echo "$DETAIL" | python3 -c "import sys,json; d=json.load(sys.stdin); items=d.get('data',{}).get('items',[]); print(items[0].get('validation_status','?') if items else 'NO_ITEMS')" 2>/dev/null || echo "")
    if [[ "$ITEM_VS" == "INVALID" ]]; then
      pass "PENDING_MATCH gov_mapping -> INVALID"
    else
      fail "PENDING_MATCH gov_mapping should be INVALID, got: $ITEM_VS"
    fi
  else
    fail "create submission for PENDING_MATCH returned $STATUS"
  fi
else
  skip "No PENDING_MATCH mapping in DB — skipping check 5"
fi

# ---------------------------------------------------------------------------
# 6. REJECTED government_mapping → INVALID
# ---------------------------------------------------------------------------
info "6. REJECTED government_mapping → INVALID"
REJECTED_ID=$(dbq "SELECT id FROM government_external_mappings WHERE match_status='REJECTED' LIMIT 1" 2>/dev/null | tr -d '[:space:]' || echo "")
if [[ -n "$REJECTED_ID" && "$REJECTED_ID" != "" ]]; then
  STATUS=$(api_post "/governance/submissions" "{
    \"dataset_type\": \"PROJECT_PROGRESS\",
    \"source_type\": \"GOVERNMENT\",
    \"period_year\": $HARDEN_YEAR,
    \"items\": [{
      \"entity_type\": \"government_mapping\",
      \"entity_id\": \"$REJECTED_ID\",
      \"action\": \"VALIDATE_ONLY\",
      \"payload_after\": {}
    }]
  }")
  if [[ "$STATUS" == "201" ]]; then
    REJ_SUB_ID=$(field "['data']['id']")
    api_post "/governance/submissions/$REJ_SUB_ID/submit" '{}' > /dev/null
    api_post "/governance/submissions/$REJ_SUB_ID/review" '{}' > /dev/null
    DETAIL=$(curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" "$API/governance/submissions/$REJ_SUB_ID")
    ITEM_VS=$(echo "$DETAIL" | python3 -c "import sys,json; d=json.load(sys.stdin); items=d.get('data',{}).get('items',[]); print(items[0].get('validation_status','?') if items else 'NO_ITEMS')" 2>/dev/null || echo "")
    if [[ "$ITEM_VS" == "INVALID" ]]; then
      pass "REJECTED gov_mapping -> INVALID"
    else
      fail "REJECTED gov_mapping should be INVALID, got: $ITEM_VS"
    fi
  else
    fail "create submission for REJECTED mapping returned $STATUS"
  fi
else
  skip "No REJECTED mapping in DB — skipping check 6"
fi

# ---------------------------------------------------------------------------
# 7. MATCHED government_mapping → VALID (happy path)
# ---------------------------------------------------------------------------
info "7. MATCHED government_mapping → VALID"
MATCHED_ID=$(dbq "SELECT id FROM government_external_mappings WHERE match_status='MATCHED' LIMIT 1" 2>/dev/null | tr -d '[:space:]' || echo "")
if [[ -n "$MATCHED_ID" && "$MATCHED_ID" != "" ]]; then
  STATUS=$(api_post "/governance/submissions" "{
    \"dataset_type\": \"PROJECT_PROGRESS\",
    \"source_type\": \"GOVERNMENT\",
    \"period_year\": $HARDEN_YEAR,
    \"items\": [{
      \"entity_type\": \"government_mapping\",
      \"entity_id\": \"$MATCHED_ID\",
      \"action\": \"VALIDATE_ONLY\",
      \"payload_after\": {}
    }]
  }")
  if [[ "$STATUS" == "201" ]]; then
    MATCHED_SUB_ID=$(field "['data']['id']")
    api_post "/governance/submissions/$MATCHED_SUB_ID/submit" '{}' > /dev/null
    api_post "/governance/submissions/$MATCHED_SUB_ID/review" '{}' > /dev/null
    DETAIL=$(curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" "$API/governance/submissions/$MATCHED_SUB_ID")
    ITEM_VS=$(echo "$DETAIL" | python3 -c "import sys,json; d=json.load(sys.stdin); items=d.get('data',{}).get('items',[]); print(items[0].get('validation_status','?') if items else 'NO_ITEMS')" 2>/dev/null || echo "")
    if [[ "$ITEM_VS" == "VALID" ]]; then
      pass "MATCHED gov_mapping -> VALID"
    else
      fail "MATCHED gov_mapping should be VALID, got: $ITEM_VS"
    fi
  else
    fail "create submission for MATCHED mapping returned $STATUS"
  fi
else
  fail "No MATCHED mapping in DB (fixture setup should have created one) — check 7"
fi

# ---------------------------------------------------------------------------
# 8. Lock submission memblokir submission baru di periode sama
# ---------------------------------------------------------------------------
info "8. Lock submission blocks new submission in same period"
# Gunakan periode unik agar test idempotent (tidak bentrok antar-run).
LOCK_YEAR=$((2090 + RANDOM % 5))
LOCK_MONTH=$((1 + RANDOM % 12))
LOCK_DATASET="RISK"

# Buat submission baru untuk periode ini dan lock-kan
STATUS=$(api_post "/governance/submissions" "{
  \"dataset_type\": \"$LOCK_DATASET\",
  \"source_type\": \"MANUAL\",
  \"period_year\": $LOCK_YEAR,
  \"period_month\": $LOCK_MONTH,
  \"items\": [{\"entity_type\": \"project\", \"action\": \"VALIDATE_ONLY\", \"payload_after\": {\"x\": 1}}]
}")
if [[ "$STATUS" == "201" ]]; then
  LOCK_SUB_ID=$(field "['data']['id']")
  api_post "/governance/submissions/$LOCK_SUB_ID/submit" '{}' > /dev/null
  api_post "/governance/submissions/$LOCK_SUB_ID/review" '{}' > /dev/null
  api_post "/governance/submissions/$LOCK_SUB_ID/approve" '' > /dev/null
  STATUS_LOCK=$(api_post "/governance/submissions/$LOCK_SUB_ID/lock" '{"lock_reason":"smoke test"}')
  if [[ "$STATUS_LOCK" == "200" ]]; then
    pass "locked submission for period $LOCK_YEAR/$LOCK_MONTH $LOCK_DATASET"
    # Now creating another submission for same period should be 409
    STATUS2=$(api_post "/governance/submissions" "{
      \"dataset_type\": \"$LOCK_DATASET\",
      \"source_type\": \"MANUAL\",
      \"period_year\": $LOCK_YEAR,
      \"period_month\": $LOCK_MONTH,
      \"items\": [{\"entity_type\": \"project\", \"action\": \"VALIDATE_ONLY\", \"payload_after\": {\"x\": 1}}]
    }")
    if [[ "$STATUS2" == "409" ]]; then
      pass "new submission in locked period -> 409"
    else
      fail "new submission in locked period returned $STATUS2 (expected 409)"
    fi
  else
    fail "lock submission returned $STATUS_LOCK (expected 200)"
  fi
else
  fail "create submission for lock test returned $STATUS (expected 201)"
fi

# ---------------------------------------------------------------------------
# 9. Full-year lock tidak bisa duplicate
# ---------------------------------------------------------------------------
info "9. Full-year lock (period_month=null) cannot be duplicated"
FULLYEAR_YEAR=2099
FULLYEAR_DATASET="BENEFIT"

STATUS=$(api_post "/governance/lock-periods" "{
  \"dataset_type\": \"$FULLYEAR_DATASET\",
  \"period_year\": $FULLYEAR_YEAR,
  \"lock_reason\": \"smoke test full-year\"
}")
if [[ "$STATUS" == "201" ]]; then
  pass "created full-year lock period ($FULLYEAR_YEAR $FULLYEAR_DATASET)"
  # Duplicate should be 409
  STATUS2=$(api_post "/governance/lock-periods" "{
    \"dataset_type\": \"$FULLYEAR_DATASET\",
    \"period_year\": $FULLYEAR_YEAR,
    \"lock_reason\": \"duplicate attempt\"
  }")
  if [[ "$STATUS2" == "409" ]]; then
    pass "duplicate full-year lock -> 409"
  else
    fail "duplicate full-year lock returned $STATUS2 (expected 409)"
  fi
elif [[ "$STATUS" == "409" ]]; then
  pass "full-year lock already exists (409 = correct)"
  STATUS2=$(api_post "/governance/lock-periods" "{
    \"dataset_type\": \"$FULLYEAR_DATASET\",
    \"period_year\": $FULLYEAR_YEAR,
    \"lock_reason\": \"duplicate attempt\"
  }")
  if [[ "$STATUS2" == "409" ]]; then
    pass "duplicate full-year lock -> 409"
  else
    fail "duplicate full-year lock returned $STATUS2 (expected 409)"
  fi
else
  fail "create full-year lock period returned $STATUS (expected 201 or 409)"
fi

# ---------------------------------------------------------------------------
# 10. Malformed JSON → 400 di action endpoints
# ---------------------------------------------------------------------------
info "10. Malformed JSON -> 400 at action endpoints"
# Gunakan tahun unik agar tidak bentrok dengan lock period dari run sebelumnya.
MALFORM_YEAR=$((2080 + RANDOM % 5))

# Buat fresh submission untuk dipakai malformed JSON test
STATUS=$(api_post "/governance/submissions" "{
  \"dataset_type\": \"DOCUMENT\",
  \"source_type\": \"MANUAL\",
  \"period_year\": $MALFORM_YEAR,
  \"items\": [{\"entity_type\": \"project\", \"action\": \"VALIDATE_ONLY\", \"payload_after\": {\"x\": 1}}]
}")
if [[ "$STATUS" == "201" ]]; then
  MALFORM_SUB_ID=$(field "['data']['id']")
  pass "created submission for malformed JSON test"

  # Submit terlebih dahulu (empty body valid)
  STATUS_SUBMIT=$(api_post "/governance/submissions/$MALFORM_SUB_ID/submit" '{}')
  if [[ "$STATUS_SUBMIT" == "200" ]]; then
    # Malformed JSON on review
    STATUS_MAL=$(api_post_raw "/governance/submissions/$MALFORM_SUB_ID/review" '{"invalid_json": true, "bad":')
    if [[ "$STATUS_MAL" == "400" ]]; then
      pass "malformed JSON on review -> 400"
    else
      fail "malformed JSON on review returned $STATUS_MAL (expected 400)"
    fi
  else
    fail "submit for malformed JSON test returned $STATUS_SUBMIT (expected 200)"
  fi

  # Test malformed JSON on lock-periods/:id/lock endpoint (tahun unik)
  MALFORM_LP_YEAR=$((2075 + RANDOM % 5))
  STATUS_LP=$(api_post "/governance/lock-periods" "{\"dataset_type\":\"DOCUMENT\",\"period_year\":$MALFORM_LP_YEAR,\"lock_reason\":\"malform test\"}")
  if [[ "$STATUS_LP" == "201" ]]; then
    LP_ID=$(field "['data']['id']")
    STATUS_MAL2=$(api_post_raw "/governance/lock-periods/$LP_ID/lock" '{"bad": json is here')
    if [[ "$STATUS_MAL2" == "400" ]]; then
      pass "malformed JSON on lock-period/lock -> 400"
    else
      fail "malformed JSON on lock-period/lock returned $STATUS_MAL2 (expected 400)"
    fi
  else
    skip "Skipping lock-period malformed JSON test (create lock-period returned $STATUS_LP)"
  fi
else
  fail "create submission for malformed JSON test returned $STATUS (expected 201)"
fi

# ---------------------------------------------------------------------------
# 11. submitted_by = NULL in DRAFT (DB check)
# ---------------------------------------------------------------------------
info "11. submitted_by = NULL in DRAFT"
STATUS=$(api_post "/governance/submissions" '{
  "dataset_type": "ISSUE",
  "source_type": "MANUAL",
  "period_year": 2026,
  "items": [{"entity_type": "project", "action": "VALIDATE_ONLY", "payload_after": {"x": 1}}]
}')
if [[ "$STATUS" == "201" ]]; then
  DRAFT_ID=$(field "['data']['id']")
  pass "created DRAFT submission $DRAFT_ID"
  # Check submitted_by in DB
  SUBMITTED_BY=$(dbq "SELECT submitted_by FROM data_submissions WHERE id='$DRAFT_ID'" | tr -d '[:space:]')
  if [[ -z "$SUBMITTED_BY" ]]; then
    pass "submitted_by IS NULL in DRAFT"
  else
    fail "submitted_by should be NULL in DRAFT, got: '$SUBMITTED_BY'"
  fi
else
  fail "create DRAFT submission returned $STATUS"
fi

# ---------------------------------------------------------------------------
# 12. created_by diset di DRAFT (DB check)
# ---------------------------------------------------------------------------
info "12. created_by set in DRAFT"
STATUS=$(api_post "/governance/submissions" '{
  "dataset_type": "CONTRACT",
  "source_type": "MANUAL",
  "period_year": 2026,
  "items": [{"entity_type": "project", "action": "VALIDATE_ONLY", "payload_after": {"x": 1}}]
}')
if [[ "$STATUS" == "201" ]]; then
  DRAFT2_ID=$(field "['data']['id']")
  CREATED_BY=$(dbq "SELECT created_by FROM data_submissions WHERE id='$DRAFT2_ID'" | tr -d '[:space:]')
  if [[ -n "$CREATED_BY" ]]; then
    pass "created_by IS SET in DRAFT: $CREATED_BY"
  else
    fail "created_by should be set in DRAFT, got NULL"
  fi
else
  fail "create DRAFT for created_by check returned $STATUS"
fi

# ---------------------------------------------------------------------------
# 13. UPDATE/DELETE/UPSERT require entity_id; CREATE/VALIDATE_ONLY may omit it
# ---------------------------------------------------------------------------
info "13. item action entity_id rules"
ERULE_YEAR=$((2060 + RANDOM % 10))

# 13a. UPDATE without entity_id -> 400
STATUS=$(api_post "/governance/submissions" "{
  \"dataset_type\": \"DOCUMENT\",
  \"source_type\": \"MANUAL\",
  \"period_year\": $ERULE_YEAR,
  \"items\": [{\"entity_type\": \"project\", \"action\": \"UPDATE\", \"payload_after\": {\"x\": 1}}]
}")
if [[ "$STATUS" == "400" ]]; then pass "UPDATE without entity_id -> 400"; else fail "UPDATE without entity_id returned $STATUS (expected 400)"; fi

# 13b. DELETE without entity_id -> 400
STATUS=$(api_post "/governance/submissions" "{
  \"dataset_type\": \"DOCUMENT\",
  \"source_type\": \"MANUAL\",
  \"period_year\": $ERULE_YEAR,
  \"items\": [{\"entity_type\": \"project\", \"action\": \"DELETE\", \"payload_after\": {} }]
}")
if [[ "$STATUS" == "400" ]]; then pass "DELETE without entity_id -> 400"; else fail "DELETE without entity_id returned $STATUS (expected 400)"; fi

# 13c. UPSERT without entity_id -> 400
STATUS=$(api_post "/governance/submissions" "{
  \"dataset_type\": \"DOCUMENT\",
  \"source_type\": \"MANUAL\",
  \"period_year\": $ERULE_YEAR,
  \"items\": [{\"entity_type\": \"project\", \"action\": \"UPSERT\", \"payload_after\": {\"x\": 1}}]
}")
if [[ "$STATUS" == "400" ]]; then pass "UPSERT without entity_id -> 400"; else fail "UPSERT without entity_id returned $STATUS (expected 400)"; fi

# 13d. CREATE without entity_id -> allowed (201)
STATUS=$(api_post "/governance/submissions" "{
  \"dataset_type\": \"DOCUMENT\",
  \"source_type\": \"MANUAL\",
  \"period_year\": $ERULE_YEAR,
  \"items\": [{\"entity_type\": \"project\", \"action\": \"CREATE\", \"payload_after\": {\"x\": 1}}]
}")
if [[ "$STATUS" == "201" ]]; then pass "CREATE without entity_id -> 201"; else fail "CREATE without entity_id returned $STATUS (expected 201)"; fi

# 13e. VALIDATE_ONLY without entity_id (payload-only) -> allowed (201)
STATUS=$(api_post "/governance/submissions" "{
  \"dataset_type\": \"DOCUMENT\",
  \"source_type\": \"MANUAL\",
  \"period_year\": $ERULE_YEAR,
  \"items\": [{\"entity_type\": \"project\", \"action\": \"VALIDATE_ONLY\", \"payload_after\": {\"progress_pct\": 60}}]
}")
if [[ "$STATUS" == "201" ]]; then pass "VALIDATE_ONLY without entity_id -> 201"; else fail "VALIDATE_ONLY without entity_id returned $STATUS (expected 201)"; fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
TOTAL=$((PASS_COUNT + FAILURES))
if [[ "$FAILURES" -eq 0 ]]; then
  echo -e "${GREEN}============================================${RESET}"
  echo -e "${GREEN} ALL HARDENING SMOKE CHECKS PASSED ✓${RESET}"
  echo -e "${GREEN} $PASS_COUNT/$TOTAL checks passed${RESET}"
  echo -e "${GREEN}============================================${RESET}"
  exit 0
else
  echo -e "${RED}============================================${RESET}"
  echo -e "${RED} $FAILURES HARDENING SMOKE CHECK(S) FAILED ✗${RESET}"
  echo -e "${RED} $PASS_COUNT/$TOTAL checks passed${RESET}"
  echo -e "${RED}============================================${RESET}"
  exit 1
fi
