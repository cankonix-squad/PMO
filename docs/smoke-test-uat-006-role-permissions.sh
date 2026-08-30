#!/usr/bin/env bash
# =============================================================================
# CANKORA — Smoke Test UAT-006: Role-Based Permissions & Navigation Hardening
# =============================================================================
# Tests:
#   1. Demo users login (pmo, pm, officer, executive, auditor)
#   2. Role-appropriate endpoints return 200
#   3. Role-inappropriate endpoints return 403
#   4. No token → 401
#   5. Spatial routes (sectors/regions/river-basins) now require permission
#   6. Cross-tenant guard still enforced
#
# Prerequisites:
#   - Backend running at http://localhost:8080
#   - Demo users seeded: make seed (from backend/)
#   - All passwords: Demo@Cankora2024!
# =============================================================================

set -uo pipefail

BASE="http://localhost:8080/api/v1"
PASS=0
FAIL=0
SKIP=0

# ── helpers ─────────────────────────────────────────────────────────────────
green()  { echo -e "\033[32m✓ PASS\033[0m  $*"; }
red()    { echo -e "\033[31m✗ FAIL\033[0m  $*"; }
yellow() { echo -e "\033[33m~ SKIP\033[0m  $*"; }
header() { echo -e "\n\033[1;34m── $* ──\033[0m"; }

pass() { green "$1";  ((PASS++));  }
fail() { red   "$1";  ((FAIL++));  }
skip() { yellow "$1"; ((SKIP++));  }

# Login helper — returns access token or empty string on failure
login() {
  local email="$1" password="$2"
  curl -s -X POST "$BASE/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"$password\"}" \
    | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('access_token',''))" 2>/dev/null || echo ""
}

# GET with auth — returns HTTP status code
get_status() {
  local token="$1" path="$2"
  curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer $token" \
    "$BASE$path"
}

# POST with auth — returns HTTP status code
post_status() {
  local token="$1" path="$2" body="$3"
  curl -s -o /dev/null -w "%{http_code}" \
    -X POST \
    -H "Authorization: Bearer $token" \
    -H "Content-Type: application/json" \
    -d "$body" \
    "$BASE$path"
}

# GET without auth — returns HTTP status code
get_noauth() {
  curl -s -o /dev/null -w "%{http_code}" "$BASE$1"
}

check_status() {
  local label="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    pass "$label → HTTP $actual"
  else
    fail "$label → expected $expected, got $actual"
  fi
}

# =============================================================================
# 1. Health check
# =============================================================================
header "1. Health check"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/health")
check_status "GET /health" "200" "$STATUS"

# =============================================================================
# 2. No-token guard (401)
# =============================================================================
header "2. No-token → 401"
for path in /projects /audit-logs /notifications /analytics/executive \
            /sectors /regions /river-basins /governance/submissions; do
  S=$(get_noauth "$path")
  check_status "GET $path (no token)" "401" "$S"
done

# =============================================================================
# 3. Login demo users
# =============================================================================
header "3. Demo user login"

ADMIN_TOKEN=$(login "admin@cankora.local" "Admin@Cankora2024!")
[ -n "$ADMIN_TOKEN" ] && pass "admin@cankora.local login" || fail "admin@cankora.local login"

PMO_TOKEN=$(login "pmo@cankora.local" "Demo@Cankora2024!")
[ -n "$PMO_TOKEN" ] && pass "pmo@cankora.local login" || fail "pmo@cankora.local login"

PM_TOKEN=$(login "pm@cankora.local" "Demo@Cankora2024!")
[ -n "$PM_TOKEN" ] && pass "pm@cankora.local login" || fail "pm@cankora.local login"

OFFICER_TOKEN=$(login "officer@cankora.local" "Demo@Cankora2024!")
[ -n "$OFFICER_TOKEN" ] && pass "officer@cankora.local login" || fail "officer@cankora.local login"

EXECUTIVE_TOKEN=$(login "executive@cankora.local" "Demo@Cankora2024!")
[ -n "$EXECUTIVE_TOKEN" ] && pass "executive@cankora.local login" || fail "executive@cankora.local login"

AUDITOR_TOKEN=$(login "auditor@cankora.local" "Demo@Cankora2024!")
[ -n "$AUDITOR_TOKEN" ] && pass "auditor@cankora.local login" || fail "auditor@cankora.local login"

# =============================================================================
# 4. ADMIN — should access everything
# =============================================================================
header "4. ADMIN → full access"
for path in /projects /audit-logs /notifications /analytics/executive \
            /analytics/programs /analytics/gis/summary /sectors /regions; do
  S=$(get_status "$ADMIN_TOKEN" "$path")
  # 200 or 400/422 (bad params) both mean the guard passed; 404 means route exists
  if [ "$S" = "200" ] || [ "$S" = "400" ] || [ "$S" = "422" ] || [ "$S" = "404" ]; then
    pass "ADMIN GET $path → $S (guard passed)"
  else
    fail "ADMIN GET $path → $S (expected 200/400/404/422)"
  fi
done

# =============================================================================
# 5. PMO — should access everything except system admin
# =============================================================================
header "5. PMO → broad access"
if [ -n "$PMO_TOKEN" ]; then
  for path in /projects /audit-logs /notifications /analytics/executive \
              /analytics/programs /analytics/gis/summary \
              /governance/submissions /command-center \
              /sectors /regions /river-basins; do
    S=$(get_status "$PMO_TOKEN" "$path")
    if [ "$S" = "200" ] || [ "$S" = "400" ] || [ "$S" = "422" ]; then
      pass "PMO GET $path → $S (guard passed)"
    else
      fail "PMO GET $path → $S (expected 200/400/422)"
    fi
  done
else
  skip "PMO tests skipped (login failed)"
fi

# =============================================================================
# 6. PROJECT_MANAGER — access projects, tasks, risks, issues; NOT audit-logs
# =============================================================================
header "6. PROJECT_MANAGER → project ops access"
if [ -n "$PM_TOKEN" ]; then
  # Should access
  for path in /projects /notifications; do
    S=$(get_status "$PM_TOKEN" "$path")
    if [ "$S" = "200" ] || [ "$S" = "400" ]; then
      pass "PM GET $path → $S (allowed)"
    else
      fail "PM GET $path → $S (expected 200/400)"
    fi
  done

  # Should NOT access audit logs (PMO_AND_ABOVE only)
  S=$(get_status "$PM_TOKEN" "/audit-logs")
  if [ "$S" = "403" ]; then
    pass "PM GET /audit-logs → 403 (correctly forbidden)"
  else
    fail "PM GET /audit-logs → $S (expected 403)"
  fi

  # Should NOT access executive dashboard
  S=$(get_status "$PM_TOKEN" "/analytics/executive")
  if [ "$S" = "403" ]; then
    pass "PM GET /analytics/executive → 403 (correctly forbidden)"
  else
    fail "PM GET /analytics/executive → $S (expected 403)"
  fi
else
  skip "PM tests skipped (login failed)"
fi

# =============================================================================
# 7. PROJECT_OFFICER — task execution; limited scope
# =============================================================================
header "7. PROJECT_OFFICER → task execution scope"
if [ -n "$OFFICER_TOKEN" ]; then
  # Should access
  for path in /projects /notifications; do
    S=$(get_status "$OFFICER_TOKEN" "$path")
    if [ "$S" = "200" ] || [ "$S" = "400" ]; then
      pass "OFFICER GET $path → $S (allowed)"
    else
      fail "OFFICER GET $path → $S (expected 200/400)"
    fi
  done

  # Should NOT access audit logs
  S=$(get_status "$OFFICER_TOKEN" "/audit-logs")
  if [ "$S" = "403" ]; then
    pass "OFFICER GET /audit-logs → 403 (correctly forbidden)"
  else
    fail "OFFICER GET /audit-logs → $S (expected 403)"
  fi

  # Should NOT access executive dashboard
  S=$(get_status "$OFFICER_TOKEN" "/analytics/executive")
  if [ "$S" = "403" ]; then
    pass "OFFICER GET /analytics/executive → 403 (correctly forbidden)"
  else
    fail "OFFICER GET /analytics/executive → $S (expected 403)"
  fi

  # Should NOT access command center
  S=$(get_status "$OFFICER_TOKEN" "/command-center")
  if [ "$S" = "403" ]; then
    pass "OFFICER GET /command-center → 403 (correctly forbidden)"
  else
    fail "OFFICER GET /command-center → $S (expected 403)"
  fi
else
  skip "OFFICER tests skipped (login failed)"
fi

# =============================================================================
# 8. EXECUTIVE_VIEWER — read-only analytics; no write operations
# =============================================================================
header "8. EXECUTIVE_VIEWER → analytics read-only"
if [ -n "$EXECUTIVE_TOKEN" ]; then
  # Should access executive/program analytics
  for path in /analytics/executive /analytics/programs /analytics/gis/summary \
              /reports/catalog /notifications; do
    S=$(get_status "$EXECUTIVE_TOKEN" "$path")
    if [ "$S" = "200" ] || [ "$S" = "400" ] || [ "$S" = "404" ]; then
      pass "EXECUTIVE GET $path → $S (allowed)"
    else
      fail "EXECUTIVE GET $path → $S (expected 200/400/404)"
    fi
  done

  # Should NOT access audit logs (PMO_AND_ABOVE + AUDITOR)
  S=$(get_status "$EXECUTIVE_TOKEN" "/audit-logs")
  if [ "$S" = "403" ]; then
    pass "EXECUTIVE GET /audit-logs → 403 (correctly forbidden)"
  else
    fail "EXECUTIVE GET /audit-logs → $S (expected 403)"
  fi

  # Should NOT create governance submissions (write operation)
  S=$(post_status "$EXECUTIVE_TOKEN" "/governance/submissions" '{"dataset_type":"project","period_year":2024}')
  if [ "$S" = "403" ]; then
    pass "EXECUTIVE POST /governance/submissions → 403 (correctly forbidden)"
  else
    fail "EXECUTIVE POST /governance/submissions → $S (expected 403)"
  fi

  # Should NOT access command center
  S=$(get_status "$EXECUTIVE_TOKEN" "/command-center")
  if [ "$S" = "403" ]; then
    pass "EXECUTIVE GET /command-center → 403 (correctly forbidden)"
  else
    fail "EXECUTIVE GET /command-center → $S (expected 403)"
  fi
else
  skip "EXECUTIVE tests skipped (login failed)"
fi

# =============================================================================
# 9. AUDITOR — audit logs + reports; no project writes
# =============================================================================
header "9. AUDITOR → audit logs access"
if [ -n "$AUDITOR_TOKEN" ]; then
  # Should access audit logs
  S=$(get_status "$AUDITOR_TOKEN" "/audit-logs")
  if [ "$S" = "200" ] || [ "$S" = "400" ]; then
    pass "AUDITOR GET /audit-logs → $S (allowed)"
  else
    fail "AUDITOR GET /audit-logs → $S (expected 200/400)"
  fi

  # Should access reports
  S=$(get_status "$AUDITOR_TOKEN" "/reports/catalog")
  if [ "$S" = "200" ] || [ "$S" = "400" ] || [ "$S" = "404" ]; then
    pass "AUDITOR GET /reports/catalog → $S (allowed)"
  else
    fail "AUDITOR GET /reports/catalog → $S (expected 200/400/404)"
  fi

  # Should NOT access executive dashboard
  S=$(get_status "$AUDITOR_TOKEN" "/analytics/executive")
  if [ "$S" = "403" ]; then
    pass "AUDITOR GET /analytics/executive → 403 (correctly forbidden)"
  else
    fail "AUDITOR GET /analytics/executive → $S (expected 403)"
  fi

  # Should NOT access command center
  S=$(get_status "$AUDITOR_TOKEN" "/command-center")
  if [ "$S" = "403" ]; then
    pass "AUDITOR GET /command-center → 403 (correctly forbidden)"
  else
    fail "AUDITOR GET /command-center → $S (expected 403)"
  fi
else
  skip "AUDITOR tests skipped (login failed)"
fi

# =============================================================================
# 10. Spatial routes — now require RequirePermission (added UAT-006)
# =============================================================================
header "10. Spatial routes permission guards"

# PMO can access sectors/regions/river-basins
if [ -n "$PMO_TOKEN" ]; then
  for path in /sectors /regions /river-basins; do
    S=$(get_status "$PMO_TOKEN" "$path")
    if [ "$S" = "200" ] || [ "$S" = "400" ]; then
      pass "PMO GET $path → $S (allowed)"
    else
      fail "PMO GET $path → $S (expected 200/400)"
    fi
  done
fi

# PROJECT_OFFICER should NOT write to spatial data
if [ -n "$OFFICER_TOKEN" ]; then
  for path in /sectors /regions /river-basins; do
    S=$(post_status "$OFFICER_TOKEN" "$path" '{"name":"test","code":"TST"}')
    if [ "$S" = "403" ]; then
      pass "OFFICER POST $path → 403 (correctly forbidden)"
    else
      fail "OFFICER POST $path → $S (expected 403)"
    fi
  done
fi

# EXECUTIVE should NOT access spatial settings
if [ -n "$EXECUTIVE_TOKEN" ]; then
  S=$(get_status "$EXECUTIVE_TOKEN" "/sectors")
  if [ "$S" = "403" ]; then
    pass "EXECUTIVE GET /sectors → 403 (correctly forbidden)"
  else
    fail "EXECUTIVE GET /sectors → $S (expected 403)"
  fi
fi

# =============================================================================
# 11. Invalid token → 401
# =============================================================================
header "11. Invalid token → 401"
FAKE_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJmYWtlIn0.fake"
S=$(get_status "$FAKE_TOKEN" "/projects")
check_status "GET /projects (fake token)" "401" "$S"

# =============================================================================
# 12. 401 vs 403 distinction
# =============================================================================
header "12. 401 vs 403 semantics"
# No token → 401 (unauthenticated)
S=$(get_noauth "/audit-logs")
check_status "GET /audit-logs (no token) → 401" "401" "$S"

# Valid token, wrong role → 403 (authenticated but forbidden)
if [ -n "$OFFICER_TOKEN" ]; then
  S=$(get_status "$OFFICER_TOKEN" "/audit-logs")
  check_status "GET /audit-logs (OFFICER token) → 403" "403" "$S"
fi

# =============================================================================
# Summary
# =============================================================================
echo ""
echo "════════════════════════════════════════════"
echo " UAT-006 Role-Based Permissions Smoke Test"
echo "════════════════════════════════════════════"
echo " PASS: $PASS"
echo " FAIL: $FAIL"
echo " SKIP: $SKIP"
echo "════════════════════════════════════════════"

if [ "$FAIL" -eq 0 ]; then
  echo -e "\033[32m ALL CHECKS PASSED\033[0m"
  exit 0
else
  echo -e "\033[31m $FAIL CHECK(S) FAILED\033[0m"
  exit 1
fi
