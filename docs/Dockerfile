FROM node:20-alpine as deps

RUN echo "deps"

FROM deps AS final

WORKDIR /app
COPY . .
RUN npm install -g mint

EXPOSE 3000

FROM deps AS dev

WORKDIR /app
COPY . .
RUN npm install -g mint

EXPOSE 3000
