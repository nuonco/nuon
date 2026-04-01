#!/usr/bin/env bash
set -e

if [ "$NUON_DASHBOARD_SPA" = "true" ]; then
    echo "Building dashboard server..."
    go build -C "$NUON_ROOT/nuon" -o /tmp/dashboard-server ./services/dashboard-ui/server

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

    npm run dev:spa
else
    npm run dev
fi
