import cors from "cors";
import express, { ErrorRequestHandler, Express, RequestHandler } from "express";
import { auth } from "express-oauth2-jwt-bearer";
import { createSchema, createYoga } from "graphql-yoga";
import { version } from "../package.json";
import { resolvers } from "./gql-resolvers";
import { schema as typeDefs } from "./gql-schema";
import { TServiceClients, initServiceClients } from "./grpc-clients";
import { IGQLContext } from "./types";
import { env, logger } from "./utils";

const checkJwt = auth({
  audience: env.AUTH_AUDIENCE,
  issuerBaseURL: env.AUTH_ISSUER,
});

// eslint-disable-next-line
export const errorHandler: ErrorRequestHandler = (err, req, res, _) => {
  logger.error(err);
  res
    .status(err?.statusCode || 500)
    .json({ error: err?.message || "Somethings busted" });
};

export const healthzHandler: RequestHandler = (_, res) => {
  res.status(200).json({ status: "ok", version });
};

/**
 * Higher-order function that returns a context handler function with initiated service clients.
 *
 * The goal is to init the grpc clients once during start up
 * & have the rest of the GQL context created per request.
 */
export const gqlContextHandler =
  (clients) =>
  (context): IGQLContext => {
    logger.debug("Creating GQL context");

    let user = {};
    if (context.req?.auth) {
      user = {
        id: context.req?.auth?.payload?.sub,
      };
    }

    return { clients, user, ...context };
  };

export function initServer(
  clients: TServiceClients = initServiceClients()
): Express {
  logger.debug("Initializing server");
  const server = express();
  const gqlServer = createYoga({
    context: gqlContextHandler(clients),
    schema: createSchema({
      resolvers,
      typeDefs,
    }),
  });

  // add middleware
  server.use(cors());

  // add gql server
  server.use("/graphql", checkJwt, gqlServer);

  // add error handler
  server.use(errorHandler);

  // healthz routes
  server.get("/health", healthzHandler);
  server.get("/readyz", healthzHandler);
  server.get("/livez", healthzHandler);

  return server;
}

export function serverListenHandler(): void {
  logger.info(`Server running at port ${env.PORT}`);
}

export function startServer(): void {
  logger.debug("Attempting to start server");
  const server = initServer();
  server.listen(env.PORT, serverListenHandler);
}
