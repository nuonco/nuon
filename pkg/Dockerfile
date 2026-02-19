FROM golang:1.25 AS base

WORKDIR /src

# Copy module files and local replace targets, then download dependencies
COPY --from=nuon go.mod go.sum ./
COPY --from=nuon sdks ./sdks/
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-pkg \
    go mod download

# Copy pkg source
COPY --from=nuon pkg ./pkg/

# Generate code
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-pkg \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-pkg \
    go generate ./pkg/...

FROM base AS test
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-pkg \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-pkg \
    go test -v ./pkg/...
