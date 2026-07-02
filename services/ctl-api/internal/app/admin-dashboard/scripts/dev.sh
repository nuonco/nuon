#!/usr/bin/env bash
set -e

cd "$(dirname "$0")/.."

DEV_PGID=$(ps -o pgid= -p $$ | tr -d ' ')
PGID_FILE="/tmp/nuon-admin-dashboard-dev.pgid"

if [ -f "$PGID_FILE" ]; then
    OLD_PGID=$(cat "$PGID_FILE" 2>/dev/null || true)
    if [ -n "$OLD_PGID" ] && [ "$OLD_PGID" != "$DEV_PGID" ]; then
        kill -TERM -- "-$OLD_PGID" 2>/dev/null || true
    fi
fi
pkill -f 'esbuild client/index.tsx --bundle --outfile=dist/app.js' 2>/dev/null || true
pkill -f 'admin-dashboard/node_modules/.bin/postcss' 2>/dev/null || true
pkill -f 'admin-dashboard/node_modules/.bin/browser-sync' 2>/dev/null || true
pkill -f 'admin-dashboard/node_modules/.bin/run-p' 2>/dev/null || true

echo "$DEV_PGID" > "$PGID_FILE"

cleanup() {
    kill -TERM -- "-$DEV_PGID" 2>/dev/null || true
    wait 2>/dev/null || true
}
trap cleanup EXIT INT TERM

PARENT_PID=$PPID
(
    while kill -0 "$PARENT_PID" 2>/dev/null; do
        sleep 2
    done
    kill -TERM -- "-$DEV_PGID" 2>/dev/null || true
) &

mkdir -p dist
npm run dev:html
npm run dev:static
npm run dev:watch &
wait $!
