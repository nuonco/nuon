VERSION --use-cache-command --use-copy-link 0.6

IMPORT github.com/powertoolsdev/shared-configs:main

FROM ghcr.io/powertoolsdev/ci-go-builder

WORKDIR /app

ARG GITHUB_ACTIONS=
# repo corresponds to the buf repository being built, linted, formatted etc
ARG --required REPO=

ARG EARTHLY_GIT_PROJECT_NAME
ARG GHCR_IMAGE=ghcr.io/$EARTHLY_GIT_PROJECT_NAME

ARG BUF_USER=jonmorehouse
ARG BUF_API_TOKEN=4c51e8481ed34404b7ab6a0c62dc7b2db82757d8f86e4caa853750973e2c5083

proto:
    FROM bufbuild/buf
    WORKDIR /work
    COPY --dir $REPO/ .
    WORKDIR $REPO
    RUN echo $BUF_API_TOKEN | buf registry login --username $BUF_USER --token-stdin

push-proto:
    FROM +proto
    RUN buf push

# TODO: reenable linting once we've figured out how to get golangci-lint working
lint:
    BUILD +lint-standard
    BUILD +lint-proto

lint-standard:
    FROM ghcr.io/powertoolsdev/ci-reviewdog
    WORKDIR /work
    COPY --dir . .
    DO shared-configs+LINT \
        --GITHUB_ACTIONS=$GITHUB_ACTIONS \
        --GOCACHE=$GOCACHE \
        --GOMODCACHE=$GOMODCACHE
    SAVE IMAGE --push $GHCR_IMAGE:lint

lint-proto:
    FROM +proto
    RUN sh -c "buf lint && buf format -d --exit-code" \
            || sh -c "printf '%s\n' 'Buf Format changes exist in current branch' >&2 && exit 1"

################################### UDCs ######################################

################################### LOCAL #####################################
