#!/bin/bash
set -e  # Остановить скрипт при любой ошибке

# Создаём пользователя только для репликации
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE USER replicator WITH REPLICATION ENCRYPTED PASSWORD 'postgres';
EOSQL

# Включаем репликацию в настройках PostgreSQL
cat >> /var/lib/postgresql/data/postgresql.conf <<EOF

wal_level = replica          # Включает лог изменений для реплик
max_wal_senders = 3          # Максимум 3 реплики
wal_keep_size = 100MB        # Хранить WAL-файлы, пока реплика не заберёт
hot_standby = on             # Разрешить чтение на реплике
EOF

# Разрешаем реплике подключаться к Primary
cat >> /var/lib/postgresql/data/pg_hba.conf <<EOF
host replication replicator 0.0.0.0/0 md5  # Только для репликации, с паролем
EOF