FROM golang:1.24 as code

WORKDIR /src
ARG SERVICE=ctl-api
ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG PKG_ROOT=github.com/powertoolsdev/mono/services/ctl-api

# copy deps, tools and fetch dependencies
COPY tools.go go.mod go.sum .
RUN go mod download

# copy package and generate it
COPY pkg pkg
RUN cd pkg && \
  go generate ./...

# copy each file that is specific to the api into the correct directory
WORKDIR /src/services/${SERVICE}

COPY cmd  ./cmd/
COPY internal ./internal/
COPY docs ./docs/
COPY main.go tools.go admin.go public.go runner.go .
RUN go generate ./...

FROM code as build

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service

FROM code as test

FROM code as lint

FROM golang:1.24-alpine as final

WORKDIR /src/services/ctl-api

COPY --from=build /bin/service /bin/service

ENV CGO_ENABLED=0
RUN apk add --no-cache \
  libc6-compat \
  build-base \
  make \
  ca-certificates-bundle \
  coreutils \
  curl \
  protoc \
  bash \
  jq \
  git \
  python3 \
  py3-pip \
  zip && \
  pip3 install awscli --upgrade --user --break-system-packages

ENV PATH=/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/root/.local/bin

ENTRYPOINT ["/bin/service"]
