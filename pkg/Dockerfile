FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.25 AS base

WORKDIR /src

# Copy module files and local replace targets, then download dependencies
COPY --from=nuon go.mod go.sum ./
COPY --from=nuon sdks ./sdks/
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    go mod download


# Copy pkg source
COPY --from=nuon pkg ./pkg/

# Generate code
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    go generate ./pkg/...

FROM base AS test
# Only packages with tests (test-less ones serialize go test); -p 16 overlaps idle test runs with links.
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    PKGS="$(go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./pkg/...)" && \
    go test -v -p 16 ${PKGS}
