VERSION --shell-out-anywhere 0.6
FROM golang:1.18.3-alpine3.16


RUN apk add --update --no-cache \
    bash \
    binutils \
    ca-certificates \
    coreutils \
    curl \
    git \
    grep \
    less \
    openssl \
    openssh-client

WORKDIR pkg

ARG GOCACHE=/go-cache
ENV CGO_ENABLED=0

deps:
    FROM +base
    RUN git config --global url."git@github.com:".insteadOf "https://github.com/"
    RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
    COPY go.mod go.sum ./

artifact:
    FROM +deps
    COPY --dir . .
    RUN --ssh go mod download
    SAVE ARTIFACT .

lint:
    FROM golangci/golangci-lint
    COPY +artifact/* /pkg/go-common
    COPY  ../../+lintcfg/cfg /pkg/go-common/.golangci.yml
    RUN golangci-lint run

vet:
    FROM +artifact
    RUN go vet ./...

test:
    FROM +artifact
    RUN go test ./...
