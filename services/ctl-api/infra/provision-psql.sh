#!/bin/bash

set -eou pipefail
set -x

OUTPUT=$(terraform output -json)

# NOTE: we use the instance host for both this script, and connecting to the DB since aws iam auth doesn't seem to work
# with any user created host records.
DB_ADDR="$(echo $OUTPUT | jq -r .db.value.instance.host)"
DB_PORT="$(echo $OUTPUT | jq -r .db.value.instance.port)"

DB_NAME="$(echo $OUTPUT | jq -r .db.value.instance.name)"
DB_USER="$(echo $OUTPUT | jq -r .db.value.instance.username)"

ADMIN_USER="$(echo $OUTPUT | jq -r .db.value.admin.username )"
ADMIN_PW="$(echo $OUTPUT | jq -r .db.value.admin.password )"
ADMIN_DB="$(echo $OUTPUT | jq -r .db.value.admin.name)"

echo "executing setup iam..."
cat <<EOF | PGPASSWORD="$ADMIN_PW" psql \
    -h "$DB_ADDR" \
    -p "$DB_PORT" \
    -U "$ADMIN_USER" \
    -d "$ADMIN_DB" \
    --no-psqlrc \
    -f -

DO \$\$
BEGIN
CREATE EXTENSION IF NOT EXISTS hstore;
CREATE USER ctl_api WITH LOGIN;
EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
END
\$\$;

GRANT rds_iam TO ctl_api;
GRANT CREATE TO ctl_api;
CREATE DATABASE ctl_api;
EOF

echo "executing enable hstore..."
cat <<EOF | PGPASSWORD="$ADMIN_PW" psql \
    -h "$DB_ADDR" \
    -p "$DB_PORT" \
    -U "$ADMIN_USER" \
    -d "ctl_api" \
    --no-psqlrc \
    -f -

DO \$\$
BEGIN
CREATE EXTENSION IF NOT EXISTS hstore;
END
\$\$;
EOF
