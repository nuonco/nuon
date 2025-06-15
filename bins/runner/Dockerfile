FROM golang:1.24 as code

WORKDIR /src
ARG SERVICE=runner
ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG PKG_ROOT=github.com/powertoolsdev/mono/bins/runner

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
COPY generate.sh .
RUN go generate ./...

FROM code as build

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/runner

FROM code as test

FROM code as lint

FROM golang:1.24-alpine as nuon-cli

RUN apk add --update --no-cache curl ca-certificates
RUN curl https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli/0.19.519/nuon_linux_amd64 -o /bin/nuon && chmod +x /bin/nuon

FROM gcr.io/kaniko-project/executor as kaniko

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

RUN apk add --no-cache \
  libc6-compat \
  build-base \
  make \
  ca-certificates-bundle \
  coreutils \
  curl \
  protoc \
  jq \
  git \
  aws-cli \
  kubectl \
  zip \
  openssl \
  helm

WORKDIR /src/bins/runner

# add dependencies
COPY --from=kaniko /kaniko/* /kaniko/
COPY --from=kaniko /kaniko/.docker /kaniko/.docker
COPY --from=kaniko /etc/nsswitch.conf /etc/nsswitch.conf
COPY --from=nuon-cli /bin/nuon /usr/bin/nuon
COPY --from=aws-iam-authenticator /bin/aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
COPY --from=certs /global-bundle.pem /etc/ssl/cert.pem

COPY ./bundle /bundle
COPY --from=build /bin/runner /bin/runner

ENTRYPOINT ["/bin/runner"]

FROM scratch as artifacts
