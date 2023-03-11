VERSION --use-copy-link 0.6

IMPORT github.com/powertoolsdev/shared-configs:main

FROM ghcr.io/powertoolsdev/ci-go-builder


WORKDIR /pkg

ARG EARTHLY_GIT_PROJECT_NAME

ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG CGO_ENABLED=0
ARG GOPRIVATE=github.com/powertoolsdev/*

ARG GITHUB_ACTIONS=

ARG GHCR_IMAGE=ghcr.io/${EARTHLY_GIT_PROJECT_NAME}

test:
    DO +DEPS
    RUN go test ./...
    SAVE IMAGE --push $GHCR_IMAGE:test

test-integration:
    DO +DEPS
    RUN go test -tags=integration ./...
    SAVE IMAGE --push $GHCR_IMAGE:test-integration

lint:
    FROM ghcr.io/powertoolsdev/ci-reviewdog
    DO +DEPS
    WORKDIR /work
    COPY --dir . .
    DO github.com/powertoolsdev/shared-configs+LINT \
        --GITHUB_ACTIONS=$GITHUB_ACTIONS \
        --GOCACHE=$GOCACHE \
        --GOMODCACHE=$GOMODCACHE
    SAVE IMAGE --push $GHCR_IMAGE:lint

DEPS:
    COMMAND
    COPY go.mod go.sum *.go .
    COPY --dir pkg .
    DO shared-configs+SETUP_SSH --GITHUB_ACTIONS=$GITHUB_ACTIONS
    RUN --ssh git config --global --add safe.directory "$(pwd)" \
        && go mod download
