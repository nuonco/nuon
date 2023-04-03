import { GetDeploymentRequest } from "@buf/nuon_apis.grpc_node/deployment/v1/messages_pb";
import { GraphQLError } from "graphql";
import { TDeployment, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

export const deployment: TResolverFn<{ id: string }, TDeployment> = (
  _,
  { id },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.deployment) {
      const request = new GetDeploymentRequest().setId(id);

      clients.deployment.getDeployment(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(getNodeFields<TDeployment>(res.toObject().deployment));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
