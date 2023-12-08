VERSION --use-cache-command 0.6

FROM ./images/go-base+build-base

WORKDIR /src

ARG ETCSSL=/etc/ssl/cert.pem
ARG EARTHLY_GIT_PROJECT_NAME
ARG GHCR_IMAGE=ghcr.io/${EARTHLY_GIT_PROJECT_NAME}

deps:
    # TODO(jm): we shouldn't be installing these dependencies here in this way.
    # Probably should be in tools.go or even
    RUN go install github.com/bufbuild/buf/cmd/buf@v1.12.0 \
        && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
        && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
        && go install github.com/srikrsna/protoc-gen-gotag@v0.6.2

    COPY go.mod go.sum ./
    COPY .golangci-lint.yaml ./
    RUN go mod download

    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum
    SAVE IMAGE --cache-hint

clean:
    LOCALLY
    RUN git clean -f -x
