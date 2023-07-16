import { GraphQLError } from "graphql";
import { GetOrgRequest } from "../../build/api/org/v1/messages_pb";
import type { Org, Query, QueryOrgArgs, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

export const org: TResolverFn<QueryOrgArgs, Query["org"]> = (
  _,
  { id },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.org) {
      const request = new GetOrgRequest().setId(id);

      clients.org.getOrg(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(getNodeFields<Org>(res.toObject().org));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
