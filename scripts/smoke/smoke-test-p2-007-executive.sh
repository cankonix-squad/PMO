#!/usr/bin/env bash
# ============================================================
# Smoke Test P2-007 — Level 1 Executive Dashboard
# Endpoint: GET /api/v1/analytics/executive
# ============================================================
set -uo pipefail

BASE="http://localhost:8080/api/v1"
PASS=0; FAIL=0

green() { printf "\033[32m[PASS]\033[0m %s\n" "$1"; }
red()   { printf "\033[31m[FAIL]\033[0m %s\n" "$1"; }

check() {
  local desc="$1" result="$2"
  if [ "$result" = "true" ]; then
    green "$desc"; ((PASS++))
  else
    red   "$desc"; ((FAIL++))
  fi
}

# ── Section 1: Auth ───────────────────────────────────────────────────────────
echo ""
echo "=== Section 1: Auth ==="

LOGIN=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}')

TOKEN=$(echo "$LOGIN" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('access_token',''))" 2>/dev/null)

check "T01 — login returns 200 with access_token" \
  "$([ -n "$TOKEN" ] && echo true || echo false)"

check "T02 — token is a non-empty JWT string" \
  "$(echo "$TOKEN" | grep -qE '^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$' && echo true || echo false)"

# ── Section 2: GET /analytics/executive — Top-level shape ────────────────────
echo ""
echo "=== Section 2: Dashboard top-level ==="

RESP=$(curl -s -o /tmp/exec_resp.json -w "%{http_code}" \
  -H "Authorization: Bearer $TOKEN" \
  "$BASE/analytics/executive")

check "T03 — HTTP 200" \
  "$([ "$RESP" = "200" ] && echo true || echo false)"

DATA=$(cat /tmp/exec_resp.json)

check "T04 — response has 'success: true'" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print(str(d.get('success')==True).lower())" 2>/dev/null)"

check "T05 — data.summary exists" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print('true' if 'summary' in d.get('data',{}) else 'false')" 2>/dev/null)"

check "T06 — summary.total_projects is integer >= 0" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); v=d['data']['summary']['total_projects']; print('true' if isinstance(v,int) and v>=0 else 'false')" 2>/dev/null)"

check "T07 — summary.total_budget is number >= 0" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); v=d['data']['summary']['total_budget']; print('true' if (isinstance(v,(int,float))) and v>=0 else 'false')" 2>/dev/null)"

check "T08 — summary.health_green key present" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print('true' if 'health_green' in d['data']['summary'] else 'false')" 2>/dev/null)"

check "T09 — summary.open_risks key present" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print('true' if 'open_risks' in d['data']['summary'] else 'false')" 2>/dev/null)"

check "T10 — summary.open_escalations key present" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print('true' if 'open_escalations' in d['data']['summary'] else 'false')" 2>/dev/null)"

check "T11 — summary.pending_decisions key present" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print('true' if 'pending_decisions' in d['data']['summary'] else 'false')" 2>/dev/null)"

check "T12 — data.as_of key present" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print('true' if 'as_of' in d.get('data',{}) else 'false')" 2>/dev/null)"

# ── Section 3: Critical Projects ─────────────────────────────────────────────
echo ""
echo "=== Section 3: Critical Projects ==="

check "T13 — data.critical_projects is array" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print('true' if isinstance(d['data']['critical_projects'],list) else 'false')" 2>/dev/null)"

CRIT_LEN=$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d['data']['critical_projects']))" 2>/dev/null)

check "T14 — critical_projects length <= 20" \
  "$([ "${CRIT_LEN:-0}" -le 20 ] && echo true || echo false)"

check "T15 — critical_projects item has project_id field (if non-empty)" \
  "$(echo "$DATA" | python3 -c "
import sys,json
d=json.load(sys.stdin)
items=d['data']['critical_projects']
if not items: print('true')
else: print('true' if 'project_id' in items[0] else 'false')
" 2>/dev/null)"

check "T16 — critical_projects item has health_class field (if non-empty)" \
  "$(echo "$DATA" | python3 -c "
import sys,json
d=json.load(sys.stdin)
items=d['data']['critical_projects']
if not items: print('true')
else: print('true' if 'health_class' in items[0] else 'false')
" 2>/dev/null)"

# ── Section 4: Escalations ────────────────────────────────────────────────────
echo ""
echo "=== Section 4: Escalations ==="

check "T17 — data.escalations is array" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print('true' if isinstance(d['data']['escalations'],list) else 'false')" 2>/dev/null)"

check "T18 — escalations item has reason field (if non-empty)" \
  "$(echo "$DATA" | python3 -c "
import sys,json
d=json.load(sys.stdin)
items=d['data']['escalations']
if not items: print('true')
else: print('true' if 'reason' in items[0] else 'false')
" 2>/dev/null)"

check "T19 — escalations item has status field (if non-empty)" \
  "$(echo "$DATA" | python3 -c "
import sys,json
d=json.load(sys.stdin)
items=d['data']['escalations']
if not items: print('true')
else: print('true' if 'status' in items[0] else 'false')
" 2>/dev/null)"

# ── Section 5: Pending Decisions ──────────────────────────────────────────────
echo ""
echo "=== Section 5: Pending Decisions ==="

check "T20 — data.pending_decisions is array" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print('true' if isinstance(d['data']['pending_decisions'],list) else 'false')" 2>/dev/null)"

check "T21 — pending_decisions item has subject field (if non-empty)" \
  "$(echo "$DATA" | python3 -c "
import sys,json
d=json.load(sys.stdin)
items=d['data']['pending_decisions']
if not items: print('true')
else: print('true' if 'subject' in items[0] else 'false')
" 2>/dev/null)"

check "T22 — pending_decisions item has is_overdue field (if non-empty)" \
  "$(echo "$DATA" | python3 -c "
import sys,json
d=json.load(sys.stdin)
items=d['data']['pending_decisions']
if not items: print('true')
else: print('true' if 'is_overdue' in items[0] else 'false')
" 2>/dev/null)"

# ── Section 6: Programs ───────────────────────────────────────────────────────
echo ""
echo "=== Section 6: Programs ==="

check "T23 — data.programs is array" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print('true' if isinstance(d['data']['programs'],list) else 'false')" 2>/dev/null)"

check "T24 — programs item has program_id field (if non-empty)" \
  "$(echo "$DATA" | python3 -c "
import sys,json
d=json.load(sys.stdin)
items=d['data']['programs']
if not items: print('true')
else: print('true' if 'program_id' in items[0] else 'false')
" 2>/dev/null)"

check "T25 — programs item has total_budget field (if non-empty)" \
  "$(echo "$DATA" | python3 -c "
import sys,json
d=json.load(sys.stdin)
items=d['data']['programs']
if not items: print('true')
else: print('true' if 'total_budget' in items[0] else 'false')
" 2>/dev/null)"

# ── Section 7: Benefits ───────────────────────────────────────────────────────
echo ""
echo "=== Section 7: Benefits ==="

check "T26 — data.benefits exists" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); print('true' if 'benefits' in d.get('data',{}) else 'false')" 2>/dev/null)"

check "T27 — benefits.total_indicators is integer >= 0" \
  "$(echo "$DATA" | python3 -c "import sys,json; d=json.load(sys.stdin); v=d['data']['benefits']['total_indicators']; print('true' if isinstance(v,int) and v>=0 else 'false')" 2>/dev/null)"

# ── Section 8: Auth guard ─────────────────────────────────────────────────────
echo ""
echo "=== Section 8: Auth guard ==="

UNAUTH=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/analytics/executive")

check "T28 — GET without token returns 401" \
  "$([ "$UNAUTH" = "401" ] && echo true || echo false)"

BAD_TOKEN_RESP=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer invalidtoken123" \
  "$BASE/analytics/executive")

check "T29 — GET with invalid token returns 401" \
  "$([ "$BAD_TOKEN_RESP" = "401" ] && echo true || echo false)"

# ── Section 9: Period filter ──────────────────────────────────────────────────
echo ""
echo "=== Section 9: Period filter ==="

FILTER_RESP=$(curl -s -o /tmp/exec_filter.json -w "%{http_code}" \
  -H "Authorization: Bearer $TOKEN" \
  "$BASE/analytics/executive?year=2024&month=6")

check "T30 — GET with year+month filter returns 200" \
  "$([ "$FILTER_RESP" = "200" ] && echo true || echo false)"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "============================================"
TOTAL=$((PASS + FAIL))
echo "RESULT: $PASS/$TOTAL PASS"
if [ "$FAIL" -eq 0 ]; then
  printf "\033[32mAll tests passed!\033[0m\n"
else
  printf "\033[31m$FAIL test(s) failed.\033[0m\n"
  exit 1
fi
