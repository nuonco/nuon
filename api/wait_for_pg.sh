#!/usr/bin/env sh

COUNTER=15
PG_ISREADY=$(find /usr -type f -name pg_isready -executable -print -quit)
printf "Found pg_isready: %s\n" "$PG_ISREADY"

while ! sh -c "$PG_ISREADY -h 127.0.0.1";
do
    COUNTER=$((COUNTER-1))
    if [ $COUNTER -lt 0 ];
    then
        exit 1;
    fi
    sleep 1;
    counter=$((counter-1));
done

printf "Connected to postgres\n"
exit 0
