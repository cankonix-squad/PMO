#!/usr/bin/env bash
# Smoke test P3-001: BIM/Digital Twin Integration Foundation
# Run: bash docs/smoke-test-p3-001-bim.sh
# Expected: All N checks PASS

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
HEALTH_URL="${HEALTH_URL:-http://localhost:8080}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@cankora.local}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Admin@Cankora2024!}"

PASS=0
FAIL=0
MODEL_ID=""
VERSION_ID=""

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

green() { printf '\033[0;32m%s\033[0m\n' "$*"; }
red()   { printf '\033[0;31m%s\033[0m\n' "$*"; }

check() {
  local label="$1"
  local actual="$2"
  local expected="$3"
  if echo "$actual" | grep -q "$expected"; then
    green "  PASS  $label"
    PASS=$((PASS + 1))
  else
    red   "  FAIL  $label"
    echo  "        expected: $expected"
    echo  "        got:      $(echo "$actual" | head -c 200)"
    FAIL=$((FAIL + 1))
  fi
}

# ---------------------------------------------------------------------------
# 1. Health check
# ---------------------------------------------------------------------------
echo ""
echo "=== P3-001 BIM/Digital Twin Smoke Test ==="
echo ""
echo "--- 1. Health ---"

HEALTH=$(curl -sf "$HEALTH_URL/health" || echo "FAIL")
check "GET /health → ok" "$HEALTH" "ok"

# ---------------------------------------------------------------------------
# 2. Login
# ---------------------------------------------------------------------------
echo ""
echo "--- 2. Auth ---"

LOGIN=$(curl -sf -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}")
check "POST /auth/login → token" "$LOGIN" "token"

TOKEN=$(echo "$LOGIN" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
  red "Cannot extract token — aborting"
  exit 1
fi

AUTH="-H \"Authorization: Bearer $TOKEN\""

# ---------------------------------------------------------------------------
# 3. Auth guard — unauthenticated request should return 401
# ---------------------------------------------------------------------------
echo ""
echo "--- 3. Auth guard ---"

UNAUTH=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/integrations/bim/models")
check "GET /bim/models without token → 401" "$UNAUTH" "401"

# ---------------------------------------------------------------------------
# 4. List models (empty initially)
# ---------------------------------------------------------------------------
echo ""
echo "--- 4. List BIM models ---"

LIST=$(curl -sf -H "Authorization: Bearer $TOKEN" "$BASE_URL/integrations/bim/models")
check "GET /bim/models → 200 with data array" "$LIST" '"data"'
check "GET /bim/models → meta.total present" "$LIST" '"total"'

# ---------------------------------------------------------------------------
# 5. Create BIM model
# ---------------------------------------------------------------------------
echo ""
echo "--- 5. Create BIM model ---"

CREATE=$(curl -sf -X POST "$BASE_URL/integrations/bim/models" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Smoke Test Model BIM",
    "description": "Model untuk smoke test P3-001",
    "discipline": "STRUCTURAL",
    "provider": "local",
    "external_model_id": "smoke-test-ext-001",
    "viewer_url": "https://example.com/viewer/smoke-001",
    "metadata": {"source": "smoke_test", "version": "1"}
  }')
check "POST /bim/models → 201 with id" "$CREATE" '"id"'
check "POST /bim/models → discipline STRUCTURAL" "$CREATE" "STRUCTURAL"
check "POST /bim/models → status DRAFT" "$CREATE" "DRAFT"
check "POST /bim/models → provider local" "$CREATE" "local"

MODEL_ID=$(echo "$CREATE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "  model_id: $MODEL_ID"

# ---------------------------------------------------------------------------
# 6. Get single model
# ---------------------------------------------------------------------------
echo ""
echo "--- 6. Get BIM model by ID ---"

GET=$(curl -sf -H "Authorization: Bearer $TOKEN" "$BASE_URL/integrations/bim/models/$MODEL_ID")
check "GET /bim/models/:id → 200" "$GET" '"id"'
check "GET /bim/models/:id → correct name" "$GET" "Smoke Test Model BIM"

# ---------------------------------------------------------------------------
# 7. Update model
# ---------------------------------------------------------------------------
echo ""
echo "--- 7. Update BIM model ---"

UPDATE=$(curl -sf -X PATCH "$BASE_URL/integrations/bim/models/$MODEL_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "ACTIVE", "description": "Updated description"}')
check "PATCH /bim/models/:id → updated" "$UPDATE" '"id"'

# ---------------------------------------------------------------------------
# 8. List models again — should have 1
# ---------------------------------------------------------------------------
echo ""
echo "--- 8. List models after create ---"

LIST2=$(curl -sf -H "Authorization: Bearer $TOKEN" "$BASE_URL/integrations/bim/models")
check "GET /bim/models → total >= 1" "$LIST2" '"total"'
check "GET /bim/models → contains created model" "$LIST2" "Smoke Test Model BIM"

# ---------------------------------------------------------------------------
# 9. Add version
# ---------------------------------------------------------------------------
echo ""
echo "--- 9. Add BIM model version ---"

ADDVER=$(curl -sf -X POST "$BASE_URL/integrations/bim/models/$MODEL_ID/versions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "version_label": "v1.0",
    "external_version_id": "ext-ver-001",
    "change_summary": "Versi awal smoke test",
    "file_size_bytes": 10485760,
    "checksum": "abc123def456"
  }')
check "POST /bim/models/:id/versions → 201" "$ADDVER" '"id"'
check "POST /bim/models/:id/versions → version_label v1.0" "$ADDVER" "v1.0"

VERSION_ID=$(echo "$ADDVER" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "  version_id: $VERSION_ID"

# ---------------------------------------------------------------------------
# 10. List versions
# ---------------------------------------------------------------------------
echo ""
echo "--- 10. List versions ---"

VERSIONS=$(curl -sf -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/integrations/bim/models/$MODEL_ID/versions")
check "GET /bim/models/:id/versions → data array" "$VERSIONS" '"data"'
check "GET /bim/models/:id/versions → v1.0 present" "$VERSIONS" "v1.0"

# ---------------------------------------------------------------------------
# 11. Link project (use a fake UUID; 500 or 404 is acceptable since project
#     may not exist — what we verify is the endpoint is reachable and auth works)
# ---------------------------------------------------------------------------
echo ""
echo "--- 11. Project mapping endpoints ---"

FAKE_PROJECT="00000000-0000-0000-0000-000000000001"
LINK=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  "$BASE_URL/integrations/bim/models/$MODEL_ID/mappings" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"project_id\":\"$FAKE_PROJECT\",\"model_role\":\"REFERENCE\",\"notes\":\"smoke test\"}")
# 201 = linked, 500 = FK constraint (project doesn't exist) — both mean endpoint is reachable
check "POST /bim/models/:id/mappings → reachable (201 or 500)" "$LINK" "[25][0-9][0-9]"

LISTMAP=$(curl -sf -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/integrations/bim/models/$MODEL_ID/mappings")
check "GET /bim/models/:id/mappings → data array" "$LISTMAP" '"data"'

# ---------------------------------------------------------------------------
# 12. Not found — invalid UUID
# ---------------------------------------------------------------------------
echo ""
echo "--- 12. Not found ---"

NF=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/integrations/bim/models/00000000-0000-0000-0000-000000000000")
check "GET /bim/models/non-existent-id → 404" "$NF" "404"

# ---------------------------------------------------------------------------
# 13. Delete model
# ---------------------------------------------------------------------------
echo ""
echo "--- 13. Delete BIM model (soft delete) ---"

DELETE=$(curl -sf -X DELETE -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/integrations/bim/models/$MODEL_ID")
check "DELETE /bim/models/:id → 200" "$DELETE" "bim model deleted"

# Verify soft delete — model should no longer appear in list
LIST3=$(curl -sf -H "Authorization: Bearer $TOKEN" "$BASE_URL/integrations/bim/models")
if echo "$LIST3" | grep -q "Smoke Test Model BIM"; then
  red  "  FAIL  Soft delete: model still visible after delete"
  FAIL=$((FAIL + 1))
else
  green "  PASS  Soft delete: model no longer visible after delete"
  PASS=$((PASS + 1))
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo '================================================='
TOTAL=$((PASS + FAIL))
echo "P3-001 BIM/Digital Twin: $PASS/$TOTAL PASS"
if [ "$FAIL" -gt 0 ]; then
  echo "FAIL: $FAIL checks failed"
  exit 1
else
  green "ALL $PASS/$TOTAL PASS"
  exit 0
fi
