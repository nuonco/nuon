#!/bin/bash
# usage bash /usr/bin/backup.sh otel_log_records

set -e
set -o pipefail
set -u

TABLE=$1
TIMESTAMP=`date -Iseconds -u | sed 's/-//g; s/://g; s/T//g; s/+//g'`
LOCATION="$BUCKET_URL/backups/"$TABLE
INITIAL_BACKUP="$LOCATION/initial"
CURRENT_BACKUP="$LOCATION/$TIMESTAMP"

echo
echo "[clickhouse backups to s3] Preparing to create a backup:"
echo
echo "    table = "$TABLE
echo "   backup = "$TIMESTAMP
echo " location = "$LOCATION
echo "  initial = "$INITIAL_BACKUP
echo "  current = "$CURRENT_BACKUP
echo


set +e
# 1. check for an existing initial backup. if it does not exist, create it.

# NOTE(fd): we are allowing this to fail w/out breaking the script.
#  - if it fails because the initial backup exists, that's fine
#  - if it fails for other reasons, the backup command will fail
#    and we can rely on catching that error

# the query checks for entries w/ a base backup. these can only exist if the base backup exists. so if the count is > 0, we have
# an initial backup. it is a bit indirect but it works well.
HAS_INITIAL_QUERY="SELECT count(*) FROM system.backups WHERE status = 'BACKUP_CREATED' AND position(base_backup_name, '$INITIAL_BACKUP') != 0;"
HAS_INITIAL=`clickhouse client --host $CLICKHOUSE_URL --user $CLICKHOUSE_USERNAME --password $CLICKHOUSE_PASSWORD -q "$HAS_INITIAL_QUERY"`

# if it does not exit, create it
if [ "$HAS_INITIAL" == "0" ]; then
  CREATE_INITIAL_BACKUP_CMD="BACKUP TABLE $TABLE TO S3('$INITIAL_BACKUP');"
  echo
  echo "[clickhouse backups to s3] Creating initial backup: "$CREATE_INITIAL_BACKUP_CMD
  echo
  RESPONSE=`clickhouse client --host $CLICKHOUSE_URL --user $CLICKHOUSE_USERNAME --password $CLICKHOUSE_PASSWORD -q "$CREATE_INITIAL_BACKUP_CMD"`
  if [[ $RESPONSE == *"BACKUP_FAILED"* ]]; then
      echo
      echo "[clickhouse backups to s3] failed create the initial backup"
      echo
      echo $RESPONSE >&2
      echo
      echo "[clickhouse backups to s3] we will continue anyway"
      echo
  fi
fi
set -e

# 2. create the backup w/ the initital backup as its base

COMMAND="BACKUP TABLE $TABLE TO S3('$CURRENT_BACKUP') SETTINGS base_backup = S3('$INITIAL_BACKUP');"

echo
echo '[clickhouse backups to s3] creating current backup: '$COMMAND
echo

RESPONSE=`clickhouse client --host $CLICKHOUSE_URL --user $CLICKHOUSE_USERNAME --password $CLICKHOUSE_PASSWORD -q "$COMMAND"`
if [[ $RESPONSE == *"BACKUP_FAILED"* ]]; then
    echo
    echo "[clickhouse backups to s3] failed to create the current backup"
    echo
    echo $RESPONSE
    exit 1
fi
