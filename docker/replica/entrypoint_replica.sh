#!/usr/bin/env bash
set -e

# Wait for master to be ready
until pg_isready -h pg-master -p 5432 -U postgres; do
  echo "Waiting for pg-master..."
  sleep 1
done

# If data folder empty — basebackup
if [ ! -s "/var/lib/postgresql/data/PG_VERSION" ]; then
  echo "Bootstrapping replica basebackup..."
  PGPASSWORD=replpass pg_basebackup -h pg-master -D /var/lib/postgresql/data -U repluser -v -P --wal-method=stream
  # Create standby.signal for Postgres >=12
  touch /var/lib/postgresql/data/standby.signal
  cat >> /var/lib/postgresql/data/postgresql.auto.conf <<EOF
primary_conninfo = 'host=pg-master port=5432 user=repluser password=replpass'
recovery_target_timeline = 'latest'
EOF
  chown -R postgres:postgres /var/lib/postgresql/data
fi

# exec postgres
exec docker-entrypoint.sh postgres