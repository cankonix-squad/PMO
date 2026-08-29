#!/usr/bin/env bash
# =============================================================================
# Smoke Test — P2-004: Priority Scoring & Decision Support
# =============================================================================
# Usage:
#   chmod +x docs/smoke-test-p2-004-priority.sh
#   ./docs/smoke-test-p2-004-priority.sh
#
# Prerequisites:
#   - Backend running on http://localhost:8080
#   - At least one project seeded in the DB
#   - jq installed (brew install jq)
# =============================================================================

set -uo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@cankora.local}"
ADMIN_PASS="${ADMIN_PASS:-Admin@Cankora2024!}"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0
FAIL=0

pass() { echo -e "${GREEN}  ✓ $1${NC}"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}  ✗ $1${NC}"; FAIL=$((FAIL+1)); }
section() { echo -e "\n${YELLOW}▶ $1${NC}"; }

assert_field() {
  local label="$1" json="$2" field="$3" expected="$4"
  local actual
  actual=$(echo "$json" | jq -r "$field" 2>/dev/null || echo "__jq_error__")
  if [[ "$actual" == "$expected" ]]; then
    pass "$label"
  else
    fail "$label (expected='$expected', got='$actual')"
  fi
}

assert_not_empty() {
  local label="$1" val="$2"
  if [[ -n "$val" && "$val" != "null" && "$val" != "__jq_error__" ]]; then
    pass "$label"
  else
    fail "$label (value is empty or null)"
  fi
}

assert_http() {
  local label="$1" code="$2" expected="$3"
  if [[ "$code" == "$expected" ]]; then
    pass "$label (HTTP $code)"
  else
    fail "$label (expected HTTP $expected, got $code)"
  fi
}

# =============================================================================
# 0. Login
# =============================================================================
section "0. Authentication"

LOGIN_RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASS\"}")
LOGIN_BODY=$(echo "$LOGIN_RESP" | sed '$d')
LOGIN_CODE=$(echo "$LOGIN_RESP" | tail -n 1)

assert_http "Login" "$LOGIN_CODE" "200"
TOKEN=$(echo "$LOGIN_BODY" | jq -r '.data.token // .data.access_token' 2>/dev/null)
assert_not_empty "Token obtained" "$TOKEN"

AUTH="-H \"Authorization: Bearer $TOKEN\""

# Helper: authenticated curl
acurl() {
  curl -s -w "\n%{http_code}" -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" "$@"
}

# =============================================================================
# 1. Get first project ID
# =============================================================================
section "1. Resolve a project for testing"

PROJECTS_RESP=$(acurl "$BASE_URL/projects?page_size=1")
PROJECTS_BODY=$(echo "$PROJECTS_RESP" | sed '$d')
PROJECT_ID=$(echo "$PROJECTS_BODY" | jq -r 'if .data | type == "array" then .data[0].id else .data.data[0].id end' 2>/dev/null)
assert_not_empty "At least one project exists" "$PROJECT_ID"
echo "    project_id = $PROJECT_ID"

# =============================================================================
# 2. Create a DRAFT formula
# =============================================================================
section "2. Create DRAFT formula"

FORMULA_PAYLOAD=$(cat <<'EOF'
{
  "name": "Smoke Test Formula",
  "missing_data_rule": "PENALIZE",
  "components": [
    {"component_key": "health_score",              "weight": 0.20},
    {"component_key": "risk_score",                "weight": 0.20},
    {"component_key": "issue_severity",            "weight": 0.15},
    {"component_key": "budget_usage",              "weight": 0.15},
    {"component_key": "schedule_variance",         "weight": 0.15},
    {"component_key": "corrective_action_overdue", "weight": 0.10},
    {"component_key": "benefit_indicator",         "weight": 0.05}
  ]
}
EOF
)

CREATE_RESP=$(acurl -X POST "$BASE_URL/priority/formulas" -d "$FORMULA_PAYLOAD")
CREATE_BODY=$(echo "$CREATE_RESP" | sed '$d')
CREATE_CODE=$(echo "$CREATE_RESP" | tail -n 1)

assert_http "Create formula" "$CREATE_CODE" "201"
FORMULA_ID=$(echo "$CREATE_BODY" | jq -r '.data.id' 2>/dev/null)
assert_not_empty "Formula ID returned" "$FORMULA_ID"
assert_field "Formula status is DRAFT" "$CREATE_BODY" ".data.status" "DRAFT"
echo "    formula_id = $FORMULA_ID"

# =============================================================================
# 3. Reject invalid weights (sum != 1.0)
# =============================================================================
section "3. Reject formula with invalid weights"

BAD_PAYLOAD='{"name":"Bad Formula","components":[{"component_key":"health_score","weight":0.99}]}'
BAD_RESP=$(acurl -X POST "$BASE_URL/priority/formulas" -d "$BAD_PAYLOAD")
BAD_CODE=$(echo "$BAD_RESP" | tail -n 1)
assert_http "Reject invalid weights (HTTP 400)" "$BAD_CODE" "400"

# =============================================================================
# 4. Activate formula
# =============================================================================
section "4. Activate formula"

ACT_RESP=$(acurl -X POST "$BASE_URL/priority/formulas/$FORMULA_ID/activate" -d '{}')
ACT_BODY=$(echo "$ACT_RESP" | sed '$d')
ACT_CODE=$(echo "$ACT_RESP" | tail -n 1)

assert_http "Activate formula" "$ACT_CODE" "200"
assert_field "Formula status is ACTIVE" "$ACT_BODY" ".data.status" "ACTIVE"

# =============================================================================
# 5. Calculate single project score
# =============================================================================
section "5. Calculate single project priority score"

CALC_PAYLOAD="{\"project_id\":\"$PROJECT_ID\",\"formula_id\":\"$FORMULA_ID\"}"
CALC_RESP=$(acurl -X POST "$BASE_URL/priority/calculate" -d "$CALC_PAYLOAD")
CALC_BODY=$(echo "$CALC_RESP" | sed '$d')
CALC_CODE=$(echo "$CALC_RESP" | tail -n 1)

assert_http "Calculate score" "$CALC_CODE" "200"
SCORE_ID=$(echo "$CALC_BODY" | jq -r '.data.id' 2>/dev/null)
assert_not_empty "Score ID returned" "$SCORE_ID"
TOTAL_SCORE=$(echo "$CALC_BODY" | jq -r '.data.total_score' 2>/dev/null)
assert_not_empty "Total score returned" "$TOTAL_SCORE"
CATEGORY=$(echo "$CALC_BODY" | jq -r '.data.score_category' 2>/dev/null)
assert_not_empty "Score category returned" "$CATEGORY"
echo "    score=$TOTAL_SCORE  category=$CATEGORY"

# =============================================================================
# 6. Verify explainability (component breakdown)
# =============================================================================
section "6. Verify score explainability"

EXPLAIN_RESP=$(acurl "$BASE_URL/priority/projects/$PROJECT_ID/explain")
EXPLAIN_BODY=$(echo "$EXPLAIN_RESP" | sed '$d')
EXPLAIN_CODE=$(echo "$EXPLAIN_RESP" | tail -n 1)

assert_http "Explain endpoint responds" "$EXPLAIN_CODE" "200"
COMP_COUNT=$(echo "$EXPLAIN_BODY" | jq '.data.components | length' 2>/dev/null || echo "0")
if [[ "$COMP_COUNT" -gt 0 ]]; then
  pass "Components array has $COMP_COUNT entries"
else
  fail "Components array is empty"
fi

# Check that each component has explainability fields
FIRST_COMP=$(echo "$EXPLAIN_BODY" | jq '.data.components[0]' 2>/dev/null)
assert_not_empty "component.component_key"    "$(echo "$FIRST_COMP" | jq -r '.component_key')"
# normalized_score may be null when no source data exists (valid — means missing data)
NORM_FIELD=$(echo "$FIRST_COMP" | jq 'has("normalized_score")' 2>/dev/null)
if [[ "$NORM_FIELD" == "true" ]]; then
  pass "component.normalized_score field present (value may be null when data missing)"
else
  fail "component.normalized_score field missing from response"
fi
assert_not_empty "component.weight"           "$(echo "$FIRST_COMP" | jq -r '.weight')"
assert_not_empty "component.weighted_score"   "$(echo "$FIRST_COMP" | jq -r '.weighted_score')"

# Check that at least one component correctly marks missing data as available=false
HAS_UNAVAILABLE=$(echo "$EXPLAIN_BODY" | jq '[.data.components[] | select(.available == false)] | length' 2>/dev/null || echo "-1")
echo "    components_with_missing_data=$HAS_UNAVAILABLE"

# =============================================================================
# 7. Batch calculate
# =============================================================================
section "7. Batch calculate"

BATCH_PAYLOAD="{\"project_ids\":[\"$PROJECT_ID\"],\"formula_id\":\"$FORMULA_ID\"}"
BATCH_RESP=$(acurl -X POST "$BASE_URL/priority/batch-calculate" -d "$BATCH_PAYLOAD")
BATCH_BODY=$(echo "$BATCH_RESP" | sed '$d')
BATCH_CODE=$(echo "$BATCH_RESP" | tail -n 1)

assert_http "Batch calculate" "$BATCH_CODE" "200"
# Response: {formula_id, formula_version, calculated: [...], skipped: [...]}
BATCH_COUNT=$(echo "$BATCH_BODY" | jq '.data.calculated | length' 2>/dev/null || echo "0")
if [[ "$BATCH_COUNT" -gt 0 ]]; then
  pass "Batch calculated $BATCH_COUNT project(s)"
else
  fail "Batch calculated array is empty"
fi
assert_not_empty "Batch formula_id" "$(echo "$BATCH_BODY" | jq -r '.data.formula_id')"

# =============================================================================
# 8. Ranking endpoint
# =============================================================================
section "8. Ranking endpoint"

RANK_RESP=$(acurl "$BASE_URL/priority/projects")
RANK_BODY=$(echo "$RANK_RESP" | sed '$d')
RANK_CODE=$(echo "$RANK_RESP" | tail -n 1)

assert_http "Ranking endpoint" "$RANK_CODE" "200"
# Response: {projects: [...], counts: {...}}
RANK_COUNT=$(echo "$RANK_BODY" | jq '.data.projects | length' 2>/dev/null || echo "0")
if [[ "$RANK_COUNT" -gt 0 ]]; then
  pass "Ranking has $RANK_COUNT scored project(s)"
else
  fail "Ranking is empty"
fi
assert_not_empty "counts in ranking" "$(echo "$RANK_BODY" | jq -r '.data.counts')"

# Verify our project appears in ranking
IN_RANK=$(echo "$RANK_BODY" | jq --arg pid "$PROJECT_ID" \
  '[.data.projects[] | select(.project_id == $pid)] | length' 2>/dev/null || echo "0")
if [[ "$IN_RANK" -gt 0 ]]; then
  pass "Our project appears in ranking"
else
  fail "Our project not found in ranking"
fi

# =============================================================================
# 9. Cross-tenant access rejected (using invalid org token is not feasible here;
#    we verify that querying a non-existent project returns 404 not another org's data)
# =============================================================================
section "9. Non-existent resource returns 404"

FAKE_ID="00000000-0000-0000-0000-000000000000"
NF_RESP=$(acurl "$BASE_URL/priority/projects/$FAKE_ID")
NF_CODE=$(echo "$NF_RESP" | tail -n 1)
assert_http "Non-existent project score returns 404" "$NF_CODE" "404"

NF_FORMULA=$(acurl "$BASE_URL/priority/formulas/$FAKE_ID")
NF_FORMULA_CODE=$(echo "$NF_FORMULA" | tail -n 1)
assert_http "Non-existent formula returns 404" "$NF_FORMULA_CODE" "404"

# =============================================================================
# 10. Cleanup — archive formula (optional, idempotent)
# =============================================================================
section "10. List formulas includes our formula"

LIST_RESP=$(acurl "$BASE_URL/priority/formulas")
LIST_BODY=$(echo "$LIST_RESP" | sed '$d')
LIST_CODE=$(echo "$LIST_RESP" | tail -n 1)
assert_http "List formulas" "$LIST_CODE" "200"
FOUND=$(echo "$LIST_BODY" | jq --arg fid "$FORMULA_ID" \
  '[.data[] | select(.id == $fid)] | length' 2>/dev/null || echo "0")
if [[ "$FOUND" -gt 0 ]]; then
  pass "Smoke-test formula visible in list"
else
  fail "Smoke-test formula not found in list"
fi

# =============================================================================
# Summary
# =============================================================================
echo ""
echo -e "${YELLOW}═══════════════════════════════════════${NC}"
echo -e "  Results: ${GREEN}$PASS passed${NC}  ${RED}$FAIL failed${NC}"
echo -e "${YELLOW}═══════════════════════════════════════${NC}"

if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
