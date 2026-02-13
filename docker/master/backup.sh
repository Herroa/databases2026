#!/bin/bash

BACKUP_DIR="/backups"
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")

echo "[INFO] Starting backup at ${TIMESTAMP}"

# Проверяем, существует ли папка
mkdir -p "$BACKUP_DIR"

pg_dumpall -U postgres > "${BACKUP_DIR}/dump_${TIMESTAMP}.sql"

cp /var/lib/postgresql/data/postgresql.conf "${BACKUP_DIR}/postgresql_${TIMESTAMP}.conf"
cp /var/lib/postgresql/data/pg_hba.conf "${BACKUP_DIR}/pg_hba_${TIMESTAMP}.conf"
