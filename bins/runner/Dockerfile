ARG BUILD_IMAGE=buildimage

FROM ${BUILD_IMAGE} AS code

WORKDIR /src
ARG SERVICE=runner
ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG PKG_ROOT=github.com/powertoolsdev/mono/bins/${SERVICE}


# copy each file that is specific to the api into the correct directory
WORKDIR /src/bins/${SERVICE}

COPY cmd  ./cmd/
COPY internal ./internal/
COPY main.go .
COPY generate.sh .
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-runner \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-runner \
    go generate ./...

FROM code AS build

ARG tag=

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-runner \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-runner \
    CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /bin/runner

FROM code AS test
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild-runner \
    go test -v ./...

FROM code AS lint
COPY .golangci.yml .
RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-runner \
    golangci-lint run -c .golangci.yml --issues-exit-code 0 -v

FROM golang:1.24-alpine AS nuon-cli

RUN apk add --update --no-cache curl ca-certificates
RUN curl https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli/0.19.602/nuon_linux_amd64 -o /bin/nuon && chmod +x /bin/nuon

FROM gcr.io/kaniko-project/executor AS kaniko

FROM golang:1.24-alpine AS aws-iam-authenticator

RUN apk add --update --no-cache curl ca-certificates
WORKDIR /bin
# iam authenticator download url
ARG version=0.6.2
RUN curl \
    -Lo aws-iam-authenticator \
    "https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v${version}/aws-iam-authenticator_${version}_linux_amd64" \
    && chmod +x aws-iam-authenticator

FROM golang:1.24-alpine AS kubelogin

RUN apk add --update --no-cache curl ca-certificates
WORKDIR /bin
# kubelogin download url
RUN curl -LO \
    https://github.com/Azure/kubelogin/releases/download/v0.2.9/kubelogin-linux-amd64.zip \
    && unzip kubelogin-linux-amd64.zip \
    && mv bin/linux_amd64/kubelogin . \
    && rm kubelogin-linux-amd64.zip \
    && rm -r bin \
    && chmod +x kubelogin

FROM golang:1.24-alpine AS certs

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

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=darwin \
    GOARCH=arm64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_arm64 .

FROM code AS build-darwin-amd64
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=darwin \
    GOARCH=amd64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_amd64 .

FROM code AS build-linux-arm
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=linux \
    GOARCH=arm \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm .

FROM code AS build-linux-arm64
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=linux \
    GOARCH=arm64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm64 .

FROM code AS build-linux-386
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=linux \
    GOARCH=386 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_386 .

FROM code AS build-linux-amd64
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_amd64 .

# END

FROM golang:1.24-alpine AS final

ARG tag

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
    helm

WORKDIR /src/bins/runner

# add dependencies
COPY --from=kaniko /kaniko/* /kaniko/
COPY --from=kaniko /kaniko/.docker /kaniko/.docker
COPY --from=kaniko /etc/nsswitch.conf /etc/nsswitch.conf
COPY --from=nuon-cli /bin/nuon /usr/bin/nuon
COPY --from=aws-iam-authenticator /bin/aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
COPY --from=kubelogin /bin/kubelogin /usr/local/bin/kubelogin
COPY --from=certs /global-bundle.pem /etc/ssl/cert.pem

COPY ./bundle /bundle
COPY --from=build /bin/runner /bin/runner

# udpate the ca certs so python's urllib cert verification passes
RUN update-ca-certificates


ENV GIT_REF=$tag

ENTRYPOINT ["/bin/runner"]

FROM build AS build-artifacts

ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-runner \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-runner \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_amd64 .

FROM scratch AS artifacts-build
COPY --from=build-artifacts /out /
COPY --from=build-darwin-arm64 /out/ /
COPY --from=build-darwin-amd64 /out/ /
COPY --from=build-linux-arm /out/ /
COPY --from=build-linux-arm64 /out/ /
COPY --from=build-linux-386 /out/ /
COPY --from=build-linux-amd64 /out/ /

FROM amazon/aws-cli:latest AS artifacts

ARG S3_BUCKET_NAME=nuon-artifacts/runner
ARG VERSION
ARG SERVICE=runner
ARG BIN_NAME=runner

COPY --from=artifacts-build / /artifacts/

RUN cd /artifacts && sha256sum * > checksums.txt

RUN --mount=type=secret,id=aws_config,target=/root/.aws/config \
    echo "Verifying AWS credentials" && \
    aws sts get-caller-identity

RUN --mount=type=secret,id=aws_config,target=/root/.aws/config \
    for file in $(ls /artifacts/ | grep -v "checksums.txt"); do \
      echo "Uploading to s3://$S3_BUCKET_NAME/$VERSION/$file" && \
      aws s3 cp "/artifacts/$file" "s3://$S3_BUCKET_NAME/$VERSION/$file"; \
    done

RUN --mount=type=secret,id=aws_config,target=/root/.aws/config \
    echo "Uploading checksums to s3://$S3_BUCKET_NAME/$VERSION/checksums.txt" && \
    aws s3 cp /artifacts/checksums.txt "s3://$S3_BUCKET_NAME/$VERSION/checksums.txt"
