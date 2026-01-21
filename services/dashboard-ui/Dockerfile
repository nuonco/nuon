# =============================================================================
# OpenAPI spec generation stage (only used with gen-api-types label)
# =============================================================================
FROM golang:1.25.4-alpine AS openapi-gen

WORKDIR /src

RUN apk add --no-cache git

# Copy go.mod/go.sum (copied to build context by sources)
COPY go.mod go.sum ./

# Copy pkg directory (copied to build context by sources)
COPY pkg pkg

# Copy sdks directory (required by go.mod replace directive)
COPY sdks sdks

# Copy ctl-api service - maintain proper path for module resolution
COPY ctl-api services/ctl-api

WORKDIR /src/services/ctl-api

# Generate OpenAPI specs (public, admin, runner) including v3
RUN go run cmd/gen/main.go

# =============================================================================
# Base deps stage (shared by all targets)
# =============================================================================
FROM node:20-alpine AS deps

WORKDIR /src

RUN apk add --no-cache build-base \
    make \
    python3 \
    py3-pip

# Copy only package files first for better caching
COPY \
    .eslintrc.json \
    .prettierrc \
    vitest.config.ts \
    vitest.setup.ts \
    package-lock.json \
    next.config.mjs \
    postcss.config.js \
    tailwind.config.ts \
    package.json \
    tsconfig.json \
    ./
RUN --mount=type=cache,target=/root/.npm,id=npm-dashboard-ui \
    npm install

COPY test test
COPY scripts scripts
COPY public public
COPY src src

# =============================================================================
# Deps with local spec (for gen-api-types label)
# =============================================================================
FROM deps AS deps-local-spec

COPY --from=openapi-gen /src/services/ctl-api/docs/public/swagger-v3.json /openapi/swagger-v3.json

# =============================================================================
# Standard targets (default - fetch from staging)
# =============================================================================
FROM deps AS lint

RUN npm run lint

FROM deps AS test

ENV NUON_API_URL=https://api.stage.nuon.co

RUN npm run test -- -t=30000

FROM deps AS build

ARG NUON_API_URL=https://api.stage.nuon.co
ENV NEXT_PUBLIC_PROFILE_ROUTE=/api/auth/profile
ENV NEXT_PUBLIC_ACCESS_TOKEN_ROUTE=/api/auth/access-token

RUN --mount=type=cache,target=/root/.npm,id=npm-dashboard-ui \
    npm run build -- --no-lint

# =============================================================================
# Local spec targets (for gen-api-types label)
# =============================================================================
FROM deps-local-spec AS lint-local-spec

RUN npm run lint

FROM deps-local-spec AS test-local-spec

ENV NUON_OPENAPI_SPEC_FILE=/openapi/swagger-v3.json

RUN npm run test -- -t=30000

FROM deps-local-spec AS build-local-spec

ENV NUON_OPENAPI_SPEC_FILE=/openapi/swagger-v3.json
ENV NEXT_PUBLIC_PROFILE_ROUTE=/api/auth/profile
ENV NEXT_PUBLIC_ACCESS_TOKEN_ROUTE=/api/auth/access-token

RUN --mount=type=cache,target=/root/.npm,id=npm-dashboard-ui \
    npm run build -- --no-lint

# =============================================================================
# Final stage
# =============================================================================
FROM build AS final

ENV NEXT_PUBLIC_PROFILE_ROUTE=/api/auth/profile
ENV NEXT_PUBLIC_ACCESS_TOKEN_ROUTE=/api/auth/access-token

ARG VERSION
ARG GIT_REF
ENV VERSION=$VERSION
ENV DD_VERSION=$VERSION
ENV GIT_REF=$GIT_REF

EXPOSE 4000
ENTRYPOINT ["npm", "run", "start"]
