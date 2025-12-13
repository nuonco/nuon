FROM node:20-alpine AS deps
WORKDIR /src

COPY package-lock.json package.json ./
RUN npm ci

FROM deps AS build
COPY . .
RUN npm run build

FROM node:20-alpine AS final
WORKDIR /src
COPY --from=build /src/.next .next
COPY --from=build /src/public public
COPY --from=build /src/node_modules node_modules
COPY --from=build /src/package.json package.json

ARG VERSION
ENV VERSION=$VERSION

EXPOSE 4100
ENTRYPOINT ["npm", "run", "start"]
