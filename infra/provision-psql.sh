#!/bin/bash

set -eou pipefail
set -x

# NOTE: we use the instance host for both this script, and connecting to the DB since aws iam auth doesn't seem to work
# with any user created host records.
INSTANCE_ADDR="$(terraform output -raw db_instance_host)"
INSTANCE_DB="$(terraform output -raw db_instance_name)"
INSTANCE_PORT="$(terraform output -raw db_instance_port)"
INSTANCE_USER="$(terraform output -raw db_instance_username)"
INSTANCE_PW="$(terraform output -raw db_instance_password)"

cat <<EOF | PGPASSWORD="$INSTANCE_PW" psql \
    -h "$INSTANCE_ADDR" \
    -p "$INSTANCE_PORT" \
    -U "$INSTANCE_USER" \
    -d "$INSTANCE_DB" \
    --no-psqlrc \
    -f -

DO \$\$
BEGIN
CREATE USER api WITH LOGIN;
EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
END
\$\$;

GRANT rds_iam TO api;
CREATE DATABASE api;
EOF
