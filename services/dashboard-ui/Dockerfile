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
COPY public public
COPY src src

FROM deps AS lint

# Run linting with node_modules from deps stage
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

FROM build AS final

ENV NEXT_PUBLIC_PROFILE_ROUTE=/api/auth/profile
ENV NEXT_PUBLIC_ACCESS_TOKEN_ROUTE=/api/auth/access-token

ARG VERSION
ARG GIT_REF
ENV VERSION=$VERSION
ENV GIT_REF=$GIT_REF

EXPOSE 4000
ENTRYPOINT ["npm", "run", "start"]
