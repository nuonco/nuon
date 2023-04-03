import { CreateDeploymentRequest } from "@buf/nuon_apis.grpc_node/deployment/v1/messages_pb";
import { GraphQLError } from "graphql";
import { TDeployment, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

export const createDeployment: TResolverFn<
  { componentId: string },
  TDeployment
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
          resolve(getNodeFields(res.toObject().deployment));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
