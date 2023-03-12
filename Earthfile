VERSION --use-cache-command 0.6

IMPORT github.com/powertoolsdev/shared-configs:main

FROM ghcr.io/powertoolsdev/ci-go-builder

WORKDIR /src

ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG CGO_ENABLED=0
ARG GOPRIVATE=github.com/powertoolsdev/*
ARG ETCSSL=/etc/ssl/cert.pem
ARG BUF_USER=jonmorehouse
ARG BUF_API_TOKEN=4c51e8481ed34404b7ab6a0c62dc7b2db82757d8f86e4caa853750973e2c5083
ARG GITHUB_ACTIONS=
ARG EARTHLY_GIT_PROJECT_NAME
ARG GHCR_IMAGE=ghcr.io/${EARTHLY_GIT_PROJECT_NAME}

CACHE $GOCACHE
CACHE $GOMODCACHE

code:
  DO +DEPS
  DO +GEN
  SAVE ARTIFACT .

deps:
    DO shared-configs+SETUP_SSH --GITHUB_ACTIONS=$GITHUB_ACTIONS

    # TODO(jm): we shouldn't be installing these dependencies here in this way. Probably should be in tools.go or even
    # better, in our base image.
    RUN go install github.com/bufbuild/buf/cmd/buf@v1.12.0 \
        && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
        && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
        && go install github.com/srikrsna/protoc-gen-gotag@v0.6.2

    COPY go.mod go.sum ./

    # NOTE(jm): once we have finished migrating go-waypoint we can remove the ssh step
    # RUN go mod download
    RUN --ssh git config --global --add safe.directory "$(pwd)" \
        && go mod download

    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

lint:
    FROM ghcr.io/powertoolsdev/ci-reviewdog
    DO +DEPS
    DO +GEN
    DO github.com/powertoolsdev/shared-configs+LINT \
        --GITHUB_ACTIONS=$GITHUB_ACTIONS \
        --GOCACHE=$GOCACHE \
        --GOMODCACHE=$GOMODCACHE

    SAVE IMAGE --push $GHCR_IMAGE:lint

################################### UDCs ######################################

################################### LOCAL #####################################
