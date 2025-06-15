FROM node:18-alpine as deps

WORKDIR /src

RUN apk add --update build-base \
  make \
  python3 \
  py3-pip

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
RUN \
npm install

COPY test test
COPY public public
COPY src src

FROM deps AS test

RUN echo "running unit test..." && \
  npm run test

FROM deps AS lint

RUN echo "running lint..." && \
  npm run lint && \
  npm run tsc

FROM deps as build

ARG NUON_API_URL=https://api.stage.nuon.co
RUN npm run build

FROM build as final

EXPOSE 4000
ENTRYPOINT ["npm", "run", "start"]
