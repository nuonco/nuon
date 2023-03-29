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
ARG EARTHLY_GIT_PROJECT_NAME
ARG GHCR_IMAGE=ghcr.io/${EARTHLY_GIT_PROJECT_NAME}

CACHE $GOCACHE
CACHE $GOMODCACHE

deps:
    # TODO(jm): we shouldn't be installing these dependencies here in this way. Probably should be in tools.go or even
    # better, in our base image.
    RUN go install github.com/bufbuild/buf/cmd/buf@v1.12.0 \
        && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
        && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
        && go install github.com/srikrsna/protoc-gen-gotag@v0.6.2

    COPY go.mod go.sum ./
    RUN go mod download

    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum
    SAVE IMAGE --cache-hint

clean:
    LOCALLY
    RUN git clean -f -x
