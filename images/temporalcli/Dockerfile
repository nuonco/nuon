FROM alpine:latest as builder

# Install required packages
RUN mkdir -p /go/bin && \
  apk add --no-cache curl

# Download and extract temporal cli
ARG TARGETPLATFORM
ARG TARGETARCH
RUN echo "building for ${TARGETPLATFORM}_${TARGETARCH}" && \
  curl -L "https://github.com/temporalio/cli/releases/download/v1.2.0/temporal_cli_1.2.0_linux_${TARGETARCH}.tar.gz" | \
  tar -xz -C /go/bin && \
  ls /go/bin

FROM gcr.io/distroless/base-debian11

COPY --from=builder /go/bin/temporal /
EXPOSE 7233
EXPOSE 8233

ENTRYPOINT [   \
  "/temporal", \
  "server",    \
  "start-dev", \
  "--ip",      \
  "0.0.0.0",   \
  "-n", "general",     \
  "-n", "actions",     \
  "-n", "installs",    \
  "-n", "orgs",        \
  "-n", "apps",        \
  "-n", "components",  \
  "-n", "releases",    \
  "-n", "runners",     \
  "-n", "canary",      \
  "-n", "infra-tests", \
  "--db-filename", "/tmp/nuon-dev-temporal.db" \
]
