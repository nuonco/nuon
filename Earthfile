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

WORKDIR /pkg

ARG GOCACHE=/go-cache
ENV CGO_ENABLED=0

deps:
    FROM +base
    RUN git config --global url."git@github.com:".insteadOf "https://github.com/"
    RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
    COPY go.mod go.sum ./

artifact:
    FROM +deps
    COPY --dir config .
    RUN --ssh go mod download
    SAVE ARTIFACT .
    SAVE IMAGE --cache-hint

lint:
    FROM golangci/golangci-lint
    COPY +artifact/* .
    COPY .golangci.yml .
    RUN golangci-lint run

# remove when golangci-lint works again
vet:
    FROM +artifact
    RUN go vet ./...

test:
    FROM +artifact
    RUN go test ./...

ci:
    BUILD +vet
    BUILD +test


# for other go repos to consume
lintcfg:
    COPY .golangci.yml .
    SAVE ARTIFACT .golangci.yml /cfg
