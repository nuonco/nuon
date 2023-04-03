import { GetStatusRequest } from "@buf/nuon_orgs-api.grpc_node/deployments/v1/status_pb";
import { GraphQLError } from "graphql";
import { TResolverFn } from "../../types";
import { STATUS_ENUM } from "../../utils";

export const deploymentStatus: TResolverFn<
  { appId: string; componentId: string; deploymentId: string; orgId: string },
  string
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
