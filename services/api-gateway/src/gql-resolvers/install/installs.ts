import { GetInstallsByAppRequest } from "@buf/nuon_apis.grpc_node/install/v1/messages_pb";
import { GraphQLError } from "graphql";
import {
  IConnectionResolver,
  TConnection,
  TInstall,
  TResolverFn,
} from "../../types";
import { formatInstall } from "./utils";

interface IInstallsResolver extends IConnectionResolver {
  appId: string;
}

export const installs: TResolverFn<IInstallsResolver, TConnection<TInstall>> = (
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
