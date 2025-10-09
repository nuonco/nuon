#!/bin/bash
# Docker command wrapper for nerdctl with buildx support
# Routes buildx commands to real buildx plugin, everything else to nerdctl

if [ "$1" = "buildx" ]; then
    # Use real buildx plugin
    shift

    # Ensure buildx is configured to use remote BuildKit
    if [ ! -z "$BUILDKIT_HOST" ]; then
        # Check if our builder exists
        if ! /usr/local/lib/docker/cli-plugins/docker-buildx ls 2>/dev/null | grep -q "^remote-buildkit"; then
            # Create builder pointing to remote BuildKit
            /usr/local/lib/docker/cli-plugins/docker-buildx create \
                --name remote-buildkit \
                --driver remote \
                --use \
                "$BUILDKIT_HOST" >/dev/null 2>&1 || true
        fi

        # Use the remote builder
        export BUILDX_BUILDER=remote-buildkit
    fi

    exec /usr/local/lib/docker/cli-plugins/docker-buildx "$@"
else
    # Use nerdctl for all other commands
    exec /usr/local/bin/nerdctl "$@"
fi