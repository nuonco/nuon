FROM golang:1.24 as code

ARG BIN_NAME=nuon
ARG SERVICE=cli
ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG PKG_ROOT=github.com/powertoolsdev/mono/bins/${SERVICE}

WORKDIR /src

# copy deps, tools and fetch dependencies
COPY tools.go go.mod go.sum .
RUN go mod download

# copy package and generate it
COPY pkg pkg
RUN cd pkg && \
  go generate ./...

# copy each file that is specific to the api into the correct directory
WORKDIR /src/bins/${SERVICE}

COPY cmd  ./cmd/
COPY internal ./internal/
COPY main.go .
RUN go generate ./...

FROM code as build

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/${BIN_NAME}

FROM code as test

FROM code as lint

FROM golang:1.24-alpine as final

WORKDIR /src/bins/${SERVICE}

COPY --from=build /bin/${SERVICE} /bin/${BIN_NAME}

ENTRYPOINT ["/bin/nuon"]

FROM code as artifacts

ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out

# mac builds
RUN GOOS=darwin \
    GOARCH=arm64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_arm64 .
RUN GOOS=darwin \
    GOARCH=amd64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_amd64 .

# linux builds
RUN GOOS=linux \
    GOARCH=arm \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm .
RUN GOOS=linux \
    GOARCH=arm64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm64 .

RUN GOOS=linux \
    GOARCH=386 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_386 .
RUN GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_amd64 .
