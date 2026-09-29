FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.25 AS base

WORKDIR /src

COPY --from=nuon go.mod go.sum ./
COPY --from=nuon sdks ./sdks/
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    go mod download


COPY --from=nuon pkg ./pkg/

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    go generate ./pkg/...

FROM base AS test
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    go test -v ./pkg/...
