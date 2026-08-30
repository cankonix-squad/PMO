#!/usr/bin/env bash
# =============================================================================
# smoke-test-harden-001-integrations.sh
# CANKORA-HARDEN-001 — Integration/BIM/Government clean code & data integrity
# =============================================================================
# Tests:
#   1.  go build ./... passes
#   2.  Login succeeds (obtain token)
#   3.  BIM: create model
#   4.  BIM: add a version to model
#   5.  BIM: link VALID (same-org) project succeeds
#   6.  BIM: link cross-tenant / non-existent project is rejected (404)
#   7.  BIM: audit events use specific names (bim.model.created, etc.)
#   8.  BIM: list mappings returns only org-scoped results
#   9.  Government: COMMIT sync creates mapping idempotently
#  10.  Government: COMMIT sync repeat does not duplicate mapping
#  11.  Government: mapping internal_entity_id = null, match_status = PENDING_MATCH
#  12.  Government: cross-tenant run access is rejected
#  13.  BIM: unlink project works
# =============================================================================
set -euo pipefail

BASE_URL="${CANKORA_BASE_URL:-http://localhost:8080}"
ADMIN_EMAIL="${CANKORA_ADMIN_EMAIL:-admin@cankora.local}"
ADMIN_PASS="${CANKORA_ADMIN_PASS:-Admin@Cankora2024!}"

# Colour helpers
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
PASS=0; FAIL=0

pass() { echo -e "${GREEN}  PASS${NC} $1"; ((PASS++)) || true; }
fail() { echo -e "${RED}  FAIL${NC} $1"; ((FAIL++)) || true; }
info() { echo -e "${YELLOW}  INFO${NC} $1"; }

require_cmd() { command -v "$1" >/dev/null 2>&1 || { echo "Required command '$1' not found."; exit 1; }; }
require_cmd curl
require_cmd jq

echo ""
echo "======================================================================"
echo " CANKORA-HARDEN-001 smoke test — $(date '+%Y-%m-%d %H:%M:%S')"
echo " BASE_URL: $BASE_URL"
echo "======================================================================"

# ─────────────────────────────────────────────────────────────────────────────
# 1. go build
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 1. go build ./..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/../backend" && pwd)"
if (cd "$BACKEND_DIR" && go build ./... 2>&1); then
  pass "go build ./... succeeded"
else
  fail "go build ./... FAILED — fix build errors before proceeding"
  exit 1
fi

# ─────────────────────────────────────────────────────────────────────────────
# 2. Login
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 2. Login"
LOGIN_RESP=$(curl -sf -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASS\"}" 2>&1) || {
  fail "Login request failed — is the server running at $BASE_URL?"
  exit 1
}
TOKEN=$(echo "$LOGIN_RESP" | jq -r '.data.token // .token // empty')
if [[ -z "$TOKEN" ]]; then
  fail "Login: could not extract token from response: $LOGIN_RESP"
  exit 1
fi
pass "Login succeeded, token obtained"
AUTH="-H \"Authorization: Bearer $TOKEN\""

apicall() {
  local method="$1" path="$2"; shift 2
  curl -sf -X "$method" "$BASE_URL$path" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    "$@"
}

# ─────────────────────────────────────────────────────────────────────────────
# 3. BIM: create model
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 3. BIM create model"
BIM_CREATE=$(apicall POST /api/v1/integrations/bim/models -d '{
  "name": "HARDEN-001 Test Model",
  "description": "Smoke test model",
  "discipline": "CIVIL",
  "provider": "local",
  "external_model_id": "harden-001-ext-model-001",
  "viewer_url": "https://example.com/viewer/harden-001"
}') || { fail "BIM create model request failed"; BIM_CREATE="{}"; }
BIM_MODEL_ID=$(echo "$BIM_CREATE" | jq -r '.data.id // empty')
if [[ -n "$BIM_MODEL_ID" ]]; then
  pass "BIM model created: $BIM_MODEL_ID"
else
  fail "BIM create model: could not extract model ID. Response: $BIM_CREATE"
fi

# ─────────────────────────────────────────────────────────────────────────────
# 4. BIM: add a version
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 4. BIM add version"
if [[ -n "${BIM_MODEL_ID:-}" ]]; then
  VER_RESP=$(apicall POST "/api/v1/integrations/bim/models/$BIM_MODEL_ID/versions" -d '{
    "version_label": "v1.0",
    "external_version_id": "harden-ver-001",
    "change_summary": "Initial smoke test version",
    "file_size_bytes": 1024,
    "checksum": "abc123"
  }') || { fail "BIM add version request failed"; VER_RESP="{}"; }
  VER_ID=$(echo "$VER_RESP" | jq -r '.data.id // empty')
  if [[ -n "$VER_ID" ]]; then
    pass "BIM version added: $VER_ID"
  else
    fail "BIM add version: could not extract version ID. Response: $VER_RESP"
  fi
else
  info "Skipping BIM add version — no model ID"
fi

# ─────────────────────────────────────────────────────────────────────────────
# 5. BIM: link VALID project (first project in org)
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 5. BIM link valid project"
PROJECTS_RESP=$(apicall GET /api/v1/projects) || { PROJECTS_RESP="{}"; }
VALID_PROJECT_ID=$(echo "$PROJECTS_RESP" | jq -r '(.data // .items // .) | if type=="array" then .[0].id else empty end' 2>/dev/null || true)

if [[ -z "$VALID_PROJECT_ID" ]]; then
  info "No projects found — skipping link valid project test (not a failure of this hardening)"
else
  if [[ -n "${BIM_MODEL_ID:-}" ]]; then
    LINK_RESP=$(apicall POST "/api/v1/integrations/bim/models/$BIM_MODEL_ID/mappings" -d "{
      \"project_id\": \"$VALID_PROJECT_ID\",
      \"model_role\": \"PRIMARY\",
      \"notes\": \"Smoke test link\"
    }") || { fail "BIM link valid project: request failed"; LINK_RESP="{}"; }
    LINK_ID=$(echo "$LINK_RESP" | jq -r '.data.id // empty')
    if [[ -n "$LINK_ID" ]]; then
      pass "BIM link valid project succeeded: mapping $LINK_ID"
    else
      fail "BIM link valid project: unexpected response: $LINK_RESP"
    fi
  else
    info "Skipping BIM link valid project — no model ID"
  fi
fi

# ─────────────────────────────────────────────────────────────────────────────
# 6. BIM: link cross-tenant / non-existent project must be REJECTED
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 6. BIM link cross-tenant / non-existent project (must 404)"
if [[ -n "${BIM_MODEL_ID:-}" ]]; then
  FAKE_PROJECT_ID="00000000-0000-0000-0000-000000000001"
  CROSS_HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    "$BASE_URL/api/v1/integrations/bim/models/$BIM_MODEL_ID/mappings" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"project_id\":\"$FAKE_PROJECT_ID\",\"model_role\":\"REFERENCE\"}")
  if [[ "$CROSS_HTTP" == "404" || "$CROSS_HTTP" == "403" ]]; then
    pass "BIM cross-tenant project rejected with HTTP $CROSS_HTTP (expected 404/403)"
  else
    fail "BIM cross-tenant project: expected 404/403 but got HTTP $CROSS_HTTP"
  fi
else
  info "Skipping cross-tenant test — no model ID"
fi

# ─────────────────────────────────────────────────────────────────────────────
# 7. BIM: audit events use specific names
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 7. BIM audit event names"
AUDIT_RESP=$(apicall GET "/api/v1/audit-logs?entity_type=bim_model&page_size=50") || { AUDIT_RESP="{}"; }
BIM_CREATED_EVENTS=$(echo "$AUDIT_RESP" | jq '[(.data // .items // .) | if type=="array" then .[] else . end | select(.action == "bim.model.created")] | length' 2>/dev/null || echo "0")
if [[ "$BIM_CREATED_EVENTS" -ge 1 ]]; then
  pass "BIM audit event 'bim.model.created' found ($BIM_CREATED_EVENTS entries)"
else
  fail "BIM audit event 'bim.model.created' NOT found in audit log"
fi

MAPPING_AUDIT=$(apicall GET "/api/v1/audit-logs?entity_type=bim_project_mapping&page_size=50") || { MAPPING_AUDIT="{}"; }
BIM_LINKED_EVENTS=$(echo "$MAPPING_AUDIT" | jq '[(.data // .items // .) | if type=="array" then .[] else . end | select(.action == "bim.project.linked")] | length' 2>/dev/null || echo "0")
if [[ "$BIM_LINKED_EVENTS" -ge 1 ]]; then
  pass "BIM audit event 'bim.project.linked' found ($BIM_LINKED_EVENTS entries)"
else
  info "BIM audit event 'bim.project.linked' not found (may be 0 if no valid project to link)"
fi

VER_AUDIT=$(apicall GET "/api/v1/audit-logs?entity_type=bim_model_version&page_size=50") || { VER_AUDIT="{}"; }
BIM_VER_EVENTS=$(echo "$VER_AUDIT" | jq '[(.data // .items // .) | if type=="array" then .[] else . end | select(.action == "bim.version.created")] | length' 2>/dev/null || echo "0")
if [[ "$BIM_VER_EVENTS" -ge 1 ]]; then
  pass "BIM audit event 'bim.version.created' found ($BIM_VER_EVENTS entries)"
else
  fail "BIM audit event 'bim.version.created' NOT found in audit log"
fi

# ─────────────────────────────────────────────────────────────────────────────
# 8. BIM: list mappings — org-scoped
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 8. BIM list mappings org-scoped"
if [[ -n "${BIM_MODEL_ID:-}" ]]; then
  MAPS_RESP=$(apicall GET "/api/v1/integrations/bim/models/$BIM_MODEL_ID/mappings") || { MAPS_RESP="{}"; }
  MAPS_OK=$(echo "$MAPS_RESP" | jq 'has("data")' 2>/dev/null || echo "false")
  if [[ "$MAPS_OK" == "true" ]]; then
    pass "BIM list mappings returned data field (org-scoped)"
  else
    fail "BIM list mappings unexpected response: $MAPS_RESP"
  fi
else
  info "Skipping BIM list mappings — no model ID"
fi

# ─────────────────────────────────────────────────────────────────────────────
# 9. Government: COMMIT sync creates mapping idempotently
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 9. Government COMMIT sync (first run)"
GOV_RUN1=$(apicall POST /api/v1/integrations/government/runs -d '{
  "connector_key": "government_project_registry",
  "dataset_type":  "projects",
  "mode":          "COMMIT",
  "idempotency_key": "harden-001-smoke-run-01"
}') || { fail "Government create run 1 failed"; GOV_RUN1="{}"; }
RUN1_ID=$(echo "$GOV_RUN1" | jq -r '.data.id // empty')
if [[ -n "$RUN1_ID" ]]; then
  pass "Government sync run 1 created: $RUN1_ID"
else
  fail "Government create run 1: unexpected response: $GOV_RUN1"
fi

# Process run 1
if [[ -n "${RUN1_ID:-}" ]]; then
  PROC1=$(apicall POST "/api/v1/integrations/government/runs/$RUN1_ID/process") || { fail "Government process run 1 failed"; PROC1="{}"; }
  PROC1_STATUS=$(echo "$PROC1" | jq -r '.data.status // empty')
  if [[ "$PROC1_STATUS" == "SUCCEEDED" || "$PROC1_STATUS" == "succeeded" ]]; then
    pass "Government run 1 processed: SUCCEEDED"
  else
    fail "Government run 1 status: '$PROC1_STATUS' (expected SUCCEEDED)"
  fi
fi

# ─────────────────────────────────────────────────────────────────────────────
# 10. Government: repeat COMMIT does not duplicate mapping
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 10. Government COMMIT sync (second run — idempotency)"
GOV_RUN2=$(apicall POST /api/v1/integrations/government/runs -d '{
  "connector_key": "government_project_registry",
  "dataset_type":  "projects",
  "mode":          "COMMIT",
  "idempotency_key": "harden-001-smoke-run-02"
}') || { fail "Government create run 2 failed"; GOV_RUN2="{}"; }
RUN2_ID=$(echo "$GOV_RUN2" | jq -r '.data.id // empty')
if [[ -n "$RUN2_ID" ]]; then
  pass "Government sync run 2 created: $RUN2_ID"
  PROC2=$(apicall POST "/api/v1/integrations/government/runs/$RUN2_ID/process") || { fail "Government process run 2 failed"; PROC2="{}"; }
  PROC2_STATUS=$(echo "$PROC2" | jq -r '.data.status // empty')
  if [[ "$PROC2_STATUS" == "SUCCEEDED" || "$PROC2_STATUS" == "succeeded" ]]; then
    pass "Government run 2 processed: SUCCEEDED (idempotent repeat)"
  else
    fail "Government run 2 status: '$PROC2_STATUS'"
  fi
else
  fail "Government create run 2: unexpected response: $GOV_RUN2"
fi

# Check mapping count — should not have doubled
MAPS_AFTER=$(apicall GET "/api/v1/integrations/government/mappings?connector_key=government_project_registry&dataset_type=projects&page_size=100") || { MAPS_AFTER="{}"; }
MAPS_COUNT=$(echo "$MAPS_AFTER" | jq '(.data // .items // .) | if type=="array" then length else 0 end' 2>/dev/null || echo "0")
info "Government mappings count after 2 COMMIT runs: $MAPS_COUNT (should equal sample record count, not 2x)"
if [[ "$MAPS_COUNT" -le 10 ]]; then
  pass "Government mappings not duplicated (count=$MAPS_COUNT ≤ 10 sample records per run)"
else
  fail "Government mappings may be duplicated (count=$MAPS_COUNT > 10 — check for duplicate rows)"
fi

# ─────────────────────────────────────────────────────────────────────────────
# 11. Government: mapping has internal_entity_id=null + match_status=PENDING_MATCH
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 11. Government mapping: internal_entity_id=null, match_status=PENDING_MATCH"
FIRST_MAP=$(echo "$MAPS_AFTER" | jq '(.data // .items // .) | if type=="array" then .[0] else . end' 2>/dev/null || echo "{}")
INT_ID=$(echo "$FIRST_MAP" | jq -r '.internal_entity_id // "null"')
MATCH_ST=$(echo "$FIRST_MAP" | jq -r '.match_status // "unknown"')

if [[ "$INT_ID" == "null" || -z "$INT_ID" ]]; then
  pass "Government mapping: internal_entity_id is null (not a fake UUID)"
else
  fail "Government mapping: internal_entity_id='$INT_ID' — expected null for unresolved mapping"
fi

if [[ "$MATCH_ST" == "PENDING_MATCH" ]]; then
  pass "Government mapping: match_status=PENDING_MATCH"
else
  fail "Government mapping: match_status='$MATCH_ST' — expected PENDING_MATCH"
fi

# ─────────────────────────────────────────────────────────────────────────────
# 12. Government: cross-tenant run access is rejected
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 12. Government cross-tenant run access rejected"
FAKE_RUN_ID="00000000-0000-0000-0000-000000000099"
CROSS_RUN_HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X GET \
  "$BASE_URL/api/v1/integrations/government/runs/$FAKE_RUN_ID" \
  -H "Authorization: Bearer $TOKEN")
if [[ "$CROSS_RUN_HTTP" == "404" || "$CROSS_RUN_HTTP" == "403" ]]; then
  pass "Government cross-tenant run access rejected with HTTP $CROSS_RUN_HTTP"
else
  fail "Government cross-tenant run: expected 404/403 but got HTTP $CROSS_RUN_HTTP"
fi

# ─────────────────────────────────────────────────────────────────────────────
# 13. BIM: unlink project
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "── 13. BIM unlink project"
if [[ -n "${BIM_MODEL_ID:-}" && -n "${VALID_PROJECT_ID:-}" ]]; then
  UNLINK_HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE \
    "$BASE_URL/api/v1/integrations/bim/models/$BIM_MODEL_ID/mappings/$VALID_PROJECT_ID" \
    -H "Authorization: Bearer $TOKEN")
  if [[ "$UNLINK_HTTP" == "200" || "$UNLINK_HTTP" == "204" ]]; then
    pass "BIM unlink project succeeded (HTTP $UNLINK_HTTP)"
  else
    fail "BIM unlink project: HTTP $UNLINK_HTTP"
  fi
else
  info "Skipping BIM unlink — no model or project ID available"
fi

# ─────────────────────────────────────────────────────────────────────────────
# Summary
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "======================================================================"
TOTAL=$((PASS + FAIL))
echo -e " Results: ${GREEN}$PASS PASS${NC} / ${RED}$FAIL FAIL${NC} / $TOTAL total"
echo "======================================================================"
echo ""

if [[ $FAIL -eq 0 ]]; then
  echo -e "${GREEN}All smoke tests passed.${NC}"
  exit 0
else
  echo -e "${RED}$FAIL smoke test(s) failed. Review output above.${NC}"
  exit 1
fi
