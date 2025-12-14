#!/bin/bash
set -e

docker-entrypoint.sh postgres &

until pg_isready -U postgres; do
  sleep 2
done

psql -U postgres <<EOF
ALTER SYSTEM SET wal_level = replica;
ALTER SYSTEM SET max_wal_senders = 5;
ALTER SYSTEM SET wal_keep_size = '64MB';
ALTER SYSTEM SET listen_addresses = '*';
SELECT pg_reload_conf();
EOF

psql -U postgres -f /docker/master/init-replication.sql

wait