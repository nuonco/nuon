FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/oven/bun:1.3-alpine AS client

WORKDIR /src/services/bundle-portal/client

COPY services/bundle-portal/client/package.json services/bundle-portal/client/bun.lock ./
RUN --mount=type=cache,target=/root/.bun/install/cache,id=bun-bundle-portal bun install --frozen-lockfile

COPY services/bundle-portal/client ./
RUN bun run build

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.25.8-alpine AS build

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
COPY pkg pkg
COPY services/bundle-portal services/bundle-portal
COPY --from=client /src/services/bundle-portal/client/dist services/bundle-portal/client/dist

RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild-bundle-portal \
    CGO_ENABLED=0 go build -o /bundle-portal ./services/bundle-portal

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/alpine:3.21 AS final

RUN apk add --no-cache ca-certificates

COPY --from=build /bundle-portal /app/bundle-portal

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/bundle-portal"]
