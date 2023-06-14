import {
  GetBuildRequest,
  ListBuildsByInstanceRequest,
} from "@buf/nuon_apis.grpc_node/build/v1/messages_pb";
import { GraphQLError } from "graphql";
import type { Build, Query, QueryBuildArgs, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

export const build: TResolverFn<QueryBuildArgs, Query["build"]> = (
  _,
  { id, instanceId },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.build) {
      if (!id && !instanceId) {
        reject(
          new GraphQLError(
            "Invalid query parameters, please provide either a build ID or instance ID"
          )
        );
      }

      if (id) {
        const request = new GetBuildRequest().setId(id);

        clients.build.getBuild(request, (err, res) => {
          if (err) {
            reject(new GraphQLError(err?.message));
          } else {
            resolve(getNodeFields<Build>(res.toObject().build));
          }
        });
      }

      if (instanceId) {
        const request = new ListBuildsByInstanceRequest().setInstanceId(
          instanceId
        );

        clients.build.listBuildsByInstance(request, (err, res) => {
          if (err) {
            reject(new GraphQLError(err?.message));
          } else {
            const { buildsList } = res.toObject();
            const recentBuild = buildsList.reverse()[0];
            resolve(getNodeFields(recentBuild));
          }
        });
      }
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
