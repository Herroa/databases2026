#!/bin/bash
set -e

until pg_isready -h pg-master -U postgres; do
  sleep 1
done

if [ ! -f "$PGDATA/PG_VERSION" ]; then
  rm -rf "$PGDATA"/*
  PGPASSWORD=replpass pg_basebackup \
    -h pg-master \
    -U repluser \
    -D "$PGDATA" \
    -Fp -Xs -P
  touch "$PGDATA/standby.signal"
fi

exec docker-entrypoint.sh postgres