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

ENV NUON_API_URL=https://api.stage.nuon.co

FROM deps as build

ARG NUON_API_URL=https://api.stage.nuon.co
ENV NEXT_PUBLIC_PROFILE_ROUTE=/api/auth/profile
ENV NEXT_PUBLIC_ACCESS_TOKEN_ROUTE=/api/auth/access-token

RUN npm run build

FROM build as final

ENV NEXT_PUBLIC_PROFILE_ROUTE=/api/auth/profile
ENV NEXT_PUBLIC_ACCESS_TOKEN_ROUTE=/api/auth/access-token

EXPOSE 4000
ENTRYPOINT ["npm", "run", "start"]
