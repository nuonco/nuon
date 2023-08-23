#!/bin/bash

set -eou pipefail
set -x

# NOTE: we use the instance host for both this script, and connecting to the DB since aws iam auth doesn't seem to work
# with any user created host records.
DB_ADDR="$(terraform output -json | jq .instance.host)"
DB_PORT="$(terraform output -json | jq .instance.port)"

DB_NAME="$(terraform output -json | jq .instance.name)"
DB_USER="$(terraform output -json | jq .instance.username)"

ADMIN_USER="$(terraform output -json db | jq -c .admin.username )"
ADMIN_PW="$(terraform output -json | jq -c .admin.password )"
ADMIN_DB="$(terraform output -json db | jq -c .admin.name)"

cat <<EOF | PGPASSWORD="$INSTANCE_PW" psql \
    -h "$DB_ADDR" \
    -p "$DB_PORT" \
    -U "$ADMIN_USER" \
    -d "$ADMIN_DB" \
    --no-psqlrc \
    -f -

DO \$\$
BEGIN
CREATE USER ctl-api WITH LOGIN;
EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
END
\$\$;

GRANT rds_iam TO ctl-api;
CREATE DATABASE ctl-api;
CREATE EXTENSION IF NOT EXISTS hstore;
EOF
