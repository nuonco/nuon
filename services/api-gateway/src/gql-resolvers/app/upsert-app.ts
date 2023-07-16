import { GraphQLError } from "graphql";
import { UpsertAppRequest } from "../../build/api/app/v1/messages_pb";
import type {
  App,
  Mutation,
  MutationUpsertAppArgs,
  TResolverFn,
} from "../../types";
import { getNodeFields } from "../../utils";

export const upsertApp: TResolverFn<
  MutationUpsertAppArgs,
  Mutation["upsertApp"]
> = (_, { input }, { clients, user }) =>
  new Promise((resolve, reject) => {
    if (clients.app) {
      const request = new UpsertAppRequest()
        .setId(input.id)
        .setOrgId(input.orgId)
        .setName(input.name)
        .setCreatedById(user?.id);

      clients.app.upsertApp(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err.message));
        } else {
          resolve(getNodeFields<App>(res.toObject().app));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
