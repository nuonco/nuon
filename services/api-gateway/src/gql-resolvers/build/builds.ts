import { GraphQLError } from "graphql";
import { QueryBuildsRequest } from "../../build/api/build/v1/messages_pb";
import type { Query, QueryBuildsArgs, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

export const builds: TResolverFn<QueryBuildsArgs, Query["builds"]> = (
  _,
  { componentId },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.build) {
      const request = new QueryBuildsRequest().setComponentId(componentId);

      clients.build.queryBuilds(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          const { buildsList } = res.toObject();
          resolve(buildsList.reverse().map(getNodeFields) || []);
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
