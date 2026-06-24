#!/usr/bin/env bash
set -e

DEV_PGID=$(ps -o pgid= -p $$ | tr -d ' ')
PGID_FILE="/tmp/nuon-dashboard-dev.pgid"

# A previous session's bun build/css --watch hold no port, so a port-based
# cleanup misses them and they pile up across restarts. Every child stays in
# dev.sh's process group, so kill the previous session's whole group, then
# sweep any strays whose marker was lost (e.g. clean-dist wiped an old dist/).
if [ -f "$PGID_FILE" ]; then
    OLD_PGID=$(cat "$PGID_FILE" 2>/dev/null || true)
    if [ -n "$OLD_PGID" ] && [ "$OLD_PGID" != "$DEV_PGID" ]; then
        kill -TERM -- "-$OLD_PGID" 2>/dev/null || true
    fi
fi
pkill -f 'bun build client/index.tsx --outdir=dist/assets' 2>/dev/null || true
pkill -f 'bun scripts/build-css.js --watch' 2>/dev/null || true
pkill -f 'bun scripts/dev-server.js' 2>/dev/null || true

echo "$DEV_PGID" > "$PGID_FILE"

cleanup() {
    kill -TERM -- "-$DEV_PGID" 2>/dev/null || true
    wait 2>/dev/null || true
}
trap cleanup EXIT INT TERM

echo "Building dashboard server..."
go build -C "${NUON_DIR:-$NUON_ROOT/nuon}" -o /tmp/dashboard-server ./services/dashboard-ui/server

rm -f dist/.port
/tmp/dashboard-server serve &

# Wait for the Go server to write its port file
for i in $(seq 1 50); do
    [ -f dist/.port ] && break
    sleep 0.1
done

if [ -f dist/.port ]; then
    export HTTP_PORT=$(cat dist/.port)
    echo "Go BFF listening on port $HTTP_PORT"
else
    echo "Warning: port file not found, falling back to default"
fi

bun run dev
