#!/bin/bash
# =============================================================================
# PMO P0-013 — Local End-to-End Smoke Test (documented artifact)
# Usage:  ./docs/smoke-test-p0-013.sh
# Prereq: backend running on :8080, frontend on :3000, DB migrated + seeded
# Result: all checks print "PASS"; exit code 0 = all green
#
# P0-014 hardening (2026-08-20):
#   - Cleanup deletes child rows (sub-task) BEFORE the parent task, and
#     verifies that deleting a project cascades soft-delete to its children.
#   - No orphan active child rows are left behind in soft-deleted projects.
# =============================================================================
set -e
BASE=http://localhost:8080/api/v1
GREEN='\033[0;32m'; RED='\033[0;31m'; NC='\033[0m'
PASS=0; FAIL=0
ok()   { echo -e "${GREEN}✓ PASS${NC} $1"; PASS=$((PASS+1)); }
bad()  { echo -e "${RED}✗ FAIL${NC} $1"; FAIL=$((FAIL+1)); }
EMAIL="smoke.ver.$(date +%s)@cankora.local"
CODE="PRJ-VER-$(date +%s)"

echo "== 1. Login =="
TOKEN=$(curl -s -X POST $BASE/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['access_token'])")
[ -n "$TOKEN" ] && ok "login admin dapat token" || bad "login admin"
AUTH="Authorization: Bearer $TOKEN"
CT="Content-Type: application/json"

echo "== 2. Dashboard =="
curl -s $BASE/dashboard -H "$AUTH" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success']" && ok "GET /dashboard" || bad "GET /dashboard"

echo "== 3. Project CRUD =="
PROJ=$(curl -s -X POST $BASE/projects -H "$AUTH" -H "$CT" -d "{
  \"code\":\"$CODE\",\"name\":\"Smoke Verified P0-013\",
  \"description\":\"verified smoke\",\"priority\":\"HIGH\",\"category\":\"Testing\",
  \"start_date\":\"2026-09-01\",\"end_date\":\"2026-12-31\",
  \"budget_total\":250000000,\"currency\":\"IDR\"}")
PID=$(echo "$PROJ" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
[ -n "$PID" ] && ok "create project ($PID)" || bad "create project: $PROJ"
curl -s $BASE/projects/$PID -H "$AUTH" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['code']=='$CODE'" && ok "get project by id" || bad "get project"
curl -s -X PUT $BASE/projects/$PID -H "$AUTH" -H "$CT" -d '{"name":"Smoke Verified P0-013 Updated","priority":"CRITICAL"}' | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['name'].endswith('Updated')" && ok "update project (name+priority)" || bad "update project"
curl -s "$BASE/projects?page=1&page_size=5" -H "$AUTH" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and len(d['data'])>=1" && ok "list projects paginated" || bad "list projects"
# duplicate code → 409
CODE2=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE/projects -H "$AUTH" -H "$CT" -d "{\"code\":\"$CODE\",\"name\":\"Duplicate\"}")
[ "$CODE2" = "409" ] && ok "duplicate project code → 409 Conflict" || bad "duplicate code (got $CODE2)"

echo "== 4. Project transition =="
curl -s -X POST $BASE/projects/$PID/transition -H "$AUTH" -H "$CT" -d '{"to_status":"PLANNING"}' | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['status']=='PLANNING'" && ok "transition DRAFT→PLANNING" || bad "transition PLANNING"
curl -s -X POST $BASE/projects/$PID/transition -H "$AUTH" -H "$CT" -d '{"to_status":"ACTIVE"}' | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['status']=='ACTIVE'" && ok "transition PLANNING→ACTIVE" || bad "transition ACTIVE"
curl -s -X POST $BASE/projects/$PID/transition -H "$AUTH" -H "$CT" -d '{"to_status":"CANCELLED"}' | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['status']=='CANCELLED'" && ok "transition ACTIVE→CANCELLED (valid FSM)" || bad "transition CANCELLED"
curl -s -X POST $BASE/projects/$PID/transition -H "$AUTH" -H "$CT" -d '{"to_status":"PLANNING"}' | python3 -c "import sys,json; d=json.load(sys.stdin); assert not d['success']" && ok "transition invalid ditolak FSM (CANCELLED→PLANNING)" || bad "transition invalid seharusnya ditolak"

echo "== 5. Progress history =="
curl -s -X PUT $BASE/projects/$PID -H "$AUTH" -H "$CT" -d '{"progress_pct":35}' > /dev/null
curl -s $BASE/projects/$PID/progress-history -H "$AUTH" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success']" && ok "GET progress-history" || bad "progress-history"

echo "== 6. Milestone CRUD =="
MS=$(curl -s -X POST $BASE/projects/$PID/milestones -H "$AUTH" -H "$CT" -d '{"title":"M1 Initiation","due_date":"2026-10-01"}')
MSID=$(echo "$MS" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
[ -n "$MSID" ] && ok "create milestone ($MSID)" || bad "create milestone: $MS"
curl -s -X PUT $BASE/projects/$PID/milestones/$MSID -H "$AUTH" -H "$CT" -d '{"title":"M1 Initiation (updated)","status":"IN_PROGRESS","progress_pct":20}' | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['status']=='IN_PROGRESS'" && ok "update milestone status+progress" || bad "update milestone"
curl -s $BASE/projects/$PID/milestones -H "$AUTH" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and len(d['data'])>=1" && ok "list milestones" || bad "list milestones"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE $BASE/projects/$PID/milestones/$MSID -H "$AUTH")
[ "$CODE" = "204" ] && ok "soft delete milestone (HTTP 204)" || bad "delete milestone (got $CODE)"

echo "== 7. Task CRUD =="
TK=$(curl -s -X POST $BASE/projects/$PID/tasks -H "$AUTH" -H "$CT" -d '{"title":"T1 Requirement Analysis","description":"gather requirements","priority":"HIGH","start_date":"2026-09-01","due_date":"2026-09-30"}')
TKID=$(echo "$TK" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
[ -n "$TKID" ] && ok "create task ($TKID)" || bad "create task: $TK"
curl -s -X PUT $BASE/projects/$PID/tasks/$TKID -H "$AUTH" -H "$CT" -d '{"title":"T1 Requirement Analysis","status":"IN_PROGRESS","progress_pct":50}' | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['progress_pct']==50" && ok "update task progress 50%" || bad "update task"
STK=$(curl -s -X POST $BASE/projects/$PID/tasks -H "$AUTH" -H "$CT" -d "{\"title\":\"T1.1 Sub Analysis\",\"parent_id\":\"$TKID\",\"wbs_code\":\"1.1\"}")
STKID=$(echo "$STK" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
[ -n "$STKID" ] && ok "create sub-task (WBS level 2)" || bad "create sub-task: $STK"
curl -s $BASE/projects/$PID/tasks -H "$AUTH" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and len(d['data'])>=1" && ok "list tasks" || bad "list tasks"
# Delete child FIRST, then parent (P0-014: cleanup order)
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE $BASE/projects/$PID/tasks/$STKID -H "$AUTH")
[ "$CODE" = "204" ] && ok "soft delete sub-task (HTTP 204)" || bad "delete sub-task (got $CODE)"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE $BASE/projects/$PID/tasks/$TKID -H "$AUTH")
[ "$CODE" = "204" ] && ok "soft delete task (HTTP 204)" || bad "delete task (got $CODE)"

echo "== 8. User management =="
U=$(curl -s -X POST $BASE/users -H "$AUTH" -H "$CT" -d "{\"first_name\":\"Smoke\",\"last_name\":\"User\",\"email\":\"$EMAIL\",\"password\":\"Smoke1234!\",\"job_title\":\"Analyst\"}")
USERID=$(echo "$U" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
[ -n "$USERID" ] && ok "create user ($EMAIL → $USERID)" || bad "create user: $U"
curl -s -X PUT $BASE/users/$USERID -H "$AUTH" -H "$CT" -d '{"first_name":"Smoke","last_name":"User Updated","job_title":"Senior Analyst"}' | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['last_name']=='User Updated'" && ok "update user profile" || bad "update user"
curl -s $BASE/users/$USERID -H "$AUTH" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['email']=='$EMAIL'" && ok "get user by id" || bad "get user"
curl -s -X POST $BASE/users/$USERID/deactivate -H "$AUTH" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['message']=='user deactivated'" && ok "deactivate user (message OK)" || bad "deactivate user"
LOGIN_DEACT=$(curl -s -X POST $BASE/auth/login -H "$CT" -d "{\"email\":\"$EMAIL\",\"password\":\"Smoke1234!\"}")
echo "$LOGIN_DEACT" | python3 -c "import sys,json; d=json.load(sys.stdin); assert not d['success']" && ok "user deactivated tidak bisa login" || bad "deactivated user masih bisa login"
curl -s "$BASE/users?page=1&page_size=5" -H "$AUTH" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and len(d['data'])>0" && ok "list users paginated" || bad "list users"

echo "== 9. Auth me + logout =="
curl -s $BASE/auth/me -H "$AUTH" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success'] and d['data']['email']=='admin@cankora.local'" && ok "GET /auth/me" || bad "auth/me"
curl -s -X POST $BASE/auth/logout -H "$AUTH" -H "$CT" -d '{}' | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['success']" && ok "logout" || bad "logout"

echo "== 10. Cleanup =="
TOKEN2=$(curl -s -X POST $BASE/auth/login -H 'Content-Type: application/json' -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['access_token'])")
AUTH2="Authorization: Bearer $TOKEN2"
# Create a child task to verify project delete cascades to it (P0-014)
CHILD_TK=$(curl -s -X POST $BASE/projects/$PID/tasks -H "$AUTH2" -H "$CT" -d '{"title":"T-Cascade Check","priority":"LOW"}')
CHILD_TKID=$(echo "$CHILD_TK" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
[ -n "$CHILD_TKID" ] && ok "create task untuk verifikasi cascade" || bad "create cascade task: $CHILD_TK"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE $BASE/projects/$PID -H "$AUTH2")
[ "$CODE" = "204" ] && ok "soft delete project cleanup (HTTP 204)" || bad "delete project (got $CODE)"
# After project delete: child task must be soft-deleted too (no orphan)
CHILD_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE/projects/$PID/tasks/$CHILD_TKID -H "$AUTH2")
[ "$CHILD_STATUS" = "404" ] && ok "child task tidak lagi diakses setelah project delete (cascade OK)" || bad "child task masih bisa diakses setelah project delete (got $CHILD_STATUS)"

echo ""
echo "======================================"
echo "SMOKE RESULT VERIFIED: $PASS passed, $FAIL failed"
echo "======================================"
[ $FAIL -eq 0 ]
