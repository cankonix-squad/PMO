#!/usr/bin/env bash
# =============================================================================
# CANKORA UAT Stop Script
# Menghentikan backend + frontend yang dijalankan oleh uat-start.sh
# Usage: bash scripts/uat-stop.sh
# =============================================================================

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_FILE="$BASE_DIR/.uat-pids"

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()  { echo -e "${GREEN}[UAT-STOP]${NC} $1"; }
warn()  { echo -e "${YELLOW}[UAT-STOP]${NC} $1"; }
error() { echo -e "${RED}[UAT-STOP]${NC} $1"; }

echo
echo "── Stopping CANKORA UAT Stack ──"

STOPPED=0
FAILED=0

if [ -f "$PID_FILE" ]; then
  while IFS= read -r line; do
    LABEL=$(echo "$line" | cut -d: -f1)
    PID=$(echo "$line" | cut -d: -f2)
    if [ -z "$PID" ]; then continue; fi
    if kill -0 "$PID" 2>/dev/null; then
      kill "$PID" 2>/dev/null
      sleep 1
      if kill -0 "$PID" 2>/dev/null; then
        kill -9 "$PID" 2>/dev/null
        warn "$LABEL PID $PID required SIGKILL"
      else
        info "Stopped $LABEL (PID $PID)"
      fi
      ((STOPPED++))
    else
      warn "$LABEL PID $PID was not running"
    fi
  done < "$PID_FILE"
  rm -f "$PID_FILE"
  info "Removed PID file"
else
  warn "No PID file found at $PID_FILE"
  warn "Looking for processes by port..."

  # Fallback: kill by port — only kill processes listening on our ports
  for PORT in 8080 3000; do
    PIDS=$(lsof -ti tcp:"$PORT" 2>/dev/null || true)
    if [ -n "$PIDS" ]; then
      for P in $PIDS; do
        CMD=$(ps -p "$P" -o comm= 2>/dev/null || echo "?")
        # Only kill known CANKORA processes
        if echo "$CMD" | grep -qE "go|node|next|api|cankora"; then
          kill "$P" 2>/dev/null && info "Killed $CMD PID $P (port $PORT)" && ((STOPPED++)) || warn "Failed to kill PID $P"
        else
          warn "Port $PORT in use by $CMD (PID $P) — skipping (not a CANKORA process)"
        fi
      done
    fi
  done
fi

echo
echo "═══════════════════════════════════════════════════"
echo " UAT Stack Stop: $STOPPED process(es) stopped"
echo "═══════════════════════════════════════════════════"
echo " Status: bash scripts/uat-status.sh"
echo " Start:  bash scripts/uat-start.sh"
echo "═══════════════════════════════════════════════════"
