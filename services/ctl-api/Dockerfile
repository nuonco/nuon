FROM golang:1.24 as code

WORKDIR /src
ARG SERVICE=ctl-api
ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG PKG_ROOT=github.com/powertoolsdev/mono/services/ctl-api
ARG ETCSSL=/etc/ssl/cert.pem

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

FROM code AS build

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service

FROM code AS test

FROM code AS lint

FROM golang:1.24-alpine as aws-iam-authenticator

RUN apk add --update --no-cache curl ca-certificates
WORKDIR /bin
# iam authenticator download url
ARG version=0.6.2
RUN curl \
        -Lo aws-iam-authenticator \
        "https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v${version}/aws-iam-authenticator_${version}_linux_amd64" \
    && chmod +x aws-iam-authenticator

FROM golang:1.24-alpine as certs

# NOTE(jdt): ideally we would wget -a but busybox doesn't support it
# or use `update-ca-certificate` but it can't handle a single file w/ multiple certs
RUN wget \
        https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem \
    && cat global-bundle.pem >> /global-bundle.pem

FROM golang:1.24-alpine as final

WORKDIR /src/services/ctl-api

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

COPY --from=aws-iam-authenticator /bin/aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
COPY --from=certs /global-bundle.pem /etc/ssl/cert.pem
COPY bundle /bundle
COPY --from=build /bin/service /bin/service

ENTRYPOINT ["/bin/service"]
