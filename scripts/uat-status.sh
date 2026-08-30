#!/usr/bin/env bash
# =============================================================================
# CANKORA UAT Status Script
# Cek status backend + frontend
# Usage: bash scripts/uat-status.sh
# =============================================================================

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_FILE="$BASE_DIR/.uat-pids"

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
ok()   { echo -e "${GREEN}✓ UP${NC}    $1"; }
down() { echo -e "${RED}✗ DOWN${NC}  $1"; }
warn() { echo -e "${YELLOW}~ WARN${NC}  $1"; }

echo
echo "═══════════════════════════════════════════════════"
echo " CANKORA UAT Stack Status"
echo "═══════════════════════════════════════════════════"

# Backend health
BACKEND_STATUS=$(curl -sf http://localhost:8080/health 2>/dev/null)
if [ $? -eq 0 ]; then
  VERSION=$(echo "$BACKEND_STATUS" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('version','?'))" 2>/dev/null || echo "?")
  ok "Backend  http://localhost:8080/health  (v$VERSION)"
else
  down "Backend  http://localhost:8080  — not responding"
fi

# Frontend
FE_CODE=$(curl -sf -o /dev/null -w "%{http_code}" http://localhost:3000/ 2>/dev/null)
if [[ "$FE_CODE" == "200" || "$FE_CODE" == "307" ]]; then
  ok "Frontend http://localhost:3000  (HTTP $FE_CODE)"
else
  down "Frontend http://localhost:3000  — not responding (got: ${FE_CODE:-no response})"
fi

# PID file
if [ -f "$PID_FILE" ]; then
  echo
  echo "── Recorded PIDs ($PID_FILE) ──"
  while IFS= read -r line; do
    LABEL=$(echo "$line" | cut -d: -f1)
    PID=$(echo "$line" | cut -d: -f2)
    if kill -0 "$PID" 2>/dev/null; then
      echo -e "  ${GREEN}●${NC} $LABEL  PID $PID  (running)"
    else
      echo -e "  ${RED}●${NC} $LABEL  PID $PID  (not running)"
    fi
  done < "$PID_FILE"
fi

# PostgreSQL
echo
if nc -z localhost 5432 2>/dev/null; then
  ok "PostgreSQL :5432  reachable"
else
  down "PostgreSQL :5432  not reachable"
fi

echo
echo "── Quick login check ──"
TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cankora.local","password":"Admin@Cankora2024!"}' \
  2>/dev/null | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('access_token',''))" 2>/dev/null)
if [ -n "$TOKEN" ]; then
  ok "Login: admin@cankora.local ✓"
else
  down "Login: admin@cankora.local ✗ — backend may not be ready"
fi

echo
echo "═══════════════════════════════════════════════════"
echo " Start:  bash scripts/uat-start.sh"
echo " Stop:   bash scripts/uat-stop.sh"
echo " Env:    bash scripts/uat-env-check.sh"
echo "═══════════════════════════════════════════════════"
