ARG BUILD_IMAGE=buildimage
ARG NUON_CLI_VERSION=0.19.602
ARG AWS_IAM_AUTHENTICATOR_VERSION=0.6.2
ARG KUBELOGIN_VERSION=0.2.9
ARG GCLOUD_VERSION=516.0.0
ARG AZURE_CLI_VERSION=2.73.0
ARG PULUMI_VERSION=3.234.0

FROM ${BUILD_IMAGE} AS code

WORKDIR /src
ARG SERVICE=runner
ARG PKG_ROOT=github.com/nuonco/nuon/bins/${SERVICE}


# copy each file that is specific to the api into the correct directory
WORKDIR /src/bins/${SERVICE}

# Copy monorepo shared sources (SDK go.mod/go.sum are in the base image)
COPY --link sdks /src/sdks/
COPY --link pkg /src/pkg/

# Copy service-specific files (most frequently changed, last for cache efficiency)
COPY --link cmd  ./cmd/
COPY --link internal ./internal/
COPY --link main.go .

FROM code AS build

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    CGO_ENABLED=0 \
    GOOS=linux \
    go build -o /bin/runner

FROM code AS build-versioned

ARG tag=

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /bin/runner

FROM code AS test
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    go test -v ./...

FROM code AS lint
COPY .golangci.yml .
RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-runner \
    --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    golangci-lint run -c .golangci.yml -v

FROM golang:1.25-alpine AS nuon-cli

ARG NUON_CLI_VERSION

RUN apk add --update --no-cache curl ca-certificates
RUN curl "https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli/${NUON_CLI_VERSION}/nuon_linux_amd64" -o /bin/nuon && chmod +x /bin/nuon

FROM public.ecr.aws/p7e3r5y0/kaniko-project/executor:latest  AS kaniko

FROM golang:1.25-alpine AS aws-iam-authenticator

ARG AWS_IAM_AUTHENTICATOR_VERSION

RUN apk add --update --no-cache curl ca-certificates
WORKDIR /bin
RUN curl \
    -Lo aws-iam-authenticator \
    "https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v${AWS_IAM_AUTHENTICATOR_VERSION}/aws-iam-authenticator_${AWS_IAM_AUTHENTICATOR_VERSION}_linux_amd64" \
    && chmod +x aws-iam-authenticator

FROM golang:1.25-alpine AS kubelogin

ARG KUBELOGIN_VERSION

RUN apk add --update --no-cache curl ca-certificates
WORKDIR /bin
# kubelogin download url
RUN curl -LO \
    "https://github.com/Azure/kubelogin/releases/download/v${KUBELOGIN_VERSION}/kubelogin-linux-amd64.zip" \
    && unzip kubelogin-linux-amd64.zip \
    && mv bin/linux_amd64/kubelogin . \
    && rm kubelogin-linux-amd64.zip \
    && rm -r bin \
    && chmod +x kubelogin

FROM golang:1.25-alpine AS gcloud

ARG GCLOUD_VERSION

RUN apk add --update --no-cache curl ca-certificates python3
WORKDIR /opt
RUN curl -sSL "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-${GCLOUD_VERSION}-linux-x86_64.tar.gz" | tar xz \
    && /opt/google-cloud-sdk/install.sh --quiet --usage-reporting false --path-update false \
    && /opt/google-cloud-sdk/bin/gcloud components install gke-gcloud-auth-plugin --quiet

FROM python:3.12-alpine AS azure-cli

ARG AZURE_CLI_VERSION

RUN apk add --update --no-cache gcc musl-dev libffi-dev
RUN pip install --no-cache-dir --prefix=/azure-cli azure-cli==${AZURE_CLI_VERSION}

FROM golang:1.25-alpine AS pulumi

ARG PULUMI_VERSION

RUN apk add --update --no-cache curl ca-certificates
WORKDIR /opt
RUN curl -sSL "https://get.pulumi.com/releases/sdk/pulumi-v${PULUMI_VERSION}-linux-x64.tar.gz" | tar xz

FROM golang:1.25-alpine AS certs

# NOTE(jdt): ideally we would wget -a but busybox doesn't support it
# or use `update-ca-certificate` but it can't handle a single file w/ multiple certs
RUN wget \
    https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem \
    && cat global-bundle.pem >> /global-bundle.pem

# START

FROM code AS build-darwin-arm64

ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=darwin \
    GOARCH=arm64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_arm64 .

FROM code AS build-darwin-amd64
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=darwin \
    GOARCH=amd64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_amd64 .

FROM code AS build-linux-arm
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=linux \
    GOARCH=arm \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm .

FROM code AS build-linux-arm64
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=linux \
    GOARCH=arm64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm64 .

FROM code AS build-linux-386
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=linux \
    GOARCH=386 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_386 .

FROM code AS build-linux-amd64
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_amd64 .

# END

FROM golang:1.25-alpine AS runtime-base

RUN apk add --no-cache \
    bash \
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
    helm \
    python3 \
    oras-cli

# add dependencies
COPY --from=kaniko /kaniko/* /kaniko/
COPY --from=kaniko /kaniko/.docker /kaniko/.docker
COPY --from=kaniko /etc/nsswitch.conf /etc/nsswitch.conf
COPY --from=nuon-cli /bin/nuon /usr/bin/nuon
COPY --from=aws-iam-authenticator /bin/aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
COPY --from=kubelogin /bin/kubelogin /usr/local/bin/kubelogin
COPY --from=gcloud /opt/google-cloud-sdk /opt/google-cloud-sdk
COPY --from=azure-cli /azure-cli /usr/local
COPY --from=pulumi /opt/pulumi /opt/pulumi
COPY --from=certs /global-bundle.pem /etc/ssl/cert.pem

ENV PATH="/opt/pulumi:/opt/google-cloud-sdk/bin:${PATH}"

# udpate the ca certs so python's urllib cert verification passes
RUN update-ca-certificates

FROM runtime-base AS final

WORKDIR /src/bins/runner

COPY ./bundle /bundle
COPY --from=build-versioned /bin/runner /bin/runner

ARG tag
ENV GIT_REF=$tag

ENTRYPOINT ["/bin/runner"]

FROM scratch AS artifacts-build
COPY --from=build-darwin-arm64 /out/ /
COPY --from=build-darwin-amd64 /out/ /
COPY --from=build-linux-arm /out/ /
COPY --from=build-linux-arm64 /out/ /
COPY --from=build-linux-386 /out/ /
COPY --from=build-linux-amd64 /out/ /

# Prepare artifacts for upload
# Go code will handle the S3 upload part
FROM alpine:latest AS artifacts

ARG BIN_NAME=runner

COPY --from=artifacts-build / /artifacts/
COPY install.sh /artifacts/install.sh
