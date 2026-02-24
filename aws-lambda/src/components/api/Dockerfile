FROM debian:bookworm-slim AS build

# --- Install a clean Go toolchain into /opt/go ---
ARG GO_VERSION=1.21.13
ARG GO_TARBALL=go${GO_VERSION}.linux-amd64.tar.gz
ARG GO_SHA256=502fc16d5910562461e6a6631fb6377de2322aad7304bf2bcd23500ba9dab4a7

RUN apt-get update && \
    rm -f /etc/ssl/openssl.cnf || true && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      ca-certificates curl git xz-utils && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /tmp
RUN curl -fsSLO https://go.dev/dl/${GO_TARBALL} \
 && echo "${GO_SHA256}  ${GO_TARBALL}" | sha256sum -c -

RUN mkdir -p /opt && tar -C /opt -xzf ${GO_TARBALL} && mv /opt/go /opt/go${GO_VERSION}

# Go env (avoid any runner overlays)
ENV GOROOT=/opt/go${GO_VERSION}
ENV GOPATH=/go
ENV PATH=$GOROOT/bin:$GOPATH/bin:$PATH
ENV GOTOOLCHAIN=local
ENV GOEXPERIMENT=
ENV GOFLAGS=
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOMODCACHE=/go/pkg/mod

# --- Build the api module (context is this folder) ---
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Defensive: if any stray iter sources exist, remove them
RUN rm -rf ${GOROOT}/src/iter || true \
 && rm -f  ${GOROOT}/src/bytes/iter.go || true

RUN go version && CGO_ENABLED=0 GOOS=linux go build -o /api .

# --- Minimal runtime ---
FROM alpine:3.19
COPY --from=build /api /main
ENTRYPOINT ["/main"]
