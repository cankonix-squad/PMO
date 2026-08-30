#!/usr/bin/env bash
# =============================================================================
# CANKORA PMO — Database Backup Script
# Target: /root/cankonix-node/backups/cankonix-pmo/
# Usage : ./scripts/backup.sh
# =============================================================================

set -euo pipefail

APP_DIR="/root/cankonix-node/apps/cankonix-pmo"
BACKUP_DIR="/root/cankonix-node/backups/cankonix-pmo"
TIMESTAMP=$(date '+%Y%m%d-%H%M%S')
BACKUP_FILE="pmo-db-${TIMESTAMP}.sql"
COMPOSE_FILE="compose.yml"

# Load env to get DB credentials
if [[ -f "$APP_DIR/.env" ]]; then
  # shellcheck source=/dev/null
  set -a; source "$APP_DIR/.env"; set +a
fi

DB_USER="${DB_USER:-cankora}"
DB_NAME="${DB_NAME:-cankora_db}"

mkdir -p "$BACKUP_DIR"

echo "============================================================"
echo " CANKORA PMO — Database Backup $(date '+%Y-%m-%d %H:%M:%S %Z')"
echo "============================================================"
echo ""
echo "Target: $BACKUP_DIR/$BACKUP_FILE"
echo ""

cd "$APP_DIR"

docker compose -f "$COMPOSE_FILE" exec -T cankonix-pmo-db \
  pg_dump -U "$DB_USER" "$DB_NAME" > "$BACKUP_DIR/$BACKUP_FILE"

BACKUP_SIZE=$(du -sh "$BACKUP_DIR/$BACKUP_FILE" | cut -f1)

echo "✅  Backup complete: $BACKUP_DIR/$BACKUP_FILE ($BACKUP_SIZE)"
echo ""

# ── Retention: keep last 7 daily backups ─────────────────────────────────────
echo "Cleaning up backups older than 7 days..."
find "$BACKUP_DIR" -name "pmo-db-*.sql" -mtime +7 -delete
echo "Remaining backups:"
ls -lh "$BACKUP_DIR"/pmo-db-*.sql 2>/dev/null || echo "  (none)"
echo "============================================================"
