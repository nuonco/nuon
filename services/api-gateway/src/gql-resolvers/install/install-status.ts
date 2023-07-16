import { GraphQLError } from "graphql";
import { GetStatusRequest } from "../../build/orgs-api/installs/v1/status_pb";
import type { Query, QueryInstallStatusArgs, TResolverFn } from "../../types";
import { STATUS_ENUM } from "../../utils";

export const installStatus: TResolverFn<
  QueryInstallStatusArgs,
  Query["installStatus"]
> = (_, { appId, installId, orgId }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.installStatus) {
      const request = new GetStatusRequest()
        .setAppId(appId)
        .setInstallId(installId)
        .setOrgId(orgId);

      clients.installStatus.getStatus(request, (err, res) => {
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
