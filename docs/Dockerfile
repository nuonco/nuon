FROM node:20-alpine as deps

RUN echo "deps"

FROM deps AS final

WORKDIR /app
COPY . .
RUN npm install -g mint

EXPOSE 3000

FROM deps AS build

RUN echo "noop"

FROM deps AS lint

RUN echo "noop"

FROM deps AS test

RUN echo "noop"

FROM deps AS dev

WORKDIR /app
COPY . .
RUN npm install -g mint

EXPOSE 3000

FROM deps as build-artifacts

RUN echo "noop"

FROM deps as artifacts

RUN echo "noop"
