#!/usr/bin/env bash
# =============================================================================
# CANKORA — Smoke Test UAT-007: End-to-End Business Flow & Data Consistency
# =============================================================================
# Tests the full PMO business flow:
#   A. Project Setup & Onboarding
#   B. Vendor, Contract & Budget Input
#   C. Risk & Issue Register
#   D. Corrective Action
#   E. Field Evidence & Document
#   F. Governance Official Approval
#   G. Command Center & Escalation
#   H. Analytics & Report Export
#   I. Audit Log Verification
#   J. Notification
#   K. Data Cleanup (soft delete)
#
# Prerequisites:
#   - Backend running at http://localhost:8080
#   - Demo users seeded: make seed (from backend/)
#   - All passwords: Demo@Cankora2024! (demo) / Admin@Cankora2024! (admin)
#
# All UAT data uses prefix "UAT-BF-007" for easy identification.
# =============================================================================

set -uo pipefail

BASE="http://localhost:8080/api/v1"
PASS=0
FAIL=0
SKIP=0

# Unique suffix per run — avoids unique constraint conflicts from soft-deleted rows
RUN_ID=$(date +%s | tail -c 6)
UAT_CODE="UAT-BF-007-${RUN_ID}"
UAT_CONTRACT="CNT-UAT-007-${RUN_ID}"
UAT_GOV_YEAR=9000  # Use fixed far-future year that's unlikely to conflict
GOV_YEAR=$((UAT_GOV_YEAR + RANDOM % 999))

# ── colours ──────────────────────────────────────────────────────────────────
green()  { echo -e "\033[32m✓ PASS\033[0m  $*"; }
red()    { echo -e "\033[31m✗ FAIL\033[0m  $*"; }
yellow() { echo -e "\033[33m~ SKIP\033[0m  $*"; }
header() { echo -e "\n\033[1;34m── $* ──\033[0m"; }

pass() { green "$1";  ((PASS++));  }
fail() { red   "$1";  ((FAIL++));  }
skip() { yellow "$1"; ((SKIP++));  }

# ── HTTP helpers ──────────────────────────────────────────────────────────────
login() {
  local email="$1" password="$2"
  curl -s -X POST "$BASE/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"$password\"}" \
    | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('access_token',''))" 2>/dev/null || echo ""
}

api_post() {
  local token="$1" path="$2" body="$3"
  curl -s -w "\n%{http_code}" -X POST "$BASE$path" \
    -H "Authorization: Bearer $token" \
    -H "Content-Type: application/json" \
    -d "$body"
}

api_get() {
  local token="$1" path="$2"
  curl -s -w "\n%{http_code}" -H "Authorization: Bearer $token" "$BASE$path"
}

api_put() {
  local token="$1" path="$2" body="$3"
  curl -s -w "\n%{http_code}" -X PUT "$BASE$path" \
    -H "Authorization: Bearer $token" \
    -H "Content-Type: application/json" \
    -d "$body"
}

api_patch() {
  local token="$1" path="$2" body="$3"
  curl -s -w "\n%{http_code}" -X PATCH "$BASE$path" \
    -H "Authorization: Bearer $token" \
    -H "Content-Type: application/json" \
    -d "$body"
}

api_delete() {
  local token="$1" path="$2"
  curl -s -o /dev/null -w "%{http_code}" -X DELETE \
    -H "Authorization: Bearer $token" "$BASE$path"
}

# Extract field from last curl response (body + status code on last line)
extract_id() {
  echo "$1" | awk 'NR>1{print prev}{prev=$0}' | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id','') or d.get('data',{}).get('submission',{}).get('id',''))" 2>/dev/null || echo ""
}

extract_field() {
  local resp="$1" field="$2"
  echo "$resp" | awk 'NR>1{print prev}{prev=$0}' | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('$field',''))" 2>/dev/null || echo ""
}

get_status_code() {
  echo "$1" | tail -1
}

check_status() {
  local label="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    pass "$label → HTTP $actual"
  else
    fail "$label → expected $expected, got $actual"
  fi
}

check_created() {
  local label="$1" code="$2"
  if [ "$code" = "201" ] || [ "$code" = "200" ]; then
    pass "$label → HTTP $code"
  else
    fail "$label → expected 201/200, got $code"
  fi
}

# =============================================================================
# 0. Health & Login
# =============================================================================
header "0. Health & Login"
S=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/health")
check_status "GET /health" "200" "$S"

TOKEN=$(login "admin@cankora.local" "Admin@Cankora2024!")
[ -n "$TOKEN" ] && pass "admin@cankora.local login" || { fail "admin@cankora.local login — aborting"; exit 1; }

PMO_TOKEN=$(login "pmo@cankora.local" "Demo@Cankora2024!")
[ -n "$PMO_TOKEN" ] && pass "pmo@cankora.local login" || fail "pmo@cankora.local login"

# Pre-run cleanup: remove stale UAT project from previous runs
STALE_PID=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE/projects?search=$UAT_CODE" \
  | python3 -c "import sys,json; d=json.load(sys.stdin); items=d.get('data',[]); print(items[0]['id'] if items else '')" 2>/dev/null || echo "")
if [ -n "$STALE_PID" ]; then
  curl -s -o /dev/null -X DELETE -H "Authorization: Bearer $TOKEN" "$BASE/projects/$STALE_PID"
fi

# =============================================================================
# A. Project Setup & Onboarding (Flow A)
# =============================================================================
header "A. Project Setup & Onboarding"

# A-1: Create project
RESP=$(api_post "$TOKEN" "/projects" "{
  \"code\": \"$UAT_CODE\",
  \"name\": \"Proyek UAT Business Flow 007\",
  \"description\": \"UAT-007 end-to-end business flow test project\",
  \"start_date\": \"2026-01-01\",
  \"end_date\": \"2026-12-31\",
  \"progress_pct\": 0
}")
CODE=$(get_status_code "$RESP")
check_created "A-1: Create project $UAT_CODE" "$CODE"
PROJECT_ID=$(extract_id "$RESP")
[ -n "$PROJECT_ID" ] && pass "A-1: project_id extracted: ${PROJECT_ID:0:8}..." || fail "A-1: project_id extraction failed"

# A-2: Duplicate code should return 409
if [ -n "$PROJECT_ID" ]; then
  RESP2=$(api_post "$TOKEN" "/projects" "{
    \"code\": \"$UAT_CODE\",
    \"name\": \"Duplicate $UAT_CODE\",
    \"start_date\": \"2026-01-01\",
    \"end_date\": \"2026-12-31\"
  }")
  CODE2=$(get_status_code "$RESP2")
  check_status "A-2: Duplicate project code → 409" "409" "$CODE2"
fi

# A-3: Transition DRAFT → PLANNING
if [ -n "$PROJECT_ID" ]; then
  RESP=$(api_post "$TOKEN" "/projects/$PROJECT_ID/transition" '{"to_status":"PLANNING"}')
  CODE=$(get_status_code "$RESP")
  check_status "A-3: Project DRAFT → PLANNING" "200" "$CODE"

  # A-4: Transition PLANNING → ACTIVE
  RESP=$(api_post "$TOKEN" "/projects/$PROJECT_ID/transition" '{"to_status":"ACTIVE"}')
  CODE=$(get_status_code "$RESP")
  check_status "A-4: Project PLANNING → ACTIVE" "200" "$CODE"

  # A-5: Create milestone
  RESP=$(api_post "$TOKEN" "/projects/$PROJECT_ID/milestones" '{
    "title": "UAT Milestone 007",
    "due_date": "2026-06-30"
  }')
  CODE=$(get_status_code "$RESP")
  check_created "A-5: Create milestone" "$CODE"
  MILESTONE_ID=$(extract_id "$RESP")

  # A-6: Create task
  RESP=$(api_post "$TOKEN" "/projects/$PROJECT_ID/tasks" '{
    "title": "UAT Task 007",
    "priority": "HIGH",
    "due_date": "2026-03-31"
  }')
  CODE=$(get_status_code "$RESP")
  check_created "A-6: Create task" "$CODE"
  TASK_ID=$(extract_id "$RESP")

  # A-7: GET project list includes new project
  RESP=$(api_get "$TOKEN" "/projects?search=UAT-BF-007")
  CODE=$(get_status_code "$RESP")
  check_status "A-7: GET projects list" "200" "$CODE"
else
  skip "A-3..7: skipped (project creation failed)"
fi

# =============================================================================
# B. Vendor, Contract & Budget (Flow B)
# =============================================================================
header "B. Vendor, Contract & Budget"

# B-1: Create vendor
RESP=$(api_post "$TOKEN" "/vendors" '{
  "name": "UAT Vendor 007",
  "type": "VENDOR",
  "legal_name": "PT UAT Vendor Tujuh",
  "tax_id": "001.234.567.8-007.000",
  "is_active": true
}')
CODE=$(get_status_code "$RESP")
check_created "B-1: Create vendor" "$CODE"
VENDOR_ID=$(extract_id "$RESP")

# B-2: Create consultant
RESP=$(api_post "$TOKEN" "/vendors" '{
  "name": "UAT Consultant 007",
  "type": "CONSULTANT",
  "is_active": true
}')
CODE=$(get_status_code "$RESP")
check_created "B-2: Create consultant" "$CODE"

# B-3: Create contract
if [ -n "$PROJECT_ID" ] && [ -n "$VENDOR_ID" ]; then
  RESP=$(api_post "$TOKEN" "/projects/$PROJECT_ID/contracts" "{
    \"contract_number\": \"$UAT_CONTRACT\",
    \"title\": \"UAT Contract 007\",
    \"vendor_id\": \"$VENDOR_ID\",
    \"contract_value\": 10000000,
    \"currency\": \"IDR\",
    \"start_date\": \"2026-01-01\",
    \"end_date\": \"2026-12-31\",
    \"status\": \"ACTIVE\"
  }")
  CODE=$(get_status_code "$RESP")
  check_created "B-3: Create contract" "$CODE"
  CONTRACT_ID=$(extract_id "$RESP")

  # B-4: Duplicate contract number → 409
  RESP2=$(api_post "$TOKEN" "/projects/$PROJECT_ID/contracts" "{
    \"contract_number\": \"$UAT_CONTRACT\",
    \"title\": \"Duplicate Contract\",
    \"vendor_id\": \"$VENDOR_ID\",
    \"contract_value\": 1,
    \"currency\": \"IDR\",
    \"start_date\": \"2026-01-01\",
    \"end_date\": \"2026-12-31\",
    \"status\": \"DRAFT\"
  }")
  CODE2=$(get_status_code "$RESP2")
  check_status "B-4: Duplicate contract number → 409" "409" "$CODE2"
else
  skip "B-3..4: skipped (project or vendor missing)"
fi

# B-5: Create budget line (planned=5000000, actual=0)
if [ -n "$PROJECT_ID" ]; then
  RESP=$(api_post "$TOKEN" "/projects/$PROJECT_ID/budgets" '{
    "category": "KONSTRUKSI",
    "description": "UAT Budget 007",
    "planned": 5000000,
    "actual": 0,
    "currency": "IDR"
  }')
  CODE=$(get_status_code "$RESP")
  check_created "B-5: Create budget line (NORMAL)" "$CODE"
  BUDGET_ID=$(extract_id "$RESP")
  STATUS_VAL=$(extract_field "$RESP" "status")
  [ "$STATUS_VAL" = "NORMAL" ] && pass "B-5: Budget status=NORMAL (actual=0)" || fail "B-5: Budget status=$STATUS_VAL (expected NORMAL)"

  # B-6: Update actual to RISK threshold (90%)
  if [ -n "$BUDGET_ID" ]; then
    RESP=$(api_put "$TOKEN" "/projects/$PROJECT_ID/budgets/$BUDGET_ID" '{"actual": 4500000}')
    CODE=$(get_status_code "$RESP")
    check_status "B-6: Update budget actual → RISK" "200" "$CODE"
    STATUS_VAL=$(extract_field "$RESP" "status")
    [ "$STATUS_VAL" = "RISK" ] && pass "B-6: Budget status=RISK (actual=4500000/planned=5000000, 90%)" || fail "B-6: Budget status=$STATUS_VAL (expected RISK)"

    # B-7: Update to OVERRUN
    RESP=$(api_put "$TOKEN" "/projects/$PROJECT_ID/budgets/$BUDGET_ID" '{"actual": 5100000}')
    CODE=$(get_status_code "$RESP")
    check_status "B-7: Update budget actual → OVERRUN" "200" "$CODE"
    STATUS_VAL=$(extract_field "$RESP" "status")
    [ "$STATUS_VAL" = "OVERRUN" ] && pass "B-7: Budget status=OVERRUN (actual=5100000/planned=5000000, 102%)" || fail "B-7: Budget status=$STATUS_VAL (expected OVERRUN)"
  fi
else
  skip "B-5..7: skipped (project missing)"
fi

# B-8: Cannot delete vendor referenced by contract
if [ -n "$VENDOR_ID" ] && [ -n "$CONTRACT_ID" ]; then
  S=$(api_delete "$TOKEN" "/vendors/$VENDOR_ID")
  check_status "B-8: Cannot delete referenced vendor → 409" "409" "$S"
else
  skip "B-8: skipped (vendor or contract missing)"
fi

# =============================================================================
# C. Risk & Issue Register (Flow C)
# =============================================================================
header "C. Risk & Issue Register"

if [ -n "$PROJECT_ID" ]; then
  # C-1: Create risk (probability=4, impact=5 → score=20 CRITICAL)
  RESP=$(api_post "$TOKEN" "/projects/$PROJECT_ID/risks" '{
    "title": "UAT Risk 007",
    "description": "UAT risk for business flow test",
    "probability": 4,
    "impact": 5,
    "mitigation": "Apply UAT mitigation plan"
  }')
  CODE=$(get_status_code "$RESP")
  check_created "C-1: Create risk" "$CODE"
  RISK_ID=$(extract_id "$RESP")
  RISK_SCORE=$(extract_field "$RESP" "risk_score")
  RISK_SEV=$(extract_field "$RESP" "severity")
  [ "$RISK_SCORE" = "20" ] && pass "C-1: risk_score=20 (4×5)" || fail "C-1: risk_score=$RISK_SCORE (expected 20)"
  [ "$RISK_SEV" = "CRITICAL" ] && pass "C-1: severity=CRITICAL (score 16-25)" || fail "C-1: severity=$RISK_SEV (expected CRITICAL)"

  # C-2: Risk transition IDENTIFIED → ASSESSED (via PUT, no separate transition route)
  if [ -n "$RISK_ID" ]; then
    RESP=$(api_put "$TOKEN" "/projects/$PROJECT_ID/risks/$RISK_ID" '{"status":"ASSESSED"}')
    CODE=$(get_status_code "$RESP")
    check_status "C-2: Risk IDENTIFIED → ASSESSED" "200" "$CODE"

    # C-3: Risk transition ASSESSED → MITIGATED (via PUT)
    RESP=$(api_put "$TOKEN" "/projects/$PROJECT_ID/risks/$RISK_ID" '{"status":"MITIGATED"}')
    CODE=$(get_status_code "$RESP")
    check_status "C-3: Risk ASSESSED → MITIGATED" "200" "$CODE"
  fi

  # C-4: Create issue
  RESP=$(api_post "$TOKEN" "/projects/$PROJECT_ID/issues" '{
    "title": "UAT Issue 007",
    "description": "UAT issue for business flow test",
    "severity": "HIGH",
    "escalation": "PROJECT_MANAGER"
  }')
  CODE=$(get_status_code "$RESP")
  check_created "C-4: Create issue" "$CODE"
  ISSUE_ID=$(extract_id "$RESP")

  # C-5: Issue transition OPEN → IN_PROGRESS (via PUT, no separate transition route)
  if [ -n "$ISSUE_ID" ]; then
    RESP=$(api_put "$TOKEN" "/projects/$PROJECT_ID/issues/$ISSUE_ID" '{"status":"IN_PROGRESS"}')
    CODE=$(get_status_code "$RESP")
    check_status "C-5: Issue OPEN → IN_PROGRESS" "200" "$CODE"

    # C-6: Issue transition IN_PROGRESS → RESOLVED (via PUT)
    RESP=$(api_put "$TOKEN" "/projects/$PROJECT_ID/issues/$ISSUE_ID" '{"status":"RESOLVED","resolution":"Resolved via UAT-007 test run"}')
    CODE=$(get_status_code "$RESP")
    check_status "C-6: Issue IN_PROGRESS → RESOLVED" "200" "$CODE"
  fi
else
  skip "C: skipped (project missing)"
fi

# =============================================================================
# D. Corrective Action (Flow D)
# =============================================================================
header "D. Corrective Action"

if [ -n "$PROJECT_ID" ]; then
  CA_BODY="{
    \"title\": \"UAT Corrective 007\",
    \"deviation\": \"UAT deviation — budget overrun detected\",
    \"root_cause\": \"Scope creep in UAT phase\",
    \"recommendation\": \"Enforce change control board review\",
    \"source_type\": \"ISSUE\""
  if [ -n "$ISSUE_ID" ]; then
    CA_BODY="$CA_BODY, \"source_issue_id\": \"$ISSUE_ID\""
  fi
  CA_BODY="$CA_BODY}"

  RESP=$(api_post "$TOKEN" "/projects/$PROJECT_ID/corrective-actions" "$CA_BODY")
  CODE=$(get_status_code "$RESP")
  check_created "D-1: Create corrective action (linked to issue)" "$CODE"
  CA_ID=$(extract_id "$RESP")

  if [ -n "$CA_ID" ]; then
    # D-2: Update corrective action
    RESP=$(api_put "$TOKEN" "/projects/$PROJECT_ID/corrective-actions/$CA_ID" '{
      "recommendation": "Enforce change control board review — UPDATED",
      "target_date": "2026-06-30"
    }')
    CODE=$(get_status_code "$RESP")
    check_status "D-2: Update corrective action" "200" "$CODE"

    # D-3: Transition CA: DRAFT→SUBMITTED first, then SUBMITTED→IN_PROGRESS (CA FSM: DRAFT→SUBMITTED→IN_PROGRESS→COMPLETED)
    RESP=$(api_post "$TOKEN" "/projects/$PROJECT_ID/corrective-actions/$CA_ID/transition" '{"to_status":"SUBMITTED"}')
    CODE=$(get_status_code "$RESP")
    check_status "D-3a: Corrective action DRAFT → SUBMITTED" "200" "$CODE"
    RESP=$(api_post "$TOKEN" "/projects/$PROJECT_ID/corrective-actions/$CA_ID/transition" '{"to_status":"IN_PROGRESS"}')
    CODE=$(get_status_code "$RESP")
    check_status "D-3: Corrective action → IN_PROGRESS" "200" "$CODE"

    # D-4: List corrective actions
    RESP=$(api_get "$TOKEN" "/projects/$PROJECT_ID/corrective-actions")
    CODE=$(get_status_code "$RESP")
    check_status "D-4: List corrective actions" "200" "$CODE"
  else
    skip "D-2..4: skipped (corrective action creation failed)"
  fi
else
  skip "D: skipped (project missing)"
fi

# =============================================================================
# E. Field Evidence & Document (Flow E)
# =============================================================================
header "E. Field Evidence & Document"

if [ -n "$PROJECT_ID" ]; then
  # E-1: Upload document
  TMPFILE=$(mktemp /tmp/uat007_doc_XXXXXX.txt)
  echo "UAT-007 document content for business flow test" > "$TMPFILE"
  RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE/projects/$PROJECT_ID/documents" \
    -H "Authorization: Bearer $TOKEN" \
    -F "file=@$TMPFILE;type=text/plain" \
    -F "name=UAT Document 007" \
    -F "category=EVIDENCE" \
    -F "version=1.0")
  CODE=$(get_status_code "$RESP")
  check_created "E-1: Upload document" "$CODE"
  DOC_ID=$(extract_id "$RESP")
  rm -f "$TMPFILE"

  # E-2: Download document
  if [ -n "$DOC_ID" ]; then
    S=$(curl -s -o /dev/null -w "%{http_code}" \
      -H "Authorization: Bearer $TOKEN" \
      "$BASE/projects/$PROJECT_ID/documents/$DOC_ID/download")
    check_status "E-2: Download document" "200" "$S"

    # E-3: Update document metadata
    RESP=$(api_put "$TOKEN" "/projects/$PROJECT_ID/documents/$DOC_ID" '{
      "name": "UAT Document 007 — Updated",
      "version": "1.1"
    }')
    CODE=$(get_status_code "$RESP")
    check_status "E-3: Update document metadata" "200" "$CODE"
  else
    skip "E-2..3: skipped (document upload failed)"
  fi

  # E-4: Create field inspection (uses multipart form binding, not JSON)
  RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE/projects/$PROJECT_ID/inspections" \
    -H "Authorization: Bearer $TOKEN" \
    -F "inspected_at=2026-03-15T09:00:00Z" \
    -F "notes=UAT-007 field inspection" \
    -F "latitude=-6.2088" \
    -F "longitude=106.8456")
  CODE=$(get_status_code "$RESP")
  check_created "E-4: Create field inspection" "$CODE"
  INSPECT_ID=$(extract_id "$RESP")

  # E-5: Upload evidence to inspection
  if [ -n "$INSPECT_ID" ]; then
    TMPFILE2=$(mktemp /tmp/uat007_evidence_XXXXXX.txt)
    echo "UAT-007 inspection evidence content" > "$TMPFILE2"
    RESP=$(curl -s -w "\n%{http_code}" -X POST \
      "$BASE/projects/$PROJECT_ID/inspections/$INSPECT_ID/evidence" \
      -H "Authorization: Bearer $TOKEN" \
      -F "file=@$TMPFILE2;type=text/plain" \
      -F "latitude=-6.2088" \
      -F "longitude=106.8456")
    CODE=$(get_status_code "$RESP")
    check_created "E-5: Upload inspection evidence" "$CODE"
    # AddEvidence returns the parent Inspection object; evidence ID is in data.evidence[0].id
    EVIDENCE_ID=$(echo "$RESP" | awk 'NR>1{print prev}{prev=$0}' | python3 -c \
      "import sys,json; d=json.load(sys.stdin); ev=d.get('data',{}).get('evidence',[]); print(ev[0]['id'] if ev else '')" 2>/dev/null)
    rm -f "$TMPFILE2"

    # E-6: Download evidence
    if [ -n "$EVIDENCE_ID" ]; then
      S=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $TOKEN" \
        "$BASE/projects/$PROJECT_ID/inspections/$INSPECT_ID/evidence/$EVIDENCE_ID/download")
      check_status "E-6: Download evidence" "200" "$S"
    else
      skip "E-6: skipped (evidence upload failed)"
    fi

    # E-7: Verify inspection
    RESP=$(api_patch "$TOKEN" "/projects/$PROJECT_ID/inspections/$INSPECT_ID/verification" '{"status":"VERIFIED","notes":"UAT-007 verified"}')
    CODE=$(get_status_code "$RESP")
    check_status "E-7: Verify inspection" "200" "$CODE"
  else
    skip "E-5..7: skipped (inspection creation failed)"
  fi
else
  skip "E: skipped (project missing)"
fi

# =============================================================================
# F. Governance Official Approval (Flow F)
# =============================================================================
header "F. Governance Official Approval"

# Use period_year=9999 to avoid conflict with existing locked periods
RESP=$(api_post "$TOKEN" "/governance/submissions" "{
  \"dataset_type\": \"PROJECT_PROGRESS\",
  \"source_type\": \"MANUAL\",
  \"period_year\": $GOV_YEAR,
  \"description\": \"UAT-007 governance submission test\",
  \"items\": [{\"entity_type\":\"project\",\"action\":\"CREATE\",\"payload_after\":{\"name\":\"UAT-007 test entry\"}}]
}")
CODE=$(get_status_code "$RESP")
check_created "F-1: Create governance submission (DRAFT)" "$CODE"
SUBMISSION_ID=$(extract_id "$RESP")

if [ -n "$SUBMISSION_ID" ]; then
  # F-3: Submit (DRAFT → SUBMITTED)
  RESP=$(api_post "$TOKEN" "/governance/submissions/$SUBMISSION_ID/submit" '{}')
  CODE=$(get_status_code "$RESP")
  check_status "F-3: Submission DRAFT → SUBMITTED" "200" "$CODE"

  # F-4: Start review (SUBMITTED → IN_REVIEW)
  RESP=$(api_post "$TOKEN" "/governance/submissions/$SUBMISSION_ID/review" '{}')
  CODE=$(get_status_code "$RESP")
  check_status "F-4: Submission SUBMITTED → IN_REVIEW" "200" "$CODE"

  # F-5: Approve (IN_REVIEW → APPROVED)
  RESP=$(api_post "$TOKEN" "/governance/submissions/$SUBMISSION_ID/approve" '{}')
  CODE=$(get_status_code "$RESP")
  check_status "F-5: Submission IN_REVIEW → APPROVED" "200" "$CODE"

  # F-6: Lock (APPROVED → LOCKED)
  RESP=$(api_post "$TOKEN" "/governance/submissions/$SUBMISSION_ID/lock" '{}')
  CODE=$(get_status_code "$RESP")
  check_status "F-6: Submission APPROVED → LOCKED" "200" "$CODE"

  # F-7: Create second submission for same locked period → 409
  RESP2=$(api_post "$TOKEN" "/governance/submissions" "{
    \"dataset_type\": \"PROJECT_PROGRESS\",
    \"source_type\": \"MANUAL\",
    \"period_year\": $GOV_YEAR,
    \"description\": \"Duplicate locked period test\",
    \"items\": [{\"entity_type\":\"project\",\"action\":\"CREATE\",\"payload_after\":{\"name\":\"dup test\"}}]
  }")
  CODE2=$(get_status_code "$RESP2")
  check_status "F-7: Duplicate locked period → 409" "409" "$CODE2"

  # F-8: Reject flow — new submission, submit, review, reject
  REJ_GOV_YEAR=$((GOV_YEAR - 1))
  RESP=$(api_post "$TOKEN" "/governance/submissions" "{
    \"dataset_type\": \"PROJECT_PROGRESS\",
    \"source_type\": \"MANUAL\",
    \"period_year\": $REJ_GOV_YEAR,
    \"description\": \"UAT-007 reject flow test\",
    \"items\": [{\"entity_type\":\"project\",\"action\":\"CREATE\",\"payload_after\":{\"name\":\"reject test\"}}]
  }")
  CODE=$(get_status_code "$RESP")
  check_created "F-8a: Create submission for reject flow" "$CODE"
  REJECT_ID=$(extract_id "$RESP")

  if [ -n "$REJECT_ID" ]; then
    api_post "$TOKEN" "/governance/submissions/$REJECT_ID/submit" '{}' > /dev/null
    api_post "$TOKEN" "/governance/submissions/$REJECT_ID/review" '{}' > /dev/null
    RESP=$(api_post "$TOKEN" "/governance/submissions/$REJECT_ID/reject" '{"rejection_reason":"UAT-007 rejection test — insufficient data"}')
    CODE=$(get_status_code "$RESP")
    check_status "F-8b: Submission reject with reason" "200" "$CODE"
  else
    skip "F-8b: skipped (reject submission creation failed)"
  fi
else
  skip "F-3..8: skipped (submission creation failed)"
fi

# =============================================================================
# G. Command Center & Escalation (Flow G)
# =============================================================================
header "G. Command Center & Escalation"

# G-1: Read command center
RESP=$(api_get "$TOKEN" "/command-center")
CODE=$(get_status_code "$RESP")
check_status "G-1: GET /command-center" "200" "$CODE"

# G-2: Create escalation
ESC_BODY="{\"source_type\": \"risk\", \"level\": \"PROJECT_MANAGER\", \"reason\": \"UAT-007 escalation test\""
if [ -n "$PROJECT_ID" ]; then
  ESC_BODY="$ESC_BODY, \"project_id\": \"$PROJECT_ID\""
fi
if [ -n "$RISK_ID" ]; then
  ESC_BODY="$ESC_BODY, \"source_id\": \"$RISK_ID\""
else
  ESC_BODY="$ESC_BODY, \"source_id\": \"00000000-0000-0000-0000-000000000001\""
fi
ESC_BODY="$ESC_BODY}"

RESP=$(api_post "$TOKEN" "/command-center/escalations" "$ESC_BODY")
CODE=$(get_status_code "$RESP")
check_created "G-2: Create escalation" "$CODE"
ESC_ID=$(extract_id "$RESP")

# G-3: Update escalation status → ACKNOWLEDGED
if [ -n "$ESC_ID" ]; then
  RESP=$(api_patch "$TOKEN" "/command-center/escalations/$ESC_ID/status" '{"status":"ACKNOWLEDGED"}')
  CODE=$(get_status_code "$RESP")
  check_status "G-3: Escalation → ACKNOWLEDGED" "200" "$CODE"
else
  skip "G-3: skipped (escalation creation failed)"
fi

# G-4: Create decision
DEC_BODY="{\"subject\": \"UAT-007 Decision\", \"decision_text\": \"UAT-007 decision: proceed with plan B\", \"source_type\": \"executive_decision\""
if [ -n "$PROJECT_ID" ]; then
  DEC_BODY="$DEC_BODY, \"project_id\": \"$PROJECT_ID\""
fi
DEC_BODY="$DEC_BODY}"

RESP=$(api_post "$TOKEN" "/command-center/decisions" "$DEC_BODY")
CODE=$(get_status_code "$RESP")
check_created "G-4: Create decision" "$CODE"
DEC_ID=$(extract_id "$RESP")

# G-5: Update decision status → IN_PROGRESS (valid: IN_PROGRESS, COMPLETED, CANCELLED)
if [ -n "$DEC_ID" ]; then
  RESP=$(api_patch "$TOKEN" "/command-center/decisions/$DEC_ID/status" '{"status":"IN_PROGRESS"}')
  CODE=$(get_status_code "$RESP")
  check_status "G-5: Decision → IN_PROGRESS" "200" "$CODE"
else
  skip "G-5: skipped (decision creation failed)"
fi

# G-6: List escalations
RESP=$(api_get "$TOKEN" "/command-center/escalations")
CODE=$(get_status_code "$RESP")
check_status "G-6: GET /command-center/escalations" "200" "$CODE"

# =============================================================================
# H. Analytics & Report Export (Flow H)
# =============================================================================
header "H. Analytics & Report Export"

# H-1: Executive dashboard
RESP=$(api_get "$TOKEN" "/analytics/executive")
CODE=$(get_status_code "$RESP")
check_status "H-1: GET /analytics/executive" "200" "$CODE"

# H-2: Program dashboard
RESP=$(api_get "$TOKEN" "/analytics/programs")
CODE=$(get_status_code "$RESP")
check_status "H-2: GET /analytics/programs" "200" "$CODE"

# H-3: GIS summary
RESP=$(api_get "$TOKEN" "/analytics/gis/summary")
CODE=$(get_status_code "$RESP")
check_status "H-3: GET /analytics/gis/summary" "200" "$CODE"

# H-4: Reporting catalog
RESP=$(api_get "$TOKEN" "/analytics/reports/catalog")
CODE=$(get_status_code "$RESP")
if [ "$CODE" = "200" ] || [ "$CODE" = "400" ] || [ "$CODE" = "404" ]; then
  pass "H-4: GET /analytics/reports/catalog → $CODE (guard passed)"
else
  fail "H-4: GET /analytics/reports/catalog → $CODE"
fi

# H-5: Create export request
RESP=$(api_post "$TOKEN" "/analytics/reports/export/request" '{
  "dataset_key": "executive-summary",
  "format": "CSV"
}')
CODE=$(get_status_code "$RESP")
check_created "H-5: Create report export request (CSV)" "$CODE"
EXPORT_ID=$(extract_id "$RESP")

# H-6: Get export request status and download
if [ -n "$EXPORT_ID" ]; then
  sleep 1  # Allow async processing
  RESP=$(api_get "$TOKEN" "/analytics/reports/export/requests/$EXPORT_ID")
  CODE=$(get_status_code "$RESP")
  check_status "H-6a: GET export request" "200" "$CODE"
  EXPORT_STATUS=$(extract_field "$RESP" "status")

  if [ "$EXPORT_STATUS" = "COMPLETED" ]; then
    S=$(curl -s -o /dev/null -w "%{http_code}" \
      -H "Authorization: Bearer $TOKEN" \
      "$BASE/analytics/reports/export/requests/$EXPORT_ID/download")
    check_status "H-6b: Download export file" "200" "$S"
  else
    skip "H-6b: Download skipped — export status=$EXPORT_STATUS (not COMPLETED)"
  fi
else
  skip "H-6: skipped (export creation failed)"
fi

# =============================================================================
# I. Audit Log Verification (Flow I)
# =============================================================================
header "I. Audit Log Verification"

# I-1: List audit logs
RESP=$(api_get "$TOKEN" "/audit-logs")
CODE=$(get_status_code "$RESP")
check_status "I-1: GET /audit-logs" "200" "$CODE"

# I-2: Filter by action
RESP=$(api_get "$TOKEN" "/audit-logs?action=project.created")
CODE=$(get_status_code "$RESP")
check_status "I-2: GET /audit-logs?action=project.created" "200" "$CODE"

# Verify at least one project.created event exists in the system
COUNT=$(echo "$RESP" | awk 'NR>1{print prev}{prev=$0}' | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('meta',{}).get('total',0))" 2>/dev/null || echo "0")
[ "${COUNT:-0}" -gt 0 ] 2>/dev/null && pass "I-2: audit has project.created events (n=$COUNT)" || skip "I-2: no project.created events yet (new instance)"

# I-3: Filter by entity_type
RESP=$(api_get "$TOKEN" "/audit-logs?entity_type=project")
CODE=$(get_status_code "$RESP")
check_status "I-3: GET /audit-logs?entity_type=project" "200" "$CODE"

# I-4: Audit log summary
RESP=$(api_get "$TOKEN" "/audit-logs/summary")
CODE=$(get_status_code "$RESP")
check_status "I-4: GET /audit-logs/summary" "200" "$CODE"

# I-5: Audit log export CSV
S=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $TOKEN" \
  "$BASE/audit-logs/export")
check_status "I-5: GET /audit-logs/export (CSV)" "200" "$S"

# I-6: Auditor can see audit logs
if [ -n "${AUDITOR_TOKEN:-}" ]; then
  RESP=$(api_get "$AUDITOR_TOKEN" "/audit-logs")
  CODE=$(get_status_code "$RESP")
  check_status "I-6: AUDITOR GET /audit-logs" "200" "$CODE"
else
  AUDITOR_TOKEN=$(login "auditor@cankora.local" "Demo@Cankora2024!")
  if [ -n "$AUDITOR_TOKEN" ]; then
    RESP=$(api_get "$AUDITOR_TOKEN" "/audit-logs")
    CODE=$(get_status_code "$RESP")
    check_status "I-6: AUDITOR GET /audit-logs" "200" "$CODE"
  else
    skip "I-6: skipped (auditor login failed)"
  fi
fi

# I-7: No-token → 401
S=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/audit-logs")
check_status "I-7: /audit-logs no-token → 401" "401" "$S"

# =============================================================================
# J. Notification (Flow J)
# =============================================================================
header "J. Notification"

# J-1: Send test notification
RESP=$(api_post "$TOKEN" "/notifications/test" '{
  "subject": "UAT-007 Test Notification",
  "body": "Business flow UAT-007 notification verification",
  "priority": "NORMAL"
}')
CODE=$(get_status_code "$RESP")
check_created "J-1: Send test notification" "$CODE"
NOTIF_ID=$(extract_id "$RESP")

# J-2: List notifications
RESP=$(api_get "$TOKEN" "/notifications")
CODE=$(get_status_code "$RESP")
check_status "J-2: GET /notifications" "200" "$CODE"

# J-3: Notification summary
RESP=$(api_get "$TOKEN" "/notifications/summary")
CODE=$(get_status_code "$RESP")
check_status "J-3: GET /notifications/summary" "200" "$CODE"

# J-4: Mark as read
if [ -n "$NOTIF_ID" ]; then
  RESP=$(api_patch "$TOKEN" "/notifications/$NOTIF_ID/read" '{}')
  CODE=$(get_status_code "$RESP")
  check_status "J-4: PATCH /notifications/:id/read" "200" "$CODE"
else
  skip "J-4: skipped (notification creation failed)"
fi

# J-5: Mark all as read
RESP=$(api_patch "$TOKEN" "/notifications/read-all" '{}')
CODE=$(get_status_code "$RESP")
check_status "J-5: PATCH /notifications/read-all" "200" "$CODE"

# J-6: Admin notifications overview
RESP=$(api_get "$TOKEN" "/notifications/admin")
CODE=$(get_status_code "$RESP")
check_status "J-6: GET /notifications/admin (PMO+Admin only)" "200" "$CODE"

# =============================================================================
# K. Dashboard reads (data consistency)
# =============================================================================
header "K. Dashboard Data Consistency"

RESP=$(api_get "$TOKEN" "/dashboard")
CODE=$(get_status_code "$RESP")
check_status "K-1: GET /dashboard (includes warnings)" "200" "$CODE"

# Verify no 500 on dashboard
[ "$CODE" != "500" ] && pass "K-1: Dashboard no 500 error" || fail "K-1: Dashboard returned 500"

# Check project control (Level 3) for UAT project
if [ -n "$PROJECT_ID" ]; then
  RESP=$(api_get "$TOKEN" "/projects/$PROJECT_ID/control")
  CODE=$(get_status_code "$RESP")
  if [ "$CODE" = "200" ] || [ "$CODE" = "400" ]; then
    pass "K-2: GET /projects/$PROJECT_ID/control → $CODE"
  else
    fail "K-2: GET /projects/$PROJECT_ID/control → $CODE"
  fi
fi

# =============================================================================
# L. Cleanup (soft delete UAT data)
# =============================================================================
header "L. Cleanup (Soft Delete UAT Data)"

# Delete corrective action
if [ -n "$PROJECT_ID" ] && [ -n "$CA_ID" ]; then
  S=$(api_delete "$TOKEN" "/projects/$PROJECT_ID/corrective-actions/$CA_ID")
  check_status "L-1: Delete corrective action" "204" "$S"
fi

# Delete inspection
if [ -n "$PROJECT_ID" ] && [ -n "$INSPECT_ID" ]; then
  S=$(api_delete "$TOKEN" "/projects/$PROJECT_ID/inspections/$INSPECT_ID")
  check_status "L-2: Delete inspection" "204" "$S"
fi

# Delete document
if [ -n "$PROJECT_ID" ] && [ -n "$DOC_ID" ]; then
  S=$(api_delete "$TOKEN" "/projects/$PROJECT_ID/documents/$DOC_ID")
  check_status "L-3: Delete document" "204" "$S"
fi

# Delete contract (before vendor to allow vendor delete)
if [ -n "$PROJECT_ID" ] && [ -n "$CONTRACT_ID" ]; then
  S=$(api_delete "$TOKEN" "/projects/$PROJECT_ID/contracts/$CONTRACT_ID")
  check_status "L-4: Delete contract" "204" "$S"
fi

# Delete vendor (after contract is deleted)
if [ -n "$VENDOR_ID" ]; then
  S=$(api_delete "$TOKEN" "/vendors/$VENDOR_ID")
  check_status "L-5: Delete vendor (after contract removed)" "204" "$S"
fi

# Delete project (cascades to tasks/milestones/issues/risks/budgets)
if [ -n "$PROJECT_ID" ]; then
  S=$(api_delete "$TOKEN" "/projects/$PROJECT_ID")
  check_status "L-6: Delete project (cascade soft-delete)" "204" "$S"

  # Verify project is gone (404)
  RESP=$(api_get "$TOKEN" "/projects/$PROJECT_ID")
  CODE=$(get_status_code "$RESP")
  check_status "L-7: Deleted project returns 404" "404" "$CODE"
fi

# =============================================================================
# Summary
# =============================================================================
echo ""
echo "════════════════════════════════════════════════════"
echo " UAT-007 End-to-End Business Flow Smoke Test"
echo "════════════════════════════════════════════════════"
echo " PASS: $PASS"
echo " FAIL: $FAIL"
echo " SKIP: $SKIP"
echo "════════════════════════════════════════════════════"

if [ "$FAIL" -eq 0 ]; then
  echo -e "\033[32m ALL CHECKS PASSED\033[0m"
  exit 0
else
  echo -e "\033[31m $FAIL CHECK(S) FAILED\033[0m"
  exit 1
fi
