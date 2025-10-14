#!/bin/bash
set -e

echo "🏁 Initializing master PostgreSQL..."

# Стандартная инициализация
docker-entrypoint.sh postgres &

# Ждём, пока сервер запустится
until pg_isready -U postgres; do
  sleep 2
done

# Создаём пользователя и слот репликации
psql -U postgres -c "CREATE ROLE replica WITH REPLICATION PASSWORD 'replica' LOGIN;"
psql -U postgres -c "SELECT * FROM pg_create_physical_replication_slot('replica_slot');"

echo "✅ Master ready for replication"
wait