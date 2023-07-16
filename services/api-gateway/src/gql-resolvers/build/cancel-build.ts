import { GraphQLError } from "graphql";
import { CancelBuildRequest } from "../../build/api/build/v1/messages_pb";
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
