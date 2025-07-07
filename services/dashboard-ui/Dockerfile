FROM node:18-alpine as deps

WORKDIR /src

RUN apk add --update build-base \
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

# Add cache mounts for npm dependencies
RUN --mount=type=cache,target=/src/node_modules \
    --mount=type=cache,target=/root/.npm \
    npm install

COPY test test
COPY public public
COPY src src

FROM deps AS lint

# Run linting with cache mounts
RUN --mount=type=cache,target=/src/node_modules \
    npm run lint

FROM deps AS test

ENV NUON_API_URL=https://api.stage.nuon.co

# Run tests with cache mounts
RUN --mount=type=cache,target=/src/node_modules \
npm run test -- -t=30000

FROM deps as build

ARG NUON_API_URL=https://api.stage.nuon.co
ENV NEXT_PUBLIC_PROFILE_ROUTE=/api/auth/profile
ENV NEXT_PUBLIC_ACCESS_TOKEN_ROUTE=/api/auth/access-token

# Add cache mount for the build process
RUN --mount=type=cache,target=/src/node_modules \
    --mount=type=cache,target=/src/.next/cache \
    npm run build

FROM build as final

ENV NEXT_PUBLIC_PROFILE_ROUTE=/api/auth/profile
ENV NEXT_PUBLIC_ACCESS_TOKEN_ROUTE=/api/auth/access-token

ARG VERSION
ARG GIT_REF
ENV VERSION=$VERSION
ENV GIT_REF=$GIT_REF

EXPOSE 4000
ENTRYPOINT ["npm", "run", "start"]
