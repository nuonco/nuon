VERSION --use-cache-command 0.6

FROM ghcr.io/powertoolsdev/ci-go-builder


WORKDIR /pkg

ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG CGO_ENABLED=0
ARG GOPRIVATE=github.com/powertoolsdev/*

ARG GITHUB_ACTIONS=

ARG GHCR_IMAGE=ghcr.io/powertoolsdev/go-common

deps:
    ARG from=+base
    FROM $from
    WORKDIR /app
    CACHE $GOCACHE
    CACHE $GOMODCACHE
    COPY go.mod go.sum .
    COPY --dir config shortid temporalzap .
    IF [ -z "$GITHUB_ACTIONS" ]     # local
        RUN git config --global \
                url."git@github.com:powertoolsdev/".insteadOf \
                "https://github.com/powertoolsdev/" \
            && ssh-keyscan github.com 2> /dev/null >> ~/.ssh/known_hosts
    ELSE
        RUN --secret clone_token \
            git config --global \
                url."https://x-access-token:$clone_token@github.com/".insteadOf \
                https://github.com/
    END
    RUN --ssh git config --global --add safe.directory /work \
        && go mod download

test:
    FROM +deps
    RUN go test ./...
    SAVE IMAGE --push $GHCR_IMAGE:test

test-integration:
    FROM +deps
    RUN go test -tags=integration ./...
    SAVE IMAGE --push $GHCR_IMAGE:test-integration

lint:
    FROM --platform=linux/amd64 +deps --from=ghcr.io/powertoolsdev/ci-reviewdog
    WORKDIR /work
    COPY --dir . .
    CACHE /root/.cache
    DO github.com/powertoolsdev/shared-configs+LINT \
        --GITHUB_ACTIONS=$GITHUB_ACTIONS \
        --GOCACHE=$GOCACHE \
        --GOMODCACHE=$GOMODCACHE
    SAVE IMAGE --push $GHCR_IMAGE:lint
