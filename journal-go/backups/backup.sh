#!/usr/bin/env bash
# Sauvegarde complete : base de donnees + fichiers medias.
# Usage : ./backups/backup.sh
# A lancer depuis la racine du projet (la ou se trouve docker-compose.yml).
# Peut etre planifie avec cron, ex: tous les jours a 3h :
#   0 3 * * * cd /chemin/vers/journal-go && ./backups/backup.sh >> backups/backup.log 2>&1

set -euo pipefail
cd "$(dirname "$0")/.."

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
OUT_DIR="backups"
mkdir -p "$OUT_DIR"

echo "[$TIMESTAMP] Dump de la base de donnees..."
docker compose exec -T db pg_dump -U journal journal > "$OUT_DIR/db-$TIMESTAMP.sql"

echo "[$TIMESTAMP] Archivage des fichiers medias (volume uploads)..."
docker run --rm \
  -v journal-go_uploads:/data \
  -v "$(pwd)/$OUT_DIR:/backup" \
  alpine tar czf "/backup/uploads-$TIMESTAMP.tar.gz" -C /data .

echo "[$TIMESTAMP] Termine. Fichiers crees :"
echo "  - $OUT_DIR/db-$TIMESTAMP.sql"
echo "  - $OUT_DIR/uploads-$TIMESTAMP.tar.gz"
echo ""
echo "Pense a copier ces fichiers ailleurs (NAS, cloud, cle USB...) pour une vraie replication."
