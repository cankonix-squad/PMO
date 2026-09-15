#!/usr/bin/env bash
# =============================================================================
# PMO — Local Reset (Super Admin only)
#
# Mengosongkan data demo/transaksional dari database LOKAL (dari backend/.env)
# dan menyisakan hanya:
#   - schema migration (tabel schema_migrations tidak disentuh)
#   - organisasi default, RBAC/permissions/roles, master data dasar
#   - user Super Admin: admin@cankora.local / Admin@Cankora2024!
#
# TIDAK menjalankan seed-demo. Demo users (pmo@, pm@, officer@, executive@,
# auditor@) dan project DEMO-SDA-* ikut terhapus dan tidak dibuat ulang.
#
# Usage:
#   bash scripts/reset-local-admin-only.sh              # stop app + reset + verify
#   bash scripts/reset-local-admin-only.sh --yes        # tanpa prompt konfirmasi
#   bash scripts/reset-local-admin-only.sh --keep-running   # jangan stop app
#   bash scripts/reset-local-admin-only.sh --no-migrate # lewati migrate-up
#
# Keamanan:
#   - Hanya untuk database LOKAL. Script menolak DB_NAME selain cankora_db
#     (override: RESET_LOCAL_EXPECTED_DB=...).
#   - cmd/reset-local menolak APP_ENV=production dan host cloud/managed.
# =============================================================================
set -euo pipefail

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$BASE_DIR/backend"
ENV_FILE="$BACKEND_DIR/.env"
BACKEND_PORT="${BACKEND_PORT:-8083}"
FRONTEND_PORT="${FRONTEND_PORT:-3003}"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${GREEN}[RESET]${NC} $1"; }
warn()  { echo -e "${YELLOW}[RESET]${NC} $1"; }
error() { echo -e "${RED}[RESET]${NC} $1"; exit 1; }
header(){ echo; echo "── $1 ──"; }

ASSUME_YES=0
KEEP_RUNNING=0
RUN_MIGRATE=1

for arg in "$@"; do
  case "$arg" in
    --yes|-y)        ASSUME_YES=1 ;;
    --keep-running)  KEEP_RUNNING=1 ;;
    --no-migrate)    RUN_MIGRATE=0 ;;
    -h|--help)
      sed -n '2,22p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
      exit 0 ;;
    *) error "Unknown argument: $arg" ;;
  esac
done

header "0. Prasyarat"
[ -f "$ENV_FILE" ] || error "backend/.env tidak ditemukan — jalankan: cp backend/.env.example backend/.env"
command -v go >/dev/null 2>&1 || error "go tidak ditemukan di PATH"

# --- Baca konfigurasi DB dari backend/.env (tanpa mengekspor apa pun) -------
env_value() {
  local key="$1"
  grep -E "^[[:space:]]*${key}=" "$ENV_FILE" 2>/dev/null | tail -1 | cut -d= -f2- \
    | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' -e 's/^"//' -e 's/"$//' -e "s/^'//" -e "s/'$//"
}

DB_HOST="$(env_value DB_HOST)";   DB_HOST="${DB_HOST:-localhost}"
DB_PORT="$(env_value DB_PORT)";   DB_PORT="${DB_PORT:-5432}"
DB_NAME="$(env_value DB_NAME)";   DB_NAME="${DB_NAME:-cankora_db}"
DB_USER="$(env_value DB_USER)"
APP_ENV="$(env_value APP_ENV)";   APP_ENV="${APP_ENV:-development}"
# .env values may carry an inline comment (e.g. "development  # dev | prod")
# — keep only the first whitespace-delimited token.
APP_ENV="${APP_ENV%%[[:space:]]*}"

export DB_HOST DB_PORT DB_NAME DB_USER APP_ENV

EXPECTED_DB="${RESET_LOCAL_EXPECTED_DB:-cankora_db}"
if [ "$DB_NAME" != "$EXPECTED_DB" ]; then
  error "DB_NAME '$DB_NAME' != RESET_LOCAL_EXPECTED_DB '$EXPECTED_DB' — menolak reset"
fi

case "$(echo "$APP_ENV" | tr '[:upper:]' '[:lower:]')" in
  production|prod)
    error "APP_ENV=$APP_ENV — script ini hanya untuk database lokal/UAT" ;;
esac

case "$DB_HOST" in
  localhost|127.0.0.1|::1|0.0.0.0|host.docker.internal|db|postgres|*.local) ;;
  *) error "DB_HOST '$DB_HOST' bukan host lokal — menolak reset" ;;
esac

info "Target: db=$DB_NAME user=$DB_USER host=$DB_HOST:$DB_PORT app_env=$APP_ENV"
info "Ports : backend=$BACKEND_PORT frontend=$FRONTEND_PORT"

# --- Konfirmasi --------------------------------------------------------------
if [ "$ASSUME_YES" -ne 1 ]; then
  echo
  warn "Semua data proyek/transaksional di '$DB_NAME' akan DIHAPUS permanen."
  warn "Hanya Super Admin (admin@cankora.local) yang disisakan."
  read -r -p "Lanjutkan? [y/N] " yn
  [[ "${yn:-}" =~ ^[Yy]$ ]] || { info "Dibatalkan."; exit 0; }
fi

# --- 1. Hentikan aplikasi ----------------------------------------------------
if [ "$KEEP_RUNNING" -eq 1 ]; then
  warn "Melewati stop aplikasi (--keep-running)"
else
  header "1. Menghentikan backend & frontend"
  if [ -f "$BASE_DIR/scripts/uat-stop.sh" ]; then
    bash "$BASE_DIR/scripts/uat-stop.sh" | sed 's/^/  /' || warn "uat-stop.sh selesai dengan error (diabaikan)"
  fi
  # Bersihkan proses tersisa yang masih memegang port UAT/frontend.
  for PORT in "$BACKEND_PORT" "$FRONTEND_PORT"; do
    PIDS="$(lsof -ti tcp:"$PORT" 2>/dev/null || true)"
    if [ -n "$PIDS" ]; then
      for P in $PIDS; do
        CMD="$(ps -p "$P" -o comm= 2>/dev/null || echo '?')"
        if echo "$CMD" | grep -qE "go|node|next|api|cankora|main"; then
          kill "$P" 2>/dev/null && info "Killed $CMD (PID $P, port $PORT)" || warn "Gagal kill PID $P"
        else
          warn "Port $PORT dipakai $CMD (PID $P) — dilewati (bukan proses PMO)"
        fi
      done
    fi
  done
  sleep 1
fi

cd "$BACKEND_DIR"

# --- 2. Migration up (schema tetap up-to-date) -------------------------------
if [ "$RUN_MIGRATE" -eq 1 ]; then
  header "2. Menjalankan migration (up)"
  make migrate-up 2>&1 | tail -5 | sed 's/^/  /'
else
  warn "Melewati migrate-up (--no-migrate)"
fi

# --- 3. Hapus data demo/transaksional ---------------------------------------
header "3. Menghapus data demo/transaksional"
make reset-local 2>&1 | sed 's/^/  /'

# --- 4. Seed dasar TANPA demo users -----------------------------------------
header "4. Seed dasar (Super Admin only, tanpa seed-demo)"
SEED_SKIP_DEMO_USERS=true go run ./cmd/seed 2>&1 | tail -5 | sed 's/^/  /'

# --- 5. Verifikasi -----------------------------------------------------------
header "5. Verifikasi"
go run ./cmd/reset-local --verify 2>&1 | sed 's/^/  /'

echo
echo "═══════════════════════════════════════════════════"
echo " RESET SELESAI — Super Admin only"
echo "═══════════════════════════════════════════════════"
echo " Login: admin@cankora.local / Admin@Cankora2024!"
echo "───────────────────────────────────────────────────"
echo " Start UAT:  bash scripts/uat-start.sh"
echo " Verify:     cd backend && make reset-local-verify"
echo "═══════════════════════════════════════════════════"
