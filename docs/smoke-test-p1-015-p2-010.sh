#!/bin/bash
# =============================================================================
# CANKORA P1-015 / P2-010 — Command Center + Benefit smoke
# Usage: bash docs/smoke-test-p1-015-p2-010.sh
# Prereq: backend running on :8080, DB migrated + seeded
# =============================================================================
set -e

BASE=${BASE:-http://localhost:8080/api/v1}
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'
PASS=0
FAIL=0

ok() { echo -e "${GREEN}✓ PASS${NC} $1"; PASS=$((PASS+1)); }
bad() { echo -e "${RED}✗ FAIL${NC} $1"; FAIL=$((FAIL+1)); }

TOKEN=$(curl -s -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['access_token'])")
AUTH="Authorization: Bearer $TOKEN"
CT="Content-Type: application/json"
STAMP=$(date +%s)

PROJECT_A=$(curl -s -X POST "$BASE/projects" -H "$AUTH" -H "$CT" -d "{
  \"code\":\"CMD-BEN-A-$STAMP\",\"name\":\"Smoke Command Benefit A\",\"priority\":\"HIGH\",
  \"start_date\":\"2026-01-01\",\"end_date\":\"2026-12-31\",\"budget_total\":1000000
}")
PROJECT_B=$(curl -s -X POST "$BASE/projects" -H "$AUTH" -H "$CT" -d "{
  \"code\":\"CMD-BEN-B-$STAMP\",\"name\":\"Smoke Command Benefit B\",\"priority\":\"MEDIUM\",
  \"start_date\":\"2026-01-01\",\"end_date\":\"2026-12-31\",\"budget_total\":1000000
}")
PID_A=$(echo "$PROJECT_A" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")
PID_B=$(echo "$PROJECT_B" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")
[ -n "$PID_A" ] && [ -n "$PID_B" ] && ok "created two projects for source mismatch smoke" || bad "create smoke projects"

RISK=$(curl -s -X POST "$BASE/projects/$PID_A/risks" -H "$AUTH" -H "$CT" -d '{
  "title":"Smoke strategic risk","probability":5,"impact":5,"description":"source for command center"
}')
RISK_ID=$(echo "$RISK" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")
[ -n "$RISK_ID" ] && ok "created source risk" || bad "create source risk"

HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/command-center/escalations" -H "$AUTH" -H "$CT" -d "{
  \"project_id\":\"$PID_B\",\"source_type\":\"risk\",\"source_id\":\"$RISK_ID\",
  \"level\":\"EXECUTIVE\",\"reason\":\"must reject source/project mismatch\"
}")
[ "$HTTP" = "404" ] && ok "command escalation rejects wrong project/source reference" || bad "source/project mismatch expected 404 got $HTTP"

ESC=$(curl -s -X POST "$BASE/command-center/escalations" -H "$AUTH" -H "$CT" -d "{
  \"project_id\":\"$PID_A\",\"source_type\":\"risk\",\"source_id\":\"$RISK_ID\",
  \"level\":\"EXECUTIVE\",\"reason\":\"valid command center smoke\"
}")
ESC_ID=$(echo "$ESC" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")
[ -n "$ESC_ID" ] && ok "created valid escalation" || bad "create valid escalation: $ESC"

curl -s -X PATCH "$BASE/command-center/escalations/$ESC_ID/status" -H "$AUTH" -H "$CT" -d '{"status":"ACKNOWLEDGED"}' \
  | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['acknowledged_by']" \
  && ok "acknowledge escalation sets acknowledged_by" || bad "acknowledge escalation"

DEC=$(curl -s -X POST "$BASE/command-center/decisions" -H "$AUTH" -H "$CT" -d "{
  \"escalation_id\":\"$ESC_ID\",\"subject\":\"Smoke executive direction\",
  \"decision_text\":\"Proceed with mitigation follow-up\",\"due_date\":\"2026-10-01\"
}")
DEC_ID=$(echo "$DEC" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")
[ -n "$DEC_ID" ] && ok "created executive decision with decided_by" || bad "create decision: $DEC"

curl -s -X PATCH "$BASE/command-center/decisions/$DEC_ID/status" -H "$AUTH" -H "$CT" -d '{"status":"COMPLETED"}' \
  | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['status']=='COMPLETED'" \
  && ok "complete executive decision" || bad "complete decision"

IND=$(curl -s -X POST "$BASE/benefits" -H "$AUTH" -H "$CT" -d "{
  \"project_id\":\"$PID_A\",\"name\":\"Luas Layanan Irigasi Smoke\",\"unit\":\"Ha\",
  \"aggregation_method\":\"SUM\",\"source\":\"smoke\",\"description\":\"P2-010 benefit smoke\"
}")
IND_ID=$(echo "$IND" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")
[ -n "$IND_ID" ] && ok "created benefit indicator" || bad "create benefit indicator: $IND"

curl -s -X POST "$BASE/benefits/$IND_ID/measurements" -H "$AUTH" -H "$CT" -d '{
  "period_year":2026,"period_month":8,"baseline":100,"target":150,"actual":125,
  "source":"smoke","validation_status":"VALID"
}' > /dev/null
curl -s -X POST "$BASE/benefits/$IND_ID/measurements" -H "$AUTH" -H "$CT" -d '{
  "period_year":2026,"period_month":9,"baseline":100,"target":170,"actual":135,
  "source":"smoke","validation_status":"DRAFT"
}' > /dev/null

curl -s "$BASE/benefits/$IND_ID/aggregate" -H "$AUTH" \
  | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['count']==1 and d['data']['value']==125" \
  && ok "benefit aggregate uses VALID measurement only" || bad "benefit aggregate"

curl -s "$BASE/benefits/summary" -H "$AUTH" \
  | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and any(x['unit']=='HA' for x in d['data'])" \
  && ok "benefit summary groups compatible unit/method" || bad "benefit summary"

curl -s -X DELETE "$BASE/projects/$PID_A" -H "$AUTH" > /dev/null
HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/benefits/$IND_ID" -H "$AUTH")
[ "$HTTP" = "404" ] && ok "project delete cascades project benefit indicator" || bad "benefit indicator still active after project delete (got $HTTP)"
curl -s -X DELETE "$BASE/projects/$PID_B" -H "$AUTH" > /dev/null

echo ""
echo "SMOKE RESULT P1-015/P2-010: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ]
