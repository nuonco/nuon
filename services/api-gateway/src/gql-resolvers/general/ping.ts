import { GraphQLError } from "graphql";
import { PingRequest } from "../../build/shared/status/v1/ping_pb";
import type { Query, TResolverFn } from "../../types";

export const ping: TResolverFn<undefined, Query["ping"]> = (
  _,
  args,
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.status) {
      const request = new PingRequest();
      clients.status.ping(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err.message));
        } else {
          resolve(res.getStatus());
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
