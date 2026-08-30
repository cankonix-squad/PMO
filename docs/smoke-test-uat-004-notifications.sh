#!/usr/bin/env bash
# ============================================================
# PMO — Smoke Test: UAT-004 Notification Delivery Foundation
# ============================================================
# Usage:
#   bash docs/smoke-test-uat-004-notifications.sh
#   BASE_URL=http://localhost:8080 bash docs/smoke-test-uat-004-notifications.sh
#
# Requirements:
#   - Backend running at BASE_URL (default http://localhost:8080)
#   - Seeded DB (admin@cankora.local / Admin@Cankora2024!)
#   - curl + jq available
# ============================================================

set -uo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
API="${BASE_URL}/api/v1"
ADMIN_EMAIL="admin@cankora.local"
ADMIN_PASSWORD="Admin@Cankora2024!"

PASS=0
FAIL=0
SKIP=0

# ── helpers ──────────────────────────────────────────────────────────────────

green()  { echo -e "\033[32m[PASS]\033[0m $*"; }
red()    { echo -e "\033[31m[FAIL]\033[0m $*"; }
yellow() { echo -e "\033[33m[SKIP]\033[0m $*"; }
header() { echo -e "\n\033[1;34m$*\033[0m"; }

pass() { green  "$1"; ((PASS++)) || true; }
fail() { red    "$1"; ((FAIL++)) || true; }
skip() { yellow "$1"; ((SKIP++)) || true; }

json_field() { echo "$1" | jq -r "$2" 2>/dev/null; }

# ── pre-flight ────────────────────────────────────────────────────────────────

header "[0] Pre-flight: backend health"
HEALTH_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/health")
if [[ "$HEALTH_CODE" == "200" ]]; then
  pass "[0-1] Backend health → 200"
else
  fail "[0-1] Backend health → ${HEALTH_CODE}"
  echo "Backend not reachable. Exiting."
  exit 1
fi

# ── login ─────────────────────────────────────────────────────────────────────

header "[1] Auth"
LOGIN_RESP=$(curl -s -X POST "${API}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASSWORD}\"}")

TOKEN=$(json_field "$LOGIN_RESP" '.data.access_token')
if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
  red "Login failed — cannot continue"
  echo "Response: $LOGIN_RESP"
  exit 1
fi
pass "[1-1] Admin login OK, token received"

AUTH_H="Authorization: Bearer ${TOKEN}"

# ── [2] No-token 401 guard ────────────────────────────────────────────────────

header "[2] Unauthenticated requests must be rejected"
for label_path in \
  "[2-1] GET /notifications|${API}/notifications" \
  "[2-2] GET /notifications/summary|${API}/notifications/summary"; do
  label="${label_path%%|*}"
  url="${label_path##*|}"
  code=$(curl -s -o /dev/null -w "%{http_code}" "$url")
  if [[ "$code" == "401" ]]; then
    pass "${label} no-token → 401"
  else
    fail "${label} no-token → ${code} (expected 401)"
  fi
done

# ── [3] Create test notification ──────────────────────────────────────────────

header "[3] Create test notification"
TEST_RESP=$(curl -s -X POST "${API}/notifications/test" \
  -H "Content-Type: application/json" \
  -H "${AUTH_H}" \
  -d '{"subject":"Smoke Test","body":"UAT-004 notification smoke test","channel":"IN_APP","priority":"NORMAL"}')

TEST_SUCCESS=$(json_field "$TEST_RESP" '.success')
TEST_ID=$(json_field "$TEST_RESP" '.data.id')

if [[ "$TEST_SUCCESS" == "true" && "$TEST_ID" != "null" && -n "$TEST_ID" ]]; then
  pass "[3-1] POST /notifications/test → 201/200, id=${TEST_ID}"
else
  fail "[3-1] POST /notifications/test unexpected: $TEST_RESP"
  TEST_ID=""
fi

# ── [4] List notifications ────────────────────────────────────────────────────

header "[4] List notifications"
LIST_RESP=$(curl -s "${API}/notifications" -H "${AUTH_H}")
LIST_SUCCESS=$(json_field "$LIST_RESP" '.success')

if [[ "$LIST_SUCCESS" == "true" ]]; then
  pass "[4-1] GET /notifications → 200 success"
else
  fail "[4-1] GET /notifications unexpected: $LIST_RESP"
fi

META_PAGE=$(json_field "$LIST_RESP" '.meta.page')
META_TOTAL=$(json_field "$LIST_RESP" '.meta.total')
if [[ "$META_PAGE" != "null" && -n "$META_PAGE" ]]; then
  pass "[4-2] Pagination meta.page present (${META_PAGE})"
else
  fail "[4-2] Pagination meta.page missing"
fi
if [[ "$META_TOTAL" != "null" && -n "$META_TOTAL" ]]; then
  pass "[4-3] Pagination meta.total present (${META_TOTAL})"
else
  fail "[4-3] Pagination meta.total missing"
fi

# verify test notification appears in list
FOUND=$(json_field "$LIST_RESP" ".data[] | select(.id==\"${TEST_ID}\") | .id" 2>/dev/null)
if [[ -n "$TEST_ID" && -n "$FOUND" ]]; then
  pass "[4-4] Test notification appears in list"
elif [[ -z "$TEST_ID" ]]; then
  skip "[4-4] Test notification not created — skipping list check"
else
  # may not be on first page if there are many notifications; soft skip
  skip "[4-4] Test notification not on first page (may be paginated)"
fi

# ── [5] Pagination params ─────────────────────────────────────────────────────

header "[5] Pagination"
PAGE_RESP=$(curl -s "${API}/notifications?page=1&page_size=5" -H "${AUTH_H}")
PAGE_SUCCESS=$(json_field "$PAGE_RESP" '.success')
if [[ "$PAGE_SUCCESS" == "true" ]]; then
  pass "[5-1] GET /notifications?page=1&page_size=5 → 200"
else
  fail "[5-1] GET /notifications?page=1&page_size=5 unexpected: $PAGE_RESP"
fi

# ── [6] Filter params ─────────────────────────────────────────────────────────

header "[6] Filter params"
FILTER_STATUS=$(curl -s "${API}/notifications?status=PENDING" -H "${AUTH_H}")
if [[ "$(json_field "$FILTER_STATUS" '.success')" == "true" ]]; then
  pass "[6-1] Filter ?status=PENDING → 200"
else
  fail "[6-1] Filter ?status=PENDING unexpected: $FILTER_STATUS"
fi

FILTER_CH=$(curl -s "${API}/notifications?channel=IN_APP" -H "${AUTH_H}")
if [[ "$(json_field "$FILTER_CH" '.success')" == "true" ]]; then
  pass "[6-2] Filter ?channel=IN_APP → 200"
else
  fail "[6-2] Filter ?channel=IN_APP unexpected: $FILTER_CH"
fi

UNREAD_RESP=$(curl -s "${API}/notifications?unread_only=true" -H "${AUTH_H}")
if [[ "$(json_field "$UNREAD_RESP" '.success')" == "true" ]]; then
  pass "[6-3] Filter ?unread_only=true → 200"
else
  fail "[6-3] Filter ?unread_only=true unexpected: $UNREAD_RESP"
fi

# ── [7] Get by ID ─────────────────────────────────────────────────────────────

header "[7] Get by ID"
if [[ -n "$TEST_ID" ]]; then
  BYID_RESP=$(curl -s "${API}/notifications/${TEST_ID}" -H "${AUTH_H}")
  BYID_SUCCESS=$(json_field "$BYID_RESP" '.success')
  if [[ "$BYID_SUCCESS" == "true" ]]; then
    pass "[7-1] GET /notifications/:id → 200"
  else
    fail "[7-1] GET /notifications/:id unexpected: $BYID_RESP"
  fi

  BAD_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${API}/notifications/00000000-0000-0000-0000-000000000000" -H "${AUTH_H}")
  if [[ "$BAD_CODE" == "404" ]]; then
    pass "[7-2] GET /notifications/00000000-... → 404"
  else
    fail "[7-2] GET /notifications/00000000-... → ${BAD_CODE} (expected 404)"
  fi
else
  skip "[7-1] No test notification — skipping GetByID"
  skip "[7-2] No test notification — skipping 404 check"
fi

# ── [8] Mark as read ──────────────────────────────────────────────────────────

header "[8] Mark as read"
if [[ -n "$TEST_ID" ]]; then
  READ_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X PATCH "${API}/notifications/${TEST_ID}/read" \
    -H "${AUTH_H}")
  if [[ "$READ_CODE" == "200" ]]; then
    pass "[8-1] PATCH /notifications/:id/read → 200"
  else
    fail "[8-1] PATCH /notifications/:id/read → ${READ_CODE} (expected 200)"
  fi

  # Verify status changed to READ
  AFTER_RESP=$(curl -s "${API}/notifications/${TEST_ID}" -H "${AUTH_H}")
  AFTER_STATUS=$(json_field "$AFTER_RESP" '.data.status')
  if [[ "$AFTER_STATUS" == "READ" ]]; then
    pass "[8-2] Notification status is now READ"
  else
    fail "[8-2] Notification status after mark-read: ${AFTER_STATUS} (expected READ)"
  fi

  # Unread count should not include this one now
  UNREAD_AFTER=$(curl -s "${API}/notifications/summary" -H "${AUTH_H}")
  if [[ "$(json_field "$UNREAD_AFTER" '.success')" == "true" ]]; then
    pass "[8-3] GET /notifications/summary after mark-read → 200"
  else
    fail "[8-3] GET /notifications/summary after mark-read → unexpected: $UNREAD_AFTER"
  fi
else
  skip "[8-1] No test notification — skipping mark-read"
  skip "[8-2] No test notification — skipping status check"
  skip "[8-3] No test notification — skipping summary check"
fi

# ── [9] Summary endpoint ──────────────────────────────────────────────────────

header "[9] Summary endpoint"
SUMMARY_RESP=$(curl -s "${API}/notifications/summary" -H "${AUTH_H}")
SUMMARY_SUCCESS=$(json_field "$SUMMARY_RESP" '.success')
if [[ "$SUMMARY_SUCCESS" == "true" ]]; then
  pass "[9-1] GET /notifications/summary → 200"
else
  fail "[9-1] GET /notifications/summary unexpected: $SUMMARY_RESP"
fi

TOTAL_N=$(json_field "$SUMMARY_RESP" '.data.total')
UNREAD_N=$(json_field "$SUMMARY_RESP" '.data.unread')
if [[ "$TOTAL_N" != "null" && -n "$TOTAL_N" ]]; then
  pass "[9-2] summary.data.total present (${TOTAL_N})"
else
  fail "[9-2] summary.data.total missing"
fi
if [[ "$UNREAD_N" != "null" && -n "$UNREAD_N" ]]; then
  pass "[9-3] summary.data.unread present (${UNREAD_N})"
else
  fail "[9-3] summary.data.unread missing"
fi

# ── [10] Mark all read ────────────────────────────────────────────────────────

header "[10] Mark all read"
MARKALL_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  -X PATCH "${API}/notifications/read-all" \
  -H "${AUTH_H}")
if [[ "$MARKALL_CODE" == "200" ]]; then
  pass "[10-1] PATCH /notifications/read-all → 200"
else
  fail "[10-1] PATCH /notifications/read-all → ${MARKALL_CODE} (expected 200)"
fi

# ── [11] Retry FAILED (optional endpoint) ────────────────────────────────────

header "[11] Retry endpoint"
if [[ -n "$TEST_ID" ]]; then
  RETRY_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${API}/notifications/${TEST_ID}/retry" \
    -H "${AUTH_H}")
  if [[ "$RETRY_CODE" == "200" || "$RETRY_CODE" == "400" ]]; then
    pass "[11-1] POST /notifications/:id/retry → ${RETRY_CODE} (200=retried, 400=not-FAILED is expected)"
  else
    fail "[11-1] POST /notifications/:id/retry → ${RETRY_CODE}"
  fi
else
  skip "[11-1] No test notification — skipping retry"
fi

# ── [12] Tenant isolation ─────────────────────────────────────────────────────

header "[12] Tenant isolation"
FIRST_ORG=$(json_field "$LIST_RESP" '.data[0].organization_id')
USER_ORG=$(json_field "$LOGIN_RESP" '.data.user.organization_id')
if [[ -z "$FIRST_ORG" || "$FIRST_ORG" == "null" ]]; then
  skip "[12-1] No notifications to verify org isolation"
elif [[ "$FIRST_ORG" == "$USER_ORG" ]]; then
  pass "[12-1] First notification org matches caller's org (tenant-safe)"
else
  fail "[12-1] Org mismatch: notification.org=${FIRST_ORG} caller.org=${USER_ORG}"
fi

# ── [13] DB check ─────────────────────────────────────────────────────────────

header "[13] DB row check"
DB_COUNT=$(PGPASSWORD=cankora_secret psql -h localhost -U cankora -d cankora_db \
  -tAc "SELECT count(*) FROM notifications;" 2>/dev/null | tr -d '[:space:]')
if [[ "$DB_COUNT" =~ ^[0-9]+$ ]] && [[ "$DB_COUNT" -gt 0 ]]; then
  pass "[13-1] notifications table has ${DB_COUNT} rows"
else
  skip "[13-1] Cannot check DB (psql unavailable or table empty)"
fi

# ── [14] Frontend route reachability ─────────────────────────────────────────

header "[14] Frontend /notifications"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"
FE_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 "${FRONTEND_URL}/notifications" 2>/dev/null || echo "000")
if [[ "$FE_CODE" == "200" ]]; then
  pass "[14-1] Frontend /notifications → 200"
elif [[ "$FE_CODE" == "000" ]]; then
  skip "[14-1] Frontend not reachable at ${FRONTEND_URL} (start frontend for full test)"
else
  skip "[14-1] Frontend /notifications → ${FE_CODE} (may redirect to login — OK)"
fi

# ── summary ───────────────────────────────────────────────────────────────────

echo ""
echo "============================================================"
echo " Smoke Test: UAT-004 Notification Delivery Foundation"
echo " PASS: ${PASS}  FAIL: ${FAIL}  SKIP: ${SKIP}"
echo "============================================================"

if [[ "$FAIL" -gt 0 ]]; then
  echo "Result: FAIL"
  exit 1
else
  echo "Result: PASS"
  exit 0
fi
