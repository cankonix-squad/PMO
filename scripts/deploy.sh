#!/usr/bin/env bash
# =============================================================================
# PMO — Production Deployment Script
# Target: /root/cankonix-node/apps/cankonix-pmo on 187.77.127.202
# Usage : ./scripts/deploy.sh [--skip-build]
# =============================================================================

set -euo pipefail

APP_DIR="/root/cankonix-node/apps/cankonix-pmo"
COMPOSE_FILE="compose.yml"
DEPLOY_REMOTE_URL="${DEPLOY_REMOTE_URL:-https://github.com/cankonix-squad/PMO.git}"
SKIP_BUILD="${1:-}"

cd "$APP_DIR"

echo "============================================================"
echo " PMO — Deployment $(date '+%Y-%m-%d %H:%M:%S %Z')"
echo "============================================================"

# ── 1. Show current commit ────────────────────────────────────────────────────
echo ""
echo "[1/5] Current commit before update:"
git rev-parse HEAD

# ── 2. Pull latest source ─────────────────────────────────────────────────────
echo ""
echo "[2/5] Fetching latest source from main..."
git remote set-url origin "$DEPLOY_REMOTE_URL"
git fetch origin
git checkout main
git reset --hard origin/main

echo "New commit:"
git rev-parse HEAD

# ── 3. Build containers ───────────────────────────────────────────────────────
if [[ "$SKIP_BUILD" != "--skip-build" ]]; then
  echo ""
  echo "[3/5] Building containers..."
  docker compose -f "$COMPOSE_FILE" build --pull
else
  echo ""
  echo "[3/5] Skipping build (--skip-build flag set)"
fi

# ── 4. Start / update containers ─────────────────────────────────────────────
echo ""
echo "[4/5] Starting containers..."
docker compose -f "$COMPOSE_FILE" up -d --remove-orphans

# ── 5. Health check ───────────────────────────────────────────────────────────
echo ""
echo "[5/5] Waiting for API health check..."

MAX_WAIT=60
INTERVAL=5
ELAPSED=0
API_HEALTHY=false

while [[ $ELAPSED -lt $MAX_WAIT ]]; do
  if docker compose -f "$COMPOSE_FILE" exec -T cankonix-pmo-api \
    wget -qO- http://localhost:8080/health >/dev/null 2>&1; then
    API_HEALTHY=true
    break
  fi
  echo "    API not ready yet, retrying in ${INTERVAL}s... (${ELAPSED}s elapsed)"
  sleep "$INTERVAL"
  ELAPSED=$((ELAPSED + INTERVAL))
done

echo ""
echo "============================================================"
if [[ "$API_HEALTHY" == "true" ]]; then
  echo " ✅  Deployment SUCCESSFUL"
  echo "     API health: OK"
else
  echo " ❌  Deployment WARNING: API health check timed out after ${MAX_WAIT}s"
  echo "     Check logs: docker compose -f $COMPOSE_FILE logs --tail=50 cankonix-pmo-api"
  exit 1
fi

echo ""
echo "Container status:"
docker compose -f "$COMPOSE_FILE" ps
echo "============================================================"
