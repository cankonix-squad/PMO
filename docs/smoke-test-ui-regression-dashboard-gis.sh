#!/usr/bin/env bash
# smoke-test-ui-regression-dashboard-gis.sh
# Verifikasi UI regression: /dashboard, /dashboard#monitoring, /gis, /governance
# Pastikan sidebar/topbar tampil, map bukan full-screen, tidak ada horizontal overflow.
#
# Prasyarat:
#   - Frontend dev server jalan di :3000
#   - Backend API jalan di :8080
#   - Node.js terinstall (untuk Playwright runner inline)
#   npm install -g playwright / npx playwright install chromium
#
# Cara pakai:
#   chmod +x docs/smoke-test-ui-regression-dashboard-gis.sh
#   ./docs/smoke-test-ui-regression-dashboard-gis.sh
#
# Exit code: 0 = semua PASS, 1 = ada FAIL

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:3000}"
API_URL="${API_URL:-http://localhost:8080}"
EMAIL="${SMOKE_EMAIL:-admin@cankora.local}"
PASS="${SMOKE_PASS:-Admin@Cankora2024!}"

PASS_COUNT=0
FAIL_COUNT=0
ERRORS=()

pass() { echo "  ✅ PASS: $1"; ((PASS_COUNT++)); return 0; }
fail() { echo "  ❌ FAIL: $1"; ERRORS+=("$1"); ((FAIL_COUNT++)); return 0; }

section() { echo ""; echo "▶ $1"; }

# ── 1. Health checks ──────────────────────────────────────────────────────────
section "1. Health checks"

HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/health" 2>/dev/null || echo "000")
[[ "$HTTP" == "200" ]] && pass "Backend $API_URL/health → 200" || fail "Backend health returned $HTTP (expected 200)"

HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL" 2>/dev/null || echo "000")
[[ "$HTTP" == "200" || "$HTTP" == "307" || "$HTTP" == "308" ]] && pass "Frontend $BASE_URL → $HTTP" || fail "Frontend returned $HTTP"

# ── 2. Auth ───────────────────────────────────────────────────────────────────
section "2. Auth — login & token"

RESP=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}" 2>/dev/null)

TOKEN=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('access_token',''))" 2>/dev/null || echo "")
[[ -n "$TOKEN" ]] && pass "Login berhasil, token diterima" || fail "Login gagal — cek email/password atau endpoint"

# ── 3. Playwright UI checks ───────────────────────────────────────────────────
section "3. Playwright UI — layout & overflow checks"

if ! command -v node &>/dev/null; then
  fail "Node.js tidak ditemukan — skip Playwright checks"
else
  # Write Playwright script into frontend dir so ESM can resolve playwright from node_modules
  REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
  FRONTEND_DIR="$REPO_ROOT/frontend"
  TMP_SCRIPT=$(mktemp "$FRONTEND_DIR/smoke-gis-XXXXXX.mjs")
  trap "rm -f '$TMP_SCRIPT'" EXIT

  cat > "$TMP_SCRIPT" << 'PLAYWRIGHT_EOF'
import { chromium } from 'playwright';

const BASE = process.env.BASE_URL || 'http://localhost:3000';
const EMAIL = process.env.SMOKE_EMAIL || 'admin@cankora.local';
const PASS  = process.env.SMOKE_PASS  || 'Admin@Cankora2024!';

const VIEWPORTS = [
  { name: 'desktop-1366', width: 1366, height: 768 },
  { name: 'desktop-1920', width: 1920, height: 1080 },
  { name: 'mobile-390',   width: 390,  height: 844 },
];

let passCount = 0, failCount = 0;
const errors = [];

function pass(msg)  { console.log(`  ✅ PASS: ${msg}`); passCount++; }
function fail(msg)  { console.error(`  ❌ FAIL: ${msg}`); errors.push(msg); failCount++; }

async function checkNoHOverflow(page, label) {
  const { sw, cw } = await page.evaluate(() => ({
    sw: document.documentElement.scrollWidth,
    cw: document.documentElement.clientWidth,
  }));
  if (sw <= cw + 1) pass(`${label} — no horizontal overflow (scrollWidth ${sw} ≤ clientWidth ${cw})`);
  else fail(`${label} — horizontal overflow: scrollWidth ${sw} > clientWidth ${cw}`);
}

async function checkLayoutPresent(page, label) {
  const sidebar = await page.evaluate(() => {
    const el = document.querySelector('aside, [class*="sidebar"], nav[aria-label]');
    if (!el) return null;
    const r = el.getBoundingClientRect();
    return { w: Math.round(r.width), visible: r.width > 0 && r.height > 0 };
  });
  if (sidebar && sidebar.visible) pass(`${label} — sidebar present (w=${sidebar.w}px)`);
  else fail(`${label} — sidebar NOT present`);

  const topbar = await page.evaluate(() => {
    const el = document.querySelector('header, [class*="topbar"], [class*="top-bar"]');
    if (!el) return null;
    const r = el.getBoundingClientRect();
    return { h: Math.round(r.height), visible: r.width > 0 && r.height > 0 };
  });
  if (topbar && topbar.visible) pass(`${label} — topbar present (h=${topbar.h}px)`);
  else fail(`${label} — topbar NOT present`);
}

async function checkMapNotFullScreen(page, label) {
  const result = await page.evaluate(() => {
    const imgs = Array.from(document.querySelectorAll('img'));
    const mapImg = imgs.find(i => i.src.includes('indonesia') || i.src.includes('map'));
    if (!mapImg) {
      // Check Leaflet map
      const leaflet = document.querySelector('.leaflet-container');
      if (!leaflet) return { type: 'none', ok: true };
      const r = leaflet.getBoundingClientRect();
      const vw = document.documentElement.clientWidth;
      const vh = document.documentElement.clientHeight;
      return { type: 'leaflet', w: Math.round(r.width), h: Math.round(r.height), vw, vh,
        ok: r.width < vw * 0.95 || r.height < vh * 0.95 };
    }
    const r = mapImg.getBoundingClientRect();
    const vw = document.documentElement.clientWidth;
    const vh = document.documentElement.clientHeight;
    return { type: 'img', w: Math.round(r.width), h: Math.round(r.height), vw, vh,
      ok: r.width < vw * 0.95 || r.height < vh * 0.95 };
  });
  if (result.ok) pass(`${label} — map NOT full-screen (${result.type}: ${result.w}x${result.h})`);
  else fail(`${label} — map IS full-screen (${result.type}: ${result.w}x${result.h} vs ${result.vw}x${result.vh})`);
}

async function checkMonitoringAnchor(page) {
  const result = await page.evaluate(() => {
    const el = document.getElementById('monitoring');
    if (!el) return { found: false };
    const r = el.getBoundingClientRect();
    return { found: true, top: Math.round(r.top), visible: r.top < window.innerHeight };
  });
  if (result.found) pass(`/dashboard#monitoring — anchor id="monitoring" exists`);
  else fail(`/dashboard#monitoring — anchor id="monitoring" NOT found`);
}

async function checkGISMapVisible(page) {
  await page.waitForTimeout(2000); // wait for Leaflet init
  const result = await page.evaluate(() => {
    const el = document.querySelector('.leaflet-container');
    if (!el) return { found: false };
    const r = el.getBoundingClientRect();
    return { found: true, w: Math.round(r.width), h: Math.round(r.height), visible: r.width > 100 && r.height > 100 };
  });
  if (result.found && result.visible) pass(`/gis — Leaflet map visible (${result.w}x${result.h})`);
  else if (result.found) fail(`/gis — Leaflet map found but too small (${result.w}x${result.h})`);
  else fail(`/gis — Leaflet .leaflet-container NOT found`);
}

async function checkNoFatalError(page, label) {
  const errors_js = await page.evaluate(() => {
    return window.__smokeErrors || [];
  });
  // Also check for visible error messages in DOM
  const hasError = await page.evaluate(() => {
    const body = document.body.innerText || '';
    return /Application error|Unhandled Runtime Error|500 Internal/.test(body);
  });
  if (hasError) fail(`${label} — fatal JS/render error visible in DOM`);
  else pass(`${label} — no fatal error in DOM`);
}

// ── Main ──────────────────────────────────────────────────────────────────────
const browser = await chromium.launch({ headless: true });

// Login via UI form to get proper auth cookie (middleware reads cankora-auth cookie)
const context = await browser.newContext({ viewport: { width: 1366, height: 768 } });
const authPage = await context.newPage();
await authPage.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded', timeout: 20000 });
await authPage.waitForTimeout(1000);
await authPage.fill('input[type="email"], input[name="email"]', EMAIL);
await authPage.fill('input[type="password"], input[name="password"]', PASS);
await authPage.click('button[type="submit"]');
try {
  await authPage.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 15000 });
} catch (_) { /* redirect may not be immediate; storage state is what matters */ }
await authPage.waitForTimeout(1000);

// ── Per-viewport tests ────────────────────────────────────────────────────────
const storage = await context.storageState();
for (const vp of VIEWPORTS) {
  console.log(`\n  📐 Viewport: ${vp.name} (${vp.width}x${vp.height})`);
  const vpContext = await browser.newContext({ viewport: { width: vp.width, height: vp.height }, storageState: storage });
  const vpPage = await vpContext.newPage();

  const isMobile = vp.width < 768;

  // /dashboard
  console.log(`\n    🔍 ${BASE}/dashboard`);
  await vpPage.goto(`${BASE}/dashboard`, { waitUntil: 'domcontentloaded', timeout: 20000 });
  await vpPage.waitForTimeout(1000);
  if (!isMobile) await checkLayoutPresent(vpPage, `/dashboard [${vp.name}]`);
  await checkMapNotFullScreen(vpPage, `/dashboard [${vp.name}]`);
  await checkNoHOverflow(vpPage, `/dashboard [${vp.name}]`);
  await checkNoFatalError(vpPage, `/dashboard [${vp.name}]`);

  // /dashboard#monitoring
  console.log(`\n    🔍 ${BASE}/dashboard#monitoring`);
  await vpPage.goto(`${BASE}/dashboard#monitoring`, { waitUntil: 'domcontentloaded', timeout: 20000 });
  await vpPage.waitForTimeout(500);
  await checkMonitoringAnchor(vpPage);
  await checkMapNotFullScreen(vpPage, `/dashboard#monitoring [${vp.name}]`);
  await checkNoHOverflow(vpPage, `/dashboard#monitoring [${vp.name}]`);

  // /gis
  console.log(`\n    🔍 ${BASE}/gis`);
  await vpPage.goto(`${BASE}/gis`, { waitUntil: 'domcontentloaded', timeout: 20000 });
  await vpPage.waitForTimeout(2500);
  if (!isMobile) await checkLayoutPresent(vpPage, `/gis [${vp.name}]`);
  await checkGISMapVisible(vpPage);
  await checkNoHOverflow(vpPage, `/gis [${vp.name}]`);
  await checkNoFatalError(vpPage, `/gis [${vp.name}]`);

  // /governance
  console.log(`\n    🔍 ${BASE}/governance`);
  await vpPage.goto(`${BASE}/governance`, { waitUntil: 'domcontentloaded', timeout: 20000 });
  await vpPage.waitForTimeout(500);
  if (!isMobile) await checkLayoutPresent(vpPage, `/governance [${vp.name}]`);
  await checkNoHOverflow(vpPage, `/governance [${vp.name}]`);
  await checkNoFatalError(vpPage, `/governance [${vp.name}]`);

  await vpContext.close();
}

await browser.close();

// ── Summary ───────────────────────────────────────────────────────────────────
console.log('\n' + '─'.repeat(60));
console.log(`Results: ${passCount} PASS, ${failCount} FAIL`);
if (errors.length > 0) {
  console.error('\nFailed checks:');
  errors.forEach(e => console.error(`  • ${e}`));
  process.exit(1);
}
console.log('All checks PASSED ✅');
process.exit(0);
PLAYWRIGHT_EOF

  # Jalankan Playwright script (node is installed; playwright resolvable from frontend dir)
  BASE_URL="$BASE_URL" SMOKE_EMAIL="$EMAIL" SMOKE_PASS="$PASS" \
    node "$TMP_SCRIPT" 2>&1 || {
      fail "Playwright script exited with error"
    }
fi

# ── Final summary ─────────────────────────────────────────────────────────────
echo ""
echo "════════════════════════════════════════════════════════════"
echo "  SMOKE TEST SUMMARY"
echo "  PASS: $PASS_COUNT  |  FAIL: $FAIL_COUNT"
echo "════════════════════════════════════════════════════════════"

if [[ ${#ERRORS[@]} -gt 0 ]]; then
  echo ""
  echo "Failed:"
  for e in "${ERRORS[@]}"; do echo "  • $e"; done
  echo ""
  exit 1
fi

echo ""
echo "  ✅ Semua smoke checks PASSED"
echo ""
exit 0
