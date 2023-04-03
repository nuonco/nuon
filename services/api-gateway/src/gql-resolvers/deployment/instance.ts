import { GetInfoRequest } from "@buf/nuon_orgs-api.grpc_node/instances/v1/info_pb";
import { GetStatusRequest } from "@buf/nuon_orgs-api.grpc_node/instances/v1/status_pb";
import { GraphQLError } from "graphql";
import type { Query, QueryInstanceArgs, TResolverFn } from "../../types";
import { STATUS_ENUM } from "../../utils";

export const instance: TResolverFn<QueryInstanceArgs, Query["instance"]> = (
  _,
  { appId, componentId, deploymentId, installId, orgId },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.instance) {
      const infoRequest = new GetInfoRequest()
        .setAppId(appId)
        .setComponentId(componentId)
        .setDeploymentId(deploymentId)
        .setInstallId(installId)
        .setOrgId(orgId);

      clients.instance.getInfo(infoRequest, (infoError, infoRes) => {
        if (infoError) {
          reject(new GraphQLError(infoError?.message));
        } else {
          const statusRequest = new GetStatusRequest()
            .setAppId(appId)
            .setComponentId(componentId)
            .setDeploymentId(deploymentId)
            .setInstallId(installId)
            .setOrgId(orgId);

          clients.instance.getStatus(
            statusRequest,
            (statusError, statusRes) => {
              if (statusError) {
                reject(new GraphQLError(statusError?.message));
              } else {
                resolve({
                  __typename: "Instance",
                  hostname: infoRes.toObject()?.response?.hostname,
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
