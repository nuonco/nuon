#!/bin/bash
# usage bash /usr/bin/backup.sh otel_log_records

set -e
set -o pipefail
set -u

TABLE="otel_log_records"
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

# check if the initial backup exists
QUERY="SELECT count(*) FROM system.backups WHERE status = 'BACKUP_CREATED' AND position(name, '$INITIAL_BACKUP') != 0;"
HAS_INITIAL=`clickhouse client -h $CLICKHOUSE_URL -q "$QUERY"`

# if it does not exit, create it
if [ "$HAS_INITIAL" == "0" ]; then
  CREATE_INITIAL_BACKUP_CMD="BACKUP TABLE $TABLE TO S3('$INITIAL_BACKUP');"
  echo
  echo "[clickhouse backups to s3] Creating initial backup: "$CREATE_INITIAL_BACKUP_CMD
  echo
  RESPONSE=`clickhouse client -h $CLICKHOUSE_URL -q "$CREATE_INITIAL_BACKUP_CMD"`
  if [[ $RESPONSE == *"BACKUP_FAILED"* ]]; then
      echo
      echo "[clickhouse backups to s3] failed create the initial backup"
      echo
      echo $RESPONSE >&2
      echo
      exit 126
  fi
elif [ "$HAS_INITIAL" == "1" ]; then
  echo
  echo "[clickhouse backups to s3] Found initial backup: "$INITIAL_CMD
  echo
else
  echo
  echo "[clickhouse backups to s3] Query returned unexpected value: "$INITIAL_CMD
  echo
  echo "  unsure what to do. exiting. "
  echo
  exit 1
fi

# finally: create the backup
COMMAND="BACKUP TABLE $TABLE TO S3('$CURRENT_BACKUP') SETTINGS base_backup = S3('$INITIAL_BACKUP');"

echo
echo '[clickhouse backups to s3] creating current backup: '$COMMAND
echo

RESPONSE=`clickhouse client -h $CLICKHOUSE_URL -q "$COMMAND"`
if [[ $RESPONSE == *"BACKUP_FAILED"* ]]; then
    echo
    echo "[clickhouse backups to s3] failed to create the current backup"
    echo
    echo $RESPONSE
    exit 1
fi
