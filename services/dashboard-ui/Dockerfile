FROM alpine:3.21 AS openapi-spec

COPY ctl-api/docs/public/swagger-v3.json /openapi/swagger-v3.json

FROM node:20-alpine AS deps

WORKDIR /src

RUN apk add --no-cache build-base make python3 py3-pip

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

RUN --mount=type=cache,target=/root/.npm,id=npm-dashboard-ui npm install

COPY test test
COPY scripts scripts
COPY public public
COPY src src

FROM deps AS lint

RUN npm run lint

FROM deps AS test

ENV NUON_API_URL=https://api.stage.nuon.co

RUN npm run test -- -t=30000

FROM deps AS build

ARG NUON_API_URL=https://api.stage.nuon.co
ENV NEXT_PUBLIC_PROFILE_ROUTE=/api/auth/profile
ENV NEXT_PUBLIC_ACCESS_TOKEN_ROUTE=/api/auth/access-token

RUN --mount=type=cache,target=/root/.npm,id=npm-dashboard-ui npm run build -- --no-lint

FROM deps AS lint-local-spec

RUN npm run lint

FROM deps AS test-local-spec

COPY --from=openapi-spec /openapi/swagger-v3.json /openapi/swagger-v3.json
ENV NUON_OPENAPI_SPEC_FILE=/openapi/swagger-v3.json

RUN npm run test -- -t=30000

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
