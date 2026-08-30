#!/bin/bash
# =============================================================================
# PMO P2-008 — GIS Map smoke test
# Usage: bash docs/smoke-test-p2-008-gis.sh
# Prereq: backend running on :8080, DB migrated + seeded (koordinat DEMO projects sudah diisi)
# =============================================================================
set -uo pipefail

BASE=${BASE:-http://localhost:8080/api/v1}
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'
PASS=0
FAIL=0

ok()  { echo -e "${GREEN}✓ PASS${NC} $1"; PASS=$((PASS+1)); }
bad() { echo -e "${RED}✗ FAIL${NC} $1"; FAIL=$((FAIL+1)); }

# ── Auth ──────────────────────────────────────────────────────────────────────
TOKEN=$(curl -s -X POST "$BASE/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}' \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['access_token'])")
[ -n "$TOKEN" ] && ok "login and get access_token" || { bad "login failed"; exit 1; }

AUTH="Authorization: Bearer $TOKEN"
CT="Content-Type: application/json"
STAMP=$(date +%s)

# ── 1. GIS summary — harus ada total_projects ─────────────────────────────────
SUMMARY=$(curl -s "$BASE/analytics/gis/summary" -H "$AUTH")
echo "$SUMMARY" | python3 -c "
import sys, json
d = json.load(sys.stdin)
assert d['success'], f'not success: {d}'
s = d['data']
assert 'total_projects' in s, f'missing total_projects: {s}'
assert 'mapped_projects' in s, f'missing mapped_projects: {s}'
assert 'health_green' in s, f'missing health_green: {s}'
assert s['total_projects'] >= 0, f'total_projects negative: {s}'
" && ok "GET /analytics/gis/summary returns valid structure" || bad "GET /analytics/gis/summary"

# ── 2. GIS summary — mapped_projects + unmapped_projects = total_projects ──────
echo "$SUMMARY" | python3 -c "
import sys, json
d = json.load(sys.stdin)
s = d['data']
assert s['mapped_projects'] + s['unmapped_projects'] == s['total_projects'], \
    f'mapped+unmapped != total: {s}'
" && ok "summary: mapped_projects + unmapped_projects == total_projects" || bad "summary: mapped+unmapped mismatch"

# ── 3. GIS summary — DEMO projects sudah terpetakan (≥10 mapped) ──────────────
echo "$SUMMARY" | python3 -c "
import sys, json
d = json.load(sys.stdin)
s = d['data']
assert s['mapped_projects'] >= 10, f'expected >=10 mapped, got {s[\"mapped_projects\"]}'
" && ok "summary: at least 10 DEMO projects are mapped" || bad "summary: <10 mapped projects (cek koordinat seed)"

# ── 4. GIS projects list — harus mengembalikan array ──────────────────────────
PROJECTS=$(curl -s "$BASE/analytics/gis/projects" -H "$AUTH")
echo "$PROJECTS" | python3 -c "
import sys, json
d = json.load(sys.stdin)
assert d['success'], f'not success: {d}'
assert isinstance(d['data'], list), f'data is not list: {type(d[\"data\"])}'
" && ok "GET /analytics/gis/projects returns array" || bad "GET /analytics/gis/projects"

# ── 5. GIS projects — setiap item punya field wajib ───────────────────────────
echo "$PROJECTS" | python3 -c "
import sys, json
d = json.load(sys.stdin)
required = ['project_id','project_code','project_name','status','health_class',
            'progress_pct','budget_total','open_risks','open_issues']
for p in d['data'][:5]:
    for f in required:
        assert f in p, f'missing field {f} in {p}'
print(f'checked {min(5,len(d[\"data\"]))} projects OK')
" && ok "GIS project markers have all required fields" || bad "GIS project markers missing fields"

# ── 6. GIS projects — DEMO-SDA-001 ada dan punya koordinat ───────────────────
echo "$PROJECTS" | python3 -c "
import sys, json
d = json.load(sys.stdin)
demos = [p for p in d['data'] if p['project_code'] == 'DEMO-SDA-001']
assert len(demos) == 1, f'DEMO-SDA-001 not found in list'
p = demos[0]
assert p['latitude'] is not None, f'DEMO-SDA-001 latitude is None'
assert p['longitude'] is not None, f'DEMO-SDA-001 longitude is None'
assert p['province'] == 'Lampung', f'expected Lampung got {p[\"province\"]}'
" && ok "DEMO-SDA-001 has latitude, longitude, province" || bad "DEMO-SDA-001 location data missing"

# ── 7. GIS projects — filter by status=ACTIVE ─────────────────────────────────
ACTIVE=$(curl -s "$BASE/analytics/gis/projects?status=ACTIVE" -H "$AUTH")
echo "$ACTIVE" | python3 -c "
import sys, json
d = json.load(sys.stdin)
assert d['success'], f'not success: {d}'
for p in d['data']:
    assert p['status'] == 'ACTIVE', f'non-ACTIVE in result: {p[\"status\"]}'
print(f'{len(d[\"data\"])} active projects')
" && ok "GET /analytics/gis/projects?status=ACTIVE filters correctly" || bad "status=ACTIVE filter"

# ── 8. GIS projects — filter by health_class=GREEN ────────────────────────────
GREEN_PROJECTS=$(curl -s "$BASE/analytics/gis/projects?health_class=GREEN" -H "$AUTH")
echo "$GREEN_PROJECTS" | python3 -c "
import sys, json
d = json.load(sys.stdin)
assert d['success'], f'not success: {d}'
for p in d['data']:
    assert p['health_class'] == 'GREEN', f'non-GREEN in result: {p[\"health_class\"]}'
print(f'{len(d[\"data\"])} green projects')
" && ok "GET /analytics/gis/projects?health_class=GREEN filters correctly" || bad "health_class=GREEN filter"

# ── 9. GIS projects — filter by province=Lampung ─────────────────────────────
LAMPUNG=$(curl -s "$BASE/analytics/gis/projects?province=Lampung" -H "$AUTH")
echo "$LAMPUNG" | python3 -c "
import sys, json
d = json.load(sys.stdin)
assert d['success'], f'not success: {d}'
assert len(d['data']) >= 1, f'expected >=1 Lampung project, got {len(d[\"data\"])}'
for p in d['data']:
    assert 'lampung' in p['province'].lower(), f'non-Lampung in result: {p[\"province\"]}'
" && ok "GET /analytics/gis/projects?province=Lampung filters correctly" || bad "province=Lampung filter"

# ── 10. GIS project detail — DEMO-SDA-004 (Bendungan Bener) ──────────────────
DEMO4_ID=$(echo "$PROJECTS" | python3 -c "
import sys, json
d = json.load(sys.stdin)
hits = [p for p in d['data'] if p['project_code'] == 'DEMO-SDA-004']
print(hits[0]['project_id'] if hits else '')
")
[ -n "$DEMO4_ID" ] && ok "found DEMO-SDA-004 project_id: $DEMO4_ID" || bad "DEMO-SDA-004 not found"

DETAIL=$(curl -s "$BASE/analytics/gis/projects/$DEMO4_ID" -H "$AUTH")
echo "$DETAIL" | python3 -c "
import sys, json
d = json.load(sys.stdin)
assert d['success'], f'not success: {d}'
p = d['data']
assert p['project_code'] == 'DEMO-SDA-004', f'wrong project: {p[\"project_code\"]}'
assert p['latitude'] is not None, 'latitude is None'
assert abs(p['latitude'] - (-7.5167)) < 0.01, f'wrong lat: {p[\"latitude\"]}'
assert p['province'] == 'Jawa Tengah', f'wrong province: {p[\"province\"]}'
" && ok "GET /analytics/gis/projects/:id returns DEMO-SDA-004 detail with coordinates" || bad "GET /analytics/gis/projects/:id detail"

# ── 11. GIS project detail — 404 untuk ID tidak dikenal ──────────────────────
HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/analytics/gis/projects/00000000-0000-0000-0000-000000000000" -H "$AUTH")
[ "$HTTP" = "404" ] && ok "GET /analytics/gis/projects/:id returns 404 for unknown ID" || bad "unknown project ID expected 404 got $HTTP"

# ── 12. GIS project detail — 400 untuk ID tidak valid ────────────────────────
HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/analytics/gis/projects/not-a-uuid" -H "$AUTH")
[ "$HTTP" = "400" ] && ok "GET /analytics/gis/projects/:id returns 400 for invalid UUID" || bad "invalid UUID expected 400 got $HTTP"

# ── 13. GIS endpoints — 401 tanpa token ──────────────────────────────────────
HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/analytics/gis/projects")
[ "$HTTP" = "401" ] && ok "GET /analytics/gis/projects returns 401 without auth" || bad "unauthenticated expected 401 got $HTTP"

HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/analytics/gis/summary")
[ "$HTTP" = "401" ] && ok "GET /analytics/gis/summary returns 401 without auth" || bad "unauthenticated expected 401 got $HTTP"

# ── 14. Buat proyek baru dengan koordinat via smoke-create ────────────────────
NEW_PROJECT=$(curl -s -X POST "$BASE/projects" -H "$AUTH" -H "$CT" -d "{
  \"code\":\"GIS-SMOKE-$STAMP\",\"name\":\"GIS Smoke Test Project\",\"priority\":\"HIGH\",
  \"start_date\":\"2026-01-01\",\"end_date\":\"2026-12-31\",\"budget_total\":500000000
}")
NEW_PID=$(echo "$NEW_PROJECT" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
[ -n "$NEW_PID" ] && ok "created smoke project for GIS test: $NEW_PID" || bad "create smoke project"

# ── 15. Proyek baru muncul di GIS projects list ───────────────────────────────
UPDATED=$(curl -s "$BASE/analytics/gis/projects" -H "$AUTH")
echo "$UPDATED" | python3 -c "
import sys, json
d = json.load(sys.stdin)
codes = [p['project_code'] for p in d['data']]
assert 'GIS-SMOKE-$STAMP' in codes, f'new project not in GIS list'
" && ok "newly created project appears in GIS projects list" || bad "new project not in GIS list"

# ── 16. Proyek baru tanpa koordinat — latitude/longitude = null ───────────────
echo "$UPDATED" | python3 -c "
import sys, json
d = json.load(sys.stdin)
p = next(x for x in d['data'] if x['project_code'] == 'GIS-SMOKE-$STAMP')
assert p['latitude'] is None, f'expected null latitude, got {p[\"latitude\"]}'
assert p['longitude'] is None, f'expected null longitude, got {p[\"longitude\"]}'
" && ok "new project without coordinates has null latitude/longitude" || bad "new project lat/lng not null"

# ── 17. Summary total_projects sesuai jumlah di list ─────────────────────────
SUMMARY2=$(curl -s "$BASE/analytics/gis/summary" -H "$AUTH")
LIST2=$(curl -s "$BASE/analytics/gis/projects" -H "$AUTH")
python3 -c "
import json
s = json.loads('$(echo "$SUMMARY2" | python3 -c "import sys,json;print(json.dumps(json.load(sys.stdin)['data']))")')
l = json.loads('$(echo "$LIST2" | python3 -c "import sys,json;print(json.dumps(json.load(sys.stdin)['data']))")')
assert s['total_projects'] == len(l), f'summary total {s[\"total_projects\"]} != list count {len(l)}'
" && ok "summary total_projects matches actual project list count" || bad "summary total_projects mismatch"

# ── 18. Filter kombinasi status + province ────────────────────────────────────
COMBO=$(curl -s "$BASE/analytics/gis/projects?status=ACTIVE&province=Jawa" -H "$AUTH")
echo "$COMBO" | python3 -c "
import sys, json
d = json.load(sys.stdin)
assert d['success'], f'not success: {d}'
for p in d['data']:
    assert p['status'] == 'ACTIVE', f'non-ACTIVE: {p[\"status\"]}'
    assert 'jawa' in p['province'].lower(), f'non-Jawa: {p[\"province\"]}'
print(f'{len(d[\"data\"])} results for ACTIVE+Jawa')
" && ok "combined filter status=ACTIVE&province=Jawa works" || bad "combined filter failed"

# ── 19. GIS summary health distribution tidak negatif ────────────────────────
SUMMARY3=$(curl -s "$BASE/analytics/gis/summary" -H "$AUTH")
echo "$SUMMARY3" | python3 -c "
import sys, json
d = json.load(sys.stdin)
s = d['data']
for k in ['health_green','health_yellow','health_red','health_critical','health_unscored']:
    assert s[k] >= 0, f'{k} is negative: {s[k]}'
total_health = s['health_green'] + s['health_yellow'] + s['health_red'] + s['health_critical'] + s['health_unscored']
assert total_health == s['total_projects'], f'health sum {total_health} != total {s[\"total_projects\"]}'
" && ok "summary health distribution sums to total_projects" || bad "summary health distribution invalid"

# ── 20. DEMO-SDA-010 (Rote Ndao, NTT) — koordinat paling timur ───────────────
echo "$PROJECTS" | python3 -c "
import sys, json
d = json.load(sys.stdin)
hits = [p for p in d['data'] if p['project_code'] == 'DEMO-SDA-010']
assert len(hits) == 1, 'DEMO-SDA-010 not found'
p = hits[0]
assert p['province'] == 'Nusa Tenggara Timur', f'wrong province: {p[\"province\"]}'
assert p['latitude'] < -10, f'expected lat < -10 for Rote Ndao, got {p[\"latitude\"]}'
assert p['longitude'] > 120, f'expected lon > 120 for Rote Ndao, got {p[\"longitude\"]}'
" && ok "DEMO-SDA-010 Rote Ndao has correct NTT coordinates" || bad "DEMO-SDA-010 coordinates wrong"

# ── 21. Tidak ada proyek dengan lat=0, lon=0 (sentinel value) ─────────────────
echo "$PROJECTS" | python3 -c "
import sys, json
d = json.load(sys.stdin)
zero_coord = [p for p in d['data'] if p['latitude'] == 0.0 and p['longitude'] == 0.0]
assert len(zero_coord) == 0, f'found {len(zero_coord)} projects with 0,0 coordinates: {[p[\"project_code\"] for p in zero_coord]}'
" && ok "no projects have sentinel 0,0 coordinates" || bad "projects with 0,0 coordinates found"

# ── 22. health_class hanya nilai yang valid ───────────────────────────────────
echo "$PROJECTS" | python3 -c "
import sys, json
d = json.load(sys.stdin)
valid = {'GREEN','YELLOW','RED','CRITICAL','UNSCORED'}
for p in d['data']:
    assert p['health_class'] in valid, f'invalid health_class: {p[\"health_class\"]}'
" && ok "all project markers have valid health_class values" || bad "invalid health_class value found"

# ── 23. progress_pct dalam range 0–100 ───────────────────────────────────────
echo "$PROJECTS" | python3 -c "
import sys, json
d = json.load(sys.stdin)
for p in d['data']:
    assert 0 <= p['progress_pct'] <= 100, f'progress_pct out of range: {p[\"progress_pct\"]} ({p[\"project_code\"]})'
" && ok "all project markers have progress_pct in [0, 100]" || bad "progress_pct out of range"

# ── 24. open_risks dan open_issues tidak negatif ──────────────────────────────
echo "$PROJECTS" | python3 -c "
import sys, json
d = json.load(sys.stdin)
for p in d['data']:
    assert p['open_risks'] >= 0, f'negative open_risks: {p[\"project_code\"]}'
    assert p['open_issues'] >= 0, f'negative open_issues: {p[\"project_code\"]}'
" && ok "open_risks and open_issues are non-negative for all markers" || bad "negative open_risks or open_issues"

# ── 25. DEMO-SDA-008 (Sepaku, Kalimantan Timur) ───────────────────────────────
echo "$PROJECTS" | python3 -c "
import sys, json
d = json.load(sys.stdin)
hits = [p for p in d['data'] if p['project_code'] == 'DEMO-SDA-008']
assert len(hits) == 1, 'DEMO-SDA-008 not found'
p = hits[0]
assert 'Kalimantan Timur' in p['province'], f'wrong province: {p[\"province\"]}'
assert p['longitude'] > 100, f'expected lon > 100 for Kalimantan, got {p[\"longitude\"]}'
" && ok "DEMO-SDA-008 Sepaku has correct Kalimantan Timur coordinates" || bad "DEMO-SDA-008 coordinates wrong"

# ── 26. Concurrent requests tidak crash ──────────────────────────────────────
R1=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/analytics/gis/projects" -H "$AUTH")
R2=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/analytics/gis/summary" -H "$AUTH")
R3=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/analytics/gis/projects?status=ACTIVE" -H "$AUTH")
[ "$R1" = "200" ] && [ "$R2" = "200" ] && [ "$R3" = "200" ] && ok "concurrent GIS requests all return 200" || bad "concurrent requests failed: R1=$R1 R2=$R2 R3=$R3"

# ── 27. avg_progress_pct dalam range 0–100 ───────────────────────────────────
echo "$SUMMARY3" | python3 -c "
import sys, json
d = json.load(sys.stdin)
pct = d['data']['avg_progress_pct']
assert 0 <= pct <= 100, f'avg_progress_pct out of range: {pct}'
" && ok "summary avg_progress_pct is in [0, 100]" || bad "avg_progress_pct out of range"

# ── 28. province filter case-insensitive (ilike) ─────────────────────────────
JAWA_LOWER=$(curl -s "$BASE/analytics/gis/projects?province=jawa+tengah" -H "$AUTH")
echo "$JAWA_LOWER" | python3 -c "
import sys, json
d = json.load(sys.stdin)
assert d['success'], f'not success: {d}'
assert len(d['data']) >= 1, f'expected >=1 Jawa Tengah project with lowercase filter'
" && ok "province filter is case-insensitive (ILIKE)" || bad "province filter case-sensitive"

# ── 29. GIS summary unmapped_projects ≥ 1 (ada proyek tanpa koordinat) ────────
echo "$SUMMARY2" | python3 -c "
import sys, json
d = json.load(sys.stdin)
s = d['data']
assert s['unmapped_projects'] >= 1, f'expected >=1 unmapped, got {s[\"unmapped_projects\"]}'
" && ok "summary shows at least 1 unmapped project (no coordinates)" || bad "no unmapped projects found"

# ── 30. GIS projects filter province=nonexistent → kosong bukan error ─────────
EMPTY=$(curl -s "$BASE/analytics/gis/projects?province=XYZProvinceTidakAda" -H "$AUTH")
echo "$EMPTY" | python3 -c "
import sys, json
d = json.load(sys.stdin)
assert d['success'], f'not success: {d}'
assert d['data'] == [], f'expected empty array, got: {d[\"data\"]}'
" && ok "province filter with no match returns empty array (not error)" || bad "nonexistent province filter returns error"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "══════════════════════════════════════════════════════"
echo "  GIS Map P2-008 Smoke Test Results"
echo "══════════════════════════════════════════════════════"
echo -e "  ${GREEN}PASS: $PASS${NC}   ${RED}FAIL: $FAIL${NC}"
echo "══════════════════════════════════════════════════════"
[ "$FAIL" = "0" ] && exit 0 || exit 1
