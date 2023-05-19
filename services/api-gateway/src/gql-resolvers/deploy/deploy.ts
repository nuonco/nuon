import { GetDeployRequest } from "@buf/nuon_apis.grpc_node/deploy/v1/messages_pb";
import { GraphQLError } from "graphql";
import type { Query, QueryDeployArgs, TResolverFn } from "../../types";

export const deploy: TResolverFn<QueryDeployArgs, Query["deploy"]> = (
  _,
  { id },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.deploy) {
      const request = new GetDeployRequest().setId(id);
      clients.deploy.getDeploy(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(res.toObject().deploy);
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
