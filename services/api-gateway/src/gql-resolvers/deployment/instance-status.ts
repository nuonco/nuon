import { GetInfoRequest } from "@buf/nuon_orgs-api.grpc_node/instances/v1/info_pb";
import { GetStatusRequest } from "@buf/nuon_orgs-api.grpc_node/instances/v1/status_pb";
import { GraphQLError } from "graphql";
import type { Query, QueryInstanceStatusArgs, TResolverFn } from "../../types";
import { STATUS_ENUM } from "../../utils";

export const instanceStatus: TResolverFn<
  QueryInstanceStatusArgs,
  Query["instanceStatus"]
> = (_, { appId, buildId, componentId, installId, orgId }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.instanceStatus) {
      const infoRequest = new GetInfoRequest()
        .setAppId(appId)
        .setComponentId(componentId)
        .setDeploymentId(buildId)
        .setInstallId(installId)
        .setOrgId(orgId);

      clients.instanceStatus.getInfo(infoRequest, (infoError, infoRes) => {
        if (infoError) {
          reject(new GraphQLError(infoError?.message));
        } else {
          const statusRequest = new GetStatusRequest()
            .setAppId(appId)
            .setComponentId(componentId)
            .setDeploymentId(buildId)
            .setInstallId(installId)
            .setOrgId(orgId);

          clients.instanceStatus.getStatus(
            statusRequest,
            (statusError, statusRes) => {
              if (statusError) {
                reject(new GraphQLError(statusError?.message));
              } else {
                resolve({
                  __typename: "InstanceStatus",
                  hostname: infoRes.toObject()?.hostname,
                  status: STATUS_ENUM[statusRes.toObject()?.status],
                });
              }
            }
          );
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
