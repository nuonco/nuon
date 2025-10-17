ARG BUILD_IMAGE=buildimage

FROM ${BUILD_IMAGE} AS code

ARG BIN_NAME=nuon
ARG SERVICE=cli
ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG PKG_ROOT=github.com/powertoolsdev/mono/bins/${SERVICE}


# copy each file that is specific to the api into the correct directory
WORKDIR /src/bins/${SERVICE}

COPY cmd  ./cmd/
COPY internal ./internal/
COPY main.go .
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    go generate ./...

FROM code AS build

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/${BIN_NAME}

FROM code AS test
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    go test -v ./...

FROM code AS lint
COPY .golangci.yml .
RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-cli \
    golangci-lint run -c .golangci.yml --issues-exit-code 0 -v

FROM golang:1.24-alpine AS final

ARG BIN_NAME=nuon
ARG SERVICE=cli

WORKDIR /src/bins/${BIN_NAME}

COPY --from=build /bin/${BIN_NAME} /bin/${BIN_NAME}

ENTRYPOINT ["/bin/nuon"]

FROM code AS build-darwin-arm64
ARG BIN_NAME=nuon
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
ARG BIN_NAME=nuon
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
ARG BIN_NAME=nuon
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
ARG BIN_NAME=nuon
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
ARG BIN_NAME=nuon
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
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_amd64 .

FROM scratch AS artifacts-build
COPY --from=build-darwin-arm64 /out/ /
COPY --from=build-darwin-amd64 /out/ /
COPY --from=build-linux-arm /out/ /
COPY --from=build-linux-arm64 /out/ /
COPY --from=build-linux-386 /out/ /
COPY --from=build-linux-amd64 /out/ /

FROM amazon/aws-cli:latest AS artifacts

ARG S3_BUCKET_NAME=nuon-artifacts/cli
ARG VERSION
ARG SERVICE=cli
ARG BIN_NAME=nuon

COPY --from=artifacts-build / /artifacts/
COPY install.sh /artifacts/install.sh

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

RUN --mount=type=secret,id=aws_config,target=/root/.aws/config \
    echo "Uploading install.sh to root s3://$S3_BUCKET_NAME/install.sh" && \
    aws s3 cp /artifacts/install.sh "s3://$S3_BUCKET_NAME/install.sh"

FROM amazon/aws-cli:latest AS export-linux-amd64

ARG S3_BUCKET_NAME=nuon-artifacts/cli
ARG VERSION
ARG BIN_NAME=nuon

COPY --from=build-linux-amd64 /out/${BIN_NAME}_linux_amd64 /artifacts/${BIN_NAME}_linux_amd64
RUN sha256sum /artifacts/* > /artifacts/checksums.txt

RUN --mount=type=secret,id=aws_config,target=/root/.aws/config \
    echo "Verifying AWS credentials" && \
    aws sts get-caller-identity

RUN --mount=type=secret,id=aws_config,target=/root/.aws/config \
    echo "Uploading to s3://$S3_BUCKET_NAME/$VERSION/${BIN_NAME}_linux_amd64" && \
    aws s3 cp /artifacts/${BIN_NAME}_linux_amd64 "s3://$S3_BUCKET_NAME/$VERSION/${BIN_NAME}_linux_amd64"

RUN --mount=type=secret,id=aws_config,target=/root/.aws/config \
    echo "Uploading checksums to s3://$S3_BUCKET_NAME/$VERSION/checksums.txt" && \
    aws s3 cp /artifacts/checksums.txt "s3://$S3_BUCKET_NAME/$VERSION/checksums.txt"
