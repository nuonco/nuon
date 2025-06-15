FROM golang:1.24 as code

RUN apt-get update && apt-get install -y ca-certificates \
    coreutils \
    make \
    git \
    curl \
    jq

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
  | sh -s -- -b $(go env GOPATH)/bin v2.1.6

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

# run golangci-lint to see if we can cache any tools
COPY .golangci.yml .
# RUN golangci-lint run -c .golangci.yml --issues-exit-code 0 -v .

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

COPY .golangci.yml .

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

RUN wget \
        https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem \
    && cat global-bundle.pem >> /global-bundle.pem

FROM golang:1.24-alpine as final

WORKDIR /src/services/ctl-api

ENV PATH=/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/root/.local/bin

# NOTE(JM): this can be removed once `inline: true` is set on all uses of the kube config package
COPY --from=aws-iam-authenticator /bin/aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
COPY --from=certs /global-bundle.pem /etc/ssl/cert.pem
COPY bundle /bundle
COPY --from=build /bin/service /bin/service

ENTRYPOINT ["/bin/service"]
