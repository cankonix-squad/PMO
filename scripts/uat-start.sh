#!/usr/bin/env bash
# =============================================================================
# PMO UAT Start Script
# Menjalankan backend + frontend untuk sesi UAT
# Usage: bash scripts/uat-start.sh
# =============================================================================
set -euo pipefail

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$BASE_DIR/backend"
FRONTEND_DIR="$BASE_DIR/frontend"
LOG_DIR="$BASE_DIR/.uat-logs"
PID_FILE="$BASE_DIR/.uat-pids"
BACKEND_PORT="${BACKEND_PORT:-8083}"
FRONTEND_PORT="${FRONTEND_PORT:-3003}"

mkdir -p "$LOG_DIR"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${GREEN}[UAT-START]${NC} $1"; }
warn()  { echo -e "${YELLOW}[UAT-START]${NC} $1"; }
error() { echo -e "${RED}[UAT-START]${NC} $1"; exit 1; }

# --- Pre-flight check ---
info "Running environment check first..."
bash "$BASE_DIR/scripts/uat-env-check.sh" || error "Environment check failed. Fix issues above before starting."

# --- Check if already running ---
if nc -z localhost "$BACKEND_PORT" 2>/dev/null; then
  warn "Port $BACKEND_PORT already in use. Backend may already be running."
  warn "Run 'bash scripts/uat-stop.sh' first if you want a clean start."
  read -r -p "Continue anyway? [y/N] " yn
  [[ "${yn,,}" == "y" ]] || { info "Aborted."; exit 0; }
fi

# --- Backend migration + seed ---
info "Applying migrations..."
cd "$BACKEND_DIR"
make migrate-up 2>&1 | tail -5
info "Running seed (idempotent)..."
SEED_SKIP_DEMO_USERS=true make seed 2>&1 | tail -3

# --- Start backend ---
info "Starting backend on :$BACKEND_PORT..."
nohup env \
  PORT="$BACKEND_PORT" \
  APP_URL="http://localhost:$FRONTEND_PORT" \
  CORS_ALLOWED_ORIGINS="http://localhost:$FRONTEND_PORT" \
  make dev > "$LOG_DIR/backend.log" 2>&1 &
BACKEND_PID=$!
echo "backend:$BACKEND_PID" > "$PID_FILE"
info "Backend PID: $BACKEND_PID → log: $LOG_DIR/backend.log"

# Wait for backend to be ready
info "Waiting for backend health..."
MAX_WAIT=30; WAITED=0
while ! curl -sf "http://localhost:$BACKEND_PORT/health" > /dev/null 2>&1; do
  sleep 1; ((WAITED++))
  if [ $WAITED -ge $MAX_WAIT ]; then
    error "Backend did not start within ${MAX_WAIT}s. Check $LOG_DIR/backend.log"
  fi
done
info "Backend ready ✓ (took ${WAITED}s)"

# --- Start frontend ---
info "Starting frontend on :$FRONTEND_PORT..."
cd "$FRONTEND_DIR"
# Clean .next to prevent CSS corruption
if [ -d ".next" ]; then
  warn "Removing stale .next directory to prevent CSS corruption..."
  rm -rf .next
fi
if [ ! -d "node_modules" ]; then
  info "Installing frontend dependencies..."
  npm install --silent
fi
nohup env FRONTEND_PORT="$FRONTEND_PORT" npm run dev > "$LOG_DIR/frontend.log" 2>&1 &
FRONTEND_PID=$!
echo "frontend:$FRONTEND_PID" >> "$PID_FILE"
info "Frontend PID: $FRONTEND_PID → log: $LOG_DIR/frontend.log"

# Wait for frontend to be ready
info "Waiting for frontend..."
MAX_WAIT=60; WAITED=0
while ! curl -sf -o /dev/null -w "%{http_code}" "http://localhost:$FRONTEND_PORT/" 2>/dev/null | grep -qE "200|307"; do
  sleep 2; WAITED=$((WAITED+2))
  if [ $WAITED -ge $MAX_WAIT ]; then
    error "Frontend did not start within ${MAX_WAIT}s. Check $LOG_DIR/frontend.log"
  fi
done
info "Frontend ready ✓ (took ${WAITED}s)"

echo
echo "═══════════════════════════════════════════════════"
echo " PMO UAT Stack READY"
echo "═══════════════════════════════════════════════════"
echo " Backend:  http://localhost:$BACKEND_PORT/health"
echo " Frontend: http://localhost:$FRONTEND_PORT"
echo " Login:    admin@cankora.local / Admin@Cankora2024!"
echo "───────────────────────────────────────────────────"
echo " Logs:     $LOG_DIR/"
echo " PIDs:     $PID_FILE"
echo " Stop:     bash scripts/uat-stop.sh"
echo " Status:   bash scripts/uat-status.sh"
echo "═══════════════════════════════════════════════════"
