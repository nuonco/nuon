import { GraphQLError } from "graphql";
import { GetInstallsByAppRequest } from "../../build/api/install/v1/messages_pb";
import type { Query, QueryInstallsArgs, TResolverFn } from "../../types";
import { formatInstall } from "./utils";

export const installs: TResolverFn<QueryInstallsArgs, Query["installs"]> = (
  _,
  { appId },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.install) {
      const request = new GetInstallsByAppRequest().setAppId(appId);

      clients.install.getInstallsByApp(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          const { installsList } = res.toObject();

          resolve({
            edges:
              installsList?.map((install) => ({
                cursor: install?.id,
                node: formatInstall(install),
              })) || [],
            pageInfo: {
              endCursor: null,
              hasNextPage: false,
              hasPreviousPage: false,
              startCursor: null,
            },
            totalCount: installsList?.length || 0,
          });
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
