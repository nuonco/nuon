import { GraphQLError } from "graphql";
import { GetStatusRequest } from "../../build/orgs-api/deployments/v1/status_pb";
import type {
  Query,
  QueryDeploymentStatusArgs,
  TResolverFn,
} from "../../types";
import { STATUS_ENUM } from "../../utils";

export const deploymentStatus: TResolverFn<
  QueryDeploymentStatusArgs,
  Query["deploymentStatus"]
> = (_, { appId, componentId, deploymentId, orgId }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.deploymentStatus) {
      const request = new GetStatusRequest()
        .setAppId(appId)
        .setComponentId(componentId)
        .setDeploymentId(deploymentId)
        .setOrgId(orgId);

      clients.deploymentStatus.getStatus(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(STATUS_ENUM[res.toObject().status]);
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
