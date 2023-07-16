import { GraphQLError } from "graphql";
import { CreateDeploymentRequest } from "../../build/api/deployment/v1/messages_pb";
import type {
  Deployment,
  Mutation,
  MutationCreateDeploymentArgs,
  TResolverFn,
} from "../../types";
import { getNodeFields } from "../../utils";

export const createDeployment: TResolverFn<
  MutationCreateDeploymentArgs,
  Mutation["createDeployment"]
> = (_, { componentId }, { clients, user }) =>
  new Promise((resolve, reject) => {
    if (clients.deployment) {
      const request = new CreateDeploymentRequest()
        .setComponentId(componentId)
        .setCreatedById(user?.id);

      clients.deployment.createDeployment(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err.message));
        } else {
          resolve(getNodeFields<Deployment>(res.toObject().deployment));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
