#!/bin/bash
set -e

echo "⏳ Waiting for pg-master..."

until pg_isready -h pg-master -U postgres; do
  sleep 2
done

echo "pg-master is ready, starting backup loop"

while true; do
  TS=$(date +"%Y-%m-%d_%H-%M-%S")
  echo "[$(date)] running backup..."

  pg_dumpall -h pg-master -U postgres > /backups/globals_$TS.sql
  pg_dump -h pg-master -U postgres -Fc sports_club > /backups/sports_club_$TS.dump

  find /backups -type f -mtime +7 -delete
  sleep 43200
done