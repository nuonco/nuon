FROM alpine:3.21 AS openapi-spec

COPY ctl-api/docs/public/swagger-v3.json /openapi/swagger-v3.json

# --- Node deps + SPA source ---

FROM node:20-alpine AS deps

RUN npm i -g bun

WORKDIR /src

RUN apk add --no-cache build-base make python3 py3-pip

COPY \
    .prettierrc \
    vitest.config.ts \
    vitest.setup.ts \
    bun.lock \
    postcss.config.js \
    package.json \
    ./

RUN --mount=type=cache,target=/root/.bun/install/cache,id=bun-dashboard-ui bun install --frozen-lockfile

COPY scripts scripts
COPY public public
COPY client client

# --- Lint targets ---

FROM deps AS lint

RUN bun run lint && bunx tsc --noEmit --project client/tsconfig.json

FROM deps AS lint-local-spec

RUN bun run lint && bunx tsc --noEmit --project client/tsconfig.json

# --- Test targets ---

FROM deps AS test

RUN bun run test

FROM deps AS test-local-spec

COPY --from=openapi-spec /openapi/swagger-v3.json /openapi/swagger-v3.json
ENV NUON_OPENAPI_SPEC_FILE=/openapi/swagger-v3.json

RUN bun run test

# --- Build SPA static assets ---

FROM deps AS build-spa

ARG NUON_API_URL=https://api.stage.nuon.co

RUN --mount=type=cache,target=/root/.bun/install/cache,id=bun-dashboard-ui bun run build

# --- Build Go server ---

FROM public.ecr.aws/p7e3r5y0/mirror/golang:1.25.8-alpine AS build-server

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
COPY pkg pkg
COPY sdks sdks
COPY server services/dashboard-ui/server

RUN go build -o /dashboard-server ./services/dashboard-ui/server

# --- Final image ---

FROM alpine:3.21 AS final

RUN apk add --no-cache ca-certificates

COPY --from=build-spa /src/dist /app/dist
COPY --from=deps /src/public /app/public
COPY --from=build-server /dashboard-server /app/dashboard-server

WORKDIR /app

ARG VERSION
ARG GIT_REF
ENV VERSION=$VERSION
ENV DD_VERSION=$VERSION
ENV GIT_REF=$GIT_REF
ENV DIST_DIR=./dist
ENV PUBLIC_DIR=./public
ENV HTTP_PORT=4000

EXPOSE 4000
ENTRYPOINT ["/app/dashboard-server", "serve"]
