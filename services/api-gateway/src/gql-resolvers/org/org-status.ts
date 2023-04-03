import { GetStatusRequest } from "@buf/nuon_orgs-api.grpc_node/orgs/v1/status_pb";
import { GraphQLError } from "graphql";
import type { Query, QueryOrgStatusArgs, TResolverFn } from "../../types";
import { STATUS_ENUM } from "../../utils";

export const orgStatus: TResolverFn<QueryOrgStatusArgs, Query["orgStatus"]> = (
  _,
  { id },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.orgStatus) {
      const request = new GetStatusRequest().setOrgId(id);

      clients.orgStatus.getStatus(request, (err, res) => {
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
