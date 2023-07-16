import { GraphQLError } from "graphql";
import { StartBuildRequest } from "../../build/api/build/v1/messages_pb";
import type {
  Mutation,
  MutationStartBuildArgs,
  TResolverFn,
} from "../../types";

export const startBuild: TResolverFn<
  MutationStartBuildArgs,
  Mutation["startBuild"]
> = (_, { input }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.build) {
      const request = new StartBuildRequest()
        .setComponentId(input.componentId)
        .setGitRef(input.gitRef);

      clients.build.startBuild(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(res.toObject().build);
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
