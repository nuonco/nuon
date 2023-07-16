import { GraphQLError } from "graphql";
import { StartDeployRequest } from "../../build/api/deploy/v1/messages_pb";
import type {
  Mutation,
  MutationStartDeployArgs,
  TResolverFn,
} from "../../types";

export const startDeploy: TResolverFn<
  MutationStartDeployArgs,
  Mutation["startDeploy"]
> = (_, { input }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.deploy) {
      const request = new StartDeployRequest()
        .setBuildId(input.buildId)
        .setComponentId(input.componentId)
        .setInstallId(input.installId);

      clients.deploy.startDeploy(request, (err, res) => {
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
