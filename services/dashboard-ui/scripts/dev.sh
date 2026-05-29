#!/usr/bin/env bash
set -e

cleanup() {
    kill $(jobs -p) 2>/dev/null
    wait 2>/dev/null
}
trap cleanup EXIT INT TERM

# Kill any leftover processes on our ports from previous runs
if [ -f dist/.port ]; then
    OLD_PORT=$(cat dist/.port)
    lsof -ti :"$OLD_PORT" 2>/dev/null | xargs kill 2>/dev/null || true
    DEV_PORT=$((OLD_PORT + 1))
    lsof -ti :"$DEV_PORT" 2>/dev/null | xargs kill 2>/dev/null || true
fi

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
