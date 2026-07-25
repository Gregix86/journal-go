#!/usr/bin/env bash
# Restauration a partir d'une sauvegarde (utile sur une nouvelle machine en cas de pepin).
# Usage : ./backups/restore.sh db-20260101-030000.sql uploads-20260101-030000.tar.gz

set -euo pipefail
cd "$(dirname "$0")/.."

DB_DUMP="$1"
UPLOADS_ARCHIVE="$2"

echo "Demarrage des services (base uniquement pour l'instant)..."
docker compose up -d db
sleep 5

echo "Restauration de la base de donnees depuis backups/$DB_DUMP ..."
cat "backups/$DB_DUMP" | docker compose exec -T db psql -U journal journal

echo "Restauration des fichiers medias depuis backups/$UPLOADS_ARCHIVE ..."
docker run --rm \
  -v journal-go_uploads:/data \
  -v "$(pwd)/backups:/backup" \
  alpine sh -c "rm -rf /data/* && tar xzf /backup/$UPLOADS_ARCHIVE -C /data"

echo "Demarrage du site..."
docker compose up -d

echo "Restauration terminee : http://localhost:8000"
