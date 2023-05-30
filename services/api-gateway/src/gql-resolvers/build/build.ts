import { GetBuildRequest } from "@buf/nuon_apis.grpc_node/build/v1/messages_pb";
import { GraphQLError } from "graphql";
import type { Build, Query, QueryBuildArgs, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

export const build: TResolverFn<QueryBuildArgs, Query["build"]> = (
  _,
  { id },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.build) {
      const request = new GetBuildRequest().setId(id);

      clients.build.getBuild(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(getNodeFields<Build>(res.toObject().build));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
