#!/usr/bin/env bash
set -e

DEV_PGID=$(ps -o pgid= -p $$ | tr -d ' ')
PGID_FILE="/tmp/nuon-dashboard-dev.pgid"

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

# why: Build from this script's own checkout, not $NUON_ROOT/nuon: running dev from a
# worktree otherwise compiled the server from the wrong source. NUON_DIR overrides.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"

echo "Building dashboard server from ${NUON_DIR:-$REPO_DIR}..."
go build -C "${NUON_DIR:-$REPO_DIR}" -o /tmp/dashboard-server ./services/dashboard-ui/server

rm -f dist/.port
/tmp/dashboard-server serve &

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
