import { GraphQLError } from "graphql";
import { DeleteAppRequest } from "../../build/api/app/v1/messages_pb";
import type { Mutation, MutationDeleteAppArgs, TResolverFn } from "../../types";

export const deleteApp: TResolverFn<
  MutationDeleteAppArgs,
  Mutation["deleteApp"]
> = (_, { id }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.app) {
      const request = new DeleteAppRequest().setId(id);

      clients.app.deleteApp(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(res.toObject().deleted);
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
