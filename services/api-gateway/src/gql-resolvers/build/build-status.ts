import { GraphQLError } from "graphql";
import { GetStatusRequest } from "../../build/orgs-api/builds/v1/status_pb";
import type { Query, QueryBuildStatusArgs, TResolverFn } from "../../types";
import { STATUS_ENUM } from "../../utils";

export const buildStatus: TResolverFn<
  QueryBuildStatusArgs,
  Query["buildStatus"]
> = (_, { appId, buildId, componentId, orgId }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.buildStatus) {
      const request = new GetStatusRequest()
        .setAppId(appId)
        .setBuildId(buildId)
        .setComponentId(componentId)
        .setOrgId(orgId);

      clients.buildStatus.getStatus(request, (err, res) => {
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
