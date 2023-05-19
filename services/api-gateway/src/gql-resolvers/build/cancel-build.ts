import { CancelBuildRequest } from "@buf/nuon_apis.grpc_node/build/v1/messages_pb";
import { GraphQLError } from "graphql";
import type {
  Mutation,
  MutationCancelBuildArgs,
  TResolverFn,
} from "../../types";

export const cancelBuild: TResolverFn<
  MutationCancelBuildArgs,
  Mutation["cancelBuild"]
> = (_, { id }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.build) {
      const request = new CancelBuildRequest().setId(id);

      clients.build.cancelBuild(request, (err) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(true);
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
