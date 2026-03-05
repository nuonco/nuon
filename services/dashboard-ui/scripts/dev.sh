#!/usr/bin/env bash
set -e

if [ "$NUON_DASHBOARD_SPA" = "true" ]; then
    echo "Building dashboard server..."
    go build -C "$NUON_ROOT/nuon" -o /tmp/dashboard-server ./services/dashboard-ui/server
    /tmp/dashboard-server serve &
    npm run dev:spa
else
    npm run dev
fi
