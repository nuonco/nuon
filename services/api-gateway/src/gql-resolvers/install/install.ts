import { GetInstallRequest } from "@buf/nuon_apis.grpc_node/install/v1/messages_pb";
import { GraphQLError } from "graphql";
import type { Query, QueryInstallArgs, TResolverFn } from "../../types";
import { formatInstall } from "./utils";

export const install: TResolverFn<QueryInstallArgs, Query["install"]> = (
  _,
  { id },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.install) {
      const request = new GetInstallRequest().setId(id);

      clients.install.getInstall(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(formatInstall(res.toObject().install));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
