import { GraphQLError } from "graphql";
import { GetDeploymentRequest } from "../../build/api/deployment/v1/messages_pb";
import type {
  Deployment,
  Query,
  QueryDeploymentArgs,
  TResolverFn,
} from "../../types";
import { getNodeFields } from "../../utils";

export const deployment: TResolverFn<
  QueryDeploymentArgs,
  Query["deployment"]
> = (_, { id }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.deployment) {
      const request = new GetDeploymentRequest().setId(id);

      clients.deployment.getDeployment(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(getNodeFields<Deployment>(res.toObject().deployment));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
