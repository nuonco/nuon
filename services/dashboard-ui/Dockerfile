FROM golang:1.25.4-alpine AS openapi-gen

WORKDIR /src

RUN apk add --no-cache git

# Copy go.mod/go.sum from monorepo root for dependency resolution
COPY go.mod go.sum ./

# Copy pkg directory (required by ctl-api)
COPY pkg pkg

# Copy ctl-api service for spec generation
COPY services/ctl-api services/ctl-api

WORKDIR /src/services/ctl-api

# Generate OpenAPI specs (public, admin, runner) including v3
RUN go run cmd/gen/main.go

FROM node:20-alpine AS deps

WORKDIR /src

RUN apk add --no-cache build-base \
    make \
    python3 \
    py3-pip

# Copy only package files first for better caching
COPY services/dashboard-ui/.eslintrc.json .eslintrc.json
COPY services/dashboard-ui/.prettierrc .prettierrc
COPY services/dashboard-ui/vitest.config.ts vitest.config.ts
COPY services/dashboard-ui/vitest.setup.ts vitest.setup.ts
COPY services/dashboard-ui/package-lock.json package-lock.json
COPY services/dashboard-ui/next.config.mjs next.config.mjs
COPY services/dashboard-ui/postcss.config.js postcss.config.js
COPY services/dashboard-ui/tailwind.config.ts tailwind.config.ts
COPY services/dashboard-ui/package.json package.json
COPY services/dashboard-ui/tsconfig.json tsconfig.json

RUN --mount=type=cache,target=/root/.npm,id=npm-dashboard-ui \
    npm install

COPY services/dashboard-ui/test test
COPY services/dashboard-ui/scripts scripts
COPY services/dashboard-ui/public public
COPY services/dashboard-ui/src src

# Copy generated OpenAPI spec from openapi-gen stage
COPY --from=openapi-gen /src/services/ctl-api/docs/public/swagger-v3.json /openapi/swagger-v3.json

FROM deps AS lint

# Run linting with node_modules from deps stage
RUN npm run lint

FROM deps AS test

# Use locally generated spec instead of fetching from staging
ENV NUON_OPENAPI_SPEC_FILE=/openapi/swagger-v3.json

RUN npm run test -- -t=30000

FROM deps AS build

# Use locally generated spec for type generation
ENV NUON_OPENAPI_SPEC_FILE=/openapi/swagger-v3.json

ENV NEXT_PUBLIC_PROFILE_ROUTE=/api/auth/profile
ENV NEXT_PUBLIC_ACCESS_TOKEN_ROUTE=/api/auth/access-token

RUN --mount=type=cache,target=/root/.npm,id=npm-dashboard-ui \
    npm run build -- --no-lint

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
