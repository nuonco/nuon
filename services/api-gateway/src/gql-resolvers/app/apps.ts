import { GetAppsByOrgRequest } from "@buf/nuon_apis.grpc_node/app/v1/messages_pb";
import { GraphQLError } from "graphql";
import {
  IConnectionResolver,
  TApp,
  TConnection,
  TResolverFn,
} from "../../types";
import { getNodeFields } from "../../utils";

interface IAppsResolver extends IConnectionResolver {
  orgId: string;
}

export const apps: TResolverFn<IAppsResolver, TConnection<TApp>> = (
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
                node: getNodeFields(app),
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
