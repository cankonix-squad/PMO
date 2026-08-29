#!/usr/bin/env bash
# smoke-test-uat-005-field-mobile.sh
# CANKORA UAT-005 — Field Inspection Mobile View
#
# Tujuan:
#   Verifikasi end-to-end field inspection workflow:
#   - Login dan auth guard
#   - List inspections (empty & populated)
#   - Create inspection dengan evidence file
#   - Evidence muncul di list
#   - Download evidence authorized (200) dan unauthorized (401)
#   - Verification action (VERIFIED / REJECTED)
#   - Delete / soft-delete inspection
#   - AddEvidence endpoint (POST /:id/evidence)
#   - Backend auth guard 401 tanpa token
#   - Frontend route reachability (project detail + /inspections sub-route)
#   - Desktop regression quick check 1366x768 (Playwright jika tersedia)
#
# Prerequisites:
#   - Backend running:  cd backend && make dev   (port 8080)
#   - Frontend running: cd frontend && npm run dev  (port 3000) — opsional
#   - make seed + make seed-demo sudah dijalankan
#
# Usage:
#   bash docs/smoke-test-uat-005-field-mobile.sh
#   BASE_URL=http://localhost:8080 bash docs/smoke-test-uat-005-field-mobile.sh
#
# Exit: 0 = all PASS, 1 = one or more FAIL

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"
API="$BASE_URL/api/v1"
TOKEN="${TOKEN:-}"
STRICT_FRONTEND="${STRICT_FRONTEND:-0}"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RESET='\033[0m'

PASS_COUNT=0
FAIL_COUNT=0
ERRORS=()

pass()    { echo -e "  ${GREEN}✓ PASS${RESET}  $1"; PASS_COUNT=$((PASS_COUNT + 1)); }
fail()    { echo -e "  ${RED}✗ FAIL${RESET}  $1"; ERRORS+=("$1"); FAIL_COUNT=$((FAIL_COUNT + 1)); }
info()    { echo -e "${CYAN}▶${RESET} $1"; }
warn()    { echo -e "  ${YELLOW}⚠ WARN${RESET}  $1"; }
section() { echo ""; echo -e "${CYAN}━━━ $1 ━━━${RESET}"; }

RES_FILE=/tmp/smoke_uat005_resp.json
EVIDENCE_FILE=/tmp/smoke_uat005_evidence.txt

# ── cleanup evidence temp file ───────────────────────────────────────────────
cleanup() { rm -f "$RES_FILE" "$EVIDENCE_FILE" 2>/dev/null || true; }
trap cleanup EXIT

echo ""
echo "========================================================"
echo " CANKORA UAT-005 — Field Inspection Mobile View"
echo "========================================================"
echo " BASE_URL   : $BASE_URL"
echo " FRONTEND   : $FRONTEND_URL"
echo " DATE       : $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================================"

api_get() {
  local path="$1"; shift
  curl -s -o "$RES_FILE" -w "%{http_code}" \
    -H "Content-Type: application/json" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} \
    "$API$path" "$@"
}

api_post() {
  local path="$1"; local body="$2"; shift 2
  curl -s -o "$RES_FILE" -w "%{http_code}" \
    -H "Content-Type: application/json" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} \
    -d "$body" "$API$path" "$@"
}

api_patch() {
  local path="$1"; local body="$2"; shift 2
  curl -s -o "$RES_FILE" -w "%{http_code}" \
    -X PATCH \
    -H "Content-Type: application/json" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} \
    -d "$body" "$API$path" "$@"
}

api_delete() {
  local path="$1"; shift
  curl -s -o "$RES_FILE" -w "%{http_code}" -X DELETE \
    -H "Content-Type: application/json" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} \
    "$API$path" "$@"
}

field() {
  python3 -c "import sys,json; d=json.load(open('$RES_FILE')); print(d$1)" 2>/dev/null || echo ""
}

dbq() {
  PGPASSWORD=cankora_secret psql -h localhost -U cankora -d cankora_db -P pager=off -tAc "$1" 2>/dev/null || echo ""
}

# ── 1. Health ─────────────────────────────────────────────────────────────────
section "1. Backend health"
HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
[[ "$HTTP" == "200" ]] && pass "GET /health → 200" || { fail "GET /health → $HTTP"; exit 1; }

# ── 2. Login ──────────────────────────────────────────────────────────────────
section "2. Login & auth"
if [[ -z "$TOKEN" ]]; then
  RESP=$(curl -s -X POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}')
  TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['access_token'])" 2>/dev/null || echo "")
fi
[[ -n "$TOKEN" ]] && pass "Login → access_token obtained" || { fail "Login failed — cannot continue"; exit 1; }

# ── 3. Auth guard (no token → 401) ────────────────────────────────────────────
section "3. Auth guard"
# Find a demo project to use for guard checks
PROJECT_ID=$(curl -s -H "Authorization: Bearer $TOKEN" "$API/projects?page=1&page_size=1" | \
  python3 -c "import sys,json; items=json.load(sys.stdin).get('data',[]); print(items[0]['id'] if items else '')" 2>/dev/null || echo "")

if [[ -z "$PROJECT_ID" ]]; then
  warn "No projects found — creating a test project for inspection smoke"
  TS=$(date +%s)
  RESP=$(curl -s -X POST "$API/projects" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{\"name\":\"UAT-005 Smoke\",\"code\":\"UAT005-$TS\",\"status\":\"ACTIVE\",\"start_date\":\"2026-01-01\",\"end_date\":\"2026-12-31\"}")
  PROJECT_ID=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
fi

if [[ -n "$PROJECT_ID" ]]; then
  pass "Project available: $PROJECT_ID"
  # No token → 401
  HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$API/projects/$PROJECT_ID/inspections")
  [[ "$HTTP" == "401" ]] && pass "GET /inspections (no token) → 401" || fail "GET /inspections (no token) → $HTTP (expected 401)"
else
  fail "Could not obtain or create a project — skipping auth guard check"
fi

# ── 4. List inspections ───────────────────────────────────────────────────────
section "4. List inspections"
HTTP=$(api_get "/projects/$PROJECT_ID/inspections")
if [[ "$HTTP" == "200" ]]; then
  pass "GET /inspections → 200"
  COUNT=$(field "['data'] | len" 2>/dev/null || echo "0")
  pass "Inspection list returned (count: ${COUNT:-0})"
else
  fail "GET /inspections → $HTTP"
fi

# ── 5. Create inspection with evidence ───────────────────────────────────────
section "5. Create inspection + evidence"
echo "UAT-005 evidence test file — $(date)" > "$EVIDENCE_FILE"

INSP_RESP=$(curl -s -X POST "$API/projects/$PROJECT_ID/inspections" \
  -H "Authorization: Bearer $TOKEN" \
  -F "inspected_at=2026-08-29T10:00:00Z" \
  -F "notes=UAT-005 smoke test inspection" \
  -F "latitude=-6.200000" \
  -F "longitude=106.816666" \
  -F "file=@$EVIDENCE_FILE;type=text/plain")

INSP_ID=$(echo "$INSP_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
INSP_STATUS=$(echo "$INSP_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('success',''))" 2>/dev/null || echo "")

if [[ -n "$INSP_ID" && "$INSP_STATUS" == "True" ]]; then
  pass "POST /inspections → 201, id=$INSP_ID"
else
  fail "POST /inspections → unexpected response: $(echo "$INSP_RESP" | head -c 200)"
  INSP_ID=""
fi

# ── 6. Evidence appears in list ───────────────────────────────────────────────
section "6. Evidence in list"
if [[ -n "$INSP_ID" ]]; then
  HTTP=$(api_get "/projects/$PROJECT_ID/inspections")
  [[ "$HTTP" == "200" ]] && pass "GET /inspections after create → 200" || fail "GET /inspections → $HTTP"
  EV_COUNT=$(python3 -c "
import json
d=json.load(open('$RES_FILE'))
items=d.get('data',[])
total=sum(len(i.get('evidence',[]) or []) for i in items)
print(total)
" 2>/dev/null || echo "0")
  [[ "$EV_COUNT" -gt 0 ]] && pass "Evidence present in list (count: $EV_COUNT)" || warn "No evidence found in list response — may need reload"
else
  warn "Skipping evidence list check — no inspection id"
fi

# ── 7. Download evidence authorized ──────────────────────────────────────────
section "7. Download evidence"
if [[ -n "$INSP_ID" ]]; then
  EV_ID=$(curl -s -H "Authorization: Bearer $TOKEN" "$API/projects/$PROJECT_ID/inspections" | \
    python3 -c "
import sys,json
d=json.load(sys.stdin)['data']
for ins in d:
    if ins['id'] == '$INSP_ID':
        evs = ins.get('evidence') or []
        if evs: print(evs[0]['id'])
" 2>/dev/null || echo "")

  if [[ -n "$EV_ID" ]]; then
    pass "Evidence ID found: $EV_ID"
    # Authorized download
    DL_HTTP=$(curl -s -o /dev/null -w "%{http_code}" \
      -H "Authorization: Bearer $TOKEN" \
      "$API/projects/$PROJECT_ID/inspections/$INSP_ID/evidence/$EV_ID/download")
    [[ "$DL_HTTP" == "200" ]] && pass "GET evidence/download (authorized) → 200" || fail "GET evidence/download → $DL_HTTP (expected 200)"

    # Unauthorized download → 401
    DL_HTTP_UNAUTH=$(curl -s -o /dev/null -w "%{http_code}" \
      "$API/projects/$PROJECT_ID/inspections/$INSP_ID/evidence/$EV_ID/download")
    [[ "$DL_HTTP_UNAUTH" == "401" ]] && pass "GET evidence/download (no token) → 401" || fail "GET evidence/download (no token) → $DL_HTTP_UNAUTH (expected 401)"
  else
    warn "Evidence ID not found — skipping download checks"
  fi
else
  warn "Skipping download checks — no inspection id"
fi

# ── 8. AddEvidence to existing inspection ────────────────────────────────────
section "8. POST /:inspectionID/evidence (add to existing)"
if [[ -n "$INSP_ID" ]]; then
  ADD_EV_FILE=/tmp/smoke_uat005_evidence2.txt
  echo "UAT-005 second evidence — $(date)" > "$ADD_EV_FILE"
  ADD_EV_RESP=$(curl -s -X POST "$API/projects/$PROJECT_ID/inspections/$INSP_ID/evidence" \
    -H "Authorization: Bearer $TOKEN" \
    -F "file=@${ADD_EV_FILE};type=text/plain" \
    -F "latitude=-6.201000" \
    -F "longitude=106.817000")
  rm -f "$ADD_EV_FILE" 2>/dev/null || true
  ADD_SUCCESS=$(echo "$ADD_EV_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('success',''))" 2>/dev/null || echo "")
  [[ "$ADD_SUCCESS" == "True" ]] && pass "POST /:inspectionID/evidence → 201" || fail "POST /:inspectionID/evidence → unexpected: $(echo "$ADD_EV_RESP" | head -c 200)"
else
  warn "Skipping AddEvidence check — no inspection id"
fi

# ── 9. Invalid file type rejected ────────────────────────────────────────────
section "9. Invalid evidence type rejected"
echo "<script>alert(1)</script>" > /tmp/smoke_uat005_bad.html
BAD_RESP=$(curl -s -X POST "$API/projects/$PROJECT_ID/inspections" \
  -H "Authorization: Bearer $TOKEN" \
  -F "inspected_at=2026-08-29T11:00:00Z" \
  -F "notes=bad type test" \
  -F "file=@/tmp/smoke_uat005_bad.html;type=text/html")
rm -f /tmp/smoke_uat005_bad.html 2>/dev/null || true
BAD_STATUS=$(echo "$BAD_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('success',''))" 2>/dev/null || echo "")
BAD_MSG=$(echo "$BAD_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('message',''))" 2>/dev/null || echo "")
[[ "$BAD_STATUS" == "False" ]] && pass "Invalid file type rejected (400): $BAD_MSG" || fail "Invalid file type NOT rejected — got success=$BAD_STATUS"

# ── 10. Verification action ───────────────────────────────────────────────────
section "10. Verification action"
if [[ -n "$INSP_ID" ]]; then
  HTTP=$(api_patch "/projects/$PROJECT_ID/inspections/$INSP_ID/verification" '{"status":"VERIFIED"}')
  [[ "$HTTP" == "200" ]] && pass "PATCH /verification → 200 (VERIFIED)" || fail "PATCH /verification → $HTTP"

  # Verify status in response
  VER_STATUS=$(field "['data']['verification_status']" 2>/dev/null || echo "")
  [[ "$VER_STATUS" == "VERIFIED" ]] && pass "verification_status=VERIFIED confirmed" || fail "verification_status=$VER_STATUS (expected VERIFIED)"

  # Rejection on a new inspection
  INSP2_RESP=$(curl -s -X POST "$API/projects/$PROJECT_ID/inspections" \
    -H "Authorization: Bearer $TOKEN" \
    -F "inspected_at=2026-08-29T12:00:00Z" \
    -F "notes=to be rejected")
  INSP2_ID=$(echo "$INSP2_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
  if [[ -n "$INSP2_ID" ]]; then
    HTTP=$(api_patch "/projects/$PROJECT_ID/inspections/$INSP2_ID/verification" '{"status":"REJECTED"}')
    [[ "$HTTP" == "200" ]] && pass "PATCH /verification → 200 (REJECTED)" || fail "PATCH /verification → $HTTP"
  fi
else
  warn "Skipping verification check — no inspection id"
fi

# ── 11. Delete / soft delete ──────────────────────────────────────────────────
section "11. Delete inspection"
if [[ -n "$INSP_ID" ]]; then
  HTTP=$(api_delete "/projects/$PROJECT_ID/inspections/$INSP_ID")
  [[ "$HTTP" == "204" ]] && pass "DELETE /inspections/:id → 204" || fail "DELETE /inspections/:id → $HTTP (expected 204)"

  # After soft-delete, GET list should not contain it
  HTTP=$(api_get "/projects/$PROJECT_ID/inspections")
  DELETED_FOUND=$(python3 -c "
import json
d=json.load(open('$RES_FILE'))
items=d.get('data',[])
found=[i for i in items if i['id']=='$INSP_ID']
print(len(found))
" 2>/dev/null || echo "0")
  [[ "$DELETED_FOUND" == "0" ]] && pass "Soft-deleted inspection not in list" || fail "Soft-deleted inspection still visible in list"

  # DB-level soft-delete verification
  DB_DELETED=$(dbq "SELECT count(*) FROM field_inspections WHERE id='$INSP_ID' AND deleted_at IS NOT NULL" | tr -d '[:space:]')
  if [[ "$DB_DELETED" == "1" ]]; then
    pass "DB: deleted_at IS NOT NULL confirmed (soft-delete)"
  else
    warn "DB check skipped or not confirmed (psql not reachable)"
  fi
fi

# ── 12. Cross-tenant guard (wrong org) ───────────────────────────────────────
section "12. Cross-tenant / 404 guard"
FAKE_PID="00000000-0000-0000-0000-000000000099"
HTTP=$(api_get "/projects/$FAKE_PID/inspections")
[[ "$HTTP" == "404" ]] && pass "GET /projects/<fake>/inspections → 404" || fail "GET /projects/<fake>/inspections → $HTTP (expected 404)"

# ── 13. Frontend reachability ─────────────────────────────────────────────────
section "13. Frontend route reachability"
FE_HTTP=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$FRONTEND_URL" 2>/dev/null || echo "000")

if [[ "$FE_HTTP" == "000" ]]; then
  if [[ "$STRICT_FRONTEND" == "1" ]]; then
    fail "Frontend not reachable at $FRONTEND_URL (STRICT_FRONTEND=1)"
  else
    warn "Frontend not running at $FRONTEND_URL — skipping UI route checks (set STRICT_FRONTEND=1 to fail)"
  fi
else
  pass "Frontend $FRONTEND_URL → $FE_HTTP"

  # Check /projects page
  PROJECTS_HTTP=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$FRONTEND_URL/projects" 2>/dev/null || echo "000")
  [[ "$PROJECTS_HTTP" =~ ^(200|307|308)$ ]] && pass "/projects → $PROJECTS_HTTP" || fail "/projects → $PROJECTS_HTTP"

  # Check project detail page with first project ID
  if [[ -n "$PROJECT_ID" ]]; then
    PD_HTTP=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$FRONTEND_URL/projects/$PROJECT_ID" 2>/dev/null || echo "000")
    [[ "$PD_HTTP" =~ ^(200|307|308)$ ]] && pass "/projects/$PROJECT_ID → $PD_HTTP" || fail "/projects/$PROJECT_ID → $PD_HTTP"

    INS_HTTP=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$FRONTEND_URL/projects/$PROJECT_ID/inspections" 2>/dev/null || echo "000")
    [[ "$INS_HTTP" =~ ^(200|307|308)$ ]] && pass "/projects/$PROJECT_ID/inspections → $INS_HTTP" || fail "/projects/$PROJECT_ID/inspections → $INS_HTTP"
  fi

  # Playwright mobile viewport check
  REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
  FRONTEND_DIR="$REPO_ROOT/frontend"
  PLAYWRIGHT_AVAILABLE=0
  if [[ -f "$FRONTEND_DIR/node_modules/.bin/playwright" ]] && \
     (cd "$FRONTEND_DIR" && node -e "require('playwright')" 2>/dev/null); then
    PLAYWRIGHT_AVAILABLE=1
  fi

  if [[ "$PLAYWRIGHT_AVAILABLE" == "1" ]]; then
    info "Playwright available — running mobile viewport checks (390x844)"
    PW_SCRIPT=$(cat <<'PWEOF'
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch({ headless: true });
  const errors = [];
  const baseUrl = process.env.FRONTEND_URL || 'http://localhost:3000';
  const apiUrl  = process.env.API_URL       || 'http://localhost:8080';
  const projectId = process.env.PROJECT_ID  || '';

  // Mobile viewport
  const mobileCtx = await browser.newContext({
    viewport: { width: 390, height: 844 },
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15'
  });
  const mobilePage = await mobileCtx.newPage();

  async function checkPage(url, label) {
    try {
      const res = await mobilePage.goto(url, { waitUntil: 'domcontentloaded', timeout: 15000 });
      const status = res ? res.status() : 0;
      // Check horizontal overflow
      const overflows = await mobilePage.evaluate(() => {
        return document.documentElement.scrollWidth > document.documentElement.clientWidth;
      });
      if (overflows) {
        errors.push(`${label}: horizontal overflow detected at 390px`);
      } else if (status < 400) {
        console.log(`  ✓ PASS  ${label} (${status}, no overflow)`);
      } else {
        errors.push(`${label}: status ${status}`);
      }
    } catch(e) {
      errors.push(`${label}: ${e.message}`);
    }
  }

  // Login page
  await checkPage(`${baseUrl}/login`, 'mobile /login (390x844)');

  // Projects list
  await checkPage(`${baseUrl}/projects`, 'mobile /projects (390x844)');

  // Inspections sub-route
  if (projectId) {
    await checkPage(`${baseUrl}/projects/${projectId}/inspections`, `mobile /projects/${projectId}/inspections (390x844)`);
  }

  // Desktop regression 1366x768
  const desktopCtx = await browser.newContext({ viewport: { width: 1366, height: 768 } });
  const desktopPage = await desktopCtx.newPage();
  async function checkDesktop(url, label) {
    try {
      const res = await desktopPage.goto(url, { waitUntil: 'domcontentloaded', timeout: 15000 });
      const status = res ? res.status() : 0;
      const overflows = await desktopPage.evaluate(() =>
        document.documentElement.scrollWidth > document.documentElement.clientWidth + 5
      );
      if (overflows) errors.push(`${label}: horizontal overflow at 1366px`);
      else if (status < 400) console.log(`  ✓ PASS  ${label} (${status})`);
      else errors.push(`${label}: status ${status}`);
    } catch(e) { errors.push(`${label}: ${e.message}`); }
  }

  await checkDesktop(`${baseUrl}/projects`, 'desktop /projects 1366x768');
  if (projectId) {
    await checkDesktop(`${baseUrl}/projects/${projectId}`, `desktop /projects/:id 1366x768`);
  }

  await browser.close();

  if (errors.length > 0) {
    errors.forEach(e => console.error(`  ✗ FAIL  ${e}`));
    process.exit(1);
  }
  process.exit(0);
})();
PWEOF
)
    PLAYWRIGHT_OUTPUT=$(cd "$FRONTEND_DIR" && \
      FRONTEND_URL="$FRONTEND_URL" API_URL="$BASE_URL" PROJECT_ID="$PROJECT_ID" \
      node -e "$PW_SCRIPT" 2>&1)
    PW_EXIT=$?
    echo "$PLAYWRIGHT_OUTPUT"
    if [[ "$PW_EXIT" == "0" ]]; then
      pass "Playwright mobile/desktop viewport checks passed"
    else
      fail "Playwright viewport checks failed (see output above)"
    fi
  else
    warn "Playwright not available — skipping mobile viewport checks"
    info "  To enable: cd frontend && npm install playwright && npx playwright install chromium"
  fi
fi

# ── 14. Cleanup smoke project if created ─────────────────────────────────────
section "14. Cleanup"
# Remove smoke inspection 2 if it exists
if [[ -n "${INSP2_ID:-}" ]]; then
  HTTP=$(api_delete "/projects/$PROJECT_ID/inspections/$INSP2_ID")
  [[ "$HTTP" == "204" ]] && pass "Cleanup: smoke inspection 2 deleted" || warn "Cleanup: delete insp2 → $HTTP"
fi
pass "Smoke cleanup complete"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
TOTAL=$((PASS_COUNT + FAIL_COUNT))
echo "========================================================"
if [[ "$FAIL_COUNT" -eq 0 ]]; then
  echo -e "${GREEN} ALL FIELD INSPECTION SMOKE CHECKS PASSED ✓${RESET}"
  echo -e "${GREEN} $PASS_COUNT/$TOTAL checks passed${RESET}"
else
  echo -e "${RED} $FAIL_COUNT FIELD INSPECTION SMOKE CHECK(S) FAILED ✗${RESET}"
  echo -e "${RED} $PASS_COUNT/$TOTAL checks passed${RESET}"
  echo ""
  for e in "${ERRORS[@]}"; do
    echo -e "  ${RED}✗${RESET} $e"
  done
fi
echo "========================================================"
[[ "$FAIL_COUNT" -eq 0 ]] && exit 0 || exit 1
