FROM node:alpine as deps

WORKDIR /src

RUN apk add --update build-base \
  make \
  python3 \
  py3-pip

# Copy only package files first for better caching
COPY \
  package.json \
  package-lock.json \
  astro.config.mjs \
  tailwind.config.mjs \
  tsconfig.json \
 ./
RUN --mount=type=cache,target=/root/.npm,id=npm-wiki \
npm install

COPY src src
COPY public public

FROM deps AS lint

# Run linting
RUN npm run lint

FROM deps AS test

# Run tests
RUN npm run test -- run

FROM deps as build

RUN rm -rf src/content/docs
COPY wiki/ src/content/docs

RUN --mount=type=cache,target=/root/.npm,id=npm-wiki \
npm run build

FROM build as final

ENV HOST=0.0.0.0
ENV PORT=4321

ARG VERSION
ARG GIT_REF
ENV VERSION=$VERSION
ENV GIT_REF=$GIT_REF

EXPOSE 4321
ENTRYPOINT ["node", "./dist/server/entry.mjs"]
