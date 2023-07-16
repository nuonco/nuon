import { GraphQLError } from "graphql";
import { GetAppsByOrgRequest } from "../../build/api/app/v1/messages_pb";
import type { App, Query, QueryAppsArgs, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

export const apps: TResolverFn<QueryAppsArgs, Query["apps"]> = (
  _,
  { orgId },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.app) {
      const request = new GetAppsByOrgRequest().setOrgId(orgId);

      clients.app.getAppsByOrg(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          const { appsList } = res.toObject();

          resolve({
            edges:
              appsList?.map((app) => ({
                cursor: app?.id,
                node: getNodeFields<App>(app),
              })) || [],
            pageInfo: {
              endCursor: null,
              hasNextPage: false,
              hasPreviousPage: false,
              startCursor: null,
            },
            totalCount: appsList.length || 0,
          });
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
